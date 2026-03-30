package graph

import (
	"fmt"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformToGraph_Empty tests conversion with empty resource list
func TestTerraformToGraph_Empty(t *testing.T) {
	resources := []*terraform.Resource{}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 0, db.NodeCount())
	assert.Equal(t, 0, db.RelationshipCount())
}

// TestTerraformToGraph_SingleEC2Instance tests conversion with single EC2 instance
func TestTerraformToGraph_SingleEC2Instance(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "web_server",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                     "i-1234567890abcdef0",
			"instance_type":          "t2.micro",
			"ami":                    "ami-0c55b159cbfafe1f0",
			"private_ip":             "10.0.1.100",
			"public_ip":              "52.1.2.3",
			"instance_state":         "running",
			"subnet_id":              "subnet-12345678",
			"vpc_security_group_ids": []interface{}{"sg-12345678"},
		},
	}

	resources := []*terraform.Resource{resource}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 1, db.NodeCount())

	node := db.GetNode("i-1234567890abcdef0")
	assert.NotNil(t, node)
	assert.Equal(t, "i-1234567890abcdef0", node.ID)
	assert.Equal(t, "aws_instance", node.Properties["type"])
	assert.Equal(t, "t2.micro", node.Properties["instance_type"])
	assert.Equal(t, "10.0.1.100", node.Properties["private_ip"])
	assert.Equal(t, "52.1.2.3", node.Properties["public_ip"])
	assert.Equal(t, "running", node.Properties["state"])
	assert.Contains(t, node.Labels, "EC2")
	assert.Contains(t, node.Labels, "Compute")
	assert.Contains(t, node.Labels, "Resource")
}

// TestTerraformToGraph_MultipleResourcesWithRelationships tests conversion with multiple related resources
func TestTerraformToGraph_MultipleResourcesWithRelationships(t *testing.T) {
	vpc := &terraform.Resource{
		Type:     "aws_vpc",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "vpc-12345678",
			"cidr_block": "10.0.0.0/16",
		},
	}

	subnet := &terraform.Resource{
		Type:     "aws_subnet",
		Name:     "private",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                "subnet-12345678",
			"vpc_id":            "vpc-12345678",
			"cidr_block":        "10.0.1.0/24",
			"availability_zone": "us-east-1a",
		},
	}

	ec2 := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "app",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":        "i-0987654321fedcba0",
			"subnet_id": "subnet-12345678",
		},
	}

	resources := []*terraform.Resource{vpc, subnet, ec2}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 3, db.NodeCount())

	// Verify nodes exist
	vpcNode := db.GetNode("vpc-12345678")
	assert.NotNil(t, vpcNode)
	assert.Contains(t, vpcNode.Labels, "VPC")

	subnetNode := db.GetNode("subnet-12345678")
	assert.NotNil(t, subnetNode)
	assert.Contains(t, subnetNode.Labels, "Subnet")

	ec2Node := db.GetNode("i-0987654321fedcba0")
	assert.NotNil(t, ec2Node)
	assert.Contains(t, ec2Node.Labels, "EC2")

	// Verify relationships
	assert.Greater(t, db.RelationshipCount(), 0)

	// Check for subnet-vpc relationship
	subnetVPCRelationships := db.GetRelationshipsByType(PART_OF)
	assert.Greater(t, len(subnetVPCRelationships), 0)

	// Check for ec2-subnet relationship
	ec2SubnetRelationships := db.GetRelationshipsByType(DEPENDS_ON)
	assert.Greater(t, len(ec2SubnetRelationships), 0)
}

// TestTerraformToGraph_WithDriftedIDs tests conversion with drifted resource IDs
func TestTerraformToGraph_WithDriftedIDs(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "web",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id": "i-drifted123456789",
		},
	}

	resources := []*terraform.Resource{resource}
	driftedIDs := map[string]bool{
		"i-drifted123456789": true,
	}

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 1, db.NodeCount())

	node := db.GetNode("i-drifted123456789")
	assert.NotNil(t, node)
	assert.Contains(t, node.Labels, "Drifted")
	assert.Equal(t, true, node.Properties["has_drift"])
}

// TestTerraformToGraph_NilDriftedIDs tests conversion with nil drifted IDs map
func TestTerraformToGraph_NilDriftedIDs(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "test",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id": "i-test123",
		},
	}

	resources := []*terraform.Resource{resource}

	db := TerraformToGraph(resources, nil)

	assert.NotNil(t, db)
	assert.Equal(t, 1, db.NodeCount())

	node := db.GetNode("i-test123")
	assert.NotNil(t, node)
	assert.Equal(t, false, node.Properties["has_drift"])
	assert.NotContains(t, node.Labels, "Drifted")
}

