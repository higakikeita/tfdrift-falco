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
