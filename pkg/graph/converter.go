package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// TerraformToGraph converts Terraform resources to a GraphDatabase
func TerraformToGraph(resources []*terraform.Resource, driftedIDs map[string]bool) *GraphDatabase {
	db := NewGraphDatabase()

	log.Info("Converting Terraform resources to graph database...")

	// First pass: Create nodes for all resources
	for _, resource := range resources {
		node := resourceToNode(resource, driftedIDs)
		if node != nil {
			db.AddNode(node)
		}
	}

	// Second pass: Create relationships
	for _, resource := range resources {
		relationships := extractRelationships(resource)
		for _, rel := range relationships {
			if err := db.AddRelationship(rel); err != nil {
				log.Debugf("Skipping relationship %s: %v", rel.ID, err)
			}
		}
	}

	log.Infof("Graph database created: %d nodes, %d relationships",
		db.NodeCount(), db.RelationshipCount())

	return db
}

// resourceToNode converts a Terraform resource to a graph Node
func resourceToNode(resource *terraform.Resource, driftedIDs map[string]bool) *Node {
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return nil
	}

	// Determine labels
	labels := []string{"Resource"}

	// Add resource type as label
	switch resource.Type {
	case "aws_vpc":
		labels = append(labels, "VPC", "Network")
	case "aws_subnet":
		labels = append(labels, "Subnet", "Network")
	case "aws_instance":
		labels = append(labels, "EC2", "Compute")
	case "aws_db_instance":
		labels = append(labels, "RDS", "Database")
	case "aws_nat_gateway":
		labels = append(labels, "NATGateway", "Network")
	case "aws_internet_gateway":
		labels = append(labels, "InternetGateway", "Network")
	case "aws_route_table":
		labels = append(labels, "RouteTable", "Network")
	case "aws_security_group":
		labels = append(labels, "SecurityGroup", "Security")
	default:
		labels = append(labels, resource.Type)
	}

	// Add drift label if applicable
	if driftedIDs[resourceID] {
		labels = append(labels, "Drifted")
	}

	// Build properties
	properties := make(map[string]interface{})
	properties["id"] = resourceID
	properties["type"] = resource.Type
	properties["name"] = extractResourceName(resource)
	properties["has_drift"] = driftedIDs[resourceID]
	properties["mode"] = resource.Mode
	properties["provider"] = resource.Provider

	// Add resource-specific properties
	addResourceSpecificProperties(resource, properties)

	return &Node{
		ID:         resourceID,
		Labels:     labels,
		Properties: properties,
	}
}

// addResourceSpecificProperties adds type-specific properties
func addResourceSpecificProperties(resource *terraform.Resource, properties map[string]interface{}) {
	switch resource.Type {
	case "aws_vpc":
		if cidr, ok := resource.Attributes["cidr_block"].(string); ok {
			properties["cidr"] = cidr
		}

	case "aws_subnet":
		if cidr, ok := resource.Attributes["cidr_block"].(string); ok {
			properties["cidr"] = cidr
		}
		if az, ok := resource.Attributes["availability_zone"].(string); ok {
			properties["availability_zone"] = az
		}
		if mapPublicIP, ok := resource.Attributes["map_public_ip_on_launch"].(bool); ok {
			properties["public"] = mapPublicIP
		}

	case "aws_instance":
		if instanceType, ok := resource.Attributes["instance_type"].(string); ok {
			properties["instance_type"] = instanceType
		}
		if state, ok := resource.Attributes["instance_state"].(string); ok {
			properties["state"] = state
		}
		if privateIP, ok := resource.Attributes["private_ip"].(string); ok {
			properties["private_ip"] = privateIP
		}
		if publicIP, ok := resource.Attributes["public_ip"].(string); ok {
			properties["public_ip"] = publicIP
		}

	case "aws_db_instance":
		if engine, ok := resource.Attributes["engine"].(string); ok {
			properties["engine"] = engine
		}
		if instanceClass, ok := resource.Attributes["instance_class"].(string); ok {
			properties["instance_class"] = instanceClass
		}
		if status, ok := resource.Attributes["status"].(string); ok {
			properties["status"] = status
		}

	case "aws_security_group":
		if name, ok := resource.Attributes["name"].(string); ok {
			properties["sg_name"] = name
		}
		if description, ok := resource.Attributes["description"].(string); ok {
			properties["description"] = description
		}
	}

	// Add tags if present
	if tags, ok := resource.Attributes["tags"].(map[string]interface{}); ok {
		properties["tags"] = tags
	}
}

// extractRelationships extracts relationships from a resource
func extractRelationships(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

	switch resource.Type {
	case "aws_subnet":
		// Subnet PART_OF VPC
		if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-PART_OF-%s", resourceID, vpcID),
				Type:      PART_OF,
				StartNode: resourceID,
				EndNode:   vpcID,
				Properties: map[string]interface{}{
					"type": "hierarchical",
				},
			})
		}

	case "aws_instance":
		// EC2 DEPENDS_ON Subnet
		if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetID),
				Type:      DEPENDS_ON,
				StartNode: resourceID,
				EndNode:   subnetID,
				Properties: map[string]interface{}{
					"type": "network_placement",
				},
			})
		}

		// EC2 SECURES SecurityGroup
		if sgIDs, ok := resource.Attributes["vpc_security_group_ids"].([]interface{}); ok {
			for _, sgID := range sgIDs {
				if sgIDStr, ok := sgID.(string); ok && sgIDStr != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-SECURES-%s", sgIDStr, resourceID),
						Type:      SECURES,
						StartNode: sgIDStr,
						EndNode:   resourceID,
						Properties: map[string]interface{}{
							"type": "security",
						},
					})
				}
			}
		}

	case "aws_nat_gateway":
		// NAT Gateway DEPENDS_ON Subnet
		if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetID),
				Type:      DEPENDS_ON,
				StartNode: resourceID,
				EndNode:   subnetID,
				Properties: map[string]interface{}{
					"type": "network_placement",
				},
			})
		}

	case "aws_route_table":
		// Route table ROUTES_TO VPC
		if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-ROUTES_TO-%s", resourceID, vpcID),
				Type:      ROUTES_TO,
				StartNode: resourceID,
				EndNode:   vpcID,
				Properties: map[string]interface{}{
					"type": "network_routing",
				},
			})
		}

	case "aws_security_group":
		// Security Group SECURES VPC
		if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-SECURES-%s", resourceID, vpcID),
				Type:      SECURES,
				StartNode: resourceID,
				EndNode:   vpcID,
				Properties: map[string]interface{}{
					"type": "security",
				},
			})
		}

	case "aws_internet_gateway":
		// IGW CONNECTS_TO VPC
		if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-CONNECTS_TO-%s", resourceID, vpcID),
				Type:      CONNECTS_TO,
				StartNode: resourceID,
				EndNode:   vpcID,
				Properties: map[string]interface{}{
					"type": "network_gateway",
				},
			})
		}
	}

	return relationships
}
