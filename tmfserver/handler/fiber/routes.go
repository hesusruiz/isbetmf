package fiber

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers the routes for the TMF API.
func (h *Handler) RegisterRoutes(app *fiber.App) {

	// Mock listener for local testing (accepts any listener path)
	app.Post("/listener/*", h.MockListener)

	// Health check
	app.Get("/health", h.Health)

	// Group routes for TMF API V5
	tmfApiV5 := app.Group("/tmf-api/:apiFamily/v5")

	// Notifications Hub routes for V5
	tmfApiV5.Post("/hub", h.CreateHubSubscription)
	tmfApiV5.Delete("/hub/:id", h.DeleteHubSubscription)

	// Generalized routes for TMF API V5 resources
	// Collection operations (List and Create)
	tmfApiV5.Get("/:resourceName", h.ListGenericObjects)
	tmfApiV5.Post("/:resourceName", h.CreateGenericObject)

	// Individual resource operations (Get, Update, Delete)
	tmfApiV5.Get("/:resourceName/:id", h.GetGenericObject)
	tmfApiV5.Patch("/:resourceName/:id", h.UpdateGenericObject)
	tmfApiV5.Delete("/:resourceName/:id", h.DeleteGenericObject)

	// Group routes for TMF API V4
	tmfApiV4 := app.Group("/tmf-api/:apiFamily/v4")

	// Notifications Hub routes for V4
	tmfApiV4.Post("/hub", h.CreateHubSubscription)
	tmfApiV4.Delete("/hub/:id", h.DeleteHubSubscription)

	// Generalized routes for TMF API V4 resources
	// Collection operations (List and Create)
	tmfApiV4.Get("/:resourceName", h.ListGenericObjects)
	tmfApiV4.Post("/:resourceName", h.CreateGenericObject)

	// Individual resource operations (Get, Update, Delete)
	tmfApiV4.Get("/:resourceName/:id", h.GetGenericObject)
	tmfApiV4.Patch("/:resourceName/:id", h.UpdateGenericObject)
	tmfApiV4.Delete("/:resourceName/:id", h.DeleteGenericObject)

}
