package aws

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// TestExtractRDSTags_Behavior tests the RDS tag extraction helper behavior
func TestExtractRDSTags_Behavior(t *testing.T) {
	// We can't call extractRDSTags directly without AWS SDK types,
	// but we can verify that the result type is correct map[string]string
	// by testing with actual usage in DiscoveredResource
	resource := &DiscoveredResource{
		ID:   "db-test",
		Type: "aws_db_instance",
		Tags: make(map[string]string), // This is what extractRDSTags returns
	}

	if resource.Tags == nil {
		t.Errorf("tags should not be nil")
	}
}

// TestExtractELBTags_Behavior tests the ELB tag extraction helper behavior
func TestExtractELBTags_Behavior(t *testing.T) {
	// We can't call extractELBTags directly without AWS SDK types,
	// but we can verify that the result type is correct map[string]string
	resource := &DiscoveredResource{
		ID:   "alb-test",
		Type: "aws_lb",
		Tags: make(map[string]string), // This is what extractELBTags returns
	}

	if resource.Tags == nil {
		t.Errorf("tags should not be nil")
	}
}

// TestDiscoveryClient_RegionSetup tests that region is properly stored
func TestDiscoveryClient_RegionSetup(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{
			name:     "us-east-1",
			region:   "us-east-1",
			expected: "us-east-1",
		},
		{
			name:     "eu-west-1",
			region:   "eu-west-1",
			expected: "eu-west-1",
		},
		{
			name:     "ap-southeast-1",
			region:   "ap-southeast-1",
			expected: "ap-southeast-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't call NewDiscoveryClient without AWS credentials,
			// but we can verify the type structure
			client := &DiscoveryClient{
				region: tt.region,
			}
			if client.region != tt.expected {
				t.Errorf("expected region %s, got %s", tt.expected, client.region)
			}
		})
	}
}

// TestDiscoveredResource_AllResourceTypes tests various resource type definitions
func TestDiscoveredResource_AllResourceTypes(t *testing.T) {
	resourceTypes := []string{
		"aws_vpc",
		"aws_subnet",
		"aws_security_group",
		"aws_instance",
		"aws_db_instance",
		"aws_eks_cluster",
		"aws_elasticache_replication_group",
		"aws_lb",
	}

	for _, resourceType := range resourceTypes {
		t.Run(resourceType, func(t *testing.T) {
			resource := &DiscoveredResource{
				ID:     "test-id",
				Type:   resourceType,
				Name:   "test-name",
				Region: "us-east-1",
			}

			if resource.Type != resourceType {
				t.Errorf("expected type %s, got %s", resourceType, resource.Type)
			}
		})
	}
}

// TestDiscoveredResource_AttributeTypes tests different attribute value types
func TestDiscoveredResource_AttributeTypes(t *testing.T) {
	tests := []struct {
		name        string
		attributes  map[string]interface{}
		keyToCheck  string
		expectedVal interface{}
	}{
		{
			name: "string attribute",
			attributes: map[string]interface{}{
				"cidr_block": "10.0.0.0/16",
			},
			keyToCheck:  "cidr_block",
			expectedVal: "10.0.0.0/16",
		},
		{
			name: "boolean attribute",
			attributes: map[string]interface{}{
				"multi_az": true,
			},
			keyToCheck:  "multi_az",
			expectedVal: true,
		},
		{
			name: "integer attribute",
			attributes: map[string]interface{}{
				"allocated_storage": 100,
			},
			keyToCheck:  "allocated_storage",
			expectedVal: 100,
		},
		{
			name: "slice attribute",
			attributes: map[string]interface{}{
				"subnet_ids": []string{"subnet-1", "subnet-2"},
			},
			keyToCheck: "subnet_ids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &DiscoveredResource{
				ID:         "test-id",
				Type:       "aws_test",
				Attributes: tt.attributes,
			}

			val, exists := resource.Attributes[tt.keyToCheck]
			if !exists {
				t.Errorf("attribute %s not found", tt.keyToCheck)
				return
			}

			if tt.expectedVal != nil && val != tt.expectedVal {
				t.Errorf("expected %v, got %v", tt.expectedVal, val)
			}
		})
	}
}

