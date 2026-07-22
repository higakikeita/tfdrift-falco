package handlers

import (
	"net/http"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
)

// ReadinessProbe reports whether the service's real processing path is healthy.
// It returns overall readiness plus per-check detail (e.g. "falco": "connected").
type ReadinessProbe func() (ready bool, checks map[string]string)

// HealthHandler handles health check requests
type HealthHandler struct {
	version   string
	readiness ReadinessProbe
}

// NewHealthHandler creates a new health handler with no readiness probe
// (liveness only; always reports "ok" when the process is up).
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

// NewHealthHandlerWithReadiness creates a health handler that folds a readiness
// probe into the response, so /health surfaces a "degraded" signal when the
// real processing path (e.g. the Falco subscription) is down — instead of
// silently reporting "ok" (pus #9).
func NewHealthHandlerWithReadiness(version string, probe ReadinessProbe) *HealthHandler {
	return &HealthHandler{version: version, readiness: probe}
}

// GetHealth handles GET /health.
//
// It always returns HTTP 200 while the process is alive (so it stays usable as
// a liveness probe and does not trigger restart flapping when a dependency is
// down), but the body's Status becomes "degraded" and Checks explains which
// dependency is unhealthy. Monitoring/readiness should alert on Status != "ok".
func (h *HealthHandler) GetHealth(w http.ResponseWriter, _ *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Version:   h.version,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if h.readiness != nil {
		ready, checks := h.readiness()
		response.Checks = checks
		if !ready {
			response.Status = "degraded"
		}
	}

	respondJSON(w, http.StatusOK, response)
}
