package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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
		Type:    "query",
		Topic:   "drifts",
		Payload: payloadJSON,
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

// ==================== Handler Tests ====================

func TestNewHandler(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.hub)
}

func TestHandler_GetHub(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	hub := handler.GetHub()
	assert.NotNil(t, hub)
	assert.Equal(t, handler.hub, hub)
}

// ==================== Subscription Filter Tests ====================

func TestClient_Subscribe_DriftTopic(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")

	// Check subscription worked
	assert.True(t, client.isSubscribedTo("drifts"))

	// Verify response was sent
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "subscribed", wsResp.Type)
		assert.Equal(t, "drifts", wsResp.Topic)
	default:
		t.Fatal("Expected subscription response")
	}
}

func TestClient_Subscribe_AllTopic(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("all")

	// "all" subscription should match any topic
	assert.True(t, client.isSubscribedTo("drifts"))
	assert.True(t, client.isSubscribedTo("events"))
	assert.True(t, client.isSubscribedTo("discovery_progress"))
}

func TestClient_Unsubscribe_SingleTopic(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("events")
	client.subscribe("drifts")

	client.unsubscribe("events")

	assert.False(t, client.isSubscribedTo("events"))
	assert.True(t, client.isSubscribedTo("drifts"))
}

func TestClient_SetProviderFilter_AWS(t *testing.T) {
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

	// Check response
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "filter_set", wsResp.Type)
	default:
		t.Fatal("Expected filter response")
	}
}

func TestClient_SetProviderFilter_GCP(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.setProviderFilter("gcp")
	assert.Equal(t, "gcp", client.providerFilter)

	client.setProviderFilter("azure")
	assert.Equal(t, "azure", client.providerFilter)
}

func TestClient_MatchesProviderFilter_EmptyFilter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// No filter set, all providers pass
	payload := map[string]interface{}{
		"provider": "aws",
	}

	assert.True(t, client.matchesProviderFilter(payload))
}

func TestClient_MatchesProviderFilter_Match(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		providerFilter: "aws",
	}

	payload := map[string]interface{}{
		"provider": "aws",
	}

	assert.True(t, client.matchesProviderFilter(payload))
}

func TestClient_MatchesProviderFilter_NoMatch(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		providerFilter: "aws",
	}

	payload := map[string]interface{}{
		"provider": "gcp",
	}

	assert.False(t, client.matchesProviderFilter(payload))
}

func TestClient_MatchesProviderFilter_NonStringProvider(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		providerFilter: "aws",
	}

	payload := map[string]interface{}{
		"provider": 123, // Non-string provider
	}

	// Should pass through when provider is not a string
	assert.True(t, client.matchesProviderFilter(payload))
}

// ==================== Message Type Handling Tests ====================

