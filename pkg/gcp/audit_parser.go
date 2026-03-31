// Package gcp provides Google Cloud Platform audit log parsing and resource mapping
// for TFDrift-Falco. It processes GCP Audit Logs received via Falco's gcpaudit plugin
// and maps them to Terraform resource types for drift detection.
//
// The package supports 200+ GCP event types across 25+ services including:
//   - Compute Engine (instances, firewalls, networks, disks, VPN, load balancers, security policies)
//   - Cloud Storage (buckets, bucket IAM)
//   - Cloud SQL (database instances, databases, users)
//   - IAM (project-level policies, service accounts, service account keys)
//   - GKE (clusters, node pools)
//   - Cloud Functions (v1 and v2), Cloud Run, Pub/Sub
//   - BigQuery, Cloud KMS, Secret Manager
//   - Dataproc, Dataflow, Cloud Spanner, Firestore, Redis
//   - Cloud DNS, Cloud Logging, Cloud Monitoring
//   - App Engine, Cloud Armor / VPC Service Controls
//   - Artifact Registry, Cloud Composer, Cloud Tasks, Cloud Scheduler
//   - Cloud Build, Cloud Endpoints, Vertex AI, and more
//
// Example usage:
//
//	parser := gcp.NewAuditParser()
//	event := parser.Parse(falcoResponse)
//	if event != nil {
//	    // Process drift detection event
//	}
package gcp

