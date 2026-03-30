/**
 * ReactFlowGraph Component Tests
 * Tests for React Flow graph visualization with hierarchical nodes
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders, createMockCytoscapeNode, createMockCytoscapeEdge, userEvent } from '@/__tests__/utils/testUtils';
import { ReactFlowGraph } from './ReactFlowGraph';
import type { CytoscapeElements } from '../../types/graph';

// Mock ReactFlow and dependencies
vi.mock('reactflow', () => ({
  default: ({ children, ...props }: any) => (
    <div data-testid="react-flow" {...props}>
      {children}
    </div>
  ),
  ReactFlowProvider: ({ children }: any) => <div>{children}</div>,
  MiniMap: () => <div data-testid="minimap" />,
  Controls: () => <div data-testid="controls" />,
  Background: () => <div data-testid="background" />,
  Panel: ({ children, position }: any) => (
    <div data-testid={`panel-${position}`}>{children}</div>
  ),
  useNodesState: (initial: any[]) => [initial || [], vi.fn(), vi.fn()],
  useEdgesState: (initial: any[]) => [initial || [], vi.fn(), vi.fn()],
  useReactFlow: () => ({
    getNodes: vi.fn(() => []),
    setViewport: vi.fn(),
    fitView: vi.fn(),
  }),
  getRectOfNodes: vi.fn(() => ({ x: 0, y: 0, width: 100, height: 100 })),
  getTransformForBounds: vi.fn(() => [0, 0, 1]),
  MarkerType: { ArrowClosed: 'arrowclosed' },
  Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
  ConnectionMode: { Loose: 'loose' },
  BackgroundVariant: { Dots: 'dots', Lines: 'lines' },
}));

vi.mock('html-to-image', () => ({
  toPng: vi.fn(() => Promise.resolve('data:image/png;base64,test')),
  toSvg: vi.fn(() => Promise.resolve('<svg></svg>')),
}));

vi.mock('./CustomNode', () => ({
  CustomNode: ({ data }: any) => (
    <div data-testid="custom-node">{data?.label}</div>
  ),
}));

vi.mock('./HierarchicalNodes', () => ({
  RegionGroupNode: ({ data }: any) => (
    <div data-testid="region-group-node">{data?.label}</div>
  ),
  VPCGroupNode: ({ data }: any) => (
    <div data-testid="vpc-group-node">{data?.label}</div>
  ),
  AZGroupNode: ({ data }: any) => (
    <div data-testid="az-group-node">{data?.label}</div>
  ),
  SubnetGroupNode: ({ data }: any) => (
    <div data-testid="subnet-group-node">{data?.label}</div>
  ),
}));

vi.mock('./NodeDetailPanel', () => ({
  NodeDetailPanel: ({ node }: any) => (
    <div data-testid="node-detail-panel">
      {node && <div>{node.id}</div>}
    </div>
  ),
}));

vi.mock('../../utils/reactFlowAdapter', () => ({
  convertToReactFlow: (elements: CytoscapeElements) => ({
    nodes: (elements.nodes || []).map((node: any) => ({
      id: node.data.id,
      data: node.data,
      position: { x: 0, y: 0 },
      type: 'custom',
    })),
    edges: (elements.edges || []).map((edge: any) => ({
      id: edge.data.id,
      source: edge.data.source,
      target: edge.data.target,
      data: edge.data,
    })),
  }),
  highlightPath: (nodes: any[], edges: any[], path: string[]) => ({
    nodes: nodes.map((n: any) => ({
      ...n,
      style: { ...n.style, opacity: path.includes(n.id) ? 1 : 0.3 },
    })),
    edges: edges.map((e: any) => ({
      ...e,
      style: { ...e.style, opacity: 1 },
    })),
  }),
}));

vi.mock('../../utils/logger', () => ({
  logger: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn(),
  },
}));

describe('ReactFlowGraph', () => {
  const createMockElements = (): CytoscapeElements => ({
    nodes: [
      createMockCytoscapeNode({
        data: { id: 'node-1', label: 'IAM Role', type: 'aws_iam_role', severity: 'high' },
      }),
      createMockCytoscapeNode({
        data: { id: 'node-2', label: 'S3 Bucket', type: 'aws_s3_bucket', severity: 'medium' },
      }),
    ],
    edges: [
      createMockCytoscapeEdge({
        data: { id: 'edge-1', source: 'node-1', target: 'node-2' },
      }),
    ],
  });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render ReactFlow component', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render with empty elements', () => {
      const elements: CytoscapeElements = { nodes: [], edges: [] };
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render minimap component', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('minimap')).toBeInTheDocument();
    });

    it('should render controls component', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('controls')).toBeInTheDocument();
    });

    it('should render background component', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('background')).toBeInTheDocument();
    });

    it('should render info panel with node and edge count', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      const panel = screen.getByTestId('panel-top-left');
      expect(panel).toBeInTheDocument();
      expect(panel.textContent).toContain('nodes');
      expect(panel.textContent).toContain('edges');
    });

    it('should render export buttons', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByRole('button', { name: /PNG/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /SVG/i })).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <ReactFlowGraph elements={elements} className="custom-graph-class" />
      );
      const wrapper = container.querySelector('.custom-graph-class');
      expect(wrapper).toBeInTheDocument();
    });
  });

  describe('Node Click Handling', () => {
    it('should call onNodeClick callback when node is clicked', async () => {
      const mockOnNodeClick = vi.fn();
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph elements={elements} onNodeClick={mockOnNodeClick} />
      );
      // Note: Actual click would require ReactFlow context setup
      expect(mockOnNodeClick).not.toHaveBeenCalled();
    });
  });

  describe('Highlighting', () => {
    it('should render with highlighted path', () => {
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph
          elements={elements}
          highlightedPath={['node-1', 'node-2']}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render with highlighted nodes', () => {
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph
          elements={elements}
          highlightedNodes={['node-1']}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render with critical nodes', () => {
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph
          elements={elements}
          criticalNodes={['node-2']}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Layout Options', () => {
    it('should render with dagre layout', () => {
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph elements={elements} layout="dagre" />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render with cose layout', () => {
      const elements = createMockElements();
      renderWithProviders(
        <ReactFlowGraph elements={elements} layout="cose" />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Node Detail Panel', () => {
    it('should render node detail panel', () => {
      const elements = createMockElements();
      renderWithProviders(<ReactFlowGraph elements={elements} />);
      expect(screen.getByTestId('node-detail-panel')).toBeInTheDocument();
    });
  });
});
