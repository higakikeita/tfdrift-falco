import type { Meta, StoryObj } from '@storybook/react';
import { ReactFlowProvider } from 'reactflow';
import 'reactflow/dist/style.css';
import { ReactFlowGraph } from './ReactFlowGraph';
import type { CytoscapeElements } from '../../types/graph';

const meta = {
  title: 'Components/ReactFlow/ReactFlowGraph',
  component: ReactFlowGraph,
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <ReactFlowProvider>
        <div style={{ width: '100%', height: '600px', backgroundColor: '#0f172a' }}>
          <Story />
        </div>
      </ReactFlowProvider>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta<typeof ReactFlowGraph>;

export default meta;
type Story = StoryObj<typeof meta>;

// Small graph with 5 nodes
const smallGraphData: CytoscapeElements = {
  nodes: [
    { data: { id: 'node1', label: 'EC2-1', severity: 'low', provider: 'AWS', resource_type: 'ec2:instance' }, position: { x: 0, y: 0 } },
    { data: { id: 'node2', label: 'RDS-1', severity: 'medium', provider: 'AWS', resource_type: 'rds:instance' }, position: { x: 200, y: 0 } },
    { data: { id: 'node3', label: 'S3-1', severity: 'low', provider: 'AWS', resource_type: 's3:bucket' }, position: { x: 400, y: 0 } },
    { data: { id: 'node4', label: 'VPC-1', severity: 'low', provider: 'AWS', resource_type: 'ec2:vpc' }, position: { x: 100, y: 150 } },
    { data: { id: 'node5', label: 'Subnet-1', severity: 'low', provider: 'AWS', resource_type: 'ec2:subnet' }, position: { x: 300, y: 150 } },
  ],
  edges: [
    { data: { id: 'edge1', source: 'node1', target: 'node2', label: '' } },
    { data: { id: 'edge2', source: 'node2', target: 'node3', label: '' } },
    { data: { id: 'edge3', source: 'node4', target: 'node1', label: '' } },
    { data: { id: 'edge4', source: 'node5', target: 'node2', label: '' } },
  ],
};

// Medium graph with 15 nodes showing more complexity
const mediumGraphData: CytoscapeElements = {
  nodes: [
    { data: { id: 'n1', label: 'LB-1', severity: 'low', provider: 'AWS', resource_type: 'elb:load-balancer' }, position: { x: 0, y: 0 } },
    { data: { id: 'n2', label: 'ASG-1', severity: 'medium', provider: 'AWS', resource_type: 'autoscaling:group' }, position: { x: 200, y: 0 } },
    { data: { id: 'n3', label: 'EC2-A', severity: 'low', provider: 'AWS', resource_type: 'ec2:instance' }, position: { x: 400, y: -100 } },
    { data: { id: 'n4', label: 'EC2-B', severity: 'low', provider: 'AWS', resource_type: 'ec2:instance' }, position: { x: 400, y: 0 } },
    { data: { id: 'n5', label: 'EC2-C', severity: 'high', provider: 'AWS', resource_type: 'ec2:instance' }, position: { x: 400, y: 100 } },
    { data: { id: 'n6', label: 'RDS-Primary', severity: 'high', provider: 'AWS', resource_type: 'rds:instance' }, position: { x: 600, y: 0 } },
    { data: { id: 'n7', label: 'RDS-Replica', severity: 'low', provider: 'AWS', resource_type: 'rds:instance' }, position: { x: 600, y: 100 } },
    { data: { id: 'n8', label: 'Cache-1', severity: 'low', provider: 'AWS', resource_type: 'elasticache:cluster' }, position: { x: 800, y: 0 } },
    { data: { id: 'n9', label: 'S3-Bucket', severity: 'critical', provider: 'AWS', resource_type: 's3:bucket' }, position: { x: 1000, y: 0 } },
    { data: { id: 'n10', label: 'VPC-Main', severity: 'low', provider: 'AWS', resource_type: 'ec2:vpc' }, position: { x: 200, y: 200 } },
    { data: { id: 'n11', label: 'Subnet-Public', severity: 'low', provider: 'AWS', resource_type: 'ec2:subnet' }, position: { x: 400, y: 200 } },
    { data: { id: 'n12', label: 'Subnet-Private', severity: 'low', provider: 'AWS', resource_type: 'ec2:subnet' }, position: { x: 400, y: 300 } },
    { data: { id: 'n13', label: 'NAT-Gateway', severity: 'low', provider: 'AWS', resource_type: 'ec2:nat-gateway' }, position: { x: 200, y: 300 } },
    { data: { id: 'n14', label: 'Route-53', severity: 'low', provider: 'AWS', resource_type: 'route53:zone' }, position: { x: 0, y: 150 } },
    { data: { id: 'n15', label: 'CloudFront', severity: 'low', provider: 'AWS', resource_type: 'cloudfront:distribution' }, position: { x: -200, y: 0 } },
  ],
  edges: [
    { data: { id: 'e1', source: 'n1', target: 'n2', label: '' } },
    { data: { id: 'e2', source: 'n2', target: 'n3', label: '' } },
    { data: { id: 'e3', source: 'n2', target: 'n4', label: '' } },
    { data: { id: 'e4', source: 'n2', target: 'n5', label: '' } },
    { data: { id: 'e5', source: 'n3', target: 'n6', label: '' } },
    { data: { id: 'e6', source: 'n4', target: 'n6', label: '' } },
    { data: { id: 'e7', source: 'n5', target: 'n6', label: '' } },
    { data: { id: 'e8', source: 'n6', target: 'n7', label: '' } },
    { data: { id: 'e9', source: 'n6', target: 'n8', label: '' } },
    { data: { id: 'e10', source: 'n8', target: 'n9', label: '' } },
    { data: { id: 'e11', source: 'n10', target: 'n11', label: '' } },
    { data: { id: 'e12', source: 'n10', target: 'n12', label: '' } },
    { data: { id: 'e13', source: 'n11', target: 'n2', label: '' } },
    { data: { id: 'e14', source: 'n12', target: 'n13', label: '' } },
    { data: { id: 'e15', source: 'n14', target: 'n1', label: '' } },
    { data: { id: 'e16', source: 'n15', target: 'n14', label: '' } },
  ],
};

export const SmallGraph: Story = {
  args: {
    elements: smallGraphData,
    layout: 'dagre',
  },
};

export const MediumGraph: Story = {
  args: {
    elements: mediumGraphData,
    layout: 'dagre',
  },
};

export const WithHighlightedNodes: Story = {
  args: {
    elements: mediumGraphData,
    layout: 'dagre',
    highlightedNodes: ['n5', 'n6', 'n9'],
  },
};

export const WithCriticalNodes: Story = {
  args: {
    elements: mediumGraphData,
    layout: 'dagre',
    criticalNodes: ['n9'],
  },
};

export const WithHighlightedPath: Story = {
  args: {
    elements: mediumGraphData,
    layout: 'dagre',
    highlightedPath: ['n1', 'n2', 'n5', 'n6', 'n8', 'n9'],
  },
};

export const GridLayout: Story = {
  args: {
    elements: smallGraphData,
    layout: 'grid',
  },
};

export const ConccentricLayout: Story = {
  args: {
    elements: smallGraphData,
    layout: 'concentric',
  },
};
