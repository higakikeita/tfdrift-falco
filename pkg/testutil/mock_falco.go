package testutil

import (
	"context"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// MockFalcoClient is a mock implementation of Falco client for testing
type MockFalcoClient struct {
	mu           sync.Mutex
	events       []*types.Event
	eventIndex   int
	connected    bool
	connectError error
	streamError  error
	disconnected bool
}

// NewMockFalcoClient creates a new mock Falco client
func NewMockFalcoClient() *MockFalcoClient {
	return &MockFalcoClient{
		events:    make([]*types.Event, 0),
		connected: false,
	}
}

// AddEvent adds an event to be returned by the mock
func (m *MockFalcoClient) AddEvent(event *types.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

// AddEvents adds multiple events to be returned by the mock
func (m *MockFalcoClient) AddEvents(events []*types.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, events...)
}

// Connect simulates connecting to Falco
func (m *MockFalcoClient) Connect(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connectError != nil {
		return m.connectError
	}

	m.connected = true
	m.disconnected = false
	return nil
}

// StreamEvents simulates streaming events from Falco
func (m *MockFalcoClient) StreamEvents(ctx context.Context, eventChan chan<- *types.Event) error {
	m.mu.Lock()
	if m.streamError != nil {
		m.mu.Unlock()
		return m.streamError
	}
	m.mu.Unlock()

	// Send all configured events
	for {
		m.mu.Lock()
		if m.eventIndex >= len(m.events) {
			m.mu.Unlock()
			// Wait for context cancellation
			<-ctx.Done()
			return ctx.Err()
		}

		event := m.events[m.eventIndex]
		m.eventIndex++
		m.mu.Unlock()

		select {
		case eventChan <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Disconnect simulates disconnecting from Falco
func (m *MockFalcoClient) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	m.disconnected = true
}

// IsConnected returns whether the mock is connected
func (m *MockFalcoClient) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// SetConnectError sets an error to return on Connect
func (m *MockFalcoClient) SetConnectError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectError = err
}

// SetStreamError sets an error to return on StreamEvents
func (m *MockFalcoClient) SetStreamError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamError = err
}

// Reset resets the mock to initial state
func (m *MockFalcoClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make([]*types.Event, 0)
	m.eventIndex = 0
	m.connected = false
	m.connectError = nil
	m.streamError = nil
	m.disconnected = false
}

// GetEventCount returns the number of events configured
func (m *MockFalcoClient) GetEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

// GetProcessedEventCount returns the number of events that have been streamed
func (m *MockFalcoClient) GetProcessedEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.eventIndex
}
