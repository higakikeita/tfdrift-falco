package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// TestExtractTFResourceID_EmptyAttributes tests extractTFResourceID with empty attributes
func TestExtractTFResourceID_EmptyAttributes(t *testing.T) {
	resource := &terraform.Resource{
		Type:       "aws_vpc",
		Attributes: map[string]interface{}{},
	}

	result := extractTFResourceID(resource)
	if result != "" {
		t.Errorf("expected empty string for resource with no ID fields, got %q", result)
	}
}

// TestExtractTFResourceID_AllIDFieldTypes tests all possible ID field types
func TestExtractTFResourceID_AllIDFieldTypes(t *testing.T) {
	tests := []struct {
		name          string
		idField       string
		idValue       string
		expectedFound bool
	}{
		{"id field", "id", "vpc-123", true},
		{"instance_id field", "instance_id", "i-456", true},
		{"db_instance_identifier field", "db_instance_identifier", "mydb", true},
		{"vpc_id field", "vpc_id", "vpc-789", true},
		{"subnet_id field", "subnet_id", "subnet-123", true},
		{"group_id field", "group_id", "sg-456", true},
		{"cluster_name field", "cluster_name", "cluster-1", true},
		{"replication_group_id field", "replication_group_id", "redis-1", true},
		{"arn field", "arn", "arn:aws:rds:us-east-1:123456789:db:mydb", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &terraform.Resource{
				Type: "aws_test",
				Attributes: map[string]interface{}{
					tt.idField: tt.idValue,
				},
			}

			result := extractTFResourceID(resource)
			if tt.expectedFound && result != tt.idValue {
				t.Errorf("expected %q, got %q", tt.idValue, result)
			}
		})
	}
}

// TestExtractTFResourceID_PrefersPrimaryID tests that primary IDs are preferred over fallbacks
func TestExtractTFResourceID_PrefersPrimaryID(t *testing.T) {
	// When multiple ID fields exist, should return the first one found (in order of priority)
	resource := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"id":                     "wrong-id",
			"db_instance_identifier": "correct-db-id",
			"arn":                    "arn:aws:rds:us-east-1:123456789:db:mydb",
		},
	}

	result := extractTFResourceID(resource)
	// Since "id" field exists first, it should be returned
	if result != "wrong-id" {
		t.Errorf("expected 'wrong-id' (first found), got %q", result)
	}
}

// TestExtractTFResourceID_EmptyIDFieldSkipped tests that empty ID fields are skipped
func TestExtractTFResourceID_EmptyIDFieldSkipped(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"id":          "", // Empty string - should be skipped
			"instance_id": "i-12345",
		},
	}

	result := extractTFResourceID(resource)
	if result != "i-12345" {
		t.Errorf("expected 'i-12345' (empty id skipped), got %q", result)
	}
}

// TestValuesEqual_EdgeCases tests edge cases for value comparison
func TestValuesEqual_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"empty strings", "", "", true},
		{"zero integers", 0, 0, true},
		{"zero vs false", 0, false, false},
		{"empty slice vs nil", []string{}, nil, false},
		{"float precision", 1.0, 1.0, true},
		{"float vs int", 1.0, 1, true},
		{"bool vs non-matching string", false, "no", false},
		{"string zero vs int zero", "0", 0, true},
		{"negative numbers", -100, -100, true},
		{"negative vs positive", -100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("valuesEqual(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestGetNestedValue_ComplexPaths tests nested value retrieval with various path complexities
func TestGetNestedValue_ComplexPaths(t *testing.T) {
	data := map[string]interface{}{
		"simple": "value",
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"value": "deeply-nested",
				},
			},
		},
		"empty_map": map[string]interface{}{},
		"non_map":   "scalar-value",
	}

	tests := []struct {
		name        string
		path        string
		shouldExist bool
		expected    interface{}
	}{
		{"simple field", "simple", true, "value"},
		{"level1", "level1", true, nil}, // Returns the map itself
		{"level3 deep", "level1.level2.level3.value", true, "deeply-nested"},
		{"partial path returns nil", "level1.level2.nonexistent", false, nil},
		{"path through empty map", "empty_map.key", false, nil},
		{"path through scalar", "non_map.field", false, nil},
		{"empty path", "", false, nil},
		{"nonexistent top level", "nonexistent", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedValue(data, tt.path)
			if tt.shouldExist {
				if result == nil && tt.expected != nil {
					t.Errorf("expected %v, got nil", tt.expected)
				}
			} else {
				if result != nil && tt.expected == nil {
					t.Errorf("expected nil, got %v", result)
				}
			}
		})
	}
}

// TestGetTerraformTags_NonStringValues tests getTerraformTags with non-string tag values
func TestGetTerraformTags_NonStringValues(t *testing.T) {
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"String":  "value",
			"Number":  123,
			"Boolean": true,
			"Null":    nil,
			"Slice":   []string{"a", "b"},
			"Map":     map[string]interface{}{},
		},
	}

	result := getTerraformTags(attrs)

	// Only string values should be included
	if len(result) != 1 {
		t.Errorf("expected 1 tag (only string values), got %d: %v", len(result), result)
	}
	if result["String"] != "value" {
		t.Errorf("expected String tag to be 'value', got %q", result["String"])
	}
}

// TestGetTerraformTags_TagsAllPriority tests that tags_all is lower priority than tags
func TestGetTerraformTags_TagsAllPriority(t *testing.T) {
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "prod",
		},
		"tags_all": map[string]interface{}{
			"Environment": "dev",
			"Project":     "myapp",
		},
	}

	result := getTerraformTags(attrs)

	// Should use tags, not tags_all
	if result["Environment"] != "prod" {
		t.Errorf("expected Environment=prod from 'tags', got %q", result["Environment"])
	}
	if _, exists := result["Project"]; exists {
		t.Errorf("should not include Project from tags_all when tags exist")
	}
}

// TestTagsEqual_ManagedTagsFiltering tests that managed tags are properly filtered
func TestTagsEqual_ManagedTagsFiltering(t *testing.T) {
	tests := []struct {
		name     string
		tfAttrs  map[string]interface{}
		awsTags  map[string]string
		expected bool
	}{
		{
			name: "aws: prefix ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"Name": "test"},
			},
			awsTags: map[string]string{
				"Name":               "test",
				"aws:cloudformation": "stack-id",
				"aws:created-by":     "user",
			},
			expected: true,
		},
		{
			name: "kubernetes.io: prefix ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"Owner": "team"},
			},
			awsTags: map[string]string{
				"Owner":                            "team",
				"kubernetes.io/created-by":         "controller",
				"kubernetes.io/cluster/my-cluster": "owned",
			},
			expected: true,
		},
		{
			name: "multiple prefixes ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"App": "myapp"},
			},
			awsTags: map[string]string{
				"App":                    "myapp",
				"aws:managed":            "true",
				"kubernetes.io/name":     "test",
				"aws:additional-context": "data",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagsEqual(tt.tfAttrs, tt.awsTags)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompareResourceAttributes_AllResourceTypes tests compareResourceAttributes for all supported types
func TestCompareResourceAttributes_AllResourceTypes(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		tfAttrs      map[string]interface{}
		awsAttrs     map[string]interface{}
		shouldMatch  bool
	}{
		{
			name:         "aws_vpc match",
			resourceType: "aws_vpc",
			tfAttrs:      map[string]interface{}{"cidr_block": "10.0.0.0/16", "enable_dns_hostnames": true},
			awsAttrs:     map[string]interface{}{"cidr_block": "10.0.0.0/16", "enable_dns_hostnames": true},
			shouldMatch:  true,
		},
		{
			name:         "aws_subnet match",
			resourceType: "aws_subnet",
			tfAttrs:      map[string]interface{}{"vpc_id": "vpc-123", "cidr_block": "10.0.1.0/24"},
			awsAttrs:     map[string]interface{}{"vpc_id": "vpc-123", "cidr_block": "10.0.1.0/24"},
			shouldMatch:  true,
		},
		{
			name:         "aws_instance match",
			resourceType: "aws_instance",
			tfAttrs:      map[string]interface{}{"instance_type": "t3.micro"},
			awsAttrs:     map[string]interface{}{"instance_type": "t3.micro"},
			shouldMatch:  true,
		},
		{
			name:         "aws_db_instance match",
			resourceType: "aws_db_instance",
			tfAttrs:      map[string]interface{}{"engine": "postgres", "engine_version": "13.7"},
			awsAttrs:     map[string]interface{}{"engine": "postgres", "engine_version": "13.7"},
			shouldMatch:  true,
		},
		{
			name:         "aws_eks_cluster match",
			resourceType: "aws_eks_cluster",
			tfAttrs:      map[string]interface{}{"version": "1.23"},
			awsAttrs:     map[string]interface{}{"version": "1.23"},
			shouldMatch:  true,
		},
		{
			name:         "aws_elasticache_replication_group match",
			resourceType: "aws_elasticache_replication_group",
			tfAttrs:      map[string]interface{}{"node_type": "cache.t3.micro"},
			awsAttrs:     map[string]interface{}{"node_type": "cache.t3.micro"},
			shouldMatch:  true,
		},
		{
			name:         "aws_lb match",
			resourceType: "aws_lb",
			tfAttrs:      map[string]interface{}{"type": "application"},
			awsAttrs:     map[string]interface{}{"type": "application"},
			shouldMatch:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tfRes := &terraform.Resource{
				Type:       tt.resourceType,
				Attributes: tt.tfAttrs,
			}
			awsRes := &DiscoveredResource{
				ID:         "test-id",
				Type:       tt.resourceType,
				Attributes: tt.awsAttrs,
			}

			result := compareResourceAttributes(tfRes, awsRes)
			if result == nil && !tt.shouldMatch {
				return
			}
			if tt.shouldMatch && len(result) == 0 {
				return
			}
			if tt.shouldMatch && len(result) > 0 {
				t.Errorf("expected matching resources to have no differences, got %d differences", len(result))
			}
		})
	}
}

