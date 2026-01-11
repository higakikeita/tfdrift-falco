/**
 * Mock data for Drift Detection
 *
 * This file provides mock data for Drift Detection API responses
 * to enable UI development before the backend is fully implemented.
 */

import type { DriftSummary, DriftDetectionResult } from '../api/types';

/**
 * Mock Drift Summary - Scenario with detected drift
 */
export const mockDriftSummaryWithDrift: DriftSummary = {
  region: 'us-east-1',
  timestamp: new Date().toISOString(),
  counts: {
    terraform_resources: 45,
    aws_resources: 53,
    unmanaged: 8,
    missing: 3,
    modified: 5,
  },
  breakdown: {
    unmanaged_by_type: {
      'AWS::EC2::Instance': 3,
      'AWS::S3::Bucket': 2,
      'AWS::Lambda::Function': 2,
      'AWS::IAM::Role': 1,
    },
    missing_by_type: {
      'AWS::RDS::DBInstance': 2,
      'AWS::ElastiCache::ReplicationGroup': 1,
    },
    modified_by_type: {
      'AWS::EC2::SecurityGroup': 3,
      'AWS::EKS::Cluster': 1,
      'AWS::LoadBalancer': 1,
    },
  },
};

/**
 * Mock Drift Summary - Clean scenario (no drift)
 */
export const mockDriftSummaryNoDrift: DriftSummary = {
  region: 'us-east-1',
  timestamp: new Date().toISOString(),
  counts: {
    terraform_resources: 45,
    aws_resources: 45,
    unmanaged: 0,
    missing: 0,
    modified: 0,
  },
  breakdown: {
    unmanaged_by_type: {},
    missing_by_type: {},
    modified_by_type: {},
  },
};

/**
 * Mock Drift Detection Details - With drift
 */
