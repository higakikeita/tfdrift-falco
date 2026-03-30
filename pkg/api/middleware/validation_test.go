package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInputValidation_ContentType(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		expectedStatus int
		body           []byte
	}{
		{
			name:           "POST with valid JSON content-type",
			method:         http.MethodPost,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			body:           []byte(`{"test": "data"}`),
		},
		{
			name:           "POST with valid form content-type",
			method:         http.MethodPost,
			contentType:    "application/x-www-form-urlencoded",
			expectedStatus: http.StatusOK,
			body:           []byte(`key=value`),
		},
		{
			name:           "POST without content-type",
			method:         http.MethodPost,
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
			body:           []byte(`{"test": "data"}`),
		},
		{
			name:           "POST with invalid content-type",
			method:         http.MethodPost,
			contentType:    "application/xml",
			expectedStatus: http.StatusBadRequest,
			body:           []byte(`<test>data</test>`),
		},
		{
			name:           "GET request skips content-type check",
			method:         http.MethodGet,
			contentType:    "",
			expectedStatus: http.StatusOK,
			body:           nil,
		},
		{
			name:           "PUT with valid JSON content-type",
			method:         http.MethodPut,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			body:           []byte(`{"test": "data"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with middleware
			handler := InputValidation(DefaultValidationConfig())(testHandler)

			// Create request
			var body io.Reader
			if tt.body != nil {
				body = bytes.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, "/test", body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Record response
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestInputValidation_BodySize(t *testing.T) {
	tests := []struct {
		name           string
		maxBodySize    int64
		bodySize       int64
		expectedStatus int
	}{
		{
			name:           "Body size within limit",
			maxBodySize:    2048,
			bodySize:       512,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Body size at limit",
			maxBodySize:    1024,
			bodySize:       1024,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with middleware
			cfg := &ValidationConfig{MaxBodySize: tt.maxBodySize}
			handler := InputValidation(cfg)(testHandler)

			// Create request with specific body size
			bodyData := make([]byte, tt.bodySize)
			for i := range bodyData {
				bodyData[i] = 'a'
			}

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyData))
			req.Header.Set("Content-Type", "application/json")
			req.ContentLength = tt.bodySize

			// Record response
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestInputValidation_PathTraversal(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "Normal path",
			path:           "/api/v1/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Path with double dots",
			path:           "/api/../config",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Deep nested path without traversal",
			path:           "/api/v1/users/123/data",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with middleware
			handler := InputValidation(DefaultValidationConfig())(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			// Record response
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestInputValidation_RequestID(t *testing.T) {
	// Create a test handler that checks for request ID
	requestIDFound := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != "" {
			requestIDFound = true
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := InputValidation(DefaultValidationConfig())(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !requestIDFound {
		t.Error("request ID was not added to context")
	}

	// Check that X-Request-ID header is set
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header was not set")
	}
}

func TestIsValidContentType(t *testing.T) {
	tests := []struct {
		name     string
		ct       string
		expected bool
	}{
		{"JSON", "application/json", true},
		{"JSON with charset", "application/json; charset=utf-8", true},
		{"Form data", "application/x-www-form-urlencoded", true},
		{"Multipart", "multipart/form-data; boundary=xyz", true},
		{"XML", "application/xml", false},
		{"Plain text", "text/plain", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidContentType(tt.ct)
			if result != tt.expected {
				t.Errorf("isValidContentType(%q) = %v, expected %v", tt.ct, result, tt.expected)
			}
		})
	}
}

func TestIsValidPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Normal path", "/api/v1/test", true},
		{"Root path", "/", true},
		{"Multiple segments", "/api/v1/users/123", true},
		{"Path with traversal", "/api/../config", false},
		{"Path with double traversal", "../../etc/passwd", false},
		{"Path with null byte", "/api/test\x00", false},
		{"Path with hyphen", "/api-v1/test", true},
		{"Path with underscore", "/api_v1/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPath(tt.path)
			if result != tt.expected {
				t.Errorf("isValidPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
