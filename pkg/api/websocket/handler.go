package websocket

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	log "github.com/sirupsen/logrus"
)

func getAllowedOrigins() []string {
	originsStr := os.Getenv("WEBSOCKET_ALLOWED_ORIGINS")
	if originsStr == "" {
		// Default to * (all origins) for development
		return []string{"*"}
	}
	// Parse comma-separated list of allowed origins
	var origins []string
	for _, origin := range strings.Split(originsStr, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func checkOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Allow requests without Origin header (same-origin)
		}
		allowedOrigins := getAllowedOrigins()
		return checkOriginAllowed(origin, allowedOrigins)
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(bc *broadcaster.Broadcaster) *Handler {
	hub := NewHub(bc)

	// Start the hub in a goroutine
	go hub.Run()

	return &Handler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	client := newClient(h.hub, conn)
	h.hub.register <- client

	// Send welcome message
	client.sendResponse(WSResponse{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id": client.id,
			"message":   "Connected to TFDrift-Falco WebSocket",
			"topics":    []string{"drifts", "events", "state", "all"},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})

	// Start client pumps
	go client.writePump()
	go client.readPump()
}

// GetHub returns the WebSocket hub (for metrics/monitoring)
func (h *Handler) GetHub() *Hub {
	return h.hub
}

// encodeEvent encodes a broadcaster event as JSON for WebSocket transmission
func encodeEvent(event broadcaster.Event) []byte {
	resp := WSResponse{
		Type:      "data",
		Topic:     event.Type,
		Data:      event.Payload,
		Timestamp: event.Timestamp,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("Failed to encode event: %v", err)
		return []byte{}
	}

	return data
}
