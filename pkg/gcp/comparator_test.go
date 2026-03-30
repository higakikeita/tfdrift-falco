package gcp

import (
	"testing"
)

func TestCompareStateWithActual_UnmanagedResources(t *testing.T) {
	tfResources := []*TerraformResource{}

	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/my-project/global/networks/default",
			Type: "google_compute_network",
			Name: "default",
			Attributes: map[string]interface{}{
				"name":                    "default",
				"auto_create_subnetworks": true,
			},
		},
		{
			ID:   "my-bucket-123",
			Type: "google_storage_bucket",
			Name: "my-bucket-123",
			Attributes: map[string]interface{}{
				"name":     "my-bucket-123",
				"location": "us-central1",
			},
		},
	}

	result := CompareStateWithActual(tfResources, gcpResources)

	if len(result.UnmanagedResources) != 2 {
		t.Errorf("expected 2 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

func TestCompareStateWithActual_MissingResources(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_network",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":   "projects/my-project/global/networks/main-vpc",
				"name": "main-vpc",
			},
		},
		{
			Type: "google_storage_bucket",
			Name: "data-bucket",
			Attributes: map[string]interface{}{
				"id":   "data-bucket",
				"name": "data-bucket",
			},
		},
	}

	gcpResources := []*DiscoveredResource{}

	result := CompareStateWithActual(tfResources, gcpResources)

	if len(result.MissingResources) != 2 {
		t.Errorf("expected 2 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
}

func TestCompareStateWithActual_ModifiedResources(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_network",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":                      "projects/my-project/global/networks/main-vpc",
				"name":                    "main-vpc",
				"auto_create_subnetworks": true,
				"routing_mode":            "REGIONAL",
			},
		},
	}

	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/my-project/global/networks/main-vpc",
			Type: "google_compute_network",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"name":                    "main-vpc",
				"auto_create_subnetworks": false,
				"routing_mode":            "GLOBAL",
				"description":             "",
			},
		},
	}

	result := CompareStateWithActual(tfResources, gcpResources)

	if len(result.ModifiedResources) != 1 {
		t.Fatalf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}

	diff := result.ModifiedResources[0]
	if diff.ResourceType != "google_compute_network" {
		t.Errorf("expected resource type google_compute_network, got %s", diff.ResourceType)
	}

	// Should detect auto_create_subnetworks and routing_mode changes
	if len(diff.Differences) < 2 {
		t.Errorf("expected at least 2 differences, got %d", len(diff.Differences))
	}

	// Verify specific differences
	foundSubnetwork := false
	foundRouting := false
	for _, d := range diff.Differences {
		if d.Field == "auto_create_subnetworks" {
			foundSubnetwork = true
		}
		if d.Field == "routing_mode" {
			foundRouting = true
		}
	}
	if !foundSubnetwork {
		t.Error("expected auto_create_subnetworks difference not found")
	}
	if !foundRouting {
		t.Error("expected routing_mode difference not found")
	}
}

func TestCompareStateWithActual_MatchByName(t *testing.T) {
	// Test that resources can be matched by name when IDs differ
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"name":         "web-server",
				"machine_type": "e2-medium",
				"zone":         "us-central1-a",
				"status":       "RUNNING",
			},
		},
	}

	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/my-project/zones/us-central1-a/instances/web-server",
			Type: "google_compute_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"name":         "web-server",
				"machine_type": "e2-medium",
				"zone":         "us-central1-a",
				"status":       "RUNNING",
			},
		},
	}

	result := CompareStateWithActual(tfResources, gcpResources)

	// Should match by name - no unmanaged or missing
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged (matched by name), got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing (matched by name), got %d", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_MixedDrift(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_network",
			Name: "existing-vpc",
			Attributes: map[string]interface{}{
				"id":   "projects/p/global/networks/existing-vpc",
				"name": "existing-vpc",
			},
		},
		{
			Type: "google_storage_bucket",
			Name: "deleted-bucket",
			Attributes: map[string]interface{}{
				"id":   "deleted-bucket",
				"name": "deleted-bucket",
			},
		},
	}

	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/p/global/networks/existing-vpc",
			Type: "google_compute_network",
			Name: "existing-vpc",
			Attributes: map[string]interface{}{
				"name": "existing-vpc",
			},
		},
		{
			ID:   "manual-bucket",
			Type: "google_storage_bucket",
			Name: "manual-bucket",
			Attributes: map[string]interface{}{
				"name": "manual-bucket",
			},
		},
	}

	result := CompareStateWithActual(tfResources, gcpResources)

	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource (manual-bucket), got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource (deleted-bucket), got %d", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_EmptyInputs(t *testing.T) {
	result := CompareStateWithActual(nil, nil)

	if result == nil {
		t.Fatal("expected non-nil result for empty inputs")
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified, got %d", len(result.ModifiedResources))
	}
}

