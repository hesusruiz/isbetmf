package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	pdp "github.com/hesusruiz/isbetmf/pdp"
	"github.com/hesusruiz/isbetmf/tmfserver/notifications"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/jmoiron/sqlx"
)

// AuthUser represents the authenticated user's information from the JWT mandator object.
type AuthUser struct {
	CommonName             string `json:"commonName"`
	Country                string `json:"country"`
	EmailAddress           string `json:"emailAddress"`
	Organization           string `json:"organization"`
	OrganizationIdentifier string `json:"organizationIdentifier"`
	SerialNumber           string `json:"serialNumber"`
	isAuthenticated        bool
	isLEAR                 bool
	isOwner                bool
}

func (u *AuthUser) ToMap() map[string]any {
	return map[string]any{
		"commonName":             u.CommonName,
		"country":                u.Country,
		"emailAddress":           u.EmailAddress,
		"organization":           u.Organization,
		"organizationIdentifier": u.OrganizationIdentifier,
		"serialNumber":           u.SerialNumber,
		"isAuthenticated":        u.isAuthenticated,
		"isLEAR":                 u.isLEAR,
		"isOwner":                u.isOwner,
	}
}

// Request represents a generic HTTP request. Handlers must convert to this representation.
// In this way, we support easily any HTTP framework (currently Fiber and Echo), but also other
// future channels like JSON-RPC or even non-HTTP channels like GRPC.
type Request struct {
	Method       string
	Action       string
	APIfamily    string
	ResourceName string
	ID           string
	QueryParams  url.Values
	Body         []byte
	AuthUser     *AuthUser
	AccessToken  string
}

func (r *Request) ToMap() map[string]any {
	return map[string]any{
		"method":   r.Method,
		"action":   r.Action,
		"api":      r.APIfamily,
		"resource": r.ResourceName,
		"id":       r.ID,
	}
}

// These are more friendly names for the writers of policy rules and can be used interchangeably
var HttpMethodAliases = map[string]string{
	"GET":    "READ",
	"POST":   "CREATE",
	"PATCH":  "UPDATE",
	"DELETE": "DELETE",
}

// Response represents a generic HTTP response.
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       any
}

// Service is the service for the API.
type Service struct {
	db         *sqlx.DB
	ruleEngine *pdp.PDP
	// The public key used to verify the Access Tokens. In DOME they belong to the Verifier,
	// and the PDP retrieves it dynamically depending on the environment.
	// The caller is able to provide a function to retrieve the key from a different place.
	verifierServer     string
	verifierJWK        *jose.JSONWebKey
	verificationKeyFun func(verifierServer string) (*jose.JSONWebKey, error)

	// The OpenID configuration of the Verifier Server
	oid *OpenIDConfig

	// Notifications manager
	notif *notifications.Manager

	// Pluggable storage backend (optional). When nil, falls back to built-in SQLite via db
	storage Storage
}

// NewService creates a new service.
func NewService(db *sqlx.DB, ruleEngine *pdp.PDP, verifierServer string) *Service {
	svc := &Service{
		db:             db,
		ruleEngine:     ruleEngine,
		verifierServer: verifierServer,
	}

	err := svc.initializeService()
	if err != nil {
		panic(err)
	}

	// Initialize notifications with in-memory store and HTTP delivery
	store := notifications.NewMemoryStore()
	deliver := notifications.NewHTTPDelivery(5 * time.Second)
	svc.notif = notifications.NewManager(store, deliver)

	return svc
}

func (svc *Service) initializeService() error {

	// Create the server operator identity, in case it is not yet in the database
	org := &repository.Organization{
		CommonName:             config.ServerOperatorName,
		Country:                config.ServerOperatorCountry,
		EmailAddress:           "",
		Organization:           config.ServerOperatorName,
		OrganizationIdentifier: config.ServerOperatorOrganizationIdentifier,
	}

	obj, _ := repository.TMFOrganizationFromToken(nil, org)

	if err := svc.createObject(obj); err != nil {
		if errors.Is(err, &ErrObjectExists{}) {
			slog.Debug("server operator organization already exists", "organizationIdentifier", config.ServerOperatorOrganizationIdentifier)
		} else {
			err = errl.Error(err)
			panic("error creatingserver operator organization")
		}
	}

	// Retrieve the OpenId configuration of the Verifier server
	oid, err := NewOpenIDConfig(svc.verifierServer)
	if err != nil {
		return errl.Errorf("failed to retrieve OpenID configuration: %w", err)
	}
	svc.oid = oid

	return nil

}

