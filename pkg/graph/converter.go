package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// TerraformToGraph converts Terraform resources to a Database
func TerraformToGraph(resources []*terraform.Resource, driftedIDs map[string]bool) *Database {
	db := NewDatabase()

	log.Infof("Converting Terraform resources to graph database... (input: %d resources)", len(resources))

	// First pass: Create nodes for all resources
	for i, resource := range resources {
		log.Debugf("Processing resource %d/%d: %s.%s", i+1, len(resources), resource.Type, resource.Name)
		node := resourceToNode(resource, driftedIDs)
		if node != nil {
			log.Debugf("Created node with ID: %s", node.ID)
			db.AddNode(node)
		} else {
			log.Warnf("resourceToNode returned nil for %s.%s", resource.Type, resource.Name)
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
		// Fallback: use type + name as ID if no ID found
		resourceID = resource.Type + "." + resource.Name
		log.Debugf("No ID found for resource %s.%s, using fallback ID: %s", resource.Type, resource.Name, resourceID)
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

// extractRelationships extracts relationships from a resource using the registry pattern
func extractRelationships(resource *terraform.Resource) []*Relationship {
	registry := NewRelationshipExtractorRegistry()
	relationships := registry.Extract(resource)

	// Generic fallback: Extract common reference patterns
	relationships = append(relationships, extractGenericDependencies(resource)...)

	return relationships
}

// extractGenericDependencies extracts dependencies using common AWS attribute patterns
func extractGenericDependencies(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

	// Common VPC reference
	if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" && resource.Type != "aws_vpc" {
		relationships = append(relationships, &Relationship{
			ID:        fmt.Sprintf("%s-GENERIC_DEPENDS_ON-%s", resourceID, vpcID),
			Type:      DEPENDS_ON,
			StartNode: resourceID,
			EndNode:   vpcID,
			Properties: map[string]interface{}{
				"type": "vpc_dependency",
			},
		})
	}

	return relationships
}