func TestCompareStateWithActual_LabelDrift(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_instance",
			Name: "labeled-vm",
			Attributes: map[string]interface{}{
				"id":     "projects/p/zones/us-central1-a/instances/labeled-vm",
				"name":   "labeled-vm",
				"labels": map[string]interface{}{"env": "prod", "team": "platform"},
			},
		},
	}

	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/p/zones/us-central1-a/instances/labeled-vm",
			Type: "google_compute_instance",
			Name: "labeled-vm",
			Attributes: map[string]interface{}{
				"name": "labeled-vm",
			},
			Labels: map[string]string{"env": "staging", "team": "platform", "goog-managed": "true"},
		},
	}

	result := CompareStateWithActual(tfResources, gcpResources)

	if len(result.ModifiedResources) != 1 {
		t.Fatalf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}

	// Should detect label drift (env changed, goog-managed ignored)
	foundLabelDiff := false
	for _, d := range result.ModifiedResources[0].Differences {
		if d.Field == "labels" {
			foundLabelDiff = true
		}
	}
	if !foundLabelDiff {
		t.Error("expected labels difference not found")
	}
}

func TestGetComparableFields(t *testing.T) {
	tests := []struct {
		resourceType string
		minFields    int
	}{
		{"google_compute_network", 3},
		{"google_compute_subnetwork", 3},
		{"google_compute_firewall", 4},
		{"google_compute_instance", 3},
		{"google_storage_bucket", 3},
		{"google_sql_database_instance", 4},
		{"google_container_cluster", 4},
		{"google_cloud_run_v2_service", 2},
		{"unknown_type", 0},
	}

	for _, tt := range tests {
		fields := getComparableFields(tt.resourceType)
		if len(fields) < tt.minFields {
			t.Errorf("getComparableFields(%s): expected at least %d fields, got %d",
				tt.resourceType, tt.minFields, len(fields))
		}
	}
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     interface{}
		expected bool
	}{
		{"nil-nil", nil, nil, true},
		{"nil-string", nil, "value", false},
		{"string-nil", "value", nil, false},
		{"same-string", "hello", "hello", true},
		{"diff-string", "hello", "world", false},
		{"bool-true", true, true, true},
		{"bool-false", false, false, true},
		{"bool-diff", true, false, false},
		{"bool-string-true", true, "true", true},
		{"bool-string-false", false, "false", true},
		{"int-same", 42, 42, true},
		{"int-diff", 42, 43, false},
	}

	for _, tt := range tests {
		result := valuesEqual(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("valuesEqual(%v, %v) [%s]: expected %v, got %v",
				tt.a, tt.b, tt.name, tt.expected, result)
		}
	}
}

func TestLabelsEqual(t *testing.T) {
	tests := []struct {
		name      string
		tfAttrs   map[string]interface{}
		gcpLabels map[string]string
		expected  bool
	}{
		{
			name:      "both-empty",
			tfAttrs:   map[string]interface{}{},
			gcpLabels: map[string]string{},
			expected:  true,
		},
		{
			name: "equal-labels",
			tfAttrs: map[string]interface{}{
				"labels": map[string]interface{}{"env": "prod"},
			},
			gcpLabels: map[string]string{"env": "prod"},
			expected:  true,
		},
		{
			name: "gcp-managed-labels-ignored",
			tfAttrs: map[string]interface{}{
				"labels": map[string]interface{}{"env": "prod"},
			},
			gcpLabels: map[string]string{"env": "prod", "goog-managed": "true", "gke-cluster": "test"},
			expected:  true,
		},
		{
			name: "different-labels",
			tfAttrs: map[string]interface{}{
				"labels": map[string]interface{}{"env": "prod"},
			},
			gcpLabels: map[string]string{"env": "staging"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		result := labelsEqual(tt.tfAttrs, tt.gcpLabels)
		if result != tt.expected {
			t.Errorf("labelsEqual [%s]: expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

func TestExtractTFResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource *TerraformResource
		expected string
	}{
		{
			name: "id-field",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{"id": "projects/p/global/networks/vpc"},
			},
			expected: "projects/p/global/networks/vpc",
		},
		{
			name: "self-link-field",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{"self_link": "https://compute.googleapis.com/..."},
			},
			expected: "https://compute.googleapis.com/...",
		},
		{
			name: "name-field-fallback",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{"name": "my-resource"},
			},
			expected: "my-resource",
		},
		{
			name: "empty-attributes",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		result := extractTFResourceID(tt.resource)
		if result != tt.expected {
			t.Errorf("extractTFResourceID [%s]: expected %q, got %q", tt.name, tt.expected, result)
		}
	}
}

// TestDiscoveryHelpers tests the helper functions from discovery.go
func TestZoneToRegion(t *testing.T) {
	tests := []struct {
		zone     string
		expected string
	}{
		{"us-central1-a", "us-central1"},
		{"europe-west1-b", "europe-west1"},
		{"asia-east1-c", "asia-east1"},
		{"us-central1", "us-central1"}, // already a region
	}

	for _, tt := range tests {
		result := zoneToRegion(tt.zone)
		if result != tt.expected {
			t.Errorf("zoneToRegion(%s): expected %s, got %s", tt.zone, tt.expected, result)
		}
	}
}

func TestExtractRegionFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{
			"https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1",
			"us-central1",
		},
		{
			"https://www.googleapis.com/compute/v1/projects/my-project/regions/europe-west1/subnetworks/default",
			"europe-west1",
		},
		{"us-central1", "us-central1"}, // no regions/ prefix
	}

	for _, tt := range tests {
		result := extractRegionFromURL(tt.url)
		if result != tt.expected {
			t.Errorf("extractRegionFromURL(%s): expected %s, got %s", tt.url, tt.expected, result)
		}
	}
}

