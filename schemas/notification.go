package schemas

// Notification represents a notification to be sent to a device
type Notification struct {
	Token string            `json:"token"`
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"` // Additional data to be sent with the notification (optional)
}
