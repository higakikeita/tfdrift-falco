package gcp

import (
	"fmt"
	"strings"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// AuditParser parses GCP Audit Log events from Falco
type AuditParser struct {
	mapper *ResourceMapper
}

// NewAuditParser creates a new GCP Audit Log parser
func NewAuditParser() *AuditParser {
	return &AuditParser{
		mapper: NewResourceMapper(),
	}
}

// Parse parses a Falco output response into a TFDrift event
func (p *AuditParser) Parse(res *outputs.Response) *types.Event {
	// Handle nil response
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	// Check if this is a GCP Audit Log event
	if res.Source != "gcpaudit" {
		return nil
	}

	// Parse output fields
	fields := res.OutputFields

	// Extract GCP method name (equivalent to CloudTrail event name)
	methodName, ok := fields["gcp.methodName"]
	if !ok || methodName == "" {
		log.Warnf("Missing gcp.methodName in Falco output")
		return nil
	}

	// Check if this is a relevant event for drift detection
	if !p.isRelevantEvent(methodName) {
		log.Debugf("Event %s is not relevant for drift detection", methodName)
		return nil
	}

	// Extract resource information
	resourceName, ok := fields["gcp.resource.name"]
	if !ok || resourceName == "" {
		log.Debugf("Missing gcp.resource.name for event %s", methodName)
		return nil
	}

	// Extract resource ID from resource name
	resourceID := p.extractResourceID(resourceName)
	if resourceID == "" {
		log.Debugf("Could not extract resource ID from %s", resourceName)
		return nil
	}

	// Map event to Terraform resource type
	resourceType := p.mapper.MapEventToResource(methodName)
	if resourceType == "" {
		log.Debugf("No resource mapping for event %s", methodName)
		return nil
	}

	// Extract project ID
	projectID := p.extractProjectID(resourceName, fields)

	// Extract zone/region
	zone := p.extractZone(resourceName, fields)
	region := p.extractRegion(zone)

	// Extract user identity
	userIdentity := types.UserIdentity{
		Type:      "ServiceAccount", // GCP typically uses service accounts
		UserName:  getStringField(fields, "gcp.authenticationInfo.principalEmail"),
		AccountID: projectID,
	}

	// Extract service name
	serviceName := getStringField(fields, "gcp.serviceName")

	// Extract changes based on event type
	changes := p.extractChanges(methodName, fields)

	return &types.Event{
		Provider:     "gcp",
		EventName:    methodName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Region:       region,
		ProjectID:    projectID,
		ServiceName:  serviceName,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     res,
	}
}

// isRelevantEvent checks if a GCP Audit Log event is relevant for drift detection
func (p *AuditParser) isRelevantEvent(methodName string) bool {
	relevantEvents := map[string]bool{
		// Compute Engine - Instances (Phase 1)
		"compute.instances.insert":          true,
		"compute.instances.delete":          true,
		"compute.instances.setMetadata":     true,
		"compute.instances.setLabels":       true,
		"compute.instances.setTags":         true,
		"compute.instances.setMachineType":  true,
		"compute.instances.setServiceAccount": true,
		"compute.instances.setDeletionProtection": true,

		// Compute Engine - Firewall (Phase 1)
		"compute.firewalls.insert": true,
		"compute.firewalls.delete": true,
		"compute.firewalls.update": true,
		"compute.firewalls.patch":  true,

		// Compute Engine - Networks (Phase 1)
		"compute.networks.insert": true,
		"compute.networks.delete": true,
		"compute.networks.patch":  true,

		// Compute Engine - Subnetworks (Phase 1)
		"compute.subnetworks.insert": true,
		"compute.subnetworks.delete": true,
		"compute.subnetworks.patch":  true,

		// IAM (Phase 1)
		"SetIamPolicy": true,

		// Cloud Storage - Buckets (Phase 2)
		"storage.buckets.create": true,
		"storage.buckets.delete": true,
		"storage.buckets.update": true,
		"storage.buckets.patch":  true,

		// Cloud SQL - Instances (Phase 2)
		"cloudsql.instances.create": true,
		"cloudsql.instances.delete": true,
		"cloudsql.instances.update": true,
		"cloudsql.instances.patch":  true,

		// Compute Engine - Disks (Phase 2)
		"compute.disks.insert": true,
		"compute.disks.delete": true,

		// GKE - Clusters (Phase 3)
		"container.clusters.create": true,
		"container.clusters.delete": true,
		"container.clusters.update": true,

		// Cloud Run - Services (Phase 3)
		"run.services.create": true,
		"run.services.delete": true,
		"run.services.update": true,

		// Cloud Functions (Phase 3)
		"cloudfunctions.functions.create": true,
		"cloudfunctions.functions.delete": true,
		"cloudfunctions.functions.update": true,
	}

	return relevantEvents[methodName]
}

// extractResourceID extracts the resource ID from GCP resource name
// Example: "projects/123/zones/us-central1-a/instances/vm-1" -> "vm-1"
func (p *AuditParser) extractResourceID(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractProjectID extracts the project ID from resource name or fields
// Example: "projects/my-project-123/zones/..." -> "my-project-123"
func (p *AuditParser) extractProjectID(resourceName string, fields map[string]string) string {
	// Try to extract from resource name
	if strings.HasPrefix(resourceName, "projects/") {
		parts := strings.Split(resourceName, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Fallback to fields
	return getStringField(fields, "gcp.resource.labels.project_id")
}

// extractZone extracts the zone from resource name
// Example: "projects/123/zones/us-central1-a/instances/vm-1" -> "us-central1-a"
func (p *AuditParser) extractZone(resourceName string, fields map[string]string) string {
	// Try to extract from resource name
	if strings.Contains(resourceName, "/zones/") {
		parts := strings.Split(resourceName, "/zones/")
		if len(parts) >= 2 {
			zoneParts := strings.Split(parts[1], "/")
			if len(zoneParts) > 0 {
				return zoneParts[0]
			}
		}
	}

	// Fallback to fields
	return getStringField(fields, "gcp.resource.labels.zone")
}

// extractRegion extracts region from zone
// Example: "us-central1-a" -> "us-central1"
func (p *AuditParser) extractRegion(zone string) string {
	if zone == "" {
		return ""
	}

	// Zone format: {region}-{zone_letter} (e.g., us-central1-a)
	parts := strings.Split(zone, "-")
	if len(parts) >= 3 {
		return strings.Join(parts[:len(parts)-1], "-")
	}

	return zone
}

// extractChanges extracts attribute changes from GCP Audit Log
func (p *AuditParser) extractChanges(methodName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	// Extract request and response
	request := getStringField(fields, "gcp.request")
	response := getStringField(fields, "gcp.response")

	// Event-specific extraction
	switch {
	case strings.HasSuffix(methodName, ".setMetadata"):
		// Extract metadata changes
		if request != "" {
			changes["metadata"] = request
		}

	case strings.HasSuffix(methodName, ".setLabels"):
		// Extract label changes
		if request != "" {
			changes["labels"] = request
		}

	case strings.HasSuffix(methodName, ".setTags"):
		// Extract tag changes
		if request != "" {
			changes["tags"] = request
		}

	case strings.HasSuffix(methodName, ".setMachineType"):
		// Extract machine type change
		if request != "" {
			changes["machine_type"] = request
		}

	case strings.HasSuffix(methodName, ".setServiceAccount"):
		// Extract service account change
		if request != "" {
			changes["service_account"] = request
		}

	case strings.HasSuffix(methodName, ".setDeletionProtection"):
		// Extract deletion protection change
		if request != "" {
			changes["deletion_protection"] = request
		}

	case methodName == "SetIamPolicy":
		// Extract IAM policy changes
		if request != "" {
			changes["policy"] = request
		}

	case strings.Contains(methodName, ".insert") || strings.Contains(methodName, ".create"):
		// Resource creation
		changes["_action"] = "create"
		if response != "" {
			changes["_created_resource"] = response
		}

	case strings.Contains(methodName, ".delete"):
		// Resource deletion
		changes["_action"] = "delete"

	case strings.Contains(methodName, ".update") || strings.Contains(methodName, ".patch"):
		// Resource update
		changes["_action"] = "update"
		if request != "" {
			changes["_update_request"] = request
		}
	}

	// Always include raw request/response for debugging
	if request != "" {
		changes["_raw_request"] = request
	}
	if response != "" {
		changes["_raw_response"] = response
	}

	return changes
}

// getStringField safely gets a string field from Falco output fields
func getStringField(fields map[string]string, key string) string {
	if val, ok := fields[key]; ok {
		return val
	}
	return ""
}

// ValidateEvent performs validation on parsed event
func (p *AuditParser) ValidateEvent(event *types.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.Provider != "gcp" {
		return fmt.Errorf("invalid provider: %s (expected: gcp)", event.Provider)
	}

	if event.EventName == "" {
		return fmt.Errorf("event name is empty")
	}

	if event.ResourceType == "" {
		return fmt.Errorf("resource type is empty")
	}

	if event.ResourceID == "" {
		return fmt.Errorf("resource ID is empty")
	}

	return nil
}