// TestResourceToNode_VPC tests node creation for VPC resource
func TestResourceToNode_VPC(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_vpc",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "vpc-abc123",
			"cidr_block": "10.0.0.0/16",
		},
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.Equal(t, "vpc-abc123", node.ID)
	assert.Equal(t, "aws_vpc", node.Properties["type"])
	assert.Equal(t, "10.0.0.0/16", node.Properties["cidr"])
	assert.Contains(t, node.Labels, "VPC")
	assert.Contains(t, node.Labels, "Network")
}

// TestResourceToNode_Subnet tests node creation for Subnet resource
func TestResourceToNode_Subnet(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_subnet",
		Name:     "private_a",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                      "subnet-def456",
			"vpc_id":                  "vpc-abc123",
			"cidr_block":              "10.0.1.0/24",
			"availability_zone":       "us-east-1a",
			"map_public_ip_on_launch": true,
		},
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.Equal(t, "subnet-def456", node.ID)
	assert.Equal(t, "10.0.1.0/24", node.Properties["cidr"])
	assert.Equal(t, "us-east-1a", node.Properties["availability_zone"])
	assert.Equal(t, true, node.Properties["public"])
	assert.Contains(t, node.Labels, "Subnet")
}

// TestResourceToNode_SecurityGroup tests node creation for security group resource
func TestResourceToNode_SecurityGroup(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_security_group",
		Name:     "web",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":          "sg-12345678",
			"name":        "web-sg",
			"description": "Security group for web servers",
			"vpc_id":      "vpc-abc123",
		},
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.Equal(t, "sg-12345678", node.ID)
	assert.Equal(t, "aws_security_group", node.Properties["type"])
	assert.Equal(t, "web-sg", node.Properties["sg_name"])
	assert.Equal(t, "Security group for web servers", node.Properties["description"])
	assert.Contains(t, node.Labels, "SecurityGroup")
	assert.Contains(t, node.Labels, "Security")
}

// TestResourceToNode_RDS tests node creation for RDS instance
func TestResourceToNode_RDS(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_db_instance",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                "mydb",
			"engine":            "mysql",
			"instance_class":    "db.t2.micro",
			"status":            "available",
			"allocated_storage": 20,
		},
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.Equal(t, "mydb", node.ID)
	assert.Equal(t, "mysql", node.Properties["engine"])
	assert.Equal(t, "db.t2.micro", node.Properties["instance_class"])
	assert.Equal(t, "available", node.Properties["status"])
	assert.Contains(t, node.Labels, "RDS")
	assert.Contains(t, node.Labels, "Database")
}

// TestResourceToNode_MissingID tests node creation with missing ID (uses fallback)
func TestResourceToNode_MissingID(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "test",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"instance_type": "t2.micro",
		},
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	// Should use fallback ID format
	assert.Equal(t, "aws_instance.test", node.ID)
	assert.Equal(t, "aws_instance", node.Properties["type"])
}

// TestResourceToNode_NilAttributes tests node creation with nil attributes
func TestResourceToNode_NilAttributes(t *testing.T) {
	resource := &terraform.Resource{
		Type:       "aws_instance",
		Name:       "test",
		Mode:       "managed",
		Provider:   "aws",
		Attributes: nil,
	}

	driftedIDs := make(map[string]bool)
	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.NotNil(t, node.Properties)
	assert.Equal(t, "aws_instance", node.Properties["type"])
}

// TestResourceToNode_WithDrift tests node creation with drift flag
func TestResourceToNode_WithDrift(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "drifted",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id": "i-drift123",
		},
	}

	driftedIDs := map[string]bool{
		"i-drift123": true,
	}

	node := resourceToNode(resource, driftedIDs)

	assert.NotNil(t, node)
	assert.Contains(t, node.Labels, "Drifted")
	assert.Equal(t, true, node.Properties["has_drift"])
}

// TestAddResourceSpecificProperties_VPC tests VPC-specific properties
func TestAddResourceSpecificProperties_VPC(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	assert.Equal(t, "10.0.0.0/16", properties["cidr"])
}

