package handlers

import (
	"net/http"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/provider"
)

// ProvidersHandler handles provider status and capability endpoints (v0.6.0)
type ProvidersHandler struct {
	registry *provider.Registry
}

// NewProvidersHandler creates a new providers handler
func NewProvidersHandler(registry *provider.Registry) *ProvidersHandler {
	return &ProvidersHandler{
		registry: registry,
	}
}

// GetProviders returns all registered providers with their capabilities
// GET /api/v1/providers
func (h *ProvidersHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	capabilities := h.registry.GetAllCapabilities()

	providers := make([]map[string]interface{}, 0)
	for name, caps := range capabilities {
		p, ok := h.registry.Get(name)
		if !ok {
			continue
		}

		providerInfo := map[string]interface{}{
			"name":           name,
			"event_count":    p.SupportedEventCount(),
			"resource_types": p.SupportedResourceTypes(),
			"has_discovery":  caps.Discovery,
			"has_comparison": caps.Comparison,
		}

		if caps.Discovery {
			if d, ok := h.registry.GetDiscoverer(name); ok {
				providerInfo["discovery_types"] = d.SupportedDiscoveryTypes()
			}
		}

		providers = append(providers, providerInfo)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetProviderCapabilities returns capabilities for a specific provider
// GET /api/v1/providers/{name}/capabilities
func (h *ProvidersHandler) GetProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		respondError(w, http.StatusBadRequest, "provider name is required")
		return
	}

	p, ok := h.registry.Get(name)
	if !ok {
		respondError(w, http.StatusNotFound, "provider not found: "+name)
		return
	}

	caps := provider.GetCapabilities(p)

	result := map[string]interface{}{
		"name":           name,
		"event_count":    p.SupportedEventCount(),
		"resource_types": p.SupportedResourceTypes(),
		"has_discovery":  caps.Discovery,
		"has_comparison": caps.Comparison,
	}

	if caps.Discovery {
		if d, ok := h.registry.GetDiscoverer(name); ok {
			result["discovery_types"] = d.SupportedDiscoveryTypes()
		}
	}

	respondJSON(w, http.StatusOK, result)
}