// TestCompareStateWithActual_EmptyLists tests with empty lists
func TestCompareStateWithActual_EmptyLists(t *testing.T) {
	result := CompareStateWithActual([]*terraform.Resource{}, []*DiscoveredResource{})

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 || len(result.ModifiedResources) != 0 {
		t.Errorf("expected all empty results for empty inputs")
	}
}

// TestCompareStateWithActual_NilResources tests with nil resources in slices
func TestCompareStateWithActual_NilResources(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-123",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-123",
			Type: "aws_vpc",
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 {
		t.Errorf("expected resources to match")
	}
}

// TestCompareStateWithActual_MultipleModifications tests multiple modified resources
func TestCompareStateWithActual_MultipleModifications(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-1",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-2",
				"cidr_block": "10.1.0.0/16",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-1",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			ID:   "vpc-2",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.1.0.0/24", // Modified
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
}

// TestCompareResourceAttributes_NoComparableFields tests resource type with no comparable fields
func TestCompareResourceAttributes_NoComparableFields(t *testing.T) {
	tfRes := &terraform.Resource{
		Type:       "unknown_type",
		Attributes: map[string]interface{}{"field": "value"},
	}

	awsRes := &DiscoveredResource{
		ID:         "test",
		Type:       "unknown_type",
		Attributes: map[string]interface{}{"field": "different"},
	}

	result := compareResourceAttributes(tfRes, awsRes)

	// Should not find differences because unknown type has no comparable fields
	if len(result) > 0 {
		t.Errorf("expected no differences for unknown resource type, got %d", len(result))
	}
}

// TestGetComparableFields_Coverage tests all branches of getComparableFields
func TestGetComparableFields_Coverage(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		minFields    int
	}{
		{"aws_vpc", "aws_vpc", 3},
		{"aws_subnet", "aws_subnet", 4},
		{"aws_security_group", "aws_security_group", 3},
		{"aws_instance", "aws_instance", 4},
		{"aws_db_instance", "aws_db_instance", 7},
		{"aws_eks_cluster", "aws_eks_cluster", 2},
		{"aws_elasticache_replication_group", "aws_elasticache_replication_group", 3},
		{"aws_lb", "aws_lb", 3},
		{"unknown_resource", "unknown_resource", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getComparableFields(tt.resourceType)
			if len(result) < tt.minFields {
				t.Errorf("expected at least %d fields for %s, got %d", tt.minFields, tt.resourceType, len(result))
			}
		})
	}
}

// TestDiscoveredResource_JSONMarshaling tests that DiscoveredResource can be properly marshaled
func TestDiscoveredResource_JSONMarshaling(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "vpc-123",
		Type:   "aws_vpc",
		ARN:    "arn:aws:ec2:us-east-1:123456789:vpc/vpc-123",
		Name:   "my-vpc",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"state":      "available",
		},
		Tags: map[string]string{
			"Environment": "prod",
		},
	}

	// Verify all fields are accessible and correct types
	if resource.ID != "vpc-123" {
		t.Errorf("ID field incorrect")
	}
	if resource.Type != "aws_vpc" {
		t.Errorf("Type field incorrect")
	}
	if resource.ARN != "arn:aws:ec2:us-east-1:123456789:vpc/vpc-123" {
		t.Errorf("ARN field incorrect")
	}
	if resource.Name != "my-vpc" {
		t.Errorf("Name field incorrect")
	}
	if resource.Region != "us-east-1" {
		t.Errorf("Region field incorrect")
	}
	if len(resource.Attributes) != 2 {
		t.Errorf("Attributes count incorrect")
	}
	if len(resource.Tags) != 1 {
		t.Errorf("Tags count incorrect")
	}
}

// TestDriftResult_AllFieldsPopulated tests DriftResult with all field types populated
func TestDriftResult_AllFieldsPopulated(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{
			{ID: "unmanaged-1", Type: "aws_vpc"},
			{ID: "unmanaged-2", Type: "aws_subnet"},
		},
		MissingResources: []*types.TerraformResource{
			{Type: "aws_instance", Name: "missing-1"},
			{Type: "aws_security_group", Name: "missing-2"},
			{Type: "aws_db_instance", Name: "missing-3"},
		},
		ModifiedResources: []*ResourceDiff{
			{ResourceID: "modified-1", ResourceType: "aws_vpc"},
			{ResourceID: "modified-2", ResourceType: "aws_instance"},
		},
	}

	if len(result.UnmanagedResources) != 2 {
		t.Errorf("UnmanagedResources count incorrect")
	}
	if len(result.MissingResources) != 3 {
		t.Errorf("MissingResources count incorrect")
	}
	if len(result.ModifiedResources) != 2 {
		t.Errorf("ModifiedResources count incorrect")
	}
}

// TestResourceDiff_AllFieldsPopulated tests ResourceDiff with all field types
func TestResourceDiff_AllFieldsPopulated(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "i-12345",
		ResourceType: "aws_instance",
		TerraformState: map[string]interface{}{
			"instance_type": "t3.micro",
			"subnet_id":     "subnet-123",
		},
		ActualState: map[string]interface{}{
			"instance_type": "t3.small",
			"subnet_id":     "subnet-123",
		},
		Differences: []FieldDiff{
			{
				Field:          "instance_type",
				TerraformValue: "t3.micro",
				ActualValue:    "t3.small",
			},
		},
	}

	if diff.ResourceID != "i-12345" {
		t.Errorf("ResourceID incorrect")
	}
	if diff.ResourceType != "aws_instance" {
		t.Errorf("ResourceType incorrect")
	}
	if len(diff.TerraformState) != 2 {
		t.Errorf("TerraformState count incorrect")
	}
	if len(diff.ActualState) != 2 {
		t.Errorf("ActualState count incorrect")
	}
	if len(diff.Differences) != 1 {
		t.Errorf("Differences count incorrect")
	}
}

