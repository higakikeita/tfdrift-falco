package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders_AllHeaders(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Test each header
	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Content-Security-Policy", "default-src 'self'"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "geolocation=(), microphone=(), camera=()"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			actual := w.Header().Get(tt.header)
			if actual != tt.expected {
				t.Errorf("header %s: expected %q, got %q", tt.header, tt.expected, actual)
			}
		})
	}
}

func TestSecurityHeaders_HTTPSRequest(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create HTTPS request
	req := httptest.NewRequest(http.MethodGet, "https://localhost/test", nil)
	req.TLS = &tls.ConnectionState{}

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check HSTS header
	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("Strict-Transport-Security header not set for HTTPS request")
	}

	if hsts != "max-age=31536000; includeSubDomains" {
		t.Errorf("expected HSTS header %q, got %q", "max-age=31536000; includeSubDomains", hsts)
	}
}

func TestSecurityHeaders_HTTPRequest(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create HTTP request (no TLS)
	req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check HSTS header should not be set
	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Errorf("Strict-Transport-Security header should not be set for HTTP request, got %q", hsts)
	}
}

func TestSecurityHeaders_XForwardedProto(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create request with X-Forwarded-Proto header
	req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check HSTS header
	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Error("Strict-Transport-Security header not set when X-Forwarded-Proto is https")
	}
}

func TestSecurityHeaders_ContentType(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify Content-Type is preserved and X-Content-Type-Options is set
	if w.Header().Get("Content-Type") == "" {
		t.Error("Content-Type header was removed")
	}

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("X-Content-Type-Options header not properly set")
	}
}

func TestSecurityHeaders_CSPHeader(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := SecurityHeaders(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Record response
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check CSP header
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("Content-Security-Policy header not set")
	}

	if csp != "default-src 'self'" {
		t.Errorf("CSP header incorrect: expected %q, got %q", "default-src 'self'", csp)
	}
}
