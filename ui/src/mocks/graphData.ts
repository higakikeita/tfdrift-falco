/**
 * Mock Graph Data for Storybook
 *
 * CytoscapeGraph Stories用のモックデータ
 */

import type { CytoscapeElements } from '../types/graph';

// ==================== Small Graph (10 nodes) ====================

export const mockSmallGraph: CytoscapeElements = {
  nodes: [
    // VPC
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'prod-vpc' }},

    // Subnets
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'subnet-a', parent: 'vpc-1' }},
    { data: { id: 'subnet-2', resource_type: 'aws_subnet', label: 'subnet-b', parent: 'vpc-1' }},

    // Resources in subnet-1
    { data: { id: 'eks-1', resource_type: 'aws_eks_cluster', label: 'prod-eks', parent: 'subnet-1' }},
    { data: { id: 'rds-1', resource_type: 'aws_db_instance', label: 'prod-db', parent: 'subnet-1' }},
    { data: { id: 'sg-1', resource_type: 'aws_security_group', label: 'eks-sg', parent: 'subnet-1' }},

    // Resources in subnet-2
    { data: { id: 'lb-1', resource_type: 'aws_lb', label: 'prod-alb', parent: 'subnet-2' }},
    { data: { id: 'igw-1', resource_type: 'aws_internet_gateway', label: 'prod-igw', parent: 'subnet-2' }},

    // Other resources
    { data: { id: 'iam-1', resource_type: 'aws_iam_role', label: 'eks-role' }},
    { data: { id: 'kms-1', resource_type: 'aws_kms_key', label: 'prod-key' }},
  ],
  edges: [
    { data: { id: 'e1', source: 'eks-1', target: 'sg-1', label: 'uses' }},
    { data: { id: 'e2', source: 'eks-1', target: 'iam-1', label: 'assumes' }},
    { data: { id: 'e3', source: 'rds-1', target: 'kms-1', label: 'encrypted_by' }},
    { data: { id: 'e4', source: 'lb-1', target: 'eks-1', label: 'forwards_to' }},
    { data: { id: 'e5', source: 'lb-1', target: 'igw-1', label: 'routes_through' }},
  ]
};

// ==================== Medium Graph (30 nodes) ====================

