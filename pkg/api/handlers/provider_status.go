package handlers

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/provider"
)

// ProviderStatusHandler handles provider status and health endpoints.
// It tracks real-time event counts, last event timestamps, and uptime
// for each registered cloud provider (AWS, GCP, Azure).
type ProviderStatusHandler struct {
	registry *provider.Registry
	mu       sync.RWMutex
	stats    map[string]*ProviderStats
	startAt  time.Time
}

// ProviderStats holds runtime statistics for a single provider.
type ProviderStats struct {
	EventsReceived int64     `json:"events_received"`
	EventsMatched  int64     `json:"events_matched"`
	LastEventAt    time.Time `json:"last_event_at,omitempty"`
	ErrorCount     int64     `json:"error_count"`
	Status         string    `json:"status"` // "active", "idle", "error"
}

// NewProviderStatusHandler creates a new provider status handler.
func NewProviderStatusHandler(registry *provider.Registry) *ProviderStatusHandler {
	h := &ProviderStatusHandler{
		registry: registry,
		stats:    make(map[string]*ProviderStats),
		startAt:  time.Now(),
	}
	// Initialize stats for all registered providers
	for _, name := range []string{"aws", "gcp", "azure"} {
		if _, ok := registry.Get(name); ok {
			h.stats[name] = &ProviderStats{Status: "idle"}
		}
	}
	return h
}

// RecordEvent records an event for a provider (called from event pipeline).
func (h *ProviderStatusHandler) RecordEvent(providerName string, matched bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.stats[providerName]
	if !ok {
		s = &ProviderStats{Status: "idle"}
		h.stats[providerName] = s
	}
	s.EventsReceived++
	if matched {
		s.EventsMatched++
	}
	s.LastEventAt = time.Now()
	s.Status = "active"
}

// RecordError records an error for a provider.
func (h *ProviderStatusHandler) RecordError(providerName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.stats[providerName]
	if !ok {
		s = &ProviderStats{}
		h.stats[providerName] = s
	}
	s.ErrorCount++
	s.Status = "error"
}

// GetProviderStatus returns status for all providers.
// GET /api/v1/providers/status
func (h *ProviderStatusHandler) GetProviderStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	capabilities := h.registry.GetAllCapabilities()

	result := make([]map[string]interface{}, 0)
	for name, caps := range capabilities {
		p, ok := h.registry.Get(name)
		if !ok {
			continue
		}

		stats := h.stats[name]
		if stats == nil {
			stats = &ProviderStats{Status: "idle"}
		}

		// Determine health status
		status := stats.Status
		if stats.ErrorCount > 0 && stats.EventsReceived == 0 {
			status = "error"
		} else if !stats.LastEventAt.IsZero() && time.Since(stats.LastEventAt) < 5*time.Minute {
			status = "active"
		} else if stats.EventsReceived > 0 {
			status = "idle"
		}

		entry := map[string]interface{}{
			"name":            name,
			"status":          status,
			"event_count":     p.SupportedEventCount(),
			"resource_types":  len(p.SupportedResourceTypes()),
			"has_discovery":   caps.Discovery,
			"has_comparison":  caps.Comparison,
			"events_received": stats.EventsReceived,
			"events_matched":  stats.EventsMatched,
			"error_count":     stats.ErrorCount,
			"uptime_seconds":  int64(time.Since(h.startAt).Seconds()),
		}

		if !stats.LastEventAt.IsZero() {
			entry["last_event_at"] = stats.LastEventAt.Format(time.RFC3339)
			entry["seconds_since_last_event"] = int64(time.Since(stats.LastEventAt).Seconds())
		}

		result = append(result, entry)
	}

	// Sort by name for deterministic output
	sort.Slice(result, func(i, j int) bool {
		nameI, _ := result[i]["name"].(string)
		nameJ, _ := result[j]["name"].(string)
		return nameI < nameJ
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": result,
		"count":     len(result),
		"uptime":    int64(time.Since(h.startAt).Seconds()),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetProviderSummary returns a compact summary for dashboard widgets.
// GET /api/v1/providers/summary
func (h *ProviderStatusHandler) GetProviderSummary(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	capabilities := h.registry.GetAllCapabilities()

	var totalEvents, totalMatched, totalErrors int64
	activeCount := 0
	providerNames := make([]string, 0)

	for name := range capabilities {
		providerNames = append(providerNames, name)
		if s, ok := h.stats[name]; ok {
			totalEvents += s.EventsReceived
			totalMatched += s.EventsMatched
			totalErrors += s.ErrorCount
			if s.Status == "active" {
				activeCount++
			}
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_providers":    len(capabilities),
		"active_providers":   activeCount,
		"total_events":       totalEvents,
		"total_matched":      totalMatched,
		"total_errors":       totalErrors,
		"providers":          providerNames,
		"match_rate_percent": matchRate(totalEvents, totalMatched),
		"uptime_seconds":     int64(time.Since(h.startAt).Seconds()),
	})
}

func matchRate(total, matched int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(matched) / float64(total) * 100
}
