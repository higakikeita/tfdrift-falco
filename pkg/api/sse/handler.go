package sse

import (
	"net/http"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	log "github.com/sirupsen/logrus"
)

// Handler handles Server-Sent Events connections
type Handler struct {
	broadcaster *broadcaster.Broadcaster
	streams     map[string]*Stream
	mu          sync.RWMutex
}

// NewHandler creates a new SSE handler
func NewHandler(bc *broadcaster.Broadcaster) *Handler {
	return &Handler{
		broadcaster: bc,
		streams:     make(map[string]*Stream),
	}
}

// HandleSSE handles SSE connection requests
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Create new stream
	stream, err := NewStream(w, r, h.broadcaster)
	if err != nil {
		log.Errorf("Failed to create SSE stream: %v", err)
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Register stream
	h.mu.Lock()
	h.streams[stream.ID()] = stream
	streamCount := len(h.streams)
	h.mu.Unlock()

	log.Infof("SSE client connected: %s (total: %d)", stream.ID(), streamCount)

	// Start stream (blocks until client disconnects)
	stream.Start()

	// Cleanup on disconnect
	h.mu.Lock()
	delete(h.streams, stream.ID())
	streamCount = len(h.streams)
	h.mu.Unlock()

	log.Infof("SSE client disconnected: %s (total: %d)", stream.ID(), streamCount)
}

// StreamCount returns the number of active streams
func (h *Handler) StreamCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.streams)
}

// CloseAll closes all active streams
func (h *Handler) CloseAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, stream := range h.streams {
		stream.Stop()
	}

	h.streams = make(map[string]*Stream)
	log.Info("All SSE streams closed")
}
