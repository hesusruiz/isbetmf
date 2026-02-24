package fiber

import (
	"net/http"
	"net/url"
	"strings"

	"encoding/json"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/hesusruiz/isbetmf/internal/errl"
	svc "github.com/hesusruiz/isbetmf/tmfserver/service"
	"github.com/hesusruiz/isbetmf/types"
)

// Handler is the handler for the TMF API (both V4 and V5).
type Handler struct {
	service *svc.Service
}

// extractAPIVersion extracts the API version from the URL path
func extractAPIVersion(path string) string {
	// Expected path format: /tmf-api/{apiFamily}/v{version}/...
	path = strings.Trim(path, "/")
	for part := range strings.SplitSeq(path, "/") {
		if strings.EqualFold(part, "v5") {
			return "v5"
		}
		if strings.EqualFold(part, "v4") {
			return "v4"
		}
	}
	// Default to v5 if not found
	return "v5"
}

// ExtractJWTToken extracts the JWT token from the Authorization header.
// It handles both "Bearer <token>" and raw token formats.
// If the token is not found, it returns an empty string.
func ExtractJWTToken(authHeader string) string {
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return tokenString
}

// NewHandler creates a new handler.
func NewHandler(s *svc.Service) *Handler {
	return &Handler{service: s}
}

// Health is a simple hello world handler.
// To exercise the full path as much as possible, we simulate a LIST request with limit = 1
func (h *Handler) Health(c *fiber.Ctx) error {

	// Create a request to list a catalog with limit=1
	queryParams, _ := url.ParseQuery("limit=1")
	req := &svc.Request{
		HealthRequest: true,
		Method:        "GET",
		Action:        svc.HttpActions["LIST"],
		APIfamily:     "productCatalogManagement",
		APIVersion:    "v4",
		ResourceName:  "catalog",
		QueryParams:   queryParams,
		AccessToken:   "",
	}

	resp := h.service.ListGenericObjects(req)
	return SendResponse(c, resp)

}

// CreateHubSubscription creates a new notification subscription (hub)
func (h *Handler) CreateHubSubscription(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	req := &svc.Request{
		Method:      c.Method(),
		Action:      svc.HttpActions[c.Method()],
		APIfamily:   c.Params("apiFamily"),
		APIVersion:  apiVersion,
		Body:        c.Body(),
		AccessToken: jwtToken,
	}

	resp := h.service.CreateHubSubscription(req)
	return SendResponse(c, resp)
}

// DeleteHubSubscription deletes an existing notification subscription (hub)
func (h *Handler) DeleteHubSubscription(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	idParam, _ := url.QueryUnescape(c.Params("id"))
	req := &svc.Request{
		Method:      c.Method(),
		Action:      svc.HttpActions[c.Method()],
		APIfamily:   c.Params("apiFamily"),
		APIVersion:  apiVersion,
		ID:          idParam,
		AccessToken: jwtToken,
	}

	resp := h.service.DeleteHubSubscription(req)
	return SendResponse(c, resp)
}

// MockListener is a minimal endpoint to receive notifications locally for testing
func (h *Handler) MockListener(c *fiber.Ctx) error {
	path := string(c.Request().URI().Path())
	body := c.Body()
	if len(body) > 0 {
		var payload any
		if err := json.Unmarshal(body, &payload); err == nil {
			slog.Info("listener received event", slog.String("path", path), slog.Int("bytes", len(body)), slog.Any("body", payload))
		} else {
			slog.Info("listener received event", slog.String("path", path), slog.Int("bytes", len(body)), slog.String("bodyRaw", string(body)))
		}
	} else {
		slog.Info("listener received event", slog.String("path", path), slog.Int("bytes", 0))
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// CreateGenericObject creates a new TMF object using generalized parameters.
func (h *Handler) CreateGenericObject(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	resourceName := c.Params("resourceName")
	if resourceName == "hub" {
		slog.Debug("handling hub creation")
		return h.CreateHubSubscription(c)
	}

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	// The tokenMap may be nil, which means that the user did not send any authorization header, and
	// this will be checked in the service downstream.
	tokenMap, authUser, err := h.service.ProcessAccessToken(jwtToken)
	if err != nil {
		resp := svc.ErrorResponsef(http.StatusUnauthorized, "invalid access token: %w", errl.Error(err))
		return SendResponse(c, resp)
	}

	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpActions[c.Method()],
		APIfamily:    c.Params("apiFamily"),
		APIVersion:   apiVersion,
		ResourceName: resourceName,
		Body:         c.Body(),
		AccessToken:  jwtToken, // Store the raw JWT token
		AuthUser:     *authUser,
		TokenMap:     tokenMap,
	}

	resp := h.service.CreateGenericObject(req)
	return SendResponse(c, resp)
}

// GetGenericObject retrieves a TMF object using generalized parameters.
func (h *Handler) GetGenericObject(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	// The tokenMap may be nil, which means that the user did not send any authorization header, and
	// this will be checked in the service downstream.
	tokenMap, authUser, err := h.service.ProcessAccessToken(jwtToken)
	if err != nil {
		resp := svc.ErrorResponsef(http.StatusUnauthorized, "invalid access token: %w", errl.Error(err))
		return SendResponse(c, resp)
	}

	queryParams, _ := url.ParseQuery(string(c.Request().URI().QueryString()))
	idParam, _ := url.QueryUnescape(c.Params("id"))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpActions[c.Method()],
		APIfamily:    c.Params("apiFamily"),
		APIVersion:   apiVersion,
		ResourceName: c.Params("resourceName"),
		ID:           idParam,
		QueryParams:  queryParams,
		AccessToken:  jwtToken,
		AuthUser:     *authUser,
		TokenMap:     tokenMap,
	}

	resp := h.service.GetGenericObject(req)
	return SendResponse(c, resp)
}

