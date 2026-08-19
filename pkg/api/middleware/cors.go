package middleware

import (
	"os"
	"strings"

	"github.com/go-chi/cors"
)

// NewCORS creates CORS middleware with configurable allowed origins.
// Origins can be set via DRIFTWIRE_ALLOWED_ORIGINS environment variable (comma-separated).
// Defaults to localhost development origins if not set.
func NewCORS() *cors.Cors {
	origins := []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"}
	env := os.Getenv("DRIFTWIRE_ALLOWED_ORIGINS")
	if env == "" {
		// Renamed from tfdrift-falco: honour the pre-rename spelling too.
		env = os.Getenv("TFDRIFT_ALLOWED_ORIGINS")
	}
	if env != "" {
		origins = strings.Split(env, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}
	return cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
