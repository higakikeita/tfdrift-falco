package middleware

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// RequestIDKey is the context key for request ID
type RequestIDKey struct{}

// ValidationConfig contains validation settings
type ValidationConfig struct {
	MaxBodySize int64
}

// DefaultValidationConfig returns default validation settings
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxBodySize: 1024 * 1024, // 1MB default
	}
}

// InputValidation middleware validates HTTP requests
func InputValidation(cfg *ValidationConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultValidationConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add request ID to context
			requestID := uuid.New().String()
			ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
			w.Header().Set("X-Request-ID", requestID)

			// Sanitize path parameters - check this for ALL requests first
			if !isValidPath(r.URL.Path) {
				log.WithFields(log.Fields{
					"request_id": requestID,
					"path":       r.URL.Path,
				}).Warn("Invalid path with potential traversal attempt")
				http.Error(w, "Invalid request path", http.StatusBadRequest)
				return
			}

			// Skip further validation for GET, HEAD, DELETE requests
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Validate Content-Type for POST/PUT requests
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				contentType := r.Header.Get("Content-Type")
				if contentType == "" {
					log.WithFields(log.Fields{
						"request_id": requestID,
						"path":       r.URL.Path,
						"method":     r.Method,
					}).Warn("Missing Content-Type header")
					http.Error(w, "Content-Type header is required", http.StatusBadRequest)
					return
				}

				// Validate Content-Type is JSON or form data
				if !isValidContentType(contentType) {
					log.WithFields(log.Fields{
						"request_id":   requestID,
						"path":         r.URL.Path,
						"method":       r.Method,
						"content_type": contentType,
					}).Warn("Invalid Content-Type header")
					http.Error(w, "Invalid Content-Type header", http.StatusBadRequest)
					return
				}
			}

			// Validate body size
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodySize)
				if err := validateBodySize(r, cfg.MaxBodySize); err != nil {
					log.WithFields(log.Fields{
						"request_id": requestID,
						"path":       r.URL.Path,
						"error":      err.Error(),
					}).Warn("Request body size validation failed")
					http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isValidContentType checks if content type is acceptable
func isValidContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	acceptedTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, accepted := range acceptedTypes {
		if strings.Contains(contentType, accepted) {
			return true
		}
	}
	return false
}

// validateBodySize checks that body doesn't exceed max size
func validateBodySize(r *http.Request, maxSize int64) error {
	// Try to get Content-Length header
	if r.Header.Get("Content-Length") != "" {
		contentLength, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err == nil && contentLength > maxSize {
			return io.EOF
		}
	}
	return nil
}

// isValidPath checks for path traversal attempts
func isValidPath(path string) bool {
	// Reject paths containing ".." which could indicate path traversal
	if strings.Contains(path, "..") {
		return false
	}

	// Reject paths with null bytes
	if strings.Contains(path, "\x00") {
		return false
	}

	return true
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}
