package aws

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

func TestDiscoveredResource_Structure(t *testing.T) {
	// Test that DiscoveredResource properly marshals all fields
	resource := &DiscoveredResource{
		ID:   "vpc-12345",
		Type: "aws_vpc",
		ARN:  "arn:aws:ec2:us-east-1:123456789:vpc/vpc-12345",
		Name: "production-vpc",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"state":      "available",
		},
		Tags: map[string]string{
			"Environment": "production",
			"Owner":       "platform-team",
		},
	}

	if resource.ID != "vpc-12345" {
		t.Errorf("expected ID vpc-12345, got %s", resource.ID)
	}
	if resource.Type != "aws_vpc" {
		t.Errorf("expected Type aws_vpc, got %s", resource.Type)
	}
	if resource.ARN != "arn:aws:ec2:us-east-1:123456789:vpc/vpc-12345" {
		t.Errorf("unexpected ARN")
	}
	if resource.Region != "us-east-1" {
		t.Errorf("expected Region us-east-1, got %s", resource.Region)
	}
	if len(resource.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(resource.Tags))
	}
}

func TestDriftResult_Structure(t *testing.T) {
	// Test that DriftResult properly holds all result types
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{
			{
				ID:   "vpc-unmanaged",
				Type: "aws_vpc",
			},
		},
		MissingResources: []*terraform.Resource{
			{
				Type: "aws_subnet",
				Name: "missing-subnet",
			},
		},
		ModifiedResources: []*ResourceDiff{
			{
				ResourceID:   "i-12345",
				ResourceType: "aws_instance",
			},
		},
	}

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

func TestResourceDiff_Structure(t *testing.T) {
	// Test ResourceDiff structure with field differences
	diff := &ResourceDiff{
		ResourceID:     "db-prod",
		ResourceType:   "aws_db_instance",
		TerraformState: map[string]interface{}{
			"allocated_storage": 100,
			"engine":            "postgres",
		},
		ActualState: map[string]interface{}{
			"allocated_storage": 200,
			"engine":            "postgres",
		},
		Differences: []FieldDiff{
			{
				Field:          "allocated_storage",
				TerraformValue: 100,
				ActualValue:    200,
			},
		},
	}

	if diff.ResourceID != "db-prod" {
		t.Errorf("expected ResourceID db-prod, got %s", diff.ResourceID)
	}
	if diff.ResourceType != "aws_db_instance" {
		t.Errorf("expected ResourceType aws_db_instance, got %s", diff.ResourceType)
	}
	if len(diff.Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(diff.Differences))
	}

	fieldDiff := diff.Differences[0]
	if fieldDiff.Field != "allocated_storage" {
		t.Errorf("expected field allocated_storage, got %s", fieldDiff.Field)
	}
	if fieldDiff.TerraformValue != 100 {
		t.Errorf("expected TF value 100, got %v", fieldDiff.TerraformValue)
	}
	if fieldDiff.ActualValue != 200 {
		t.Errorf("expected actual value 200, got %v", fieldDiff.ActualValue)
	}
}

func TestFieldDiff_Structure(t *testing.T) {
	// Test individual FieldDiff structure
	diff := FieldDiff{
		Field:          "enable_dns_hostnames",
		TerraformValue: true,
		ActualValue:    false,
	}

	if diff.Field != "enable_dns_hostnames" {
		t.Errorf("expected field enable_dns_hostnames, got %s", diff.Field)
	}
	if diff.TerraformValue != true {
		t.Errorf("expected TF value true")
	}
	if diff.ActualValue != false {
		t.Errorf("expected actual value false")
	}
}

func TestDiscoveredResource_VPC(t *testing.T) {
	// Test typical VPC discovery result
	vpc := &DiscoveredResource{
		ID:     "vpc-12345",
		Type:   "aws_vpc",
		Name:   "prod-vpc",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"state":      "available",
		},
		Tags: map[string]string{
			"Name":        "prod-vpc",
			"Environment": "production",
		},
	}

	if vpc.Type != "aws_vpc" {
		t.Errorf("expected type aws_vpc")
	}
	if vpc.Attributes["cidr_block"] != "10.0.0.0/16" {
		t.Errorf("expected cidr_block 10.0.0.0/16")
	}
}