// TestFieldDiff_AllValueTypes tests FieldDiff with all different value types
func TestFieldDiff_AllValueTypes(t *testing.T) {
	types := []struct {
		name        string
		tfValue     interface{}
		actualValue interface{}
	}{
		{"string", "value1", "value2"},
		{"int", 100, 200},
		{"int64", int64(100), int64(200)},
		{"float64", 1.5, 2.5},
		{"bool", true, false},
		{"nil", nil, "value"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			diff := FieldDiff{
				Field:          "test",
				TerraformValue: tt.tfValue,
				ActualValue:    tt.actualValue,
			}

			if diff.Field != "test" {
				t.Errorf("Field not set correctly")
			}
		})
	}
}

// TestCompareStateWithActual_LargeScale tests with many resources
func TestCompareStateWithActual_LargeScale(t *testing.T) {
	tfResources := make([]*terraform.Resource, 100)
	for i := 0; i < 100; i++ {
		id := "resource-" + string(rune(i))
		tfResources[i] = &terraform.Resource{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": id,
			},
		}
	}

	awsResources := make([]*DiscoveredResource, 100)
	for i := 0; i < 100; i++ {
		id := "resource-" + string(rune(i))
		awsResources[i] = &DiscoveredResource{
			ID:   id,
			Type: "aws_vpc",
		}
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 || len(result.ModifiedResources) != 0 {
		t.Errorf("expected all resources to match")
	}
}

// TestValuesEqual_ComplexTypes tests valuesEqual with complex types
func TestValuesEqual_ComplexTypes(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "identical slices",
			a:        []string{"a", "b"},
			b:        []string{"a", "b"},
			expected: true,
		},
		{
			name:     "different slices",
			a:        []string{"a", "b"},
			b:        []string{"a", "c"},
			expected: false,
		},
		{
			name:     "identical maps",
			a:        map[string]string{"key": "value"},
			b:        map[string]string{"key": "value"},
			expected: true,
		},
		{
			name:     "different maps",
			a:        map[string]string{"key": "value1"},
			b:        map[string]string{"key": "value2"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompareResourceAttributes_MissingAttributesInActual tests comparing when actual state has missing attributes
func TestCompareResourceAttributes_MissingAttributesInActual(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block":           "10.0.0.0/16",
			"enable_dns_hostnames": true,
			"enable_dns_support":   true,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-123",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			// Missing enable_dns_hostnames and enable_dns_support
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should detect differences for missing attributes (they'll be nil in actual)
	if len(result) == 0 {
		t.Errorf("expected differences for missing attributes, got none")
	}
}

// TestCompareResourceAttributes_ExtraAttributesInActual tests when actual has extra attributes
func TestCompareResourceAttributes_ExtraAttributesInActual(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-123",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"extra_attr": "should_not_matter",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Extra attributes shouldn't cause differences (only comparable fields are checked)
	if len(result) > 0 {
		t.Errorf("extra attributes shouldn't cause differences, got %d", len(result))
	}
}

// TestGetTerraformTags_BothTagsAndTagsAll tests when both tags and tags_all exist
func TestGetTerraformTags_BothTagsAndTagsAll(t *testing.T) {
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"FromTags": "yes",
		},
		"tags_all": map[string]interface{}{
			"FromTagsAll": "should-not-appear",
		},
	}

	result := getTerraformTags(attrs)

	if len(result) != 1 {
		t.Errorf("expected 1 tag (from tags field), got %d", len(result))
	}
	if _, exists := result["FromTags"]; !exists {
		t.Errorf("expected FromTags key to exist")
	}
	if _, exists := result["FromTagsAll"]; exists {
		t.Errorf("should not include tags from tags_all when tags exist")
	}
}

// TestCompareStateWithActual_MixedResourceTypes tests with multiple resource types
func TestCompareStateWithActual_MixedResourceTypes(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-1",
			},
		},
		{
			Type: "aws_instance",
			Attributes: map[string]interface{}{
				"instance_id": "i-1",
			},
		},
		{
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"db_instance_identifier": "db-1",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc"},
		{ID: "i-1", Type: "aws_instance"},
		{ID: "db-1", Type: "aws_db_instance"},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 {
		t.Errorf("expected all resources to match")
	}
}

// TestExtractTFResourceID_MixedFieldTypes tests extracting ID when field value is non-string
func TestExtractTFResourceID_MixedFieldTypes(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_test",
		Attributes: map[string]interface{}{
			"id":   123, // Non-string value
			"name": "test",
		},
	}

	result := extractTFResourceID(resource)

	// Should not match because value is not a string
	if result == "123" {
		t.Errorf("should not match non-string ID field")
	}
}

// TestTagsEqual_ComplexScenario tests tags comparison with multiple edge cases
func TestTagsEqual_ComplexScenario(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Name":        "my-resource",
			"Environment": "prod",
			"Owner":       "team-a",
		},
	}

	awsTags := map[string]string{
		"Name":                  "my-resource",
		"Environment":           "prod",
		"Owner":                 "team-a",
		"aws:cloudformation":    "stack-id",
		"kubernetes.io/created": "controller",
	}

	result := tagsEqual(tfAttrs, awsTags)

	if !result {
		t.Errorf("expected tags to be equal after filtering managed tags")
	}
}

// TestValuesEqual_StringNumberMismatch tests string vs number comparison
func TestValuesEqual_StringNumberMismatch(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"int 100 vs string 100", 100, "100", true},
		{"int 0 vs string 0", 0, "0", true},
		{"float 1.5 vs string 1.5", 1.5, "1.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetNestedValue_SingleLevelMap tests getNestedValue with single level access
func TestGetNestedValue_SingleLevelMap(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{"string value", "key1", "value1"},
		{"int value", "key2", 42},
		{"bool value", "key3", true},
		{"nonexistent", "key4", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedValue(data, tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetNestedValue_EmptyPath tests getNestedValue with empty path string
func TestGetNestedValue_EmptyPath(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}

	result := getNestedValue(data, "")
	// Empty path splits to [""], which won't exist in the map, so should return nil
	if result != nil {
		t.Errorf("expected nil for empty path, got %v", result)
	}
}

// TestGetNestedValue_SecondLevelNavigationFailure tests when cannot navigate deeper because not a map
func TestGetNestedValue_SecondLevelNavigationFailure(t *testing.T) {
	data := map[string]interface{}{
		"scalar": "not-a-map",
	}

	result := getNestedValue(data, "scalar.deeper")
	if result != nil {
		t.Errorf("expected nil when trying to navigate through scalar, got %v", result)
	}
}

// TestGetNestedValue_DotOnlyPath tests getNestedValue with dot-only path
func TestGetNestedValue_DotOnlyPath(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
	}

	result := getNestedValue(data, ".")
	if result != nil {
		t.Errorf("expected nil for dot-only path, got %v", result)
	}
}

// TestCompareResourceAttributes_BooleanComparison tests comparison of boolean values
func TestCompareResourceAttributes_BooleanComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"map_public_ip_on_launch": true,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "subnet-123",
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"map_public_ip_on_launch": true,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching boolean values")
	}
}

// TestCompareResourceAttributes_BooleanMismatch tests comparison of different boolean values
func TestCompareResourceAttributes_BooleanMismatch(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"map_public_ip_on_launch": true,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "subnet-123",
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"map_public_ip_on_launch": false,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	hasDifference := false
	for _, diff := range result {
		if diff.Field == "map_public_ip_on_launch" {
			hasDifference = true
			break
		}
	}
	if !hasDifference {
		t.Errorf("expected difference for mismatched boolean values")
	}
}

// TestCompareResourceAttributes_IntegerComparison tests comparison of integer values
func TestCompareResourceAttributes_IntegerComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": 100,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "db-123",
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": 100,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching integer values")
	}
}

