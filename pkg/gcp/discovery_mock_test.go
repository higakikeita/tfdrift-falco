package gcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// TestDiscoverAll_WithMocks tests DiscoverAll using function overrides
func TestDiscoverAll_WithMocks(t *testing.T) {
	client := NewDiscoveryClientForTesting("test-project", []string{"us-central1"})

	// Set up mocks for all discovery functions
	client.discoverNetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/global/networks/vpc-1", Type: "google_compute_network", Name: "vpc-1", Region: "global"},
			{ID: "projects/test-project/global/networks/vpc-2", Type: "google_compute_network", Name: "vpc-2", Region: "global"},
		}, nil
	}
	client.discoverSubnetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/regions/us-central1/subnetworks/sub-1", Type: "google_compute_subnetwork", Name: "sub-1", Region: "us-central1"},
		}, nil
	}
	client.discoverFirewallsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/global/firewalls/fw-1", Type: "google_compute_firewall", Name: "fw-1", Region: "global"},
		}, nil
	}
	client.discoverInstancesFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/zones/us-central1-a/instances/vm-1", Type: "google_compute_instance", Name: "vm-1", Region: "us-central1"},
		}, nil
	}
	client.discoverBucketsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "my-bucket", Type: "google_storage_bucket", Name: "my-bucket", Region: "us-central1"},
		}, nil
	}
	client.discoverSQLFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/instances/db-1", Type: "google_sql_database_instance", Name: "db-1", Region: "us-central1"},
		}, nil
	}
	client.discoverGKEFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/locations/us-central1/clusters/k8s-1", Type: "google_container_cluster", Name: "k8s-1", Region: "us-central1"},
		}, nil
	}
	client.discoverCloudRunFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "projects/test-project/locations/us-central1/services/svc-1", Type: "google_cloud_run_v2_service", Name: "svc-1", Region: "us-central1"},
		}, nil
	}

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	if len(resources) != 9 {
		t.Errorf("expected 9 resources, got %d", len(resources))
	}

	// Verify resource types
	typeCounts := make(map[string]int)
	for _, r := range resources {
		typeCounts[r.Type]++
	}

	expected := map[string]int{
		"google_compute_network":      2,
		"google_compute_subnetwork":   1,
		"google_compute_firewall":     1,
		"google_compute_instance":     1,
		"google_storage_bucket":       1,
		"google_sql_database_instance": 1,
		"google_container_cluster":    1,
		"google_cloud_run_v2_service": 1,
	}

	for typ, count := range expected {
		if typeCounts[typ] != count {
			t.Errorf("expected %d %s, got %d", count, typ, typeCounts[typ])
		}
	}
}

// TestDiscoverAll_WithPartialErrors tests DiscoverAll when some discovery functions fail
func TestDiscoverAll_WithPartialErrors(t *testing.T) {
	client := NewDiscoveryClientForTesting("test-project", nil)

	// Networks succeed
	client.discoverNetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "net-1", Type: "google_compute_network", Name: "net-1"},
		}, nil
	}
	// Subnetworks fail
	client.discoverSubnetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, fmt.Errorf("permission denied: compute.subnetworks.list")
	}
	// Firewalls succeed
	client.discoverFirewallsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "fw-1", Type: "google_compute_firewall", Name: "fw-1"},
		}, nil
	}
	// Instances fail
	client.discoverInstancesFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, fmt.Errorf("quota exceeded")
	}
	// Buckets succeed
	client.discoverBucketsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "bucket-1", Type: "google_storage_bucket", Name: "bucket-1"},
		}, nil
	}
	// SQL fail
	client.discoverSQLFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, fmt.Errorf("API disabled")
	}
	// GKE succeed
	client.discoverGKEFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{ID: "cluster-1", Type: "google_container_cluster", Name: "cluster-1"},
		}, nil
	}
	// Cloud Run fail
	client.discoverCloudRunFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, fmt.Errorf("service not available")
	}

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll should not return error for partial failures: %v", err)
	}

	// Should get results from successful functions only: 1+1+1+1 = 4
	if len(resources) != 4 {
		t.Errorf("expected 4 resources (from successful discovers), got %d", len(resources))
	}
}

