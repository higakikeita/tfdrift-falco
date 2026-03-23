package websocket

import (
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	log "github.com/sirupsen/logrus"
)

// Hub manages WebSocket client connections
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcaster for receiving events
	broadcaster *broadcaster.Broadcaster

	// Subscription channel for broadcaster events
	eventCh chan broadcaster.Event

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(bc *broadcaster.Broadcaster) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcaster: bc,
		eventCh:     make(chan broadcaster.Event, 100),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Subscribe to broadcaster events
	h.broadcaster.Subscribe(h.eventCh)
	defer h.broadcaster.Unsubscribe(h.eventCh)

	log.Info("WebSocket Hub started")

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Infof("WebSocket client registered: %s (total: %d)", client.id, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.CloseOnce()
				log.Infof("WebSocket client unregistered: %s (total: %d)", client.id, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// FIX: Collect clients to remove under RLock, then remove under Lock
			h.mu.RLock()
			var toRemove []*Client
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full, mark for removal
					toRemove = append(toRemove, client)
				}
			}
			h.mu.RUnlock()

			// Remove stale clients under write lock
			if len(toRemove) > 0 {
				h.mu.Lock()
				for _, client := range toRemove {
					if _, ok := h.clients[client]; ok {
						delete(h.clients, client)
						client.CloseOnce()
					}
				}
				h.mu.Unlock()
			}

		case event := <-h.eventCh:
			// Broadcast event to all connected clients
			h.broadcastEvent(event)
		}
	}
}

// broadcastEvent sends an event to all connected clients
func (h *Hub) broadcastEvent(event broadcaster.Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Check if client is subscribed to this event type
		if client.isSubscribedTo(event.Type) {
			select {
			case client.send <- encodeEvent(event):
			default:
				// Client's send buffer is full, skip
				log.Warnf("Client %s send buffer full, skipping event", client.id)
			}
		}
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
