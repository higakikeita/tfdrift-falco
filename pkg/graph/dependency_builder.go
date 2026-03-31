package graph

import (
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// DependencyBuilder defines how to extract dependency edges for resources
type DependencyBuilder interface {
	// BuildEdges builds dependency edges from resources
	BuildEdges(resources []*terraform.Resource) []models.CytoscapeEdge
}

// AWSDependencyBuilder builds dependency edges for AWS resources
type AWSDependencyBuilder struct{}

// NewAWSDependencyBuilder creates a new AWS dependency builder
func NewAWSDependencyBuilder() *AWSDependencyBuilder {
	return &AWSDependencyBuilder{}
}

// BuildEdges builds dependency edges from AWS resources
func (b *AWSDependencyBuilder) BuildEdges(resources []*terraform.Resource) []models.CytoscapeEdge {
	edges := []models.CytoscapeEdge{}
	resourceMap := make(map[string]*terraform.Resource)

	// Index resources by ID
	for _, resource := range resources {
		resourceID := extractResourceIDFromAttributes(resource.Attributes)
		if resourceID != "" {
			resourceMap[resourceID] = resource
		}
	}

	// Extract dependencies
	for _, resource := range resources {
		resourceID := extractResourceIDFromAttributes(resource.Attributes)
		if resourceID == "" {
			continue
		}

		switch resource.Type {
		case "aws_instance":
			// EC2 depends on Subnet
			if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + subnetID,
							Source: resourceID,
							Target: subnetID,
							Label:  "in subnet",
						},
					})
				}
			}
			// EC2 depends on Security Groups
			if sgIDs, ok := resource.Attributes["vpc_security_group_ids"].([]interface{}); ok {
				for _, sgID := range sgIDs {
					if sgIDStr, ok := sgID.(string); ok && sgIDStr != "" {
						if _, exists := resourceMap[sgIDStr]; exists {
							edges = append(edges, models.CytoscapeEdge{
								Data: models.EdgeData{
									ID:     resourceID + "->" + sgIDStr,
									Source: resourceID,
									Target: sgIDStr,
									Label:  "uses",
								},
							})
						}
					}
				}
			}

		case "aws_subnet":
			// Subnet depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "in vpc",
						},
					})
				}
			}

		case "aws_nat_gateway":
			// NAT Gateway depends on Subnet
			if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + subnetID,
							Source: resourceID,
							Target: subnetID,
							Label:  "in subnet",
						},
					})
				}
			}

		case "aws_route_table":
			// Route table depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "routes for",
						},
					})
				}
			}

		case "aws_security_group":
			// Security Group depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "secures",
						},
					})
				}
			}

		case "aws_internet_gateway":
			// IGW depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "gateway for",
						},
					})
				}
			}
		}
	}

	return edges
}
