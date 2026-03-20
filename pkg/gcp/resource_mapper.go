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
		"compute.subnetworks.insert":                   "google_compute_subnetwork",
		"compute.subnetworks.delete":                   "google_compute_subnetwork",
		"compute.subnetworks.patch":                    "google_compute_subnetwork",
		"compute.subnetworks.setPrivateIpGoogleAccess": "google_compute_subnetwork",

		// Compute Engine - Disks
		"compute.disks.insert":    "google_compute_disk",
		"compute.disks.delete":    "google_compute_disk",
		"compute.disks.resize":    "google_compute_disk",
		"compute.disks.setLabels": "google_compute_disk",

		// Compute Engine - Routes
		"compute.routes.insert": "google_compute_route",
		"compute.routes.delete": "google_compute_route",

		// Compute Engine - Routers
		"compute.routers.insert": "google_compute_router",
		"compute.routers.delete": "google_compute_router",
		"compute.routers.patch":  "google_compute_router",
		"compute.routers.update": "google_compute_router",

		// Compute Engine - VPN
		"compute.vpnTunnels.insert":  "google_compute_vpn_tunnel",
		"compute.vpnTunnels.delete":  "google_compute_vpn_tunnel",
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

		// ========== NEW SERVICES (v0.6.0) ==========

		// Compute Engine - Security Policies (Cloud Armor)
		"compute.securityPolicies.insert": "google_compute_security_policy",
		"compute.securityPolicies.delete": "google_compute_security_policy",
		"compute.securityPolicies.patch":  "google_compute_security_policy",

		// Compute Engine - Router NAT (note: compute.routers.insert already mapped to google_compute_router above)
		// NAT is configured as part of the router resource in GCP

		// Compute Engine - Global Forwarding Rules
		"compute.globalForwardingRules.insert": "google_compute_global_forwarding_rule",
		"compute.globalForwardingRules.delete": "google_compute_global_forwarding_rule",

		// Compute Engine - URL Maps
		"compute.urlMaps.insert": "google_compute_url_map",
		"compute.urlMaps.delete": "google_compute_url_map",
		"compute.urlMaps.patch":  "google_compute_url_map",

		// Compute Engine - Target HTTPS Proxies
		"compute.targetHttpsProxies.insert": "google_compute_target_https_proxy",
		"compute.targetHttpsProxies.delete": "google_compute_target_https_proxy",

		// Compute Engine - Managed SSL Certificates
		"compute.managedSslCertificates.insert": "google_compute_managed_ssl_certificate",
		"compute.managedSslCertificates.delete": "google_compute_managed_ssl_certificate",

		// Cloud DNS - Managed Zones
		"dns.managedZones.create": "google_dns_managed_zone",
		"dns.managedZones.delete": "google_dns_managed_zone",
		"dns.managedZones.update": "google_dns_managed_zone",
		"dns.managedZones.patch":  "google_dns_managed_zone",

		// Cloud DNS - Record Sets
		"dns.changes.create": "google_dns_record_set",

		// Memorystore (Redis)
		"google.cloud.redis.v1.CloudRedis.CreateInstance":  "google_redis_instance",
		"google.cloud.redis.v1.CloudRedis.DeleteInstance":  "google_redis_instance",
		"google.cloud.redis.v1.CloudRedis.UpdateInstance":  "google_redis_instance",
		"google.cloud.redis.v1.CloudRedis.UpgradeInstance": "google_redis_instance",

		// Cloud Spanner - Instances
		"google.spanner.admin.instance.v1.InstanceAdmin.CreateInstance": "google_spanner_instance",
		"google.spanner.admin.instance.v1.InstanceAdmin.DeleteInstance": "google_spanner_instance",
		"google.spanner.admin.instance.v1.InstanceAdmin.UpdateInstance": "google_spanner_instance",

		// Cloud Spanner - Databases
		"google.spanner.admin.database.v1.DatabaseAdmin.CreateDatabase": "google_spanner_database",
		"google.spanner.admin.database.v1.DatabaseAdmin.DropDatabase":   "google_spanner_database",
		"google.spanner.admin.database.v1.DatabaseAdmin.UpdateDatabase": "google_spanner_database",

		// Artifact Registry
		"google.devtools.artifactregistry.v1.ArtifactRegistry.CreateRepository": "google_artifact_registry_repository",
		"google.devtools.artifactregistry.v1.ArtifactRegistry.DeleteRepository": "google_artifact_registry_repository",
		"google.devtools.artifactregistry.v1.ArtifactRegistry.UpdateRepository": "google_artifact_registry_repository",

		// Cloud Scheduler
		"google.cloud.scheduler.v1.CloudScheduler.CreateJob": "google_cloud_scheduler_job",
		"google.cloud.scheduler.v1.CloudScheduler.DeleteJob": "google_cloud_scheduler_job",
		"google.cloud.scheduler.v1.CloudScheduler.UpdateJob": "google_cloud_scheduler_job",

		// Cloud Tasks
		"google.cloud.tasks.v2.CloudTasks.CreateQueue": "google_cloud_tasks_queue",
		"google.cloud.tasks.v2.CloudTasks.DeleteQueue": "google_cloud_tasks_queue",
		"google.cloud.tasks.v2.CloudTasks.UpdateQueue": "google_cloud_tasks_queue",

		// Filestore
		"google.cloud.filestore.v1.CloudFilestoreManager.CreateInstance": "google_filestore_instance",
		"google.cloud.filestore.v1.CloudFilestoreManager.DeleteInstance": "google_filestore_instance",
		"google.cloud.filestore.v1.CloudFilestoreManager.UpdateInstance": "google_filestore_instance",

		// Cloud Logging - Sinks
		"google.logging.v2.ConfigServiceV2.CreateSink": "google_logging_project_sink",
		"google.logging.v2.ConfigServiceV2.DeleteSink": "google_logging_project_sink",
		"google.logging.v2.ConfigServiceV2.UpdateSink": "google_logging_project_sink",

		// Cloud Logging - Metrics
		"google.logging.v2.MetricsServiceV2.CreateLogMetric": "google_logging_metric",
		"google.logging.v2.MetricsServiceV2.DeleteLogMetric": "google_logging_metric",
		"google.logging.v2.MetricsServiceV2.UpdateLogMetric": "google_logging_metric",

		// Cloud Monitoring - Alert Policies
		"google.monitoring.v3.AlertPolicyService.CreateAlertPolicy": "google_monitoring_alert_policy",
		"google.monitoring.v3.AlertPolicyService.DeleteAlertPolicy": "google_monitoring_alert_policy",
		"google.monitoring.v3.AlertPolicyService.UpdateAlertPolicy": "google_monitoring_alert_policy",

		// Cloud Monitoring - Notification Channels
		"google.monitoring.v3.NotificationChannelService.CreateNotificationChannel": "google_monitoring_notification_channel",
		"google.monitoring.v3.NotificationChannelService.DeleteNotificationChannel": "google_monitoring_notification_channel",
		"google.monitoring.v3.NotificationChannelService.UpdateNotificationChannel": "google_monitoring_notification_channel",

		// Dataproc - Clusters
		"google.cloud.dataproc.v1.ClusterController.CreateCluster": "google_dataproc_cluster",
		"google.cloud.dataproc.v1.ClusterController.DeleteCluster": "google_dataproc_cluster",
		"google.cloud.dataproc.v1.ClusterController.UpdateCluster": "google_dataproc_cluster",

		// Cloud Build - Triggers
		"google.devtools.cloudbuild.v1.CloudBuild.CreateBuildTrigger": "google_cloudbuild_trigger",
		"google.devtools.cloudbuild.v1.CloudBuild.DeleteBuildTrigger": "google_cloudbuild_trigger",
		"google.devtools.cloudbuild.v1.CloudBuild.UpdateBuildTrigger": "google_cloudbuild_trigger",

		// Workflows
		"google.cloud.workflows.v1.Workflows.CreateWorkflow": "google_workflows_workflow",
		"google.cloud.workflows.v1.Workflows.DeleteWorkflow": "google_workflows_workflow",
		"google.cloud.workflows.v1.Workflows.UpdateWorkflow": "google_workflows_workflow",

		// VPC Service Controls
		"google.identity.accesscontextmanager.v1.AccessContextManager.CreateServicePerimeter": "google_access_context_manager_service_perimeter",
		"google.identity.accesscontextmanager.v1.AccessContextManager.DeleteServicePerimeter": "google_access_context_manager_service_perimeter",
		"google.identity.accesscontextmanager.v1.AccessContextManager.UpdateServicePerimeter": "google_access_context_manager_service_perimeter",
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
