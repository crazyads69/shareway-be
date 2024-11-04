package ws

import "encoding/json"

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[string]*Client // Use userID as key for quick lookups
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// // Global hub instance
// var hub *Hub

// // Initialize the hub and start its main loop
// func init() {
// 	hub = NewHub()
// 	go hub.Run()
// }

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

// Run starts the main loop for the Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client // Register new client
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID) // Unregister client
				close(client.send)
			}
		case message := <-h.broadcast:
			// Broadcast message to all clients
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, messageType string, data interface{}) {
	if client, ok := h.clients[userID]; ok {
		message := map[string]interface{}{
			"type": messageType,
			"data": data,
		}
		jsonMessage, _ := json.Marshal(message)
		select {
		case client.send <- jsonMessage:
		default:
			close(client.send)
			delete(h.clients, userID)
		}
	}
}