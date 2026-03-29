package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/stretchr/testify/assert"
)

// ==================== Hub Tests ====================

func TestNewHub(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_ClientCount(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_BroadcasterSubscription(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	// Before running hub, eventCh should be created
	assert.NotNil(t, hub.eventCh)
	assert.Equal(t, 0, bc.SubscriberCount())
}

// ==================== Client Tests ====================

func TestNewClient(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	// We'll create a mock connection since we can't use real websocket in unit test
	client := &Client{
		id:            "test-client-1",
		hub:           hub,
		conn:          nil,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	assert.NotEmpty(t, client.id)
	assert.Equal(t, hub, client.hub)
	assert.NotNil(t, client.subscriptions)
}

func TestClient_Subscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")

	assert.True(t, client.isSubscribedTo("drifts"))
}

func TestClient_MultipleSubscriptions(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")
	client.subscribe("events")
	client.subscribe("state")

	assert.True(t, client.isSubscribedTo("drifts"))
	assert.True(t, client.isSubscribedTo("events"))
	assert.True(t, client.isSubscribedTo("state"))
}

func TestClient_Unsubscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")
	assert.True(t, client.isSubscribedTo("drifts"))

	client.unsubscribe("drifts")
	assert.False(t, client.isSubscribedTo("drifts"))
}

func TestClient_IsSubscribedTo_All(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("all")

	// Should return true for any topic when subscribed to "all"
	assert.True(t, client.isSubscribedTo("drifts"))
	assert.True(t, client.isSubscribedTo("events"))
	assert.True(t, client.isSubscribedTo("any_topic"))
}

func TestClient_ProviderFilter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.setProviderFilter("aws")

	assert.Equal(t, "aws", client.providerFilter)
}

func TestClient_MatchesProviderFilter_NoFilter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	payload := map[string]interface{}{
		"provider": "gcp",
	}

	// With no filter, all providers pass
	assert.True(t, client.matchesProviderFilter(payload))
}

func TestClient_MatchesProviderFilter_WithFilter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.setProviderFilter("aws")

	awsPayload := map[string]interface{}{
		"provider": "aws",
	}

	gcpPayload := map[string]interface{}{
		"provider": "gcp",
	}

	assert.True(t, client.matchesProviderFilter(awsPayload))
	assert.False(t, client.matchesProviderFilter(gcpPayload))
}

func TestClient_MatchesProviderFilter_NoProviderInPayload(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.setProviderFilter("aws")

	payload := map[string]interface{}{
		"data": "value",
	}

	// If no provider in payload, pass through
	assert.True(t, client.matchesProviderFilter(payload))
}

// ==================== Message Handling Tests ====================

func TestClient_HandleMessage_Subscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"subscribe","topic":"drifts"}`)
	client.handleMessage(message)

	assert.True(t, client.isSubscribedTo("drifts"))
}

func TestClient_HandleMessage_Unsubscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")
	message := []byte(`{"type":"unsubscribe","topic":"drifts"}`)
	client.handleMessage(message)

	assert.False(t, client.isSubscribedTo("drifts"))
}

func TestClient_HandleMessage_Filter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"filter","provider":"aws"}`)
	client.handleMessage(message)

	assert.Equal(t, "aws", client.providerFilter)
}

func TestClient_HandleMessage_Ping(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"ping"}`)
	client.handleMessage(message)

	// Should have sent a pong response
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "pong", wsResp.Type)
	default:
		t.Fatal("Expected pong response")
	}
}

func TestClient_HandleMessage_InvalidJSON(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`invalid json`)
	client.handleMessage(message)

	// Should have sent an error response
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "error", wsResp.Type)
	default:
		t.Fatal("Expected error response")
	}
}

func TestClient_HandleMessage_UnknownType(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"unknown"}`)
	client.handleMessage(message)

	// Should have sent an error response
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "error", wsResp.Type)
	default:
		t.Fatal("Expected error response")
	}
}

// ==================== Response Tests ====================