import (
	"fmt"
	"strings"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/parser"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// AuditParser parses GCP Audit Log events from Falco into TFDrift events.
//
// The parser extracts relevant information from Falco's gcpaudit plugin output,
// including resource identifiers, user identity, project/zone/region information,
// and maps GCP method names to corresponding Terraform resource types.
//
// Thread-safe: Multiple goroutines can safely call Parse concurrently.
type AuditParser struct {
	mapper *ResourceMapper
	config *ResourceConfig
	base   *parser.BaseEventParser
}

// NewAuditParser creates a new GCP Audit Log parser with pre-initialized resource mappings.
//
// The parser is configured with 100+ event-to-resource mappings covering major GCP services.
// Returns a ready-to-use parser instance that can process Falco gcpaudit events.
func NewAuditParser() *AuditParser {
	mapper := NewResourceMapper()
	ap := &AuditParser{
		mapper: mapper,
		config: mapper.config,
	}
	ap.base = parser.NewBaseEventParser(ap.createConfig())
	return ap
}

// Parse converts a Falco output response into a TFDrift event for drift detection.
//
// The method performs the following operations:
//   - Validates the response is from the gcpaudit source
//   - Extracts GCP method name (e.g., "compute.instances.setMetadata")
//   - Filters irrelevant events (read-only operations, non-infrastructure changes)
//   - Maps the method to a Terraform resource type (e.g., "google_compute_instance")
//   - Extracts resource identifiers, project/zone/region, and user identity
//   - Parses event-specific changes from the audit log
//
// Parameters:
//   - res: Falco output response containing GCP audit log data
//
// Returns:
//   - *types.Event: Parsed drift detection event, or nil if:
//   - Response is nil
//   - Source is not "gcpaudit"
//   - Event is not relevant for drift detection
//   - Required fields are missing
//   - No resource mapping exists for the event type
//
// Example:
//
//	parser := NewAuditParser()
//	event := parser.Parse(falcoResponse)
//	if event != nil {
//	    fmt.Printf("Drift detected: %s on %s.%s\n",
//	        event.EventName, event.ResourceType, event.ResourceID)
//	}
func (p *AuditParser) Parse(res *outputs.Response) *types.Event {
	event := p.base.Parse(res)
	if event == nil {
		return nil
	}

	// Set deprecated fields for backward compatibility
	if event.Metadata != nil {
		if region, ok := event.Metadata["region"]; ok {
			event.Region = region
		}
		if projectID, ok := event.Metadata["project_id"]; ok {
			event.ProjectID = projectID
		}
		if serviceName, ok := event.Metadata["service_name"]; ok {
			event.ServiceName = serviceName
		}
	}

	return event
}

// createConfig creates the parser configuration for the BaseEventParser
func (p *AuditParser) createConfig() parser.EventParserConfig {
	return parser.EventParserConfig{
		Provider:       "gcp",
		ExpectedSource: "gcpaudit",
		ExtractEventName: func(fields map[string]string) string {
			return parser.GetStringField(fields, "gcp.methodName")
		},
		IsRelevantEvent: p.isRelevantEvent,
		ExtractResourceID: func(eventName string, fields map[string]string) string {
			resourceName := parser.GetStringField(fields, "gcp.resource.name")
			if resourceName == "" {
				return ""
			}
			return p.extractResourceIDFromName(resourceName)
		},
		MapResourceType: func(eventName string, fields map[string]string) string {
			return p.mapper.MapEventToResource(eventName)
		},
		ExtractUserIdentity: func(fields map[string]string) types.UserIdentity {
			resourceName := parser.GetStringField(fields, "gcp.resource.name")
			projectID := p.extractProjectIDFromName(resourceName, fields)
			return types.UserIdentity{
				Type:      "ServiceAccount",
				UserName:  parser.GetStringField(fields, "gcp.authenticationInfo.principalEmail"),
				AccountID: projectID,
			}
		},
		ExtractChanges: p.extractChanges,
		ExtractMetadata: func(eventName string, fields map[string]string) map[string]string {
			resourceName := parser.GetStringField(fields, "gcp.resource.name")
			projectID := p.extractProjectIDFromName(resourceName, fields)
			zone := p.extractZoneFromName(resourceName, fields)
			region := p.extractRegionFromZone(zone)

			metadata := make(map[string]string)
			if region != "" {
				metadata["region"] = region
			}
			if projectID != "" {
				metadata["project_id"] = projectID
			}
			if serviceName := parser.GetStringField(fields, "gcp.serviceName"); serviceName != "" {
				metadata["service_name"] = serviceName
			}
			return metadata
		},
	}
}

// isRelevantEvent checks if a GCP Audit Log event is relevant for drift detection
func (p *AuditParser) isRelevantEvent(methodName string) bool {
	if p.config != nil {
		return p.config.IsRelevantEvent(methodName)
	}
	// Fallback if config is not available (should not happen in normal operation)
	return false
}

// extractResourceIDFromName extracts the resource ID from GCP resource name
// Example: "projects/123/zones/us-central1-a/instances/vm-1" -> "vm-1"
func (p *AuditParser) extractResourceIDFromName(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractProjectIDFromName extracts the project ID from resource name or fields
// Example: "projects/my-project-123/zones/..." -> "my-project-123"
func (p *AuditParser) extractProjectIDFromName(resourceName string, fields map[string]string) string {
	// Try to extract from resource name
	if strings.HasPrefix(resourceName, "projects/") {
		parts := strings.Split(resourceName, "/")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Fallback to fields
	return parser.GetStringField(fields, "gcp.resource.labels.project_id")
}

// extractZoneFromName extracts the zone from resource name
// Example: "projects/123/zones/us-central1-a/instances/vm-1" -> "us-central1-a"
func (p *AuditParser) extractZoneFromName(resourceName string, fields map[string]string) string {
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
	return parser.GetStringField(fields, "gcp.resource.labels.zone")
}

// extractRegionFromZone extracts region from zone
// Example: "us-central1-a" -> "us-central1"
func (p *AuditParser) extractRegionFromZone(zone string) string {
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
func (p *AuditParser) extractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	// Extract request and response
	request := parser.GetStringField(fields, "gcp.request")
	response := parser.GetStringField(fields, "gcp.response")

	// Event-specific extraction
	switch {
	case strings.HasSuffix(eventName, ".setMetadata"):
		// Extract metadata changes
		if request != "" {
			changes["metadata"] = request
		}

	case strings.HasSuffix(eventName, ".setLabels"):
		// Extract label changes
		if request != "" {
			changes["labels"] = request
		}

	case strings.HasSuffix(eventName, ".setTags"):
		// Extract tag changes
		if request != "" {
			changes["tags"] = request
		}

	case strings.HasSuffix(eventName, ".setMachineType"):
		// Extract machine type change
		if request != "" {
			changes["machine_type"] = request
		}

	case strings.HasSuffix(eventName, ".setServiceAccount"):
		// Extract service account change
		if request != "" {
			changes["service_account"] = request
		}

	case strings.HasSuffix(eventName, ".setDeletionProtection"):
		// Extract deletion protection change
		if request != "" {
			changes["deletion_protection"] = request
		}

	case eventName == "SetIamPolicy" || strings.HasSuffix(eventName, ".SetIamPolicy"):
		// Extract IAM policy changes
		if request != "" {
			changes["policy"] = request
		}

	case strings.Contains(eventName, "UpdateDatabaseDdl"):
		// Extract Spanner DDL updates
		if request != "" {
			changes["ddl"] = request
		}

	case strings.Contains(eventName, "UpdatePolicy") || strings.Contains(eventName, "SetIamPolicy"):
		// Extract policy/binding changes
		if request != "" {
			changes["bindings"] = request
		}

	case strings.Contains(eventName, "changes.create") || strings.Contains(eventName, "dns.changes"):
		// Extract DNS record changes (rrdata)
		if request != "" {
			changes["rrdata"] = request
		}

	case strings.Contains(eventName, "SecurityPolicy") || strings.Contains(eventName, "securityPolicies"):
		// Extract security policy rule changes
		if request != "" {
			changes["rules"] = request
		}

	case strings.HasSuffix(eventName, ".ModifyPushConfig"):
		// Extract Pub/Sub push config changes
		if request != "" {
			changes["push_config"] = request
		}

	case strings.Contains(eventName, ".insert") || strings.Contains(eventName, ".create") ||
		strings.Contains(eventName, ".Create") || strings.Contains(eventName, "Create"):
		// Resource creation (covers compute .insert, gRPC Create* methods, Dataproc CreateCluster, etc.)
		changes["_action"] = "create"
		if response != "" {
			changes["_created_resource"] = response
		}

	case strings.Contains(eventName, ".delete") || strings.Contains(eventName, ".Drop") ||
		strings.Contains(eventName, ".Delete") || strings.Contains(eventName, "Delete"):
		// Resource deletion (covers compute .delete, gRPC Delete* methods, Redis DeleteInstance, etc.)
		changes["_action"] = "delete"

	case strings.Contains(eventName, ".update") || strings.Contains(eventName, ".patch") ||
		strings.Contains(eventName, ".Update") || strings.Contains(eventName, "Update"):
		// Resource update (covers compute .update/.patch, gRPC Update* methods)
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