// TestAddResourceSpecificProperties_Subnet tests Subnet-specific properties
func TestAddResourceSpecificProperties_Subnet(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"cidr_block":              "10.0.1.0/24",
			"availability_zone":       "us-west-2a",
			"map_public_ip_on_launch": false,
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	assert.Equal(t, "10.0.1.0/24", properties["cidr"])
	assert.Equal(t, "us-west-2a", properties["availability_zone"])
	assert.Equal(t, false, properties["public"])
}

// TestAddResourceSpecificProperties_EC2 tests EC2-specific properties
func TestAddResourceSpecificProperties_EC2(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"instance_type":  "t3.large",
			"instance_state": "running",
			"private_ip":     "10.0.0.50",
			"public_ip":      "203.0.113.10",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	assert.Equal(t, "t3.large", properties["instance_type"])
	assert.Equal(t, "running", properties["state"])
	assert.Equal(t, "10.0.0.50", properties["private_ip"])
	assert.Equal(t, "203.0.113.10", properties["public_ip"])
}

// TestAddResourceSpecificProperties_RDS tests RDS-specific properties
func TestAddResourceSpecificProperties_RDS(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"engine":         "postgres",
			"instance_class": "db.t3.small",
			"status":         "creating",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	assert.Equal(t, "postgres", properties["engine"])
	assert.Equal(t, "db.t3.small", properties["instance_class"])
	assert.Equal(t, "creating", properties["status"])
}

// TestAddResourceSpecificProperties_SecurityGroup tests security group properties
func TestAddResourceSpecificProperties_SecurityGroup(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"name":        "app-sg",
			"description": "Application security group",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	assert.Equal(t, "app-sg", properties["sg_name"])
	assert.Equal(t, "Application security group", properties["description"])
}

// TestAddResourceSpecificProperties_WithTags tests tags extraction
func TestAddResourceSpecificProperties_WithTags(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"tags": map[string]interface{}{
				"Name":        "prod-server",
				"Environment": "production",
			},
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	tags, ok := properties["tags"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "prod-server", tags["Name"])
	assert.Equal(t, "production", tags["Environment"])
}

// TestExtractRelationships_EC2WithSubnet tests EC2 to subnet relationship
func TestExtractRelationships_EC2WithSubnet(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"id":        "i-test123",
			"subnet_id": "subnet-test456",
		},
	}

	relationships := extractRelationships(resource)

	assert.Greater(t, len(relationships), 0)

	found := false
	for _, rel := range relationships {
		if rel.Type == DEPENDS_ON && rel.StartNode == "i-test123" && rel.EndNode == "subnet-test456" {
			found = true
			assert.Equal(t, "network_placement", rel.Properties["type"])
		}
	}
	assert.True(t, found)
}

// TestExtractRelationships_SubnetWithVPC tests subnet to VPC relationship
func TestExtractRelationships_SubnetWithVPC(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"id":     "subnet-test123",
			"vpc_id": "vpc-main456",
		},
	}

	relationships := extractRelationships(resource)

	assert.Greater(t, len(relationships), 0)

	found := false
	for _, rel := range relationships {
		if rel.Type == PART_OF && rel.StartNode == "subnet-test123" && rel.EndNode == "vpc-main456" {
			found = true
		}
	}
	assert.True(t, found)
}

// TestExtractRelationships_EC2WithSecurityGroups tests EC2 to security group relationships
func TestExtractRelationships_EC2WithSecurityGroups(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"id": "i-sg-test",
			"vpc_security_group_ids": []interface{}{
				"sg-123456",
				"sg-789012",
			},
		},
	}

	relationships := extractRelationships(resource)

	sgRelationships := 0
	for _, rel := range relationships {
		if rel.Type == SECURES {
			sgRelationships++
		}
	}
	assert.Equal(t, 2, sgRelationships)
}

// TestExtractRelationships_SecurityGroupWithVPC tests security group to VPC relationship
func TestExtractRelationships_SecurityGroupWithVPC(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"id":     "sg-test123",
			"vpc_id": "vpc-secure456",
		},
	}

	relationships := extractRelationships(resource)

	found := false
	for _, rel := range relationships {
		if rel.Type == SECURES && rel.StartNode == "sg-test123" && rel.EndNode == "vpc-secure456" {
			found = true
		}
	}
	assert.True(t, found)
}

// TestExtractRelationships_NATGatewayWithSubnet tests NAT gateway to subnet relationship
func TestExtractRelationships_NATGatewayWithSubnet(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_nat_gateway",
		Attributes: map[string]interface{}{
			"id":        "nat-12345",
			"subnet_id": "subnet-nat-test",
		},
	}

	relationships := extractRelationships(resource)

	found := false
	for _, rel := range relationships {
		if rel.Type == DEPENDS_ON && rel.StartNode == "nat-12345" && rel.EndNode == "subnet-nat-test" {
			found = true
		}
	}
	assert.True(t, found)
}

