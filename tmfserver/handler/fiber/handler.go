package fiber

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	svc "github.com/hesusruiz/isbetmf/tmfserver/service"
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
func (h *Handler) HelloWorld(c *fiber.Ctx) error {
	resp := &svc.Response{
		StatusCode: 200,
		Body:       "Hello, World!",
	}
	return sendResponse(c, resp)
}

// CreateGenericObject creates a new TMF object using generalized parameters.
func (h *Handler) CreateGenericObject(c *fiber.Ctx) error {
	jwtToken := svc.ExtractJWTToken(c.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpMethodAliases[c.Method()],
		APIfamily:    c.Params("apiFamily"),
		ResourceName: c.Params("resourceName"),
		Body:         c.Body(),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.CreateGenericObject(req)
	return sendResponse(c, resp)
}

// GetGenericObject retrieves a TMF object using generalized parameters.
func (h *Handler) GetGenericObject(c *fiber.Ctx) error {
	jwtToken := svc.ExtractJWTToken(c.Get("Authorization"))

	queryParams, _ := url.ParseQuery(string(c.Request().URI().QueryString()))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpMethodAliases[c.Method()],
		ResourceName: c.Params("resourceName"),
		ID:           c.Params("id"),
		QueryParams:  queryParams,
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.GetGenericObject(req)
	return sendResponse(c, resp)
}

// UpdateGenericObject updates an existing TMF object using generalized parameters.
func (h *Handler) UpdateGenericObject(c *fiber.Ctx) error {
	jwtToken := svc.ExtractJWTToken(c.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpMethodAliases[c.Method()],
		ResourceName: c.Params("resourceName"),
		ID:           c.Params("id"),
		Body:         c.Body(),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.UpdateGenericObject(req)
	return sendResponse(c, resp)
}

// DeleteGenericObject deletes a TMF object using generalized parameters.
func (h *Handler) DeleteGenericObject(c *fiber.Ctx) error {
	jwtToken := svc.ExtractJWTToken(c.Get("Authorization"))

	req := &svc.Request{
		Method:       c.Method(),
		Action:       svc.HttpMethodAliases[c.Method()],
		ResourceName: c.Params("resourceName"),
		ID:           c.Params("id"),
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.DeleteGenericObject(req)
	return sendResponse(c, resp)
}

// ListGenericObjects retrieves all TMF objects of a given type using generalized parameters.
func (h *Handler) ListGenericObjects(c *fiber.Ctx) error {
	jwtToken := svc.ExtractJWTToken(c.Get("Authorization"))

	queryParams, _ := url.ParseQuery(string(c.Request().URI().QueryString()))
	req := &svc.Request{
		Method:       c.Method(),
		Action:       "LIST",
		ResourceName: c.Params("resourceName"),
		QueryParams:  queryParams,
		AccessToken:  jwtToken, // Store the raw JWT token
	}

	resp := h.service.ListGenericObjects(req)
	return sendResponse(c, resp)
}

func sendResponse(c *fiber.Ctx, resp *svc.Response) error {
	for key, value := range resp.Headers {
		c.Set(key, value)
	}
	if resp.Body != nil {
		return c.Status(resp.StatusCode).JSON(resp.Body)
	}
	return c.SendStatus(resp.StatusCode)
}
