package handlers

import (
	"net/http"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

// GetHealth handles GET /health
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Version:   h.version,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondJSON(w, http.StatusOK, response)
}
