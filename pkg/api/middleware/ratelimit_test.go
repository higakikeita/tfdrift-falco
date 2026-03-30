package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucket_Allow(t *testing.T) {
	// Create a token bucket with 2 max tokens and 1 token per second refill
	bucket := NewTokenBucket(2, 1.0)

	tests := []struct {
		name     string
		expected bool
		wait     time.Duration
	}{
		{"First request", true, 0},
		{"Second request", true, 0},
		{"Third request (over limit)", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}
			result := bucket.Allow()
			if result != tt.expected {
				t.Errorf("expected Allow() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	// Create a token bucket with 1 max token and 1 token per second refill
	bucket := NewTokenBucket(1, 1.0)

	// Consume the initial token
	if !bucket.Allow() {
		t.Fatal("First request should be allowed")
	}

	// Next request should fail (no tokens)
	if bucket.Allow() {
		t.Error("Second request should fail (no tokens)")
	}

	// Wait 1+ second for refill
	time.Sleep(1100 * time.Millisecond)

	// Next request should succeed (token refilled)
	if !bucket.Allow() {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimit_PerIP(t *testing.T) {
	// Create handler with low rate limit for testing
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	cfg := &RateLimitConfig{
		RequestsPerSecond: 2.0,
		BurstSize:         2,
	}
	handler := RateLimit(cfg)(testHandler)

	// Create requests from same IP
	makeRequest := func() int {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:8080"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w.Code
	}

	// First two requests should succeed
	if status := makeRequest(); status != http.StatusOK {
		t.Errorf("first request: expected status 200, got %d", status)
	}
	if status := makeRequest(); status != http.StatusOK {
		t.Errorf("second request: expected status 200, got %d", status)
	}

	// Third request should fail (rate limited)
	if status := makeRequest(); status != http.StatusTooManyRequests {
		t.Errorf("third request: expected status 429, got %d", status)
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	// Create handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RateLimitConfig{
		RequestsPerSecond: 2.0,
		BurstSize:         1,
	}
	handler := RateLimit(cfg)(testHandler)

	// Request from IP1
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// Request from IP2 should not be rate limited (different bucket)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:8080"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w1.Code != http.StatusOK {
		t.Errorf("IP1 first request: expected 200, got %d", w1.Code)
	}
	if w2.Code != http.StatusOK {
		t.Errorf("IP2 request: expected 200, got %d", w2.Code)
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	// Create handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RateLimitConfig{
		RequestsPerSecond: 1.0,
		BurstSize:         1,
	}
	handler := RateLimit(cfg)(testHandler)

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// Second request (should be rate limited)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:8080"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}

	// Check Retry-After header
	if w2.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header not set")
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:8080"

	ip := getClientIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")

	ip := getClientIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("expected IP 192.168.1.100, got %s", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:8080"
	req.Header.Set("X-Real-IP", "192.168.1.200")

	ip := getClientIP(req)
	if ip != "192.168.1.200" {
		t.Errorf("expected IP 192.168.1.200, got %s", ip)
	}
}

func TestRateLimit_BurstAllowance(t *testing.T) {
	// Create handler with burst size of 3
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &RateLimitConfig{
		RequestsPerSecond: 1.0,
		BurstSize:         3,
	}
	handler := RateLimit(cfg)(testHandler)

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:8080"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Fourth request should fail
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("fourth request: expected 429, got %d", w.Code)
	}
}
