package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
)

// CorrelationsHandler handles cross-cloud correlation endpoints.
type CorrelationsHandler struct {
	correlator *detector.CrossCloudCorrelator
}

// NewCorrelationsHandler creates a new correlations handler.
func NewCorrelationsHandler(correlator *detector.CrossCloudCorrelator) *CorrelationsHandler {
	return &CorrelationsHandler{
		correlator: correlator,
	}
}

// GetCorrelations returns all correlation groups.
// GET /api/v1/correlations
func (h *CorrelationsHandler) GetCorrelations(w http.ResponseWriter, r *http.Request) {
	groups := h.correlator.GetGroups()

	// Filter by provider if specified
	providerFilter := r.URL.Query().Get("provider")
	if providerFilter != "" {
		groups = h.correlator.GetGroupsByProvider(providerFilter)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"correlations": groups,
			"count":        len(groups),
			"timestamp":    time.Now().Format(time.RFC3339),
		},
	})
}

// GetCorrelationStats returns correlation statistics.
// GET /api/v1/correlations/stats
func (h *CorrelationsHandler) GetCorrelationStats(w http.ResponseWriter, r *http.Request) {
	stats := h.correlator.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    stats,
	})
}
