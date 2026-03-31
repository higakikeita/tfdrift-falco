/**
 * Tests for NodeDetailPanelSubs sub-components
 * Tests NodeProperties, NodeRelationships, and NodeMetadata components
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { NodeProperties, NodeRelationships, NodeMetadata } from './NodeDetailPanelSubs';
import type { Node } from '../types/api';

// Mock API hooks
vi.mock('../api/hooks', () => ({
  useNode: vi.fn(),
  useDependencies: vi.fn(),
  useDependents: vi.fn(),
  useNodeNeighbors: vi.fn(),
  useImpactRadius: vi.fn(),
}));

import { useNode, useDependencies, useDependents, useNodeNeighbors, useImpactRadius } from '../api/hooks';

const mockNodeData = (overrides: Partial<Node> = {}): Node => ({
  id: 'node-123',
  labels: ['Resource', 'AWS'],
  properties: {
    type: 'aws_instance',
    name: 'production-server',
    region: 'us-east-1',
  },
  ...overrides,
});

describe('NodeProperties', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state when data is loading', () => {
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    const { container } = render(<NodeProperties nodeId="node-123" />);

    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('renders "data not found" message when node is null', () => {
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node: null } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    const text = screen.getByText(/ãã¼ãæå ±ãè¦ã¤ããã¾ãã/);
    expect(text).toBeInTheDocument();
  });

  it('renders node properties when data is available', () => {
    const node = mockNodeData();
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    expect(screen.getByText('node-123')).toBeInTheDocument();
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
    expect(screen.getByText('production-server')).toBeInTheDocument();
  });

  it('renders labels as badges', () => {
    const node = mockNodeData({ labels: ['Resource', 'AWS', 'Critical'] });
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    expect(screen.getByText('Resource')).toBeInTheDocument();
    expect(screen.getByText('AWS')).toBeInTheDocument();
    expect(screen.getByText('Critical')).toBeInTheDocument();
  });

  it('handles empty labels array', () => {
    const node = mockNodeData({ labels: [] });
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });

  it('handles missing properties', () => {
    const node = mockNodeData({ properties: {} });
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    expect(screen.getByText('node-123')).toBeInTheDocument();
  });

  it('handles undefined property values', () => {
    const node = mockNodeData({
      properties: { type: undefined, name: undefined },
    });
    (useNode as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { node } },
      isLoading: false,
    });

    render(<NodeProperties nodeId="node-123" />);

    expect(screen.getByText('node-123')).toBeInTheDocument();
  });
});

describe('NodeRelationships', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders dependencies section with loading state', () => {
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: true,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    expect(screen.getByText(/Dependencies|ãã­ã³ãã³ã/)).toBeInTheDocument();
  });

  it('renders dependencies list when available', () => {
    const dependencies: Node[] = [
      mockNodeData({ id: 'dep-1', properties: { name: 'Database', type: 'rds' } }),
      mockNodeData({ id: 'dep-2', properties: { name: 'Cache', type: 'elasticache' } }),
    ];
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    expect(screen.getByText('Database')).toBeInTheDocument();
    expect(screen.getByText('Cache')).toBeInTheDocument();
  });

  it('renders "no dependencies" message when empty', () => {
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies: [] } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    const texts = screen.getAllByText(/ä¾å­åãªã/);
    expect(texts.length).toBeGreaterThan(0);
  });

  it('renders dependents section', () => {
    const dependents: Node[] = [
      mockNodeData({ id: 'dpt-1', properties: { name: 'Service', type: 'lambda' } }),
    ];
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies: [] } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    expect(screen.getByText('Service')).toBeInTheDocument();
  });

  it('renders neighbors section with pagination', () => {
    const neighbors: Node[] = Array.from({ length: 15 }, (_, i) =>
      mockNodeData({
        id: `neighbor-${i}`,
        properties: { name: `Neighbor ${i}`, type: 'resource' },
      })
    );
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies: [] } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    expect(screen.getByText('Neighbor 0')).toBeInTheDocument();
    expect(screen.getByText('Neighbor 9')).toBeInTheDocument();
    expect(screen.queryByText('Neighbor 10')).not.toBeInTheDocument();
    expect(screen.getByText(/... ä» 5 ä»¶/)).toBeInTheDocument();
  });

  it('handles node click when onNodeSelect is provided', async () => {
    const user = userEvent.setup();
    const onNodeSelect = vi.fn();
    const dependencies: Node[] = [
      mockNodeData({ id: 'dep-1', properties: { name: 'Dependency', type: 'resource' } }),
    ];
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" onNodeSelect={onNodeSelect} />);

    const depButton = screen.getByText('Dependency').closest('button');
    if (depButton) {
      await user.click(depButton);
    }

    expect(onNodeSelect).toHaveBeenCalledWith('dep-1');
  });

  it('does not call onNodeSelect if not provided', async () => {
    const user = userEvent.setup();
    const dependencies: Node[] = [
      mockNodeData({ id: 'dep-1', properties: { name: 'Dependency', type: 'resource' } }),
    ];
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependencies } },
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { dependents: [] } },
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { neighbors: [] } },
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    const depButton = screen.getByText('Dependency').closest('button');
    if (depButton) {
      await user.click(depButton);
    }
  });

  it('handles missing relationship data gracefully', () => {
    (useDependencies as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
    });
    (useDependents as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
    });
    (useNodeNeighbors as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    render(<NodeRelationships nodeId="node-123" />);

    const noDataTexts = screen.getAllByText(/ä¾å­åãªã/);
    expect(noDataTexts.length).toBeGreaterThan(0);
  });
});

describe('NodeMetadata', () => {
  const defaultProps = {
    nodeId: 'node-123',
    impactDepth: 2,
    onImpactDepthChange: vi.fn(),
    showImpact: false,
    onShowImpact: vi.fn(),
    onHideImpact: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders impact radius section', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} />);

    const headers = screen.getAllByText(/ã¤ã³ãã¯ãåå¾/);
    expect(headers.length).toBeGreaterThan(0);
  });

  it('renders depth selector dropdown', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} />);

    const select = screen.getByDisplayValue('2 hops');
    expect(select).toBeInTheDocument();
  });

  it('handles depth change', async () => {
    const user = userEvent.setup();
    const onImpactDepthChange = vi.fn();
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(
      <NodeMetadata {...defaultProps} onImpactDepthChange={onImpactDepthChange} />
    );

    const select = screen.getByDisplayValue('2 hops');
    await user.selectOptions(select, '3');

    expect(onImpactDepthChange).toHaveBeenCalledWith(3);
  });

  it('shows "Show Impact" button when impact is not shown', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} showImpact={false} />);

    expect(screen.getByText(/ã¤ã³ãã¯ãåå¾ãè¡¨ç¤º/)).toBeInTheDocument();
  });

  it('calls onShowImpact when show button is clicked', async () => {
    const user = userEvent.setup();
    const onShowImpact = vi.fn();
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} onShowImpact={onShowImpact} />);

    const showButton = screen.getByText(/ã¤ã³ãã¯ãåå¾ãè¡¨ç¤º/);
    await user.click(showButton);

    expect(onShowImpact).toHaveBeenCalled();
  });

  it('shows hide button when impact is shown', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} showImpact={true} />);

    expect(screen.getByText(/ãã¤ã©ã¤ããè§£é¤/)).toBeInTheDocument();
  });

  it('calls onHideImpact when hide button is clicked', async () => {
    const user = userEvent.setup();
    const onHideImpact = vi.fn();
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} showImpact={true} onHideImpact={onHideImpact} />);

    const hideButton = screen.getByText(/ãã¤ã©ã¤ããè§£é¤/);
    await user.click(hideButton);

    expect(onHideImpact).toHaveBeenCalled();
  });

  it('shows loading state while fetching impact data', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    const { container } = render(<NodeMetadata {...defaultProps} showImpact={true} />);

    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('displays impact node count when available', () => {
    const impactNodes: Node[] = [
      mockNodeData({ id: 'impact-1' }),
      mockNodeData({ id: 'impact-2' }),
      mockNodeData({ id: 'impact-3' }),
    ];
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: impactNodes } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} showImpact={true} />);

    const countText = screen.getByText(/å½±é¿ç¯å²:/);
    expect(countText).toBeInTheDocument();
  });

  it('displays zero impact nodes', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} showImpact={true} />);

    expect(screen.getByText(/0/)).toBeInTheDocument();
  });

  it('calls onShowImpactRadius with correct parameters when impact nodes are available', () => {
    const onShowImpactRadius = vi.fn();
    const impactNodes: Node[] = [
      mockNodeData({ id: 'impact-1' }),
      mockNodeData({ id: 'impact-2' }),
    ];
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: impactNodes } },
      isLoading: false,
    });

    render(
      <NodeMetadata
        {...defaultProps}
        showImpact={true}
        onShowImpactRadius={onShowImpactRadius}
        impactDepth={2}
      />
    );

    expect(onShowImpactRadius).toHaveBeenCalledWith(['impact-1', 'impact-2'], 2);
  });

  it('handles all depth options', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: { data: { nodes: [] } },
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} />);

    const select = screen.getByDisplayValue('2 hops');
    expect(select).toHaveProperty('children');
    const options = within(select).getAllByRole('option');
    expect(options).toHaveLength(5);
  });

  it('displays missing impact radius section gracefully', () => {
    (useImpactRadius as ReturnType<typeof vi.fn>).mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    render(<NodeMetadata {...defaultProps} />);

    const headers = screen.getAllByText(/ã¤ã³ãã¯ãåå¾/);
    expect(headers.length).toBeGreaterThan(0);
  });
});
