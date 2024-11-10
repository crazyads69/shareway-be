package schemas

import "github.com/google/uuid"

// Notification represents a notification to be sent to a device
type Notification struct {
	Token string            `json:"token"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"` // Additional data to be sent with the notification (optional)
}

// CreateNotificationRequest represents the request to create a new notification
type CreateNotificationRequest struct {
	Title string            `json:"title" binding:"required" validate:"required"`
	Body  string            `json:"body" binding:"required" validate:"required"`
	Data  map[string]string `json:"data,omitempty"` // Additional data to be sent with the notification (optional)
}

// CreateNotificationResponse represents the response to a create notification request
type CreateNotificationResponse struct {
	NotificationID uuid.UUID `json:"notification_id" binding:"required,uuid"`
}

// Define CreateTestWebsocketRequest
type CreateTestWebsocketRequest struct {
	Message string `json:"message" binding:"required" validate:"required"`
}

type NotificationPayload struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}
