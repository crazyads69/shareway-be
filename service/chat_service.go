package service

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"shareway/repository"
	"shareway/util"
	"sync"
)

var webSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

type Client struct {
	conn *websocket.Conn
	room *Room
	send chan []byte
}

type Room struct {
	id         string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewRoom(id string) *Room {
	return &Room{
		id:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

type Manager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// Retrieve or create a room
func (m *Manager) getRoom(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if room, exists := m.rooms[roomID]; exists {
		return room
	}
	newRoom := NewRoom(roomID)
	m.rooms[roomID] = newRoom
	go newRoom.run()
	return newRoom
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case message := <-r.broadcast:
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	otherId := r.URL.Query().Get("otherId")

	if userId == "" || otherId == "" {
		http.Error(w, "Missing userId or otherId", http.StatusBadRequest)
		return
	}

	// For multi-room between 2 users, we can cat with the ride offer with same logic
	if userId < otherId {
		userId, otherId = otherId, userId
	}
	roomId := userId + "-" + otherId

	if roomId == "" {
		http.Error(w, "Missing roomId", http.StatusBadRequest)
		return
	}

	log.Printf("New connection for room: %s", roomId)
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Find or create the chat room
	room := m.getRoom(roomId)

	// Create a new client
	client := &Client{
		conn: conn,
		room: room,
		send: make(chan []byte),
	}

	// Register the client in the room
	room.register <- client

	go client.readMessages()
	go client.writeMessages()
}

func (c *Client) readMessages() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		// Broadcast the message to all clients in the room
		c.room.broadcast <- message

		// Handle Store message here
		/*
			1. Store message directly to DB
			2. Pub to message broker (pub/sub)
			3. Store in cache and worker will handle it
			4. Internal channel for manager to async handle it (* recommended)
		*/
	}
}

func (c *Client) writeMessages() {
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

type IChatService interface {
	HandleWSService(w http.ResponseWriter, r *http.Request)
}

type ChatService struct {
	repo    repository.IChatRepository
	cfg     util.Config
	Manager *Manager
}

func NewChatService(repo repository.IChatRepository, cfg util.Config, manager *Manager) IChatService {
	return &ChatService{
		repo:    repo,
		cfg:     cfg,
		Manager: manager,
	}
}

func (cs *ChatService) HandleWSService(w http.ResponseWriter, r *http.Request) {
	cs.Manager.ServeWs(w, r)
}
