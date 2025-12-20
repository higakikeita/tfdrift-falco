package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	log "github.com/sirupsen/logrus"
)

// Stream represents an SSE stream for a single client
type Stream struct {
	id          string
	w           http.ResponseWriter
	flusher     http.Flusher
	ctx         context.Context
	cancel      context.CancelFunc
	eventCh     chan broadcaster.Event
	broadcaster *broadcaster.Broadcaster
	mu          sync.Mutex
}

// NewStream creates a new SSE stream
func NewStream(w http.ResponseWriter, r *http.Request, bc *broadcaster.Broadcaster) (*Stream, error) {
	// Check if the response writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	ctx, cancel := context.WithCancel(r.Context())

	stream := &Stream{
		id:          uuid.New().String(),
		w:           w,
		flusher:     flusher,
		ctx:         ctx,
		cancel:      cancel,
		eventCh:     make(chan broadcaster.Event, 100),
		broadcaster: bc,
	}

	return stream, nil
}

// Start starts the SSE stream
func (s *Stream) Start() {
	// Subscribe to broadcaster events
	s.broadcaster.Subscribe(s.eventCh)
	defer s.broadcaster.Unsubscribe(s.eventCh)

	log.Infof("SSE stream started: %s", s.id)

	// Send initial connection event
	s.sendEvent("connected", map[string]interface{}{
		"stream_id": s.id,
		"message":   "Connected to TFDrift-Falco SSE stream",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Keep-alive ticker (send comment every 30 seconds)
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			log.Infof("SSE stream stopped: %s", s.id)
			return

		case event := <-s.eventCh:
			// Forward broadcaster event to SSE client
			s.sendBroadcasterEvent(event)

		case <-keepAliveTicker.C:
			// Send keep-alive comment
			s.sendComment("keep-alive")
		}
	}
}

// sendEvent sends an SSE event with custom event type
func (s *Stream) sendEvent(eventType string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Failed to marshal SSE event data: %v", err)
		return
	}

	// Write SSE format:
	// event: <eventType>
	// data: <jsonData>
	// <blank line>
	fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, jsonData)
	s.flusher.Flush()
}

// sendBroadcasterEvent converts and sends a broadcaster event as SSE
func (s *Stream) sendBroadcasterEvent(event broadcaster.Event) {
	s.sendEvent(event.Type, event.Payload)
}

// sendComment sends an SSE comment (for keep-alive)
func (s *Stream) sendComment(comment string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintf(s.w, ": %s\n\n", comment)
	s.flusher.Flush()
}

// Stop stops the SSE stream
func (s *Stream) Stop() {
	s.cancel()
}

// ID returns the stream ID
func (s *Stream) ID() string {
	return s.id
}
