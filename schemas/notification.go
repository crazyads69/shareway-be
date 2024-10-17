package schemas

// Notification represents a notification to be sent to a device
type Notification struct {
	Token string `json:"token"`
	Title string `json:"title"`
	Body  string `json:"body"`
}