// TestDiscoverAll_AllErrors tests DiscoverAll when all discovery functions fail
func TestDiscoverAll_AllErrors(t *testing.T) {
	client := NewDiscoveryClientForTesting("test-project", nil)

	errFunc := func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, fmt.Errorf("service unavailable")
	}

	client.discoverNetworksFunc = errFunc
	client.discoverSubnetworksFunc = errFunc
	client.discoverFirewallsFunc = errFunc
	client.discoverInstancesFunc = errFunc
	client.discoverBucketsFunc = errFunc
	client.discoverSQLFunc = errFunc
	client.discoverGKEFunc = errFunc
	client.discoverCloudRunFunc = errFunc

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll should not return error: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

// TestDiscoverAll_EmptyResults tests DiscoverAll when discovery returns empty results
func TestDiscoverAll_EmptyResults(t *testing.T) {
	client := NewDiscoveryClientForTesting("empty-project", []string{"us-east1"})

	emptyFunc := func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}

	client.discoverNetworksFunc = emptyFunc
	client.discoverSubnetworksFunc = emptyFunc
	client.discoverFirewallsFunc = emptyFunc
	client.discoverInstancesFunc = emptyFunc
	client.discoverBucketsFunc = emptyFunc
	client.discoverSQLFunc = emptyFunc
	client.discoverGKEFunc = emptyFunc
	client.discoverCloudRunFunc = emptyFunc

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	if resources != nil && len(resources) != 0 {
		t.Errorf("expected nil or empty resources, got %d", len(resources))
	}
}