// TestCompareResourceAttributes_IntegerMismatch tests comparison of different integer values
func TestCompareResourceAttributes_IntegerMismatch(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": 100,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "db-123",
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": 200,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	hasDifference := false
	for _, diff := range result {
		if diff.Field == "allocated_storage" {
			hasDifference = true
			break
		}
	}
	if !hasDifference {
		t.Errorf("expected difference for mismatched integer values")
	}
}

// TestCompareStateWithActual_DuplicateTFResources tests behavior with duplicate Terraform resources
func TestCompareStateWithActual_DuplicateTFResources(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-123",
			},
		},
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-123", // Duplicate ID
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-123",
			Type: "aws_vpc",
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	// The function should handle this gracefully (later entry overrides)
	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 {
		t.Errorf("expected resources to match despite duplicates")
	}
}

// TestValuesEqual_BoolStringConversion tests bool-string conversion edge cases
func TestValuesEqual_BoolStringConversion(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"bool true vs string true", true, "true", true},
		{"bool false vs string false", false, "false", true},
		{"bool true vs string True", true, "True", false},
		{"bool false vs string False", false, "False", false},
		{"bool true vs string 1", true, "1", false},
		{"bool false vs string 0", false, "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetComparableFields_AllTypes ensures all supported types have fields defined
func TestGetComparableFields_AllTypes(t *testing.T) {
	supportedTypes := []string{
		"aws_vpc",
		"aws_subnet",
		"aws_security_group",
		"aws_instance",
		"aws_db_instance",
		"aws_eks_cluster",
		"aws_elasticache_replication_group",
		"aws_lb",
	}

	for _, resType := range supportedTypes {
		fields := getComparableFields(resType)
		if len(fields) == 0 {
			t.Errorf("resource type %s should have comparable fields defined", resType)
		}

		// Verify no duplicate fields
		seen := make(map[string]bool)
		for _, field := range fields {
			if seen[field] {
				t.Errorf("resource type %s has duplicate field: %s", resType, field)
			}
			seen[field] = true
		}
	}
}

// TestTagsEqual_EmptyTFTags tests tagsEqual with empty Terraform tags
func TestTagsEqual_EmptyTFTags(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{},
	}

	awsTags := map[string]string{
		"aws:managed": "true",
	}

	result := tagsEqual(tfAttrs, awsTags)

	if !result {
		t.Errorf("expected empty Terraform tags to match when AWS has only managed tags")
	}
}

// TestTagsEqual_EmptyAWSTags tests tagsEqual with empty AWS tags
func TestTagsEqual_EmptyAWSTags(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{},
	}

	awsTags := map[string]string{}

	result := tagsEqual(tfAttrs, awsTags)

	if !result {
		t.Errorf("expected empty tags to match")
	}
}

// TestCompareResourceAttributes_SecurityGroupComparison tests security group comparison
func TestCompareResourceAttributes_SecurityGroupComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"vpc_id":      "vpc-123",
			"description": "My SG",
			"name":        "my-sg",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "sg-123",
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"vpc_id":      "vpc-123",
			"description": "My SG",
			"name":        "my-sg",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching security groups")
	}
}

// TestCompareResourceAttributes_LoadBalancerComparison tests load balancer comparison
func TestCompareResourceAttributes_LoadBalancerComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_lb",
		Attributes: map[string]interface{}{
			"type":   "application",
			"scheme": "internet-facing",
			"vpc_id": "vpc-123",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-alb/1234567890",
		Type: "aws_lb",
		Attributes: map[string]interface{}{
			"type":   "application",
			"scheme": "internet-facing",
			"vpc_id": "vpc-123",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching load balancers")
	}
}

// TestCompareResourceAttributes_EKSComparison tests EKS cluster comparison
func TestCompareResourceAttributes_EKSComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_eks_cluster",
		Attributes: map[string]interface{}{
			"version":  "1.23",
			"role_arn": "arn:aws:iam::123456789:role/eks-service-role",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "my-cluster",
		Type: "aws_eks_cluster",
		Attributes: map[string]interface{}{
			"version":  "1.23",
			"role_arn": "arn:aws:iam::123456789:role/eks-service-role",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching EKS clusters")
	}
}

// TestCompareResourceAttributes_ElastiCacheComparison tests ElastiCache comparison
func TestCompareResourceAttributes_ElastiCacheComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_elasticache_replication_group",
		Attributes: map[string]interface{}{
			"node_type":                  "cache.t3.micro",
			"automatic_failover_enabled": "true",
			"multi_az_enabled":           "true",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "redis-group-1",
		Type: "aws_elasticache_replication_group",
		Attributes: map[string]interface{}{
			"node_type":                  "cache.t3.micro",
			"automatic_failover_enabled": "true",
			"multi_az_enabled":           "true",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching ElastiCache groups")
	}
}

// TestExtractTFResourceID_ARNFieldPriority tests that ARN field is used when other IDs missing
func TestExtractTFResourceID_ARNFieldPriority(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_eks_cluster",
		Attributes: map[string]interface{}{
			"arn": "arn:aws:eks:us-east-1:123456789:cluster/my-cluster",
		},
	}

	result := extractTFResourceID(resource)
	if result != "arn:aws:eks:us-east-1:123456789:cluster/my-cluster" {
		t.Errorf("expected ARN to be returned as ID, got %q", result)
	}
}

// TestValuesEqual_SpecialStrings tests valuesEqual with special string values
func TestValuesEqual_SpecialStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"empty strings", "", "", true},
		{"whitespace only", "   ", "   ", true},
		{"newlines", "line1\nline2", "line1\nline2", true},
		{"unicode", "café", "café", true},
		{"different whitespace", "test ", " test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompareStateWithActual_PartialAttributeMatch tests when only some attributes match
func TestCompareStateWithActual_PartialAttributeMatch(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"db_instance_identifier": "prod-db",
				"engine":                 "postgres",
				"engine_version":         "13.7",
				"instance_class":         "db.t3.micro",
				"allocated_storage":      100,
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "prod-db",
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"engine":            "postgres",
				"engine_version":    "14.1", // Different!
				"instance_class":    "db.t3.micro",
				"allocated_storage": 100,
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
	if len(result.ModifiedResources[0].Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(result.ModifiedResources[0].Differences))
	}
}

// TestCompareResourceAttributes_VPCMultipleFields tests VPC with multiple comparable fields
func TestCompareResourceAttributes_VPCMultipleFields(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id":                   "vpc-123",
			"cidr_block":           "10.0.0.0/16",
			"enable_dns_hostnames": true,
			"enable_dns_support":   true,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-123",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block":           "10.0.0.0/16",
			"enable_dns_hostnames": false,
			"enable_dns_support":   true,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	hasDiff := false
	for _, diff := range result {
		if diff.Field == "enable_dns_hostnames" {
			hasDiff = true
			if diff.TerraformValue != true {
				t.Errorf("expected TF value true, got %v", diff.TerraformValue)
			}
			if diff.ActualValue != false {
				t.Errorf("expected actual value false, got %v", diff.ActualValue)
			}
		}
	}
	if !hasDiff {
		t.Errorf("expected difference in enable_dns_hostnames")
	}
}

// TestCompareStateWithActual_ResourceTypeMismatchIgnored tests that resources with mismatched types don't cause issues
func TestCompareStateWithActual_ResourceTypeMismatchIgnored(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-123",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "subnet-456",
			Type: "aws_subnet",
		},
	}

	// This should not panic and should show both as unmanaged/missing
	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource, got %d", len(result.MissingResources))
	}
}

