package broadcaster

import (
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Event represents a broadcast event
type Event struct {
	Type      string                 `json:"type"`      // "drift", "falco", "state_change"
	Timestamp string                 `json:"timestamp"` // ISO 8601 timestamp
	Payload   map[string]interface{} `json:"payload"`   // Event-specific data
}

// Broadcaster manages event subscriptions and broadcasting
type Broadcaster struct {
	subscribers map[chan Event]bool
	mu          sync.RWMutex
}

// NewBroadcaster creates a new event broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan Event]bool),
	}
}

// Subscribe adds a new subscriber channel
func (b *Broadcaster) Subscribe(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[ch] = true
}

// Unsubscribe removes a subscriber channel
func (b *Broadcaster) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, ch)
	close(ch)
}

// Broadcast sends an event to all subscribers (non-blocking)
func (b *Broadcaster) Broadcast(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Non-blocking send, drop if channel full
		}
	}
}

// BroadcastDriftAlert broadcasts a drift alert event
func (b *Broadcaster) BroadcastDriftAlert(alert types.DriftAlert) {
	event := Event{
		Type:      "drift",
		Timestamp: alert.Timestamp,
		Payload: map[string]interface{}{
			"severity":      alert.Severity,
			"resource_type": alert.ResourceType,
			"resource_name": alert.ResourceName,
			"resource_id":   alert.ResourceID,
			"attribute":     alert.Attribute,
			"old_value":     alert.OldValue,
			"new_value":     alert.NewValue,
			"user_identity": alert.UserIdentity,
			"matched_rules": alert.MatchedRules,
			"alert_type":    alert.AlertType,
		},
	}
	b.Broadcast(event)
}

// BroadcastFalcoEvent broadcasts a Falco event
func (b *Broadcaster) BroadcastFalcoEvent(event types.Event) {
	broadcastEvent := Event{
		Type:      "falco",
		Timestamp: "", // Will be set by caller if needed
		Payload: map[string]interface{}{
			"provider":      event.Provider,
			"event_name":    event.EventName,
			"resource_type": event.ResourceType,
			"resource_id":   event.ResourceID,
			"user_identity": event.UserIdentity,
			"changes":       event.Changes,
			"region":        event.Region,
			"project_id":    event.ProjectID,
			"service_name":  event.ServiceName,
		},
	}
	b.Broadcast(broadcastEvent)
}

// BroadcastStateChange broadcasts a state change event
func (b *Broadcaster) BroadcastStateChange(resourceType, resourceID string, changes map[string]interface{}) {
	event := Event{
		Type:      "state_change",
		Timestamp: "",
		Payload: map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"changes":       changes,
		},
	}
	b.Broadcast(event)
}

// SubscriberCount returns the number of active subscribers
func (b *Broadcaster) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
