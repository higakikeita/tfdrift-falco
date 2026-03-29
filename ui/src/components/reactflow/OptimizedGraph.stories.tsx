import type { Meta, StoryObj } from '@storybook/react';
import { ReactFlowProvider } from 'reactflow';
import type { Node, Edge } from 'reactflow';
import 'reactflow/dist/style.css';
import { OptimizedGraph } from './OptimizedGraph';

const meta = {
  title: 'Components/ReactFlow/OptimizedGraph',
  component: OptimizedGraph,
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
} satisfies Meta<typeof OptimizedGraph>;

export default meta;
type Story = StoryObj<typeof meta>;

// Generate small dataset (50 nodes)
function generateSmallDataset(): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const providers = ['aws', 'gcp', 'kubernetes'];
  const severities = ['critical', 'high', 'medium', 'low'];

  for (let i = 0; i < 50; i++) {
    const provider = providers[i % providers.length];
    const severity = severities[i % severities.length];
    nodes.push({
      id: `node-${i}`,
      data: {
        label: `Resource-${i}`,
        provider,
        severity,
        type: `${provider}-service`,
      },
      position: {
        x: Math.random() * 1000,
        y: Math.random() * 800,
      },
      type: 'custom',
    });
  }

  for (let i = 0; i < nodes.length - 1; i++) {
    if (Math.random() > 0.6) {
      edges.push({
        id: `edge-${i}`,
        source: `node-${i}`,
        target: `node-${i + 1}`,
      });
    }
  }

  return { nodes, edges };
}

// Generate medium dataset (500 nodes)
function generateMediumDataset(): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const providers = ['aws', 'gcp', 'kubernetes'];
  const severities = ['critical', 'high', 'medium', 'low'];

  for (let i = 0; i < 500; i++) {
    const provider = providers[i % providers.length];
    const severity = severities[i % severities.length];
    nodes.push({
      id: `node-${i}`,
      data: {
        label: `Resource-${i}`,
        provider,
        severity,
        type: `${provider}-service`,
      },
      position: {
        x: Math.random() * 3000,
        y: Math.random() * 2000,
      },
      type: 'custom',
    });
  }

  for (let i = 0; i < nodes.length - 1; i++) {
    if (Math.random() > 0.85) {
      edges.push({
        id: `edge-${i}`,
        source: `node-${i}`,
        target: `node-${Math.min(i + Math.floor(Math.random() * 10) + 1, nodes.length - 1)}`,
      });
    }
  }

  return { nodes, edges };
}

// Generate large dataset (2000+ nodes)
function generateLargeDataset(): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const providers = ['aws', 'gcp', 'kubernetes'];
  const severities = ['critical', 'high', 'medium', 'low'];

  for (let i = 0; i < 2000; i++) {
    const provider = providers[i % providers.length];
    const severity = severities[i % severities.length];
    nodes.push({
      id: `node-${i}`,
      data: {
        label: `Resource-${i}`,
        provider,
        severity,
        type: `${provider}-service`,
      },
      position: {
        x: Math.random() * 5000,
        y: Math.random() * 3500,
      },
      type: 'custom',
    });
  }

  for (let i = 0; i < nodes.length - 1; i++) {
    if (Math.random() > 0.95) {
      edges.push({
        id: `edge-${i}`,
        source: `node-${i}`,
        target: `node-${Math.min(i + Math.floor(Math.random() * 20) + 1, nodes.length - 1)}`,
      });
    }
  }

  return { nodes, edges };
}

const smallData = generateSmallDataset();
const mediumData = generateMediumDataset();
const largeData = generateLargeDataset();

export const SmallDataset: Story = {
  args: {
    nodes: smallData.nodes,
    edges: smallData.edges,
    enableClustering: false,
    enableProgressiveLoading: false,
    enableLOD: false,
  },
};

export const MediumDataset: Story = {
  args: {
    nodes: mediumData.nodes,
    edges: mediumData.edges,
    enableClustering: true,
    enableProgressiveLoading: false,
    enableLOD: false,
    clusteringOptions: {
      groupBy: 'provider',
      minClusterSize: 5,
      maxClusterSize: 50,
    },
  },
};

export const LargeDatasetWithLOD: Story = {
  args: {
    nodes: largeData.nodes,
    edges: largeData.edges,
    enableClustering: true,
    enableProgressiveLoading: true,
    enableLOD: true,
    clusteringOptions: {
      groupBy: 'provider',
      minClusterSize: 10,
      maxClusterSize: 100,
    },
    progressiveOptions: {
      batchSize: 200,
      batchDelay: 50,
    },
  },
};

export const EmptyState: Story = {
  args: {
    nodes: [],
    edges: [],
    enableClustering: true,
    enableProgressiveLoading: true,
    enableLOD: true,
  },
};

export const WithOnNodeClick: Story = {
  args: {
    nodes: smallData.nodes,
    edges: smallData.edges,
    enableClustering: false,
    enableProgressiveLoading: false,
    enableLOD: false,
    onNodeClick: (node) => {
      console.log('Node clicked:', node.id, node.data);
    },
  },
};

export const ClusteringBySeverity: Story = {
  args: {
    nodes: mediumData.nodes,
    edges: mediumData.edges,
    enableClustering: true,
    enableProgressiveLoading: false,
    enableLOD: false,
    clusteringOptions: {
      groupBy: 'severity',
      minClusterSize: 5,
      maxClusterSize: 50,
    },
  },
};

export const ProgressiveLoadingOnly: Story = {
  args: {
    nodes: mediumData.nodes,
    edges: mediumData.edges,
    enableClustering: false,
    enableProgressiveLoading: true,
    enableLOD: false,
    progressiveOptions: {
      batchSize: 50,
      batchDelay: 100,
    },
  },
};

export const AllFeaturesDisabled: Story = {
  args: {
    nodes: smallData.nodes,
    edges: smallData.edges,
    enableClustering: false,
    enableProgressiveLoading: false,
    enableLOD: false,
  },
};
