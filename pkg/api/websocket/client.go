package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a WebSocket client connection
type Client struct {
	// Unique client ID
	id string

	// The hub managing this client
	hub *Hub

	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Subscriptions - set of event types the client is subscribed to
	subscriptions map[string]bool

	// Mutex for subscription operations
	mu sync.RWMutex
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string          `json:"type"`    // "subscribe", "unsubscribe", "query", "ping"
	Topic   string          `json:"topic"`   // "drifts", "events", "state", "all"
	Payload json.RawMessage `json:"payload"` // Additional data
}

// WSResponse represents a WebSocket response
type WSResponse struct {
	Type      string      `json:"type"`      // "subscribed", "unsubscribed", "data", "error", "pong"
	Topic     string      `json:"topic"`     // Topic name
	Data      interface{} `json:"data"`      // Response data
	Timestamp string      `json:"timestamp"` // ISO 8601 timestamp
	Error     string      `json:"error,omitempty"`
}

// newClient creates a new WebSocket client
func newClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		id:            uuid.New().String(),
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warnf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Warnf("Invalid WebSocket message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	switch msg.Type {
	case "subscribe":
		c.subscribe(msg.Topic)
	case "unsubscribe":
		c.unsubscribe(msg.Topic)
	case "ping":
		c.sendPong()
	case "query":
		// Handle query requests (future enhancement)
		c.sendError("Query not implemented yet")
	default:
		c.sendError("Unknown message type: " + msg.Type)
	}
}

// subscribe subscribes the client to a topic
func (c *Client) subscribe(topic string) {
	c.mu.Lock()
	c.subscriptions[topic] = true
	c.mu.Unlock()

	log.Infof("Client %s subscribed to topic: %s", c.id, topic)
	c.sendResponse(WSResponse{
		Type:      "subscribed",
		Topic:     topic,
		Data:      map[string]string{"status": "success"},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// unsubscribe unsubscribes the client from a topic
func (c *Client) unsubscribe(topic string) {
	c.mu.Lock()
	delete(c.subscriptions, topic)
	c.mu.Unlock()

	log.Infof("Client %s unsubscribed from topic: %s", c.id, topic)
	c.sendResponse(WSResponse{
		Type:      "unsubscribed",
		Topic:     topic,
		Data:      map[string]string{"status": "success"},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// isSubscribedTo checks if the client is subscribed to a topic
func (c *Client) isSubscribedTo(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check for "all" subscription
	if c.subscriptions["all"] {
		return true
	}

	return c.subscriptions[topic]
}

// sendResponse sends a response to the client
func (c *Client) sendResponse(resp WSResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("Failed to marshal response: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		log.Warnf("Client %s send buffer full", c.id)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	c.sendResponse(WSResponse{
		Type:      "error",
		Error:     errMsg,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// sendPong sends a pong response to the client
func (c *Client) sendPong() {
	c.sendResponse(WSResponse{
		Type:      "pong",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