export const mockMediumGraph: CytoscapeElements = {
  nodes: [
    // VPC
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'prod-vpc' }},

    // 3 Subnets
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'private-a', parent: 'vpc-1' }},
    { data: { id: 'subnet-2', resource_type: 'aws_subnet', label: 'private-b', parent: 'vpc-1' }},
    { data: { id: 'subnet-3', resource_type: 'aws_subnet', label: 'public-a', parent: 'vpc-1' }},

    // Subnet 1 resources (private-a)
    { data: { id: 'eks-1', resource_type: 'aws_eks_cluster', label: 'prod-eks', parent: 'subnet-1' }},
    { data: { id: 'node-group-1', resource_type: 'aws_eks_node_group', label: 'general', parent: 'subnet-1' }},
    { data: { id: 'addon-1', resource_type: 'aws_eks_addon', label: 'coredns', parent: 'subnet-1' }},
    { data: { id: 'sg-1', resource_type: 'aws_security_group', label: 'eks-sg', parent: 'subnet-1' }},
    { data: { id: 'rds-1', resource_type: 'aws_db_instance', label: 'prod-db', parent: 'subnet-1' }},
    { data: { id: 'sg-2', resource_type: 'aws_security_group', label: 'db-sg', parent: 'subnet-1' }},

    // Subnet 2 resources (private-b)
    { data: { id: 'elasticache-1', resource_type: 'aws_elasticache_replication_group', label: 'prod-redis', parent: 'subnet-2' }},
    { data: { id: 'sg-3', resource_type: 'aws_security_group', label: 'cache-sg', parent: 'subnet-2' }},
    { data: { id: 'ecs-1', resource_type: 'aws_ecs_cluster', label: 'batch-ecs', parent: 'subnet-2' }},

    // Subnet 3 resources (public-a)
    { data: { id: 'lb-1', resource_type: 'aws_lb', label: 'prod-alb', parent: 'subnet-3' }},
    { data: { id: 'igw-1', resource_type: 'aws_internet_gateway', label: 'prod-igw', parent: 'subnet-3' }},
    { data: { id: 'nat-1', resource_type: 'aws_nat_gateway', label: 'nat-gw', parent: 'subnet-3' }},

    // VPC-level resources (no parent)
    { data: { id: 'rtb-1', resource_type: 'aws_route_table', label: 'private-rtb' }},
    { data: { id: 'rtb-2', resource_type: 'aws_route_table', label: 'public-rtb' }},
    { data: { id: 'route-1', resource_type: 'aws_route', label: 'to-nat' }},
    { data: { id: 'route-2', resource_type: 'aws_route', label: 'to-igw' }},

    // IAM resources
    { data: { id: 'iam-1', resource_type: 'aws_iam_role', label: 'eks-role' }},
    { data: { id: 'iam-2', resource_type: 'aws_iam_role', label: 'ecs-role' }},
    { data: { id: 'iam-3', resource_type: 'aws_iam_policy', label: 's3-policy' }},

    // Security resources
    { data: { id: 'kms-1', resource_type: 'aws_kms_key', label: 'prod-key' }},
    { data: { id: 'kms-2', resource_type: 'aws_kms_alias', label: 'alias/prod' }},
    { data: { id: 'secrets-1', resource_type: 'aws_secretsmanager_secret', label: 'db-creds' }},

    // Monitoring
    { data: { id: 'cw-1', resource_type: 'aws_cloudwatch_log_group', label: '/aws/eks/prod' }},
    { data: { id: 'cw-2', resource_type: 'aws_cloudwatch_log_group', label: '/aws/ecs/batch' }},
  ],
  edges: [
    // EKS connections
    { data: { id: 'e1', source: 'eks-1', target: 'node-group-1', label: 'manages' }},
    { data: { id: 'e2', source: 'eks-1', target: 'addon-1', label: 'has_addon' }},
    { data: { id: 'e3', source: 'eks-1', target: 'sg-1', label: 'uses' }},
    { data: { id: 'e4', source: 'eks-1', target: 'iam-1', label: 'assumes' }},
    { data: { id: 'e5', source: 'eks-1', target: 'cw-1', label: 'logs_to' }},

    // RDS connections
    { data: { id: 'e6', source: 'rds-1', target: 'sg-2', label: 'uses' }},
    { data: { id: 'e7', source: 'rds-1', target: 'kms-1', label: 'encrypted_by' }},
    { data: { id: 'e8', source: 'rds-1', target: 'secrets-1', label: 'credentials' }},

    // ElastiCache connections
    { data: { id: 'e9', source: 'elasticache-1', target: 'sg-3', label: 'uses' }},

    // ECS connections
    { data: { id: 'e10', source: 'ecs-1', target: 'iam-2', label: 'assumes' }},
    { data: { id: 'e11', source: 'ecs-1', target: 'cw-2', label: 'logs_to' }},

    // Load Balancer connections
    { data: { id: 'e12', source: 'lb-1', target: 'eks-1', label: 'forwards_to' }},
    { data: { id: 'e13', source: 'lb-1', target: 'igw-1', label: 'routes_through' }},

    // Route Table connections
    { data: { id: 'e14', source: 'rtb-1', target: 'nat-1', label: 'routes_to' }},
    { data: { id: 'e15', source: 'rtb-2', target: 'igw-1', label: 'routes_to' }},
    { data: { id: 'e16', source: 'route-1', target: 'nat-1', label: 'target' }},
    { data: { id: 'e17', source: 'route-2', target: 'igw-1', label: 'target' }},

    // IAM connections
    { data: { id: 'e18', source: 'iam-1', target: 'iam-3', label: 'attached' }},
    { data: { id: 'e19', source: 'iam-2', target: 'iam-3', label: 'attached' }},

    // KMS connections
    { data: { id: 'e20', source: 'kms-2', target: 'kms-1', label: 'alias_of' }},
  ]
};

// ==================== With Drift ====================

export const mockGraphWithDrift: CytoscapeElements = {
  nodes: [
    // VPC
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'prod-vpc' }},

    // Subnets
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'subnet-a', parent: 'vpc-1' }},

    // Resources with drift
    {
      data: {
        id: 'eks-1',
        resource_type: 'aws_eks_cluster',
        label: 'prod-eks',
        parent: 'subnet-1',
        severity: 'high'  // Modified
      }
    },
    {
      data: {
        id: 'rds-1',
        resource_type: 'aws_db_instance',
        label: 'prod-db',
        parent: 'subnet-1',
        severity: 'critical'  // Missing
      }
    },
    {
      data: {
        id: 'sg-1',
        resource_type: 'aws_security_group',
        label: 'eks-sg',
        parent: 'subnet-1',
        severity: 'medium'  // Unmanaged
      }
    },
  ],
  edges: [
    { data: { id: 'e1', source: 'eks-1', target: 'sg-1', label: 'uses' }},
    { data: { id: 'e2', source: 'rds-1', target: 'sg-1', label: 'uses' }},
  ]
};

