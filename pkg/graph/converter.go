package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// TerraformToGraph converts Terraform resources to a GraphDatabase
func TerraformToGraph(resources []*terraform.Resource, driftedIDs map[string]bool) *GraphDatabase {
	db := NewGraphDatabase()

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

	case "aws_eks_cluster":
		// EKS DEPENDS_ON VPC
		if vpcConfig, ok := resource.Attributes["vpc_config"].([]interface{}); ok && len(vpcConfig) > 0 {
			if vpcConfigMap, ok := vpcConfig[0].(map[string]interface{}); ok {
				// EKS → Subnets
				if subnetIDs, ok := vpcConfigMap["subnet_ids"].([]interface{}); ok {
					for _, subnetID := range subnetIDs {
						if subnetIDStr, ok := subnetID.(string); ok && subnetIDStr != "" {
							relationships = append(relationships, &Relationship{
								ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetIDStr),
								Type:      DEPENDS_ON,
								StartNode: resourceID,
								EndNode:   subnetIDStr,
								Properties: map[string]interface{}{
									"type": "network_placement",
								},
							})
						}
					}
				}
				// EKS → Security Groups
				if sgIDs, ok := vpcConfigMap["security_group_ids"].([]interface{}); ok {
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
			}
		}

	case "aws_eks_node_group":
		// Node Group PART_OF EKS Cluster
		if clusterName, ok := resource.Attributes["cluster_name"].(string); ok && clusterName != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-PART_OF-%s", resourceID, clusterName),
				Type:      PART_OF,
				StartNode: resourceID,
				EndNode:   clusterName,
				Properties: map[string]interface{}{
					"type": "cluster_membership",
				},
			})
		}
		// Node Group → Subnets
		if subnetIDs, ok := resource.Attributes["subnet_ids"].([]interface{}); ok {
			for _, subnetID := range subnetIDs {
				if subnetIDStr, ok := subnetID.(string); ok && subnetIDStr != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetIDStr),
						Type:      DEPENDS_ON,
						StartNode: resourceID,
						EndNode:   subnetIDStr,
						Properties: map[string]interface{}{
							"type": "network_placement",
						},
					})
				}
			}
		}

	case "aws_db_instance":
		// RDS → Subnet Group
		if subnetGroupName, ok := resource.Attributes["db_subnet_group_name"].(string); ok && subnetGroupName != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetGroupName),
				Type:      DEPENDS_ON,
				StartNode: resourceID,
				EndNode:   subnetGroupName,
				Properties: map[string]interface{}{
					"type": "network_placement",
				},
			})
		}
		// RDS → Security Groups
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

	case "aws_db_subnet_group":
		// DB Subnet Group → Subnets
		if subnetIDs, ok := resource.Attributes["subnet_ids"].([]interface{}); ok {
			for _, subnetID := range subnetIDs {
				if subnetIDStr, ok := subnetID.(string); ok && subnetIDStr != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-CONTAINS-%s", resourceID, subnetIDStr),
						Type:      CONTAINS,
						StartNode: resourceID,
						EndNode:   subnetIDStr,
						Properties: map[string]interface{}{
							"type": "subnet_group_membership",
						},
					})
				}
			}
		}

	case "aws_elasticache_replication_group":
		// ElastiCache → Subnet Group
		if subnetGroupName, ok := resource.Attributes["subnet_group_name"].(string); ok && subnetGroupName != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetGroupName),
				Type:      DEPENDS_ON,
				StartNode: resourceID,
				EndNode:   subnetGroupName,
				Properties: map[string]interface{}{
					"type": "network_placement",
				},
			})
		}
		// ElastiCache → Security Groups
		if sgIDs, ok := resource.Attributes["security_group_ids"].([]interface{}); ok {
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

	case "aws_elasticache_subnet_group":
		// ElastiCache Subnet Group → Subnets
		if subnetIDs, ok := resource.Attributes["subnet_ids"].([]interface{}); ok {
			for _, subnetID := range subnetIDs {
				if subnetIDStr, ok := subnetID.(string); ok && subnetIDStr != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-CONTAINS-%s", resourceID, subnetIDStr),
						Type:      CONTAINS,
						StartNode: resourceID,
						EndNode:   subnetIDStr,
						Properties: map[string]interface{}{
							"type": "subnet_group_membership",
						},
					})
				}
			}
		}

	case "aws_lb":
		// ALB → Subnets
		if subnetIDs, ok := resource.Attributes["subnets"].([]interface{}); ok {
			for _, subnetID := range subnetIDs {
				if subnetIDStr, ok := subnetID.(string); ok && subnetIDStr != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-DEPENDS_ON-%s", resourceID, subnetIDStr),
						Type:      DEPENDS_ON,
						StartNode: resourceID,
						EndNode:   subnetIDStr,
						Properties: map[string]interface{}{
							"type": "network_placement",
						},
					})
				}
			}
		}
		// ALB → Security Groups
		if sgIDs, ok := resource.Attributes["security_groups"].([]interface{}); ok {
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

	case "aws_lb_target_group":
		// Target Group → VPC
		if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-PART_OF-%s", resourceID, vpcID),
				Type:      PART_OF,
				StartNode: resourceID,
				EndNode:   vpcID,
				Properties: map[string]interface{}{
					"type": "vpc_membership",
				},
			})
		}

	case "aws_lb_listener":
		// Listener → Load Balancer
		if lbArn, ok := resource.Attributes["load_balancer_arn"].(string); ok && lbArn != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-PART_OF-%s", resourceID, lbArn),
				Type:      PART_OF,
				StartNode: resourceID,
				EndNode:   lbArn,
				Properties: map[string]interface{}{
					"type": "listener_attachment",
				},
			})
		}
		// Listener → Target Group
		if tgArn, ok := resource.Attributes["default_action"].([]interface{}); ok && len(tgArn) > 0 {
			if action, ok := tgArn[0].(map[string]interface{}); ok {
				if targetGroupArn, ok := action["target_group_arn"].(string); ok && targetGroupArn != "" {
					relationships = append(relationships, &Relationship{
						ID:        fmt.Sprintf("%s-ROUTES_TO-%s", resourceID, targetGroupArn),
						Type:      ROUTES_TO,
						StartNode: resourceID,
						EndNode:   targetGroupArn,
						Properties: map[string]interface{}{
							"type": "traffic_routing",
						},
					})
				}
			}
		}

	case "aws_route_table_association":
		// Route Table Association → Route Table
		if rtID, ok := resource.Attributes["route_table_id"].(string); ok && rtID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-ASSOCIATES-%s", resourceID, rtID),
				Type:      ASSOCIATES,
				StartNode: resourceID,
				EndNode:   rtID,
				Properties: map[string]interface{}{
					"type": "routing",
				},
			})
		}
		// Route Table Association → Subnet
		if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-ASSOCIATES-%s", resourceID, subnetID),
				Type:      ASSOCIATES,
				StartNode: resourceID,
				EndNode:   subnetID,
				Properties: map[string]interface{}{
					"type": "subnet_routing",
				},
			})
		}

	case "aws_ecs_cluster":
		// ECS Cluster (base resource, no dependencies typically)

	case "aws_ecs_service":
		// ECS Service → Cluster
		if clusterArn, ok := resource.Attributes["cluster"].(string); ok && clusterArn != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-RUNS_IN-%s", resourceID, clusterArn),
				Type:      RUNS_IN,
				StartNode: resourceID,
				EndNode:   clusterArn,
				Properties: map[string]interface{}{
					"type": "service_placement",
				},
			})
		}
		// ECS Service → Target Group
		if loadBalancers, ok := resource.Attributes["load_balancer"].([]interface{}); ok {
			for _, lb := range loadBalancers {
				if lbMap, ok := lb.(map[string]interface{}); ok {
					if tgArn, ok := lbMap["target_group_arn"].(string); ok && tgArn != "" {
						relationships = append(relationships, &Relationship{
							ID:        fmt.Sprintf("%s-REGISTERS_TO-%s", resourceID, tgArn),
							Type:      REGISTERS_TO,
							StartNode: resourceID,
							EndNode:   tgArn,
							Properties: map[string]interface{}{
								"type": "load_balancing",
							},
						})
					}
				}
			}
		}

	case "aws_s3_bucket":
		// S3 buckets typically don't have explicit dependencies in attributes
		// but may be referenced by other resources

	case "aws_iam_role":
		// IAM roles are referenced by other resources but don't have dependencies in attributes

	case "aws_iam_policy", "aws_iam_role_policy":
		// Policy → Role
		if roleArn, ok := resource.Attributes["role"].(string); ok && roleArn != "" {
			relationships = append(relationships, &Relationship{
				ID:        fmt.Sprintf("%s-APPLIES_TO-%s", resourceID, roleArn),
				Type:      APPLIES_TO,
				StartNode: resourceID,
				EndNode:   roleArn,
				Properties: map[string]interface{}{
					"type": "policy_attachment",
				},
			})
		}
	}

	// Generic fallback: Extract common reference patterns
	relationships = append(relationships, extractGenericDependencies(resource, resourceID)...)

	return relationships
}

// extractGenericDependencies extracts dependencies using common AWS attribute patterns
func extractGenericDependencies(resource *terraform.Resource, resourceID string) []*Relationship {
	relationships := []*Relationship{}

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
