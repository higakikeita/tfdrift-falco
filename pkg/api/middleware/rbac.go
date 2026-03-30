package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/rbac"
	log "github.com/sirupsen/logrus"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys for RBAC
var (
	userIDCtxKey    = contextKey("user_id")
	userRoleCtxKey  = contextKey("user_role")
	rbacEnabledCtxKey = contextKey("rbac_enabled")
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	Enabled bool
	Engine  *rbac.Engine
}

// GlobalRBACConfig holds the global RBAC configuration
var GlobalRBACConfig = &RBACConfig{
	Enabled: strings.ToLower(os.Getenv("RBAC_ENABLED")) == "true",
	Engine:  rbac.NewEngine(),
}

// SetRBACConfig sets the global RBAC configuration
func SetRBACConfig(cfg *RBACConfig) {
	GlobalRBACConfig = cfg
}

// GetRBACConfig returns the global RBAC configuration
func GetRBACConfig() *RBACConfig {
	return GlobalRBACConfig
}

// RequireRole returns middleware that requires the user to have one of the specified roles
func RequireRole(roles ...rbac.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := GetRBACConfig()

			// Skip RBAC check if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get user role from context (set by auth middleware)
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
				return
			}

			// Check if user has one of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}

				// Also check hierarchy - a higher role satisfies a lower role requirement
				if cfg.Engine.IsRoleAtLeast(userRole, requiredRole) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondWithError(w, http.StatusForbidden, "Forbidden", "Insufficient role privileges")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission returns middleware that requires the user to have a specific permission
func RequirePermission(resource rbac.Resource, permission rbac.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg := GetRBACConfig()

			// Skip RBAC check if disabled
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Get user ID from context
			userID, ok := GetUserID(r.Context())
			if !ok {
				respondWithError(w, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
				return
			}

			// Check if user has permission
			if !cfg.Engine.UserCanAccess(userID, resource, permission) {
				respondWithError(w, http.StatusForbidden, "Forbidden", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetUserContext sets user information in the request context
// This should be called by auth middleware after successful authentication
func SetUserContext(r *http.Request, userID string, role rbac.Role) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, userIDCtxKey, userID)
	ctx = context.WithValue(ctx, userRoleCtxKey, role)
	return r.WithContext(ctx)
}

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDCtxKey).(string)
	return userID, ok
}

// GetUserRole retrieves the user role from the request context
func GetUserRole(ctx context.Context) (rbac.Role, bool) {
	role, ok := ctx.Value(userRoleCtxKey).(rbac.Role)
	return role, ok
}

// respondWithError writes a JSON error response
func respondWithError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Errorf("Failed to encode error response: %v", err)
	}
}