// TestDiscoverAll_RichResourceAttributes tests that discovered resources have proper attributes
func TestDiscoverAll_RichResourceAttributes(t *testing.T) {
	client := NewDiscoveryClientForTesting("my-project", []string{"us-west1"})

	client.discoverNetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:       "projects/my-project/global/networks/prod-vpc",
				Type:     "google_compute_network",
				Name:     "prod-vpc",
				Region:   "global",
				SelfLink: "https://compute.googleapis.com/compute/v1/projects/my-project/global/networks/prod-vpc",
				Attributes: map[string]interface{}{
					"name":                    "prod-vpc",
					"auto_create_subnetworks": false,
					"routing_mode":            "REGIONAL",
					"description":             "Production VPC",
				},
			},
		}, nil
	}
	client.discoverSubnetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:       "projects/my-project/regions/us-west1/subnetworks/prod-subnet",
				Type:     "google_compute_subnetwork",
				Name:     "prod-subnet",
				Region:   "us-west1",
				SelfLink: "https://compute.googleapis.com/compute/v1/projects/my-project/regions/us-west1/subnetworks/prod-subnet",
				Attributes: map[string]interface{}{
					"name":          "prod-subnet",
					"ip_cidr_range": "10.0.0.0/24",
					"region":        "us-west1",
				},
			},
		}, nil
	}
	client.discoverFirewallsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:   "projects/my-project/global/firewalls/allow-ssh",
				Type: "google_compute_firewall",
				Name: "allow-ssh",
				Attributes: map[string]interface{}{
					"name":      "allow-ssh",
					"direction": "INGRESS",
					"priority":  int64(1000),
					"disabled":  false,
				},
			},
		}, nil
	}
	client.discoverInstancesFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:     "projects/my-project/zones/us-west1-a/instances/web-server",
				Type:   "google_compute_instance",
				Name:   "web-server",
				Region: "us-west1",
				Labels: map[string]string{"env": "prod", "team": "platform"},
				Attributes: map[string]interface{}{
					"name":         "web-server",
					"machine_type": "e2-medium",
					"zone":         "us-west1-a",
					"status":       "RUNNING",
				},
			},
		}, nil
	}
	client.discoverBucketsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:     "my-project-assets",
				Type:   "google_storage_bucket",
				Name:   "my-project-assets",
				Region: "us-west1",
				Labels: map[string]string{"managed-by": "terraform"},
				Attributes: map[string]interface{}{
					"name":          "my-project-assets",
					"location":      "us-west1",
					"storage_class": "STANDARD",
					"versioning":    true,
				},
			},
		}, nil
	}
	client.discoverSQLFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:     "projects/my-project/instances/prod-db",
				Type:   "google_sql_database_instance",
				Name:   "prod-db",
				Region: "us-west1",
				Attributes: map[string]interface{}{
					"name":             "prod-db",
					"database_version": "POSTGRES_15",
					"tier":             "db-custom-4-16384",
				},
			},
		}, nil
	}
	client.discoverGKEFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:     "projects/my-project/locations/us-west1/clusters/prod-gke",
				Type:   "google_container_cluster",
				Name:   "prod-gke",
				Region: "us-west1",
				Labels: map[string]string{"env": "production"},
				Attributes: map[string]interface{}{
					"name":                   "prod-gke",
					"location":               "us-west1",
					"current_master_version": "1.28.5-gke.1200",
				},
			},
		}, nil
	}
	client.discoverCloudRunFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{
			{
				ID:     "projects/my-project/locations/us-west1/services/api-svc",
				Type:   "google_cloud_run_v2_service",
				Name:   "api-svc",
				Region: "us-west1",
				Labels: map[string]string{"app": "api"},
				Attributes: map[string]interface{}{
					"name":     "api-svc",
					"location": "us-west1",
					"ingress":  "INGRESS_TRAFFIC_ALL",
				},
			},
		}, nil
	}

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	if len(resources) != 8 {
		t.Fatalf("expected 8 resources, got %d", len(resources))
	}

	// Verify network attributes
	for _, r := range resources {
		if r.Name == "" {
			t.Errorf("resource %s has empty name", r.ID)
		}
		if r.Type == "" {
			t.Errorf("resource %s has empty type", r.ID)
		}
		if r.Attributes == nil {
			t.Errorf("resource %s has nil attributes", r.ID)
		}
	}

	// Verify specific resource
	var webServer *types.DiscoveredResource
	for _, r := range resources {
		if r.Name == "web-server" {
			webServer = r
			break
		}
	}
	if webServer == nil {
		t.Fatal("web-server instance not found")
	}
	if webServer.Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got %s", webServer.Labels["env"])
	}
	if webServer.Attributes["machine_type"] != "e2-medium" {
		t.Errorf("expected machine_type e2-medium, got %v", webServer.Attributes["machine_type"])
	}
}

// TestDiscoverAll_ContextCancellation tests that DiscoverAll handles context cancellation
func TestDiscoverAll_ContextCancellation(t *testing.T) {
	client := NewDiscoveryClientForTesting("test-project", nil)

	client.discoverNetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{{ID: "net-1", Type: "google_compute_network", Name: "net-1"}}, nil
	}
	client.discoverSubnetworksFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, ctx.Err()
	}
	client.discoverFirewallsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return []*DiscoveredResource{{ID: "fw-1", Type: "google_compute_firewall", Name: "fw-1"}}, nil
	}
	client.discoverInstancesFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}
	client.discoverBucketsFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}
	client.discoverSQLFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}
	client.discoverGKEFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}
	client.discoverCloudRunFunc = func(ctx context.Context) ([]*DiscoveredResource, error) {
		return nil, nil
	}

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll should not return error: %v", err)
	}

	// Should still get results from non-cancelled functions
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

