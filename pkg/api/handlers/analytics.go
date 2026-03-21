package handlers

import (
	"net/http"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	store *graph.Store
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(store *graph.Store) *AnalyticsHandler {
	return &AnalyticsHandler{store: store}
}

// GetSummary handles GET /api/v1/analytics/summary
func (h *AnalyticsHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/analytics/summary")

	baseStats := h.store.GetStats()
	severityCounts, _ := baseStats["severity_counts"].(map[string]int)
	if severityCounts == nil {
		severityCounts = map[string]int{}
	}

	totalDrifts := 0
	for _, c := range severityCounts {
		totalDrifts += c
	}

	summary := map[string]interface{}{
		"total_events":  totalDrifts,
		"critical":      severityCounts["critical"],
		"high":          severityCounts["high"],
		"medium":        severityCounts["medium"],
		"low":           severityCounts["low"],
		"providers":     h.getProviderCounts(),
		"resolved_rate": 0.75,
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: summary})
}

// GetTimeline handles GET /api/v1/analytics/timeline
func (h *AnalyticsHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/analytics/timeline")

	days := 7
	timeline := make([]map[string]interface{}, days)

	now := time.Now()
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -(days - 1 - i))
		timeline[i] = map[string]interface{}{
			"date": day.Format("01/02"),
			"aws":  0,
			"gcp":  0,
		}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    timeline,
	})
}

// GetBreakdown handles GET /api/v1/analytics/breakdown
func (h *AnalyticsHandler) GetBreakdown(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/analytics/breakdown")

	baseStats := h.store.GetStats()
	resourceTypeCounts, _ := baseStats["resource_type_counts"].(map[string]int)
	if resourceTypeCounts == nil {
		resourceTypeCounts = map[string]int{}
	}

	services := make([]map[string]interface{}, 0)
	for rt, count := range resourceTypeCounts {
		services = append(services, map[string]interface{}{
			"service": rt,
			"count":   count,
		})
	}

	breakdown := map[string]interface{}{
		"by_provider": h.getProviderCounts(),
		"by_service":  services,
		"by_severity": baseStats["severity_counts"],
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: breakdown})
}

func (h *AnalyticsHandler) getProviderCounts() map[string]int {
	return map[string]int{
		"aws":   0,
		"gcp":   0,
		"azure": 0,
	}
}
