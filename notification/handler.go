package notification

import (
	"github.com/gofiber/fiber/v2"
)

// HubHandler handles subscription requests.
type HubHandler struct {
	hubManager *HubManager
}

// NewHubHandler creates a new HubHandler.
func NewHubHandler(hub *HubManager) *HubHandler {
	return &HubHandler{hubManager: hub}
}

// Subscribe is the handler for POST /hub
func (h *HubHandler) Subscribe(c *fiber.Ctx) error {
	var sub Hub
	if err := c.BodyParser(&sub); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	h.hubManager.Subscribe(&sub)
	return c.SendStatus(fiber.StatusCreated)
}

// Unsubscribe is the handler for DELETE /hub
func (h *HubHandler) Unsubscribe(c *fiber.Ctx) error {
	id := c.Params("id")

	h.hubManager.Unsubscribe(id)
	return c.SendStatus(fiber.StatusNoContent)
}
