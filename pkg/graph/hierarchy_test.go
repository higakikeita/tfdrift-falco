package graph

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

func makeResource(resType, name string, attrs map[string]interface{}) *terraform.Resource {
	return &terraform.Resource{
		Mode:       "managed",
		Type:       resType,
		Name:       name,
		Provider:   "aws",
		Attributes: attrs,
	}
}

func TestBuildHierarchy_Basic(t *testing.T) {
	resources := []*terraform.Resource{
		makeResource("aws_vpc", "main", map[string]interface{}{
			"id":         "vpc-1",
			"cidr_block": "10.0.0.0/16",
		}),
		makeResource("aws_subnet", "pub", map[string]interface{}{
			"id":                    "subnet-1",
			"vpc_id":               "vpc-1",
			"cidr_block":           "10.0.1.0/24",
			"availability_zone":    "us-east-1a",
			"map_public_ip_on_launch": true,
		}),
		makeResource("aws_subnet", "priv", map[string]interface{}{
			"id":                 "subnet-2",
			"vpc_id":             "vpc-1",
			"cidr_block":         "10.0.2.0/24",
			"availability_zone":  "us-east-1b",
		}),
		makeResource("aws_instance", "web", map[string]interface{}{
			"id":        "i-1",
			"subnet_id": "subnet-1",
		}),
		makeResource("aws_security_group", "sg", map[string]interface{}{
			"id":     "sg-1",
			"vpc_id": "vpc-1",
		}),
	}

	hb := NewHierarchyBuilder()
	hierarchy := hb.BuildHierarchy(resources)

	// Should have 1 region
	if len(hierarchy.Regions) != 1 {
		t.Errorf("regions = %d, want 1", len(hierarchy.Regions))
	}

	// Find the region (default us-east-1)
	var region *Region
	for _, r := range hierarchy.Regions {
		region = r
		break
	}
	if region == nil {
		t.Fatal("no region found")
	}

	// Should have 1 VPC
	if len(region.VPCs) != 1 {
		t.Errorf("VPCs = %d, want 1", len(region.VPCs))
	}

	vpc := region.VPCs["vpc-1"]
	if vpc == nil {
		t.Fatal("vpc-1 not found")
	}
	if vpc.CIDR != "10.0.0.0/16" {
		t.Errorf("VPC CIDR = %q, want 10.0.0.0/16", vpc.CIDR)
	}

	// Should have 2 AZs
	if len(vpc.AvailabilityZones) != 2 {
		t.Errorf("AZs = %d, want 2", len(vpc.AvailabilityZones))
	}

	// Public subnet
	az1 := vpc.AvailabilityZones["us-east-1a"]
	if az1 == nil {
		t.Fatal("us-east-1a not found")
	}
	sub1 := az1.Subnets["subnet-1"]
	if sub1 == nil {
		t.Fatal("subnet-1 not found")
	}
	if sub1.Type != "public" {
		t.Errorf("subnet-1 type = %q, want public", sub1.Type)
	}
	if len(sub1.Resources) != 1 {
		t.Errorf("subnet-1 resources = %d, want 1", len(sub1.Resources))
	}

	// SG assigned to VPC
	if len(vpc.Resources) != 1 {
		t.Errorf("VPC resources = %d, want 1", len(vpc.Resources))
	}
}

func TestBuildHierarchy_EmptyResources(t *testing.T) {
	hb := NewHierarchyBuilder()
	hierarchy := hb.BuildHierarchy([]*terraform.Resource{})
	if len(hierarchy.Regions) != 0 {
		t.Errorf("regions = %d, want 0", len(hierarchy.Regions))
	}
}

func TestBuildHierarchy_SubnetWithoutVPC(t *testing.T) {
	resources := []*terraform.Resource{
		makeResource("aws_subnet", "orphan", map[string]interface{}{
			"id":     "subnet-orphan",
			"vpc_id": "vpc-missing",
		}),
	}
	hb := NewHierarchyBuilder()
	hierarchy := hb.BuildHierarchy(resources)
	// Subnet should be skipped since VPC doesn't exist
	for _, r := range hierarchy.Regions {
		for _, vpc := range r.VPCs {
			for _, az := range vpc.AvailabilityZones {
				if len(az.Subnets) > 0 {
					t.Error("orphan subnet should not appear")
				}
			}
		}
	}
}

func TestConvertHierarchyToNodes(t *testing.T) {
	hierarchy := &AWSHierarchy{
		Regions: map[string]*Region{
			"us-east-1": {
				ID:   "us-east-1",
				Name: "us-east-1",
				VPCs: map[string]*VPC{
					"vpc-1": {
						ID:   "vpc-1",
						Name: "main",
						CIDR: "10.0.0.0/16",
						AvailabilityZones: map[string]*AvailabilityZone{
							"us-east-1a": {
								ID:   "us-east-1a",
								Name: "us-east-1a",
								Subnets: map[string]*Subnet{
									"subnet-1": {
										ID: "subnet-1", Name: "pub",
										CIDR: "10.0.1.0/24", Type: "public",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	nodes := ConvertHierarchyToNodes(hierarchy)
	// region + VPC + AZ + subnet = 4
	if len(nodes) != 4 {
		t.Errorf("nodes = %d, want 4", len(nodes))
	}

	// Check node types
	typeCount := map[string]int{}
	for _, n := range nodes {
		typeCount[n.Data.Type]++
	}
	if typeCount["region-group"] != 1 {
		t.Errorf("region-group = %d, want 1", typeCount["region-group"])
	}
	if typeCount["vpc-group"] != 1 {
		t.Errorf("vpc-group = %d, want 1", typeCount["vpc-group"])
	}
}

// --- extractRegion ---

func TestExtractRegion(t *testing.T) {
	hb := NewHierarchyBuilder()

	tests := []struct {
		name string
		res  *terraform.Resource
		want string
	}{
		{
			"from region attr",
			&terraform.Resource{Attributes: map[string]interface{}{"region": "ap-northeast-1"}},
			"ap-northeast-1",
		},
		{
			"from ARN",
			&terraform.Resource{Attributes: map[string]interface{}{"arn": "arn:aws:ec2:eu-west-1:123456:vpc/vpc-1"}},
			"eu-west-1",
		},
		{
			"empty",
			&terraform.Resource{Attributes: map[string]interface{}{}},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hb.extractRegion(tt.res)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
