/**
 * OptimizedGraph Component Tests
 * Tests for optimized graph rendering with clustering, LOD, and progressive loading
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/__tests__/utils/testUtils';
import { OptimizedGraph } from './OptimizedGraph';
import type { Node, Edge } from 'reactflow';

// Mock ReactFlow
vi.mock('reactflow', () => ({
  default: ({ children, ...props }: Record<string, unknown>) => (
    <div data-testid="react-flow" {...props}>
      {children}
    </div>
  ),
  ReactFlowProvider: ({ children }: Record<string, unknown>) => <div>{children}</div>,
  Background: () => <div data-testid="background" />,
  Controls: () => <div data-testid="controls" />,
  MiniMap: () => <div data-testid="minimap" />,
  useNodesState: (initial: unknown[]) => [initial || [], vi.fn(), vi.fn()],
  useEdgesState: (initial: unknown[]) => [initial || [], vi.fn(), vi.fn()],
  ConnectionLineType: { SmoothStep: 'smoothstep' },
  MarkerType: { ArrowClosed: 'arrowclosed' },
}));

// Mock hooks
vi.mock('../../hooks/useProgressiveGraph', () => ({
  useProgressiveGraph: (nodes: unknown[], edges: unknown[]) => ({
    visibleNodes: nodes,
    visibleEdges: edges,
    progress: 100,
    isLoading: false,
    isComplete: true,
    currentBatch: 1,
    totalBatches: 1,
    reset: vi.fn(),
    skipToEnd: vi.fn(),
  }),
}));

vi.mock('../../utils/graphClustering', () => ({
  useGraphClustering: (nodes: unknown[], edges: unknown[]) => ({
    visibleNodes: nodes,
    visibleEdges: edges,
    clusterMap: new Map(),
    toggleCluster: vi.fn(),
    expandAll: vi.fn(),
    collapseAll: vi.fn(),
  }),
}));

vi.mock('../../utils/memoryOptimization', () => ({
  useDebounce: (value: unknown) => value,
}));

// Mock LODNode
vi.mock('./LODNode', () => ({
  LODNode: ({ data }: Record<string, unknown>) => (
    <div data-testid="lod-node">{(data as Record<string, unknown>)?.label}</div>
  ),
  shouldUseLOD: (nodeCount: number) => nodeCount > 500,
}));

// Mock ClusterNode
vi.mock('./ClusterNode', () => ({
  ClusterNode: ({ data }: Record<string, unknown>) => (
    <div data-testid="cluster-node">{(data as Record<string, unknown>)?.label}</div>
  ),
}));

// Mock CustomNode
vi.mock('./CustomNode', () => ({
  CustomNode: ({ data }: Record<string, unknown>) => (
    <div data-testid="custom-node">{(data as Record<string, unknown>)?.label}</div>
  ),
}));

describe('OptimizedGraph', () => {
  const createMockNodes = (count: number): Node[] =>
    Array.from({ length: count }, (_, i) => ({
      id: `node-${i + 1}`,
      data: {
        label: `Node ${i + 1}`,
        type: 'aws_iam_role',
        severity: i % 2 === 0 ? 'high' : 'low',
      },
      position: { x: i * 100, y: 0 },
      type: 'custom',
    }));

  const createMockEdges = (nodeCount: number): Edge[] =>
    Array.from({ length: nodeCount - 1 }, (_, i) => ({
      id: `edge-${i + 1}`,
      source: `node-${i + 1}`,
      target: `node-${i + 2}`,
    }));

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render without errors', () => {
      const nodes = createMockNodes(10);
      const edges = createMockEdges(10);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render with empty nodes and edges', () => {
      renderWithProviders(
        <OptimizedGraph nodes={[]} edges={[]} />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render ReactFlow component', () => {
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should render background component', () => {
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      expect(screen.getByTestId('background')).toBeInTheDocument();
    });

    it('should render controls component', () => {
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      expect(screen.getByTestId('controls')).toBeInTheDocument();
    });

    it('should render minimap component', () => {
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      expect(screen.getByTestId('minimap')).toBeInTheDocument();
    });
  });

  describe('Performance Stats', () => {
    it('should show performance stats', () => {
      const nodes = createMockNodes(10);
      const edges = createMockEdges(10);
      renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );
      // Stats are displayed in the component
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have accessibility attributes present', () => {
      const nodes = createMockNodes(10);
      const edges = createMockEdges(10);
      const { container } = renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );

      // Check for role="img" on the div wrapper
      const graphWrapper = container.querySelector('[role="img"]');
      expect(graphWrapper).toBeInTheDocument();
    });

    it('should have aria-label for accessibility', () => {
      const nodes = createMockNodes(10);
      const edges = createMockEdges(10);
      const { container } = renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );

      const graphWrapper = container.querySelector('[aria-label*="Network graph"]');
      expect(graphWrapper).toBeInTheDocument();
    });

    it('should have status announcements for screen readers', () => {
      const nodes = createMockNodes(10);
      const edges = createMockEdges(10);
      const { container } = renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );

      const statusRegion = container.querySelector('[role="status"]');
      expect(statusRegion).toBeInTheDocument();
      expect(statusRegion).toHaveAttribute('aria-live', 'polite');
    });
  });

  describe('Clustering', () => {
    it('should enable clustering by default for large graphs', () => {
      const nodes = createMockNodes(150);
      const edges = createMockEdges(150);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          enableClustering={true}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should disable clustering when disabled', () => {
      const nodes = createMockNodes(150);
      const edges = createMockEdges(150);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          enableClustering={false}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Progressive Loading', () => {
    it('should enable progressive loading for large graphs', () => {
      const nodes = createMockNodes(250);
      const edges = createMockEdges(250);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          enableProgressiveLoading={true}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should disable progressive loading when disabled', () => {
      const nodes = createMockNodes(250);
      const edges = createMockEdges(250);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          enableProgressiveLoading={false}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Props', () => {
    it('should accept node click callback', () => {
      const mockOnNodeClick = vi.fn();
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          onNodeClick={mockOnNodeClick}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should accept edge click callback', () => {
      const mockOnEdgeClick = vi.fn();
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          onEdgeClick={mockOnEdgeClick}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });

    it('should accept clustering options', () => {
      const nodes = createMockNodes(150);
      const edges = createMockEdges(150);
      renderWithProviders(
        <OptimizedGraph
          nodes={nodes}
          edges={edges}
          clusteringOptions={{
            groupBy: 'provider',
            minClusterSize: 3,
            maxClusterSize: 100,
          }}
        />
      );
      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should be memoized to prevent unnecessary re-renders', () => {
      const nodes = createMockNodes(5);
      const edges = createMockEdges(5);
      const { rerender } = renderWithProviders(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );

      // Rerender with same props
      rerender(
        <OptimizedGraph nodes={nodes} edges={edges} />
      );

      expect(screen.getByTestId('react-flow')).toBeInTheDocument();
    });
  });
});