func TestClient_SendResponse(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	resp := WSResponse{
		Type:      "subscribed",
		Topic:     "drifts",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	client.sendResponse(resp)

	// Should have sent response
	select {
	case data := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(data, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "subscribed", wsResp.Type)
	default:
		t.Fatal("Expected response")
	}
}

func TestClient_SendError(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.sendError("Test error")

	// Should have sent error response
	select {
	case data := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(data, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "error", wsResp.Type)
		assert.Equal(t, "Test error", wsResp.Error)
	default:
		t.Fatal("Expected error response")
	}
}

func TestClient_SendPong(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.sendPong()

	// Should have sent pong response
	select {
	case data := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(data, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "pong", wsResp.Type)
	default:
		t.Fatal("Expected pong response")
	}
}

// ==================== WSMessage Tests ====================

func TestWSMessage_Marshaling(t *testing.T) {
	msg := WSMessage{
		Type:     "subscribe",
		Topic:    "drifts",
		Provider: "aws",
		Payload:  json.RawMessage(`{"key":"value"}`),
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var unmarshaledMsg WSMessage
	err = json.Unmarshal(data, &unmarshaledMsg)
	assert.NoError(t, err)

	assert.Equal(t, "subscribe", unmarshaledMsg.Type)
	assert.Equal(t, "drifts", unmarshaledMsg.Topic)
}

func TestWSResponse_Marshaling(t *testing.T) {
	resp := WSResponse{
		Type:      "subscribed",
		Topic:     "drifts",
		Data:      map[string]string{"status": "success"},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var unmarshaledResp WSResponse
	err = json.Unmarshal(data, &unmarshaledResp)
	assert.NoError(t, err)

	assert.Equal(t, "subscribed", unmarshaledResp.Type)
	assert.Equal(t, "drifts", unmarshaledResp.Topic)
}

// ==================== Hub Broadcast Tests ====================

func TestHub_BroadcastEvent_FilteredDelivery(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	// Create test clients
	client1 := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"drifts": true},
	}

	client2 := &Client{
		id:            "client-2",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"events": true},
	}

	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	// Broadcast drift event
	event := broadcaster.Event{
		Type:    "drifts",
		Payload: map[string]interface{}{},
	}

	hub.broadcastEvent(event)

	// Client1 should receive (subscribed to drifts)
	assert.Equal(t, 1, len(client1.send))
	// Client2 should not receive (not subscribed to drifts)
	assert.Equal(t, 0, len(client2.send))
}

func TestHub_BroadcastEvent_ProviderFilter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  map[string]bool{"drift": true},
		providerFilter: "aws",
	}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// Broadcast AWS event
	awsEvent := broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"provider": "aws",
		},
	}

	hub.broadcastEvent(awsEvent)
	assert.Equal(t, 1, len(client.send))

	// Broadcast GCP event
	gcpEvent := broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"provider": "gcp",
		},
	}

	hub.broadcastEvent(gcpEvent)
	// Should still be 1 (GCP event filtered out)
	assert.Equal(t, 1, len(client.send))
}

// ==================== Integration Tests ====================

func TestClient_ThreadSafety_Subscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	done := make(chan struct{})

	// Concurrent subscribe
	for i := 0; i < 10; i++ {
		go func(topic string) {
			client.subscribe(topic)
			done <- struct{}{}
		}("topic" + string(rune(i)))
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 subscriptions
	client.mu.RLock()
	count := len(client.subscriptions)
	client.mu.RUnlock()
	assert.Equal(t, 10, count)
}

func TestWSMessage_ComplexPayload(t *testing.T) {
	payload := map[string]interface{}{
		"filters": map[string]string{
			"status": "active",
		},
		"limit": 100,
	}

	payloadJSON, _ := json.Marshal(payload)

	msg := WSMessage{
		Type:     "query",
		Topic:    "drifts",
		Payload:  payloadJSON,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var unmarshaledMsg WSMessage
	err = json.Unmarshal(data, &unmarshaledMsg)
	assert.NoError(t, err)

	assert.Equal(t, "query", unmarshaledMsg.Type)
	assert.NotNil(t, unmarshaledMsg.Payload)
}

func TestEncodeEvent(t *testing.T) {
	event := broadcaster.Event{
		Type:      "drift",
		Timestamp: time.Now().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"severity": "high",
		},
	}

	data := encodeEvent(event)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data)

	var resp WSResponse
	err := json.Unmarshal(data, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "data", resp.Type)
	assert.Equal(t, "drift", resp.Topic)
}

func TestClient_SendBufferFullHandling(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	// Client with small buffer
	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 1),
		subscriptions: make(map[string]bool),
	}

	// Fill the buffer
	client.send <- []byte("message1")

	// Try to send another (should not panic)
	resp := WSResponse{
		Type: "data",
	}

	assert.NotPanics(t, func() {
		client.sendResponse(resp)
	})
}

func TestClient_SubscriptionState(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	topics := []string{"drifts", "events", "state", "drift_result"}

	for _, topic := range topics {
		client.subscribe(topic)
	}

	for _, topic := range topics {
		assert.True(t, client.isSubscribedTo(topic))
	}

	// Unsubscribe from one
	client.unsubscribe("state")

	assert.True(t, client.isSubscribedTo("drifts"))
	assert.True(t, client.isSubscribedTo("events"))
	assert.False(t, client.isSubscribedTo("state"))
	assert.True(t, client.isSubscribedTo("drift_result"))
}
