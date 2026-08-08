package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// PutTMFObject is like CreateTMFObject but allows an id to be specified by the caller.
func (svc *Service) PutTMFObject(ctx context.Context, req *Request) *Response {
	slog.Debug("PutTMFObject called", slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))

	// Authentication is required for create operations
	if errorResponse := svc.requiresAuthentication(req); errorResponse != nil {
		return errorResponse
	}

	// Parse request body
	incomingObjectMap, errorResponse := svc.parseCreateAndPutRequestBody(req)
	if errorResponse != nil {
		return errorResponse
	}

	if svc.LogLevel() >= 3 {
		data, _ := json.MarshalIndent(incomingObjectMap, "", "  ")
		fmt.Println(string(data))
	}

	// Ensure TMF metadata (ID, version, href, etc.)
	if errorResponse := svc.verifyObjectOnCreate(req, incomingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Authorization
	if errorResponse := svc.authorizeAction(req, incomingObjectMap); errorResponse != nil {
		return errorResponse
	}

	// Object Creation (Remote or Local)
	response := svc.createRemoteOrLocalObject(ctx, req, incomingObjectMap)

	// Notification to subscribers
	if response.StatusCode == http.StatusCreated {
		eventType := toEventType(req.ResourceName, "CreateEvent")
		eventPayload := buildEventPayload(req, eventType, response.Body)
		svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)
	}

	return response
}