func TestDiscoveredResource_EC2Instance(t *testing.T) {
	// Test typical EC2 instance discovery result
	instance := &DiscoveredResource{
		ID:     "i-12345",
		Type:   "aws_instance",
		Name:   "web-server",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"instance_type":     "t3.micro",
			"subnet_id":         "subnet-12345",
			"vpc_id":            "vpc-12345",
			"availability_zone": "us-east-1a",
			"private_ip":        "10.0.1.10",
			"public_ip":         "203.0.113.1",
			"state":             "running",
		},
		Tags: map[string]string{
			"Name": "web-server",
			"Role": "web",
		},
	}

	if instance.Type != "aws_instance" {
		t.Errorf("expected type aws_instance")
	}
	if instance.Attributes["instance_type"] != "t3.micro" {
		t.Errorf("expected instance_type t3.micro")
	}
	if instance.Attributes["state"] != "running" {
		t.Errorf("expected state running")
	}
}

func TestDiscoveredResource_RDSInstance(t *testing.T) {
	// Test typical RDS instance discovery result
	rds := &DiscoveredResource{
		ID:     "db-prod",
		Type:   "aws_db_instance",
		ARN:    "arn:aws:rds:us-east-1:123456789:db:db-prod",
		Name:   "db-prod",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"engine":                 "postgres",
			"engine_version":         "13.7",
			"instance_class":         "db.t3.micro",
			"allocated_storage":      100,
			"db_subnet_group_name":   "default",
			"vpc_security_group_ids": []string{"sg-12345"},
			"multi_az":               true,
			"publicly_accessible":    false,
			"status":                 "available",
		},
		Tags: map[string]string{
			"Environment": "production",
		},
	}

	if rds.Type != "aws_db_instance" {
		t.Errorf("expected type aws_db_instance")
	}
	if rds.Attributes["engine"] != "postgres" {
		t.Errorf("expected engine postgres")
	}
	if rds.Attributes["multi_az"] != true {
		t.Errorf("expected multi_az true")
	}
}

func TestDiscoveredResource_EKSCluster(t *testing.T) {
	// Test typical EKS cluster discovery result
	eks := &DiscoveredResource{
		ID:     "my-cluster",
		Type:   "aws_eks_cluster",
		ARN:    "arn:aws:eks:us-east-1:123456789:cluster/my-cluster",
		Name:   "my-cluster",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"version":            "1.24",
			"role_arn":           "arn:aws:iam::123456789:role/eks-service-role",
			"status":             "ACTIVE",
			"subnet_ids":         []string{"subnet-12345", "subnet-67890"},
			"security_group_ids": []string{"sg-12345"},
			"endpoint":           "https://ABC123.eks.us-east-1.amazonaws.com",
		},
		Tags: map[string]string{
			"Environment": "production",
			"Name":        "my-cluster",
		},
	}

	if eks.Type != "aws_eks_cluster" {
		t.Errorf("expected type aws_eks_cluster")
	}
	if eks.Attributes["version"] != "1.24" {
		t.Errorf("expected version 1.24")
	}
}

func TestDiscoveredResource_SecurityGroup(t *testing.T) {
	// Test typical security group discovery result
	sg := &DiscoveredResource{
		ID:     "sg-12345",
		Type:   "aws_security_group",
		Name:   "prod-sg",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"vpc_id":      "vpc-12345",
			"description": "Production security group",
			"name":        "prod-sg",
		},
		Tags: map[string]string{
			"Name": "prod-sg",
		},
	}

	if sg.Type != "aws_security_group" {
		t.Errorf("expected type aws_security_group")
	}
	if sg.Attributes["vpc_id"] != "vpc-12345" {
		t.Errorf("expected vpc_id vpc-12345")
	}
}

func TestDiscoveredResource_Subnet(t *testing.T) {
	// Test typical subnet discovery result
	subnet := &DiscoveredResource{
		ID:     "subnet-12345",
		Type:   "aws_subnet",
		Name:   "prod-subnet-1a",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"vpc_id":                          "vpc-12345",
			"cidr_block":                      "10.0.1.0/24",
			"availability_zone":               "us-east-1a",
			"map_public_ip_on_launch":         true,
			"assign_ipv6_address_on_creation": false,
		},
		Tags: map[string]string{
			"Name": "prod-subnet-1a",
		},
	}

	if subnet.Type != "aws_subnet" {
		t.Errorf("expected type aws_subnet")
	}
	if subnet.Attributes["cidr_block"] != "10.0.1.0/24" {
		t.Errorf("expected cidr_block 10.0.1.0/24")
	}
}