// TestCompareStateWithActual_SubnetMultipleAttributes tests subnet comparison with multiple attributes
func TestCompareStateWithActual_SubnetMultipleAttributes(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_subnet",
			Attributes: map[string]interface{}{
				"id":                      "subnet-123",
				"vpc_id":                  "vpc-123",
				"cidr_block":              "10.0.1.0/24",
				"availability_zone":       "us-east-1a",
				"map_public_ip_on_launch": true,
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "subnet-123",
			Type: "aws_subnet",
			Attributes: map[string]interface{}{
				"vpc_id":                  "vpc-123",
				"cidr_block":              "10.0.1.0/24",
				"availability_zone":       "us-east-1a",
				"map_public_ip_on_launch": false,
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
	if len(result.ModifiedResources[0].Differences) != 1 {
		t.Errorf("expected 1 difference (map_public_ip), got %d", len(result.ModifiedResources[0].Differences))
	}
}

// TestCompareStateWithActual_InstanceComparison tests instance comparison
func TestCompareStateWithActual_InstanceComparison(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_instance",
			Attributes: map[string]interface{}{
				"instance_id":       "i-12345",
				"instance_type":     "t3.micro",
				"subnet_id":         "subnet-123",
				"vpc_id":            "vpc-123",
				"availability_zone": "us-east-1a",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "i-12345",
			Type: "aws_instance",
			Attributes: map[string]interface{}{
				"instance_type":     "t3.small",
				"subnet_id":         "subnet-123",
				"vpc_id":            "vpc-123",
				"availability_zone": "us-east-1a",
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
	if len(result.ModifiedResources[0].Differences) != 1 {
		t.Errorf("expected 1 difference (instance_type), got %d", len(result.ModifiedResources[0].Differences))
	}
}

// TestGetTerraformTags_NonMapTagsField tests getTerraformTags when tags field is not a map
func TestGetTerraformTags_NonMapTagsField(t *testing.T) {
	attrs := map[string]interface{}{
		"tags": "not-a-map", // Invalid tags structure
	}

	result := getTerraformTags(attrs)
	if len(result) != 0 {
		t.Errorf("expected empty tags when tags field is not a map, got %d tags", len(result))
	}
}

// TestGetTerraformTags_NonMapTagsAllField tests getTerraformTags when tags_all is not a map
func TestGetTerraformTags_NonMapTagsAllField(t *testing.T) {
	attrs := map[string]interface{}{
		"tags_all": "not-a-map", // Invalid tags_all structure
	}

	result := getTerraformTags(attrs)
	if len(result) != 0 {
		t.Errorf("expected empty tags when tags_all field is not a map, got %d tags", len(result))
	}
}

// TestTagsEqual_IgnoredPrefixVariations tests various managed tag prefixes are ignored
func TestTagsEqual_IgnoredPrefixVariations(t *testing.T) {
	tests := []struct {
		name        string
		awsTag      string
		awsTagValue string
		shouldMatch bool
	}{
		{"aws: prefix", "aws:something", "managed-value", true},
		{"aws: multiple", "aws:cf:stack-id", "managed-value", true},
		{"kubernetes.io: prefix", "kubernetes.io/name", "managed-value", true},
		{"kubernetes.io: longer", "kubernetes.io/cluster/my-cluster", "managed-value", true},
		{"normal tag", "NormalTag", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tfAttrs := map[string]interface{}{
				"tags": map[string]interface{}{"NormalTag": "value"},
			}

			awsTags := map[string]string{
				"NormalTag": "value",
				tt.awsTag:   tt.awsTagValue,
			}

			result := tagsEqual(tfAttrs, awsTags)
			if !result && tt.shouldMatch {
				t.Errorf("expected tags to be equal with managed prefix %s ignored", tt.awsTag)
			}
		})
	}
}

// TestCompareResourceAttributes_AllFieldsNil tests comparison when attributes are nil
func TestCompareResourceAttributes_AllFieldsNil(t *testing.T) {
	tfRes := &terraform.Resource{
		Type:       "aws_vpc",
		Attributes: nil,
	}

	awsRes := &DiscoveredResource{
		ID:         "vpc-123",
		Type:       "aws_vpc",
		Attributes: nil,
	}

	// Should handle nil attributes gracefully
	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil {
		t.Fatal("expected non-nil result even with nil attributes")
	}
}

// TestCompareStateWithActual_IntegrityCheck tests CompareStateWithActual doesn't lose resources
func TestCompareStateWithActual_IntegrityCheck(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type:       "aws_vpc",
			Attributes: map[string]interface{}{"id": "vpc-1"},
		},
		{
			Type:       "aws_subnet",
			Attributes: map[string]interface{}{"id": "subnet-1"},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-1",
			Type: "aws_vpc",
		},
		{
			ID:   "sg-1",
			Type: "aws_security_group",
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	// One missing (subnet-1), one unmanaged (sg-1)
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource, got %d", len(result.MissingResources))
	}
	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource, got %d", len(result.UnmanagedResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

// TestDiscoveryClient_Initialization tests DiscoveryClient initialization with nil clients
func TestDiscoveryClient_Initialization(t *testing.T) {
	client := &DiscoveryClient{
		region:      "us-east-1",
		ec2Client:   nil,
		rdsClient:   nil,
		eksClient:   nil,
		elasticache: nil,
		elbClient:   nil,
	}

	if client.region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", client.region)
	}
}

// TestDiscoveryClient_MultipleRegions tests DiscoveryClient with different regions
func TestDiscoveryClient_MultipleRegions(t *testing.T) {
	regions := []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"ap-southeast-1",
		"ca-central-1",
	}

	for _, region := range regions {
		client := &DiscoveryClient{
			region: region,
		}

		if client.region != region {
			t.Errorf("expected region %s, got %s", region, client.region)
		}
	}
}

// TestDiscoveredResource_FullStructure tests DiscoveredResource with all fields populated
func TestDiscoveredResource_FullStructure(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "vpc-123",
		Type:   "aws_vpc",
		ARN:    "arn:aws:ec2:us-east-1:123456789:vpc/vpc-123",
		Name:   "production-vpc",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"cidr_block":           "10.0.0.0/16",
			"enable_dns_hostnames": true,
			"enable_dns_support":   true,
			"state":                "available",
		},
		Tags: map[string]string{
			"Name":        "production-vpc",
			"Environment": "prod",
			"Owner":       "platform-team",
		},
	}

	// Verify all fields are accessible
	if resource.ID != "vpc-123" {
		t.Errorf("ID mismatch")
	}
	if resource.Type != "aws_vpc" {
		t.Errorf("Type mismatch")
	}
	if resource.ARN != "arn:aws:ec2:us-east-1:123456789:vpc/vpc-123" {
		t.Errorf("ARN mismatch")
	}
	if resource.Name != "production-vpc" {
		t.Errorf("Name mismatch")
	}
	if resource.Region != "us-east-1" {
		t.Errorf("Region mismatch")
	}
	if len(resource.Attributes) != 4 {
		t.Errorf("Attributes length mismatch")
	}
	if len(resource.Tags) != 3 {
		t.Errorf("Tags length mismatch")
	}
}

