package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64 // Requests per second per IP
	BurstSize         int64   // Maximum burst size
}

// DefaultRateLimitConfig returns default rate limiting settings
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 100.0, // 100 requests per second
		BurstSize:         10,    // Burst of 10 requests
	}
}

// TokenBucket implements a simple token bucket rate limiter
type TokenBucket struct {
	tokens    float64
	maxTokens int64
	lastTime  time.Time
	refillRate float64
	mu        sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(maxTokens int64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(maxTokens),
		maxTokens:  maxTokens,
		lastTime:   time.Now(),
		refillRate: refillRate,
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// Refill tokens based on elapsed time
	tokensToAdd := elapsed * tb.refillRate
	maxFloat := float64(tb.maxTokens)
	if tb.tokens+tokensToAdd > maxFloat {
		tb.tokens = maxFloat
	} else {
		tb.tokens = tb.tokens + tokensToAdd
	}

	if tb.tokens >= 1.0 {
		tb.tokens--
		return true
	}

	return false
}

// GetWaitTime returns how long to wait before the next request is allowed
func (tb *TokenBucket) GetWaitTime() time.Duration {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens >= 1.0 {
		return 0
	}

	// Calculate how long until we have 1 token
	tokensNeeded := 1.0 - tb.tokens
	waitSeconds := tokensNeeded / tb.refillRate
	return time.Duration(float64(time.Second) * waitSeconds)
}

// RateLimiter implements per-IP rate limiting using token buckets
type RateLimiter struct {
	config  *RateLimitConfig
	buckets sync.Map // map[string]*TokenBucket
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *RateLimitConfig) *RateLimiter {
	if cfg == nil {
		cfg = DefaultRateLimitConfig()
	}
	return &RateLimiter{
		config: cfg,
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Get the first IP in the chain
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getOrCreateBucket gets or creates a token bucket for an IP
func (rl *RateLimiter) getOrCreateBucket(ip string) *TokenBucket {
	val, _ := rl.buckets.LoadOrStore(ip, NewTokenBucket(
		rl.config.BurstSize,
		rl.config.RequestsPerSecond,
	))
	return val.(*TokenBucket)
}

// RateLimit middleware enforces per-IP rate limiting
func RateLimit(cfg *RateLimitConfig) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			bucket := limiter.getOrCreateBucket(clientIP)

			if !bucket.Allow() {
				retryAfter := bucket.GetWaitTime()
				w.Header().Set("Retry-After", retryAfter.String())
				w.Header().Set("X-RateLimit-Limit", "exceeded")

				log.WithFields(log.Fields{
					"client_ip":  clientIP,
					"retry_after": retryAfter,
				}).Warn("Rate limit exceeded")

				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

