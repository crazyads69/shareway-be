package ws

import (
	"fmt"
	"log"
	"net/http"
	"shareway/helper"
	"shareway/middleware"

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
func WebSocketHandler(ctx *gin.Context) {
	// Get payload from context
	payload := ctx.MustGet(middleware.AuthorizationPayloadKey)

	// Convert payload to map
	data, err := helper.ConvertToPayload(payload)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			fmt.Errorf("failed to convert payload"),
			"Failed to convert payload",
			"Không thể chuyển đổi payload",
		)
		helper.GinResponse(ctx, http.StatusInternalServerError, response)
		return
	}

	// Convert userid from uuid to string
	userIDString := data.UserID.String()

	// Serve WebSocket connection
	ServeWs(hub, ctx.Writer, ctx.Request, userIDString)
}

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(message []byte) {
	hub.broadcast <- message
}