// CreateHubSubscription creates a new notification subscription (hub) for an API family.
func (svc *Service) CreateHubSubscription(req *Request) *Response {
	// Authenticate like write operations
	_, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Parse incoming body
	var body map[string]any
	if err := json.Unmarshal(req.Body, &body); err != nil {
		err = errl.Errorf("failed to bind request body: %w", err)
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	callback, _ := body["callback"].(string)
	if callback == "" {
		err = errl.Errorf("callback is required")
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	var eventTypes []string
	if raw, ok := body["eventTypes"].([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				eventTypes = append(eventTypes, s)
			}
		}
	}

	headers := make(map[string]string)
	if hmap, ok := body["headers"].(map[string]any); ok {
		for k, v := range hmap {
			if sv, ok := v.(string); ok {
				headers[strings.ToLower(k)] = sv
			}
		}
	}

	query, _ := body["query"].(string)

	// Build subscription
	id := uuid.NewString()
	sub := &notifications.Subscription{
		ID:         id,
		APIFamily:  req.APIfamily,
		Callback:   callback,
		EventTypes: eventTypes,
		Headers:    headers,
		Query:      query,
	}

	_, err = svc.notif.CreateSubscription(req.APIfamily, sub)
	if err != nil {
		err = errl.Errorf("failed to create subscription: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	resp := map[string]any{
		"id":         sub.ID,
		"callback":   sub.Callback,
		"eventTypes": sub.EventTypes,
		"query":      sub.Query,
		"href":       fmt.Sprintf("/tmf-api/%s/v5/hub/%s", req.APIfamily, sub.ID),
	}
	if token, ok := sub.Headers["x-auth-token"]; ok && token != "" {
		resp["headers"] = map[string]any{"x-auth-token": token}
	}

	return &Response{StatusCode: http.StatusCreated, Body: resp}
}

// DeleteHubSubscription deletes a subscription by id for an API family.
func (svc *Service) DeleteHubSubscription(req *Request) *Response {
	// Authenticate like write operations
	_, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	if req.ID == "" {
		err = errl.Errorf("id is required")
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	if err := svc.notif.DeleteSubscription(req.APIfamily, req.ID); err != nil {
		err = errl.Errorf("failed to delete subscription: %w", err)
		apiErr := NewApiError("404", "Not Found", err.Error(), fmt.Sprintf("%d", http.StatusNotFound), "")
		return &Response{StatusCode: http.StatusNotFound, Body: apiErr}
	}

	return &Response{StatusCode: http.StatusNoContent}
}

// CreateGenericObject creates a new TMF object using generalized parameters.
func (svc *Service) CreateGenericObject(req *Request) *Response {
	slog.Debug("CreateGenericObject called", slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// This operation can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Parse the request body, which contains the TMForum object being created
	var incomingObjectMap map[string]any
	if err := json.Unmarshal(req.Body, &incomingObjectMap); err != nil {
		err = errl.Errorf("failed to bind request body: %w", err)
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Failed to bind request body", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	// Create a new 'id' if the user did not specify it
	id, _ := incomingObjectMap["id"].(string)
	if id == "" {
		// If the incoming object does not have an 'id', we generate a new one
		// The format is "urn:ngsi-ld:{resource-in-kebab-case}:{uuid}"
		id = fmt.Sprintf("urn:ngsi-ld:%s:%s", ToKebabCase(req.ResourceName), uuid.NewString())
		incomingObjectMap["id"] = id
		slog.Debug("Generated new ID for object", "id", id)

	}

	// Create a new 'href' if the user did not specify it
	href, _ := incomingObjectMap["href"].(string)
	if href == "" {
		// Add href to the object
		incomingObjectMap["href"] = fmt.Sprintf("/tmf-api/%s/v5/%s/%s", req.APIfamily, req.ResourceName, id)
		slog.Debug("Set href", slog.String("href", incomingObjectMap["href"].(string)))
	}

	// Check and process '@type' field
	if typeVal, typeOk := incomingObjectMap["@type"].(string); typeOk {
		if !strings.EqualFold(typeVal, req.ResourceName) {
			err = errl.Errorf("@type mismatch: expected %s, got %s", req.ResourceName, typeVal)
			apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("@type mismatch", slog.String("expected", req.ResourceName), slog.String("got", typeVal), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	} else {
		// If @type is not specified, add it
		incomingObjectMap["@type"] = req.ResourceName
		slog.Debug("Added missing @type field", slog.String("type", req.ResourceName))
	}

	// Set default 'version' if not provided by the user
	version, versionOk := incomingObjectMap["version"].(string)
	if !versionOk || version == "" {
		version = "1.0"
		incomingObjectMap["version"] = version // Update data map for content marshaling
		slog.Debug("Set default version", slog.String("version", version))
	}

	// Set the lastUpdate property. We overwrite whatever the user set.
	now := time.Now()
	lastUpdate := now.Format(time.RFC3339Nano)
	incomingObjectMap["lastUpdate"] = lastUpdate

	// Add Seller and Buyer info. We overwrite whatever is in the incoming object, if any
	err = setSellerAndBuyerInfo(incomingObjectMap, req.AuthUser.OrganizationIdentifier)
	if err != nil {
		err = errl.Errorf("failed to add Seller and Buyer info: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to add Seller and Buyer info", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	incomingContent, err := json.Marshal(incomingObjectMap)
	if err != nil {
		err = errl.Errorf("failed to marshal object content: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to marshal object content", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	// Create the object in-memory, ready to be checked and stored.
	obj := repo.NewTMFObject(id, req.ResourceName, version, lastUpdate, incomingContent)

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object.
	// ************************************************************************************************

	err = takeDecision(svc.ruleEngine, req, token, obj)
	if err != nil {
		err = errl.Errorf("user not authorized: %w", err)
		apiErr := NewApiError("403", "Forbidden", err.Error(), fmt.Sprintf("%d", http.StatusForbidden), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusForbidden, Body: apiErr}
	}

	// ************************************************************************************************
	// Now we can proceed, creating an object in the database.
	// ************************************************************************************************

	if err := svc.createObject(obj); err != nil {
		err = errl.Errorf("failed to create object in service: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to create object in service", slog.Any("error", err), slog.String("id", id), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	headers := make(map[string]string)
	headers["Location"] = incomingObjectMap["href"].(string)
	slog.Info("Object created successfully", slog.String("id", id), slog.String("resourceName", req.ResourceName), slog.String("location", incomingObjectMap["href"].(string)))

	// Send TMForum notification
	eventType := toEventType(req.ResourceName, "CreateEvent")
	eventPayload := buildEventPayload(req, eventType, incomingObjectMap)
	svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)

	return &Response{
		StatusCode: http.StatusCreated,
		Headers:    headers,
		Body:       incomingObjectMap,
	}
}

// GetGenericObject retrieves a TMF object using generalized parameters.
func (svc *Service) GetGenericObject(req *Request) *Response {
	slog.Debug("GetGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	obj, err := svc.getObject(req.ID, req.ResourceName)
	if err != nil {
		err = errl.Errorf("failed to get object from service: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to get object from service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	if obj == nil {
		err = errl.Errorf("object not found")
		apiErr := NewApiError("404", "Not Found", err.Error(), fmt.Sprintf("%d", http.StatusNotFound), "")
		slog.Info("Object not found", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusNotFound, Body: apiErr}
	}

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object.
	// ************************************************************************************************

	err = takeDecision(svc.ruleEngine, req, token, obj)
	if err != nil {
		err = errl.Error(err)
		apiErr := NewApiError("403", "Forbidden", err.Error(), fmt.Sprintf("%d", http.StatusForbidden), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusForbidden, Body: apiErr}
	}

	// ************************************************************************************************
	// Now we can proceed.
	// ************************************************************************************************

	var responseData map[string]any
	err = json.Unmarshal(obj.Content, &responseData)
	if err != nil {
		err = errl.Errorf("failed to unmarshal object content: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to unmarshal object content", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	// Handle partial field selection
	fieldsParam := req.QueryParams.Get("fields")
	if fieldsParam != "" {
		var fields []string
		if fieldsParam == "none" {
			fields = []string{"id", "href", "lastUpdate", "version"}
		} else {
			fields = strings.Split(fieldsParam, ",")
		}

		// Create a set of fields for quick lookup
		fieldSet := make(map[string]bool)
		for _, f := range fields {
			fieldSet[strings.TrimSpace(f)] = true
		}

		// Always include id, href, lastUpdate, version and @type
		fieldSet["id"] = true
		fieldSet["href"] = true
		fieldSet["lastUpdate"] = true
		fieldSet["version"] = true
		fieldSet["@type"] = true

		filteredItem := make(map[string]any)
		for key, value := range responseData {
			if fieldSet[key] {
				filteredItem[key] = value
			}
		}
		responseData = filteredItem
	}

	slog.Info("Object retrieved successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return &Response{StatusCode: http.StatusOK, Body: responseData}
}

// UpdateGenericObject updates an existing TMF object using generalized parameters.
func (svc *Service) UpdateGenericObject(req *Request) *Response {
	slog.Debug("UpdateGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// This operation can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Parse the request body, which contains the TMForum object being created
	var incomingObjMap map[string]any
	if err := json.Unmarshal(req.Body, &incomingObjMap); err != nil {
		err = errl.Errorf("failed to bind request body: %w", err)
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Failed to bind request body for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	// If the ID is present in the body, ensure it matches the ID in the URL
	if bodyID, ok := incomingObjMap["id"]; ok {
		bodyIDStr, ok := bodyID.(string)
		if !ok || bodyIDStr != req.ID {
			err = errl.Errorf("ID in body must match ID in URL")
			apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("ID mismatch in update request", slog.String("url_id", req.ID), slog.Any("body_id", bodyID), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	}

	// Check and process '@type' field
	if typeVal, typeOk := incomingObjMap["@type"].(string); typeOk {
		if !strings.EqualFold(typeVal, req.ResourceName) {
			err = errl.Errorf("@type field in body must match resource name in URL")
			apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("@type mismatch in update request", slog.String("expected", req.ResourceName), slog.String("got", typeVal), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	} else {
		// If @type is not specified, add it
		incomingObjMap["@type"] = req.ResourceName
		slog.Debug("Added missing @type field to update request", slog.String("type", req.ResourceName))
	}

	// Set the lastUpdate property. We overwrite whatever the user set.
	now := time.Now()
	lastUpdate := now.Format(time.RFC3339Nano)
	incomingObjMap["lastUpdate"] = lastUpdate

	// Add Seller and Buyer info. We overwrite whatever is in the incoming object, if any
	err = setSellerAndBuyerInfo(incomingObjMap, req.AuthUser.OrganizationIdentifier)
	if err != nil {
		err = errl.Errorf("failed to add Seller and Buyer info: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to add Seller and Buyer info", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	// Retrieve existing object from database to preserve CreatedAt
	existingObj, err := svc.getObject(req.ID, req.ResourceName)
	if err != nil {
		err = errl.Errorf("failed to get existing object for update: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to get existing object for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	if existingObj == nil {
		err = errl.Errorf("object not found")
		apiErr := NewApiError("404", "Not Found", err.Error(), fmt.Sprintf("%d", http.StatusNotFound), "")
		slog.Info("Object not found for update", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusNotFound, Body: apiErr}
	}

	// Version must be specified for update operations, and it must be greater than the current incomingVersion
	// TODO: ensure that this incomingVersion is greater than the existing one
	incomingVersion, _ := incomingObjMap["version"].(string)
	if incomingVersion == "" {
		err = errl.Errorf("version field is required for update operations")
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Version missing from update request", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	existingVersion := existingObj.Version

	// incomingVersion must be lexicographically greater than existingVersion
	if incomingVersion <= existingVersion {
		err = errl.Errorf("incoming version must be greater than existing version")
		apiErr := NewApiError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("incoming version must be greater than existing versio", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	// Merge incomingObjMap into existing object using RFC7396 (JSON Merge Patch)
	var existingMap map[string]any
	if err := json.Unmarshal(existingObj.Content, &existingMap); err != nil {
		err = errl.Errorf("failed to unmarshal existing object content for merge: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to unmarshal existing object content for merge", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	// RFC7396 merge implementation: modify target in place
	var mergeRFC7396 func(target, patch map[string]any)
	mergeRFC7396 = func(target, patch map[string]any) {
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

	mergeRFC7396(existingMap, incomingObjMap)

	// update incomingObjMap to the merged result so response/notification contains the final content
	incomingObjMap = existingMap

	incomingContent, err := json.Marshal(incomingObjMap)
	if err != nil {
		err = errl.Errorf("failed to marshal object content for update: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to marshal object content for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	obj := &repo.TMFObject{
		ID:         req.ID,
		Type:       req.ResourceName,
		Version:    incomingVersion,
		LastUpdate: lastUpdate,
		Content:    incomingContent,
		CreatedAt:  existingObj.CreatedAt,
		UpdatedAt:  time.Now(),
	}

	if err := svc.updateObject(obj); err != nil {
		err = errl.Errorf("failed to update object in service: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to update object in service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	slog.Info("Object updated successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Send TMForum notification (AttributeValueChangeEvent)
	eventType := toEventType(req.ResourceName, "AttributeValueChangeEvent")
	eventPayload := buildEventPayload(req, eventType, incomingObjMap)
	svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)

	return &Response{StatusCode: http.StatusOK, Body: incomingObjMap}
}

// DeleteGenericObject deletes a TMF object using generalized parameters.
func (svc *Service) DeleteGenericObject(req *Request) *Response {
	slog.Debug("DeleteGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Deleting an object can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	if err := svc.deleteObject(req.ID, req.ResourceName); err != nil {
		err = errl.Errorf("failed to delete object from service: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to delete object from service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	slog.Info("Object deleted successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Send TMForum notification
	eventType := toEventType(req.ResourceName, "DeleteEvent")
	minimal := map[string]any{
		"id":    req.ID,
		"@type": req.ResourceName,
		"href":  fmt.Sprintf("/tmf-api/%s/v5/%s/%s", req.APIfamily, req.ResourceName, req.ID),
	}
	eventPayload := buildEventPayload(req, eventType, minimal)
	svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)

	return &Response{StatusCode: http.StatusNoContent}
}

// ListGenericObjects retrieves all TMF objects of a given type using generalized parameters.
func (svc *Service) ListGenericObjects(req *Request) *Response {
	slog.Debug("ListGenericObjects called", slog.String("resourceName", req.ResourceName))

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	_, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := NewApiError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	objs, totalCount, err := svc.listObjects(req.ResourceName, req.QueryParams)
	if err != nil {
		err = errl.Errorf("failed to list objects from service: %w", err)
		apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to list objects from service", slog.Any("error", err), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	headers := make(map[string]string)
	headers["X-Total-Count"] = strconv.Itoa(totalCount)

	var responseData []map[string]any
	for _, obj := range objs {
		var item map[string]any
		err := json.Unmarshal(obj.Content, &item)
		if err != nil {
			err = errl.Errorf("failed to unmarshal object content for listing: %w", err)
			apiErr := NewApiError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
			slog.Error("Failed to unmarshal object content for listing", slog.Any("error", err), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
		}
		responseData = append(responseData, item)
	}

	// Handle partial field selection
	fieldsParam := req.QueryParams.Get("fields")
	if fieldsParam != "" {
		var fields []string
		if fieldsParam == "none" {
			fields = []string{"id", "href", "lastUpdate", "version"}
		} else {
			fields = strings.Split(fieldsParam, ",")
		}

		// Create a set of fields for quick lookup
		fieldSet := make(map[string]bool)
		for _, f := range fields {
			fieldSet[strings.TrimSpace(f)] = true
		}

		// Always include id, href, lastUpdate, version and @type
		fieldSet["id"] = true
		fieldSet["href"] = true
		fieldSet["lastUpdate"] = true
		fieldSet["version"] = true
		fieldSet["@type"] = true

		var filteredResponseData []map[string]any
		for _, item := range responseData {
			filteredItem := make(map[string]any)
			for key, value := range item {
				if fieldSet[key] {
					filteredItem[key] = value
				}
			}
			filteredResponseData = append(filteredResponseData, filteredItem)
		}
		responseData = filteredResponseData
	}

	slog.Info("Objects listed successfully", slog.Int("count", len(responseData)), slog.String("resourceName", req.ResourceName))
	return &Response{StatusCode: http.StatusOK, Headers: headers, Body: responseData}
}

// ToKebabCase converts a camelCase string to kebab-case.
// For example: "productOffering" becomes "product-offering".
func ToKebabCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}

// toEventType composes the TMF event type name from a resource name and a suffix.
func toEventType(resourceName, suffix string) string {
	// Convert resourceName to PascalCase
	parts := strings.Split(resourceName, "-")
	if len(parts) > 1 {
		for i, p := range parts {
			if len(p) == 0 {
				continue
			}
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
		return strings.Join(parts, "") + suffix
	}
	if len(resourceName) == 0 {
		return suffix
	}
	return strings.ToUpper(resourceName[:1]) + resourceName[1:] + suffix
}

// buildEventPayload builds a generic TMF event envelope.
func buildEventPayload(req *Request, eventType string, resource any) map[string]any {
	return map[string]any{
		"eventId":      uuid.NewString(),
		"eventTime":    time.Now().Format(time.RFC3339Nano),
		"eventType":    eventType,
		"apiFamily":    req.APIfamily,
		"resourceName": req.ResourceName,
		"resourceId":   req.ID,
		"resourcePath": fmt.Sprintf("/tmf-api/%s/v5/%s", req.APIfamily, req.ResourceName),
		"event": map[string]any{
			"resource": resource,
		},
	}
}
