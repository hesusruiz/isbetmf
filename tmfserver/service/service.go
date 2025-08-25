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
	"github.com/hesusruiz/isbetmf/pkg/apierror"
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

// CreateGenericObject creates a new TMF object using generalized parameters.
func (svc *Service) CreateGenericObject(req *Request) *Response {
	slog.Debug("CreateGenericObject called", slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))

	// Before doing anything, check if we can extract the calling user info
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Creating an object can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	var data map[string]any
	if err := json.Unmarshal(req.Body, &data); err != nil {
		err = errl.Errorf("failed to bind request body: %w", err)
		apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Failed to bind request body", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	id, _ := data["id"].(string)
	if id == "" {
		// If the incoming object does not have an 'id', we generate a new one
		// The format is "urn:ngsi-ld:{resource-in-kebab-case}:{uuid}"
		id = fmt.Sprintf("urn:ngsi-ld:%s:%s", ToKebabCase(req.ResourceName), uuid.NewString())
		data["id"] = id
		slog.Debug("Generated new ID for object", "id", id)

	}

	href, _ := data["href"].(string)
	if href == "" {
		// Add href to the object
		data["href"] = fmt.Sprintf("/tmf-api/%s/v5/%s/%s", req.APIfamily, req.ResourceName, id)
		slog.Debug("Set href", slog.String("href", data["href"].(string)))
	}

	// Check and process '@type' field
	if typeVal, typeOk := data["@type"].(string); typeOk {
		if !strings.EqualFold(typeVal, req.ResourceName) {
			err = errl.Errorf("@type mismatch: expected %s, got %s", req.ResourceName, typeVal)
			apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("@type mismatch", slog.String("expected", req.ResourceName), slog.String("got", typeVal), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	} else {
		// If @type is not specified, add it
		data["@type"] = req.ResourceName
		slog.Debug("Added missing @type field", slog.String("type", req.ResourceName))
	}

	// Set default version if not provided
	version, versionOk := data["version"].(string)
	if !versionOk || version == "" {
		version = "1.0"
		data["version"] = version // Update data map for content marshaling
		slog.Debug("Set default version", slog.String("version", version))
	}

	// Set the lastUpdate property
	now := time.Now()
	lastUpdate := now.Format(time.RFC3339Nano)
	data["lastUpdate"] = lastUpdate

	// Add Seller and Buyer info. We overwrite whatever is in the incoming object, if any
	err = setSellerAndBuyerInfo(data, req.AuthUser.OrganizationIdentifier)
	if err != nil {
		err = errl.Errorf("failed to add Seller and Buyer info: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to add Seller and Buyer info", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	content, err := json.Marshal(data)
	if err != nil {
		err = errl.Errorf("failed to marshal object content: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to marshal object content", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	obj := repo.NewTMFObject(id, req.ResourceName, version, lastUpdate, content)

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object.
	// ************************************************************************************************

	err = takeDecision(svc.ruleEngine, req, token, obj)
	if err != nil {
		err = errl.Errorf("user not authorized: %w", err)
		apiErr := apierror.NewError("403", "Forbidden", err.Error(), fmt.Sprintf("%d", http.StatusForbidden), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusForbidden, Body: apiErr}
	}

	// ************************************************************************************************
	// Now we can proceed.
	// ************************************************************************************************

	if err := svc.createObject(obj); err != nil {
		err = errl.Errorf("failed to create object in service: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to create object in service", slog.Any("error", err), slog.String("id", id), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	headers := make(map[string]string)
	headers["Location"] = data["href"].(string)
	slog.Info("Object created successfully", slog.String("id", id), slog.String("resourceName", req.ResourceName), slog.String("location", data["href"].(string)))

	return &Response{
		StatusCode: http.StatusCreated,
		Headers:    headers,
		Body:       data,
	}
}

// GetGenericObject retrieves a TMF object using generalized parameters.
func (svc *Service) GetGenericObject(req *Request) *Response {
	slog.Debug("GetGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Before doing anything, check if we can extract the calling user info
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	obj, err := svc.getObject(req.ID, req.ResourceName)
	if err != nil {
		err = errl.Errorf("failed to get object from service: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to get object from service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	if obj == nil {
		err = errl.Errorf("object not found")
		apiErr := apierror.NewError("404", "Not Found", err.Error(), fmt.Sprintf("%d", http.StatusNotFound), "")
		slog.Info("Object not found", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusNotFound, Body: apiErr}
	}

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object.
	// ************************************************************************************************

	err = takeDecision(svc.ruleEngine, req, token, obj)
	if err != nil {
		err = errl.Error(err)
		apiErr := apierror.NewError("403", "Forbidden", err.Error(), fmt.Sprintf("%d", http.StatusForbidden), "")
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
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
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

	// Before doing anything, check if we can extract the calling user info
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Updating an object can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	var incomingObjMap map[string]any
	if err := json.Unmarshal(req.Body, &incomingObjMap); err != nil {
		err = errl.Errorf("failed to bind request body: %w", err)
		apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Failed to bind request body for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	// If the ID is present in the body, ensure it matches the ID in the URL
	if bodyID, ok := incomingObjMap["id"]; ok {
		bodyIDStr, ok := bodyID.(string)
		if !ok || bodyIDStr != req.ID {
			err = errl.Errorf("ID in body must match ID in URL")
			apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("ID mismatch in update request", slog.String("url_id", req.ID), slog.Any("body_id", bodyID), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	}

	// Check and process '@type' field
	if typeVal, typeOk := incomingObjMap["@type"].(string); typeOk {
		if !strings.EqualFold(typeVal, req.ResourceName) {
			err = errl.Errorf("@type field in body must match resource name in URL")
			apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
			slog.Error("@type mismatch in update request", slog.String("expected", req.ResourceName), slog.String("got", typeVal), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
			return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
		}
	} else {
		// If @type is not specified, add it
		incomingObjMap["@type"] = req.ResourceName
		slog.Debug("Added missing @type field to update request", slog.String("type", req.ResourceName))
	}

	// Version must be specified for update operations
	version, versionOk := incomingObjMap["version"].(string)
	if !versionOk || version == "" {
		err = errl.Errorf("version field is required for update operations")
		apiErr := apierror.NewError("400", "Bad Request", err.Error(), fmt.Sprintf("%d", http.StatusBadRequest), "")
		slog.Error("Version missing from update request", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusBadRequest, Body: apiErr}
	}

	// Set the lastUpdate field in the incoming object
	now := time.Now()
	lastUpdate := now.Format(time.RFC3339Nano)
	incomingObjMap["lastUpdate"] = lastUpdate

	// Add Seller and Buyer info. We overwrite whatever is in the incoming object, if any
	err = setSellerAndBuyerInfo(incomingObjMap, req.AuthUser.OrganizationIdentifier)
	if err != nil {
		err = errl.Errorf("failed to add Seller and Buyer info: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to add Seller and Buyer info", slog.Any("error", err), slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	incomingContent, err := json.Marshal(incomingObjMap)
	if err != nil {
		err = errl.Errorf("failed to marshal object content for update: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to marshal object content for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	// Get existing object to preserve CreatedAt
	existingObj, err := svc.getObject(req.ID, req.ResourceName)
	if err != nil {
		err = errl.Errorf("failed to get existing object for update: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to get existing object for update", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	if existingObj == nil {
		err = errl.Errorf("object not found")
		apiErr := apierror.NewError("404", "Not Found", err.Error(), fmt.Sprintf("%d", http.StatusNotFound), "")
		slog.Info("Object not found for update", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusNotFound, Body: apiErr}
	}

	// TODO: perform a JSON Merge operation, instead of replacing the whole object

	obj := &repo.TMFObject{
		ID:         req.ID,
		Type:       req.ResourceName,
		Version:    version,
		LastUpdate: lastUpdate,
		Content:    incomingContent,
		CreatedAt:  existingObj.CreatedAt,
		UpdatedAt:  time.Now(),
	}

	if err := svc.updateObject(obj); err != nil {
		err = errl.Errorf("failed to update object in service: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to update object in service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	slog.Info("Object updated successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return &Response{StatusCode: http.StatusOK, Body: incomingObjMap}
}

// DeleteGenericObject deletes a TMF object using generalized parameters.
func (svc *Service) DeleteGenericObject(req *Request) *Response {
	slog.Debug("DeleteGenericObject called", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))

	// Before doing anything, check if we can extract the calling user info
	token, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	// Deleting an object can not be done without authentication
	if len(token) == 0 {
		err = errl.Errorf("user not authenticated")
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request")
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	if err := svc.deleteObject(req.ID, req.ResourceName); err != nil {
		err = errl.Errorf("failed to delete object from service: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
		slog.Error("Failed to delete object from service", slog.Any("error", err), slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
		return &Response{StatusCode: http.StatusInternalServerError, Body: apiErr}
	}

	slog.Info("Object deleted successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return &Response{StatusCode: http.StatusNoContent}
}

// ListGenericObjects retrieves all TMF objects of a given type using generalized parameters.
func (svc *Service) ListGenericObjects(req *Request) *Response {
	slog.Debug("ListGenericObjects called", slog.String("resourceName", req.ResourceName))

	// Before doing anything, check if we can extract the calling user info
	_, err := svc.extractCallerInfo(req)
	if err != nil {
		err = errl.Errorf("invalid access token: %w", err)
		apiErr := apierror.NewError("401", "Unauthorized", err.Error(), fmt.Sprintf("%d", http.StatusUnauthorized), "")
		slog.Error("Unauthorized request", slog.Any("error", err))
		return &Response{StatusCode: http.StatusUnauthorized, Body: apiErr}
	}

	objs, totalCount, err := svc.listObjects(req.ResourceName, req.QueryParams)
	if err != nil {
		err = errl.Errorf("failed to list objects from service: %w", err)
		apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
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
			apiErr := apierror.NewError("500", "Internal Server Error", err.Error(), fmt.Sprintf("%d", http.StatusInternalServerError), "")
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