func TestExtractLastSegment(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://compute.googleapis.com/compute/v1/projects/p/zones/z/machineTypes/e2-medium", "e2-medium"},
		{"simple-name", "simple-name"},
		{"a/b/c", "c"},
		{"", ""},  // empty string edge case
		{"/", ""}, // single slash returns empty
	}

	for _, tt := range tests {
		result := extractLastSegment(tt.url)
		if result != tt.expected {
			t.Errorf("extractLastSegment(%s): expected %s, got %s", tt.url, tt.expected, result)
		}
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"us-central1", "europe-west1", "asia-east1"}

	if !containsString(slice, "us-central1") {
		t.Error("expected containsString to find us-central1")
	}
	if containsString(slice, "us-east1") {
		t.Error("expected containsString not to find us-east1")
	}
	if containsString(nil, "anything") {
		t.Error("expected containsString to return false for nil slice")
	}
}

// TestMatchesByName tests the matchesByName function
func TestMatchesByName_Match(t *testing.T) {
	gcpRes := &DiscoveredResource{
		Type: "google_compute_instance",
		Name: "web-server",
	}

	tfMap := map[string]interface{}{
		"tf-1": &TerraformResource{
			Type: "google_compute_instance",
			Attributes: map[string]interface{}{
				"name": "web-server",
			},
		},
	}

	result := matchesByName(gcpRes, tfMap)
	if !result {
		t.Error("expected matchesByName to find matching resource")
	}
}

// TestMatchesByName_NoMatch_DifferentName tests matchesByName with different name
func TestMatchesByName_NoMatch_DifferentName(t *testing.T) {
	gcpRes := &DiscoveredResource{
		Type: "google_compute_instance",
		Name: "web-server",
	}

	tfMap := map[string]interface{}{
		"tf-1": &TerraformResource{
			Type: "google_compute_instance",
			Attributes: map[string]interface{}{
				"name": "api-server",
			},
		},
	}

	result := matchesByName(gcpRes, tfMap)
	if result {
		t.Error("expected matchesByName not to find match with different name")
	}
}

// TestMatchesByName_NoMatch_DifferentType tests matchesByName with different type
func TestMatchesByName_NoMatch_DifferentType(t *testing.T) {
	gcpRes := &DiscoveredResource{
		Type: "google_compute_instance",
		Name: "resource-1",
	}

	tfMap := map[string]interface{}{
		"tf-1": &TerraformResource{
			Type: "google_storage_bucket",
			Attributes: map[string]interface{}{
				"name": "resource-1",
			},
		},
	}

	result := matchesByName(gcpRes, tfMap)
	if result {
		t.Error("expected matchesByName not to find match with different type")
	}
}

// TestMatchesByName_NoNameAttribute tests matchesByName with missing name attribute
func TestMatchesByName_NoNameAttribute(t *testing.T) {
	gcpRes := &DiscoveredResource{
		Type: "google_compute_instance",
		Name: "web-server",
	}

	tfMap := map[string]interface{}{
		"tf-1": &TerraformResource{
			Type: "google_compute_instance",
			Attributes: map[string]interface{}{
				"zone": "us-central1-a",
			},
		},
	}

	result := matchesByName(gcpRes, tfMap)
	if result {
		t.Error("expected matchesByName not to match when name attribute missing")
	}
}