// TestExtractRelationships_RDSWithSubnetGroup tests RDS to subnet group relationship
func TestExtractRelationships_RDSWithSubnetGroup(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"id":                     "mydb",
			"db_subnet_group_name":   "default-sg",
			"vpc_security_group_ids": []interface{}{"sg-rds-123"},
		},
	}

	relationships := extractRelationships(resource)

	assert.Greater(t, len(relationships), 0)

	foundSubnetGroup := false
	foundSG := false

	for _, rel := range relationships {
		if rel.Type == DEPENDS_ON && rel.EndNode == "default-sg" {
			foundSubnetGroup = true
		}
		if rel.Type == SECURES && rel.EndNode == "mydb" {
			foundSG = true
		}
	}

	assert.True(t, foundSubnetGroup)
	assert.True(t, foundSG)
}

// TestExtractRelationships_RouteTableWithVPC tests route table to VPC relationship
func TestExtractRelationships_RouteTableWithVPC(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_route_table",
		Attributes: map[string]interface{}{
			"id":     "rtb-test123",
			"vpc_id": "vpc-route456",
		},
	}

	relationships := extractRelationships(resource)

	found := false
	for _, rel := range relationships {
		if rel.Type == ROUTES_TO && rel.StartNode == "rtb-test123" && rel.EndNode == "vpc-route456" {
			found = true
		}
	}
	assert.True(t, found)
}

// TestExtractRelationships_NoRelationships tests resource with no relationships
func TestExtractRelationships_NoRelationships(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_iam_role",
		Attributes: map[string]interface{}{
			"id":   "role-test",
			"name": "test-role",
		},
	}

	relationships := extractRelationships(resource)

	// IAM roles may have no standard relationships extracted
	assert.NotNil(t, relationships)
}

// TestExtractRelationships_MissingID tests relationship extraction with missing resource ID
func TestExtractRelationships_MissingID(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"subnet_id": "subnet-123",
		},
	}

	relationships := extractRelationships(resource)

	// Should return empty if resource ID cannot be extracted
	assert.Equal(t, 0, len(relationships))
}

// TestTerraformToGraph_LargeResourceSet tests conversion with many resources
func TestTerraformToGraph_LargeResourceSet(t *testing.T) {
	var resources []*terraform.Resource

	// Create VPC
	vpc := &terraform.Resource{
		Type:     "aws_vpc",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "vpc-large",
			"cidr_block": "10.0.0.0/16",
		},
	}
	resources = append(resources, vpc)

	// Create 50 subnets and EC2 instances
	for i := 0; i < 50; i++ {
		subnetID := assertSprintf(t, "subnet-%d", i)
		instanceID := assertSprintf(t, "i-%d", i)

		subnet := &terraform.Resource{
			Type:     "aws_subnet",
			Name:     assertSprintf(t, "subnet_%d", i),
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":     subnetID,
				"vpc_id": "vpc-large",
			},
		}
		resources = append(resources, subnet)

		instance := &terraform.Resource{
			Type:     "aws_instance",
			Name:     assertSprintf(t, "instance_%d", i),
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":        instanceID,
				"subnet_id": subnetID,
			},
		}
		resources = append(resources, instance)
	}

	driftedIDs := make(map[string]bool)
	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 101, db.NodeCount()) // 1 VPC + 50 subnets + 50 instances
	assert.Greater(t, db.RelationshipCount(), 100)
}

// TestTerraformToGraph_ResourceWithEmptyType tests resource with empty type
func TestTerraformToGraph_ResourceWithEmptyType(t *testing.T) {
	resource := &terraform.Resource{
		Type:     "",
		Name:     "empty",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id": "empty-id",
		},
	}

	resources := []*terraform.Resource{resource}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	node := db.GetNode("empty-id")
	assert.NotNil(t, node)
	assert.Equal(t, "", node.Properties["type"])
}

