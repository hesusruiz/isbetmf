package service

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// ListGenericObjects retrieves all TMF objects of a given type.
// List is the most complex operation due to access control, which makes it very difficult to cache the requests efficiently.
func (svc *Service) ListGenericObjects(req *Request) *Response {
	var err error
	if !req.HealthRequest {
		slog.Debug("ListGenericObjects called", slog.String("resourceName", req.ResourceName))
	}
	// Authentication: anonymous access is allowed for some objects, so we do not check the existence of the access token at this moment

	// Defaut values for pagination, meaning that they were not specified by the user
	var userLimit = 10
	var userOffset = 0

	// Get the user-specified values for pagination
	limitStr := req.QueryParams.Get("limit")
	if value, err := strconv.Atoi(limitStr); err == nil {
		userLimit = value
	}

	offsetStr := req.QueryParams.Get("offset")
	if value, err := strconv.Atoi(offsetStr); err == nil {
		userOffset = value
	}

	// Respond immediately if the caller wanted zero objects
	if userLimit == 0 {
		var objs []repo.TMFRecord
		return &Response{StatusCode: http.StatusOK, Body: objs}
	}

	// Prepare the filter for partial field selection
	fieldsParam := req.QueryParams.Get("fields")
	fieldSet := make(map[string]bool, 10)

	if fieldsParam != "" {

		// These are always required
		fieldSet = map[string]bool{
			"id":              true,
			"href":            true,
			"lastUpdate":      true,
			"version":         true,
			"@type":           true,
			"lifecycleStatus": true,
		}

		// If the user specified 'none' we will send back the required fields.
		// Otherwise, we will add the ones specified by the user, if different from required ones.
		if fieldsParam != "none" {
			fields := strings.SplitSeq(fieldsParam, ",")

			// Add the fields specified by the user
			for f := range fields {
				fieldSet[strings.TrimSpace(f)] = true
			}
		}
	}

	if svc.proxyEnabled {

		// Delete the attribute selection for the query to the backend. We will receice full objects and
		// perform attribute selection ourselves. This is because we want to store the full objects in our local cache.
		req.QueryParams.Del("fields")

		// Will send the authorization header
		// TODO: set headers to avoid caching
		headers := map[string]string{
			"Authorization": "Bearer " + req.AccessToken,
		}

		// We will accumulate the objects in this array, and return it to the caller.
		// This is necessary because we need to know the total number of objects that match the query,
		// and the caller may specify a limit that is greater than the number of objects that match the query.
		// In this case, we will return the objects that match the query, and the total number of objects that match the query.
		responseObjectMaps := make([]repo.TMFObjectMap, 0)
		// diagnosticMessages := make([]string, 0)
		diagnosticObjects := make([]repo.ValidationResult, 0)
		var offsetCounter int
		invalidObjects := 0
		diagnostic := false

		// Loop retrieving pages of objects until we fulfill the request of the user. We have to do this because of access control.
		// The strategy is the following:
		//   - start from offset 0, counting the objects suitable for the user, until we reach the user-specified offset
		//   - continue retrieving objects suitable for the user until we get the number specified by the user (limit)
		// We have to do this because we can not predict how many objects will be discarded by access control until we reach offset.
		// To accelerate the process, we will use in the future our cache to serve requests from the local database.
		// TODO: implement caching of replies from the remote server, when testing has been performed on the normal flow.
		pageSize := 100
		pageOffset := 0
		for {

			// Set (maybe overwriting user-specified values) the offset and limit to the current values for the query to the remote server.
			req.QueryParams.Set("offset", strconv.Itoa(pageOffset))
			req.QueryParams.Set("limit", strconv.Itoa(pageSize))

			// Get the diagnostic flag set by the user
			diagnostic = req.QueryParams.Has("diagnostic")
			if diagnostic {
				slog.Debug("diagnostic mode enabled")
				// Delete the diagnostic flag from the query parameters. The upstream server will not understand it.
				req.QueryParams.Del("diagnostic")
			}

			// Encode all query parameters sorted by key, which is useful for caching the reply using the path as key
			parameters := req.QueryParams.Encode()
			if parameters != "" {
				parameters = "?" + parameters
			}

			// Build the path with the canonicalized parameters
			path := fmt.Sprintf("/tmf-api/%s/%s/%s%s", req.APIfamily, req.APIVersion, req.ResourceName, parameters)
			if !req.HealthRequest {
				slog.Debug("sending request to remote", "path", path)
			}
			// Retrieve the page from the remote server

			responseData, err := svc.tmfClient.ForwardTMFGetList(path, headers)
			if err != nil {
				return ErrorResponsef(http.StatusInternalServerError, "upstream server failed with error: %w", err)
			}

			// Process the list of response objects to filter depending on the user
			for _, responseObject := range responseData {

				// Prepare object for the database, validating it to check for errors
				objectMap, validations := repo.NewTMFObjectMapFromUpstream(req.ResourceName, responseObject)
				if len(validations.Errors) > 0 {
					invalidObjects++
					fmt.Println(validations.String())
					diagnosticObjects = append(diagnosticObjects, validations)
					continue
				}

				obj := objectMap.ToTMFRecord(req.ResourceName)

				// Store locally for further usage
				if err := svc.upsertObject(obj); err != nil {
					if errors.Is(err, &ErrObjectExists{}) {
						slog.Debug("object already exists", "id", obj.ID)
					} else {
						invalidObjects++
						err = errl.Errorf("error saving object %s in local database: %w", obj.ID, err)
						slog.Error("error saving object in local database", "error", err)
						continue
					}
				}

				// Check if the object can be read by the user
				authorized, err := svc.takeDecision(svc.ruleEngine, req, objectMap)
				if !authorized {
					invalidObjects++
					slog.Debug("object %s not authorized: %s", objectMap.ID(), errl.Error(err))
					// diagnosticMessages = append(diagnosticMessages, fmt.Sprintf("object %s not authorized: %s", objectMap.ID(), errl.Error(err)))
					diagnosticObjects = append(diagnosticObjects, repo.ValidationResult{
						ObjectID:   objectMap.ID(),
						ObjectType: req.ResourceName,
						Valid:      false,
						Errors: []repo.ValidationError{
							{
								Field:   objectMap.ID(),
								Message: fmt.Sprintf("object %s not authorized: %s", objectMap.ID(), errl.Error(err)),
								Code:    "NOT_AUTHORIZED",
							},
						},
					})
					continue
				}

				// We have to discard the first 'offset' objects, as specified by the user
				if offsetCounter < userOffset {
					offsetCounter++
					continue
				}

				// Handle attribute selection if required
				if fieldsParam != "" {
					// Allocate a map with initially 10 elements. This is just heuristics, as the required fields are 6 and typically the caller
					// using attribute selection is interested in few elements.
					filteredObjectMap := make(map[string]any, 10)

					// Iterate the fields that we require, and for each key, copy the value from the source object to the target
					for key := range fieldSet {
						if val, ok := objectMap[key]; ok {
							filteredObjectMap[key] = val
						}
					}

					// Replace the objectMap with the new filtered object
					objectMap = filteredObjectMap

				}

				// Now we add the object to the result array and check if we reached the limit as specified by the user
				responseObjectMaps = append(responseObjectMaps, objectMap)

				if userLimit >= 0 && len(responseObjectMaps) >= userLimit {
					break
				}

			}

			// Exit from the outer loop if we already have the number of objects that the user requested
			if userLimit >= 0 && len(responseObjectMaps) >= userLimit {
				break
			}

			// Exit the loop if the number of retrieved objects is less than the page size requested.
			// That means that there are no more objects
			if len(responseData) < pageSize {
				break
			}

			// Retrieve next page
			pageOffset += pageSize

		}

		responseHeaders := make(map[string]string)
		responseHeaders["X-Total-Count"] = strconv.Itoa(len(responseObjectMaps))

		slog.Debug("Remote objects listed", slog.Int("valid", len(responseObjectMaps)), slog.Int("invalid", invalidObjects), slog.String("resourceName", req.ResourceName))
		if diagnostic {
			return &Response{StatusCode: http.StatusOK, Headers: responseHeaders, Body: diagnosticObjects}
		} else {
			return &Response{StatusCode: http.StatusOK, Headers: responseHeaders, Body: responseObjectMaps}
		}

	} else {

		// Set the offset and limit to the user-specified values or the default
		req.QueryParams.Set("offset", strconv.Itoa(userOffset))
		req.QueryParams.Set("limit", strconv.Itoa(userLimit))

		var objs []repo.TMFRecord
		var totalCount int

		objs, totalCount, err = svc.listObjects(req, func(obj *repo.TMFRecord) bool {

			// Convert to a type-safe map representation to facilitate manipulation
			objMap, err := obj.ToTMFObjectMap()
			if err != nil {
				err = errl.Errorf("failed to unmarshal object content for listing: %w", err)
				slog.Error(err.Error())
			}

			// Check if the user can access the object
			authorized, err := svc.takeDecision(svc.ruleEngine, req, objMap)
			if !authorized {
				slog.Debug("object %s not authorized: %w", obj.ID, errl.Error(err))
				return false
			}

			return true

		})

		if err != nil {
			slog.Error("failed to list objects from service", slog.String("error", err.Error()))
			return ErrorResponsef(http.StatusInternalServerError, "failed to list objects from service: %w", err)
		}

		responseHeaders := make(map[string]string)
		responseHeaders["X-Total-Count"] = strconv.Itoa(totalCount)

		// Generate the response to the caller, which must be an array of repo.TMFObjectMap
		responseData := make([]repo.TMFObjectMap, 0, len(objs))
		for _, obj := range objs {
			objectMap, err := obj.ToTMFObjectMap()
			if err != nil {
				slog.Error("failed to unmarshal object content for listing", slog.String("error", err.Error()))
				return ErrorResponsef(http.StatusInternalServerError, "failed to unmarshal object content for listing: %w", err)
			}

			// Handle attribute selection if required
			if fieldsParam != "" {
				// Allocate a map with initially 10 elements. This is just heuristics, as the required fields are 6 and typically the caller
				// using attribute selection is interested in few elements.
				filteredObjectMap := make(map[string]any, 10)

				// Iterate the fields that we require, and for each key, copy the value from the source object to the target
				for key := range fieldSet {
					if val, ok := objectMap[key]; ok {
						filteredObjectMap[key] = val
					}
				}

				// Replace the objectMap with the new filtered object
				objectMap = filteredObjectMap

			}

			// Add the processed object to the array of objects to send back to caller
			responseData = append(responseData, objectMap)

		}

		if !req.HealthRequest {
			slog.Debug("Objects listed successfully", slog.Int("count", len(responseData)), slog.String("resourceName", req.ResourceName))
		}
		return &Response{StatusCode: http.StatusOK, Headers: responseHeaders, Body: responseData}
	}

}
