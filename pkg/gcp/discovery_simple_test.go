package gcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	compute "google.golang.org/api/compute/v1"
)

// TestDiscoverAll_HelperFunctions tests the helper functions used by discovery
func TestDiscoverAll_HelperFunctions(t *testing.T) {
	// Test extractRegionFromURL
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.googleapis.com/compute/v1/projects/p/regions/us-central1", "us-central1"},
		{"https://www.googleapis.com/compute/v1/projects/p/regions/europe-west1/subnetworks/default", "europe-west1"},
		{"us-central1", "us-central1"},
	}

	for _, tt := range tests {
		result := extractRegionFromURL(tt.url)
		assert.Equal(t, tt.expected, result)
	}
}

// TestExtractZoneFromURL_Helper tests zone extraction helper
func TestExtractZoneFromURL_Helper(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a", "us-central1-a"},
		{"https://www.googleapis.com/compute/v1/projects/my-project/zones/europe-west1-b/instances/vm-1", "europe-west1-b"},
		{"https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default", "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default"},
	}

	for _, tt := range tests {
		result := extractZoneFromURL(tt.url)
		assert.Equal(t, tt.expected, result)
	}
}

// TestZoneToRegion_Helper tests zone to region conversion
func TestZoneToRegion_Helper(t *testing.T) {
	tests := []struct {
		zone     string
		expected string
	}{
		{"us-central1-a", "us-central1"},
		{"europe-west1-b", "europe-west1"},
		{"asia-east1-c", "asia-east1"},
		{"us-central1", "us-central1"},
	}

	for _, tt := range tests {
		result := zoneToRegion(tt.zone)
		assert.Equal(t, tt.expected, result)
	}
}

// TestExtractLastSegment_Helper tests last segment extraction
func TestExtractLastSegment_Helper(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://compute.googleapis.com/compute/v1/projects/p/zones/z/machineTypes/e2-medium", "e2-medium"},
		{"simple-name", "simple-name"},
		{"a/b/c", "c"},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractLastSegment(tt.url)
		assert.Equal(t, tt.expected, result)
	}
}

// TestContainsString_Helper tests string slice contains
func TestContainsString_Helper(t *testing.T) {
	slice := []string{"us-central1", "europe-west1", "asia-east1"}

	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{"found", slice, "us-central1", true},
		{"not found", slice, "us-east1", false},
		{"nil slice", nil, "anything", false},
		{"empty slice", []string{}, "anything", false},
	}

	for _, tt := range tests {
		result := containsString(tt.slice, tt.value)
		assert.Equal(t, tt.expected, result, tt.name)
	}
}

// TestSubnetworkToDiscovered_Helper tests subnetwork conversion
func TestSubnetworkToDiscovered_Helper(t *testing.T) {
	subnetwork := &compute.Subnetwork{
		Name:                  "subnet-1",
		Network:               "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default",
		IpCidrRange:           "10.0.0.0/24",
		Region:                "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1",
		PrivateIpGoogleAccess: true,
		Purpose:               "PRIVATE",
		SelfLink:              "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1/subnetworks/subnet-1",
	}

	result := subnetworkToDiscovered("my-project", subnetwork)

	assert.Equal(t, "google_compute_subnetwork", result.Type)
	assert.Equal(t, "subnet-1", result.Name)
	assert.Equal(t, "us-central1", result.Region)
	assert.Equal(t, "10.0.0.0/24", result.Attributes["ip_cidr_range"])
	assert.Equal(t, true, result.Attributes["private_ip_google_access"])
	assert.Equal(t, "PRIVATE", result.Attributes["purpose"])
}

// TestNewDiscoveryClient_Structure tests the structure of NewDiscoveryClient without network calls
func TestNewDiscoveryClient_Structure(t *testing.T) {
	// This test verifies the function exists and is callable
	// A full test would require mocking GCP APIs
	t.Skip("Requires GCP authentication")
}