export const mockDriftDetectionWithDrift: DriftDetectionResult = {
  region: 'us-east-1',
  timestamp: new Date().toISOString(),
  summary: {
    terraform_resources: 45,
    aws_resources: 53,
    unmanaged_count: 8,
    missing_count: 3,
    modified_count: 5,
  },
  drift: {
    unmanaged_resources: [
      {
        id: 'i-0123456789abcdef0',
        name: 'manual-web-server-1',
        type: 'AWS::EC2::Instance',
        region: 'us-east-1',
        arn: 'arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef0',
        attributes: {
          instance_type: 't3.medium',
          created_at: '2025-12-15T10:30:00Z',
        },
        tags: {
          Name: 'manual-web-server-1',
          Environment: 'production',
        },
      },
      {
        id: 'i-0123456789abcdef1',
        name: 'manual-web-server-2',
        type: 'AWS::EC2::Instance',
        region: 'us-east-1',
        arn: 'arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef1',
        attributes: {
          instance_type: 't3.medium',
          created_at: '2025-12-16T14:20:00Z',
        },
        tags: {
          Name: 'manual-web-server-2',
          Environment: 'production',
        },
      },
      {
        id: 'i-0123456789abcdef2',
        name: 'manual-worker-1',
        type: 'AWS::EC2::Instance',
        region: 'us-east-1',
        arn: 'arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef2',
        attributes: {
          instance_type: 't3.small',
          created_at: '2025-12-20T09:15:00Z',
        },
        tags: {
          Name: 'manual-worker-1',
          Environment: 'staging',
        },
      },
      {
        id: 'manual-data-bucket',
        name: 'manual-data-bucket',
        type: 'AWS::S3::Bucket',
        region: 'us-east-1',
        arn: 'arn:aws:s3:::manual-data-bucket',
        attributes: {
          versioning: true,
          created_at: '2025-11-10T08:00:00Z',
        },
        tags: {
          Purpose: 'data-lake',
        },
      },
      {
        id: 'manual-logs-bucket',
        name: 'manual-logs-bucket',
        type: 'AWS::S3::Bucket',
        region: 'us-east-1',
        arn: 'arn:aws:s3:::manual-logs-bucket',
        attributes: {
          versioning: false,
          created_at: '2025-11-25T12:30:00Z',
        },
        tags: {
          Purpose: 'logging',
        },
      },
      {
        id: 'manual-processor',
        name: 'manual-processor',
        type: 'AWS::Lambda::Function',
        region: 'us-east-1',
        arn: 'arn:aws:lambda:us-east-1:123456789012:function:manual-processor',
        attributes: {
          runtime: 'python3.11',
          memory: 512,
          created_at: '2025-12-01T16:45:00Z',
        },
        tags: {
          Service: 'data-processing',
        },
      },
      {
        id: 'manual-notifier',
        name: 'manual-notifier',
        type: 'AWS::Lambda::Function',
        region: 'us-east-1',
        arn: 'arn:aws:lambda:us-east-1:123456789012:function:manual-notifier',
        attributes: {
          runtime: 'nodejs18.x',
          memory: 256,
          created_at: '2025-12-05T11:20:00Z',
        },
        tags: {
          Service: 'notifications',
        },
      },
      {
        id: 'ManualAccessRole',
        name: 'ManualAccessRole',
        type: 'AWS::IAM::Role',
        region: 'us-east-1',
        arn: 'arn:aws:iam::123456789012:role/ManualAccessRole',
        attributes: {
          created_at: '2025-10-15T09:00:00Z',
        },
        tags: {},
      },
    ],
    missing_resources: [
      {
        type: 'aws_db_instance',
        name: 'prod-mysql',
        provider: 'aws',
        mode: 'managed',
        attributes: {
          engine: 'mysql',
          instance_class: 'db.t3.medium',
          allocated_storage: 100,
        },
      },
      {
        type: 'aws_db_instance',
        name: 'staging-postgres',
        provider: 'aws',
        mode: 'managed',
        attributes: {
          engine: 'postgres',
          instance_class: 'db.t3.small',
          allocated_storage: 50,
        },
      },
      {
        type: 'aws_elasticache_replication_group',
        name: 'session-cache',
        provider: 'aws',
        mode: 'managed',
        attributes: {
          engine: 'redis',
          node_type: 'cache.t3.micro',
          num_cache_clusters: 2,
        },
      },
    ],
    modified_resources: [
      {
        resource_id: 'sg-0a1b2c3d4e5f6g7h8',
        resource_type: 'aws_security_group',
        terraform_state: {
          id: 'sg-0a1b2c3d4e5f6g7h8',
          name: 'web-sg',
          ingress_rules: [{ from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] }],
        },
        actual_state: {
          id: 'sg-0a1b2c3d4e5f6g7h8',
          name: 'web-sg',
          ingress_rules: [
            { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
            { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
          ],
        },
        differences: [
          {
            field: 'ingress_rules',
            terraform_value: [{ from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] }],
            actual_value: [
              { from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
              { from_port: 80, to_port: 80, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] },
            ],
          },
        ],
      },
      {
        resource_id: 'sg-0a1b2c3d4e5f6g7h9',
        resource_type: 'aws_security_group',
        terraform_state: {
          id: 'sg-0a1b2c3d4e5f6g7h9',
          name: 'db-sg',
          ingress_rules: [{ from_port: 5432, to_port: 5432, protocol: 'tcp', cidr_blocks: ['10.0.0.0/16'] }],
        },
        actual_state: {
          id: 'sg-0a1b2c3d4e5f6g7h9',
          name: 'db-sg',
          ingress_rules: [
            { from_port: 5432, to_port: 5432, protocol: 'tcp', cidr_blocks: ['10.0.0.0/16', '172.31.0.0/16'] },
          ],
        },
        differences: [
          {
            field: 'ingress_rules',
            terraform_value: [{ from_port: 5432, to_port: 5432, protocol: 'tcp', cidr_blocks: ['10.0.0.0/16'] }],
            actual_value: [
              { from_port: 5432, to_port: 5432, protocol: 'tcp', cidr_blocks: ['10.0.0.0/16', '172.31.0.0/16'] },
            ],
          },
        ],
      },
      {
        resource_id: 'sg-0a1b2c3d4e5f6g7h0',
        resource_type: 'aws_security_group',
        terraform_state: {
          id: 'sg-0a1b2c3d4e5f6g7h0',
          name: 'app-sg',
          egress_rules: [{ from_port: 0, to_port: 65535, protocol: '-1', cidr_blocks: ['0.0.0.0/0'] }],
        },
        actual_state: {
          id: 'sg-0a1b2c3d4e5f6g7h0',
          name: 'app-sg',
          egress_rules: [{ from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] }],
        },
        differences: [
          {
            field: 'egress_rules',
            terraform_value: [{ from_port: 0, to_port: 65535, protocol: '-1', cidr_blocks: ['0.0.0.0/0'] }],
            actual_value: [{ from_port: 443, to_port: 443, protocol: 'tcp', cidr_blocks: ['0.0.0.0/0'] }],
          },
        ],
      },
      {
        resource_id: 'prod-eks',
        resource_type: 'aws_eks_cluster',
        terraform_state: {
          id: 'prod-eks',
          name: 'prod-eks',
          version: '1.27',
        },
        actual_state: {
          id: 'prod-eks',
          name: 'prod-eks',
          version: '1.28',
        },
        differences: [
          {
            field: 'version',
            terraform_value: '1.27',
            actual_value: '1.28',
          },
        ],
      },
      {
        resource_id: 'arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb/50dc6c495c0c9188',
        resource_type: 'aws_lb',
        terraform_state: {
          id: 'prod-alb',
          name: 'prod-alb',
          access_logs: { enabled: false },
        },
        actual_state: {
          id: 'prod-alb',
          name: 'prod-alb',
          access_logs: { enabled: true },
        },
        differences: [
          {
            field: 'access_logs.enabled',
            terraform_value: false,
            actual_value: true,
          },
        ],
      },
    ],
  },
};

/**
 * Mock Drift Detection - No drift scenario
 */
export const mockDriftDetectionNoDrift: DriftDetectionResult = {
  region: 'us-east-1',
  timestamp: new Date().toISOString(),
  summary: {
    terraform_resources: 45,
    aws_resources: 45,
    unmanaged_count: 0,
    missing_count: 0,
    modified_count: 0,
  },
  drift: {
    unmanaged_resources: [],
    missing_resources: [],
    modified_resources: [],
  },
};
