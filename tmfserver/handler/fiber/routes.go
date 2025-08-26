package fiber

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registers the routes for the TMF API.
func (h *Handler) RegisterRoutes(app *fiber.App) {

	// Mock listener for local testing (accepts any listener path)
	app.Post("/listener/*", h.MockListener)

	// Group routes for TMF API
	tmfApi := app.Group("/tmf-api/:apiFamily/v5")

	// Notifications Hub routes
	tmfApi.Post("/hub", h.CreateHubSubscription)
	tmfApi.Delete("/hub/:id", h.DeleteHubSubscription)

	// Generalized routes for TMF API resources
	// Collection operations (List and Create)
	tmfApi.Get("/:resourceName", h.ListGenericObjects)
	tmfApi.Post("/:resourceName", h.CreateGenericObject)

	// Individual resource operations (Get, Update, Delete)
	tmfApi.Get("/:resourceName/:id", h.GetGenericObject)
	tmfApi.Patch("/:resourceName/:id", h.UpdateGenericObject)
	tmfApi.Delete("/:resourceName/:id", h.DeleteGenericObject)

	// HelloWorld route (health check)
	app.Get("/", h.HelloWorld)

}
