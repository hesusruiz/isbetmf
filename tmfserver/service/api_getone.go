package service

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/hesusruiz/isbetmf/internal/errl"
)

// GetGenericObject retrieves a TMF object using generalized parameters.
//
// The function performs the following checks and operations:
//
// 1.  **Authentication**:
//   - It processes the Access Token to get the caller's information. Public (unauthenticated) access is allowed.
//
// 2.  **Object Retrieval**:
//   - It first tries to retrieve the object from the local database.
//   - If the object is not found and proxy mode is enabled, it tries to fetch it from the remote TMF server.
//   - If the object is found remotely, it is cached in the local database.
//   - If the object is not found locally or remotely, it returns a 404 Not Found error.
//
// 3.  **Authorization**:
//   - It calls the Policy Decision Point (PDP) to check if the user is authorized to read the object.
//   - If the user is not authorized, it returns a 403 Forbidden error.
//
// 4.  **Response**:
//   - It handles partial field selection using the "fields" query parameter.
//   - It returns a 200 OK response with the (potentially filtered) object in the body.
func (svc *Service) GetGenericObject(req *Request) *Response {
	slog.Debug("GetGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Authentication: anonymous access is allowed, so we do not check the existence of the access token

	// Retrieve the object from the database. If it is not found, we try to get it from the remote server (if proxy is enabled).
	// If the object is not found, we return a 404 error.
	existingObject, err := svc.getLocalOrRemoteObject(req)
	if err != nil {
		// This is an unexpected error, so we return a server error
		return ErrorResponsef(http.StatusInternalServerError, "failed to get object from service: %w", errl.Error(err))
	}

	if existingObject == nil {
		return ErrorResponsef(http.StatusNotFound, "object not found")
	}

	// Convert to a type-safe map representation to facilitate manipulation
	existingObjectMap, err := existingObject.ToTMFObjectMap()
	if err != nil {
		if len(existingObject.Validations.Errors) > 0 {
			return ErrorResponsef(http.StatusInternalServerError, "validation errors found: %s", existingObject.Validations.String())
		} else {
			return ErrorResponsef(http.StatusInternalServerError, "failed to unmarshal existing object content: %w", errl.Error(err))
		}
	}

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object,
	// based on the rules defined by the user in the policy engine.
	// ************************************************************************************************

	if authorized, err := svc.takeDecision(svc.ruleEngine, req, existingObjectMap); !authorized {
		return ErrorResponsef(http.StatusForbidden,
			"user %s is not authorized, object: %s, error: %w",
			req.AuthUser.OrganizationIdentifier,
			existingObjectMap,
			err,
		)
	} else {
		slog.Debug("caller is authorized to read object", "reason", err)
	}

	// ************************************************************************************************
	// Now we can proceed.
	// ************************************************************************************************

	// Handle partial field selection
	fieldsParam := req.QueryParams.Get("fields")
	if fieldsParam != "" {
		var fields []string
		if fieldsParam != "none" {
			fields = strings.Split(fieldsParam, ",")
		}

		// Create a set of fields for quick lookup
		fieldSet := make(map[string]bool)
		for _, f := range fields {
			fieldSet[strings.TrimSpace(f)] = true
		}

		// Always include id, href, lastUpdate, version, @type and lifecycleStatus
		fieldSet["id"] = true
		fieldSet["href"] = true
		fieldSet["lastUpdate"] = true
		fieldSet["version"] = true
		fieldSet["@type"] = true
		fieldSet["lifecycleStatus"] = true

		filteredItem := make(map[string]any)
		for key, value := range existingObjectMap {
			if fieldSet[key] {
				filteredItem[key] = value
			}
		}
		existingObjectMap = filteredItem
	}

	slog.Info("Object retrieved successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return &Response{StatusCode: http.StatusOK, Body: existingObjectMap}
}
