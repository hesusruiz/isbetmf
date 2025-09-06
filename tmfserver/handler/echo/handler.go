package echo

import (
	"io"

	svc "github.com/hesusruiz/isbetmf/tmfserver/service"
	"github.com/labstack/echo/v4"
)

// Handler is the handler for the v5 API.
type Handler struct {
	service *svc.Service
}

// NewHandler creates a new handler.
func NewHandler(s *svc.Service) *Handler {
	return &Handler{service: s}
}

// HelloWorld is a simple hello world handler.
func (h *Handler) HelloWorld(c echo.Context) error {
	resp := &svc.Response{
		StatusCode: 200,
		Body:       "Hello, World!",
	}
	return SendResponse(c, resp)
}

// CreateGenericObject creates a new TMF object using generalized parameters.
func (h *Handler) CreateGenericObject(c echo.Context) error {
	body, _ := io.ReadAll(c.Request().Body)
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       svc.HttpMethodAliases[c.Request().Method],
		APIfamily:    c.Param("apiFamily"),
		ResourceName: c.Param("resourceName"),
		Body:         body,
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.CreateGenericObject(req)
	return SendResponse(c, resp)
}

// GetGenericObject retrieves a TMF object using generalized parameters.
func (h *Handler) GetGenericObject(c echo.Context) error {
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       svc.HttpMethodAliases[c.Request().Method],
		ResourceName: c.Param("resourceName"),
		ID:           c.Param("id"),
		QueryParams:  c.QueryParams(),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.GetGenericObject(req)
	return SendResponse(c, resp)
}

// UpdateGenericObject updates an existing TMF object using generalized parameters.
func (h *Handler) UpdateGenericObject(c echo.Context) error {
	body, _ := io.ReadAll(c.Request().Body)
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       svc.HttpMethodAliases[c.Request().Method],
		ResourceName: c.Param("resourceName"),
		ID:           c.Param("id"),
		Body:         body,
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.UpdateGenericObject(req)
	return SendResponse(c, resp)
}

// DeleteGenericObject deletes a TMF object using generalized parameters.
func (h *Handler) DeleteGenericObject(c echo.Context) error {
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       svc.HttpMethodAliases[c.Request().Method],
		ResourceName: c.Param("resourceName"),
		ID:           c.Param("id"),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.DeleteGenericObject(req)
	return SendResponse(c, resp)
}

// ListGenericObjects retrieves all TMF objects of a given type using generalized parameters.
func (h *Handler) ListGenericObjects(c echo.Context) error {
	jwtToken := svc.ExtractJWTToken(c.Request().Header.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Request().Method,
		Action:       "LIST",
		ResourceName: c.Param("resourceName"),
		QueryParams:  c.QueryParams(),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.ListGenericObjects(req)
	return SendResponse(c, resp)
}

func SendResponse(c echo.Context, resp *svc.Response) error {
	for key, value := range resp.Headers {
		c.Response().Header().Set(key, value)
	}
	if resp.Body != nil {
		return c.JSON(resp.StatusCode, resp.Body)
	}
	return c.NoContent(resp.StatusCode)
}
