package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	log "github.com/sirupsen/logrus"
)

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled         bool   `yaml:"enabled" mapstructure:"enabled"`
	RequestsPerMin  int    `yaml:"requests_per_minute" mapstructure:"requests_per_minute"`
	BurstSize       int    `yaml:"burst_size" mapstructure:"burst_size"`
	CleanupInterval string `yaml:"cleanup_interval" mapstructure:"cleanup_interval"`
}

// tokenBucket implements a token bucket rate limiter per client.
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// RateLimiter is the rate limiting middleware.
type RateLimiter struct {
	config  RateLimitConfig
	buckets map[string]*tokenBucket
	mu      sync.Mutex
	stopCh  chan struct{}
}

// NewRateLimiter creates a new rate limiting middleware.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.RequestsPerMin <= 0 {
		cfg.RequestsPerMin = 60
	}
	if cfg.BurstSize <= 0 {
		cfg.BurstSize = 10
	}

	rl := &RateLimiter{
		config:  cfg,
		buckets: make(map[string]*tokenBucket),
		stopCh:  make(chan struct{}),
	}

	if cfg.Enabled {
		log.Info("API rate limiting enabled")
		log.Infof("  Rate: %d requests/min, burst: %d", cfg.RequestsPerMin, cfg.BurstSize)

		// Start cleanup goroutine
		cleanupInterval := 5 * time.Minute
		if cfg.CleanupInterval != "" {
			if d, err := time.ParseDuration(cfg.CleanupInterval); err == nil {
				cleanupInterval = d
			}
		}
		go rl.cleanupLoop(cleanupInterval)
	}

	return rl
}

// Middleware returns the rate limiting middleware handler.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Determine client identifier
		clientID := rl.getClientID(r)

		// Check rate limit
		bucket := rl.getBucket(clientID)
		allowed, remaining, resetAt := rl.tryConsume(bucket)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerMin))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			retryAfter := resetAt - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
			respondTooManyRequests(w, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientID determines the client identifier for rate limiting.
// Uses API key name if authenticated, otherwise falls back to IP.
func (rl *RateLimiter) getClientID(r *http.Request) string {
	// Prefer authenticated identity
	if info := GetAuthInfo(r.Context()); info != nil {
		return "auth:" + info.Subject
	}
	// Fall back to IP
	return "ip:" + r.RemoteAddr
}

// getBucket returns or creates a token bucket for the given client.
func (rl *RateLimiter) getBucket(clientID string) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &tokenBucket{
			tokens:     float64(rl.config.BurstSize),
			maxTokens:  float64(rl.config.BurstSize),
			refillRate: float64(rl.config.RequestsPerMin) / 60.0,
			lastRefill: time.Now(),
		}
		rl.buckets[clientID] = bucket
	}

	return bucket
}

// tryConsume attempts to consume a token from the bucket.
// Returns whether the request is allowed, remaining tokens, and reset timestamp.
func (rl *RateLimiter) tryConsume(bucket *tokenBucket) (bool, int, int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * bucket.refillRate
	if bucket.tokens > bucket.maxTokens {
		bucket.tokens = bucket.maxTokens
	}
	bucket.lastRefill = now

	// Calculate reset time (when bucket will be full again)
	tokensNeeded := bucket.maxTokens - bucket.tokens
	var resetAt int64
	if tokensNeeded > 0 && bucket.refillRate > 0 {
		resetAt = now.Add(time.Duration(tokensNeeded/bucket.refillRate) * time.Second).Unix()
	} else {
		resetAt = now.Unix()
	}

	if bucket.tokens < 1 {
		return false, 0, resetAt
	}

	bucket.tokens--
	remaining := int(bucket.tokens)
	return true, remaining, resetAt
}

// cleanupLoop periodically removes stale rate limit entries.
func (rl *RateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

// cleanup removes buckets that have been fully refilled (idle clients).
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for id, bucket := range rl.buckets {
		// If bucket hasn't been used in 2x the refill period, remove it
		elapsed := now.Sub(bucket.lastRefill).Seconds()
		if elapsed > 120 { // 2 minutes idle
			delete(rl.buckets, id)
		}
	}
}

// Stop stops the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// respondTooManyRequests sends a 429 JSON response.
func respondTooManyRequests(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    http.StatusTooManyRequests,
			Message: message,
		},
	})
}