// TestTerraformToGraph_InternetGateway tests IGW creation and relationships
func TestTerraformToGraph_InternetGateway(t *testing.T) {
	igw := &terraform.Resource{
		Type:     "aws_internet_gateway",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":     "igw-12345",
			"vpc_id": "vpc-abc",
		},
	}

	resources := []*terraform.Resource{igw}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 1, db.NodeCount())

	node := db.GetNode("igw-12345")
	assert.NotNil(t, node)
	assert.Contains(t, node.Labels, "InternetGateway")

	relationships := extractRelationships(igw)
	assert.Greater(t, len(relationships), 0)

	found := false
	for _, rel := range relationships {
		if rel.Type == CONNECTS_TO && rel.StartNode == "igw-12345" && rel.EndNode == "vpc-abc" {
			found = true
		}
	}
	assert.True(t, found)
}

// TestTerraformToGraph_EKSCluster tests EKS cluster with VPC config
func TestTerraformToGraph_EKSCluster(t *testing.T) {
	eks := &terraform.Resource{
		Type:     "aws_eks_cluster",
		Name:     "demo",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":   "demo-cluster",
			"name": "demo",
			"vpc_config": []interface{}{
				map[string]interface{}{
					"subnet_ids": []interface{}{
						"subnet-1",
						"subnet-2",
					},
					"security_group_ids": []interface{}{
						"sg-eks-1",
					},
				},
			},
		},
	}

	resources := []*terraform.Resource{eks}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	node := db.GetNode("demo-cluster")
	assert.NotNil(t, node)

	relationships := extractRelationships(eks)
	subnetRels := 0
	sgRels := 0

	for _, rel := range relationships {
		if rel.Type == DEPENDS_ON && (rel.EndNode == "subnet-1" || rel.EndNode == "subnet-2") {
			subnetRels++
		}
		if rel.Type == SECURES && rel.EndNode == "demo-cluster" {
			sgRels++
		}
	}

	assert.Equal(t, 2, subnetRels)
	assert.Equal(t, 1, sgRels)
}

// TestTerraformToGraph_LoadBalancer tests ALB with subnets and security groups
func TestTerraformToGraph_LoadBalancer(t *testing.T) {
	alb := &terraform.Resource{
		Type:     "aws_lb",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":              "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/50dc6c495c0c9188",
			"name":            "my-alb",
			"subnets":         []interface{}{"subnet-a", "subnet-b"},
			"security_groups": []interface{}{"sg-alb-1", "sg-alb-2"},
		},
	}

	resources := []*terraform.Resource{alb}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	relationships := extractRelationships(alb)

	subnetDeps := 0
	sgSecures := 0

	for _, rel := range relationships {
		if rel.Type == DEPENDS_ON && (rel.EndNode == "subnet-a" || rel.EndNode == "subnet-b") {
			subnetDeps++
		}
		if rel.Type == SECURES && (rel.StartNode == "sg-alb-1" || rel.StartNode == "sg-alb-2") {
			sgSecures++
		}
	}

	assert.Equal(t, 2, subnetDeps)
	assert.Equal(t, 2, sgSecures)
}

// TestTerraformToGraph_DBSubnetGroup tests database subnet group relationships
func TestTerraformToGraph_DBSubnetGroup(t *testing.T) {
	dbsg := &terraform.Resource{
		Type:     "aws_db_subnet_group",
		Name:     "default",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "default",
			"subnet_ids": []interface{}{"subnet-db-1", "subnet-db-2", "subnet-db-3"},
		},
	}

	resources := []*terraform.Resource{dbsg}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	relationships := extractRelationships(dbsg)

	containsRels := 0
	for _, rel := range relationships {
		if rel.Type == CONTAINS {
			containsRels++
		}
	}
	assert.Equal(t, 3, containsRels)
}

// TestResourceToNode_AllResourceTypes tests node creation for all supported resource types
func TestResourceToNode_AllResourceTypes(t *testing.T) {
	resourceTypes := []string{
		"aws_vpc",
		"aws_subnet",
		"aws_instance",
		"aws_db_instance",
		"aws_nat_gateway",
		"aws_internet_gateway",
		"aws_route_table",
		"aws_security_group",
		"aws_eks_cluster",
		"aws_ecs_cluster",
		"aws_s3_bucket",
		"aws_iam_role",
	}

	for _, rType := range resourceTypes {
		resource := &terraform.Resource{
			Type:       rType,
			Name:       "test",
			Mode:       "managed",
			Provider:   "aws",
			Attributes: map[string]interface{}{"id": rType + "-123"},
		}

		driftedIDs := make(map[string]bool)
		node := resourceToNode(resource, driftedIDs)

		assert.NotNil(t, node, "Failed to create node for type %s", rType)
		assert.Equal(t, rType+"-123", node.ID)
		assert.Equal(t, rType, node.Properties["type"])
		assert.Contains(t, node.Labels, "Resource")
	}
}

