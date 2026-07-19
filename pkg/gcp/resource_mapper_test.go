package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceMapper(t *testing.T) {
	mapper := NewResourceMapper()
	assert.NotNil(t, mapper)
	assert.NotEmpty(t, mapper.eventToResource)
}

func TestResourceMapper_MapEventToResource(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		name       string
		methodName string
		want       string
	}{
		// Compute Engine - Instances
		{
			"Compute Instance SetMetadata",
			"compute.instances.setMetadata",
			"google_compute_instance",
		},
		{
			"Compute Instance Insert",
			"compute.instances.insert",
			"google_compute_instance",
		},

		// Firewall
		{
			"Firewall Insert",
			"compute.firewalls.insert",
			"google_compute_firewall",
		},

		// Networks
		{
			"Network Insert",
			"compute.networks.insert",
			"google_compute_network",
		},

		// IAM
		{
			"SetIamPolicy",
			"SetIamPolicy",
			"google_project_iam_binding",
		},

		// Cloud Storage
		{
			"Storage Bucket Create",
			"storage.buckets.create",
			"google_storage_bucket",
		},

		// Cloud SQL
		{
			"CloudSQL Instance Create",
			"cloudsql.instances.create",
			"google_sql_database_instance",
		},

		// GKE
		{
			"GKE Cluster Create",
			"container.clusters.create",
			"google_container_cluster",
		},

		// Cloud Run
		{
			"Cloud Run Service Create",
			"run.services.create",
			"google_cloud_run_service",
		},

		// Unknown Event
		{
			"Unknown Event",
			"unknown.method",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.MapEventToResource(tt.methodName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResourceMapper_GetAllSupportedEvents(t *testing.T) {
	mapper := NewResourceMapper()

	events := mapper.GetAllSupportedEvents()

	assert.NotEmpty(t, events)
	assert.Greater(t, len(events), 50, "Should have more than 50 supported events")

	// Check some known events exist
	expectedEvents := []string{
		"compute.instances.setMetadata",
		"compute.firewalls.insert",
		"SetIamPolicy",
		"storage.buckets.create",
	}

	for _, expected := range expectedEvents {
		found := false
		for _, event := range events {
			if event == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected event %s not found in supported events", expected)
	}
}

func TestResourceMapper_GetResourceTypesForService(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		name        string
		serviceName string
		wantCount   int
		wantTypes   []string
	}{
		{
			"Compute Service",
			"compute",
			10, // At least 10 resource types
			[]string{
				"google_compute_instance",
				"google_compute_firewall",
				"google_compute_network",
			},
		},
		{
			"Storage Service",
			"storage",
			2, // At least 2 resource types
			[]string{
				"google_storage_bucket",
			},
		},
		{
			"CloudSQL Service",
			"cloudsql",
			3, // At least 3 resource types
			[]string{
				"google_sql_database_instance",
			},
		},
		{
			"Unknown Service",
			"unknownservice",
			0,
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.GetResourceTypesForService(tt.serviceName)

			assert.GreaterOrEqual(t, len(got), tt.wantCount, "Expected at least %d resource types", tt.wantCount)

			// Check expected types are present
			for _, expectedType := range tt.wantTypes {
				found := false
				for _, gotType := range got {
					if gotType == expectedType {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected resource type %s not found", expectedType)
			}
		})
	}
}

func TestInitializeEventMapping(t *testing.T) {
	// initializeEventMapping is the hardcoded fallback used when config
	// loading fails, so it is exercised directly here.
	mapping := initializeEventMapping()

	assert.NotEmpty(t, mapping)
	assert.Greater(t, len(mapping), 100, "fallback mapping should cover 100+ events")

	// Spot-check representative events across a broad set of services to
	// guard against accidental deletions during future edits.
	cases := map[string]string{
		"compute.instances.insert":                                  "google_compute_instance",
		"compute.firewalls.insert":                                  "google_compute_firewall",
		"compute.subnetworks.insert":                                "google_compute_subnetwork",
		"SetIamPolicy":                                              "google_project_iam_binding",
		"storage.buckets.create":                                    "google_storage_bucket",
		"storage.buckets.setIamPolicy":                              "google_storage_bucket_iam_binding",
		"cloudsql.instances.create":                                 "google_sql_database_instance",
		"container.clusters.create":                                 "google_container_cluster",
		"run.services.create":                                       "google_cloud_run_service",
		"cloudfunctions.v2.functions.create":                        "google_cloudfunctions2_function",
		"google.pubsub.v1.Publisher.CreateTopic":                    "google_pubsub_topic",
		"google.cloud.dataproc.v1.ClusterController.CreateCluster":  "google_dataproc_cluster",
		"google.cloud.aiplatform.v1.EndpointService.CreateEndpoint": "google_vertex_ai_endpoint",
		"google.cloud.kms.v1.KeyManagementService.CreateKeyRing":    "google_kms_key_ring",
	}

	for method, want := range cases {
		got, ok := mapping[method]
		assert.True(t, ok, "expected mapping for %q", method)
		assert.Equal(t, want, got, "unexpected resource type for %q", method)
	}

	// Every mapped resource type should look like a Terraform google_ resource.
	for method, resourceType := range mapping {
		assert.NotEmpty(t, resourceType, "empty resource type for %q", method)
		assert.Contains(t, resourceType, "google_", "resource type %q for %q should be a google_ resource", resourceType, method)
	}
}

func TestResourceMapper_FallbackMapping(t *testing.T) {
	// A mapper built purely from the fallback map must still resolve events,
	// matching the behavior of NewResourceMapper when config loading fails.
	mapper := &ResourceMapper{eventToResource: initializeEventMapping()}

	assert.Equal(t, "google_compute_instance", mapper.MapEventToResource("compute.instances.insert"))
	assert.Equal(t, "", mapper.MapEventToResource("unknown.method"))
	assert.NotEmpty(t, mapper.GetAllSupportedEvents())
	assert.NotEmpty(t, mapper.GetResourceTypesForService("compute"))
}

func TestResourceMapper_Coverage(t *testing.T) {
	mapper := NewResourceMapper()

	// Test that all major GCP services are covered
	services := []string{
		"compute",        // Compute Engine
		"storage",        // Cloud Storage
		"cloudsql",       // Cloud SQL
		"container",      // GKE
		"run",            // Cloud Run
		"cloudfunctions", // Cloud Functions
	}

	for _, service := range services {
		t.Run(service, func(t *testing.T) {
			resourceTypes := mapper.GetResourceTypesForService(service)
			assert.NotEmpty(t, resourceTypes, "Service %s should have at least one resource type", service)
		})
	}
}
