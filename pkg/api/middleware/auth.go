// Package middleware provides HTTP middleware for the TFDrift-Falco API server.
package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	log "github.com/sirupsen/logrus"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const (
	// ContextKeyAuthInfo is the context key for authentication information.
	ContextKeyAuthInfo contextKey = "auth_info"
)

// AuthInfo contains the authenticated identity information stored in request context.
type AuthInfo struct {
	Method  string   // "jwt" or "api_key"
	Subject string   // user ID or API key name
	Scopes  []string // authorized scopes (future use)
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled   bool          `yaml:"enabled" mapstructure:"enabled"`
	JWTSecret string        `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	JWTIssuer string        `yaml:"jwt_issuer" mapstructure:"jwt_issuer"`
	JWTExpiry string        `yaml:"jwt_expiry" mapstructure:"jwt_expiry"` // e.g. "24h"
	APIKeys   []APIKeyEntry `yaml:"api_keys" mapstructure:"api_keys"`
}

// APIKeyEntry represents a configured API key.
type APIKeyEntry struct {
	Name      string   `yaml:"name" mapstructure:"name"`
	Key       string   `yaml:"key" mapstructure:"key"`
	Scopes    []string `yaml:"scopes" mapstructure:"scopes"`
	CreatedAt string   `yaml:"created_at" mapstructure:"created_at"`
}

// Auth is the authentication middleware.
type Auth struct {
	config    AuthConfig
	jwtSecret []byte
	apiKeys   map[string]APIKeyEntry // key hash -> entry
	mu        sync.RWMutex
}

// NewAuth creates a new authentication middleware.
func NewAuth(cfg AuthConfig) *Auth {
	a := &Auth{
		config:    cfg,
		jwtSecret: []byte(cfg.JWTSecret),
		apiKeys:   make(map[string]APIKeyEntry),
	}

	// Index API keys for O(1) lookup
	for _, entry := range cfg.APIKeys {
		a.apiKeys[hashAPIKey(entry.Key)] = entry
	}

	if cfg.Enabled {
		log.Info("API authentication enabled")
		log.Infof("  JWT issuer: %s", cfg.JWTIssuer)
		log.Infof("  API keys configured: %d", len(cfg.APIKeys))
	}

	return a
}

// Middleware returns the authentication middleware handler.
// It checks for JWT Bearer tokens and X-API-Key headers.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Try JWT Bearer token first
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				info, err := a.validateJWT(tokenStr)
				if err != nil {
					respondUnauthorized(w, fmt.Sprintf("Invalid JWT: %v", err))
					return
				}
				ctx := context.WithValue(r.Context(), ContextKeyAuthInfo, info)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Try API Key
		if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
			info, err := a.validateAPIKey(apiKey)
			if err != nil {
				respondUnauthorized(w, "Invalid API key")
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyAuthInfo, info)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		respondUnauthorized(w, "Authentication required. Provide a Bearer token or X-API-Key header.")
	})
}

// validateJWT validates a JWT token and returns auth info.
func (a *Auth) validateJWT(tokenStr string) (*AuthInfo, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	}, jwt.WithIssuer(a.config.JWTIssuer), jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	sub, _ := claims.GetSubject()
	if sub == "" {
		sub = "unknown"
	}

	return &AuthInfo{
		Method:  "jwt",
		Subject: sub,
	}, nil
}

// validateAPIKey validates an API key and returns auth info.
func (a *Auth) validateAPIKey(key string) (*AuthInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	hashedInput := hashAPIKey(key)
	for storedHash, entry := range a.apiKeys {
		if subtle.ConstantTimeCompare([]byte(hashedInput), []byte(storedHash)) == 1 {
			return &AuthInfo{
				Method:  "api_key",
				Subject: entry.Name,
				Scopes:  entry.Scopes,
			}, nil
		}
	}
	return nil, fmt.Errorf("api key not found")
}

// GenerateToken creates a JWT token for the given subject.
func (a *Auth) GenerateToken(subject string) (string, error) {
	expiry := 24 * time.Hour
	if a.config.JWTExpiry != "" {
		d, err := time.ParseDuration(a.config.JWTExpiry)
		if err == nil {
			expiry = d
		}
	}

	claims := jwt.MapClaims{
		"sub": subject,
		"iss": a.config.JWTIssuer,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

// GenerateAPIKey generates a new random API key string.
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}
	return "tfd_" + hex.EncodeToString(bytes), nil
}

// hashAPIKey returns the SHA-256 hex digest of an API key.
func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// AddAPIKey adds a new API key at runtime.
func (a *Auth) AddAPIKey(entry APIKeyEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.apiKeys[hashAPIKey(entry.Key)] = entry
	a.config.APIKeys = append(a.config.APIKeys, entry)
}

// RemoveAPIKey removes an API key by name.
func (a *Auth) RemoveAPIKey(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	for key, entry := range a.apiKeys {
		if entry.Name == name {
			delete(a.apiKeys, key)
			// Remove from config slice
			for i, e := range a.config.APIKeys {
				if e.Name == name {
					a.config.APIKeys = append(a.config.APIKeys[:i], a.config.APIKeys[i+1:]...)
					break
				}
			}
			return true
		}
	}
	return false
}

// ListAPIKeys returns a list of API key metadata (without the actual key values).
func (a *Auth) ListAPIKeys() []APIKeyInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var keys []APIKeyInfo
	for _, entry := range a.apiKeys {
		keys = append(keys, APIKeyInfo{
			Name:      entry.Name,
			Prefix:    maskKey(entry.Key),
			Scopes:    entry.Scopes,
			CreatedAt: entry.CreatedAt,
		})
	}
	return keys
}

// APIKeyInfo is the public representation of an API key (key value masked).
type APIKeyInfo struct {
	Name      string   `json:"name"`
	Prefix    string   `json:"prefix"` // e.g. "tfd_a1b2...xxxx"
	Scopes    []string `json:"scopes"`
	CreatedAt string   `json:"created_at"`
}

// GetAuthInfo extracts AuthInfo from a request context.
func GetAuthInfo(ctx context.Context) *AuthInfo {
	info, _ := ctx.Value(ContextKeyAuthInfo).(*AuthInfo)
	return info
}

// maskKey returns a masked version of an API key showing only prefix and last 4 chars.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:8] + "..." + key[len(key)-4:]
}

// respondUnauthorized sends a 401 JSON response.
func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="tfdrift-falco"`)
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    http.StatusUnauthorized,
			Message: message,
		},
	})
}