// TestExtractRelationships_ECSService tests ECS service relationships
func TestExtractRelationships_ECSService(t *testing.T) {
	ecsService := &terraform.Resource{
		Type: "aws_ecs_service",
		Attributes: map[string]interface{}{
			"id":      "my-service",
			"cluster": "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster",
			"load_balancer": []interface{}{
				map[string]interface{}{
					"target_group_arn": "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/my-tg/50dc6c495c0c9188",
				},
			},
		},
	}

	relationships := extractRelationships(ecsService)

	runInFound := false
	routesFound := false

	for _, rel := range relationships {
		if rel.Type == RUNS_IN {
			runInFound = true
		}
		if rel.Type == REGISTERS_TO {
			routesFound = true
		}
	}

	assert.True(t, runInFound)
	assert.True(t, routesFound)
}

// TestExtractRelationships_RouteTableAssociation tests route table association relationships
func TestExtractRelationships_RouteTableAssociation(t *testing.T) {
	rta := &terraform.Resource{
		Type: "aws_route_table_association",
		Attributes: map[string]interface{}{
			"id":             "rtbassoc-123",
			"route_table_id": "rtb-456",
			"subnet_id":      "subnet-789",
		},
	}

	relationships := extractRelationships(rta)

	assert.Greater(t, len(relationships), 0)

	rtFound := false
	subnetFound := false

	for _, rel := range relationships {
		if rel.Type == ASSOCIATES && rel.EndNode == "rtb-456" {
			rtFound = true
		}
		if rel.Type == ASSOCIATES && rel.EndNode == "subnet-789" {
			subnetFound = true
		}
	}

	assert.True(t, rtFound)
	assert.True(t, subnetFound)
}

// TestTerraformToGraph_ComplexNetworkTopology tests a realistic network with VPC, subnets, EC2s, RDS, and security groups
func TestTerraformToGraph_ComplexNetworkTopology(t *testing.T) {
	vpc := &terraform.Resource{
		Type:     "aws_vpc",
		Name:     "prod",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "vpc-prod",
			"cidr_block": "10.0.0.0/16",
		},
	}

	publicSubnet := &terraform.Resource{
		Type:     "aws_subnet",
		Name:     "public",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":     "subnet-public",
			"vpc_id": "vpc-prod",
		},
	}

	privateSubnet := &terraform.Resource{
		Type:     "aws_subnet",
		Name:     "private",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":     "subnet-private",
			"vpc_id": "vpc-prod",
		},
	}

	webSG := &terraform.Resource{
		Type:     "aws_security_group",
		Name:     "web",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":          "sg-web",
			"vpc_id":      "vpc-prod",
			"description": "Web servers",
		},
	}

	dbSG := &terraform.Resource{
		Type:     "aws_security_group",
		Name:     "db",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":          "sg-db",
			"vpc_id":      "vpc-prod",
			"description": "Database servers",
		},
	}

	webServer := &terraform.Resource{
		Type:     "aws_instance",
		Name:     "web_server",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                     "i-web123",
			"subnet_id":              "subnet-public",
			"instance_type":          "t3.medium",
			"vpc_security_group_ids": []interface{}{"sg-web"},
		},
	}

	rds := &terraform.Resource{
		Type:     "aws_db_instance",
		Name:     "main",
		Mode:     "managed",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":                     "mydb",
			"engine":                 "postgres",
			"instance_class":         "db.t3.micro",
			"vpc_security_group_ids": []interface{}{"sg-db"},
		},
	}

	resources := []*terraform.Resource{vpc, publicSubnet, privateSubnet, webSG, dbSG, webServer, rds}
	driftedIDs := make(map[string]bool)

	db := TerraformToGraph(resources, driftedIDs)

	assert.NotNil(t, db)
	assert.Equal(t, 7, db.NodeCount())

	vpcNode := db.GetNode("vpc-prod")
	assert.NotNil(t, vpcNode)

	webServerNode := db.GetNode("i-web123")
	assert.NotNil(t, webServerNode)

	rdsNode := db.GetNode("mydb")
	assert.NotNil(t, rdsNode)

	// Verify relationships exist
	assert.Greater(t, db.RelationshipCount(), 5)
}

// Helper function to safely format a string for testing
func assertSprintf(t *testing.T, format string, args ...interface{}) string {
	require.NotEmpty(t, format)
	return fmt.Sprintf(format, args...)
}
