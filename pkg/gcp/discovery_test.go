package gcp

import (
	"testing"

	compute "google.golang.org/api/compute/v1"
	"github.com/stretchr/testify/assert"
)

// Tests for DiscoveredResource structure and fields

func TestDiscoveredResource_ComputeInstance(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/zones/us-central1-a/instances/web-server",
		Type:   "google_compute_instance",
		Name:   "web-server",
		Region: "us-central1",
		Attributes: map[string]interface{}{
			"machine_type": "n1-standard-1",
			"status":       "RUNNING",
			"zone":         "us-central1-a",
		},
		Labels: map[string]string{
			"env": "production",
		},
	}

	if resource.Type != "google_compute_instance" {
		t.Errorf("expected type google_compute_instance, got %s", resource.Type)
	}
	if resource.Attributes["machine_type"] != "n1-standard-1" {
		t.Errorf("expected machine_type n1-standard-1")
	}
	if resource.Labels["env"] != "production" {
		t.Errorf("expected env label production")
	}
}

func TestDiscoveredResource_StorageBucket(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "my-data-bucket",
		Type:   "google_storage_bucket",
		Name:   "my-data-bucket",
		Region: "us-central1",
		Attributes: map[string]interface{}{
			"location":       "US",
			"storage_class":  "STANDARD",
			"versioning":     true,
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_storage_bucket" {
		t.Errorf("expected type google_storage_bucket")
	}
	if resource.Attributes["storage_class"] != "STANDARD" {
		t.Errorf("expected storage_class STANDARD")
	}
}

func TestDiscoveredResource_GKECluster(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/zones/us-central1-a/clusters/prod-cluster",
		Type:   "google_container_cluster",
		Name:   "prod-cluster",
		Region: "us-central1",
		SelfLink: "https://container.googleapis.com/v1/projects/my-project/zones/us-central1-a/clusters/prod-cluster",
		Attributes: map[string]interface{}{
			"location":               "us-central1",
			"network":                "default",
			"subnetwork":             "default",
			"current_master_version": "1.24.1",
		},
		Labels: map[string]string{
			"environment": "production",
		},
	}

	if resource.Type != "google_container_cluster" {
		t.Errorf("expected type google_container_cluster")
	}
	if resource.Attributes["network"] != "default" {
		t.Errorf("expected network default")
	}
}

func TestDiscoveredResource_SQLInstance(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "my-database-instance",
		Type:   "google_sql_database_instance",
		Name:   "my-database-instance",
		Region: "us-central1",
		Attributes: map[string]interface{}{
			"database_version": "POSTGRES_13",
			"region":           "us-central1",
			"tier":             "db-f1-micro",
			"availability_type": "ZONAL",
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_sql_database_instance" {
		t.Errorf("expected type google_sql_database_instance")
	}
	if resource.Attributes["database_version"] != "POSTGRES_13" {
		t.Errorf("expected POSTGRES_13")
	}
}

func TestDiscoveredResource_CloudRunService(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/locations/us-central1/services/my-service",
		Type:   "google_cloud_run_v2_service",
		Name:   "my-service",
		Region: "us-central1",
		Attributes: map[string]interface{}{
			"location": "us-central1",
			"ingress":  "INGRESS_TRAFFIC_ALL",
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_cloud_run_v2_service" {
		t.Errorf("expected type google_cloud_run_v2_service")
	}
	if resource.Attributes["location"] != "us-central1" {
		t.Errorf("expected location us-central1")
	}
}

func TestDiscoveredResource_Firewall(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/global/firewalls/allow-ssh",
		Type:   "google_compute_firewall",
		Name:   "allow-ssh",
		Region: "global",
		Attributes: map[string]interface{}{
			"network":   "default",
			"direction": "INGRESS",
			"priority":  1000,
			"disabled":  false,
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_compute_firewall" {
		t.Errorf("expected type google_compute_firewall")
	}
	if resource.Attributes["direction"] != "INGRESS" {
		t.Errorf("expected direction INGRESS")
	}
}

func TestDiscoveredResource_Network(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/global/networks/custom-vpc",
		Type:   "google_compute_network",
		Name:   "custom-vpc",
		Region: "global",
		Attributes: map[string]interface{}{
			"auto_create_subnetworks": false,
			"routing_mode":            "REGIONAL",
			"description":             "Custom VPC network",
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_compute_network" {
		t.Errorf("expected type google_compute_network")
	}
	if resource.Attributes["routing_mode"] != "REGIONAL" {
		t.Errorf("expected routing_mode REGIONAL")
	}
}

func TestDiscoveredResource_Subnetwork(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "projects/my-project/regions/us-central1/subnetworks/custom-subnet",
		Type:   "google_compute_subnetwork",
		Name:   "custom-subnet",
		Region: "us-central1",
		Attributes: map[string]interface{}{
			"ip_cidr_range":            "10.0.0.0/24",
			"network":                  "custom-vpc",
			"private_ip_google_access": true,
		},
		Labels: map[string]string{},
	}

	if resource.Type != "google_compute_subnetwork" {
		t.Errorf("expected type google_compute_subnetwork")
	}
	if resource.Attributes["ip_cidr_range"] != "10.0.0.0/24" {
		t.Errorf("expected ip_cidr_range 10.0.0.0/24")
	}
}

// Tests for DriftResult

func TestDriftResult_AllEmpty(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*TerraformResource{},
		ModifiedResources:  []*ResourceDiff{},
	}

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources")
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources")
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources")
	}
}

