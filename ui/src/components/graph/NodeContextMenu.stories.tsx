import type { Meta, StoryObj } from '@storybook/react';
import { NodeContextMenu } from './NodeContextMenu';

const meta: Meta<typeof NodeContextMenu> = {
  title: 'Graph/NodeContextMenu',
  component: NodeContextMenu,
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
      <div style={{ backgroundColor: '#0f172a', minHeight: '100vh', padding: '50px' }}>
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof NodeContextMenu>;

const mockNodeData = {
  label: 'my-s3-bucket',
  type: 'aws_s3_bucket',
  resource_type: 's3:bucket',
};

export const VisibleMenuTopLeft: Story = {
  args: {
    position: { x: 100, y: 100 },
    nodeId: 'node-001',
    nodeData: mockNodeData,
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const VisibleMenuBottomRight: Story = {
  args: {
    position: { x: window.innerWidth - 300, y: window.innerHeight - 300 },
    nodeId: 'node-002',
    nodeData: {
      label: 'production-database',
      type: 'aws_rds_instance',
      resource_type: 'rds:instance',
    },
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const VisibleMenuCentered: Story = {
  args: {
    position: { x: 400, y: 300 },
    nodeId: 'node-003',
    nodeData: {
      label: 'vpc-network',
      type: 'aws_vpc',
      resource_type: 'vpc',
    },
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const WithoutAllActions: Story = {
  args: {
    position: { x: 200, y: 200 },
    nodeId: 'node-004',
    nodeData: {
      label: 'security-group',
      type: 'aws_security_group',
      resource_type: 'security_group',
    },
    onClose: () => console.log('Menu closed'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const LongResourceName: Story = {
  args: {
    position: { x: 150, y: 150 },
    nodeId: 'node-005',
    nodeData: {
      label: 'very-long-descriptive-s3-bucket-name-for-production-environment-with-many-details',
      type: 'aws_s3_bucket',
      resource_type: 's3:bucket:with:very:long:type:definition',
    },
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const AwsRdsNode: Story = {
  args: {
    position: { x: 300, y: 250 },
    nodeId: 'node-rds-001',
    nodeData: {
      label: 'production-postgres-db',
      type: 'aws_db_instance',
      resource_type: 'rds:db_instance',
    },
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const GcpComputeNode: Story = {
  args: {
    position: { x: 250, y: 200 },
    nodeId: 'node-gcp-001',
    nodeData: {
      label: 'gke-cluster-prod',
      type: 'google_kubernetes_engine_cluster',
      resource_type: 'gke:cluster',
    },
    onClose: () => console.log('Menu closed'),
    onViewDetails: () => console.log('View details clicked'),
    onFocusView: () => console.log('Focus view clicked'),
    onShowDependencies: () => console.log('Show dependencies clicked'),
    onShowImpact: () => console.log('Show impact clicked'),
    onCopyId: () => console.log('Copy ID clicked'),
  },
};

export const MinimalMenu: Story = {
  args: {
    position: { x: 180, y: 120 },
    nodeId: 'node-minimal',
    nodeData: {
      label: 'basic-resource',
      type: 'resource_type',
      resource_type: 'minimal:type',
    },
    onClose: () => console.log('Menu closed'),
  },
};
