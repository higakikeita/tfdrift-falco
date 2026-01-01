package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// StatsHandler handles statistics-related requests
type StatsHandler struct {
	store *graph.Store
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(store *graph.Store) *StatsHandler {
	return &StatsHandler{
		store: store,
	}
}

// GetStats handles GET /api/v1/stats
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/stats")

	// Get basic stats from graph store
	baseStats := h.store.GetStats()

	// Get graph structure stats
	graphData := h.store.BuildGraph()

	// Compile comprehensive statistics
	stats := map[string]interface{}{
		// Graph structure
		"graph": map[string]interface{}{
			"total_nodes": len(graphData.Nodes),
			"total_edges": len(graphData.Edges),
		},

		// Drifts
		"drifts": map[string]interface{}{
			"total":           baseStats["total_drifts"],
			"severity_counts": baseStats["severity_counts"],
			"resource_types":  baseStats["resource_type_counts"],
		},

		// Events
		"events": map[string]interface{}{
			"total": baseStats["total_events"],
		},

		// Unmanaged resources
		"unmanaged": map[string]interface{}{
			"total": baseStats["total_unmanaged"],
		},

		// Severity breakdown
		"severity_breakdown": h.calculateSeverityPercentages(baseStats["severity_counts"].(map[string]int)),

		// Top resource types with drifts
		"top_resource_types": h.getTopResourceTypes(baseStats["resource_type_counts"].(map[string]int), 5),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// calculateSeverityPercentages calculates percentage distribution of severities
func (h *StatsHandler) calculateSeverityPercentages(severityCounts map[string]int) map[string]float64 {
	total := 0
	for _, count := range severityCounts {
		total += count
	}

	if total == 0 {
		return map[string]float64{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
		}
	}

	percentages := make(map[string]float64)
	for severity, count := range severityCounts {
		percentages[severity] = float64(count) / float64(total) * 100
	}

	// Ensure all severities are present
	for _, severity := range []string{"critical", "high", "medium", "low"} {
		if _, exists := percentages[severity]; !exists {
			percentages[severity] = 0
		}
	}

	return percentages
}

// getTopResourceTypes returns the top N resource types by count
func (h *StatsHandler) getTopResourceTypes(resourceTypeCounts map[string]int, topN int) []map[string]interface{} {
	// Convert map to slice of key-value pairs
	type kv struct {
		ResourceType string
		Count        int
	}

	var sorted []kv
	for resourceType, count := range resourceTypeCounts {
		sorted = append(sorted, kv{resourceType, count})
	}

	// Sort by count (descending)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Count > sorted[i].Count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Take top N
	if len(sorted) > topN {
		sorted = sorted[:topN]
	}

	// Convert to result format
	result := make([]map[string]interface{}, len(sorted))
	for i, item := range sorted {
		result[i] = map[string]interface{}{
			"resource_type": item.ResourceType,
			"count":         item.Count,
		}
	}

	return result
}
