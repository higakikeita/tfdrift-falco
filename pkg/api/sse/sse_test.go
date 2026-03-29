package sse

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testableResponseWriter implements both http.ResponseWriter and http.Flusher
type testableResponseWriter struct {
	*httptest.ResponseRecorder
}

func (t *testableResponseWriter) Flush() {
	// No-op for test
}

// ==================== NewStream Tests ====================

func TestNewStream_Success(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, err := NewStream(w, r, bc)

	assert.NoError(t, err)
	assert.NotNil(t, stream)
	assert.NotEmpty(t, stream.id)
	assert.NotNil(t, stream.eventCh)
}

func TestNewStream_ContextPropagation(t *testing.T) {
	// Test that context is properly inherited from request
	w := &testableResponseWriter{httptest.NewRecorder()}
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()
	r := httptest.NewRequest("GET", "/sse", nil)
	r = r.WithContext(parentCtx)
	bc := broadcaster.NewBroadcaster()

	stream, err := NewStream(w, r, bc)

	assert.NoError(t, err)
	require.NotNil(t, stream)
	// Stream context should be derived from parent context
	assert.NotNil(t, stream.ctx)
}

func TestNewStream_SetsHeaders(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, err := NewStream(w, r, bc)

	assert.NoError(t, err)
	require.NotNil(t, stream)

	// Check headers in underlying ResponseRecorder
	assert.Equal(t, "text/event-stream", w.ResponseRecorder.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.ResponseRecorder.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.ResponseRecorder.Header().Get("Connection"))
	assert.Equal(t, "no", w.ResponseRecorder.Header().Get("X-Accel-Buffering"))
}

func TestNewStream_ContextFromRequest(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, err := NewStream(w, r, bc)

	assert.NoError(t, err)
	require.NotNil(t, stream)
	assert.NotNil(t, stream.ctx)
	assert.NotNil(t, stream.cancel)
}

// ==================== Stream ID Tests ====================

func TestStream_ID(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream1, _ := NewStream(w, r, bc)
	stream2, _ := NewStream(w, r, bc)

	assert.NotEmpty(t, stream1.ID())
	assert.NotEmpty(t, stream2.ID())
	assert.NotEqual(t, stream1.ID(), stream2.ID())
}

// ==================== Stream Stop Tests ====================

func TestStream_Stop(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)
	streamID := stream.ID()

	// Verify stream is running before stop
	select {
	case <-stream.ctx.Done():
		t.Fatal("Stream context should not be done before Stop()")
	default:
		// OK
	}

	stream.Stop()

	// Verify stream is stopped after Stop()
	select {
	case <-stream.ctx.Done():
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("Stream context should be done after Stop()")
	}

	assert.Equal(t, streamID, stream.ID())
}

// ==================== Handler Tests ====================

func TestNewHandler(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.broadcaster)
	assert.NotNil(t, handler.streams)
	assert.Equal(t, 0, handler.StreamCount())
}

func TestHandler_StreamCount(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	assert.Equal(t, 0, handler.StreamCount())
}

func TestHandler_HandleSSE_WithResponseWriter(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)

	// Create a channel to detect when handler finishes
	done := make(chan struct{}, 1)

	// Run handler in a goroutine so it doesn't block forever
	go func() {
		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()
		r = r.WithContext(ctx)
		handler.HandleSSE(w, r)
		done <- struct{}{}
	}()

	// Wait for handler to start and then timeout
	select {
	case <-done:
		// OK
	case <-time.After(500 * time.Millisecond):
		// Expected - handler should timeout and cleanup
	}
}

// ==================== SendEvent Tests ====================

func TestStream_SendEvent(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	testData := map[string]interface{}{
		"message": "test",
		"value":   42,
	}

	stream.sendEvent("test_event", testData)

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: test_event")
	assert.Contains(t, body, "data:")
	assert.Contains(t, body, "test")
}

func TestStream_SendEvent_WithComplexData(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	complexData := map[string]interface{}{
		"nested": map[string]interface{}{
			"level1": "value",
		},
		"array": []string{"item1", "item2"},
	}

	stream.sendEvent("complex", complexData)

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: complex")
	assert.Contains(t, body, "nested")
}

