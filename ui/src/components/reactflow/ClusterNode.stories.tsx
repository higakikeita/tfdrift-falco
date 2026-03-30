import type { Meta, StoryObj } from '@storybook/react';
import { ReactFlowProvider } from 'reactflow';
import 'reactflow/dist/style.css';
import { ClusterNode, MinimalClusterNode } from './ClusterNode';
import type { NodeProps } from 'reactflow';

interface ClusterNodeData {
  clusterType: string;
  clusterLabel: string;
  childNodeIds: string[];
  isExpanded: boolean;
  childCount: number;
  severityCounts?: Record<string, number>;
  label: string;
}

const meta = {
  title: 'Components/ReactFlow/ClusterNode',
  component: ClusterNode,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <ReactFlowProvider>
        <div style={{
          width: '100%',
          height: '400px',
          backgroundColor: '#0f172a',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}>
          <Story />
        </div>
      </ReactFlowProvider>
    ),
  ],
} satisfies Meta<typeof ClusterNode>;

export default meta;
type Story = StoryObj<typeof meta>;

const createMockNodeProps = (data: ClusterNodeData, selected = false): NodeProps<ClusterNodeData> => ({
  id: `cluster-${data.clusterLabel}`,
  data,
  selected,
  type: 'default',
  isConnectable: true,
  xPos: 0,
  yPos: 0,
  dragging: false,
  zIndex: 1,
});

export const VPCCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'aws-vpc',
      clusterLabel: 'VPC-Main',
      childNodeIds: ['ec2-1', 'ec2-2', 'ec2-3'],
      isExpanded: false,
      childCount: 3,
      label: 'VPC-Main',
    }),
  },
};

export const SubnetCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'aws-subnet',
      clusterLabel: 'Subnet-Public-A',
      childNodeIds: ['ec2-1', 'nat-1'],
      isExpanded: true,
      childCount: 2,
      label: 'Subnet-Public-A',
    }),
  },
};

export const KubernetesCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'kubernetes',
      clusterLabel: 'k8s-default-ns',
      childNodeIds: ['pod-1', 'pod-2', 'pod-3', 'pod-4', 'pod-5'],
      isExpanded: false,
      childCount: 5,
      label: 'k8s-default-ns',
    }),
  },
};

export const GCPCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'gcp-project',
      clusterLabel: 'GCP-Production',
      childNodeIds: ['gce-1', 'gce-2', 'cloudsql-1'],
      isExpanded: true,
      childCount: 3,
      label: 'GCP-Production',
    }),
  },
};

export const WithSeverityCounts: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'aws-vpc',
      clusterLabel: 'VPC-Prod',
      childNodeIds: ['node1', 'node2', 'node3', 'node4', 'node5', 'node6'],
      isExpanded: false,
      childCount: 6,
      severityCounts: {
        critical: 1,
        high: 2,
        medium: 2,
        low: 1,
      },
      label: 'VPC-Prod',
    }),
  },
};

export const LargeCriticalCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'critical',
      clusterLabel: 'Critical-Cluster',
      childNodeIds: Array.from({ length: 25 }, (_, i) => `node-${i}`),
      isExpanded: true,
      childCount: 25,
      severityCounts: {
        critical: 8,
        high: 10,
        medium: 5,
        low: 2,
      },
      label: 'Critical-Cluster',
    }),
  },
};

export const HighSeverityCluster: Story = {
  args: {
    ...createMockNodeProps({
      clusterType: 'high',
      clusterLabel: 'High-Risk-Zone',
      childNodeIds: ['h1', 'h2', 'h3', 'h4'],
      isExpanded: false,
      childCount: 4,
      severityCounts: {
        high: 3,
        medium: 1,
      },
      label: 'High-Risk-Zone',
    }),
  },
};

export const SelectedCluster: Story = {
  args: {
    ...createMockNodeProps(
      {
        clusterType: 'aws-subnet',
        clusterLabel: 'Selected-Subnet',
        childNodeIds: ['s1', 's2', 's3'],
        isExpanded: true,
        childCount: 3,
        label: 'Selected-Subnet',
      },
      true
    ),
  },
};

export const MinimalClusterNodeExpanded: Story = {
  render: () => {
    const data: ClusterNodeData = {
      clusterType: 'aws-vpc',
      clusterLabel: 'VPC-Minimal',
      childNodeIds: ['node1', 'node2', 'node3'],
      isExpanded: true,
      childCount: 3,
      label: 'VPC-Minimal',
    };
    return <MinimalClusterNode data={data} />;
  },
} as unknown as Story;

export const MinimalClusterNodeCollapsed: Story = {
  render: () => {
    const data: ClusterNodeData = {
      clusterType: 'kubernetes',
      clusterLabel: 'k8s-cluster',
      childNodeIds: Array.from({ length: 50 }, (_, i) => `pod-${i}`),
      isExpanded: false,
      childCount: 50,
      label: 'k8s-cluster',
    };
    return <MinimalClusterNode data={data} />;
  },
} as unknown as Story;
