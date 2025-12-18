package gcp

// ResourceMapper maps GCP Audit Log method names to Terraform resource types.
//
// The mapper maintains a comprehensive mapping of 100+ GCP Audit Log method names
// (e.g., "compute.instances.setMetadata") to their corresponding Terraform resource
// types (e.g., "google_compute_instance").
//
// This enables TFDrift-Falco to correlate infrastructure changes detected in GCP
// Audit Logs with resources defined in Terraform state files.
//
// Supported services:
//   - Compute Engine: instances, firewalls, networks, disks, VPN, load balancers
//   - Cloud Storage: buckets and bucket IAM
//   - Cloud SQL: database instances
//   - IAM: project policies and service accounts
//   - GKE: clusters and node pools
//   - Cloud Functions, Cloud Run, Pub/Sub, Cloud KMS, Secret Manager, and more
//
// Thread-safe: Multiple goroutines can safely call MapEventToResource concurrently.
type ResourceMapper struct {
	eventToResource map[string]string
}

// NewResourceMapper creates a new resource mapper with pre-initialized mappings.
//
// The mapper is initialized with 100+ event-to-resource mappings covering
// all major GCP services supported by TFDrift-Falco v0.5.0+.
//
// Returns a ready-to-use mapper instance for event translation.
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{
		eventToResource: initializeEventMapping(),
	}
}

// MapEventToResource maps a GCP Audit Log method name to its Terraform resource type.
//
// The method performs case-sensitive lookup of the GCP method name in the
// pre-initialized mapping table and returns the corresponding Terraform resource type.
//
// Parameters:
//   - methodName: GCP Audit Log method name (e.g., "compute.instances.setMetadata",
//     "storage.buckets.update", "SetIamPolicy")
//
// Returns:
//   - string: Terraform resource type (e.g., "google_compute_instance",
//     "google_storage_bucket", "google_project_iam_binding"), or empty string
//     if no mapping exists for the given method name
//
// Example:
//
//	mapper := NewResourceMapper()
//	resourceType := mapper.MapEventToResource("compute.instances.setMetadata")
//	// Returns: "google_compute_instance"
//
//	resourceType = mapper.MapEventToResource("unknown.method.name")
//	// Returns: ""
func (m *ResourceMapper) MapEventToResource(methodName string) string {
	if resourceType, ok := m.eventToResource[methodName]; ok {
		return resourceType
	}
	return ""
}