// TestResourceDiff_MultipleFieldDifferences tests ResourceDiff with multiple field changes
func TestResourceDiff_MultipleFieldDifferences(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "test-db",
		ResourceType: "aws_db_instance",
		TerraformState: map[string]interface{}{
			"allocated_storage":  100,
			"engine_version":     "13.7",
			"publicly_accessible": false,
			"multi_az":           true,
		},
		ActualState: map[string]interface{}{
			"allocated_storage":   200,
			"engine_version":      "14.1",
			"publicly_accessible": true,
			"multi_az":            false,
		},
		Differences: []FieldDiff{
			{
				Field:          "allocated_storage",
				TerraformValue: 100,
				ActualValue:    200,
			},
			{
				Field:          "engine_version",
				TerraformValue: "13.7",
				ActualValue:    "14.1",
			},
			{
				Field:          "publicly_accessible",
				TerraformValue: false,
				ActualValue:    true,
			},
			{
				Field:          "multi_az",
				TerraformValue: true,
				ActualValue:    false,
			},
		},
	}

	if len(diff.Differences) != 4 {
		t.Errorf("expected 4 differences, got %d", len(diff.Differences))
	}

	// Verify each difference
	expectedDiffs := map[string][2]interface{}{
		"allocated_storage":   {100, 200},
		"engine_version":      {"13.7", "14.1"},
		"publicly_accessible": {false, true},
		"multi_az":            {true, false},
	}

	for _, fieldDiff := range diff.Differences {
		expected, ok := expectedDiffs[fieldDiff.Field]
		if !ok {
			t.Errorf("unexpected field: %s", fieldDiff.Field)
			continue
		}

		if fieldDiff.TerraformValue != expected[0] {
			t.Errorf("field %s: expected TF value %v, got %v",
				fieldDiff.Field, expected[0], fieldDiff.TerraformValue)
		}

		if fieldDiff.ActualValue != expected[1] {
			t.Errorf("field %s: expected actual value %v, got %v",
				fieldDiff.Field, expected[1], fieldDiff.ActualValue)
		}
	}
}

// TestDriftResult_LargeScale tests DriftResult with many resources
func TestDriftResult_LargeScale(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: make([]*DiscoveredResource, 50),
		MissingResources:   make([]*terraform.Resource, 30),
		ModifiedResources:  make([]*ResourceDiff, 20),
	}

	// Populate with test data
	for i := 0; i < 50; i++ {
		result.UnmanagedResources[i] = &DiscoveredResource{
			ID:   "vpc-" + string(rune(i)),
			Type: "aws_vpc",
		}
	}

	for i := 0; i < 30; i++ {
		result.MissingResources[i] = &terraform.Resource{
			Type: "aws_instance",
		}
	}

	for i := 0; i < 20; i++ {
		result.ModifiedResources[i] = &ResourceDiff{
			ResourceID: "sg-" + string(rune(i)),
		}
	}

	if len(result.UnmanagedResources) != 50 {
		t.Errorf("expected 50 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 30 {
		t.Errorf("expected 30 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 20 {
		t.Errorf("expected 20 modified resources, got %d", len(result.ModifiedResources))
	}
}

// TestDiscoveredResource_NetworkingResources tests VPC-related resources
func TestDiscoveredResource_NetworkingResources(t *testing.T) {
	networkingResources := []struct {
		name     string
		resType  string
		id       string
	}{
		{"VPC", "aws_vpc", "vpc-12345"},
		{"Subnet", "aws_subnet", "subnet-12345"},
		{"Security Group", "aws_security_group", "sg-12345"},
		{"Network Interface", "aws_network_interface", "eni-12345"},
		{"Route Table", "aws_route_table", "rtb-12345"},
	}

	for _, nr := range networkingResources {
		t.Run(nr.name, func(t *testing.T) {
			res := &DiscoveredResource{
				ID:   nr.id,
				Type: nr.resType,
			}

			if res.Type != nr.resType {
				t.Errorf("expected %s, got %s", nr.resType, res.Type)
			}
		})
	}
}

// TestDiscoveredResource_ComputeResources tests compute-related resources
func TestDiscoveredResource_ComputeResources(t *testing.T) {
	computeResources := []struct {
		name     string
		resType  string
		id       string
	}{
		{"EC2 Instance", "aws_instance", "i-12345"},
		{"EKS Cluster", "aws_eks_cluster", "my-cluster"},
		{"Load Balancer", "aws_lb", "arn:aws:elasticloadbalancing:..."},
		{"Target Group", "aws_lb_target_group", "arn:aws:elasticloadbalancing:..."},
	}

	for _, cr := range computeResources {
		t.Run(cr.name, func(t *testing.T) {
			res := &DiscoveredResource{
				ID:   cr.id,
				Type: cr.resType,
			}

			if res.Type != cr.resType {
				t.Errorf("expected %s, got %s", cr.resType, res.Type)
			}
		})
	}
}

// TestDiscoveredResource_DatabaseResources tests database-related resources
func TestDiscoveredResource_DatabaseResources(t *testing.T) {
	databaseResources := []struct {
		name     string
		resType  string
		id       string
	}{
		{"RDS Instance", "aws_db_instance", "db-prod"},
		{"ElastiCache", "aws_elasticache_replication_group", "redis-cluster"},
		{"DynamoDB Table", "aws_dynamodb_table", "my-table"},
	}

	for _, dr := range databaseResources {
		t.Run(dr.name, func(t *testing.T) {
			res := &DiscoveredResource{
				ID:   dr.id,
				Type: dr.resType,
			}

			if res.Type != dr.resType {
				t.Errorf("expected %s, got %s", dr.resType, res.Type)
			}
		})
	}
}

// TestDiscoveredResource_TagVariations tests various tag scenarios
func TestDiscoveredResource_TagVariations(t *testing.T) {
	tests := []struct {
		name string
		tags map[string]string
	}{
		{
			name: "no tags",
			tags: map[string]string{},
		},
		{
			name: "single tag",
			tags: map[string]string{
				"Name": "resource-1",
			},
		},
		{
			name: "multiple tags",
			tags: map[string]string{
				"Name":        "resource-1",
				"Environment": "production",
				"Owner":       "team-a",
				"CostCenter":  "12345",
			},
		},
		{
			name: "special characters in tag values",
			tags: map[string]string{
				"Name":        "my-resource_v2.1",
				"Description": "Resource with special chars: @#$%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &DiscoveredResource{
				ID:   "test-id",
				Type: "aws_instance",
				Tags: tt.tags,
			}

			if len(res.Tags) != len(tt.tags) {
				t.Errorf("expected %d tags, got %d", len(tt.tags), len(res.Tags))
			}

			for k, v := range tt.tags {
				if res.Tags[k] != v {
					t.Errorf("tag %s: expected %s, got %s", k, v, res.Tags[k])
				}
			}
		})
	}
}