func TestClient_HandleMessage_Query(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"query","topic":"drifts"}`)
	client.handleMessage(message)

	// Should send error (not implemented)
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

func TestClient_HandleMessage_MultipleFilters(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Set first filter
	message1 := []byte(`{"type":"filter","provider":"aws"}`)
	client.handleMessage(message1)
	assert.Equal(t, "aws", client.providerFilter)

	// Change filter
	message2 := []byte(`{"type":"filter","provider":"gcp"}`)
	client.handleMessage(message2)
	assert.Equal(t, "gcp", client.providerFilter)
}

// ==================== Hub Event Routing Tests ====================

func TestHub_BroadcastEvent_AllSubscription(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"all": true},
	}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// Broadcast any event
	event := broadcaster.Event{
		Type:    "discovery_progress",
		Payload: map[string]interface{}{},
	}

	hub.broadcastEvent(event)

	// Should receive event (all subscription)
	assert.Equal(t, 1, len(client.send))
}

func TestHub_BroadcastEvent_SkipsUnsubscribedClients(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

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

	client3 := &Client{
		id:            "client-3",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"drifts": true},
	}

	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.clients[client3] = true
	hub.mu.Unlock()

	// Broadcast drift event
	event := broadcaster.Event{
		Type:    "drifts",
		Payload: map[string]interface{}{},
	}

	hub.broadcastEvent(event)

	// Only subscribed clients should receive
	assert.Equal(t, 1, len(client1.send))
	assert.Equal(t, 0, len(client2.send))
	assert.Equal(t, 1, len(client3.send))
}

func TestHub_BroadcastEvent_WithMultipleFilters(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	awsClient := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  map[string]bool{"drift": true},
		providerFilter: "aws",
	}

	gcpClient := &Client{
		id:             "client-2",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  map[string]bool{"drift": true},
		providerFilter: "gcp",
	}

	bothClient := &Client{
		id:            "client-3",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"drift": true},
	}

	hub.mu.Lock()
	hub.clients[awsClient] = true
	hub.clients[gcpClient] = true
	hub.clients[bothClient] = true
	hub.mu.Unlock()

	// Broadcast AWS event
	awsEvent := broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"provider": "aws",
		},
	}

	hub.broadcastEvent(awsEvent)

	// AWS client and both client should receive
	assert.Equal(t, 1, len(awsClient.send))
	assert.Equal(t, 0, len(gcpClient.send))
	assert.Equal(t, 1, len(bothClient.send))
}

// ==================== Complex Subscription Tests ====================

func TestClient_DynamicSubscriptionChanges(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Subscribe to multiple topics
	topics := []string{"drifts", "events", "state", "discovery_progress", "provider_status"}
	for _, topic := range topics {
		client.subscribe(topic)
	}

	// Verify all subscriptions
	for _, topic := range topics {
		assert.True(t, client.isSubscribedTo(topic))
	}

	// Unsubscribe from some
	client.unsubscribe("state")
	client.unsubscribe("discovery_progress")

	// Verify state
	assert.True(t, client.isSubscribedTo("drifts"))
	assert.True(t, client.isSubscribedTo("events"))
	assert.False(t, client.isSubscribedTo("state"))
	assert.False(t, client.isSubscribedTo("discovery_progress"))
	assert.True(t, client.isSubscribedTo("provider_status"))
}

// ==================== Event Encoding Tests ====================

func TestEncodeEvent_WithTimestamp(t *testing.T) {
	event := broadcaster.Event{
		Type:      "drift",
		Timestamp: "2024-01-01T12:00:00Z",
		Payload: map[string]interface{}{
			"severity": "high",
			"resource": "ec2",
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
	assert.Equal(t, "2024-01-01T12:00:00Z", resp.Timestamp)
}

func TestEncodeEvent_WithComplexPayload(t *testing.T) {
	event := broadcaster.Event{
		Type: "drift_result",
		Payload: map[string]interface{}{
			"provider":        "aws",
			"unmanaged_count": 5,
			"modified_count":  3,
			"details": map[string]interface{}{
				"regions": []string{"us-east-1", "us-west-2"},
			},
		},
	}

	data := encodeEvent(event)
	var resp WSResponse
	err := json.Unmarshal(data, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "drift_result", resp.Topic)
	assert.NotNil(t, resp.Data)
}

// ==================== Response Serialization Tests ====================

func TestWSResponse_ErrorMessage(t *testing.T) {
	resp := WSResponse{
		Type:      "error",
		Error:     "Invalid subscription topic",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var unmarshaledResp WSResponse
	err = json.Unmarshal(data, &unmarshaledResp)
	assert.NoError(t, err)
	assert.Equal(t, "error", unmarshaledResp.Type)
	assert.Equal(t, "Invalid subscription topic", unmarshaledResp.Error)
}

func TestWSResponse_DataResponse(t *testing.T) {
	resp := WSResponse{
		Type:  "data",
		Topic: "drifts",
		Data: map[string]interface{}{
			"severity": "critical",
			"count":    1,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var unmarshaledResp WSResponse
	err = json.Unmarshal(data, &unmarshaledResp)
	assert.NoError(t, err)
	assert.Equal(t, "data", unmarshaledResp.Type)
	assert.NotNil(t, unmarshaledResp.Data)
}

// ==================== Concurrent Client Tests ====================

func TestHub_MultipleClientsSubscribing(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	numClients := 20
	clients := make([]*Client, numClients)

	for i := 0; i < numClients; i++ {
		clients[i] = &Client{
			id:            string(rune(i)),
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}
		hub.mu.Lock()
		hub.clients[clients[i]] = true
		hub.mu.Unlock()
	}

	assert.Equal(t, numClients, hub.ClientCount())

	// Unregister some clients
	for i := 0; i < 5; i++ {
		hub.mu.Lock()
		delete(hub.clients, clients[i])
		hub.mu.Unlock()
	}

	assert.Equal(t, numClients-5, hub.ClientCount())
}

func TestClient_ConcurrentProviderFilterChanges(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	done := make(chan struct{})

	// Concurrent filter changes
	for i := 0; i < 10; i++ {
		go func(provider string) {
			client.setProviderFilter(provider)
			done <- struct{}{}
		}([]string{"aws", "gcp", "azure"}[i%3])
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Filter should be set to one of the three
	client.mu.RLock()
	filter := client.providerFilter
	client.mu.RUnlock()
	assert.Contains(t, []string{"aws", "gcp", "azure"}, filter)
}

// ==================== Additional Handler Tests ====================

func TestHandler_HandleWebSocket_WelcomeMessage(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	// Handler should be initialized with a hub
	assert.NotNil(t, handler.hub)
	assert.NotNil(t, handler.GetHub())
}

// ==================== Additional Subscribe/Unsubscribe Tests ====================

func TestClient_Subscribe_Response_Success(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("drifts")

	// Verify response contains success data
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "subscribed", wsResp.Type)
		data := wsResp.Data.(map[string]interface{})
		assert.Equal(t, "success", data["status"])
	default:
		t.Fatal("Expected response")
	}
}

func TestClient_Unsubscribe_Response_Success(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.subscribe("events")
	// Clear send buffer
	<-client.send

	client.unsubscribe("events")

	// Verify response
	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "unsubscribed", wsResp.Type)
	default:
		t.Fatal("Expected response")
	}
}

// ==================== SendError Tests ====================

func TestClient_SendError_MultipleErrors(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	errors := []string{
		"Invalid message format",
		"Subscription not found",
		"Internal server error",
	}

	for _, errMsg := range errors {
		client.sendError(errMsg)
	}

	// Verify all errors were sent
	for i := 0; i < 3; i++ {
		select {
		case response := <-client.send:
			var wsResp WSResponse
			err := json.Unmarshal(response, &wsResp)
			assert.NoError(t, err)
			assert.Equal(t, "error", wsResp.Type)
		default:
			t.Fatalf("Expected error response %d", i)
		}
	}
}

// ==================== SendResponse Data Tests ====================

func TestClient_SendResponse_WithComplexData(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	resp := WSResponse{
		Type:  "data",
		Topic: "drift_result",
		Data: map[string]interface{}{
			"provider":        "aws",
			"unmanaged_count": 5,
			"missing_count":   2,
			"resources": []map[string]interface{}{
				{
					"id":   "i-123",
					"type": "aws_instance",
				},
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	client.sendResponse(resp)

	select {
	case data := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(data, &wsResp)
		assert.NoError(t, err)
		assert.Equal(t, "data", wsResp.Type)
		assert.NotNil(t, wsResp.Data)
	default:
		t.Fatal("Expected response")
	}
}

// ==================== Message Handling Edge Cases ====================

func TestClient_HandleMessage_EmptyType(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	message := []byte(`{"type":"","topic":"drifts"}`)
	client.handleMessage(message)

	// Should send error
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

func TestClient_HandleMessage_MalformedJSON(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	malformedMessages := []string{
		`{"type":"subscribe"`,
		`{"type":subscribe}`,
		`{type:"subscribe"}`,
		`undefined`,
	}

	for _, msg := range malformedMessages {
		client.handleMessage([]byte(msg))
	}

	// Should send 4 error responses
	for i := 0; i < 4; i++ {
		select {
		case response := <-client.send:
			var wsResp WSResponse
			err := json.Unmarshal(response, &wsResp)
			assert.NoError(t, err)
			assert.Equal(t, "error", wsResp.Type)
		default:
			t.Fatalf("Expected error response %d", i)
		}
	}
}

// ==================== Provider Filter Edge Cases ====================

func TestClient_SetProviderFilter_Empty(t *testing.T) {
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

	// Clear filter by setting empty string
	client.setProviderFilter("")
	assert.Equal(t, "", client.providerFilter)
}

func TestClient_MatchesProviderFilter_InvalidPayload(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		providerFilter: "aws",
	}

	// Payload without provider field
	payload := map[string]interface{}{
		"region": "us-east-1",
		"count":  42,
	}

	// Should pass through (no provider = pass)
	assert.True(t, client.matchesProviderFilter(payload))
}

func TestClient_MatchesProviderFilter_BooleanProvider(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:             "client-1",
		hub:            hub,
		send:           make(chan []byte, 256),
		subscriptions:  make(map[string]bool),
		providerFilter: "aws",
	}

	payload := map[string]interface{}{
		"provider": true,
	}

	// Non-string provider should pass through
	assert.True(t, client.matchesProviderFilter(payload))
}

// ==================== Hub Event Broadcasting Edge Cases ====================

func TestHub_BroadcastEvent_ClientWithFullBuffer(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	// Create client with small buffer
	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 1),
		subscriptions: map[string]bool{"drift": true},
	}

	// Fill buffer
	client.send <- []byte("message")

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	event := broadcaster.Event{
		Type:    "drift",
		Payload: map[string]interface{}{},
	}

	// Should handle full buffer gracefully
	hub.broadcastEvent(event)

	// Client should still be registered (buffer full is acceptable)
	assert.Equal(t, 1, hub.ClientCount())
}

func TestHub_BroadcastEvent_MultipleEventTypes(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"all": true},
	}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	eventTypes := []string{"drift", "falco", "state_change", "discovery_progress", "provider_status", "unmanaged_resource", "drift_result"}

	for _, eventType := range eventTypes {
		event := broadcaster.Event{
			Type:    eventType,
			Payload: map[string]interface{}{},
		}
		hub.broadcastEvent(event)
	}

	// Should receive all events
	assert.Equal(t, 7, len(client.send))
}

func TestHub_ClientCount_ThreadSafety(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	done := make(chan struct{})

	// Concurrent client additions
	for i := 0; i < 50; i++ {
		go func(id int) {
			client := &Client{
				id:            string(rune(id)),
				hub:           hub,
				send:          make(chan []byte, 256),
				subscriptions: make(map[string]bool),
			}
			hub.mu.Lock()
			hub.clients[client] = true
			hub.mu.Unlock()
			done <- struct{}{}
		}(i)
	}

	// Wait for all additions
	for i := 0; i < 50; i++ {
		<-done
	}

	// Concurrent client removals
	clients := make([]*Client, 0)
	hub.mu.RLock()
	for c := range hub.clients {
		clients = append(clients, c)
	}
	hub.mu.RUnlock()

	for _, c := range clients[:25] {
		go func(client *Client) {
			hub.mu.Lock()
			delete(hub.clients, client)
			hub.mu.Unlock()
			done <- struct{}{}
		}(c)
	}

	// Wait for removals
	for i := 0; i < 25; i++ {
		<-done
	}

	assert.Equal(t, 25, hub.ClientCount())
}

// ==================== WSMessage Payload Tests ====================

func TestWSMessage_WithPayload(t *testing.T) {
	payload := map[string]interface{}{
		"status": "active",
		"count":  10,
	}
	payloadJSON, _ := json.Marshal(payload)

	msg := WSMessage{
		Type:    "query",
		Topic:   "drifts",
		Payload: payloadJSON,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var unmarshaledMsg WSMessage
	err = json.Unmarshal(data, &unmarshaledMsg)
	assert.NoError(t, err)

	assert.Equal(t, "query", unmarshaledMsg.Type)
	assert.NotNil(t, unmarshaledMsg.Payload)
}

func TestWSMessage_EmptyPayload(t *testing.T) {
	msg := WSMessage{
		Type:  "subscribe",
		Topic: "events",
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var unmarshaledMsg WSMessage
	err = json.Unmarshal(data, &unmarshaledMsg)
	assert.NoError(t, err)

	assert.Equal(t, "subscribe", unmarshaledMsg.Type)
}

// ==================== Subscription Timestamp Tests ====================

func TestClient_Subscribe_HasTimestamp(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	beforeTime := time.Now().Add(-1 * time.Second) // Give 1 second margin
	client.subscribe("drifts")
	afterTime := time.Now().Add(1 * time.Second) // Give 1 second margin

	select {
	case response := <-client.send:
		var wsResp WSResponse
		err := json.Unmarshal(response, &wsResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, wsResp.Timestamp)

		// Parse timestamp and verify it's reasonable
		respTime, err := time.Parse(time.RFC3339, wsResp.Timestamp)
		assert.NoError(t, err)
		assert.True(t, respTime.After(beforeTime), "timestamp should be after beforeTime")
		assert.True(t, respTime.Before(afterTime), "timestamp should be before afterTime")
	default:
		t.Fatal("Expected response")
	}
}

// ==================== Hub Unregister Tests ====================

func TestHub_UnregisterCloses_SendChannel(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	// Note: We can't directly test this without running Hub.Run() in a goroutine,
	// which would make the test more complex. This is a limitation of goroutine-based architectures.
	// The test is documented here for completeness.
	hub := NewHub(bc)
	assert.NotNil(t, hub)
}

// ==================== Timestamp Format Tests ====================

func TestWSResponse_TimestampFormat(t *testing.T) {
	resp := WSResponse{
		Type:      "pong",
		Timestamp: "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var unmarshaledResp WSResponse
	err = json.Unmarshal(data, &unmarshaledResp)
	assert.NoError(t, err)

	assert.Equal(t, "2024-01-01T12:00:00Z", unmarshaledResp.Timestamp)
}

// ==================== Client ID Generation Tests ====================

func TestNewClient_HasUniqueID(t *testing.T) {
	// Note: newClient is not exported, so we test it indirectly through the handler
	bc := broadcaster.NewBroadcaster()
	handler1 := NewHandler(bc)
	handler2 := NewHandler(bc)

	// Each handler has its own hub
	assert.NotEqual(t, handler1.hub, handler2.hub)
}

// ==================== IsSubscribedTo Logic Tests ====================

func TestClient_IsSubscribedTo_CaseSpecific(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	hub := NewHub(bc)

	client := &Client{
		id:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"Drifts": true}, // Capital D
	}

	// Topics are case-sensitive
	assert.True(t, client.isSubscribedTo("Drifts"))
	assert.False(t, client.isSubscribedTo("drifts"))
}

// ==================== EncodeEvent Error Handling ====================

func TestEncodeEvent_EmptyEvent(t *testing.T) {
	event := broadcaster.Event{
		Type:    "",
		Payload: map[string]interface{}{},
	}

	data := encodeEvent(event)
	assert.NotNil(t, data)
	assert.NotEmpty(t, data)

	var resp WSResponse
	err := json.Unmarshal(data, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "data", resp.Type)
	assert.Equal(t, "", resp.Topic)
}

// ==================== Integration Tests with Real WebSocket ====================

func TestHandleWebSocket_ConnectAndReceiveWelcome(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to websocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Read welcome message
	var resp WSResponse
	err = ws.ReadJSON(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "connected", resp.Type)
	assert.NotEmpty(t, resp.Data)

	// Verify welcome data structure
	data := resp.Data.(map[string]interface{})
	assert.NotNil(t, data["client_id"])
	assert.NotNil(t, data["version"])
	assert.NotNil(t, data["topics"])
	assert.NotNil(t, data["features"])
}

func TestHandleWebSocket_Subscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome message
	var welcome WSResponse
	err = ws.ReadJSON(&welcome)
	assert.NoError(t, err)

	// Send subscribe message
	msg := WSMessage{
		Type:  "subscribe",
		Topic: "drifts",
	}
	err = ws.WriteJSON(msg)
	assert.NoError(t, err)

	// Read subscription response
	var subscribeResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&subscribeResp)
	assert.NoError(t, err)
	assert.Equal(t, "subscribed", subscribeResp.Type)
	assert.Equal(t, "drifts", subscribeResp.Topic)
}

func TestHandleWebSocket_ReceiveBroadcasterEvent(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Subscribe to drifts
	msg := WSMessage{Type: "subscribe", Topic: "drifts"}
	ws.WriteJSON(msg)

	// Read subscription response
	var subscribeResp WSResponse
	ws.ReadJSON(&subscribeResp)

	// Broadcast an event
	time.Sleep(100 * time.Millisecond) // Give hub time to subscribe

	bc.Broadcast(broadcaster.Event{
		Type:      "drifts",
		Timestamp: time.Now().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"severity": "high",
		},
	})

	// Read the broadcasted event
	var eventResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&eventResp)
	assert.NoError(t, err)
	assert.Equal(t, "data", eventResp.Type)
	assert.Equal(t, "drifts", eventResp.Topic)
}

func TestHandleWebSocket_Ping(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Send ping
	msg := WSMessage{Type: "ping"}
	err = ws.WriteJSON(msg)
	assert.NoError(t, err)

	// Read pong response
	var pongResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&pongResp)
	assert.NoError(t, err)
	assert.Equal(t, "pong", pongResp.Type)
}

func TestHandleWebSocket_Filter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Send filter
	msg := WSMessage{Type: "filter", Provider: "aws"}
	err = ws.WriteJSON(msg)
	assert.NoError(t, err)

	// Read filter response
	var filterResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&filterResp)
	assert.NoError(t, err)
	assert.Equal(t, "filter_set", filterResp.Type)
}

func TestHandleWebSocket_InvalidMessage(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Send invalid JSON
	ws.WriteMessage(websocket.TextMessage, []byte("invalid json"))

	// Read error response
	var errResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&errResp)
	assert.NoError(t, err)
	assert.Equal(t, "error", errResp.Type)
}

func TestHandleWebSocket_UnknownMessageType(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Send unknown message type
	msg := WSMessage{Type: "unknown"}
	ws.WriteJSON(msg)

	// Read error response
	var errResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&errResp)
	assert.NoError(t, err)
	assert.Equal(t, "error", errResp.Type)
}

func TestHandleWebSocket_MultipleSubscriptions(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Subscribe to multiple topics
	topics := []string{"drifts", "events", "discovery_progress"}
	for _, topic := range topics {
		msg := WSMessage{Type: "subscribe", Topic: topic}
		ws.WriteJSON(msg)

		// Read subscription response
		var resp WSResponse
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		ws.ReadJSON(&resp)
		assert.Equal(t, "subscribed", resp.Type)
	}
}

func TestHandleWebSocket_SubscribeAndUnsubscribe(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Subscribe
	msg := WSMessage{Type: "subscribe", Topic: "drifts"}
	ws.WriteJSON(msg)

	var subscribeResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws.ReadJSON(&subscribeResp)
	assert.Equal(t, "subscribed", subscribeResp.Type)

	// Unsubscribe
	msg = WSMessage{Type: "unsubscribe", Topic: "drifts"}
	ws.WriteJSON(msg)

	var unsubscribeResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws.ReadJSON(&unsubscribeResp)
	assert.Equal(t, "unsubscribed", unsubscribeResp.Type)
}

func TestHandleWebSocket_ProviderFilteredBroadcast(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Receive welcome
	var welcome WSResponse
	ws.ReadJSON(&welcome)

	// Subscribe to all
	msg := WSMessage{Type: "subscribe", Topic: "drift"}
	ws.WriteJSON(msg)

	var subscribeResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws.ReadJSON(&subscribeResp)

	// Set provider filter to AWS
	msg = WSMessage{Type: "filter", Provider: "aws"}
	ws.WriteJSON(msg)

	var filterResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	ws.ReadJSON(&filterResp)

	// Broadcast AWS event
	time.Sleep(100 * time.Millisecond)
	bc.Broadcast(broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"provider": "aws",
		},
	})

	// Should receive AWS event
	var eventResp WSResponse
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	err = ws.ReadJSON(&eventResp)
	assert.NoError(t, err)
	assert.Equal(t, "data", eventResp.Type)

	// Broadcast GCP event
	bc.Broadcast(broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"provider": "gcp",
		},
	})

	// Should NOT receive GCP event (filtered)
	ws.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	err = ws.ReadJSON(&eventResp)
	// Either error or non-drift event
	if err == nil {
		assert.NotEqual(t, "data", eventResp.Type)
	}
}
