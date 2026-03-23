package graph

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// --- TerraformToGraph ---

func TestTerraformToGraph_BasicVPCSubnet(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Mode:     "managed",
			Type:     "aws_vpc",
			Name:     "main",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":         "vpc-123",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Mode:     "managed",
			Type:     "aws_subnet",
			Name:     "public",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":         "subnet-456",
				"vpc_id":     "vpc-123",
				"cidr_block": "10.0.1.0/24",
			},
		},
	}

	db := TerraformToGraph(resources, map[string]bool{})

	if db.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", db.NodeCount())
	}

	// Subnet should have PART_OF relationship to VPC
	rels := db.GetOutgoingRelationships("subnet-456")
	found := false
	for _, rel := range rels {
		if rel.Type == PART_OF && rel.EndNode == "vpc-123" {
			found = true
		}
	}
	if !found {
		t.Error("expected PART_OF relationship from subnet to VPC")
	}
}

func TestTerraformToGraph_EC2Dependencies(t *testing.T) {
	resources := []*terraform.Resource{
		{Type: "aws_vpc", Name: "main", Attributes: map[string]interface{}{"id": "vpc-1"}},
		{Type: "aws_subnet", Name: "sub", Attributes: map[string]interface{}{"id": "subnet-1", "vpc_id": "vpc-1"}},
		{Type: "aws_security_group", Name: "sg", Attributes: map[string]interface{}{"id": "sg-1", "vpc_id": "vpc-1"}},
		{
			Type: "aws_instance", Name: "web",
			Attributes: map[string]interface{}{
				"id":                       "i-123",
				"subnet_id":               "subnet-1",
				"vpc_security_group_ids":   []interface{}{"sg-1"},
				"instance_type":            "t3.micro",
			},
		},
	}

	db := TerraformToGraph(resources, map[string]bool{})

	// EC2 should depend on subnet
	ec2Rels := db.GetOutgoingRelationships("i-123")
	hasDep := false
	for _, rel := range ec2Rels {
		if rel.Type == DEPENDS_ON && rel.EndNode == "subnet-1" {
			hasDep = true
		}
	}
	if !hasDep {
		t.Error("expected EC2 DEPENDS_ON subnet")
	}

	// SG should SECURES EC2
	sgRels := db.GetOutgoingRelationships("sg-1")
	hasSec := false
	for _, rel := range sgRels {
		if rel.Type == SECURES && rel.EndNode == "i-123" {
			hasSec = true
		}
	}
	if !hasSec {
		t.Error("expected SG SECURES EC2")
	}
}

func TestTerraformToGraph_DriftedLabel(t *testing.T) {
	resources := []*terraform.Resource{
		{Type: "aws_vpc", Name: "main", Attributes: map[string]interface{}{"id": "vpc-1"}},
	}
	drifted := map[string]bool{"vpc-1": true}

	db := TerraformToGraph(resources, drifted)
	if !db.HasLabel("vpc-1", "Drifted") {
		t.Error("expected vpc-1 to have Drifted label")
	}
}

func TestTerraformToGraph_FallbackID(t *testing.T) {
	resources := []*terraform.Resource{
		{Type: "aws_unknown", Name: "foo", Attributes: map[string]interface{}{}},
	}
	db := TerraformToGraph(resources, map[string]bool{})
	node := db.GetNode("aws_unknown.foo")
	if node == nil {
		t.Error("expected fallback ID aws_unknown.foo")
	}
}

// --- resourceToNode label assignment ---

func TestResourceToNode_Labels(t *testing.T) {
	tests := []struct {
		resType    string
		wantLabels []string
	}{
		{"aws_vpc", []string{"Resource", "VPC", "Network"}},
		{"aws_subnet", []string{"Resource", "Subnet", "Network"}},
		{"aws_instance", []string{"Resource", "EC2", "Compute"}},
		{"aws_db_instance", []string{"Resource", "RDS", "Database"}},
		{"aws_security_group", []string{"Resource", "SecurityGroup", "Security"}},
		{"aws_nat_gateway", []string{"Resource", "NATGateway", "Network"}},
		{"aws_internet_gateway", []string{"Resource", "InternetGateway", "Network"}},
		{"aws_route_table", []string{"Resource", "RouteTable", "Network"}},
	}

	for _, tt := range tests {
		t.Run(tt.resType, func(t *testing.T) {
			res := &terraform.Resource{
				Type:       tt.resType,
				Name:       "test",
				Attributes: map[string]interface{}{"id": "test-id"},
			}
			node := resourceToNode(res, map[string]bool{})
			if node == nil {
				t.Fatal("got nil node")
			}
			for _, want := range tt.wantLabels {
				found := false
				for _, got := range node.Labels {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing label %q in %v", want, node.Labels)
				}
			}
		})
	}
}

// --- extractResourceIDFromAttributes ---

func TestExtractResourceIDFromAttributes(t *testing.T) {
	tests := []struct {
		name  string
		attrs map[string]interface{}
		want  string
	}{
		{"with id", map[string]interface{}{"id": "vpc-123"}, "vpc-123"},
		{"with arn", map[string]interface{}{"arn": "arn:aws:ec2:us-east-1:123:vpc/vpc-1"}, "arn:aws:ec2:us-east-1:123:vpc/vpc-1"},
		{"with name only", map[string]interface{}{"name": "my-resource"}, "my-resource"},
		{"with self_link", map[string]interface{}{"self_link": "https://compute.googleapis.com/..."}, "https://compute.googleapis.com/..."},
		{"empty", map[string]interface{}{}, ""},
		{"id takes precedence", map[string]interface{}{"id": "vpc-1", "arn": "arn:..."}, "vpc-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractResourceIDFromAttributes(tt.attrs)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- extractResourceName ---

func TestExtractResourceName(t *testing.T) {
	tests := []struct {
		name string
		res  *terraform.Resource
		want string
	}{
		{
			"from name attr",
			&terraform.Resource{Name: "tf-name", Attributes: map[string]interface{}{"name": "my-vpc"}},
			"my-vpc",
		},
		{
			"from tags.Name",
			&terraform.Resource{Name: "tf-name", Attributes: map[string]interface{}{"tags": map[string]interface{}{"Name": "tagged-name"}}},
			"tagged-name",
		},
		{
			"fallback to tf name",
			&terraform.Resource{Name: "tf-name", Type: "aws_vpc", Attributes: map[string]interface{}{}},
			"tf-name",
		},
		{
			"fallback to type",
			&terraform.Resource{Type: "aws_vpc", Attributes: map[string]interface{}{}},
			"aws_vpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractResourceName(tt.res)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- CreateEdge ---

func TestCreateEdge(t *testing.T) {
	edge := CreateEdge("src-1", "tgt-1", "depends", "caused_by", "dependency")
	if edge.Data.Source != "src-1" {
		t.Errorf("Source = %q, want src-1", edge.Data.Source)
	}
	if edge.Data.Target != "tgt-1" {
		t.Errorf("Target = %q, want tgt-1", edge.Data.Target)
	}
	if edge.Data.ID != "src-1-tgt-1" {
		t.Errorf("ID = %q, want src-1-tgt-1", edge.Data.ID)
	}
}
