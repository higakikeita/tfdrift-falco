package graph

import (
	"fmt"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// HierarchyBuilder is the generic hierarchy builder interface
// It delegates to provider-specific implementations like AWSHierarchyBuilder
type HierarchyBuilder struct {
	awsBuilder *AWSHierarchyBuilder
	// Expose hierarchy and resources for backward compatibility
	hierarchy *AWSHierarchy
	resources map[string]*terraform.Resource
}

// NewHierarchyBuilder creates a new generic hierarchy builder
func NewHierarchyBuilder() *HierarchyBuilder {
	return &HierarchyBuilder{
		awsBuilder: NewAWSHierarchyBuilder(),
		hierarchy: &AWSHierarchy{
			Regions: make(map[string]*Region),
		},
		resources: make(map[string]*terraform.Resource),
	}
}

// BuildHierarchy builds hierarchical structure from Terraform resources
// Currently delegates to AWS-specific builder
func (hb *HierarchyBuilder) BuildHierarchy(resources []*terraform.Resource) *AWSHierarchy {
	log.Info("Building hierarchy structure...")
	hb.hierarchy = hb.awsBuilder.Build(resources)
	// Also update resources map for backward compatibility
	for _, res := range resources {
		id := extractResourceIDFromAttributes(res.Attributes)
		if id != "" {
			hb.resources[id] = res
		}
	}
	return hb.hierarchy
}

// extractRegion extracts region from resource (backward compatibility)
func (hb *HierarchyBuilder) extractRegion(res *terraform.Resource) string {
	// Use the AWS builder's method
	builder := NewAWSHierarchyBuilder()
	return builder.extractRegion(res)
}

// ConvertHierarchyToNodes converts hierarchy to React Flow nodes
func ConvertHierarchyToNodes(hierarchy *AWSHierarchy) []models.CytoscapeNode {
	nodes := make([]models.CytoscapeNode, 0)

	for regionID, region := range hierarchy.Regions {
		// Create region node
		regionNode := models.CytoscapeNode{
			Data: models.NodeData{
				ID:           fmt.Sprintf("region-%s", regionID),
				Label:        region.Name,
				Type:         "region-group",
				ResourceType: "aws_region",
				ResourceName: region.Name,
				Severity:     "low",
				Metadata: map[string]interface{}{
					"level": "region",
				},
			},
		}
		nodes = append(nodes, regionNode)

		// Create VPC nodes
		for vpcID, vpc := range region.VPCs {
			vpcNode := models.CytoscapeNode{
				Data: models.NodeData{
					ID:           vpcID,
					Label:        fmt.Sprintf("%s\n(%s)", vpc.Name, vpc.CIDR),
					Type:         "vpc-group",
					ResourceType: "aws_vpc",
					ResourceName: vpc.Name,
					Severity:     "low",
					Metadata: map[string]interface{}{
						"level":       "vpc",
						"cidr":        vpc.CIDR,
						"parent_node": fmt.Sprintf("region-%s", regionID),
					},
				},
			}
			nodes = append(nodes, vpcNode)

			// Create AZ and Subnet nodes
			for azID, az := range vpc.AvailabilityZones {
				azNode := models.CytoscapeNode{
					Data: models.NodeData{
						ID:           azID,
						Label:        az.Name,
						Type:         "az-group",
						ResourceType: "aws_availability_zone",
						ResourceName: az.Name,
						Severity:     "low",
						Metadata: map[string]interface{}{
							"level":       "az",
							"parent_node": vpcID,
						},
					},
				}
				nodes = append(nodes, azNode)

				for subnetID, subnet := range az.Subnets {
					subnetNode := models.CytoscapeNode{
						Data: models.NodeData{
							ID:           subnetID,
							Label:        fmt.Sprintf("%s\n(%s)", subnet.Name, subnet.CIDR),
							Type:         fmt.Sprintf("subnet-group-%s", subnet.Type),
							ResourceType: "aws_subnet",
							ResourceName: subnet.Name,
							Severity:     "low",
							Metadata: map[string]interface{}{
								"level":       "subnet",
								"cidr":        subnet.CIDR,
								"subnet_type": subnet.Type,
								"parent_node": azID,
							},
						},
					}
					nodes = append(nodes, subnetNode)
				}
			}
		}
	}

	return nodes
}