func TestDiscoveredResource_LoadBalancer(t *testing.T) {
	// Test typical load balancer discovery result
	lb := &DiscoveredResource{
		ID:     "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/prod-alb/1234567890abcdef",
		Type:   "aws_lb",
		ARN:    "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/prod-alb/1234567890abcdef",
		Name:   "prod-alb",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"type":             "application",
			"scheme":           "internet-facing",
			"vpc_id":           "vpc-12345",
			"subnets":          []string{"subnet-12345", "subnet-67890"},
			"security_groups":  []string{"sg-12345"},
			"dns_name":         "prod-alb-123456789.us-east-1.elb.amazonaws.com",
			"state":            "active",
		},
		Tags: map[string]string{
			"Name": "prod-alb",
		},
	}

	if lb.Type != "aws_lb" {
		t.Errorf("expected type aws_lb")
	}
	if lb.Attributes["scheme"] != "internet-facing" {
		t.Errorf("expected scheme internet-facing")
	}
}

func TestDiscoveredResource_ElastiCache(t *testing.T) {
	// Test typical ElastiCache replication group discovery result
	cache := &DiscoveredResource{
		ID:     "my-redis-cluster",
		Type:   "aws_elasticache_replication_group",
		ARN:    "arn:aws:elasticache:us-east-1:123456789:replicationgroup:my-redis-cluster",
		Name:   "my-redis-cluster",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"description":                "Redis cluster for caching",
			"node_type":                  "cache.t3.micro",
			"num_cache_clusters":         3,
			"automatic_failover_enabled": "true",
			"multi_az_enabled":           "true",
			"status":                     "available",
		},
		Tags: map[string]string{},
	}

	if cache.Type != "aws_elasticache_replication_group" {
		t.Errorf("expected type aws_elasticache_replication_group")
	}
	if cache.Attributes["node_type"] != "cache.t3.micro" {
		t.Errorf("expected node_type cache.t3.micro")
	}
}

func TestDiscoveredResourceEmptyTags(t *testing.T) {
	// Test resource with no tags
	resource := &DiscoveredResource{
		ID:     "vpc-12345",
		Type:   "aws_vpc",
		Region: "us-east-1",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
		Tags: map[string]string{},
	}

	if len(resource.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(resource.Tags))
	}
}

func TestDiscoveredResourceWithARN(t *testing.T) {
	// Test resource with ARN
	resource := &DiscoveredResource{
		ID:   "cluster-12345",
		Type: "aws_eks_cluster",
		ARN:  "arn:aws:eks:us-east-1:123456789:cluster/cluster-12345",
		Name: "cluster-12345",
	}

	if resource.ARN != "arn:aws:eks:us-east-1:123456789:cluster/cluster-12345" {
		t.Errorf("expected ARN to be set")
	}
}

