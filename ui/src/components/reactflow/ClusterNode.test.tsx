/**
 * ClusterNode Component Tests
 * Tests for ReactFlow cluster node component
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ClusterNode, MinimalClusterNode } from './ClusterNode';
import type { NodeProps } from 'reactflow';

// Mock reactflow components
vi.mock('reactflow', () => ({
  Handle: ({ type, position, className }: Record<string, unknown>) => (
    <div data-testid={`handle-${type}`} data-position={position} className={className} />
  ),
  Position: {
    Top: 'top',
    Bottom: 'bottom',
    Left: 'left',
    Right: 'right',
  },
}));

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  ChevronDown: (props: Record<string, unknown>) => <span data-testid="chevron-down" {...props} />,
  ChevronRight: (props: Record<string, unknown>) => <span data-testid="chevron-right" {...props} />,
  Package: (props: Record<string, unknown>) => <span data-testid="package-icon" {...props} />,
}));

describe('ClusterNode', () => {
  const mockNodeProps: NodeProps = {
    id: 'cluster-1',
    data: {
      clusterType: 'aws',
      clusterLabel: 'AWS Resources',
      childNodeIds: ['node-1', 'node-2', 'node-3'],
      isExpanded: false,
      childCount: 3,
      severityCounts: { critical: 1, high: 2, medium: 0, low: 0 },
      label: 'AWS Resources',
    },
    selected: false,
    isConnecting: false,
    xNode: undefined,
  } as unknown as NodeProps;

  describe('Rendering', () => {
    it('should render cluster node with label', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByText('AWS Resources')).toBeInTheDocument();
    });

    it('should render child count badge', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('should render collapsed state when isExpanded is false', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByText('Collapsed')).toBeInTheDocument();
      expect(screen.getByTestId('chevron-right')).toBeInTheDocument();
    });

    it('should render expanded state when isExpanded is true', () => {
      const expandedProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, isExpanded: true },
      };
      render(<ClusterNode {...expandedProps} />);
      expect(screen.getByText('Expanded')).toBeInTheDocument();
      expect(screen.getByTestId('chevron-down')).toBeInTheDocument();
    });

    it('should render handles for connections', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByTestId('handle-target')).toBeInTheDocument();
      expect(screen.getByTestId('handle-source')).toBeInTheDocument();
    });

    it('should render package icon', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByTestId('package-icon')).toBeInTheDocument();
    });
  });

  describe('Severity badges', () => {
    it('should render severity badges when severityCounts exists', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByText('critical')).toBeInTheDocument();
      expect(screen.getByText('high')).toBeInTheDocument();
    });

    it('should render severity counts', () => {
      render(<ClusterNode {...mockNodeProps} />);
      // Count badges should show the number
      expect(screen.getByText('1')).toBeInTheDocument(); // critical count
      expect(screen.getByText('2')).toBeInTheDocument(); // high count
    });

    it('should not render severity badges when severityCounts is empty', () => {
      const noSeverityProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, severityCounts: {} },
      };
      render(<ClusterNode {...noSeverityProps} />);
      expect(screen.queryByText('critical')).not.toBeInTheDocument();
      expect(screen.queryByText('high')).not.toBeInTheDocument();
    });

    it('should filter out zero counts from severity badges', () => {
      const props = {
        ...mockNodeProps,
        data: {
          ...mockNodeProps.data,
          severityCounts: { critical: 1, high: 0, medium: 2, low: 0 },
        },
      };
      render(<ClusterNode {...props} />);
      expect(screen.getByText('critical')).toBeInTheDocument();
      expect(screen.getByText('medium')).toBeInTheDocument();
      // High should not appear because count is 0
      const highElements = screen.queryAllByText('high');
      expect(highElements).toHaveLength(0);
    });

    it('should sort severity badges by priority', () => {
      const props = {
        ...mockNodeProps,
        data: {
          ...mockNodeProps.data,
          severityCounts: { low: 1, medium: 2, critical: 3, high: 4 },
        },
      };
      render(<ClusterNode {...props} />);
      const severityBadges = screen.getAllByText(/critical|high|medium|low/);
      const severityOrder = severityBadges.map(el => el.textContent);
      expect(severityOrder[0]).toBe('critical');
      expect(severityOrder[1]).toBe('high');
      expect(severityOrder[2]).toBe('medium');
      expect(severityOrder[3]).toBe('low');
    });
  });

  describe('Cluster color variants', () => {
    it('should apply AWS color for aws cluster type', () => {
      const { container } = render(<ClusterNode {...mockNodeProps} />);
      const clusterDiv = container.querySelector('[class*="border-orange-500"]');
      expect(clusterDiv).toBeInTheDocument();
    });

    it('should apply GCP color for gcp cluster type', () => {
      const gcpProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, clusterType: 'gcp', clusterLabel: 'GCP Resources' },
      };
      const { container } = render(<ClusterNode {...gcpProps} />);
      const clusterDiv = container.querySelector('[class*="border-blue-500"]');
      expect(clusterDiv).toBeInTheDocument();
    });

    it('should apply Kubernetes color for k8s cluster type', () => {
      const k8sProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, clusterType: 'k8s', clusterLabel: 'Kubernetes' },
      };
      const { container } = render(<ClusterNode {...k8sProps} />);
      const clusterDiv = container.querySelector('[class*="border-purple-500"]');
      expect(clusterDiv).toBeInTheDocument();
    });

    it('should apply critical severity color for critical cluster type', () => {
      const criticalProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, clusterType: 'critical', clusterLabel: 'Critical' },
      };
      const { container } = render(<ClusterNode {...criticalProps} />);
      const clusterDiv = container.querySelector('[class*="border-red-500"]');
      expect(clusterDiv).toBeInTheDocument();
    });

    it('should apply gray color for unknown cluster type', () => {
      const unknownProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, clusterType: 'unknown', clusterLabel: 'Unknown' },
      };
      const { container } = render(<ClusterNode {...unknownProps} />);
      const clusterDiv = container.querySelector('[class*="border-gray-400"]');
      expect(clusterDiv).toBeInTheDocument();
    });
  });

  describe('Selected state', () => {
    it('should not apply selected styles when selected is false', () => {
      const { container } = render(<ClusterNode {...mockNodeProps} />);
      const clusterDiv = container.querySelector('[class*="ring-4"]');
      expect(clusterDiv).not.toBeInTheDocument();
    });

    it('should apply selected styles when selected is true', () => {
      const selectedProps = {
        ...mockNodeProps,
        selected: true,
      };
      const { container } = render(<ClusterNode {...selectedProps} />);
      const clusterDiv = container.querySelector('[class*="ring-4"]');
      expect(clusterDiv).toBeInTheDocument();
    });
  });

  describe('Expansion hints', () => {
    it('should show "Click to expand" when not expanded', () => {
      render(<ClusterNode {...mockNodeProps} />);
      expect(screen.getByText('Click to expand')).toBeInTheDocument();
    });

    it('should show "Click to collapse" when expanded', () => {
      const expandedProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, isExpanded: true },
      };
      render(<ClusterNode {...expandedProps} />);
      expect(screen.getByText('Click to collapse')).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('should handle zero child count', () => {
      const zeroChildProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, childCount: 0 },
      };
      render(<ClusterNode {...zeroChildProps} />);
      expect(screen.getByText('0')).toBeInTheDocument();
    });

    it('should handle large child count', () => {
      const largeProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, childCount: 999 },
      };
      render(<ClusterNode {...largeProps} />);
      expect(screen.getByText('999')).toBeInTheDocument();
    });

    it('should handle empty child node IDs', () => {
      const emptyChildProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, childNodeIds: [] },
      };
      render(<ClusterNode {...emptyChildProps} />);
      expect(screen.getByText('AWS Resources')).toBeInTheDocument();
    });

    it('should handle long cluster labels', () => {
      const longLabelProps = {
        ...mockNodeProps,
        data: {
          ...mockNodeProps.data,
          clusterLabel: 'Very Long AWS Resources Cluster Name That Should Be Truncated',
        },
      };
      const { container } = render(<ClusterNode {...longLabelProps} />);
      const label = container.querySelector('[class*="truncate"]');
      expect(label).toBeInTheDocument();
    });

    it('should handle missing severityCounts', () => {
      const noSeverityProps = {
        ...mockNodeProps,
        data: { ...mockNodeProps.data, severityCounts: undefined },
      };
      render(<ClusterNode {...noSeverityProps} />);
      expect(screen.getByText('AWS Resources')).toBeInTheDocument();
    });
  });
});

describe('MinimalClusterNode', () => {
  const mockMinimalProps = {
    data: {
      clusterType: 'aws',
      clusterLabel: 'AWS Resources',
      childNodeIds: ['node-1', 'node-2'],
      isExpanded: false,
      childCount: 2,
      label: 'AWS',
    },
  };

  describe('Rendering', () => {
    it('should render minimal cluster node', () => {
      const { container } = render(<MinimalClusterNode {...mockMinimalProps} />);
      expect(container.querySelector('[class*="w-4"]')).toBeInTheDocument();
    });

    it('should render child count in badge', () => {
      render(<MinimalClusterNode {...mockMinimalProps} />);
      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should have title attribute with label and count', () => {
      const { container } = render(<MinimalClusterNode {...mockMinimalProps} />);
      const element = container.querySelector('[title]');
      expect(element).toHaveAttribute('title', 'AWS Resources (2 nodes)');
    });
  });

  describe('Color variants', () => {
    it('should apply AWS color', () => {
      const { container } = render(<MinimalClusterNode {...mockMinimalProps} />);
      const circle = container.querySelector('[class*="border-orange-500"]');
      expect(circle).toBeInTheDocument();
    });

    it('should apply GCP color for gcp type', () => {
      const gcpProps = {
        data: { ...mockMinimalProps.data, clusterType: 'gcp' },
      };
      const { container } = render(<MinimalClusterNode {...gcpProps} />);
      const circle = container.querySelector('[class*="border-blue-500"]');
      expect(circle).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('should handle single child', () => {
      const singleChildProps = {
        data: { ...mockMinimalProps.data, childCount: 1 },
      };
      render(<MinimalClusterNode {...singleChildProps} />);
      expect(screen.getByText('1')).toBeInTheDocument();
    });

    it('should handle large child count', () => {
      const largeProps = {
        data: { ...mockMinimalProps.data, childCount: 100 },
      };
      render(<MinimalClusterNode {...largeProps} />);
      expect(screen.getByText('100')).toBeInTheDocument();
    });

    it('should handle zero children', () => {
      const zeroProps = {
        data: { ...mockMinimalProps.data, childCount: 0 },
      };
      render(<MinimalClusterNode {...zeroProps} />);
      expect(screen.getByText('0')).toBeInTheDocument();
    });
  });
});
