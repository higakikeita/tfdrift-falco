package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// StateHandler handles Terraform state-related requests
type StateHandler struct {
	stateManager *terraform.StateManager
}

// NewStateHandler creates a new state handler
func NewStateHandler(stateManager *terraform.StateManager) *StateHandler {
	return &StateHandler{
		stateManager: stateManager,
	}
}

// GetState handles GET /api/v1/state
func (h *StateHandler) GetState(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/state")

	metadata := h.stateManager.GetStateMetadata()
	if metadata == nil {
		respondError(w, http.StatusNotFound, "Terraform state not loaded")
		return
	}

	resourceCount := h.stateManager.ResourceCount()

	summary := map[string]interface{}{
		"version":           metadata.Version,
		"terraform_version": metadata.TerraformVersion,
		"serial":            metadata.Serial,
		"lineage":           metadata.Lineage,
		"resource_count":    resourceCount,
		"outputs_count":     0, // Outputs not currently tracked
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    summary,
	})
}

// GetResources handles GET /api/v1/state/resources
func (h *StateHandler) GetResources(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/state/resources")

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	params := models.NewPaginationParams(page, limit)

	// Get all resources
	resources := h.stateManager.GetAllResources()
	allResources := make([]map[string]interface{}, 0, len(resources))
	for _, resource := range resources {
		allResources = append(allResources, map[string]interface{}{
			"type":       resource.Type,
			"name":       resource.Name,
			"provider":   resource.Provider,
			"mode":       resource.Mode,
			"attributes": resource.Attributes,
		})
	}

	// Apply pagination
	total := len(allResources)
	start := params.Offset()
	end := start + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedResources := allResources[start:end]

	response := models.PaginatedResponse{
		Data:       paginatedResources,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: models.CalculateTotalPages(total, params.Limit),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetResource handles GET /api/v1/state/resource/:id
func (h *StateHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	resourceID := chi.URLParam(r, "id")
	log.Debugf("GET /api/v1/state/resource/%s", resourceID)

	resource, exists := h.stateManager.GetResource(resourceID)
	if !exists {
		respondError(w, http.StatusNotFound, "Resource not found")
		return
	}

	resourceData := map[string]interface{}{
		"type":       resource.Type,
		"name":       resource.Name,
		"provider":   resource.Provider,
		"mode":       resource.Mode,
		"attributes": resource.Attributes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    resourceData,
	})
}

// respondError sends an error response
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}