// TestNewDiscoveryClientForTesting tests the test constructor
func TestNewDiscoveryClientForTesting(t *testing.T) {
	client := NewDiscoveryClientForTesting("my-project", []string{"us-east1", "us-west1"})

	if client.projectID != "my-project" {
		t.Errorf("expected projectID my-project, got %s", client.projectID)
	}
	if len(client.regions) != 2 {
		t.Errorf("expected 2 regions, got %d", len(client.regions))
	}
	if client.computeService != nil {
		t.Error("expected nil computeService for test client")
	}
	if client.containerService != nil {
		t.Error("expected nil containerService for test client")
	}
}

// TestDiscoverAll_LargeScale tests DiscoverAll with many resources
func TestDiscoverAll_LargeScale(t *testing.T) {
	client := NewDiscoveryClientForTesting("large-project", nil)

	// Generate large resource sets
	makeResources := func(prefix, typ string, count int) func(context.Context) ([]*DiscoveredResource, error) {
		return func(ctx context.Context) ([]*DiscoveredResource, error) {
			resources := make([]*DiscoveredResource, count)
			for i := 0; i < count; i++ {
				resources[i] = &DiscoveredResource{
					ID:   fmt.Sprintf("%s-%d", prefix, i),
					Type: typ,
					Name: fmt.Sprintf("%s-%d", prefix, i),
				}
			}
			return resources, nil
		}
	}

	client.discoverNetworksFunc = makeResources("net", "google_compute_network", 10)
	client.discoverSubnetworksFunc = makeResources("sub", "google_compute_subnetwork", 50)
	client.discoverFirewallsFunc = makeResources("fw", "google_compute_firewall", 100)
	client.discoverInstancesFunc = makeResources("vm", "google_compute_instance", 200)
	client.discoverBucketsFunc = makeResources("bucket", "google_storage_bucket", 30)
	client.discoverSQLFunc = makeResources("sql", "google_sql_database_instance", 5)
	client.discoverGKEFunc = makeResources("gke", "google_container_cluster", 3)
	client.discoverCloudRunFunc = makeResources("run", "google_cloud_run_v2_service", 15)

	resources, err := client.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	expectedTotal := 10 + 50 + 100 + 200 + 30 + 5 + 3 + 15
	if len(resources) != expectedTotal {
		t.Errorf("expected %d resources, got %d", expectedTotal, len(resources))
	}
}

// TestHelperFunctions tests the helper functions in discovery.go
func TestHelperFunctions_Extended(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{"extractRegionFromURL normal", extractRegionFromURL, "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1", "us-central1"},
		{"extractRegionFromURL no region", extractRegionFromURL, "https://www.googleapis.com/compute/v1/projects/p/global", "https://www.googleapis.com/compute/v1/projects/p/global"},
		{"extractRegionFromURL empty", extractRegionFromURL, "", ""},
		{"extractZoneFromURL normal", extractZoneFromURL, "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a", "us-central1-a"},
		{"extractZoneFromURL no zone", extractZoneFromURL, "https://www.googleapis.com/compute/v1/projects/p/global", "https://www.googleapis.com/compute/v1/projects/p/global"},
		{"extractLastSegment normal", extractLastSegment, "projects/p/machineTypes/e2-medium", "e2-medium"},
		{"extractLastSegment empty", extractLastSegment, "", ""},
		{"extractLastSegment single", extractLastSegment, "e2-medium", "e2-medium"},
		{"zoneToRegion normal", zoneToRegion, "us-central1-a", "us-central1"},
		{"zoneToRegion region input", zoneToRegion, "us-central1", "us-central1"},
		{"zoneToRegion short", zoneToRegion, "us", "us"},
		{"containsString found", func(s string) string {
			if containsString([]string{"a", "b", "c"}, s) {
				return "true"
			}
			return "false"
		}, "b", "true"},
		{"containsString not found", func(s string) string {
			if containsString([]string{"a", "b", "c"}, s) {
				return "true"
			}
			return "false"
		}, "d", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