func TestDriftResult_AllEmpty(t *testing.T) {
	// Test DriftResult with no drift
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*terraform.Resource{},
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

// Test types without terraform dependency
func TestResourceDiff_ComplexAttributes(t *testing.T) {
	// Test ResourceDiff with complex nested attributes
	diff := &ResourceDiff{
		ResourceID:   "subnet-12345",
		ResourceType: "aws_subnet",
		TerraformState: map[string]interface{}{
			"vpc_id":      "vpc-12345",
			"cidr_block":  "10.0.1.0/24",
			"route_tables": []interface{}{"rtb-123", "rtb-456"},
		},
		ActualState: map[string]interface{}{
			"vpc_id":     "vpc-12345",
			"cidr_block": "10.0.1.0/24",
			"route_tables": []interface{}{"rtb-123", "rtb-456", "rtb-789"},
		},
		Differences: []FieldDiff{
			{
				Field:          "route_tables",
				TerraformValue: []interface{}{"rtb-123", "rtb-456"},
				ActualValue:    []interface{}{"rtb-123", "rtb-456", "rtb-789"},
			},
		},
	}

	if len(diff.Differences) != 1 {
		t.Errorf("expected 1 difference, got %d", len(diff.Differences))
	}
}

// Tests for helper functions

func TestExtractTFResourceID_WithID(t *testing.T) {
	// Test extracting ID when "id" field is present
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id":         "vpc-12345",
			"cidr_block": "10.0.0.0/16",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "vpc-12345" {
		t.Errorf("expected vpc-12345, got %s", id)
	}
}

func TestExtractTFResourceID_WithInstanceID(t *testing.T) {
	// Test extracting ID when "instance_id" field is present
	tfRes := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"instance_id":   "i-12345",
			"instance_type": "t3.micro",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "i-12345" {
		t.Errorf("expected i-12345, got %s", id)
	}
}

func TestExtractTFResourceID_WithDBInstanceIdentifier(t *testing.T) {
	// Test extracting ID when "db_instance_identifier" field is present
	tfRes := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"db_instance_identifier": "db-prod",
			"engine":                 "postgres",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "db-prod" {
		t.Errorf("expected db-prod, got %s", id)
	}
}

func TestExtractTFResourceID_WithVPCID(t *testing.T) {
	// Test extracting ID when "vpc_id" field is present
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"vpc_id":     "vpc-12345",
			"cidr_block": "10.0.1.0/24",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "vpc-12345" {
		t.Errorf("expected vpc-12345, got %s", id)
	}
}

func TestExtractTFResourceID_WithSubnetID(t *testing.T) {
	// Test extracting ID when only "subnet_id" field is present (no "id" field)
	tfRes := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"subnet_id": "subnet-12345",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "subnet-12345" {
		t.Errorf("expected subnet-12345, got %s", id)
	}
}

func TestExtractTFResourceID_WithGroupID(t *testing.T) {
	// Test extracting ID when only "group_id" field is present (no "id" field)
	tfRes := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"group_id": "sg-12345",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "sg-12345" {
		t.Errorf("expected sg-12345, got %s", id)
	}
}

func TestExtractTFResourceID_WithARN(t *testing.T) {
	// Test extracting ID when "arn" field is present (fallback)
	tfRes := &terraform.Resource{
		Type: "aws_eks_cluster",
		Attributes: map[string]interface{}{
			"arn": "arn:aws:eks:us-east-1:123456789:cluster/my-cluster",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "arn:aws:eks:us-east-1:123456789:cluster/my-cluster" {
		t.Errorf("expected ARN, got %s", id)
	}
}

func TestExtractTFResourceID_NoIDField(t *testing.T) {
	// Test when no ID field is found
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "" {
		t.Errorf("expected empty string, got %s", id)
	}
}

func TestExtractTFResourceID_EmptyID(t *testing.T) {
	// Test when ID field is empty string
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id":         "",
			"cidr_block": "10.0.0.0/16",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "" {
		t.Errorf("expected empty string for empty id field, got %s", id)
	}
}

func TestExtractTFResourceID_PriorityOrder(t *testing.T) {
	// Test that "id" field takes priority over other fields
	tfRes := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"id":          "vpc-priority",
			"instance_id": "i-should-be-ignored",
		},
	}

	id := extractTFResourceID(tfRes)
	if id != "vpc-priority" {
		t.Errorf("expected id to take priority, got %s", id)
	}
}

func TestGetNestedValue_TopLevel(t *testing.T) {
	// Test retrieving a top-level value
	data := map[string]interface{}{
		"cidr_block": "10.0.0.0/16",
		"vpc_id":     "vpc-12345",
	}

	value := getNestedValue(data, "cidr_block")
	if value != "10.0.0.0/16" {
		t.Errorf("expected 10.0.0.0/16, got %v", value)
	}
}

func TestGetNestedValue_Nested(t *testing.T) {
	// Test retrieving a nested value using dot notation
	data := map[string]interface{}{
		"vpc_config": map[string]interface{}{
			"subnet_ids": []string{"subnet-123", "subnet-456"},
		},
	}

	value := getNestedValue(data, "vpc_config.subnet_ids")
	if value == nil {
		t.Errorf("expected to find nested value")
	}
}

func TestGetNestedValue_NotFound(t *testing.T) {
	// Test when value is not found
	data := map[string]interface{}{
		"cidr_block": "10.0.0.0/16",
	}

	value := getNestedValue(data, "nonexistent_field")
	if value != nil {
		t.Errorf("expected nil for nonexistent field, got %v", value)
	}
}

func TestGetNestedValue_DeepNested(t *testing.T) {
	// Test retrieving a deeply nested value
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "deep_value",
			},
		},
	}

	value := getNestedValue(data, "level1.level2.level3")
	if value != "deep_value" {
		t.Errorf("expected deep_value, got %v", value)
	}
}

