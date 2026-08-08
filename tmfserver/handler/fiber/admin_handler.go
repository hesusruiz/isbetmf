package fiber

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hesusruiz/isbetmf/internal/html"
	"github.com/hesusruiz/isbetmf/tmfserver/service"
)

//go:embed templates/*
var templatesFS embed.FS

const adminSessionCookieName = "admin_session"

type AdminHandler struct {
	service      *service.Service
	htmlRenderer *html.Renderer
}

func NewAdminHandler(app *fiber.App, s *service.Service) *AdminHandler {

	htmlRenderer, err := html.NewRenderer(true, &templatesFS, "templates", "tmfserver/handler/fiber/templates", ".html")
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

	admin.Get("/login", h.ShowLogin)
	admin.Post("/login", h.ProcessLogin)
	admin.Get("/logout", h.Logout)

	admin.Use(h.RequireAuth)

	admin.Get("/", h.Dashboard)
	admin.Get("/pages/:pageName", h.ShowPage)
	admin.Get("/:resourceName", h.ListObjects)
	admin.Get("/:resourceName/:id", h.ViewObject)
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	data := map[string]any{
		"settings": "active",
		"service":  h.service,
	}
	return h.render(c, "settings", data)
}

func (h *AdminHandler) ShowPage(c *fiber.Ctx) error {
	pageName := c.Params("pageName")
	data := map[string]any{
		pageName:  "active",
		"service": h.service,
	}
	return h.render(c, pageName, data)
}

func (h *AdminHandler) RequireAuth(c *fiber.Ctx) error {
	path := c.Path()
	if path == "/admin/login" || path == "/admin/logout" {
		return c.Next()
	}

	cookie := c.Cookies(adminSessionCookieName)
	if cookie == "" || !h.isValidSession(cookie) {
		return c.Redirect("/admin/login")
	}

	return c.Next()
}

func (h *AdminHandler) ShowLogin(c *fiber.Ctx) error {
	cookie := c.Cookies(adminSessionCookieName)
	if cookie != "" && h.isValidSession(cookie) {
		return c.Redirect("/admin/")
	}
	return h.render(c, "login", map[string]any{})
}

func (h *AdminHandler) ProcessLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	adminToken := h.service.AdminToken()
	if adminToken == "" || password != adminToken {
		data := map[string]any{
			"Error": "Invalid password",
			"Email": email,
		}
		return h.render(c, "login", data)
	}

	sessionToken := h.createSessionToken()
	c.Cookie(&fiber.Cookie{
		Name:     adminSessionCookieName,
		Value:    sessionToken,
		Path:     "/admin",
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return c.Redirect("/admin/")
}

func (h *AdminHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     adminSessionCookieName,
		Value:    "",
		Path:     "/admin",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Redirect("/admin/login")
}

func (h *AdminHandler) isValidSession(tokenStr string) bool {
	adminToken := h.service.AdminToken()
	if adminToken == "" {
		return false
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return false
	}

	expStr, sig := parts[0], parts[1]
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expUnix {
		return false
	}

	expectedSig := h.computeSessionHMAC(adminToken, expStr)
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

func (h *AdminHandler) createSessionToken() string {
	adminToken := h.service.AdminToken()
	exp := time.Now().Add(24 * time.Hour).Unix()
	expStr := strconv.FormatInt(exp, 10)
	sig := h.computeSessionHMAC(adminToken, expStr)
	return fmt.Sprintf("%s.%s", expStr, sig)
}

func (h *AdminHandler) computeSessionHMAC(adminToken string, expStr string) string {
	mac := hmac.New(sha256.New, []byte(adminToken))
	mac.Write([]byte("admin_session:" + expStr))
	return hex.EncodeToString(mac.Sum(nil))
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
		Action:       service.ActionLIST,
		APIfamily:    apiFamily,
		APIVersion:   "v4",
		ResourceName: resourceName,
		QueryParams:  url.Values{"limit": []string{"20"}}, // Default limit
	}

	// Create a context with a timeout of 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp := h.service.ListTMFObjects(ctx, req)

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
		Action:       service.ActionREAD,
		APIfamily:    apiFamily,
		APIVersion:   "v4",
		ResourceName: resourceName,
		ID:           id,
	}

	// Create a context with a timeout of 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp := h.service.GetTMFObject(ctx, req)

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
