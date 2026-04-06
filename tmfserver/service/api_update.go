package service

import (
	"log/slog"
	"net/http"
)

// UpdateGenericObject updates an existing TMF object using generalized parameters.
func (svc *Service) UpdateGenericObject(req *Request) *Response {
	slog.Debug("UpdateGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Authentication is required for update operations
	if errorResponse := svc.requiresAuthentication(req); errorResponse != nil {
		return errorResponse
	}

	// Parse request body
	incomingObjectMap, errorResponse := svc.parseRequestBody(req)
	if errorResponse != nil {
		return errorResponse
	}

	// Ensure update metadata
	if errorResponse := svc.ensureUpdateMetadata(req, incomingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Retrieve existing object
	existingRecord, err := svc.getLocalOrRemoteObject(req)
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "error retrieving existing object %s for update: %w", req.ID, err)
	}
	if existingRecord == nil {
		return ErrorResponsef(http.StatusNotFound, "object %s not found", req.ID)
	}

	// Convert from storage representation to TMF object map
	existingObjectMap, err := existingRecord.ToTMFObjectMap()
	if err != nil {
		return ErrorResponsef(http.StatusInternalServerError, "error unmarshalling existing object record: %w", err)
	}

	// Authorization (using existing object)
	if errorResponse := svc.authorizeAction(req, existingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Object Update (Local or Remote)
	response := svc.updateRemoteOrLocalObject(req, existingRecord, incomingObjectMap)

	// Notification
	if response.StatusCode == http.StatusOK {
		eventType := toEventType(req.ResourceName, "AttributeValueChangeEvent")
		eventPayload := buildEventPayload(req, eventType, response.Body)
		svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)
	}

	return response
}