// ==================== All Resource Types ====================

export const mockAllResourceTypes: CytoscapeElements = {
  nodes: [
    // Compute
    { data: { id: 'eks-1', resource_type: 'aws_eks_cluster', label: 'EKS Cluster' }},
    { data: { id: 'eks-ng-1', resource_type: 'aws_eks_node_group', label: 'Node Group' }},
    { data: { id: 'eks-addon-1', resource_type: 'aws_eks_addon', label: 'CoreDNS' }},
    { data: { id: 'ecs-1', resource_type: 'aws_ecs_cluster', label: 'ECS Cluster' }},

    // Database
    { data: { id: 'rds-1', resource_type: 'aws_db_instance', label: 'RDS Instance' }},
    { data: { id: 'cache-1', resource_type: 'aws_elasticache_replication_group', label: 'ElastiCache' }},

    // Network
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'VPC' }},
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'Subnet' }},
    { data: { id: 'lb-1', resource_type: 'aws_lb', label: 'Load Balancer' }},
    { data: { id: 'igw-1', resource_type: 'aws_internet_gateway', label: 'Internet Gateway' }},
    { data: { id: 'nat-1', resource_type: 'aws_nat_gateway', label: 'NAT Gateway' }},
    { data: { id: 'rtb-1', resource_type: 'aws_route_table', label: 'Route Table' }},
    { data: { id: 'route-1', resource_type: 'aws_route', label: 'Route' }},
    { data: { id: 'sg-1', resource_type: 'aws_security_group', label: 'Security Group' }},

    // Security
    { data: { id: 'iam-1', resource_type: 'aws_iam_role', label: 'IAM Role' }},
    { data: { id: 'iam-pol-1', resource_type: 'aws_iam_policy', label: 'IAM Policy' }},
    { data: { id: 'kms-1', resource_type: 'aws_kms_key', label: 'KMS Key' }},
    { data: { id: 'kms-alias-1', resource_type: 'aws_kms_alias', label: 'KMS Alias' }},
    { data: { id: 'secrets-1', resource_type: 'aws_secretsmanager_secret', label: 'Secrets Manager' }},

    // Monitoring
    { data: { id: 'cw-1', resource_type: 'aws_cloudwatch_log_group', label: 'CloudWatch Logs' }},
  ],
  edges: []
};

// ==================== Empty State ====================

export const mockEmptyGraph: CytoscapeElements = {
  nodes: [],
  edges: []
};

// ==================== Helper: Generate Large Graph ====================

export function generateLargeGraph(nodeCount: number): CytoscapeElements {
  const nodes: CytoscapeElements['nodes'] = [];
  const edges: CytoscapeElements['edges'] = [];

  // Add VPC
  nodes.push({ data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'main-vpc' }});

  // Add Subnets
  const subnetCount = Math.min(10, Math.floor(nodeCount / 10));
  for (let i = 1; i <= subnetCount; i++) {
    nodes.push({
      data: {
        id: `subnet-${i}`,
        resource_type: 'aws_subnet',
        label: `subnet-${i}`,
        parent: 'vpc-1'
      }
    });
  }

  // Add resources
  const resourceTypes = [
    'aws_eks_cluster',
    'aws_ecs_cluster',
    'aws_db_instance',
    'aws_elasticache_replication_group',
    'aws_lb',
    'aws_security_group',
    'aws_iam_role',
    'aws_kms_key'
  ];

  for (let i = 1; i <= nodeCount - subnetCount - 1; i++) {
    const resourceType = resourceTypes[i % resourceTypes.length];
    const subnetId = `subnet-${(i % subnetCount) + 1}`;

    nodes.push({
      data: {
        id: `resource-${i}`,
        resource_type: resourceType,
        label: `resource-${i}`,
        parent: subnetId
      }
    });

    // Add some edges
    if (i > 1 && Math.random() > 0.5) {
      edges.push({
        data: {
          id: `edge-${i}`,
          source: `resource-${i}`,
          target: `resource-${Math.floor(Math.random() * (i - 1)) + 1}`,
          label: 'depends_on'
        }
      });
    }
  }

  return { nodes, edges };
}