func TestValuesEqual_BothNil(t *testing.T) {
	// Test equality when both values are nil
	if !valuesEqual(nil, nil) {
		t.Errorf("expected nil == nil to be true")
	}
}

func TestValuesEqual_OneNil(t *testing.T) {
	// Test inequality when one value is nil
	if valuesEqual("value", nil) {
		t.Errorf("expected value != nil")
	}
	if valuesEqual(nil, "value") {
		t.Errorf("expected nil != value")
	}
}

func TestValuesEqual_SameString(t *testing.T) {
	// Test equality for same strings
	if !valuesEqual("test", "test") {
		t.Errorf("expected test == test")
	}
}

func TestValuesEqual_DifferentString(t *testing.T) {
	// Test inequality for different strings
	if valuesEqual("test1", "test2") {
		t.Errorf("expected test1 != test2")
	}
}

func TestValuesEqual_SameBool(t *testing.T) {
	// Test equality for same booleans
	if !valuesEqual(true, true) {
		t.Errorf("expected true == true")
	}
	if !valuesEqual(false, false) {
		t.Errorf("expected false == false")
	}
}

func TestValuesEqual_DifferentBool(t *testing.T) {
	// Test inequality for different booleans
	if valuesEqual(true, false) {
		t.Errorf("expected true != false")
	}
}

func TestValuesEqual_BoolString(t *testing.T) {
	// Test boolean to string comparison
	if !valuesEqual(true, "true") {
		t.Errorf("expected true == \"true\"")
	}
	if !valuesEqual(false, "false") {
		t.Errorf("expected false == \"false\"")
	}
}

func TestValuesEqual_SameNumber(t *testing.T) {
	// Test equality for same numbers
	if !valuesEqual(42, 42) {
		t.Errorf("expected 42 == 42")
	}
}

func TestValuesEqual_DifferentNumber(t *testing.T) {
	// Test inequality for different numbers
	if valuesEqual(42, 43) {
		t.Errorf("expected 42 != 43")
	}
}

func TestValuesEqual_StringConversion(t *testing.T) {
	// Test string conversion for different types
	if !valuesEqual(100, "100") {
		t.Errorf("expected 100 == \"100\" via string conversion")
	}
}

func TestGetTerraformTags_WithTags(t *testing.T) {
	// Test extracting tags from "tags" field
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "production",
			"Owner":       "team-a",
		},
	}

	tags := getTerraformTags(attrs)
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	if tags["Environment"] != "production" {
		t.Errorf("expected Environment=production")
	}
}

func TestGetTerraformTags_WithTagsAll(t *testing.T) {
	// Test extracting tags from "tags_all" field (Terraform AWS provider v4+)
	attrs := map[string]interface{}{
		"tags_all": map[string]interface{}{
			"Environment": "staging",
			"Version":     "v1",
		},
	}

	tags := getTerraformTags(attrs)
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	if tags["Environment"] != "staging" {
		t.Errorf("expected Environment=staging")
	}
}

