package fiber

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/hesusruiz/isbetmf/internal/html"
	"github.com/hesusruiz/isbetmf/tmfserver/service"
)

//go:embed templates/*.html
var templatesFS embed.FS

type AdminHandler struct {
	service      *service.Service
	htmlRenderer *html.Renderer
}

func NewAdminHandler(app *fiber.App, s *service.Service) *AdminHandler {

	htmlRenderer, err := html.NewRenderer(true, &templatesFS, "templates", "internal/admin/templates", ".html")
	if err != nil {
		panic(fmt.Errorf("failed to create admin templates renderer: %w", err))
	}

	h := &AdminHandler{
		service:      s,
		htmlRenderer: htmlRenderer,
	}

	h.registerRoutes(app)

	return h

}

func (h *AdminHandler) registerRoutes(app *fiber.App) {
	admin := app.Group("/admin")

	admin.Get("/", h.Dashboard)
	admin.Get("/:resourceName", h.ListObjects)
	admin.Get("/:resourceName/:id", h.ViewObject)
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	return h.render(c, "dashboard", nil)
}

func (h *AdminHandler) ListObjects(c *fiber.Ctx) error {
	resourceName := c.Params("resourceName")

	// Default to v4 for now, could be made configurable or inferred
	apiFamily := "productCatalogManagement"
	switch resourceName {
	case "agreement":
		apiFamily = "agreementManagement"
	case "individual", "organization":
		apiFamily = "party"
	}

	req := &service.Request{
		Method:       "GET",
		Action:       service.HttpActions["LIST"],
		APIfamily:    apiFamily,
		APIVersion:   "v4",
		ResourceName: resourceName,
		QueryParams:  url.Values{"limit": []string{"20"}}, // Default limit
	}

	// We need to bypass auth for admin screens or implement a proper admin login.
	// For now, assuming the service layer allows internal calls or we mock the token if needed.
	// However, the service layer checks for tokens.
	// Let's assume for this task we are running in a trusted environment or the service allows it.
	// Actually, the service.ListGenericObjects checks for permissions.
	// We might need to inject a "superuser" token or context if the admin is internal.
	// For now, let's try calling it directly. The service layer might fail if no token is present
	// and the policy requires it.

	resp := h.service.ListTMFObjects(req)

	if resp.StatusCode >= 400 {
		return c.Status(resp.StatusCode).SendString(fmt.Sprintf("Error fetching objects: %v", resp.Body))
	}

	data := map[string]any{
		"ResourceName": resourceName,
		"Objects":      resp.Body,
	}

	return h.render(c, "list", data)
}

func (h *AdminHandler) ViewObject(c *fiber.Ctx) error {
	resourceName := c.Params("resourceName")
	id := c.Params("id")

	apiFamily := "productCatalogManagement"
	switch resourceName {
	case "agreement":
		apiFamily = "agreementManagement"
	case "individual", "organization":
		apiFamily = "party"
	}

	req := &service.Request{
		Method:       "GET",
		Action:       service.HttpActions["READ"],
		APIfamily:    apiFamily,
		APIVersion:   "v4",
		ResourceName: resourceName,
		ID:           id,
	}

	resp := h.service.GetTMFObject(req)

	if resp.StatusCode >= 400 {
		return c.Status(resp.StatusCode).SendString(fmt.Sprintf("Error fetching object: %v", resp.Body))
	}

	// Pretty print JSON
	jsonBytes, _ := json.MarshalIndent(resp.Body, "", "  ")

	data := map[string]any{
		"ResourceName": resourceName,
		"ID":           id,
		"Object":       resp.Body,
		"JSON":         string(jsonBytes),
	}

	return h.render(c, "detail", data)
}

func (h *AdminHandler) render(c *fiber.Ctx, name string, data map[string]any) error {
	c.Set("Content-Type", "text/html")
	return h.htmlRenderer.RenderFiber(c, name, data)
}