// TestMatchesByName_EmptyMap tests matchesByName with empty map
func TestMatchesByName_EmptyMap(t *testing.T) {
	gcpRes := &DiscoveredResource{
		Type: "google_compute_instance",
		Name: "web-server",
	}

	tfMap := make(map[string]interface{})

	result := matchesByName(gcpRes, tfMap)
	if result {
		t.Error("expected matchesByName not to match empty map")
	}
}

// TestFindByName_Success tests findByName finding a matching resource
func TestFindByName_Success(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_instance",
		Attributes: map[string]interface{}{
			"name": "web-server",
		},
	}

	gcpMap := map[string]*DiscoveredResource{
		"gcp-1": {
			Type: "google_compute_instance",
			Name: "web-server",
			ID:   "projects/p/zones/us-central1-a/instances/web-server",
		},
		"gcp-2": {
			Type: "google_storage_bucket",
			Name: "my-bucket",
			ID:   "my-bucket",
		},
	}

	result := findByName(tfRes, gcpMap)
	if result == nil {
		t.Fatal("expected findByName to return a result")
	}
	if result.Name != "web-server" {
		t.Errorf("expected name web-server, got %s", result.Name)
	}
}

// TestFindByName_NoMatch tests findByName with no matching resource
func TestFindByName_NoMatch(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_instance",
		Attributes: map[string]interface{}{
			"name": "missing-server",
		},
	}

	gcpMap := map[string]*DiscoveredResource{
		"gcp-1": {
			Type: "google_compute_instance",
			Name: "web-server",
			ID:   "projects/p/zones/us-central1-a/instances/web-server",
		},
	}

	result := findByName(tfRes, gcpMap)
	if result != nil {
		t.Error("expected findByName to return nil")
	}
}

// TestFindByName_DifferentType tests findByName with different type
func TestFindByName_DifferentType(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_storage_bucket",
		Attributes: map[string]interface{}{
			"name": "my-bucket",
		},
	}

	gcpMap := map[string]*DiscoveredResource{
		"gcp-1": {
			Type: "google_compute_instance",
			Name: "my-bucket",
			ID:   "projects/p/zones/us-central1-a/instances/my-bucket",
		},
	}

	result := findByName(tfRes, gcpMap)
	if result != nil {
		t.Error("expected findByName to return nil for different type")
	}
}

// TestFindByName_NoNameAttribute tests findByName with missing name attribute
func TestFindByName_NoNameAttribute(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_instance",
		Attributes: map[string]interface{}{
			"zone": "us-central1-a",
		},
	}

	gcpMap := map[string]*DiscoveredResource{
		"gcp-1": {
			Type: "google_compute_instance",
			Name: "web-server",
			ID:   "projects/p/zones/us-central1-a/instances/web-server",
		},
	}

	result := findByName(tfRes, gcpMap)
	if result != nil {
		t.Error("expected findByName to return nil when name attribute missing")
	}
}

// TestFindByNameInterface_Success tests findByNameInterface finding a matching resource
func TestFindByNameInterface_Success(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_network",
		Attributes: map[string]interface{}{
			"name": "vpc-1",
		},
	}

	gcpMap := map[string]interface{}{
		"gcp-1": &DiscoveredResource{
			Type: "google_compute_network",
			Name: "vpc-1",
			ID:   "projects/p/global/networks/vpc-1",
		},
		"gcp-2": &DiscoveredResource{
			Type: "google_compute_network",
			Name: "default",
			ID:   "projects/p/global/networks/default",
		},
	}

	result := findByNameInterface(tfRes, gcpMap)
	if result == nil {
		t.Fatal("expected findByNameInterface to return a result")
	}
	gcpRes := result.(*DiscoveredResource)
	if gcpRes.Name != "vpc-1" {
		t.Errorf("expected name vpc-1, got %s", gcpRes.Name)
	}
}

// TestFindByNameInterface_NoMatch tests findByNameInterface with no matching resource
func TestFindByNameInterface_NoMatch(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_network",
		Attributes: map[string]interface{}{
			"name": "missing-vpc",
		},
	}

	gcpMap := map[string]interface{}{
		"gcp-1": &DiscoveredResource{
			Type: "google_compute_network",
			Name: "vpc-1",
			ID:   "projects/p/global/networks/vpc-1",
		},
	}

	result := findByNameInterface(tfRes, gcpMap)
	if result != nil {
		t.Error("expected findByNameInterface to return nil")
	}
}
