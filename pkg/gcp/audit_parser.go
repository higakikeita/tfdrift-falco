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
	base   *parser.BaseEventParser
}

// NewAuditParser creates a new GCP Audit Log parser with pre-initialized resource mappings.
//
// The parser is configured with 100+ event-to-resource mappings covering major GCP services.
// Returns a ready-to-use parser instance that can process Falco gcpaudit events.
func NewAuditParser() *AuditParser {
	ap := &AuditParser{
		mapper: NewResourceMapper(),
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
	relevantEvents := map[string]bool{
		// Compute Engine - Instances
		"compute.instances.insert":                true,
		"compute.instances.delete":                true,
		"compute.instances.setMetadata":           true,
		"compute.instances.setLabels":             true,
		"compute.instances.setTags":               true,
		"compute.instances.setMachineType":        true,
		"compute.instances.setServiceAccount":     true,
		"compute.instances.setDeletionProtection": true,

		// Compute Engine - Firewall
		"compute.firewalls.insert": true,
		"compute.firewalls.delete": true,
		"compute.firewalls.update": true,
		"compute.firewalls.patch":  true,

		// Compute Engine - Networks
		"compute.networks.insert": true,
		"compute.networks.delete": true,
		"compute.networks.patch":  true,

		// Compute Engine - Subnetworks
		"compute.subnetworks.insert": true,
		"compute.subnetworks.delete": true,
		"compute.subnetworks.patch":  true,

		// Compute Engine - Disks
		"compute.disks.insert": true,
		"compute.disks.delete": true,

		// Compute Engine - Routes
		"compute.routes.insert": true,
		"compute.routes.delete": true,

		// Compute Engine - Routers
		"compute.routers.insert": true,
		"compute.routers.delete": true,
		"compute.routers.patch":  true,
		"compute.routers.update": true,

		// Compute Engine - VPN
		"compute.vpnTunnels.insert":  true,
		"compute.vpnTunnels.delete":  true,
		"compute.vpnGateways.insert": true,
		"compute.vpnGateways.delete": true,

		// Compute Engine - Instance Templates
		"compute.instanceTemplates.insert": true,
		"compute.instanceTemplates.delete": true,

		// Compute Engine - Instance Group Managers
		"compute.instanceGroupManagers.insert": true,
		"compute.instanceGroupManagers.delete": true,
		"compute.instanceGroupManagers.patch":  true,

		// Compute Engine - Autoscalers
		"compute.autoscalers.insert": true,
		"compute.autoscalers.delete": true,
		"compute.autoscalers.patch":  true,

		// Compute Engine - Global Addresses
		"compute.globalAddresses.insert": true,
		"compute.globalAddresses.delete": true,

		// Compute Engine - Addresses
		"compute.addresses.insert": true,
		"compute.addresses.delete": true,

		// Compute Engine - Images
		"compute.images.insert": true,
		"compute.images.delete": true,

		// Compute Engine - Snapshots
		"compute.snapshots.insert": true,
		"compute.snapshots.delete": true,

		// Compute Engine - Interconnect Attachments
		"compute.interconnects.insert": true,
		"compute.interconnects.delete": true,

		// Compute Engine - Global Forwarding Rules
		"compute.globalForwardingRules.insert": true,
		"compute.globalForwardingRules.delete": true,

		// Compute Engine - Target HTTPS Proxies
		"compute.targetHttpsProxies.insert": true,
		"compute.targetHttpsProxies.delete": true,

		// Compute Engine - Target HTTP Proxies
		"compute.targetHttpProxies.insert": true,
		"compute.targetHttpProxies.delete": true,

		// Compute Engine - URL Maps
		"compute.urlMaps.insert": true,
		"compute.urlMaps.delete": true,
		"compute.urlMaps.patch":  true,

		// Compute Engine - SSL Policies
		"compute.sslPolicies.insert": true,
		"compute.sslPolicies.delete": true,
		"compute.sslPolicies.patch":  true,

		// Compute Engine - Backend Services
		"compute.backendServices.insert": true,
		"compute.backendServices.delete": true,
		"compute.backendServices.update": true,
		"compute.backendServices.patch":  true,

		// Compute Engine - Health Checks
		"compute.healthChecks.insert": true,
		"compute.healthChecks.delete": true,
		"compute.healthChecks.update": true,
		"compute.healthChecks.patch":  true,

		// Compute Engine - Target Pools
		"compute.targetPools.insert": true,
		"compute.targetPools.delete": true,

		// Compute Engine - Forwarding Rules
		"compute.forwardingRules.insert": true,
		"compute.forwardingRules.delete": true,

		// Compute Engine - SSL Certificates
		"compute.sslCertificates.insert": true,
		"compute.sslCertificates.delete": true,

		// Compute Engine - Cloud Armor / Security Policies
		"compute.securityPolicies.insert":    true,
		"compute.securityPolicies.delete":    true,
		"compute.securityPolicies.patch":     true,
		"compute.securityPolicies.addRule":   true,
		"compute.securityPolicies.removeRule": true,

		// IAM - Project Level
		"SetIamPolicy": true,

		// IAM - Service Accounts
		"google.iam.admin.v1.CreateServiceAccount": true,
		"google.iam.admin.v1.DeleteServiceAccount": true,
		"google.iam.admin.v1.PatchServiceAccount":  true,

		// IAM - Service Account Keys
		"google.iam.admin.v1.CreateServiceAccountKey": true,
		"google.iam.admin.v1.DeleteServiceAccountKey": true,

		// IAM - Policy
		"google.iam.v1.IAMPolicy.SetIamPolicy": true,

		// Cloud Storage - Buckets
		"storage.buckets.create": true,
		"storage.buckets.delete": true,
		"storage.buckets.update": true,
		"storage.buckets.patch":  true,

		// Cloud Storage - Objects
		"storage.objects.create": true,
		"storage.objects.delete": true,

		// Cloud Storage - IAM
		"storage.buckets.setIamPolicy": true,

		// Cloud SQL - Instances
		"cloudsql.instances.create": true,
		"cloudsql.instances.delete": true,
		"cloudsql.instances.update": true,
		"cloudsql.instances.patch":  true,

		// Cloud SQL - Databases
		"cloudsql.databases.create": true,
		"cloudsql.databases.delete": true,
		"cloudsql.databases.update": true,

		// Cloud SQL - Users
		"cloudsql.users.create": true,
		"cloudsql.users.delete": true,
		"cloudsql.users.update": true,

		// GKE - Clusters
		"container.clusters.create": true,
		"container.clusters.delete": true,
		"container.clusters.update": true,

		// GKE - Node Pools
		"container.projects.zones.clusters.nodePools.create": true,
		"container.projects.zones.clusters.nodePools.delete": true,
		"container.projects.zones.clusters.nodePools.update": true,

		// Cloud Run - Services
		"run.services.create": true,
		"run.services.delete": true,
		"run.services.update": true,

		// Cloud Functions
		"cloudfunctions.functions.create": true,
		"cloudfunctions.functions.delete": true,
		"cloudfunctions.functions.update": true,

		// Cloud Functions v2
		"cloudfunctions.v2.functions.create": true,
		"cloudfunctions.v2.functions.delete": true,
		"cloudfunctions.v2.functions.update": true,

		// Pub/Sub - Topics
		"google.pubsub.v1.Publisher.CreateTopic": true,
		"google.pubsub.v1.Publisher.DeleteTopic": true,
		"google.pubsub.v1.Publisher.UpdateTopic": true,

		// Pub/Sub - Subscriptions
		"google.pubsub.v1.Subscriber.CreateSubscription":   true,
		"google.pubsub.v1.Subscriber.DeleteSubscription":   true,
		"google.pubsub.v1.Subscriber.ModifyPushConfig":     true,
		"google.pubsub.v1.Subscriber.UpdateSubscription":   true,

		// BigQuery - Datasets
		"google.cloud.bigquery.v2.DatasetService.InsertDataset": true,
		"google.cloud.bigquery.v2.DatasetService.DeleteDataset": true,
		"google.cloud.bigquery.v2.DatasetService.PatchDataset":  true,

		// BigQuery - Tables
		"google.cloud.bigquery.v2.TableService.InsertTable": true,
		"google.cloud.bigquery.v2.TableService.DeleteTable": true,
		"google.cloud.bigquery.v2.TableService.PatchTable":  true,

		// Cloud KMS - KeyRings
		"google.cloud.kms.v1.KeyManagementService.CreateKeyRing": true,

		// Cloud KMS - CryptoKeys
		"google.cloud.kms.v1.KeyManagementService.CreateCryptoKey": true,
		"google.cloud.kms.v1.KeyManagementService.UpdateCryptoKey": true,

		// Secret Manager - Secrets
		"google.cloud.secretmanager.v1.SecretManagerService.CreateSecret": true,
		"google.cloud.secretmanager.v1.SecretManagerService.DeleteSecret": true,

		// Dataproc - Clusters
		"google.cloud.dataproc.v1.ClusterController.CreateCluster": true,
		"google.cloud.dataproc.v1.ClusterController.DeleteCluster": true,
		"google.cloud.dataproc.v1.ClusterController.UpdateCluster": true,

		// Dataproc - Jobs
		"google.cloud.dataproc.v1.JobController.SubmitJob": true,
		"google.cloud.dataproc.v1.JobController.DeleteJob": true,

		// Dataflow - Jobs
		"google.dataflow.v1b3.JobsV1Beta3.CreateJob": true,
		"google.dataflow.v1b3.JobsV1Beta3.UpdateJob": true,

		// Cloud Spanner - Instances
		"google.spanner.admin.instance.v1.InstanceAdmin.CreateInstance": true,
		"google.spanner.admin.instance.v1.InstanceAdmin.DeleteInstance": true,
		"google.spanner.admin.instance.v1.InstanceAdmin.UpdateInstance": true,

		// Cloud Spanner - Databases
		"google.spanner.admin.database.v1.DatabaseAdmin.CreateDatabase":     true,
		"google.spanner.admin.database.v1.DatabaseAdmin.DropDatabase":       true,
		"google.spanner.admin.database.v1.DatabaseAdmin.UpdateDatabaseDdl":  true,

		// Firestore - Indexes
		"google.firestore.admin.v1.FirestoreAdmin.CreateIndex": true,
		"google.firestore.admin.v1.FirestoreAdmin.DeleteIndex": true,

		// Firestore - Databases
		"google.firestore.admin.v1.FirestoreAdmin.CreateDatabase": true,
		"google.firestore.admin.v1.FirestoreAdmin.DeleteDatabase": true,

		// Cloud Memorystore / Redis
		"google.cloud.redis.v1.CloudRedis.CreateInstance": true,
		"google.cloud.redis.v1.CloudRedis.DeleteInstance": true,
		"google.cloud.redis.v1.CloudRedis.UpdateInstance": true,

		// Cloud DNS - Managed Zones
		"dns.managedZones.create": true,
		"dns.managedZones.delete": true,
		"dns.managedZones.update": true,
		"dns.managedZones.patch":  true,

		// Cloud DNS - Record Sets
		"dns.changes.create": true,

		// Cloud Logging - Sinks
		"google.logging.v2.ConfigServiceV2.CreateSink": true,
		"google.logging.v2.ConfigServiceV2.DeleteSink": true,
		"google.logging.v2.ConfigServiceV2.UpdateSink": true,

		// Cloud Logging - Buckets
		"google.logging.v2.ConfigServiceV2.CreateBucket": true,
		"google.logging.v2.ConfigServiceV2.DeleteBucket": true,

		// Cloud Monitoring - Alert Policies
		"google.monitoring.v3.AlertPolicyService.CreateAlertPolicy": true,
		"google.monitoring.v3.AlertPolicyService.DeleteAlertPolicy": true,
		"google.monitoring.v3.AlertPolicyService.UpdateAlertPolicy": true,

		// Cloud Monitoring - Uptime Checks
		"google.monitoring.v3.UptimeCheckService.CreateUptimeCheckConfig": true,
		"google.monitoring.v3.UptimeCheckService.DeleteUptimeCheckConfig": true,

		// Cloud Monitoring - Notification Channels
		"google.monitoring.v3.NotificationChannelService.CreateNotificationChannel": true,
		"google.monitoring.v3.NotificationChannelService.DeleteNotificationChannel": true,

		// App Engine - Services
		"google.appengine.v1.Services.UpdateService": true,
		"google.appengine.v1.Services.DeleteService": true,

		// App Engine - Versions
		"google.appengine.v1.Versions.CreateVersion": true,
		"google.appengine.v1.Versions.DeleteVersion": true,

		// App Engine - Firewall
		"google.appengine.v1.Firewall.BatchUpdateIngressRules": true,

		// VPC Service Controls - Service Perimeters
		"google.identity.accesscontextmanager.v1.AccessContextManager.CreateServicePerimeter": true,
		"google.identity.accesscontextmanager.v1.AccessContextManager.DeleteServicePerimeter": true,
		"google.identity.accesscontextmanager.v1.AccessContextManager.UpdateServicePerimeter": true,

		// VPC Service Controls - Access Levels
		"google.identity.accesscontextmanager.v1.AccessContextManager.CreateAccessLevel": true,
		"google.identity.accesscontextmanager.v1.AccessContextManager.DeleteAccessLevel": true,

		// Artifact Registry - Repositories
		"google.devtools.artifactregistry.v1.ArtifactRegistry.CreateRepository": true,
		"google.devtools.artifactregistry.v1.ArtifactRegistry.DeleteRepository": true,
		"google.devtools.artifactregistry.v1.ArtifactRegistry.UpdateRepository": true,

		// Cloud Composer - Environments
		"google.cloud.orchestration.airflow.service.v1.Environments.CreateEnvironment": true,
		"google.cloud.orchestration.airflow.service.v1.Environments.DeleteEnvironment": true,
		"google.cloud.orchestration.airflow.service.v1.Environments.UpdateEnvironment": true,

		// Cloud Tasks - Queues
		"google.cloud.tasks.v2.CloudTasks.CreateQueue": true,
		"google.cloud.tasks.v2.CloudTasks.DeleteQueue": true,
		"google.cloud.tasks.v2.CloudTasks.UpdateQueue": true,

		// Cloud Scheduler - Jobs
		"google.cloud.scheduler.v1.CloudScheduler.CreateJob": true,
		"google.cloud.scheduler.v1.CloudScheduler.DeleteJob": true,
		"google.cloud.scheduler.v1.CloudScheduler.UpdateJob": true,

		// Cloud Build - Build Triggers
		"google.devtools.cloudbuild.v1.CloudBuild.CreateBuildTrigger": true,
		"google.devtools.cloudbuild.v1.CloudBuild.DeleteBuildTrigger": true,
		"google.devtools.cloudbuild.v1.CloudBuild.UpdateBuildTrigger": true,

		// Cloud Endpoints - Services
		"google.api.servicemanagement.v1.ServiceManager.CreateService":      true,
		"google.api.servicemanagement.v1.ServiceManager.DeleteService":      true,
		"google.api.servicemanagement.v1.ServiceManager.SubmitConfigSource": true,

		// Vertex AI - Endpoints
		"google.cloud.aiplatform.v1.EndpointService.CreateEndpoint": true,
		"google.cloud.aiplatform.v1.EndpointService.DeleteEndpoint": true,

		// Vertex AI - Datasets
		"google.cloud.aiplatform.v1.DatasetService.CreateDataset": true,
		"google.cloud.aiplatform.v1.DatasetService.DeleteDataset": true,

		// Vertex AI - Featurestores
		"google.cloud.aiplatform.v1.FeaturestoreService.CreateFeaturestore": true,
		"google.cloud.aiplatform.v1.FeaturestoreService.DeleteFeaturestore": true,
	}

	return relevantEvents[methodName]
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
