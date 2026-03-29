import type { Meta, StoryObj } from '@storybook/react';
import { NodeTooltip } from './NodeTooltip';

const meta: Meta<typeof NodeTooltip> = {
  title: 'Graph/NodeTooltip',
  component: NodeTooltip,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0f172a' },
        { name: 'light', value: '#ffffff' },
      ],
    },
  },
  decorators: [
    (Story) => (
      <div style={{ backgroundColor: '#0f172a', minHeight: '100vh', padding: '100px' }}>
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof NodeTooltip>;

export const BasicTooltip: Story = {
  args: {
    data: {
      id: 'node-s3-001',
      label: 'my-bucket',
      type: 'aws_s3_bucket',
      resourceType: 's3:bucket',
      resourceName: 'my-bucket-prod',
      severity: 'low',
      metadata: {
        provider: 'aws',
        mode: 'managed',
      },
    },
    position: { x: 100, y: 100 },
  },
};

export const WithDriftMetadata: Story = {
  args: {
    data: {
      id: 'node-rds-001',
      label: 'production-db',
      type: 'aws_db_instance',
      resourceType: 'rds:instance',
      resourceName: 'prod-postgres-db',
      severity: 'high',
      metadata: {
        provider: 'aws',
        has_drift: true,
        drift_count: 3,
        mode: 'managed',
      },
    },
    position: { x: 200, y: 150 },
  },
};

export const WithFullMetadata: Story = {
  args: {
    data: {
      id: 'node-lambda-001',
      label: 'api-handler',
      type: 'aws_lambda_function',
      resourceType: 'lambda:function',
      resourceName: 'api-request-handler',
      severity: 'critical',
      metadata: {
        provider: 'aws',
        mode: 'managed',
        tf_name: 'aws_lambda_function.api_handler',
        has_drift: true,
        drift_count: 2,
        last_modified: '2024-02-20T14:30:00Z',
        user: 'ci-pipeline@example.com',
      },
    },
    position: { x: 300, y: 200 },
  },
};

export const LongContent: Story = {
  args: {
    data: {
      id: 'node-vpc-001',
      label: 'main-vpc-network-production',
      type: 'aws_vpc',
      resourceType: 'vpc',
      resourceName: 'main-production-vpc-us-east-1-with-multiple-subnets-and-routing-tables',
      severity: 'medium',
      metadata: {
        provider: 'aws',
        mode: 'managed',
        tf_name: 'aws_vpc.main_production',
        last_modified: '2024-01-15T09:45:00Z',
        user: 'terraform-automation@company.com',
      },
    },
    position: { x: 150, y: 250 },
  },
};

export const CriticalSeverity: Story = {
  args: {
    data: {
      id: 'node-sg-001',
      label: 'unrestricted-security-group',
      type: 'aws_security_group',
      resourceType: 'security_group',
      resourceName: 'allow-all-traffic',
      severity: 'critical',
      metadata: {
        provider: 'aws',
        has_drift: true,
        drift_count: 5,
        last_modified: '2024-02-21T10:00:00Z',
        user: 'manual-admin',
      },
    },
    position: { x: 250, y: 300 },
  },
};

export const MediumSeverity: Story = {
  args: {
    data: {
      id: 'node-ec2-001',
      label: 'web-server-instance',
      type: 'aws_instance',
      resourceType: 'ec2:instance',
      resourceName: 'production-web-server-az-1a',
      severity: 'medium',
      metadata: {
        provider: 'aws',
        mode: 'managed',
        has_drift: true,
        drift_count: 1,
        last_modified: '2024-02-19T16:20:00Z',
        user: 'ops-team@company.com',
      },
    },
    position: { x: 180, y: 350 },
  },
};

export const LowSeverityHealthy: Story = {
  args: {
    data: {
      id: 'node-iam-001',
      label: 'lambda-execution-role',
      type: 'aws_iam_role',
      resourceType: 'iam:role',
      resourceName: 'lambda-execution-role-prod',
      severity: 'low',
      metadata: {
        provider: 'aws',
        mode: 'managed',
        tf_name: 'aws_iam_role.lambda_exec',
      },
    },
    position: { x: 220, y: 180 },
  },
};

export const MinimalMetadata: Story = {
  args: {
    data: {
      id: 'node-minimal-001',
      label: 'resource',
      type: 'resource_type',
      resourceType: 'type:resource',
      resourceName: 'resource-name',
      severity: 'low',
    },
    position: { x: 100, y: 100 },
  },
};

export const WithoutResourceName: Story = {
  args: {
    data: {
      id: 'node-no-name-001',
      label: 'node-label-only',
      type: 'aws_s3_bucket',
      resourceType: 's3:bucket',
      resourceName: '',
      severity: 'low',
      metadata: {
        provider: 'aws',
        last_modified: '2024-02-10T12:00:00Z',
      },
    },
    position: { x: 140, y: 120 },
  },
};

export const HighSeverityWithDrift: Story = {
  args: {
    data: {
      id: 'node-high-drift-001',
      label: 'database-cluster',
      type: 'aws_rds_cluster',
      resourceType: 'rds:cluster',
      resourceName: 'production-postgres-cluster',
      severity: 'high',
      metadata: {
        provider: 'aws',
        has_drift: true,
        drift_count: 4,
        last_modified: '2024-02-21T08:15:00Z',
        user: 'database-admin@company.com',
        mode: 'managed',
      },
    },
    position: { x: 210, y: 160 },
  },
};