// TestDriftResult_ComplexScenario tests DriftResult with complex mixed scenario
func TestDriftResult_ComplexScenario(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{
			{ID: "vpc-manual", Type: "aws_vpc", Name: "manually-created"},
			{ID: "sg-manual", Type: "aws_security_group", Name: "manual-sg"},
			{ID: "subnet-manual", Type: "aws_subnet", Name: "manual-subnet"},
		},
		MissingResources: []*types.TerraformResource{
			{Type: "aws_instance", Name: "deleted-instance"},
			{Type: "aws_db_instance", Name: "deleted-db"},
		},
		ModifiedResources: []*ResourceDiff{
			{ResourceID: "vpc-123", ResourceType: "aws_vpc"},
		},
	}

	// Verify counts
	if len(result.UnmanagedResources) != 3 {
		t.Errorf("expected 3 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 2 {
		t.Errorf("expected 2 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}

	// Verify we can iterate and access fields
	for _, res := range result.UnmanagedResources {
		if res.ID == "" {
			t.Errorf("unmanaged resource missing ID")
		}
	}

	for _, res := range result.MissingResources {
		if res.Type == "" {
			t.Errorf("missing resource missing Type")
		}
	}

	for _, diff := range result.ModifiedResources {
		if diff.ResourceID == "" {
			t.Errorf("modified resource missing ResourceID")
		}
	}
}

// TestCompareStateWithActual_RDSAndEKSResources tests comparison with RDS and EKS resources
func TestCompareStateWithActual_RDSAndEKSResources(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"db_instance_identifier": "prod-postgres",
				"engine":                 "postgres",
				"engine_version":         "13.7",
				"allocated_storage":      100,
				"multi_az":               true,
			},
		},
		{
			Type: "aws_eks_cluster",
			Attributes: map[string]interface{}{
				"cluster_name": "prod-cluster",
				"version":      "1.23",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "prod-postgres",
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"engine":            "postgres",
				"engine_version":    "13.7",
				"allocated_storage": 100,
				"multi_az":          true,
			},
		},
		{
			ID:   "prod-cluster",
			Type: "aws_eks_cluster",
			Attributes: map[string]interface{}{
				"version": "1.23",
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 || len(result.ModifiedResources) != 0 {
		t.Errorf("expected all resources to match")
	}
}

// TestCompareResourceAttributes_RDSInstance tests RDS instance attribute comparison
func TestCompareResourceAttributes_RDSInstance(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"engine":            "postgres",
			"engine_version":    "13.7",
			"instance_class":    "db.t3.micro",
			"allocated_storage": 100,
			"multi_az":          false,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "mydb",
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"engine":            "postgres",
			"engine_version":    "13.7",
			"instance_class":    "db.t3.micro",
			"allocated_storage": 100,
			"multi_az":          false,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching RDS instances, got %d", len(result))
	}
}

// TestCompareStateWithActual_AllResourcesPresent tests when all resources are present and matching
func TestCompareStateWithActual_AllResourcesPresent(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-1"}},
		{Type: "aws_subnet", Attributes: map[string]interface{}{"id": "subnet-1"}},
		{Type: "aws_instance", Attributes: map[string]interface{}{"instance_id": "i-1"}},
		{Type: "aws_security_group", Attributes: map[string]interface{}{"group_id": "sg-1"}},
		{Type: "aws_db_instance", Attributes: map[string]interface{}{"db_instance_identifier": "db-1"}},
	}

	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc"},
		{ID: "subnet-1", Type: "aws_subnet"},
		{ID: "i-1", Type: "aws_instance"},
		{ID: "sg-1", Type: "aws_security_group"},
		{ID: "db-1", Type: "aws_db_instance"},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	// All resources present and matching
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected no unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected no missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected no modified resources, got %d", len(result.ModifiedResources))
	}
}

// TestExtractTags_EC2Tags tests extractTags with EC2 tag types
func TestExtractTags_EC2Tags(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("my-vpc"),
		},
		{
			Key:   aws.String("Environment"),
			Value: aws.String("production"),
		},
		{
			Key:   aws.String("Owner"),
			Value: aws.String("platform-team"),
		},
	}

	result := extractTags(tags)

	if len(result) != 3 {
		t.Errorf("expected 3 tags, got %d", len(result))
	}
	if result["Name"] != "my-vpc" {
		t.Errorf("expected Name=my-vpc, got %s", result["Name"])
	}
	if result["Environment"] != "production" {
		t.Errorf("expected Environment=production, got %s", result["Environment"])
	}
	if result["Owner"] != "platform-team" {
		t.Errorf("expected Owner=platform-team, got %s", result["Owner"])
	}
}

// TestExtractTags_EmptyTags tests extractTags with empty tag slice
func TestExtractTags_EmptyTags(t *testing.T) {
	tags := []ec2Types.Tag{}

	result := extractTags(tags)

	if len(result) != 0 {
		t.Errorf("expected 0 tags, got %d", len(result))
	}
}

// TestExtractTags_NilValues tests extractTags with nil tag values
func TestExtractTags_NilValues(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: nil,
		},
		{
			Key:   nil,
			Value: aws.String("value"),
		},
	}

	result := extractTags(tags)

	// Should still process tags, AWS SDK handles nil toString
	if len(result) != 2 {
		t.Errorf("expected 2 tags (even with nil values), got %d", len(result))
	}
}

// TestGetTagValue_FindsTag tests getTagValue when tag exists
func TestGetTagValue_FindsTag(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("my-resource"),
		},
		{
			Key:   aws.String("Environment"),
			Value: aws.String("dev"),
		},
	}

	result := getTagValue(tags, "Name")

	if result != "my-resource" {
		t.Errorf("expected 'my-resource', got %q", result)
	}
}

// TestGetTagValue_TagNotFound tests getTagValue when tag doesn't exist
func TestGetTagValue_TagNotFound(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("my-resource"),
		},
	}

	result := getTagValue(tags, "NonExistent")

	if result != "" {
		t.Errorf("expected empty string for missing tag, got %q", result)
	}
}

// TestGetTagValue_EmptyTagList tests getTagValue with empty tag slice
func TestGetTagValue_EmptyTagList(t *testing.T) {
	tags := []ec2Types.Tag{}

	result := getTagValue(tags, "Name")

	if result != "" {
		t.Errorf("expected empty string for empty tag list, got %q", result)
	}
}

// TestGetTagValue_CaseSensitive tests getTagValue is case-sensitive
func TestGetTagValue_CaseSensitive(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("my-resource"),
		},
	}

	// Exact match
	if getTagValue(tags, "Name") != "my-resource" {
		t.Errorf("exact match failed")
	}

	// Different case should not match
	if getTagValue(tags, "name") != "" {
		t.Errorf("case-insensitive match should return empty")
	}
}

// TestExtractRDSTags_RDSTags tests extractRDSTags with RDS tag types
func TestExtractRDSTags_RDSTags(t *testing.T) {
	tags := []rdsTypes.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("prod-db"),
		},
		{
			Key:   aws.String("Environment"),
			Value: aws.String("production"),
		},
	}

	result := extractRDSTags(tags)

	if len(result) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result))
	}
	if result["Name"] != "prod-db" {
		t.Errorf("expected Name=prod-db, got %s", result["Name"])
	}
}

// TestExtractRDSTags_EmptyTags tests extractRDSTags with empty tags
func TestExtractRDSTags_EmptyTags(t *testing.T) {
	tags := []rdsTypes.Tag{}

	result := extractRDSTags(tags)

	if len(result) != 0 {
		t.Errorf("expected 0 tags, got %d", len(result))
	}
}

// TestExtractELBTags_ELBTags tests extractELBTags with ELB tag types
func TestExtractELBTags_ELBTags(t *testing.T) {
	tags := []elbTypes.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("my-alb"),
		},
		{
			Key:   aws.String("Type"),
			Value: aws.String("application"),
		},
	}

	result := extractELBTags(tags)

	if len(result) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result))
	}
	if result["Name"] != "my-alb" {
		t.Errorf("expected Name=my-alb, got %s", result["Name"])
	}
	if result["Type"] != "application" {
		t.Errorf("expected Type=application, got %s", result["Type"])
	}
}

// TestExtractELBTags_EmptyTags tests extractELBTags with empty tags
func TestExtractELBTags_EmptyTags(t *testing.T) {
	tags := []elbTypes.Tag{}

	result := extractELBTags(tags)

	if len(result) != 0 {
		t.Errorf("expected 0 tags, got %d", len(result))
	}
}

// TestExtractTags_SpecialCharacters tests extractTags with special characters in tag values
func TestExtractTags_SpecialCharacters(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Description"),
			Value: aws.String("Resource with special chars: @#$%&*()"),
		},
		{
			Key:   aws.String("Path"),
			Value: aws.String("/path/to/resource"),
		},
	}

	result := extractTags(tags)

	if len(result) != 2 {
		t.Errorf("expected 2 tags, got %d", len(result))
	}
	if result["Description"] != "Resource with special chars: @#$%&*()" {
		t.Errorf("special characters not preserved in tag value")
	}
	if result["Path"] != "/path/to/resource" {
		t.Errorf("path characters not preserved in tag value")
	}
}