// initializeEventMapping initializes the GCP method name to Terraform resource type mapping table.
//
// This function creates a comprehensive map of 100+ GCP Audit Log method names
// to their corresponding Terraform resource types. The mappings cover:
//   - 12+ GCP services
//   - Infrastructure mutation operations (create, update, delete, patch, set*)
//   - Resource-specific operations (e.g., setMetadata, setLabels, setTags)
//
// Returns a map[string]string with methodName -> terraformResourceType mappings.
func initializeEventMapping() map[string]string {
	return map[string]string{
		// Compute Engine - Instances
		"compute.instances.insert":                "google_compute_instance",
		"compute.instances.delete":                "google_compute_instance",
		"compute.instances.setMetadata":           "google_compute_instance",
		"compute.instances.setLabels":             "google_compute_instance",
		"compute.instances.setTags":               "google_compute_instance",
		"compute.instances.setMachineType":        "google_compute_instance",
		"compute.instances.setServiceAccount":     "google_compute_instance",
		"compute.instances.setDeletionProtection": "google_compute_instance",
		"compute.instances.start":                 "google_compute_instance",
		"compute.instances.stop":                  "google_compute_instance",
		"compute.instances.reset":                 "google_compute_instance",

		// Compute Engine - Firewall Rules
		"compute.firewalls.insert": "google_compute_firewall",
		"compute.firewalls.delete": "google_compute_firewall",
		"compute.firewalls.update": "google_compute_firewall",
		"compute.firewalls.patch":  "google_compute_firewall",

		// Compute Engine - Networks
		"compute.networks.insert": "google_compute_network",
		"compute.networks.delete": "google_compute_network",
		"compute.networks.patch":  "google_compute_network",

		// Compute Engine - Subnetworks
		"compute.subnetworks.insert":           "google_compute_subnetwork",
		"compute.subnetworks.delete":           "google_compute_subnetwork",
		"compute.subnetworks.patch":            "google_compute_subnetwork",
		"compute.subnetworks.setPrivateIpGoogleAccess": "google_compute_subnetwork",

		// Compute Engine - Disks
		"compute.disks.insert":     "google_compute_disk",
		"compute.disks.delete":     "google_compute_disk",
		"compute.disks.resize":     "google_compute_disk",
		"compute.disks.setLabels":  "google_compute_disk",

		// Compute Engine - Routes
		"compute.routes.insert": "google_compute_route",
		"compute.routes.delete": "google_compute_route",

		// Compute Engine - Routers
		"compute.routers.insert": "google_compute_router",
		"compute.routers.delete": "google_compute_router",
		"compute.routers.patch":  "google_compute_router",
		"compute.routers.update": "google_compute_router",

		// Compute Engine - VPN
		"compute.vpnTunnels.insert": "google_compute_vpn_tunnel",
		"compute.vpnTunnels.delete": "google_compute_vpn_tunnel",
		"compute.vpnGateways.insert": "google_compute_vpn_gateway",
		"compute.vpnGateways.delete": "google_compute_vpn_gateway",

		// IAM - Project Level
		"SetIamPolicy": "google_project_iam_binding",

		// IAM - Service Accounts
		"google.iam.admin.v1.CreateServiceAccount": "google_service_account",
		"google.iam.admin.v1.DeleteServiceAccount": "google_service_account",
		"google.iam.admin.v1.PatchServiceAccount":  "google_service_account",

		// Cloud Storage - Buckets
		"storage.buckets.create": "google_storage_bucket",
		"storage.buckets.delete": "google_storage_bucket",
		"storage.buckets.update": "google_storage_bucket",
		"storage.buckets.patch":  "google_storage_bucket",

		// Cloud Storage - Objects
		"storage.objects.create": "google_storage_bucket_object",
		"storage.objects.delete": "google_storage_bucket_object",

		// Cloud Storage - IAM
		"storage.buckets.setIamPolicy": "google_storage_bucket_iam_binding",

		// Cloud SQL - Instances
		"cloudsql.instances.create": "google_sql_database_instance",
		"cloudsql.instances.delete": "google_sql_database_instance",
		"cloudsql.instances.update": "google_sql_database_instance",
		"cloudsql.instances.patch":  "google_sql_database_instance",

		// Cloud SQL - Databases
		"cloudsql.databases.create": "google_sql_database",
		"cloudsql.databases.delete": "google_sql_database",
		"cloudsql.databases.update": "google_sql_database",

		// Cloud SQL - Users
		"cloudsql.users.create": "google_sql_user",
		"cloudsql.users.delete": "google_sql_user",
		"cloudsql.users.update": "google_sql_user",

		// GKE - Clusters
		"container.clusters.create": "google_container_cluster",
		"container.clusters.delete": "google_container_cluster",
		"container.clusters.update": "google_container_cluster",

		// GKE - Node Pools
		"container.projects.zones.clusters.nodePools.create": "google_container_node_pool",
		"container.projects.zones.clusters.nodePools.delete": "google_container_node_pool",
		"container.projects.zones.clusters.nodePools.update": "google_container_node_pool",

		// Cloud Run - Services
		"run.services.create": "google_cloud_run_service",
		"run.services.delete": "google_cloud_run_service",
		"run.services.update": "google_cloud_run_service",

		// Cloud Functions - Functions
		"cloudfunctions.functions.create": "google_cloudfunctions_function",
		"cloudfunctions.functions.delete": "google_cloudfunctions_function",
		"cloudfunctions.functions.update": "google_cloudfunctions_function",

		// Cloud Functions v2
		"cloudfunctions.v2.functions.create": "google_cloudfunctions2_function",
		"cloudfunctions.v2.functions.delete": "google_cloudfunctions2_function",
		"cloudfunctions.v2.functions.update": "google_cloudfunctions2_function",

		// Pub/Sub - Topics
		"google.pubsub.v1.Publisher.CreateTopic": "google_pubsub_topic",
		"google.pubsub.v1.Publisher.DeleteTopic": "google_pubsub_topic",

		// Pub/Sub - Subscriptions
		"google.pubsub.v1.Subscriber.CreateSubscription": "google_pubsub_subscription",
		"google.pubsub.v1.Subscriber.DeleteSubscription": "google_pubsub_subscription",

		// BigQuery - Datasets
		"google.cloud.bigquery.v2.DatasetService.InsertDataset": "google_bigquery_dataset",
		"google.cloud.bigquery.v2.DatasetService.DeleteDataset": "google_bigquery_dataset",
		"google.cloud.bigquery.v2.DatasetService.PatchDataset":  "google_bigquery_dataset",

		// BigQuery - Tables
		"google.cloud.bigquery.v2.TableService.InsertTable": "google_bigquery_table",
		"google.cloud.bigquery.v2.TableService.DeleteTable": "google_bigquery_table",
		"google.cloud.bigquery.v2.TableService.PatchTable":  "google_bigquery_table",

		// Cloud KMS - KeyRings
		"google.cloud.kms.v1.KeyManagementService.CreateKeyRing": "google_kms_key_ring",

		// Cloud KMS - CryptoKeys
		"google.cloud.kms.v1.KeyManagementService.CreateCryptoKey": "google_kms_crypto_key",
		"google.cloud.kms.v1.KeyManagementService.UpdateCryptoKey": "google_kms_crypto_key",

		// Secret Manager - Secrets
		"google.cloud.secretmanager.v1.SecretManagerService.CreateSecret": "google_secret_manager_secret",
		"google.cloud.secretmanager.v1.SecretManagerService.DeleteSecret": "google_secret_manager_secret",

		// Compute Engine - Load Balancers
		"compute.backendServices.insert": "google_compute_backend_service",
		"compute.backendServices.delete": "google_compute_backend_service",
		"compute.backendServices.update": "google_compute_backend_service",
		"compute.backendServices.patch":  "google_compute_backend_service",

		// Compute Engine - Health Checks
		"compute.healthChecks.insert": "google_compute_health_check",
		"compute.healthChecks.delete": "google_compute_health_check",
		"compute.healthChecks.update": "google_compute_health_check",
		"compute.healthChecks.patch":  "google_compute_health_check",

		// Compute Engine - Target Pools
		"compute.targetPools.insert": "google_compute_target_pool",
		"compute.targetPools.delete": "google_compute_target_pool",

		// Compute Engine - Forwarding Rules
		"compute.forwardingRules.insert": "google_compute_forwarding_rule",
		"compute.forwardingRules.delete": "google_compute_forwarding_rule",

		// Compute Engine - SSL Certificates
		"compute.sslCertificates.insert": "google_compute_ssl_certificate",
		"compute.sslCertificates.delete": "google_compute_ssl_certificate",
	}
}

// GetAllSupportedEvents returns all supported event method names
func (m *ResourceMapper) GetAllSupportedEvents() []string {
	events := make([]string, 0, len(m.eventToResource))
	for event := range m.eventToResource {
		events = append(events, event)
	}
	return events
}

// GetResourceTypesForService returns all resource types for a given GCP service
func (m *ResourceMapper) GetResourceTypesForService(serviceName string) []string {
	resourceTypes := make(map[string]bool)

	for event, resourceType := range m.eventToResource {
		// Check if event starts with service name
		if len(event) > len(serviceName) && event[:len(serviceName)] == serviceName {
			resourceTypes[resourceType] = true
		}
	}

	result := make([]string, 0, len(resourceTypes))
	for rt := range resourceTypes {
		result = append(result, rt)
	}
	return result
}
