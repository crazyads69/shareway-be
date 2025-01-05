// shareway/infra/ws/client.go
package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	// Reconnection parameters
	initialReconnectDelay = 1 * time.Second
	maxReconnectDelay     = 30 * time.Second
	reconnectMultiplier   = 1.5
)

type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan []byte
	userID         string
	mu             sync.Mutex // Protects conn
	isConnected    bool
	reconnectDelay time.Duration
	done           chan struct{}
	wsURL          string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, wsURL string) *Client {
	return &Client{
		hub:            hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		isConnected:    true,
		reconnectDelay: initialReconnectDelay,
		done:           make(chan struct{}),
		wsURL:          wsURL,
	}
}

func (c *Client) reconnect(url string) {
	for {
		select {
		case <-c.done:
			return
		default:
			log.Printf("Attempting to reconnect in %v...", c.reconnectDelay)
			time.Sleep(c.reconnectDelay)

			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				log.Printf("Reconnection failed: %v", err)
				c.reconnectDelay = time.Duration(float64(c.reconnectDelay) * reconnectMultiplier)
				if c.reconnectDelay > maxReconnectDelay {
					c.reconnectDelay = maxReconnectDelay
				}
				continue
			}

			c.mu.Lock()
			c.conn = conn
			c.isConnected = true
			c.reconnectDelay = initialReconnectDelay
			c.mu.Unlock()

			// Re-register the client with the Hub
			c.hub.register <- c

			go c.readPump()
			go c.writePump()
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.mu.Lock()
		c.conn.Close()
		c.isConnected = false
		c.mu.Unlock()

		// Trigger reconnection
		// go c.reconnect(c.wsURL)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().UTC().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().UTC().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}
		c.hub.broadcast <- message
	}
}
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.mu.Lock()
		c.conn.Close()
		c.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.mu.Lock()
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			err := c.conn.SetWriteDeadline(time.Now().UTC().Add(writeWait))
			if err != nil {
				c.mu.Unlock()
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.mu.Unlock()
				return
			}

			if _, err := w.Write(message); err != nil {
				c.mu.Unlock()
				return
			}

			if err := w.Close(); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()

		case <-ticker.C:
			c.mu.Lock()
			if err := c.conn.SetWriteDeadline(time.Now().UTC().Add(writeWait)); err != nil {
				c.mu.Unlock()
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}