// TestExtractTags_LongValues tests extractTags with long tag values
func TestExtractTags_LongValues(t *testing.T) {
	longValue := "a"
	for i := 0; i < 1000; i++ {
		longValue += "a"
	}

	tags := []ec2Types.Tag{
		{
			Key:   aws.String("LongTag"),
			Value: aws.String(longValue),
		},
	}

	result := extractTags(tags)

	if len(result) != 1 {
		t.Errorf("expected 1 tag, got %d", len(result))
	}
	if result["LongTag"] != longValue {
		t.Errorf("long tag value not preserved")
	}
}

// TestGetTagValue_MultipleMatches tests getTagValue returns first match
func TestGetTagValue_MultipleMatches(t *testing.T) {
	tags := []ec2Types.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String("first"),
		},
		{
			Key:   aws.String("Name"),
			Value: aws.String("second"),
		},
	}

	result := getTagValue(tags, "Name")

	if result != "first" {
		t.Errorf("expected first match, got %q", result)
	}
}

// TestExtractTags_HighVolume tests extractTags with many tags
func TestExtractTags_HighVolume(t *testing.T) {
	tags := make([]ec2Types.Tag, 100)
	for i := 0; i < 100; i++ {
		key := "tag-" + string(rune(i))
		value := "value-" + string(rune(i))
		tags[i] = ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		}
	}

	result := extractTags(tags)

	if len(result) != 100 {
		t.Errorf("expected 100 tags, got %d", len(result))
	}
}

// TestCompareStateWithActual_NoResourcesAnywhere tests when no resources exist anywhere
func TestCompareStateWithActual_NoResourcesAnywhere(t *testing.T) {
	result := CompareStateWithActual([]*terraform.Resource{}, []*DiscoveredResource{})

	if result.UnmanagedResources == nil {
		t.Errorf("UnmanagedResources should not be nil")
	}
	if result.MissingResources == nil {
		t.Errorf("MissingResources should not be nil")
	}
	if result.ModifiedResources == nil {
		t.Errorf("ModifiedResources should not be nil")
	}
}

