package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// RelationshipExtractor defines how to extract relationships for a specific resource type.
type RelationshipExtractor interface {
	// Extract returns relationships for the given resource.
	Extract(resource *terraform.Resource) []*Relationship
}

// RelationshipExtractorRegistry maps resource types to their extractors
type RelationshipExtractorRegistry struct {
	extractors map[string]RelationshipExtractor
}

// NewRelationshipExtractorRegistry creates a new registry with default AWS extractors
func NewRelationshipExtractorRegistry() *RelationshipExtractorRegistry {
	registry := &RelationshipExtractorRegistry{
		extractors: make(map[string]RelationshipExtractor),
	}

	// Register AWS extractors
	registry.Register("aws_subnet", &AWSSubnetExtractor{})
	registry.Register("aws_instance", &AWSInstanceExtractor{})
	registry.Register("aws_nat_gateway", &AWSNATGatewayExtractor{})
	registry.Register("aws_route_table", &AWSRouteTableExtractor{})
	registry.Register("aws_security_group", &AWSSecurityGroupExtractor{})
	registry.Register("aws_internet_gateway", &AWSInternetGatewayExtractor{})
	registry.Register("aws_eks_cluster", &AWSEKSClusterExtractor{})
	registry.Register("aws_eks_node_group", &AWSEKSNodeGroupExtractor{})
	registry.Register("aws_db_instance", &AWSDBInstanceExtractor{})
	registry.Register("aws_db_subnet_group", &AWSDBSubnetGroupExtractor{})
	registry.Register("aws_elasticache_replication_group", &AWSElastiCacheExtractor{})
	registry.Register("aws_elasticache_subnet_group", &AWSElastiCacheSubnetGroupExtractor{})
	registry.Register("aws_lb", &AWSLoadBalancerExtractor{})
	registry.Register("aws_lb_target_group", &AWSTargetGroupExtractor{})
	registry.Register("aws_lb_listener", &AWSListenerExtractor{})
	registry.Register("aws_route_table_association", &AWSRouteTableAssociationExtractor{})
	registry.Register("aws_ecs_service", &AWSECSServiceExtractor{})
	registry.Register("aws_iam_policy", &AWSPolicyExtractor{})
	registry.Register("aws_iam_role_policy", &AWSRolePolicyExtractor{})

	return registry
}

// Register registers an extractor for a resource type
func (r *RelationshipExtractorRegistry) Register(resourceType string, extractor RelationshipExtractor) {
	r.extractors[resourceType] = extractor
}

// Extract extracts relationships for a resource using the appropriate extractor
func (r *RelationshipExtractorRegistry) Extract(resource *terraform.Resource) []*Relationship {
	if extractor, ok := r.extractors[resource.Type]; ok {
		return extractor.Extract(resource)
	}
	return []*Relationship{}
}

// AWSSubnetExtractor extracts relationships for aws_subnet
type AWSSubnetExtractor struct{}

func (e *AWSSubnetExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSInstanceExtractor extracts relationships for aws_instance
type AWSInstanceExtractor struct{}

func (e *AWSInstanceExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSNATGatewayExtractor extracts relationships for aws_nat_gateway
type AWSNATGatewayExtractor struct{}

func (e *AWSNATGatewayExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSRouteTableExtractor extracts relationships for aws_route_table
type AWSRouteTableExtractor struct{}

func (e *AWSRouteTableExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSSecurityGroupExtractor extracts relationships for aws_security_group
type AWSSecurityGroupExtractor struct{}

func (e *AWSSecurityGroupExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSInternetGatewayExtractor extracts relationships for aws_internet_gateway
type AWSInternetGatewayExtractor struct{}

func (e *AWSInternetGatewayExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSEKSClusterExtractor extracts relationships for aws_eks_cluster
type AWSEKSClusterExtractor struct{}

func (e *AWSEKSClusterExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSEKSNodeGroupExtractor extracts relationships for aws_eks_node_group
type AWSEKSNodeGroupExtractor struct{}

func (e *AWSEKSNodeGroupExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSDBInstanceExtractor extracts relationships for aws_db_instance
type AWSDBInstanceExtractor struct{}

func (e *AWSDBInstanceExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSDBSubnetGroupExtractor extracts relationships for aws_db_subnet_group
type AWSDBSubnetGroupExtractor struct{}

func (e *AWSDBSubnetGroupExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSElastiCacheExtractor extracts relationships for aws_elasticache_replication_group
type AWSElastiCacheExtractor struct{}

func (e *AWSElastiCacheExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSElastiCacheSubnetGroupExtractor extracts relationships for aws_elasticache_subnet_group
type AWSElastiCacheSubnetGroupExtractor struct{}

func (e *AWSElastiCacheSubnetGroupExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSLoadBalancerExtractor extracts relationships for aws_lb
type AWSLoadBalancerExtractor struct{}

func (e *AWSLoadBalancerExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSTargetGroupExtractor extracts relationships for aws_lb_target_group
type AWSTargetGroupExtractor struct{}

func (e *AWSTargetGroupExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSListenerExtractor extracts relationships for aws_lb_listener
type AWSListenerExtractor struct{}

func (e *AWSListenerExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSRouteTableAssociationExtractor extracts relationships for aws_route_table_association
type AWSRouteTableAssociationExtractor struct{}

func (e *AWSRouteTableAssociationExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSECSServiceExtractor extracts relationships for aws_ecs_service
type AWSECSServiceExtractor struct{}

func (e *AWSECSServiceExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSPolicyExtractor extracts relationships for aws_iam_policy
type AWSPolicyExtractor struct{}

func (e *AWSPolicyExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}

// AWSRolePolicyExtractor extracts relationships for aws_iam_role_policy
type AWSRolePolicyExtractor struct{}

func (e *AWSRolePolicyExtractor) Extract(resource *terraform.Resource) []*Relationship {
	relationships := []*Relationship{}
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	if resourceID == "" {
		return relationships
	}

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

	return relationships
}