// UpdateGenericObject updates an existing TMF object using generalized parameters.
func (h *Handler) UpdateGenericObject(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	// The tokenMap may be nil, which means that the user did not send any authorization header, and
	// this will be checked in the service downstream.
	tokenMap, authUser, err := h.service.ProcessAccessToken(jwtToken)
	if err != nil {
		resp := svc.ErrorResponsef(http.StatusUnauthorized, "invalid access token: %w", errl.Error(err))
		return SendResponse(c, resp)
	}

	idParam, _ := url.QueryUnescape(c.Params("id"))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpActions[c.Method()],
		APIfamily:    c.Params("apiFamily"),
		APIVersion:   apiVersion,
		ResourceName: c.Params("resourceName"),
		ID:           idParam,
		Body:         c.Body(),
		AccessToken:  jwtToken, // Store the raw JWT token
		AuthUser:     *authUser,
		TokenMap:     tokenMap,
	}

	resp := h.service.UpdateGenericObject(req)
	return SendResponse(c, resp)
}

// DeleteGenericObject deletes a TMF object using generalized parameters.
func (h *Handler) DeleteGenericObject(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	resourceName := c.Params("resourceName")
	if resourceName == "hub" {
		slog.Debug("handling hub creation")
		return h.DeleteHubSubscription(c)
	}

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	// The tokenMap may be nil, which means that the user did not send any authorization header, and
	// this will be checked in the service downstream.
	tokenMap, authUser, err := h.service.ProcessAccessToken(jwtToken)
	if err != nil {
		resp := svc.ErrorResponsef(http.StatusUnauthorized, "invalid access token: %w", errl.Error(err))
		return SendResponse(c, resp)
	}

	idParam, _ := url.QueryUnescape(c.Params("id"))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpActions[c.Method()],
		APIfamily:    c.Params("apiFamily"),
		APIVersion:   apiVersion,
		ResourceName: c.Params("resourceName"),
		ID:           idParam,
		AccessToken:  jwtToken, // Store the raw JWT token
		AuthUser:     *authUser,
		TokenMap:     tokenMap,
	}

	resp := h.service.DeleteGenericObject(req)
	return SendResponse(c, resp)
}

// ListGenericObjects retrieves all TMF objects of a given type using generalized parameters.
func (h *Handler) ListGenericObjects(c *fiber.Ctx) error {
	jwtToken := ExtractJWTToken(c.Get("Authorization"))

	// Extract API version from the URL path
	apiVersion := extractAPIVersion(c.Path())

	// Authentication: process the AccessToken to extract caller info from its claims in the payload
	// The tokenMap may be nil, which means that the user did not send any authorization header, and
	// this will be checked in the service downstream.
	tokenMap, authUser, err := h.service.ProcessAccessToken(jwtToken)
	if err != nil {
		resp := svc.ErrorResponsef(http.StatusUnauthorized, "invalid access token: %w", errl.Error(err))
		return SendResponse(c, resp)
	}

	if authUser == nil {
		authUser = &types.AuthUser{}
	}

	queryParams, _ := url.ParseQuery(string(c.Request().URI().QueryString()))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpActions["LIST"],
		APIfamily:    c.Params("apiFamily"),
		APIVersion:   apiVersion,
		ResourceName: c.Params("resourceName"),
		QueryParams:  queryParams,
		AccessToken:  jwtToken, // Store the raw JWT token
		AuthUser:     *authUser,
		TokenMap:     tokenMap,
	}

	resp := h.service.ListGenericObjects(req)
	return SendResponse(c, resp)
}

func SendResponse(c *fiber.Ctx, resp *svc.Response) error {
	for key, value := range resp.Headers {
		c.Set(key, value)
	}
	if resp.Body != nil {
		return c.Status(resp.StatusCode).JSON(resp.Body)
	}
	return c.SendStatus(resp.StatusCode)
}