func TestGetTerraformTags_NoTags(t *testing.T) {
	// Test when no tags field exists
	attrs := map[string]interface{}{
		"cidr_block": "10.0.0.0/16",
	}

	tags := getTerraformTags(attrs)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestGetTerraformTags_EmptyTags(t *testing.T) {
	// Test with empty tags map
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{},
	}

	tags := getTerraformTags(attrs)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestGetComparableFields_AWSVpc(t *testing.T) {
	// Test comparable fields for aws_vpc
	fields := getComparableFields("aws_vpc")
	expectedFields := []string{"cidr_block", "enable_dns_hostnames", "enable_dns_support"}

	if len(fields) != len(expectedFields) {
		t.Errorf("expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestGetComparableFields_AWSSubnet(t *testing.T) {
	// Test comparable fields for aws_subnet
	fields := getComparableFields("aws_subnet")
	expectedFields := []string{"vpc_id", "cidr_block", "availability_zone", "map_public_ip_on_launch"}

	if len(fields) != len(expectedFields) {
		t.Errorf("expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestGetComparableFields_AWSSecurityGroup(t *testing.T) {
	// Test comparable fields for aws_security_group
	fields := getComparableFields("aws_security_group")
	expectedFields := []string{"vpc_id", "description", "name"}

	if len(fields) != len(expectedFields) {
		t.Errorf("expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestGetComparableFields_AWSInstance(t *testing.T) {
	// Test comparable fields for aws_instance
	fields := getComparableFields("aws_instance")
	if len(fields) == 0 {
		t.Errorf("expected non-empty fields for aws_instance")
	}
}

func TestGetComparableFields_AWSDBInstance(t *testing.T) {
	// Test comparable fields for aws_db_instance
	fields := getComparableFields("aws_db_instance")
	if len(fields) == 0 {
		t.Errorf("expected non-empty fields for aws_db_instance")
	}
}

func TestGetComparableFields_AWSEKSCluster(t *testing.T) {
	// Test comparable fields for aws_eks_cluster
	fields := getComparableFields("aws_eks_cluster")
	expectedFields := []string{"version", "role_arn"}

	if len(fields) != len(expectedFields) {
		t.Errorf("expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestGetComparableFields_AWSElastiCache(t *testing.T) {
	// Test comparable fields for aws_elasticache_replication_group
	fields := getComparableFields("aws_elasticache_replication_group")
	if len(fields) == 0 {
		t.Errorf("expected non-empty fields for aws_elasticache_replication_group")
	}
}

func TestGetComparableFields_AWSLoadBalancer(t *testing.T) {
	// Test comparable fields for aws_lb
	fields := getComparableFields("aws_lb")
	expectedFields := []string{"type", "scheme", "vpc_id"}

	if len(fields) != len(expectedFields) {
		t.Errorf("expected %d fields, got %d", len(expectedFields), len(fields))
	}
}

func TestGetComparableFields_Unknown(t *testing.T) {
	// Test comparable fields for unknown resource type
	fields := getComparableFields("aws_unknown_type")
	if len(fields) != 0 {
		t.Errorf("expected 0 fields for unknown type, got %d", len(fields))
	}
}

func TestTagsEqual_SameTags(t *testing.T) {
	// Test that same tags are equal
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "prod",
			"Owner":       "team-a",
		},
	}
	awsTags := map[string]string{
		"Environment": "prod",
		"Owner":       "team-a",
	}

	if !tagsEqual(tfAttrs, awsTags) {
		t.Errorf("expected tags to be equal")
	}
}

func TestTagsEqual_DifferentTags(t *testing.T) {
	// Test that different tags are not equal
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "prod",
		},
	}
	awsTags := map[string]string{
		"Environment": "staging",
	}

	if tagsEqual(tfAttrs, awsTags) {
		t.Errorf("expected tags to be different")
	}
}

func TestTagsEqual_IgnoreAWSManagedTags(t *testing.T) {
	// Test that AWS-managed tags are ignored
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{
			"Environment": "prod",
		},
	}
	awsTags := map[string]string{
		"Environment": "prod",
		"aws:cloudformation:stack-name": "my-stack",
		"kubernetes.io/cluster/name":     "my-cluster",
	}

	if !tagsEqual(tfAttrs, awsTags) {
		t.Errorf("expected AWS-managed tags to be ignored")
	}
}

func TestTagsEqual_EmptyTags(t *testing.T) {
	// Test with empty tags
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{},
	}
	awsTags := map[string]string{}

	if !tagsEqual(tfAttrs, awsTags) {
		t.Errorf("expected empty tags to be equal")
	}
}

func TestCompareResourceAttributes_SameType(t *testing.T) {
	// Test comparing resources of the same type
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id":                    "vpc-12345",
			"cidr_block":            "10.0.0.0/16",
			"enable_dns_hostnames":  true,
			"enable_dns_support":    true,
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-12345",
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block":            "10.0.0.0/16",
			"enable_dns_hostnames":  true,
			"enable_dns_support":    true,
		},
		Tags: map[string]string{},
	}

	diff := compareResourceAttributes(tfRes, awsRes)
	if diff == nil {
		t.Errorf("expected diff to not be nil")
	}
	if diff.ResourceID != "vpc-12345" {
		t.Errorf("expected ResourceID vpc-12345")
	}
}

func TestCompareResourceAttributes_DiscoveryTypeMismatch(t *testing.T) {
	// Test that type mismatch returns nil
	tfRes := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"id": "vpc-12345",
		},
	}

	awsRes := &DiscoveredResource{
		ID:   "vpc-12345",
		Type: "aws_subnet",
		Attributes: map[string]interface{}{},
		Tags: map[string]string{},
	}

	diff := compareResourceAttributes(tfRes, awsRes)
	if diff != nil {
		t.Errorf("expected diff to be nil for type mismatch")
	}
}
