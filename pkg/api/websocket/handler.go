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

// newUpgrader creates a WebSocket upgrader with appropriate CORS settings
func newUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return checkOrigin(r)
		},
	}
}

// checkOrigin validates WebSocket origin against allowed origins
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Same-origin requests don't have an Origin header
		return true
	}

	// Check environment for development mode
	isDev := strings.ToLower(os.Getenv("ENVIRONMENT")) == "development"
	if isDev {
		log.Debugf("Development mode: allowing origin %s", origin)
		return true
	}

	// Production: check against allowed origins
	allowedOrigins := getAllowedOrigins()
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}

	log.Warnf("Rejected WebSocket connection from unauthorized origin: %s", origin)
	return false
}

// getAllowedOrigins reads allowed origins from environment variable or uses defaults
func getAllowedOrigins() []string {
	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		// Default to same-origin in production
		return []string{
			"http://localhost",
			"http://127.0.0.1",
		}
	}

	// Parse comma-separated origins
	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

var upgrader = newUpgrader()

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
			"message":   "Connected to TFDrift-Falco WebSocket v0.9.0",
			"version":   "0.9.0",
			"topics":    []string{"drifts", "events", "state", "drift_result", "discovery_progress", "provider_status", "unmanaged_resource", "all"},
			"features":  []string{"provider_filter", "drift_results", "discovery_progress", "provider_status"},
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

// TestCheckOriginFunc is exported for testing the checkOrigin logic
func TestCheckOriginFunc(r *http.Request) bool {
	return checkOrigin(r)
}

// TestGetAllowedOriginsFunc is exported for testing the getAllowedOrigins logic
func TestGetAllowedOriginsFunc() []string {
	return getAllowedOrigins()
}

// TestNewUploaderFunc is exported for testing the newUpgrader logic
func TestNewUploaderFunc() websocket.Upgrader {
	return newUpgrader()
}
