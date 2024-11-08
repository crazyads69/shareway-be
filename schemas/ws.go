package schemas

// WebSocketMessage represents a message to be sent via WebSocket
type WebSocketMessage struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"` // Additional data to be sent with the message
}
