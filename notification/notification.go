package notification

import "time"

// Notification represents a TMForum notification.
type Notification struct {
	EventID       string    `json:"eventId"`
	EventTime     time.Time `json:"eventTime"`
	EventType     string    `json:"eventType"`
	CorrelationID string    `json:"correlationId,omitempty"`
	Domain        string    `json:"domain,omitempty"`
	Title         string    `json:"title,omitempty"`
	Description   string    `json:"description,omitempty"`
	Priority      string    `json:"priority,omitempty"`
	Event         any       `json:"event"`
}
