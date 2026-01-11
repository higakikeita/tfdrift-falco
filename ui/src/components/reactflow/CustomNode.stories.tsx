/**
 * CustomNode Component Stories
 * Interactive documentation and testing for the CustomNode component
 */

import type { Meta, StoryObj } from '@storybook/react';
import { ReactFlowProvider } from 'reactflow';
import { CustomNode } from './CustomNode';

const meta: Meta<typeof CustomNode> = {
  title: 'Components/Graph/CustomNode',
  component: CustomNode,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <ReactFlowProvider>
        <div style={{ width: '300px', height: '150px' }}>
          <Story />
        </div>
      </ReactFlowProvider>
    ),
  ],
  argTypes: {
    data: {
      description: 'Node data including label, type, severity, and metadata',
    },
    selected: {
      control: 'boolean',
      description: 'Whether the node is currently selected',
    },
  },
};

export default meta;
type Story = StoryObj<typeof CustomNode>;

/**
 * Default node with basic AWS IAM role
 */
export const Default: Story = {
  args: {
    data: {
      label: 'aws_iam_role',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      resource_name: 'default-role',
    },
    selected: false,
  },
};

/**
 * Node with critical severity - highest alert level
 */
export const CriticalSeverity: Story = {
  args: {
    data: {
      label: 'critical_bucket',
      type: 'aws_s3_bucket',
      resource_type: 'aws_s3_bucket',
      severity: 'critical',
      resource_name: 'production-data-bucket',
    },
  },
};

/**
 * Node with high severity
 */
export const HighSeverity: Story = {
  args: {
    data: {
      label: 'important_role',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      severity: 'high',
      resource_name: 'admin-role',
    },
  },
};

/**
 * Node with medium severity
 */
export const MediumSeverity: Story = {
  args: {
    data: {
      label: 'moderate_function',
      type: 'aws_lambda_function',
      resource_type: 'aws_lambda_function',
      severity: 'medium',
      resource_name: 'data-processor',
    },
  },
};

/**
 * Node with low severity
 */
export const LowSeverity: Story = {
  args: {
    data: {
      label: 'minor_resource',
      type: 'aws_cloudwatch_log_group',
      resource_type: 'aws_cloudwatch_log_group',
      resource_name: 'minor-resource',
      severity: 'low',
    },
  },
};

/**
 * Selected node state - shows blue border
 */
export const Selected: Story = {
  args: {
    data: {
      label: 'selected_node',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      resource_name: 'selected-node',
      severity: 'high',
    },
    selected: true,
  },
};

/**
 * Node with very long label to test text overflow
 */
export const LongLabel: Story = {
  args: {
    data: {
      label: 'very-long-resource-name-that-exceeds-normal-width-limits',
      type: 'aws_lambda_function',
      resource_type: 'aws_lambda_function',
      resource_name: 'extremely-long-function-name-for-testing-overflow-behavior',
      severity: 'medium',
    },
  },
};

/**
 * GCP resource node
 */
export const GCPResource: Story = {
  args: {
    data: {
      label: 'gcp_storage_bucket',
      type: 'gcp_storage_bucket',
      resource_type: 'gcp_storage_bucket',
      severity: 'high',
      resource_name: 'prod-data-bucket',
    },
  },
};

/**
 * Node with rich metadata
 */
export const WithMetadata: Story = {
  args: {
    data: {
      label: 'complex_role',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      severity: 'critical',
      resource_name: 'cross-account-role',
      metadata: {
        arn: 'arn:aws:iam::123456789012:role/cross-account-role',
        created: '2024-01-15T10:30:00Z',
        lastModified: '2024-01-20T14:45:00Z',
        tags: {
          Environment: 'production',
          Team: 'platform',
        },
      },
    },
  },
};

/**
 * Minimal node - no severity, no resource name
 */
export const Minimal: Story = {
  args: {
    data: {
      label: 'simple_node',
      type: 'aws_iam_policy',
      resource_type: 'aws_iam_policy',
      resource_name: 'simple-node',
    },
  },
};

/**
 * Interactive example - all states combined
 */
export const Interactive: Story = {
  args: {
    data: {
      label: 'interactive_node',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      severity: 'high',
      resource_name: 'test-role',
    },
    selected: false,
  },
};