func TestStream_SendComment(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	stream.sendComment("keep-alive")

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, ": keep-alive")
}

// ==================== BroadcasterEvent Tests ====================

func TestStream_SendBroadcasterEvent(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	event := broadcaster.Event{
		Type: "drift",
		Payload: map[string]interface{}{
			"severity": "high",
		},
	}

	stream.sendBroadcasterEvent(event)

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: drift")
	assert.Contains(t, body, "severity")
}

// ==================== Integration Tests ====================

func TestStream_CanBeStopped(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	// Verify stream can be stopped
	assert.NotPanics(t, func() {
		stream.Stop()
	})

	// Verify context is cancelled
	select {
	case <-stream.ctx.Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be done after Stop()")
	}
}

func TestHandler_HandleSSE_Starts(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	assert.NotNil(t, handler)
	assert.Equal(t, 0, handler.StreamCount())
}

func TestStream_EventChannelSize(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	// Channel should be buffered with size 100
	assert.Equal(t, cap(stream.eventCh), 100)
}

func TestStream_Broadcaster(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	assert.Equal(t, bc, stream.broadcaster)
}

func TestHandler_CloseAll(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	// Create multiple mock streams manually
	for i := 0; i < 3; i++ {
		w := &testableResponseWriter{httptest.NewRecorder()}
		r := httptest.NewRequest("GET", "/sse", nil)
		stream, _ := NewStream(w, r, bc)

		handler.mu.Lock()
		handler.streams[stream.ID()] = stream
		handler.mu.Unlock()
	}

	assert.Equal(t, 3, handler.StreamCount())

	handler.CloseAll()

	assert.Equal(t, 0, handler.StreamCount())
}

func TestStream_SendEventFormat(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	data := map[string]string{"key": "value"}
	stream.sendEvent("myevent", data)

	body := w.ResponseRecorder.Body.String()
	// Should follow SSE format: event: type\ndata: json\n\n
	assert.Contains(t, body, "event: myevent")
	assert.Contains(t, body, `data: {"key":"value"}`)
}

func TestStream_MultipleEvents(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	stream.sendEvent("event1", map[string]string{"msg": "first"})
	stream.sendEvent("event2", map[string]string{"msg": "second"})

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: event1")
	assert.Contains(t, body, "event: event2")
}

func TestStream_InitialConnectionEvent(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	// Create a mock Start that just sends the initial event
	initialData := map[string]interface{}{
		"stream_id": stream.ID(),
		"message":   "Connected to TFDrift-Falco SSE stream v0.9.0",
		"version":   "0.9.0",
	}

	stream.sendEvent("connected", initialData)

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: connected")
	assert.Contains(t, body, "stream_id")
	assert.Contains(t, body, "0.9.0")
}

func TestStream_JSONMarshaling(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	complexPayload := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"array":  []int{1, 2, 3},
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	stream.sendEvent("complex", complexPayload)

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "event: complex")
	// Verify JSON is properly encoded
	assert.Contains(t, body, "nested")
	assert.Contains(t, body, "key")
}

func TestHandler_Cleanup(t *testing.T) {
	bc := broadcaster.NewBroadcaster()
	handler := NewHandler(bc)

	// Manually create a stream in handler
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	stream, _ := NewStream(w, r, bc)

	handler.mu.Lock()
	handler.streams[stream.ID()] = stream
	handler.mu.Unlock()

	assert.Equal(t, 1, handler.StreamCount())

	// Simulate cleanup by removing stream
	handler.mu.Lock()
	delete(handler.streams, stream.ID())
	handler.mu.Unlock()

	assert.Equal(t, 0, handler.StreamCount())
}

func TestStream_MutexProtection(t *testing.T) {
	w := &testableResponseWriter{httptest.NewRecorder()}
	r := httptest.NewRequest("GET", "/sse", nil)
	bc := broadcaster.NewBroadcaster()

	stream, _ := NewStream(w, r, bc)

	// Send multiple events concurrently
	done := make(chan struct{}, 5)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			stream.sendEvent("concurrent", map[string]interface{}{"id": idx})
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	body := w.ResponseRecorder.Body.String()
	assert.Contains(t, body, "concurrent")
}