func TestDriftResult_MultipleResourceTypes(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{
			{
				ID:   "bucket-1",
				Type: "google_storage_bucket",
			},
			{
				ID:   "instance-1",
				Type: "google_compute_instance",
			},
		},
		MissingResources: []*TerraformResource{
			{
				Type: "google_compute_network",
				Name: "deleted-vpc",
			},
		},
		ModifiedResources: []*ResourceDiff{
			{
				ResourceID:   "my-cluster",
				ResourceType: "google_container_cluster",
			},
		},
	}

	if len(result.UnmanagedResources) != 2 {
		t.Errorf("expected 2 unmanaged resources")
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource")
	}
	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource")
	}
}

// Tests for TerraformResource

func TestTerraformResource_ComputeInstance(t *testing.T) {
	resource := &TerraformResource{
		Type: "google_compute_instance",
		Name: "web-server",
		Attributes: map[string]interface{}{
			"machine_type": "n1-standard-2",
			"zone":         "us-central1-a",
			"name":         "web-server",
		},
	}

	if resource.Type != "google_compute_instance" {
		t.Errorf("expected type google_compute_instance")
	}
	if resource.Attributes["machine_type"] != "n1-standard-2" {
		t.Errorf("expected machine_type n1-standard-2")
	}
}

// Tests for ResourceDiff

func TestResourceDiff_SingleDifference(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "my-cluster",
		ResourceType: "google_container_cluster",
		TerraformState: map[string]interface{}{
			"current_master_version": "1.23.0",
		},
		ActualState: map[string]interface{}{
			"current_master_version": "1.24.0",
		},
		Differences: []FieldDiff{
			{
				Field:          "current_master_version",
				TerraformValue: "1.23.0",
				ActualValue:    "1.24.0",
			},
		},
	}

	if diff.ResourceID != "my-cluster" {
		t.Errorf("expected ResourceID my-cluster")
	}
	if len(diff.Differences) != 1 {
		t.Errorf("expected 1 difference")
	}
	if diff.Differences[0].Field != "current_master_version" {
		t.Errorf("expected field current_master_version")
	}
}

func TestResourceDiff_MultipleDifferences(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "test-network",
		ResourceType: "google_compute_network",
		TerraformState: map[string]interface{}{
			"auto_create_subnetworks": true,
			"routing_mode":            "REGIONAL",
		},
		ActualState: map[string]interface{}{
			"auto_create_subnetworks": false,
			"routing_mode":            "GLOBAL",
		},
		Differences: []FieldDiff{
			{
				Field:          "auto_create_subnetworks",
				TerraformValue: true,
				ActualValue:    false,
			},
			{
				Field:          "routing_mode",
				TerraformValue: "REGIONAL",
				ActualValue:    "GLOBAL",
			},
		},
	}

	if len(diff.Differences) != 2 {
		t.Errorf("expected 2 differences")
	}
}

// Tests for FieldDiff

func TestFieldDiff_StringValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "status",
		TerraformValue: "RUNNING",
		ActualValue:    "STOPPED",
	}

	if diff.Field != "status" {
		t.Errorf("expected field status")
	}
	if diff.TerraformValue != "RUNNING" {
		t.Errorf("expected RUNNING")
	}
	if diff.ActualValue != "STOPPED" {
		t.Errorf("expected STOPPED")
	}
}

func TestFieldDiff_NumericValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "priority",
		TerraformValue: 1000,
		ActualValue:    2000,
	}

	if diff.TerraformValue != 1000 {
		t.Errorf("expected 1000")
	}
	if diff.ActualValue != 2000 {
		t.Errorf("expected 2000")
	}
}

func TestFieldDiff_BooleanValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "disabled",
		TerraformValue: false,
		ActualValue:    true,
	}

	if diff.TerraformValue != false {
		t.Errorf("expected false")
	}
	if diff.ActualValue != true {
		t.Errorf("expected true")
	}
}


// Tests for discovery helper functions

func TestExtractZoneFromURL_Standard(t *testing.T) {
	url := "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a"
	zone := extractZoneFromURL(url)
	assert.Equal(t, "us-central1-a", zone)
}

func TestExtractZoneFromURL_WithPath(t *testing.T) {
	url := "https://www.googleapis.com/compute/v1/projects/my-project/zones/europe-west1-b/instances/vm-1"
	zone := extractZoneFromURL(url)
	assert.Equal(t, "europe-west1-b", zone)
}

func TestExtractZoneFromURL_NoZone(t *testing.T) {
	url := "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default"
	zone := extractZoneFromURL(url)
	assert.Equal(t, url, zone)
}

func TestExtractLastSegment_URL(t *testing.T) {
	url := "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a/machineTypes/n1-standard-1"
	segment := extractLastSegment(url)
	assert.Equal(t, "n1-standard-1", segment)
}

func TestExtractLastSegment_Simple(t *testing.T) {
	segment := extractLastSegment("bucket-name")
	assert.Equal(t, "bucket-name", segment)
}

func TestExtractLastSegment_Path(t *testing.T) {
	url := "projects/p/zones/z/instances/inst"
	segment := extractLastSegment(url)
	assert.Equal(t, "inst", segment)
}

func TestSubnetworkToDiscovered(t *testing.T) {
	// Test the subnetworkToDiscovered helper function
	projectID := "my-project"
	subnetwork := &compute.Subnetwork{
		Name:                  "subnet-1",
		Network:               "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default",
		IpCidrRange:           "10.0.0.0/24",
		Region:                "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1",
		PrivateIpGoogleAccess: true,
		Purpose:               "PRIVATE",
		SelfLink:              "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/subnetworks/subnet-1",
	}

	result := subnetworkToDiscovered(projectID, subnetwork)
	assert.Equal(t, "google_compute_subnetwork", result.Type)
	assert.Equal(t, "subnet-1", result.Name)
	assert.Equal(t, "10.0.0.0/24", result.Attributes["ip_cidr_range"])
	assert.Equal(t, true, result.Attributes["private_ip_google_access"])
}
