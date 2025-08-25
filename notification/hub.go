package notification

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// Hub represents a subscriber to notifications.
type Hub struct {
	ID       string `json:"id,omitempty"`
	Callback string `json:"callback"`
	Query    string `json:"query,omitempty"`
}

// HubManager manages subscribers and dispatches notifications.
type HubManager struct {
	Subscribers map[string]*Hub
	mutex       sync.Mutex
}

// NewHubManager creates a new HubManager.
func NewHubManager() *HubManager {
	return &HubManager{
		Subscribers: make(map[string]*Hub),
	}
}

// Subscribe adds a new subscriber to the hub.
func (h *HubManager) Subscribe(hub *Hub) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	hub.ID = uuid.New().String()
	h.Subscribers[hub.ID] = hub
}

// Unsubscribe removes a subscriber from the hub.
func (h *HubManager) Unsubscribe(id string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.Subscribers, id)
}

// Dispatch sends a notification to all subscribers.
func (h *HubManager) Dispatch(notification *Notification) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, sub := range h.Subscribers {
		go func(hub *Hub) {
			jsonData, err := json.Marshal(notification)
			if err != nil {
				// Handle error
				return
			}

			_, err = http.Post(hub.Callback, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				// Handle error
			}
		}(sub)
	}
}