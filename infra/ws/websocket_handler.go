// shareway/infra/ws/websocket_handler.go
package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"shareway/helper"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Note: In production, implement proper origin checks
	},
}

// ServeWs handles WebSocket requests from the peer
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: userID}
	client.hub.register <- client

	// Start goroutines for pumping messages
	go client.writePump()
	go client.readPump()
}

// WebSocketHandler is the Gin handler for WebSocket connections
func WebSocketHandler(ctx *gin.Context, hub *Hub) {
	// Get the user ID from the query parameters
	userID := ctx.Query("user_id")

	// Check if userID is provided
	if userID == "" {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("user_id is required"),
			"User ID is required",
			"Yêu cầu ID người dùng",
		)
		helper.GinResponse(
			ctx,
			http.StatusBadRequest,
			response,
		)
	}

	// Serve WebSocket connection
	ServeWs(hub, ctx.Writer, ctx.Request, userID)
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(hub *Hub, messageType string, data interface{}) {
	message := Message{
		Type: messageType,
		Data: data,
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("error marshalling broadcast message: %v", err)
		return
	}
	hub.broadcast <- jsonMessage
}
