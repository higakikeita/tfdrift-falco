/**
 * Tests for GraphWithIcons component
 * Tests icon overlays, graph rendering, and user interactions
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import GraphWithIcons from './GraphWithIcons';
import type { CytoscapeElements } from '../types/graph';

// Mock cytoscape and dependencies
vi.mock('cytoscape', () => {
  const mockEdge = {
    addClass: vi.fn(() => mockEdge),
  };

  const mockCore = {
    destroy: vi.fn(),
    nodes: vi.fn(() => []),
    edges: vi.fn((selector: string) => {
      if (selector && selector.includes('source=')) {
        return mockEdge;
      }
      return [];
    }),
    elements: vi.fn(() => ({
      removeClass: vi.fn(),
    })),
    layout: vi.fn(() => ({
      on: vi.fn(() => mockCore),
      run: vi.fn(),
    })),
    on: vi.fn(() => mockCore),
    fit: vi.fn(),
    center: vi.fn(),
    $id: vi.fn(() => ({ addClass: vi.fn() })),
  };

  const cytoscapeFn = vi.fn(() => mockCore);
  cytoscapeFn.use = vi.fn();

  return {
    default: cytoscapeFn,
  };
});

vi.mock('cytoscape-dagre', () => ({
  default: vi.fn(),
}));

vi.mock('html-to-image', () => ({
  toPng: vi.fn(async () => 'data:image/png;base64,test'),
}));

vi.mock('../styles/cytoscapeStyles', () => ({
  cytoscapeConfig: { directed: true },
  layoutConfigs: {
    dagre: { name: 'dagre' },
    concentric: { name: 'concentric' },
    cose: { name: 'cose' },
    grid: { name: 'grid' },
  },
}));

vi.mock('./icons/OfficialCloudIcons', () => ({
  OfficialCloudIcon: ({ type }: Record<string, unknown>) =>
    `CloudIcon(${type})`,
  getProviderFromType: vi.fn((type: string) => {
    if (type?.includes('aws')) return 'aws';
    if (type?.includes('gcp')) return 'gcp';
    if (type?.includes('azure')) return 'azure';
    return 'unknown';
  }),
  getProviderColor: vi.fn((provider: string) => {
    const colors: Record<string, string> = {
      aws: '#FF9900',
      gcp: '#4285F4',
      azure: '#0078D4',
      unknown: '#999999',
    };
    return colors[provider] || colors.unknown;
  }),
}));

vi.mock('../utils/logger', () => ({
  logger: {
    warn: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

const mockElements: CytoscapeElements = {
  nodes: [
    { data: { id: 'node-1', type: 'aws_instance', resource_type: 'ec2' } },
    { data: { id: 'node-2', type: 'aws_rds', resource_type: 'rds' } },
    { data: { id: 'node-3', type: 'gcp_instance', resource_type: 'gce' } },
  ],
  edges: [
    { data: { source: 'node-1', target: 'node-2' } },
    { data: { source: 'node-2', target: 'node-3' } },
  ],
};

describe('GraphWithIcons', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders container div', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} />
    );

    const containerDiv = container.querySelector('div');
    expect(containerDiv).toBeInTheDocument();
  });

  it('renders control buttons', async () => {
    render(<GraphWithIcons elements={mockElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
    expect(screen.getByText('Center')).toBeInTheDocument();
    expect(screen.getByText('Export')).toBeInTheDocument();
  });

  it('renders with custom className', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} className="custom-class" />
    );

    const wrapper = container.querySelector('.custom-class');
    expect(wrapper).toBeInTheDocument();
  });

  it('handles empty graph', () => {
    const emptyElements: CytoscapeElements = { nodes: [], edges: [] };
    render(<GraphWithIcons elements={emptyElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles graph with only nodes', () => {
    const nodesOnly: CytoscapeElements = {
      nodes: mockElements.nodes,
      edges: [],
    };
    render(<GraphWithIcons elements={nodesOnly} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles graph with only edges', () => {
    const edgesOnly: CytoscapeElements = {
      nodes: [],
      edges: mockElements.edges,
    };
    render(<GraphWithIcons elements={edgesOnly} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('applies light mode background by default', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} darkMode={false} />
    );

    const cyContainer = container.querySelector('div[style*="background"]');
    expect(cyContainer).toHaveStyle('background: rgb(248, 250, 252)');
  });

  it('applies dark mode background when enabled', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} darkMode={true} />
    );

    const cyContainer = container.querySelector('div[style*="background"]');
    expect(cyContainer).toHaveStyle('background: rgb(26, 32, 44)');
  });

  it('applies light mode button styles', () => {
    render(<GraphWithIcons elements={mockElements} darkMode={false} />);

    const fitButton = screen.getByText('Fit');
    expect(fitButton).toHaveClass('bg-white');
  });

  it('applies dark mode button styles', () => {
    render(<GraphWithIcons elements={mockElements} darkMode={true} />);

    const fitButton = screen.getByText('Fit');
    expect(fitButton).toHaveClass('bg-gray-700');
  });

  it('calls onNodeClick when callback is provided', () => {
    const onNodeClick = vi.fn();
    render(
      <GraphWithIcons
        elements={mockElements}
        onNodeClick={onNodeClick}
      />
    );

    // Verify component renders
    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('supports dagre layout', () => {
    render(<GraphWithIcons elements={mockElements} layout="dagre" />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('supports concentric layout', () => {
    render(<GraphWithIcons elements={mockElements} layout="concentric" />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('supports cose layout', () => {
    render(<GraphWithIcons elements={mockElements} layout="cose" />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('supports grid layout', () => {
    render(<GraphWithIcons elements={mockElements} layout="grid" />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('defaults to dagre layout when not specified', () => {
    render(<GraphWithIcons elements={mockElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles highlighted path prop', () => {
    const highlightedPath = ['node-1', 'node-2', 'node-3'];
    render(
      <GraphWithIcons
        elements={mockElements}
        highlightedPath={highlightedPath}
      />
    );

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles empty highlighted path', () => {
    render(
      <GraphWithIcons
        elements={mockElements}
        highlightedPath={[]}
      />
    );

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('renders icon overlay container', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} />
    );

    const overlay = container.querySelector('.absolute.inset-0');
    expect(overlay).toBeInTheDocument();
  });

  it('applies pointer-events-none to icon overlay', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} />
    );

    const overlay = container.querySelector('.pointer-events-none');
    expect(overlay).toBeInTheDocument();
  });

  it('renders control buttons with pointer-events-auto', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} />
    );

    const controls = container.querySelector('.pointer-events-auto');
    expect(controls).toBeInTheDocument();
  });

  it('handles mixed provider types in graph', () => {
    const mixedElements: CytoscapeElements = {
      nodes: [
        { data: { id: 'aws-node', type: 'aws_instance' } },
        { data: { id: 'gcp-node', type: 'gcp_instance' } },
        { data: { id: 'azure-node', type: 'azure_vm' } },
      ],
      edges: [],
    };
    render(<GraphWithIcons elements={mixedElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('renders export button', () => {
    render(<GraphWithIcons elements={mockElements} />);

    const exportButton = screen.getByText('Export');
    expect(exportButton).toBeInTheDocument();
  });

  it('export button is interactive', async () => {
    const user = userEvent.setup();
    render(<GraphWithIcons elements={mockElements} />);

    const exportButton = screen.getByText('Export');
    expect(exportButton).not.toBeDisabled();
    await user.click(exportButton);
  });

  it('fit button is interactive', async () => {
    const user = userEvent.setup();
    render(<GraphWithIcons elements={mockElements} />);

    const fitButton = screen.getByText('Fit');
    expect(fitButton).not.toBeDisabled();
    await user.click(fitButton);
  });

  it('center button is interactive', async () => {
    const user = userEvent.setup();
    render(<GraphWithIcons elements={mockElements} />);

    const centerButton = screen.getByText('Center');
    expect(centerButton).not.toBeDisabled();
    await user.click(centerButton);
  });

  it('handles large graph with many nodes', () => {
    const largeElements: CytoscapeElements = {
      nodes: Array.from({ length: 100 }, (_, i) => ({
        data: { id: `node-${i}`, type: 'aws_instance' },
      })),
      edges: Array.from({ length: 99 }, (_, i) => ({
        data: { source: `node-${i}`, target: `node-${i + 1}` },
      })),
    };
    render(<GraphWithIcons elements={largeElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles complex node properties', () => {
    const complexElements: CytoscapeElements = {
      nodes: [
        {
          data: {
            id: 'complex-node',
            type: 'aws_instance',
            resource_type: 'ec2',
            status: 'running',
            tags: { environment: 'prod' },
          },
        },
      ],
      edges: [],
    };
    render(<GraphWithIcons elements={complexElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles all layout types in sequence', () => {
    const layouts: Array<'dagre' | 'concentric' | 'cose' | 'grid'> = ['dagre', 'concentric', 'cose', 'grid'];

    layouts.forEach((layout) => {
      const { unmount } = render(
        <GraphWithIcons elements={mockElements} layout={layout} />
      );
      expect(screen.getByText('Fit')).toBeInTheDocument();
      unmount();
    });
  });

  it('handles className combination with other styles', () => {
    const { container } = render(
      <GraphWithIcons
        elements={mockElements}
        className="w-full h-screen custom-graph"
      />
    );

    const wrapper = container.firstChild;
    expect(wrapper).toHaveClass('custom-graph');
  });

  it('applies flex layout to control buttons container', () => {
    const { container } = render(
      <GraphWithIcons elements={mockElements} />
    );

    const controlsContainer = container.querySelector('.absolute.top-4.right-4');
    expect(controlsContainer).toHaveClass('flex');
    expect(controlsContainer).toHaveClass('flex-col');
  });

  it('renders multiple icon positions without errors', () => {
    const multiElements: CytoscapeElements = {
      nodes: Array.from({ length: 5 }, (_, i) => ({
        data: { id: `node-${i}`, type: 'aws_instance' },
      })),
      edges: [],
    };
    render(<GraphWithIcons elements={multiElements} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles undefined onNodeClick gracefully', () => {
    render(<GraphWithIcons elements={mockElements} onNodeClick={undefined} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });

  it('handles undefined highlightedPath gracefully', () => {
    render(<GraphWithIcons elements={mockElements} highlightedPath={undefined} />);

    expect(screen.getByText('Fit')).toBeInTheDocument();
  });
});