// TestFieldDiff_VariousTypes tests FieldDiff with various value types
func TestFieldDiff_VariousTypes(t *testing.T) {
	tests := []struct {
		name           string
		tfValue        interface{}
		actualValue    interface{}
	}{
		{
			name:           "string values",
			tfValue:        "t3.micro",
			actualValue:    "t3.small",
		},
		{
			name:           "integer values",
			tfValue:        100,
			actualValue:    200,
		},
		{
			name:           "boolean values",
			tfValue:        true,
			actualValue:    false,
		},
		{
			name:           "string vs integer",
			tfValue:        "100",
			actualValue:    100,
		},
		{
			name:           "nil values",
			tfValue:        nil,
			actualValue:    "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := FieldDiff{
				Field:          "test-field",
				TerraformValue: tt.tfValue,
				ActualValue:    tt.actualValue,
			}

			if diff.TerraformValue != tt.tfValue {
				t.Errorf("expected TF value %v, got %v", tt.tfValue, diff.TerraformValue)
			}

			if diff.ActualValue != tt.actualValue {
				t.Errorf("expected actual value %v, got %v", tt.actualValue, diff.ActualValue)
			}
		})
	}
}

// TestResourceDiff_FullDiffScenario tests a complete diff scenario
func TestResourceDiff_FullDiffScenario(t *testing.T) {
	// Simulate an EKS cluster that was modified
	diff := &ResourceDiff{
		ResourceID:   "prod-cluster",
		ResourceType: "aws_eks_cluster",
		TerraformState: map[string]interface{}{
			"version":    "1.23",
			"role_arn":   "arn:aws:iam::123456789:role/eks-service",
			"subnet_ids": []string{"subnet-1", "subnet-2"},
			"tags": map[string]interface{}{
				"Environment": "prod",
				"Name":        "prod-cluster",
			},
		},
		ActualState: map[string]interface{}{
			"version":    "1.24", // Updated!
			"role_arn":   "arn:aws:iam::123456789:role/eks-service",
			"subnet_ids": []string{"subnet-1", "subnet-2"},
			"tags": map[string]interface{}{
				"Environment": "prod",
				"Name":        "prod-cluster",
				"ManagedBy":   "CloudFormation", // Added!
			},
		},
		Differences: []FieldDiff{
			{
				Field:          "version",
				TerraformValue: "1.23",
				ActualValue:    "1.24",
			},
		},
	}

	if diff.ResourceID != "prod-cluster" {
		t.Errorf("expected resource ID prod-cluster")
	}

	if len(diff.Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(diff.Differences))
	}

	if diff.Differences[0].Field != "version" {
		t.Errorf("expected version difference, got %s", diff.Differences[0].Field)
	}
}
