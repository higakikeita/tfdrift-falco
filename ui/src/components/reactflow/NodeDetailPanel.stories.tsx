/**
 * NodeDetailPanel Component Stories
 * Interactive documentation and testing for the NodeDetailPanel component
 */

import type { Meta, StoryObj } from '@storybook/react';
import { NodeDetailPanel } from './NodeDetailPanel';
import type { Node } from 'reactflow';

const meta: Meta<typeof NodeDetailPanel> = {
  title: 'Components/Graph/NodeDetailPanel',
  component: NodeDetailPanel,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  argTypes: {
    node: {
      description: 'The selected node to display details for',
    },
    onClose: {
      description: 'Callback function when the panel is closed',
    },
  },
  args: {
    onClose: () => console.log('Panel closed'),
  },
};

export default meta;
type Story = StoryObj<typeof NodeDetailPanel>;

/**
 * Default node detail panel with basic information
 */
export const Default: Story = {
  args: {
    node: {
      id: 'node-1',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-1',
        label: 'aws_iam_role',
        type: 'aws_iam_role',
        severity: 'high',
        resourceName: 'admin-role',
        metadata: {
          arn: 'arn:aws:iam::123456789012:role/admin-role',
          created: '2024-01-15T10:30:00Z',
        },
      },
    } as Node,
  },
};

/**
 * Critical severity node with extensive metadata
 */
export const CriticalSeverity: Story = {
  args: {
    node: {
      id: 'node-2',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-2',
        label: 'production_bucket',
        type: 'aws_s3_bucket',
        severity: 'critical',
        resourceName: 'prod-data-bucket',
        metadata: {
          arn: 'arn:aws:s3:::prod-data-bucket',
          region: 'us-east-1',
          encryption: 'AES256',
          versioning: 'Enabled',
          publicAccess: false,
          tags: {
            Environment: 'production',
            DataClassification: 'confidential',
            Team: 'platform',
          },
        },
      },
    } as Node,
  },
};

/**
 * Medium severity Lambda function
 */
export const MediumSeverityLambda: Story = {
  args: {
    node: {
      id: 'node-3',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-3',
        label: 'data_processor',
        type: 'aws_lambda_function',
        severity: 'medium',
        resourceName: 'data-processor-function',
        metadata: {
          arn: 'arn:aws:lambda:us-east-1:123456789012:function:data-processor',
          runtime: 'nodejs18.x',
          memory: '512MB',
          timeout: '30s',
        },
      },
    } as Node,
  },
};

/**
 * Node without severity level
 */
export const NoSeverity: Story = {
  args: {
    node: {
      id: 'node-4',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-4',
        label: 'simple_resource',
        type: 'aws_iam_policy',
        resourceName: 'readonly-policy',
      },
    } as Node,
  },
};

/**
 * Minimal node with very little data
 */
export const MinimalData: Story = {
  args: {
    node: {
      id: 'node-5',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-5',
        label: 'minimal_node',
        type: 'aws_cloudwatch_log_group',
      },
    } as Node,
  },
};

/**
 * Node with complex nested metadata
 */
export const ComplexMetadata: Story = {
  args: {
    node: {
      id: 'node-6',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-6',
        label: 'complex_role',
        type: 'aws_iam_role',
        severity: 'high',
        resourceName: 'cross-account-assume-role',
        metadata: {
          arn: 'arn:aws:iam::123456789012:role/cross-account-assume-role',
          assumeRolePolicy: {
            Version: '2012-10-17',
            Statement: [
              {
                Effect: 'Allow',
                Principal: {
                  AWS: 'arn:aws:iam::987654321098:root',
                },
                Action: 'sts:AssumeRole',
              },
            ],
          },
          attachedPolicies: [
            'ReadOnlyAccess',
            'CloudWatchLogsReadOnlyAccess',
          ],
          tags: {
            Purpose: 'cross-account-access',
            Requester: 'security-team',
          },
        },
      },
    } as Node,
  },
};

/**
 * GCP resource node
 */
export const GCPResource: Story = {
  args: {
    node: {
      id: 'node-7',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-7',
        label: 'gcp_bucket',
        type: 'gcp_storage_bucket',
        severity: 'high',
        resourceName: 'prod-gcp-bucket',
        metadata: {
          location: 'us-central1',
          storageClass: 'STANDARD',
          uniformBucketLevelAccess: true,
        },
      },
    } as Node,
  },
};

/**
 * Node with very long metadata values
 */
export const LongMetadataValues: Story = {
  args: {
    node: {
      id: 'node-8',
      type: 'custom',
      position: { x: 0, y: 0 },
      data: {
        id: 'node-8',
        label: 'policy_node',
        type: 'aws_iam_policy',
        severity: 'low',
        resourceName: 'custom-policy-with-long-arn',
        metadata: {
          arn: 'arn:aws:iam::123456789012:policy/very-long-policy-name-that-exceeds-normal-display-width',
          description: 'This is a very long description that should be displayed properly even though it contains a lot of text and might need to wrap to multiple lines in the detail panel interface.',
        },
      },
    } as Node,
  },
};

/**
 * Closed state - no node selected
 */
export const Closed: Story = {
  args: {
    node: null,
  },
};
