package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	apimiddleware "github.com/keitahigaki/tfdrift-falco/pkg/api/middleware"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
)

// AuthHandler handles authentication-related API endpoints.
type AuthHandler struct {
	auth *apimiddleware.Auth
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(auth *apimiddleware.Auth) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// tokenRequest is the request body for generating a JWT token.
type tokenRequest struct {
	Subject string `json:"subject"`
}

// tokenResponse is the response body containing a JWT token.
type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
}

// GenerateToken generates a new JWT token.
// POST /api/v1/auth/token
func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Subject == "" {
		respondError(w, http.StatusBadRequest, "Subject is required")
		return
	}

	token, err := h.auth.GenerateToken(req.Subject)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: tokenResponse{
			Token:     token,
			ExpiresIn: "24h",
		},
	})
}

// createAPIKeyRequest is the request body for creating an API key.
type createAPIKeyRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// createAPIKeyResponse is the response body containing a new API key.
type createAPIKeyResponse struct {
	Name      string   `json:"name"`
	Key       string   `json:"key"` // Only shown once at creation time
	Scopes    []string `json:"scopes"`
	CreatedAt string   `json:"created_at"`
}

// CreateAPIKey creates a new API key.
// POST /api/v1/auth/api-keys
func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	key, err := apimiddleware.GenerateAPIKey()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate API key")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	entry := apimiddleware.APIKeyEntry{
		Name:      req.Name,
		Key:       key,
		Scopes:    req.Scopes,
		CreatedAt: now,
	}

	h.auth.AddAPIKey(entry)

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: createAPIKeyResponse{
			Name:      req.Name,
			Key:       key,
			Scopes:    req.Scopes,
			CreatedAt: now,
		},
	})
}

// ListAPIKeys lists all API keys (values masked).
// GET /api/v1/auth/api-keys
func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.auth.ListAPIKeys()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    keys,
	})
}

// deleteAPIKeyRequest is the request body for revoking an API key.
type deleteAPIKeyRequest struct {
	Name string `json:"name"`
}

// RevokeAPIKey revokes an API key by name.
// DELETE /api/v1/auth/api-keys
func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	var req deleteAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if !h.auth.RemoveAPIKey(req.Name) {
		respondError(w, http.StatusNotFound, "API key not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "API key revoked"},
	})
}
