package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// DriftsHandler handles drift alerts-related requests
type DriftsHandler struct {
	store *graph.Store
}

// NewDriftsHandler creates a new drifts handler
func NewDriftsHandler(store *graph.Store) *DriftsHandler {
	return &DriftsHandler{
		store: store,
	}
}

// GetDrifts handles GET /api/v1/drifts
func (h *DriftsHandler) GetDrifts(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/drifts")

	// Parse pagination parameters
	params := ParsePagination(r, 50)

	// Parse filter parameters
	severity := r.URL.Query().Get("severity")
	resourceType := r.URL.Query().Get("resource_type")

	// Get all drifts
	allDrifts := h.store.GetDrifts()

	// Apply filters
	filteredDrifts := make([]map[string]interface{}, 0)
	for _, drift := range allDrifts {
		// Apply severity filter
		if severity != "" && drift.Severity != severity {
			continue
		}

		// Apply resource type filter
		if resourceType != "" && drift.ResourceType != resourceType {
			continue
		}

		driftData := map[string]interface{}{
			"id":            drift.ResourceID,
			"severity":      drift.Severity,
			"resource_type": drift.ResourceType,
			"resource_name": drift.ResourceName,
			"resource_id":   drift.ResourceID,
			"attribute":     drift.Attribute,
			"old_value":     drift.OldValue,
			"new_value":     drift.NewValue,
			"user_identity": drift.UserIdentity,
			"matched_rules": drift.MatchedRules,
			"timestamp":     drift.Timestamp,
			"alert_type":    drift.AlertType,
		}
		filteredDrifts = append(filteredDrifts, driftData)
	}

	// Apply pagination
	total := len(filteredDrifts)
	paginatedDrifts := Paginate(filteredDrifts, params)

	response := PaginatedResponseData(paginatedDrifts, params, total)
	respondJSON(w, http.StatusOK, response)
}

// GetDrift handles GET /api/v1/drifts/:id
func (h *DriftsHandler) GetDrift(w http.ResponseWriter, r *http.Request) {
	driftID := chi.URLParam(r, "id")
	log.Debugf("GET /api/v1/drifts/%s", driftID)

	// Get all drifts
	allDrifts := h.store.GetDrifts()

	// Find drift by ID
	for _, drift := range allDrifts {
		if drift.ResourceID == driftID {
			driftData := map[string]interface{}{
				"id":            drift.ResourceID,
				"severity":      drift.Severity,
				"resource_type": drift.ResourceType,
				"resource_name": drift.ResourceName,
				"resource_id":   drift.ResourceID,
				"attribute":     drift.Attribute,
				"old_value":     drift.OldValue,
				"new_value":     drift.NewValue,
				"user_identity": drift.UserIdentity,
				"matched_rules": drift.MatchedRules,
				"timestamp":     drift.Timestamp,
				"alert_type":    drift.AlertType,
			}

			respondJSON(w, http.StatusOK, driftData)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Drift alert not found")
}
