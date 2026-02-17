package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// UpdateGenericObject updates an existing TMF object using generalized parameters.
//
// The function performs the following checks and operations:
//
// 1.  **Authentication**:
//   - It requires an authenticated user, obtained by processing the Access Token.
//   - If the user is not authenticated, it returns a 401 Unauthorized error.
//
// 2.  **Incoming TMF Object Checks**:
//   - It parses the request body to get the incoming TMF object patch.
//   - It ensures that if an `id` is present in the body, it matches the `id` in the URL.
//   - It sets the `seller` and `buyer` information, overwriting any existing values.
//   - It sets the `lastUpdate` field to the current time.
//
// 3.  **Existing TMF Object Checks**:
//   - It retrieves the existing object from the local database or the remote server (if in proxy mode).
//   - If the object is not found, it returns a 404 Not Found error.
//   - It checks that the object is managed by the current server operator by verifying the `sellerOperator` DID.
//   - It checks that the authenticated user is the owner of the object by verifying the `seller` DID.
//   - It checks the `lifecycleStatus` of the existing object. If it is "Launched", "Retired", or "Obsolete", it requires the incoming object to have a higher `version` number.
//
// 4.  **Authorization**:
//   - The ownership and operator checks act as an authorization mechanism.
//
// 5.  **Object Update**:
//   - If proxy mode is enabled, it forwards the PATCH request to the remote TMF server.
//   - The response from the remote server is then stored in the local database.
//   - If proxy mode is disabled, it merges the incoming patch with the existing object using JSON Merge Patch (RFC 7396).
//   - The updated object is then saved to the local database.
//
// 6.  **Response and Notification**:
//   - It returns a 200 OK response with the updated object in the body.
//   - It sends an "AttributeValueChangeEvent" notification to subscribed listeners.
func (svc *Service) UpdateGenericObject(req *Request) *Response {
	var err error
	slog.Debug("UpdateGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// ************************************************************************************************
	// Authentication: we require the user to be authenticated
	// ************************************************************************************************
	if len(req.TokenMap) == 0 {
		return ErrorResponsef(http.StatusUnauthorized, "user not authenticated")
	}

	var incomingObjMap repo.TMFObjectMap

	// ************************************************************************************************
	// Parse the request body, which contains the TMForum object being updated.
	// We perform some formal verifications and add default values if needed.
	// ************************************************************************************************
	{

		incomingObjMap, err = repo.NewTMFObjectMapFromBytes(req.ResourceName, req.Body)
		if err != nil {
			return ErrorResponsef(http.StatusBadRequest, "failed to bind request body: %w", errl.Error(err))
		}

		// Check if the caller is trying to set the lifecycleStatus of a ProductOffering to "Launched"
		if strings.EqualFold(incomingObjMap.Type(), config.ProductOffering) && strings.EqualFold(incomingObjMap.LifecycleStatus(), "Launched") {
			// If the feature is enabled, only the admin (server operator) can set the lifecycleStatus to "Launched"
			if svc.Features.OfferingLaunchOnlyByAdmin {
				caller := req.AuthUser
				if !repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {
					return ErrorResponsef(http.StatusForbidden, "offering launch is only allowed by admin")
				}
			}
		}

		// An update operation specifies the ID of the object to update in the URL.
		// To facilitate life to applications, we allow the 'id' to be also in the body.
		// But if the 'id' is present in the body, ensure it matches the 'id' in the URL
		if !DOMEHacks {
			id, _ := incomingObjMap["id"].(string)
			if id != "" && id != req.ID {
				err := errl.Errorf("ID in body must match ID in URL")
				return ErrorResponsef(http.StatusBadRequest,
					"invalid object, request id: %s, id in body: %s, error: %w",
					req.ID,
					id,
					err,
				)
			}
		}

		// TODO: check if we could set some information for the Individual and Organization objects

		// Set the lastUpdate property. We overwrite whatever the user set.
		now := time.Now()
		lastUpdate := now.Format(time.RFC3339Nano)
		incomingObjMap["lastUpdate"] = lastUpdate

	}

	// ************************************************************************************************
	// Retrieve existing object locally or from proxy
	// ************************************************************************************************

	var existingObj *repo.TMFRecord

	// Retrieve existing object, locally or remotely
	existingObj, err = svc.getLocalOrRemoteObject(req)
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to get existing object %s for update: %w", req.ID, err)
	}

	if existingObj == nil {
		return ErrorResponsef(http.StatusNotFound, "object %s not found", req.ID)
	}

	existingObjectMap, err := existingObj.ToTMFObjectMap()
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to unmarshal existing object content: %w", err)
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
		slog.Debug("caller is authorized to update object", "reason", err)
	}

	// ************************************************************************************************
	// Now we can proceed.
	// ************************************************************************************************

	if svc.proxyEnabled {

		// Proxy logic
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + req.AccessToken,
		}
		path := fmt.Sprintf("/tmf-api/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID)

		// Perform a PATCH
		resp, err := svc.tmfClient.Patch(path, req.Body, headers)
		if err != nil {
			return ErrorResponsef(http.StatusInternalServerError, "failed to proxy request: %w", err)
		}
		defer resp.Body.Close()

		// Parse response, which is the updated object contents
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return ErrorResponsef(http.StatusInternalServerError, "failed to read response body: %w", err)
		}

		if resp.StatusCode >= 300 {
			return ErrorResponsef(resp.StatusCode, "remote server returned error: %s", string(body))
		}

		var receivedObjectMap repo.TMFObjectMap
		if err := json.Unmarshal(body, &receivedObjectMap); err != nil {
			return ErrorResponsef(http.StatusBadRequest, "failed to bind request body: %w", err)
		}

		// Set the existingObjectMap to the remoteObjectMap, so we store the object as it was received from the remote server
		existingObjectMap = receivedObjectMap

	} else {

		// Merge incomingObjMap into existing object using RFC7396 (JSON Merge Patch)

		// RFC7396 merge implementation: modify target in place
		mergeRFC7396(existingObjectMap, incomingObjMap)

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
		CreatedAt:  existingObj.CreatedAt,
		UpdatedAt:  time.Now().Unix(),
	}

	if err := svc.updateObject(existingObject); err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to update object in service: %w", err)
	}

	slog.Info("Object updated successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Send TMForum notification (AttributeValueChangeEvent)
	eventType := toEventType(req.ResourceName, "AttributeValueChangeEvent")
	eventPayload := buildEventPayload(req, eventType, existingObjectMap)
	svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)

	return &Response{StatusCode: http.StatusOK, Body: existingObjectMap}
}

// mergeRFC7396 implements JSON Merge Patch (RFC 7396) to merge a patch object into a target object.
// The function modifies the target object in place according to RFC 7396 rules:
// - If the patch value is null, the member is removed from the target
// - If both values are objects, they are merged recursively
// - For arrays or scalar values, the target value is replaced
func mergeRFC7396(target, patch map[string]any) {
	for k, v := range patch {
		// If the patch value is nil -> remove the member from the target
		if v == nil {
			delete(target, k)
			continue
		}

		// If both are objects, merge recursively
		vMap, vIsMap := v.(map[string]any)
		if vIsMap {
			if existingChild, ok := target[k]; ok {
				if existingChildMap, ok2 := existingChild.(map[string]any); ok2 {
					mergeRFC7396(existingChildMap, vMap)
					target[k] = existingChildMap
					continue
				}
			}
			// Otherwise, replace with the incoming object
			target[k] = vMap
			continue
		}

		// For arrays or scalar values, replace
		target[k] = v
	}
}
