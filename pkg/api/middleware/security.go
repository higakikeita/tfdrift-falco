package middleware

import (
	"net/http"
	"os"
	"strings"
)

// SecurityHeaders middleware adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options: Prevents MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevents clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: Browser XSS protection (for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security: HTTPS enforcement (only when TLS is enabled)
		if isTLSEnabled(r) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content-Security-Policy: Restrict resource loading
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Referrer-Policy: Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Disable certain browser features
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// isTLSEnabled checks if TLS is enabled for the request or application
func isTLSEnabled(r *http.Request) bool {
	// Check if request is HTTPS
	if r.TLS != nil {
		return true
	}

	// Check X-Forwarded-Proto header (for reverse proxies)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}

	// Check environment variable
	if strings.ToLower(os.Getenv("TLS_ENABLED")) == "true" {
		return true
	}

	return false
}
