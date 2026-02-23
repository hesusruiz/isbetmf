package service

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// DeleteGenericObject deletes a TMF object using generalized parameters.
//
// The function performs the following checks and operations:
//
// 1.  **Authentication**:
//   - It requires an authenticated user, obtained by processing the Access Token.
//   - If the user is not authenticated, it returns a 401 Unauthorized error.
//
// 2.  **Existing TMF Object Checks**:
//   - It retrieves the existing object from the local database or the remote server (if in proxy mode).
//   - If the object is not found, it returns a 404 Not Found error.
//   - It checks that the object is managed by the current server operator by verifying the `sellerOperator` DID.
//   - It checks that the authenticated user is the owner of the object by verifying the `seller` DID.
//
// 3.  **Authorization**:
//   - The ownership and operator checks act as an authorization mechanism.
//
// 4.  **Object Deletion**:
//   - If proxy mode is enabled, it forwards the DELETE request to the remote TMF server.
//   - It deletes the object from the local database.
//
// 5.  **Response and Notification**:
//   - It returns a 204 No Content response on successful deletion.
//   - It sends a "DeleteEvent" notification to subscribed listeners.
func (svc *Service) DeleteGenericObject(req *Request) *Response {
	var err error
	slog.Debug("DeleteGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// ************************************************************************************************
	// Authentication: we require the user to be authenticated
	// ************************************************************************************************
	if len(req.TokenMap) == 0 {
		return ErrorResponsef(http.StatusUnauthorized, "user not authenticated")
	}

	// ************************************************************************************************
	// Retrieve existing object from database
	// ************************************************************************************************

	var existingObject *repo.TMFRecord

	// Retrieve existing object, locally or remotely
	existingObject, err = svc.getLocalOrRemoteObject(req)
	if err != nil {
		// TODO: check the return code from remote server and reply accordingly
		return ErrorResponsef(http.StatusBadRequest, "failed to get existing object for update: %w", err)
	}

	// If nothing to delete, return 404
	if existingObject == nil {
		return ErrorResponsef(http.StatusNotFound, "object %s not found", req.ID)
	}

	// Convert to a type-safe map representation to facilitate manipulation
	existingObjectMap, err := existingObject.ToTMFObjectMap()
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
		slog.Debug("caller is authorized to delete object", "reason", err)
	}

	// ##########################################################
	// Delete the object
	// ##########################################################

	// Delete the object in the remote server, if the proxy is enabled
	if svc.proxyEnabled {
		// Send the authentication header
		headers := map[string]string{
			"Authorization": "Bearer " + req.AccessToken,
		}
		path := fmt.Sprintf("/tmf-api/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID)
		resp, err := svc.tmfClient.Delete(path, headers)
		if err != nil {
			return ErrorResponsef(http.StatusInternalServerError, "failed to proxy request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return ErrorResponsef(http.StatusInternalServerError, "failed to read response body: %w", err)
			}
			return &Response{
				StatusCode: resp.StatusCode,
				Body:       body,
			}
		}

		slog.Info("Object deleted from remote server successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	}

	// Delete the object in the local database
	if err := svc.deleteObject(req.ID, req.ResourceName); err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "failed to delete object %s from service: %w", req.ID, err)
	}

	slog.Info("Object deleted successfully from local database", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Send TMForum notification
	eventType := toEventType(req.ResourceName, "DeleteEvent")
	minimal := map[string]any{
		"id":    req.ID,
		"@type": req.ResourceName,
		"href":  fmt.Sprintf("/tmf-api/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID),
	}
	eventPayload := buildEventPayload(req, eventType, minimal)
	svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)

	return &Response{StatusCode: http.StatusNoContent}

}
