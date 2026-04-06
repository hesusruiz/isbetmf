// Copyright 2023 Jesus Ruiz. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// createLocalOrRemoteObject creates an object in the remote server and then in the local database, if the proxy is enabled.
// Othewise, it just creates the object in the local database.
func (svc *Service) createLocalOrRemoteObject(req *Request, obj *repo.TMFRecord) *Response {

	objMap, err := obj.ToTMFObjectMapCreate()
	if err != nil {
		err = errl.Errorf("failed to marshal object: %w", err)
		return ErrorResponsef(http.StatusInternalServerError, "failed to marshal object: %w", err)
	}

	// Create the object only in the local database if the proxy is not enabled
	if !svc.proxyEnabled {
		if err := svc.CreateObject(obj); err != nil {
			if errors.Is(err, &ErrObjectExists{}) {
				return ErrorResponsef(http.StatusBadRequest, "object %s already exists: %w", obj.ID, err)
			} else {
				return ErrorResponsef(http.StatusInternalServerError, "failed to create object locally: %w", err)
			}
		}

		// Set the headers of the reply to the caller, as per TMF specs
		headers := map[string]string{
			"Location": objMap.Href(),
		}
		slog.Info("Object created successfully", slog.String("id", objMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", objMap.Href()))

		return &Response{StatusCode: http.StatusCreated, Headers: headers, Body: objMap}

	}

	// With the proxy enabled, first create the object in the remote server and then locally
	// We do not have to worry about transaction integrity, because if the remote server fails, we do not create the object locally
	// If the local server fails, we do not have to do anything, because the object is already in the remote server and our cache will
	// be updated later on the next call from the user

	remoteObjectMap, errs := svc.tmfClient.TMFPost(req, objMap)
	if len(errs) > 0 {
		return ErrorResponsef(http.StatusInternalServerError, "failed to proxy request: %w", errs[0])
	}

	// Prepare the object for the database
	lastUpdate := remoteObjectMap.LastUpdate()

	// It is an error if the remote server does not return a 'lastUpdate', but we just log a warning
	if lastUpdate == "" {
		slog.Warn("remote server did not return lastUpdate, fixing it", slog.String("id", remoteObjectMap.ID()))
	}

	// Create the new object in the local database
	if err := svc.CreateObject(remoteObjectMap.ToTMFRecord(req.ResourceName)); err != nil {
		// If we get an error, just log the error because we return the object created remotely
		slog.Error("failed to create local object", slog.String("id", remoteObjectMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", remoteObjectMap.Href()))
	}

	// Set the headers of the reply to the caller, as per TMF specs
	headers := map[string]string{
		"Location": remoteObjectMap.Href(),
	}
	slog.Info("Object created successfully", slog.String("id", remoteObjectMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", remoteObjectMap.Href()))

	return &Response{StatusCode: http.StatusCreated, Headers: headers, Body: remoteObjectMap}

}

// updateRemoteOrLocalObject updates an existing object in the remote server and then in the local database,
// if the proxy is enabled.
// existingRecord is only used if the proxy is not enabled.
// Otherwise, it just updates the object in the local database after merging with the RFC7396 patch.
func (svc *Service) updateRemoteOrLocalObject(req *Request, existingRecord *repo.TMFRecord, patch repo.TMFObjectMap) *Response {
	var existingObjectMap repo.TMFObjectMap
	var err error

	existingObjectMap, err = existingRecord.ToTMFObjectMap()
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to unmarshal existing object content: %w", err)
	}

	if svc.proxyEnabled {
		remoteObjectMap, errs := svc.tmfClient.TMFPatch(req, patch)
		if len(errs) > 0 {
			return ErrorResponsef(http.StatusInternalServerError, "failed to proxy request: %w", errs[0])
		}

		// Set the existingObjectMap to the remoteObjectMap, so we store the object as it was received from the remote server
		existingObjectMap = remoteObjectMap

	} else {
		// Merge patch into existing object using RFC7396
		svc.mergeRFC7396(existingObjectMap, patch)
	}

	// Prepare the object for the database
	id := existingObjectMap.ID()
	existingVersion := existingObjectMap.Version()
	existingLastUpdate := existingObjectMap.LastUpdate()

	// It is an error if the remote server does not return a 'lastUpdate', but we fix it and log a warning
	if existingLastUpdate == "" {
		slog.Warn("remote server did not return lastUpdate, fixing it", slog.String("id", id))

		now := time.Now()
		existingLastUpdate = now.Format(time.RFC3339Nano)
		existingObjectMap["lastUpdate"] = existingLastUpdate
	}

	// We will store the serialized object in the database
	existingObjectContent, err := json.Marshal(existingObjectMap)
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to marshal object content for update: %w", err)
	}

	existingObject := &repo.TMFRecord{
		ID:         req.ID,
		Type:       req.ResourceName,
		Version:    existingVersion,
		APIVersion: req.APIVersion,
		LastUpdate: existingLastUpdate,
		Content:    existingObjectContent,
		CreatedAt:  existingRecord.CreatedAt,
		UpdatedAt:  time.Now().Unix(),
	}

	if err := svc.UpdateObject(existingObject); err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to update object in service: %w", err)
	}

	return &Response{StatusCode: http.StatusOK, Body: existingObjectMap}
}

// listRemoteObjects retrieves objects from the remote TMF server, filters them based on authorization,
// caches them locally, and returns the requested page.
func (svc *Service) listRemoteObjects(req *Request, userLimit, userOffset int, fieldSet map[string]bool) (
	[]repo.TMFObjectMap, map[string]string, *Response) {

	// Delete the attribute selection for the query to the backend. We will receive full objects and
	// perform attribute selection ourselves. This is because we want to store the full objects in our local cache.
	req.QueryParams.Del("fields")

	// We forward the same access token that we received from the user
	headers := map[string]string{
		"Authorization": "Bearer " + req.AuthUser.AccessToken,
	}

	// Check if the user wants diagnostic information, which is specified in the query string as '?diagnostic=true'
	// This is not standard TMF, we use it to report on quality of data
	diagnostic := req.QueryParams.Has("diagnostic")
	if diagnostic {
		// We do not want to forward this to the remote server
		req.QueryParams.Del("diagnostic")
	}

	responseObjectMaps := make([]repo.TMFObjectMap, 0)
	diagnosticObjects := make([]repo.ValidationResult, 0)
	var offsetCounter int
	invalidObjects := 0

	// To be able to perform authorization and return the number of objects requested from the user, we must request objects from the server
	// starting from the beginning, i.e. offset=0, and continue until we have enough objects to satisfy the user's request.
	// This is because the objects are not ordered in any particular way, so we cannot just request the objects we need.
	// We request objects in pages of 100 objects at a time.

	// Delete the paging parameters as we will handle them ourselves
	req.QueryParams.Del("offset")
	req.QueryParams.Del("limit")

	// Build the new parameters to send to the remote server
	baseParams := req.QueryParams.Encode()

	// Build the base path including parametersfor the request to the remote server
	basePath := fmt.Sprintf("/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName)
	if baseParams != "" {
		basePath += "?" + baseParams + "&"
	} else {
		basePath += "?"
	}

	// Reduce logging for health requests, to avoid polluting the logs
	isHealthRequest := req.HealthRequest

	pageSize := 100
	pageOffset := 0
	if !isHealthRequest {
		slog.Info("listing remote objects", "path", basePath, "limit", userLimit, "offset", userOffset)
	}
	for {

		// Tell the server which page of objects we want
		path := fmt.Sprintf("%slimit=%d&offset=%d", basePath, pageSize, pageOffset)

		if !isHealthRequest {
			slog.Debug("sending request to remote", "path", path)
		}

		// Get one page of objects from the remote server
		receivedObjects, err := svc.tmfClient.TMFGetList(path, headers)
		if err != nil {
			return nil, nil, ErrorResponsef(http.StatusInternalServerError, "upstream server failed with error: %w", err)
		}
		if !isHealthRequest {
			slog.Debug("received objects from remote", "num_objects", len(receivedObjects))
		}

		// We check each object to see if the user can access it.
		// Additionally, we cache all the objects received independently of the user's access.
		for _, responseObject := range receivedObjects {

			// Get the internal object map from the response object, performing validation
			objectMap, validations := repo.NewTMFObjectMapFromUpstream(req.ResourceName, responseObject)
			if len(validations.Errors) > 0 {
				invalidObjects++
				diagnosticObjects = append(diagnosticObjects, validations)
				continue
			}

			// Convert object to storage representation to save it in the local database
			storageObject := objectMap.ToTMFRecord(req.ResourceName)
			if err := svc.UpsertObject(storageObject); err != nil {
				if !errors.Is(err, &ErrObjectExists{}) {
					invalidObjects++
					slog.Error("error saving object in local database", "error", err)
					continue
				}
			}

			// Check if the user is authorized to access the object
			authorized, err := svc.takeDecision(svc.ruleEngine, req, objectMap)
			if !authorized {
				slog.Debug("object not authorized", "id", objectMap.ID(), "error", err)
				// Add diagnostic info if not authorized
				invalidObjects++
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

			if offsetCounter < userOffset {
				offsetCounter++
				continue
			}

			// Apply attribute selection, according to what the user specified in the query
			objectMap = svc.applyAttributeSelection(objectMap, fieldSet)
			responseObjectMaps = append(responseObjectMaps, objectMap)

			// Stop validating objects if we have enough objects to satisfy the user's request
			if userLimit >= 0 && len(responseObjectMaps) >= userLimit {
				break
			}
		}

		// Stop requesting objects from the remote server if we have enough objects to satisfy the user's request or
		// if we have received all objects from the remote server and there is nothing more to request
		// We use the 'len(receivedObjects) < pageSize' condition to detect if we have received all objects from the remote server.
		// It may be that with this check we do an additional request if the remote server had an exact multiple of pageSize objects,
		// but the robustness of the code is more important than the performance.
		if (userLimit >= 0 && len(responseObjectMaps) >= userLimit) || len(receivedObjects) < pageSize {
			break
		}
		pageOffset += pageSize
	}

	responseHeaders := map[string]string{
		"X-Total-Count": strconv.Itoa(len(responseObjectMaps)),
	}

	slog.Debug("Remote objects listed", slog.Int("valid", len(responseObjectMaps)), slog.Int("invalid", invalidObjects), slog.String("resourceName", req.ResourceName))

	// If the user wants diagnostic information, return it
	if diagnostic {
		return nil, responseHeaders, &Response{StatusCode: http.StatusOK, Headers: responseHeaders, Body: diagnosticObjects}
	}
	return responseObjectMaps, responseHeaders, nil
}

// listLocalObjects retrieves TMF objects from the local database, filters them based on authorization,
// handles attribute selection, and returns the requested page.
func (svc *Service) listLocalObjects(req *Request, userLimit, userOffset int, fieldSet map[string]bool) ([]repo.TMFObjectMap, map[string]string, *Response) {
	// Set the offset and limit for the database query
	req.QueryParams.Set("offset", strconv.Itoa(userOffset))
	req.QueryParams.Set("limit", strconv.Itoa(userLimit))

	storageObjects, err := svc.ListObjects(req, func(storageObject *repo.TMFRecord) bool {
		// Convert to internal object representation
		objMap, err := storageObject.ToTMFObjectMap()
		if err != nil {
			slog.Error("failed to unmarshal object content for listing", "error", err)
			return false
		}

		// Check if the user is authorized to access the object
		authorized, err := svc.takeDecision(svc.ruleEngine, req, objMap)
		if !authorized {
			slog.Debug("object not authorized", "id", storageObject.ID, "error", err)
			return false
		}
		return true
	})

	if err != nil {
		return nil, nil, ErrorResponsef(http.StatusInternalServerError, "failed to list objects from local database: %w", err)
	}

	responseObjects := make([]repo.TMFObjectMap, 0, len(storageObjects))
	for _, storageObject := range storageObjects {
		// Convert to internal object representation
		objectMap, err := storageObject.ToTMFObjectMap()
		if err != nil {
			return nil, nil, ErrorResponsef(http.StatusInternalServerError, "failed to unmarshal object content for listing: %w", err)
		}

		// Apply attribute selection
		objectMap = svc.applyAttributeSelection(objectMap, fieldSet)
		responseObjects = append(responseObjects, objectMap)
	}

	responseHeaders := map[string]string{
		"X-Total-Count": strconv.Itoa(len(responseObjects)),
	}

	if !req.HealthRequest {
		slog.Debug("Objects listed successfully", slog.Int("count", len(responseObjects)), slog.String("resourceName", req.ResourceName))
	}

	return responseObjects, responseHeaders, nil
}

// getLocalOrRemoteObject retrieves the object from the database.
// If the proxy is enabled and the object is not found locally or is stale, we try to get it from the remote server.
// If the object is not found anywhere, it returns a nil object and no error.
// There is no way to force the retrieval from the remote server if the object exists locally and is fresh enough.
func (svc *Service) getLocalOrRemoteObject(req *Request) (*repo.TMFRecord, error) {

	objectID := req.ID

	// For organization resources in local database, the objectID is the organization identifier.
	// TODO: this logic is only for ISBE
	if req.ResourceName == "organization" {
		if strings.HasPrefix(req.ID, "urn:ngsi-ld:organization:") && !strings.HasPrefix(req.ID, "urn:ngsi-ld:organization:did:elsi:") {
			objectID = "urn:ngsi-ld:organization:did:elsi:" + strings.TrimPrefix(req.ID, "urn:ngsi-ld:organization:")
		}
	}

	// Check if we have the object locally
	obj, err := svc.GetObject(objectID, req.ResourceName)
	if err != nil {
		return nil, errl.Errorf("failed to get object %s from local service: %w", objectID, err)
	}

	// If proxy is not enabled, we return whatever we found, which is nil if object is not found.
	// Not finding an object is not an error at this level but the caller is responsible for returning a 404 error to the user.
	if !svc.proxyEnabled {
		return obj, nil
	}

	// If we found the object and the database record is not stale, return it
	if obj != nil {
		if time.Since(obj.GetUpdatedAt()) < svc.fressness {
			slog.Debug("Object found in local database and fresh")
			return obj, nil
		}
		slog.Debug("Object found in local database but stale, retrieving from remote")
	} else {
		slog.Debug("Object not found in local database, retrieving from remote")
	}

	// The object was not found or is stale, so we have to retrieve remotely and update the local database
	remoteObj, err := svc.getRemoteObject(req)
	if err != nil {
		return nil, errl.Errorf("failed to get object %s from remote service: %w", req.ID, err)
	}

	// Store the object locally and return it to caller
	if err := svc.CreateObject(remoteObj); err != nil {
		slog.Error("failed to cache object", slog.Any("error", err))
		// Return the stale object or nil
		return remoteObj, nil
	}

	slog.Info("Object retrieved from remote and cached successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return remoteObj, nil

}

func (svc *Service) getRemoteObject(req *Request) (*repo.TMFRecord, error) {
	slog.Debug("retrieving object from remote", slog.String("id", req.ID))

	// Send the access token
	headers := map[string]string{
		"Authorization": "Bearer " + req.AuthUser.AccessToken,
	}

	// Build the path for the request according to TMForum specs
	path := fmt.Sprintf("/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID)

	// Send the request to the remote with our HTTP Client
	resp, err := svc.tmfClient.Get(path, headers)
	if err != nil {
		return nil, errl.Errorf("failed to proxy request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errl.Errorf("failed to read response body: %w", err)
	}

	// Not found is not an error at this level, but the caller must check for a nil object
	if resp.StatusCode == 404 {
		return nil, nil
	}

	// TODO: handle caching replies, with status codes in the 3xx range
	if resp.StatusCode >= 300 {
		return nil, errl.Errorf("remote server returned error: %s", string(body))
	}

	// Build the object from the reply
	receivedObjectMap, err := repo.NewTMFObjectMapFromBytes(req.ResourceName, body)
	if err != nil {
		return nil, errl.Errorf("failed to bind request body: %w", err)
	}

	// Prepare the object for the database
	id := receivedObjectMap.ID()
	version := receivedObjectMap.Version()
	lastUpdate := receivedObjectMap.LastUpdate()

	// It is an error if the remote server does not return a 'lastUpdate', but we fix it and log a warning
	if lastUpdate == "" {
		slog.Warn("remote server did not return lastUpdate, fixing it", slog.String("id", id))

		now := time.Now()
		lastUpdate = now.Format(time.RFC3339Nano)
		receivedObjectMap["lastUpdate"] = lastUpdate
	}

	// Instead of storing what we receive, we store a compact serialization of the JSON object with possible updated fields
	receivedContent, err := json.Marshal(receivedObjectMap)
	if err != nil {
		return nil, errl.Errorf("failed to marshal object content: %w", err)
	}

	obj := repo.NewTMFRecord(id, req.ResourceName, version, req.APIVersion, lastUpdate, receivedContent)

	return obj, nil
}