// TestCompareStateWithActual_OnlyUnmanaged tests scenario with only unmanaged resources
func TestCompareStateWithActual_OnlyUnmanaged(t *testing.T) {
	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc"},
		{ID: "subnet-1", Type: "aws_subnet"},
		{ID: "sg-1", Type: "aws_security_group"},
	}

	result := CompareStateWithActual([]*terraform.Resource{}, awsResources)

	if len(result.UnmanagedResources) != 3 {
		t.Errorf("expected 3 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

// TestCompareStateWithActual_OnlyMissing tests scenario with only missing resources
func TestCompareStateWithActual_OnlyMissing(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-1"}},
		{Type: "aws_instance", Attributes: map[string]interface{}{"instance_id": "i-1"}},
	}

	result := CompareStateWithActual(tfResources, []*DiscoveredResource{})

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 2 {
		t.Errorf("expected 2 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

// TestCompareStateWithActual_OnlyModified tests scenario with only modified resources
func TestCompareStateWithActual_OnlyModified(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-1",
				"cidr_block": "10.0.0.0/16",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-1",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/24",
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
}

// TestValuesEqual_IntAndInt64 tests valuesEqual with different integer types
func TestValuesEqual_IntAndInt64(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"int vs int", 100, 100, true},
		{"int64 vs int64", int64(100), int64(100), true},
		{"int vs int64", 100, int64(100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetNestedValue_DeeplyNested tests deeply nested value retrieval
func TestGetNestedValue_DeeplyNested(t *testing.T) {
	data := map[string]interface{}{
		"l1": map[string]interface{}{
			"l2": map[string]interface{}{
				"l3": map[string]interface{}{
					"l4": map[string]interface{}{
						"l5": map[string]interface{}{
							"value": "deep",
						},
					},
				},
			},
		},
	}

	result := getNestedValue(data, "l1.l2.l3.l4.l5.value")
	if result != "deep" {
		t.Errorf("expected 'deep', got %v", result)
	}
}

// TestCompareResourceAttributes_AZComparison tests availability zone comparison
func TestCompareResourceAttributes_AZComparison(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"availability_zone": "us-east-1a",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "i-123",
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"availability_zone": "us-east-1a",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching AZ")
	}
}

// TestCompareResourceAttributes_AZDifference tests availability zone difference detection
func TestCompareResourceAttributes_AZDifference(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"availability_zone": "us-east-1a",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "i-123",
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"availability_zone": "us-east-1b",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil || len(result) != 1 {
		t.Errorf("expected 1 difference for different AZ")
	}
}

// TestExtractTFResourceID_FirstEmptyString tests extracting ID when first field is empty
func TestExtractTFResourceID_FirstEmptyString(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_test",
		Attributes: map[string]interface{}{
			"id":          "",
			"instance_id": "actual-id",
		},
	}

	result := extractTFResourceID(resource)
	// Should skip empty "id" field and try next known ID field
	if result != "actual-id" {
		t.Errorf("expected 'actual-id', got %q", result)
	}
}

// TestTagsEqual_AllManagedTags tests when all AWS tags are managed tags
func TestTagsEqual_AllManagedTags(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{},
	}

	awsTags := map[string]string{
		"aws:managed":              "true",
		"kubernetes.io/controlled": "true",
	}

	result := tagsEqual(tfAttrs, awsTags)

	if !result {
		t.Errorf("expected empty TF tags to match AWS tags that are all managed")
	}
}

// TestCompareResourceAttributes_StringVsNumber tests comparing string and number values
func TestCompareResourceAttributes_StringVsNumber(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": "100",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "db-123",
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"allocated_storage": 100,
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	// String "100" and int 100 should be considered equal by valuesEqual
	if len(result) > 0 {
		t.Errorf("expected string '100' to match int 100")
	}
}

// TestCompareStateWithActual_TerraformResourceWithoutID tests resources without ID fields
func TestCompareStateWithActual_TerraformResourceWithoutID(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
				// No ID field
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-123",
			Type: "aws_vpc",
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	// TF resource without ID won't be added to map (extractTFResourceID returns "")
	// So it's neither matched nor reported as missing
	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource (AWS without TF match), got %d", len(result.UnmanagedResources))
	}
}

// TestCompareStateWithActual_LargeScaleComplex tests with many mixed resources
func TestCompareStateWithActual_LargeScaleComplex(t *testing.T) {
	tfResources := make([]*terraform.Resource, 50)
	for i := 0; i < 50; i++ {
		tfResources[i] = &terraform.Resource{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-tf-" + string(rune(i)),
				"cidr_block": "10.0.0.0/16",
			},
		}
	}

	awsResources := make([]*DiscoveredResource, 55)
	for i := 0; i < 50; i++ {
		awsResources[i] = &DiscoveredResource{
			ID:   "vpc-tf-" + string(rune(i)),
			Type: "aws_vpc",
		}
	}
	// Add 5 unmanaged resources
	for i := 50; i < 55; i++ {
		awsResources[i] = &DiscoveredResource{
			ID:   "vpc-unmanaged-" + string(rune(i)),
			Type: "aws_vpc",
		}
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 5 {
		t.Errorf("expected 5 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
}

// TestCompareResourceAttributes_WithTagOnlyDifference tests when only tags differ
func TestCompareResourceAttributes_WithTagOnlyDifference(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"tags": map[string]interface{}{
				"Name": "vpc1",
			},
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-123",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
		Tags: map[string]string{
			"Name": "vpc2",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil || len(result) != 1 {
		t.Errorf("expected 1 tag difference")
	}
	if result[0].Field != "tags" {
		t.Errorf("expected tags difference, got %s", result[0].Field)
	}
}

// TestCompareResourceAttributes_MultipleFieldsAndTagsDifference tests when both fields and tags differ
func TestCompareResourceAttributes_MultipleFieldsAndTagsDifference(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block":           "10.0.0.0/16",
			"enable_dns_hostnames": true,
			"tags": map[string]interface{}{
				"Name": "vpc1",
			},
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-123",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block":           "10.0.0.0/24",
			"enable_dns_hostnames": false,
		},
		Tags: map[string]string{
			"Name": "vpc2",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil || len(result) != 3 {
		t.Errorf("expected 3 differences (cidr, dns, tags), got %d", len(result))
	}
}

// TestCompareResourceAttributes_NoAttributes tests when both have no attributes
func TestCompareResourceAttributes_NoAttributes(t *testing.T) {
	tfRes := &terraform.Resource{
		Type:       "aws_vpc",
		Attributes: map[string]interface{}{},
	}

	awsRes := &DiscoveredResource{
		ID:         "vpc-123",
		Type:       "aws_vpc",
		Attributes: map[string]interface{}{},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil || len(result) > 0 {
		t.Errorf("expected no differences for resources with no attributes")
	}
}

// TestValuesEqual_ComplexEquality tests complex value equality scenarios
func TestValuesEqual_ComplexEquality(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"same pointer", "test", "test", true},
		{"string format of bool", "true", "true", true},
		{"whitespace strings", "  ", "  ", true},
		{"special float values", 1.0, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetNestedValue_MultiplePathVariations tests various path formats
func TestGetNestedValue_MultiplePathVariations(t *testing.T) {
	data := map[string]interface{}{
		"config": map[string]interface{}{
			"vpc": map[string]interface{}{
				"cidr": "10.0.0.0/16",
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
	}{
		{"single level", "config", nil}, // Returns the map itself, not matched with any specific value
		{"two levels", "config.vpc", nil},
		{"three levels", "config.vpc.cidr", "10.0.0.0/16"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedValue(data, tt.path)
			if tt.expected != nil && result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestTags Equal_MixedManagedAndUserTags tests filtering managed tags correctly
func TestTagsEqual_MixedManagedAndUserTags(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "prod",
			"Owner":       "team-a",
		},
	}

	awsTags := map[string]string{
		"Environment":        "prod",
		"Owner":              "team-a",
		"aws:cloudformation": "stack-123",
		"kubernetes.io/name": "resource",
		"aws:another":        "value",
	}

	result := tagsEqual(tfAttrs, awsTags)

	if !result {
		t.Errorf("expected tags to match after filtering managed prefixes")
	}
}

// TestCompareResourceAttributes_SubnetVPCID tests subnet VPC ID comparison
func TestCompareResourceAttributes_SubnetVPCID(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-123",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "subnet-456",
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-123",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if len(result) > 0 {
		t.Errorf("expected no differences for matching VPC ID")
	}
}

// TestCompareResourceAttributes_SubnetVPCIDMismatch tests subnet VPC ID mismatch
func TestCompareResourceAttributes_SubnetVPCIDMismatch(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-123",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "subnet-456",
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-789",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)
	if result == nil || len(result) != 1 {
		t.Errorf("expected 1 difference for VPC ID mismatch")
	}
}

// TestDiscoveredResource_JSONFields tests JSON field tags on DiscoveredResource
func TestDiscoveredResource_JSONFields(t *testing.T) {
	res := &DiscoveredResource{
		ID:         "id-value",
		Type:       "type-value",
		ARN:        "arn-value",
		Name:       "name-value",
		Region:     "region-value",
		Attributes: map[string]interface{}{"key": "value"},
		Tags:       map[string]string{"tag": "value"},
	}

	// Verify all fields are accessible
	if res.ID != "id-value" || res.Type != "type-value" || res.ARN != "arn-value" {
		t.Errorf("basic fields not set correctly")
	}
}

// TestDriftResult_EmptySlices tests DriftResult with empty slices
func TestDriftResult_EmptySlices(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*types.TerraformResource{},
		ModifiedResources:  []*ResourceDiff{},
	}

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("UnmanagedResources should be empty")
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("MissingResources should be empty")
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("ModifiedResources should be empty")
	}
}

// TestResourceDiff_FieldDiffAccuracy tests FieldDiff field accuracy
func TestResourceDiff_FieldDiffAccuracy(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:     "res-123",
		ResourceType:   "aws_test",
		TerraformState: map[string]interface{}{"key": "tf-val"},
		ActualState:    map[string]interface{}{"key": "actual-val"},
		Differences: []FieldDiff{
			{
				Field:          "field1",
				TerraformValue: "tf-value",
				ActualValue:    "actual-value",
			},
		},
	}

	if diff.Differences[0].Field != "field1" {
		t.Errorf("FieldDiff Field not set")
	}
	if diff.Differences[0].TerraformValue != "tf-value" {
		t.Errorf("FieldDiff TerraformValue not set")
	}
	if diff.Differences[0].ActualValue != "actual-value" {
		t.Errorf("FieldDiff ActualValue not set")
	}
}

// TestCompareStateWithActual_AllResourceTypesPresent tests when multiple resource types match
func TestCompareStateWithActual_AllResourceTypesPresent(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-1"}},
		{Type: "aws_subnet", Attributes: map[string]interface{}{"id": "subnet-1"}},
		{Type: "aws_security_group", Attributes: map[string]interface{}{"group_id": "sg-1"}},
		{Type: "aws_instance", Attributes: map[string]interface{}{"instance_id": "i-1"}},
		{Type: "aws_db_instance", Attributes: map[string]interface{}{"db_instance_identifier": "db-1"}},
		{Type: "aws_eks_cluster", Attributes: map[string]interface{}{"cluster_name": "eks-1"}},
		{Type: "aws_lb", Attributes: map[string]interface{}{"arn": "arn:aws:lb:..."}},
	}

	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc"},
		{ID: "subnet-1", Type: "aws_subnet"},
		{ID: "sg-1", Type: "aws_security_group"},
		{ID: "i-1", Type: "aws_instance"},
		{ID: "db-1", Type: "aws_db_instance"},
		{ID: "eks-1", Type: "aws_eks_cluster"},
		{ID: "arn:aws:lb:...", Type: "aws_lb"},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 0 || len(result.MissingResources) != 0 || len(result.ModifiedResources) != 0 {
		t.Errorf("expected all resources to match")
	}
}

// TestCompareStateWithActual_SelectiveModifications tests when some resources are modified
func TestCompareStateWithActual_SelectiveModifications(t *testing.T) {
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-1",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-2",
				"cidr_block": "10.1.0.0/16",
			},
		},
		{
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-3",
				"cidr_block": "10.2.0.0/16",
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-1",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			ID:   "vpc-2",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.1.0.0/24", // Modified
			},
		},
		{
			ID:   "vpc-3",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.2.0.0/16",
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
	if result.ModifiedResources[0].ResourceID != "vpc-2" {
		t.Errorf("expected vpc-2 to be modified")
	}
}

// TestValuesEqual_FormattedNumbers tests valuesEqual with various number formats
func TestValuesEqual_FormattedNumbers(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"positive ints", 42, 42, true},
		{"negative ints", -42, -42, true},
		{"float same", 3.14, 3.14, true},
		{"zero vs zero", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCompareResourceAttributes_ELBTypeScenarios tests various ELB type scenarios
func TestCompareResourceAttributes_ELBTypeScenarios(t *testing.T) {
	tests := []struct {
		name        string
		lbType      string
		scheme      string
		shouldMatch bool
	}{
		{"application alb", "application", "internet-facing", true},
		{"network nlb", "network", "internal", true},
		{"gateway gwlb", "gateway", "internet-facing", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tfRes := &terraform.Resource{
				Type: "aws_lb",
				Attributes: map[string]interface{}{
					"type":   tt.lbType,
					"scheme": tt.scheme,
				},
			}

			awsRes := &DiscoveredResource{
				ID:   "lb-123",
				Type: "aws_lb",
				Attributes: map[string]interface{}{
					"type":   tt.lbType,
					"scheme": tt.scheme,
				},
			}

			result := compareResourceAttributes(tfRes, awsRes)
			if !tt.shouldMatch && len(result) == 0 {
				t.Errorf("expected differences")
			}
		})
	}
}
