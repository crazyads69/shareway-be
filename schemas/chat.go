package schemas

import "github.com/google/uuid"

// Chat represents a notification to be sent to a device
type Chat struct {
	Token string            `json:"token"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"` // Additional data to be sent with the notification (optional)
}

// CreateChatRequest represents the request to create a new notification
type CreateChatRequest struct {
	Title string            `json:"title" binding:"required" validate:"required"`
	Body  string            `json:"body" binding:"required" validate:"required"`
	Data  map[string]string `json:"data,omitempty"` // Additional data to be sent with the notification (optional)
}

// CreateChatResponse represents the response to a create notification request
type CreateChatResponse struct {
	ChatID uuid.UUID `json:"notification_id" binding:"required,uuid"`
}
