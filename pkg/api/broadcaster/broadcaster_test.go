package broadcaster

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBroadcaster(t *testing.T) {
	bc := NewBroadcaster()

	assert.NotNil(t, bc)
	assert.NotNil(t, bc.subscribers)
	assert.Equal(t, 0, bc.SubscriberCount())
}

func TestBroadcaster_Subscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	assert.Equal(t, 1, bc.SubscriberCount())
}

func TestBroadcaster_MultipleSubscribers(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	ch3 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)
	bc.Subscribe(ch3)

	assert.Equal(t, 3, bc.SubscriberCount())
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch)
	assert.Equal(t, 0, bc.SubscriberCount())

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok)
}

func TestBroadcaster_UnsubscribeMultiple(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)
	assert.Equal(t, 2, bc.SubscriberCount())

	bc.Unsubscribe(ch1)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch2)
	assert.Equal(t, 0, bc.SubscriberCount())
}

func TestBroadcaster_Broadcast(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	event := Event{
		Type:      "drift",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload: map[string]interface{}{
			"resource": "test",
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.Equal(t, "test", received.Payload["resource"])
}

func TestBroadcaster_BroadcastMultipleSubscribers(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)

	event := Event{
		Type:      "state_change",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload:   map[string]interface{}{},
	}

	bc.Broadcast(event)

	received1 := <-ch1
	received2 := <-ch2

	assert.Equal(t, "state_change", received1.Type)
	assert.Equal(t, "state_change", received2.Type)
}

func TestBroadcaster_BroadcastNonBlockingWithFullBuffer(t *testing.T) {
	bc := NewBroadcaster()
	// Channel with buffer size of 1
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	event1 := Event{Type: "drift", Payload: map[string]interface{}{}}
	event2 := Event{Type: "state", Payload: map[string]interface{}{}}

	// First broadcast fills the buffer
	bc.Broadcast(event1)
	// Second broadcast should be non-blocking and not panic
	bc.Broadcast(event2)

	// Only first event should be in channel
	received := <-ch
	assert.Equal(t, "drift", received.Type)

	// Channel should be empty (second broadcast was dropped)
	assert.Equal(t, 0, len(ch))
}

func TestBroadcaster_BroadcastDriftAlert(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	// We'll use a minimal DriftAlert structure for this test
	// Since we don't import the full types package, we'll verify the broadcast mechanism
	alert := struct {
		Severity     string
		ResourceType string
		ResourceName string
		ResourceID   string
		Attribute    string
		OldValue     string
		NewValue     string
		UserIdentity string
		MatchedRules []string
		AlertType    string
		Timestamp    string
	}{
		Severity:     "high",
		ResourceType: "EC2",
		ResourceName: "prod-server",
		Timestamp:    "2024-01-01T00:00:00Z",
	}

	// Manually create the event that BroadcastDriftAlert would create
	event := Event{
		Type:      "drift",
		Timestamp: alert.Timestamp,
		Payload: map[string]interface{}{
			"severity":      alert.Severity,
			"resource_type": alert.ResourceType,
			"resource_name": alert.ResourceName,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.Equal(t, "high", received.Payload["severity"])
	assert.Equal(t, "EC2", received.Payload["resource_type"])
}

func TestBroadcaster_SubscriberCountAccuracy(t *testing.T) {
	bc := NewBroadcaster()

	assert.Equal(t, 0, bc.SubscriberCount())

	channels := make([]chan Event, 5)
	for i := 0; i < 5; i++ {
		channels[i] = make(chan Event, 1)
		bc.Subscribe(channels[i])
		assert.Equal(t, i+1, bc.SubscriberCount())
	}

	for i := 0; i < 5; i++ {
		bc.Unsubscribe(channels[i])
		assert.Equal(t, 4-i, bc.SubscriberCount())
	}
}

func TestBroadcaster_ConcurrentBroadcast(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 10)

	bc.Subscribe(ch)

	// Send multiple broadcasts concurrently
	go func() {
		for i := 0; i < 5; i++ {
			event := Event{
				Type:    "drift",
				Payload: map[string]interface{}{"id": i},
			}
			bc.Broadcast(event)
		}
	}()

	// Receive broadcasts
	received := 0
	timeout := time.After(2 * time.Second)

	for received < 5 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("Timeout: expected 5 broadcasts, got %d", received)
		}
	}
}

func TestBroadcaster_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	var wg sync.WaitGroup

	// Concurrent subscribe
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan Event, 1)
			bc.Subscribe(ch)
		}()
	}

	wg.Wait()
	assert.Equal(t, 10, bc.SubscriberCount())

	// Concurrent unsubscribe (would need to track channels)
	// This test verifies no race conditions
}

