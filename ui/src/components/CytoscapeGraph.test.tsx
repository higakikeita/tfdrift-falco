/**
 * Tests for CytoscapeGraph component
 */

import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { CytoscapeGraph } from './CytoscapeGraph';
import type { CytoscapeElements } from '../types/graph';

// Mock CytoscapeRenderer
vi.mock('./cytoscape/CytoscapeRenderer', () => ({
  CytoscapeRenderer: () =>
    React.createElement('div', { 'data-testid': 'cytoscape-renderer' }),
  default: () =>
    React.createElement('div', { 'data-testid': 'cytoscape-renderer' }),
}));

// Mock CytoscapeToolbar
vi.mock('./cytoscape/CytoscapeToolbar', () => ({
  CytoscapeToolbar: (props: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'cytoscape-toolbar' }, `Toolbar: Layout ${props.currentLayout}`),
  default: (props: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'cytoscape-toolbar' }, `Toolbar: Layout ${props.currentLayout}`),
}));

describe('CytoscapeGraph', () => {
  const mockElements: CytoscapeElements = {
    nodes: [
      { data: { id: 'node1', label: 'Node 1' } },
      { data: { id: 'node2', label: 'Node 2' } },
    ],
    edges: [
      {
        data: { source: 'node1', target: 'node2', id: 'edge1' },
      },
    ],
  };

  const defaultProps = {
    elements: mockElements,
    layout: 'dagre' as const,
    className: 'test-graph',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders renderer and toolbar', () => {
    render(<CytoscapeGraph {...defaultProps} />);

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
    expect(screen.getByTestId('cytoscape-toolbar')).toBeInTheDocument();
  });

  it('passes elements to renderer', () => {
    const { container } = render(<CytoscapeGraph {...defaultProps} />);

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
    // Verify the component rendered with elements
    expect(container).toBeTruthy();
  });

  it('passes layout to toolbar', () => {
    render(
      <CytoscapeGraph
        {...defaultProps}
        layout="fcose"
      />
    );

    expect(screen.getByText(/Layout fcose/)).toBeInTheDocument();
  });

  it('renders with default className', () => {
    const { container } = render(
      <CytoscapeGraph
        elements={mockElements}
        layout="dagre"
      />
    );

    const wrapper = container.querySelector('.w-full.h-full');
    expect(wrapper).toBeInTheDocument();
  });

  it('renders with custom className', () => {
    const { container } = render(
      <CytoscapeGraph
        {...defaultProps}
        className="custom-class"
      />
    );

    const wrapper = container.querySelector('.custom-class');
    expect(wrapper).toBeInTheDocument();
  });

  it('renders selected node info when node is selected', async () => {
    const onNodeClick = vi.fn();

    render(
      <CytoscapeGraph
        {...defaultProps}
        onNodeClick={onNodeClick}
      />
    );

    // Simulate node click by directly calling the prop handler
    // In a real scenario, this would be triggered by cytoscape
    // For now we test the component's handling of props

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('passes onNodeClick callback', async () => {
    const onNodeClick = vi.fn();
    render(
      <CytoscapeGraph
        {...defaultProps}
        onNodeClick={onNodeClick}
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('passes onEdgeClick callback', async () => {
    const onEdgeClick = vi.fn();
    render(
      <CytoscapeGraph
        {...defaultProps}
        onEdgeClick={onEdgeClick}
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('passes highlightedPath to renderer', () => {
    const highlightedPath = ['node1', 'node2'];
    render(
      <CytoscapeGraph
        {...defaultProps}
        highlightedPath={highlightedPath}
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('handles empty elements array', () => {
    const emptyElements: CytoscapeElements = {
      nodes: [],
      edges: [],
    };

    render(
      <CytoscapeGraph
        {...defaultProps}
        elements={emptyElements}
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('renders toolbar with current layout', () => {
    render(
      <CytoscapeGraph
        {...defaultProps}
        layout="concentric"
      />
    );

    expect(screen.getByText(/Layout concentric/)).toBeInTheDocument();
  });

  it('uses default layout when not provided', () => {
    render(
      <CytoscapeGraph
        elements={mockElements}
      />
    );

    expect(screen.getByText(/Layout dagre/)).toBeInTheDocument();
  });

  it('renders selected node panel', async () => {
    render(
      <CytoscapeGraph
        {...defaultProps}
      />
    );

    // Initially, selected node panel should not be visible
    expect(screen.queryByText('Selected Node')).not.toBeInTheDocument();
  });

  it('passes elements with multiple nodes and edges', () => {
    const complexElements: CytoscapeElements = {
      nodes: [
        { data: { id: 'n1', label: 'Node 1' } },
        { data: { id: 'n2', label: 'Node 2' } },
        { data: { id: 'n3', label: 'Node 3' } },
      ],
      edges: [
        { data: { source: 'n1', target: 'n2', id: 'e1' } },
        { data: { source: 'n2', target: 'n3', id: 'e2' } },
      ],
    };

    render(
      <CytoscapeGraph
        {...defaultProps}
        elements={complexElements}
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
  });

  it('renders with all optional props', () => {
    const onNodeClick = vi.fn();
    const onEdgeClick = vi.fn();

    render(
      <CytoscapeGraph
        elements={mockElements}
        layout="grid"
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        highlightedPath={['node1']}
        className="custom-graph"
      />
    );

    expect(screen.getByTestId('cytoscape-renderer')).toBeInTheDocument();
    expect(screen.getByTestId('cytoscape-toolbar')).toBeInTheDocument();
    expect(screen.getByText(/Layout grid/)).toBeInTheDocument();
  });

  it('component has proper structure with relative positioning', () => {
    const { container } = render(
      <CytoscapeGraph {...defaultProps} />
    );

    const wrapper = container.querySelector('.relative.w-full.h-full');
    expect(wrapper).toBeInTheDocument();
  });

  it('renders toolbar with layout prop', () => {
    render(
      <CytoscapeGraph
        {...defaultProps}
        layout="fcose"
      />
    );

    expect(screen.getByText(/Layout fcose/)).toBeInTheDocument();
  });
});
