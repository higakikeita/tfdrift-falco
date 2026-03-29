package aws

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

func TestCompareStateWithActual_UnmanagedResources(t *testing.T) {
	// Terraform has no resources, AWS has some
	tfResources := []*terraform.Resource{}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-12345",
			Type: "aws_vpc",
			Name: "prod-vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
				"state":      "available",
			},
			Tags: map[string]string{"Environment": "prod"},
		},
		{
			ID:   "sg-12345",
			Type: "aws_security_group",
			Name: "prod-sg",
			Attributes: map[string]interface{}{
				"vpc_id":      "vpc-12345",
				"description": "Production security group",
				"name":        "prod-sg",
			},
			Tags: map[string]string{},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

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
	// Terraform has resources, AWS has none
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":        "vpc-99999",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"instance_id":   "i-12345",
				"instance_type": "t3.micro",
			},
		},
	}

	awsResources := []*DiscoveredResource{}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.MissingResources) != 2 {
		t.Errorf("expected 2 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

func TestCompareStateWithActual_ModifiedResources(t *testing.T) {
	// Resources exist in both but with different attributes
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":                    "vpc-12345",
				"cidr_block":            "10.0.0.0/16",
				"enable_dns_hostnames":  true,
				"enable_dns_support":    true,
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-12345",
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"cidr_block":           "10.0.0.0/16",
				"enable_dns_hostnames": false, // Different!
				"enable_dns_support":   true,
			},
			Tags: map[string]string{},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources, got %d", len(result.MissingResources))
	}

	if len(result.ModifiedResources[0].Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(result.ModifiedResources[0].Differences))
	}
	if result.ModifiedResources[0].Differences[0].Field != "enable_dns_hostnames" {
		t.Errorf("expected field 'enable_dns_hostnames', got '%s'", result.ModifiedResources[0].Differences[0].Field)
	}
}

func TestCompareStateWithActual_Mixed(t *testing.T) {
	// Mix of all three scenarios
	tfResources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "managed-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-managed",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_instance",
			Name: "deleted-instance",
			Attributes: map[string]interface{}{
				"instance_id": "i-deleted",
			},
		},
		{
			Type: "aws_db_instance",
			Name: "modified-db",
			Attributes: map[string]interface{}{
				"db_instance_identifier": "db-modified",
				"allocated_storage":       100,
			},
		},
	}

	awsResources := []*DiscoveredResource{
		{
			ID:   "vpc-managed",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			ID:   "vpc-unmanaged",
			Type: "aws_vpc",
			Attributes: map[string]interface{}{
				"cidr_block": "10.1.0.0/16",
			},
		},
		{
			ID:   "db-modified",
			Type: "aws_db_instance",
			Attributes: map[string]interface{}{
				"allocated_storage": 200, // Different!
			},
		},
	}

	result := CompareStateWithActual(tfResources, awsResources)

	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}
}

func TestExtractTFResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource *terraform.Resource
		expected string
	}{
		{
			name: "vpc with id field",
			resource: &terraform.Resource{
				Type: "aws_vpc",
				Attributes: map[string]interface{}{
					"id":         "vpc-12345",
					"cidr_block": "10.0.0.0/16",
				},
			},
			expected: "vpc-12345",
		},
		{
			name: "instance with instance_id field",
			resource: &terraform.Resource{
				Type: "aws_instance",
				Attributes: map[string]interface{}{
					"instance_id": "i-12345",
				},
			},
			expected: "i-12345",
		},
		{
			name: "db with db_instance_identifier field",
			resource: &terraform.Resource{
				Type: "aws_db_instance",
				Attributes: map[string]interface{}{
					"db_instance_identifier": "my-database",
				},
			},
			expected: "my-database",
		},
		{
			name: "no id found",
			resource: &terraform.Resource{
				Type: "aws_vpc",
				Attributes: map[string]interface{}{
					"cidr_block": "10.0.0.0/16",
				},
			},
			expected: "",
		},
		{
			name: "arn field as fallback",
			resource: &terraform.Resource{
				Type: "aws_eks_cluster",
				Attributes: map[string]interface{}{
					"arn": "arn:aws:eks:us-east-1:123456789:cluster/my-cluster",
				},
			},
			expected: "arn:aws:eks:us-east-1:123456789:cluster/my-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTFResourceID(tt.resource)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        "value",
			b:        nil,
			expected: false,
		},
		{
			name:     "identical strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "identical booleans",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "different booleans",
			a:        true,
			b:        false,
			expected: false,
		},
		{
			name:     "bool vs string true",
			a:        true,
			b:        "true",
			expected: true,
		},
		{
			name:     "bool vs string false",
			a:        false,
			b:        "false",
			expected: true,
		},
		{
			name:     "bool vs different string",
			a:        true,
			b:        "yes",
			expected: false,
		},
		{
			name:     "identical numbers",
			a:        100,
			b:        100,
			expected: true,
		},
		{
			name:     "different numbers",
			a:        100,
			b:        200,
			expected: false,
		},
		{
			name:     "string representation of numbers",
			a:        100,
			b:        "100",
			expected: true,
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

func TestGetNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"vpc_config": map[string]interface{}{
			"subnet_ids": []string{"subnet-1", "subnet-2"},
			"security_group_ids": map[string]interface{}{
				"primary": "sg-123",
			},
		},
	}

	tests := []struct {
		name         string
		path         string
		expectedType string
		checkValue   func(interface{}) bool
	}{
		{
			name:         "top-level field",
			path:         "name",
			expectedType: "string",
			checkValue: func(v interface{}) bool {
				return v == "test"
			},
		},
		{
			name:         "nested field - level 1",
			path:         "vpc_config.subnet_ids",
			expectedType: "slice",
			checkValue: func(v interface{}) bool {
				slice, ok := v.([]string)
				return ok && len(slice) == 2 && slice[0] == "subnet-1"
			},
		},
		{
			name:         "nested field - level 2",
			path:         "vpc_config.security_group_ids.primary",
			expectedType: "string",
			checkValue: func(v interface{}) bool {
				return v == "sg-123"
			},
		},
		{
			name:         "non-existent field",
			path:         "nonexistent",
			expectedType: "nil",
			checkValue: func(v interface{}) bool {
				return v == nil
			},
		},
		{
			name:         "non-existent nested field",
			path:         "vpc_config.nonexistent",
			expectedType: "nil",
			checkValue: func(v interface{}) bool {
				return v == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedValue(data, tt.path)
			if !tt.checkValue(result) {
				t.Errorf("expected type %s, got %v", tt.expectedType, result)
			}
		})
	}
}

