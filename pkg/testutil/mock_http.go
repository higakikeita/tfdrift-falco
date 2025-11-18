package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockHTTPServer is a mock HTTP server for testing notifications and webhooks
type MockHTTPServer struct {
	Server        *httptest.Server
	mu            sync.Mutex
	requests      []*http.Request
	requestBodies []string
	statusCode    int
	responseBody  string
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer() *MockHTTPServer {
	mock := &MockHTTPServer{
		requests:      make([]*http.Request, 0),
		requestBodies: make([]string, 0),
		statusCode:    http.StatusOK,
		responseBody:  `{"ok":true}`,
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.handleRequest(w, r)
	}))

	return mock
}

// handleRequest handles incoming HTTP requests
func (m *MockHTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store request
	m.requests = append(m.requests, r)

	// Read and store body
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		m.requestBodies = append(m.requestBodies, string(body))
	}

	// Send response
	w.WriteHeader(m.statusCode)
	w.Write([]byte(m.responseBody))
}

// SetStatusCode sets the status code to return
func (m *MockHTTPServer) SetStatusCode(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusCode = code
}

// SetResponseBody sets the response body to return
func (m *MockHTTPServer) SetResponseBody(body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseBody = body
}

// GetRequestCount returns the number of requests received
func (m *MockHTTPServer) GetRequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// GetLastRequest returns the last request received
func (m *MockHTTPServer) GetLastRequest() *http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requests) == 0 {
		return nil
	}
	return m.requests[len(m.requests)-1]
}

// GetLastRequestBody returns the body of the last request received
func (m *MockHTTPServer) GetLastRequestBody() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requestBodies) == 0 {
		return ""
	}
	return m.requestBodies[len(m.requestBodies)-1]
}

// GetAllRequests returns all requests received
func (m *MockHTTPServer) GetAllRequests() []*http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests
}

// GetAllRequestBodies returns all request bodies received
func (m *MockHTTPServer) GetAllRequestBodies() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requestBodies
}

// Reset resets the mock server state
func (m *MockHTTPServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = make([]*http.Request, 0)
	m.requestBodies = make([]string, 0)
	m.statusCode = http.StatusOK
	m.responseBody = `{"ok":true}`
}

// Close closes the mock server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// URL returns the URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}
