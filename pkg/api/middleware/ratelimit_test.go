package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRateLimiter(enabled bool, rpm int, burst int) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		Enabled:        enabled,
		RequestsPerMin: rpm,
		BurstSize:      burst,
	})
}

func TestRateLimiter_Disabled(t *testing.T) {
	rl := newTestRateLimiter(false, 10, 2)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Should allow unlimited requests when disabled
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}
}

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := newTestRateLimiter(true, 60, 5)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 5 requests should succeed (burst size)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newTestRateLimiter(true, 60, 3)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust burst
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	var resp models.APIResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error.Message, "Rate limit exceeded")
}

func TestRateLimiter_Headers(t *testing.T) {
	rl := newTestRateLimiter(true, 60, 5)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "60", rr.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	rl := newTestRateLimiter(true, 60, 1)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the single token
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Next should be rate limited with Retry-After
	req = httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))
}

func TestRateLimiter_SeparateClientsIndependent(t *testing.T) {
	rl := newTestRateLimiter(true, 60, 2)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Client A uses its burst
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Client A is now rate limited
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Client B should still have its own bucket
	req = httptest.NewRequest("GET", "/api/v1/events", nil)
	req.RemoteAddr = "10.0.0.2:5678"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimiter_DefaultValues(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		Enabled:        true,
		RequestsPerMin: 0, // Should default to 60
		BurstSize:      0, // Should default to 10
	})
	defer rl.Stop()

	assert.Equal(t, 60, rl.config.RequestsPerMin)
	assert.Equal(t, 10, rl.config.BurstSize)
}