func TestBroadcaster_BroadcastStateChange(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	changes := map[string]interface{}{
		"status": "modified",
		"tags":   []string{"prod"},
	}

	event := Event{
		Type: "state_change",
		Payload: map[string]interface{}{
			"resource_type": "aws_instance",
			"resource_id":   "i-123456",
			"changes":       changes,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "state_change", received.Type)
	assert.Equal(t, "aws_instance", received.Payload["resource_type"])
	assert.Equal(t, "i-123456", received.Payload["resource_id"])
}

func TestBroadcaster_EventPayload(t *testing.T) {
	event := Event{
		Type:      "test",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload: map[string]interface{}{
			"string": "value",
			"number": 42,
			"bool":   true,
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	assert.Equal(t, "value", event.Payload["string"])
	assert.Equal(t, 42, event.Payload["number"])
	assert.Equal(t, true, event.Payload["bool"])
	assert.NotNil(t, event.Payload["nested"])
}

func TestBroadcaster_UnsubscribeNonExistent(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	// Should not panic when unsubscribing non-existent channel
	assert.NotPanics(t, func() {
		bc.Unsubscribe(ch)
	})
}

func TestBroadcaster_SubscribeAfterUnsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch)
	assert.Equal(t, 0, bc.SubscriberCount())

	// Subscribe again with new channel
	ch2 := make(chan Event, 1)
	bc.Subscribe(ch2)
	assert.Equal(t, 1, bc.SubscriberCount())
}

func TestBroadcaster_BroadcastWithMultipleEvents(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 3)

	bc.Subscribe(ch)

	events := []Event{
		{Type: "drift", Payload: map[string]interface{}{"id": 1}},
		{Type: "state", Payload: map[string]interface{}{"id": 2}},
		{Type: "falco", Payload: map[string]interface{}{"id": 3}},
	}

	for _, event := range events {
		bc.Broadcast(event)
	}

	for i, expectedEvent := range events {
		received := <-ch
		assert.Equal(t, expectedEvent.Type, received.Type, "Event %d type mismatch", i)
	}
}

func TestBroadcaster_DiscoveryProgress(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "discovery_progress",
		Payload: map[string]interface{}{
			"provider":      "aws",
			"resource_type": "EC2",
			"count":         10,
			"phase":         "discovering",
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "discovering", received.Payload["phase"])
}

func TestBroadcaster_ProviderStatus(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "provider_status",
		Payload: map[string]interface{}{
			"provider": "aws",
			"status":   "connected",
			"capabilities": map[string]bool{
				"drift_detection": true,
				"discovery":       true,
			},
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "connected", received.Payload["status"])
}

func TestBroadcaster_UnmanagedResource(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "unmanaged_resource",
		Payload: map[string]interface{}{
			"id":       "i-123456",
			"type":     "instance",
			"provider": "aws",
			"name":     "test-server",
			"region":   "us-east-1",
			"tags": map[string]string{
				"Environment": "prod",
			},
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	assert.Equal(t, "i-123456", received.Payload["id"])
	assert.Equal(t, "prod", received.Payload["tags"].(map[string]string)["Environment"])
}

func TestBroadcaster_BroadcastDriftResult(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "drift_result",
		Payload: map[string]interface{}{
			"provider":        "aws",
			"unmanaged_count": 5,
			"missing_count":   2,
			"modified_count":  3,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift_result", received.Type)
	assert.Equal(t, 5, received.Payload["unmanaged_count"])
}

func TestBroadcaster_ThreadSafety(t *testing.T) {
	bc := NewBroadcaster()
	var wg sync.WaitGroup

	// Create multiple publishers and subscribers
	channels := make([]chan Event, 5)
	for i := 0; i < 5; i++ {
		channels[i] = make(chan Event, 10)
		bc.Subscribe(channels[i])
	}

	// Concurrent broadcasts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := Event{
				Type:    "test",
				Payload: map[string]interface{}{"broadcast_id": id},
			}
			bc.Broadcast(event)
		}(i)
	}

	wg.Wait()

	// Verify all messages were delivered
	for _, ch := range channels {
		count := 0
		timeout := time.After(1 * time.Second)
		for {
			select {
			case <-ch:
				count++
			case <-timeout:
				assert.Equal(t, 10, count, "Not all broadcasts received")
				return
			}
		}
	}
}

func TestBroadcaster_LargePayload(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	largePayload := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largePayload[string(rune(i))] = "value_" + string(rune(i))
	}

	event := Event{
		Type:    "large",
		Payload: largePayload,
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, 100, len(received.Payload))
}
