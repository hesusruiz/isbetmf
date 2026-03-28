package service

import (
	"log/slog"
	"net/http"
)

// CreateGenericObject creates a new TMF object using generalized parameters.
func (svc *Service) CreateGenericObject(req *Request) *Response {
	slog.Debug("CreateGenericObject called", slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))

	// Authentication is required for create operations
	if errorResponse := svc.requiresAuthentication(req); errorResponse != nil {
		return errorResponse
	}

	// Parse request body
	incomingObjectMap, errorResponse := svc.parseRequestBody(req)
	if errorResponse != nil {
		return errorResponse
	}

	// Ensure TMF metadata (ID, version, href, etc.)
	if errorResponse := svc.ensureCreateMetadata(req, incomingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Authorization
	if errorResponse := svc.authorizeAction(req, incomingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Object Creation (Local or Remote). First convert incoming object to a storage representation
	repoObject := incomingObjectMap.ToTMFRecord(req.ResourceName)
	response := svc.createLocalOrRemoteObject(req, repoObject)

	// Notification to subscribers
	if response.StatusCode == http.StatusCreated {
		eventType := toEventType(req.ResourceName, "CreateEvent")
		eventPayload := buildEventPayload(req, eventType, response.Body)
		svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)
	}

	return response
}