func TestGetTerraformTags(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "tags field",
			attrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Environment": "prod",
					"Owner":       "team-a",
				},
			},
			expected: map[string]string{
				"Environment": "prod",
				"Owner":       "team-a",
			},
		},
		{
			name: "tags_all field",
			attrs: map[string]interface{}{
				"tags_all": map[string]interface{}{
					"Environment": "dev",
					"Project":     "myapp",
				},
			},
			expected: map[string]string{
				"Environment": "dev",
				"Project":     "myapp",
			},
		},
		{
			name: "no tags",
			attrs: map[string]interface{}{
				"name": "test",
			},
			expected: map[string]string{},
		},
		{
			name: "tags with mixed types",
			attrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Valid": "value",
					"Invalid": 123, // Should be skipped
				},
			},
			expected: map[string]string{
				"Valid": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTerraformTags(tt.attrs)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tags, got %d", len(tt.expected), len(result))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected tag %s=%s, got %s", k, v, result[k])
				}
			}
		})
	}
}

func TestTagsEqual(t *testing.T) {
	tests := []struct {
		name     string
		tfAttrs  map[string]interface{}
		awsTags  map[string]string
		expected bool
	}{
		{
			name: "identical tags",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Environment": "prod",
				},
			},
			awsTags: map[string]string{
				"Environment": "prod",
			},
			expected: true,
		},
		{
			name: "different tags",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Environment": "prod",
				},
			},
			awsTags: map[string]string{
				"Environment": "dev",
			},
			expected: false,
		},
		{
			name: "aws-managed tags ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Environment": "prod",
				},
			},
			awsTags: map[string]string{
				"Environment": "prod",
				"aws:managed": "true",
			},
			expected: true,
		},
		{
			name: "kubernetes.io tags ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"Owner": "team",
				},
			},
			awsTags: map[string]string{
				"Owner":                    "team",
				"kubernetes.io/created-by": "controller",
			},
			expected: true,
		},
		{
			name: "empty tags",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{},
			},
			awsTags: map[string]string{},
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

func TestGetComparableFields(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		shouldHave   []string
	}{
		{
			name:         "aws_vpc",
			resourceType: "aws_vpc",
			shouldHave:   []string{"cidr_block", "enable_dns_hostnames", "enable_dns_support"},
		},
		{
			name:         "aws_subnet",
			resourceType: "aws_subnet",
			shouldHave:   []string{"vpc_id", "cidr_block", "availability_zone", "map_public_ip_on_launch"},
		},
		{
			name:         "aws_security_group",
			resourceType: "aws_security_group",
			shouldHave:   []string{"vpc_id", "description", "name"},
		},
		{
			name:         "aws_instance",
			resourceType: "aws_instance",
			shouldHave:   []string{"instance_type", "subnet_id", "vpc_id", "availability_zone"},
		},
		{
			name:         "aws_db_instance",
			resourceType: "aws_db_instance",
			shouldHave: []string{"engine", "engine_version", "instance_class",
				"allocated_storage", "db_subnet_group_name", "multi_az", "publicly_accessible"},
		},
		{
			name:         "aws_eks_cluster",
			resourceType: "aws_eks_cluster",
			shouldHave:   []string{"version", "role_arn"},
		},
		{
			name:         "aws_elasticache_replication_group",
			resourceType: "aws_elasticache_replication_group",
			shouldHave:   []string{"node_type", "automatic_failover_enabled", "multi_az_enabled"},
		},
		{
			name:         "aws_lb",
			resourceType: "aws_lb",
			shouldHave:   []string{"type", "scheme", "vpc_id"},
		},
		{
			name:         "unknown resource type",
			resourceType: "unknown_resource",
			shouldHave:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getComparableFields(tt.resourceType)
			if len(result) != len(tt.shouldHave) {
				t.Errorf("expected %d fields, got %d", len(tt.shouldHave), len(result))
				return
			}
			for _, field := range tt.shouldHave {
				found := false
				for _, f := range result {
					if f == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected field %s not found", field)
				}
			}
		})
	}
}

func TestCompareResourceAttributes_TypeMismatch(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id": "vpc-12345",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-12345",
		Type: "aws_subnet", // Different type
	}

	result := compareResourceAttributes(tfRes, awsRes)

	if result != nil {
		t.Errorf("expected nil for type mismatch, got %v", result)
	}
}

func TestCompareResourceAttributes_WithTags(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id":         "vpc-12345",
			"cidr_block": "10.0.0.0/16",
			"tags": map[string]interface{}{
				"Environment": "prod",
			},
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-12345",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
		Tags: map[string]string{
			"Environment": "dev",
		},
	}

	result := compareResourceAttributes(tfRes, awsRes)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have one difference for tags
	if len(result.Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(result.Differences))
	}

	if result.Differences[0].Field != "tags" {
		t.Errorf("expected 'tags' field difference, got %s", result.Differences[0].Field)
	}
}
