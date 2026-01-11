package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/aws"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// DiscoveryHandler handles AWS resource discovery and drift detection
type DiscoveryHandler struct {
	stateManager *terraform.StateManager
}

// NewDiscoveryHandler creates a new discovery handler
func NewDiscoveryHandler(stateManager *terraform.StateManager) *DiscoveryHandler {
	return &DiscoveryHandler{
		stateManager: stateManager,
	}
}

// DiscoverAWSResources triggers AWS resource discovery
// GET /api/v1/discovery/scan?region=us-east-1
func (h *DiscoveryHandler) DiscoverAWSResources(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "us-east-1"
	}

	log.Infof("Starting AWS resource discovery for region: %s", region)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Create discovery client
	discoveryClient, err := aws.NewDiscoveryClient(ctx, region)
	if err != nil {
		log.Errorf("Failed to create discovery client: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create discovery client: %v", err))
		return
	}

	// Discover all AWS resources
	awsResources, err := discoveryClient.DiscoverAll(ctx)
	if err != nil {
		log.Errorf("Failed to discover AWS resources: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover AWS resources: %v", err))
		return
	}

	log.Infof("Discovered %d AWS resources", len(awsResources))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"region":          region,
			"total_resources": len(awsResources),
			"resources":       awsResources,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
	})
}

// DetectDrift compares Terraform state with actual AWS resources
// GET /api/v1/discovery/drift?region=us-east-1
func (h *DiscoveryHandler) DetectDrift(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "us-east-1"
	}

	log.Infof("Starting drift detection for region: %s", region)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Get Terraform state
	tfResources := h.stateManager.GetAllResources()
	log.Infof("Loaded %d resources from Terraform state", len(tfResources))

	// Create discovery client
	discoveryClient, err := aws.NewDiscoveryClient(ctx, region)
	if err != nil {
		log.Errorf("Failed to create discovery client: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create discovery client: %v", err))
		return
	}

	// Discover all AWS resources
	awsResources, err := discoveryClient.DiscoverAll(ctx)
	if err != nil {
		log.Errorf("Failed to discover AWS resources: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover AWS resources: %v", err))
		return
	}

	log.Infof("Discovered %d AWS resources", len(awsResources))

	// Compare Terraform state with actual AWS state
	driftResult := aws.CompareStateWithActual(tfResources, awsResources)

	log.Infof("Drift detection complete: %d unmanaged, %d missing, %d modified",
		len(driftResult.UnmanagedResources),
		len(driftResult.MissingResources),
		len(driftResult.ModifiedResources))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"region":    region,
			"timestamp": time.Now().Format(time.RFC3339),
			"summary": map[string]interface{}{
				"terraform_resources": len(tfResources),
				"aws_resources":       len(awsResources),
				"unmanaged_count":     len(driftResult.UnmanagedResources),
				"missing_count":       len(driftResult.MissingResources),
				"modified_count":      len(driftResult.ModifiedResources),
			},
			"drift": driftResult,
		},
	})
}

// GetDriftSummary returns a summary of drift without full resource details
// GET /api/v1/discovery/drift/summary?region=us-east-1
func (h *DiscoveryHandler) GetDriftSummary(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "us-east-1"
	}

	log.Infof("Getting drift summary for region: %s", region)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Get Terraform state
	tfResources := h.stateManager.GetAllResources()

	// Create discovery client
	discoveryClient, err := aws.NewDiscoveryClient(ctx, region)
	if err != nil {
		log.Errorf("Failed to create discovery client: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create discovery client: %v", err))
		return
	}

	// Discover all AWS resources
	awsResources, err := discoveryClient.DiscoverAll(ctx)
	if err != nil {
		log.Errorf("Failed to discover AWS resources: %v", err)
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to discover AWS resources: %v", err))
		return
	}

	// Compare Terraform state with actual AWS state
	driftResult := aws.CompareStateWithActual(tfResources, awsResources)

	// Build resource type breakdown
	unmanagedByType := make(map[string]int)
	for _, res := range driftResult.UnmanagedResources {
		unmanagedByType[res.Type]++
	}

	missingByType := make(map[string]int)
	for _, res := range driftResult.MissingResources {
		missingByType[res.Type]++
	}

	modifiedByType := make(map[string]int)
	for _, res := range driftResult.ModifiedResources {
		modifiedByType[res.ResourceType]++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"region":    region,
			"timestamp": time.Now().Format(time.RFC3339),
			"counts": map[string]interface{}{
				"terraform_resources": len(tfResources),
				"aws_resources":       len(awsResources),
				"unmanaged":           len(driftResult.UnmanagedResources),
				"missing":             len(driftResult.MissingResources),
				"modified":            len(driftResult.ModifiedResources),
			},
			"breakdown": map[string]interface{}{
				"unmanaged_by_type": unmanagedByType,
				"missing_by_type":   missingByType,
				"modified_by_type":  modifiedByType,
			},
		},
	})
}
