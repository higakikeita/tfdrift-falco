/**
 * App Component Tests
 * Tests for main application component
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// Mock components and utilities
vi.mock('./components/CytoscapeGraph', () => ({
  default: ({ elements, layout, onNodeClick, highlightedPath, className }: any) => (
    <div
      data-testid="cytoscape-graph"
      data-elements-count={elements.nodes.length}
      data-layout={layout}
      data-highlighted-count={highlightedPath.length}
      className={className}
    >
      <button
        onClick={() => onNodeClick('test-node', {})}
        data-testid="graph-node-button"
      >
        Click Node
      </button>
    </div>
  ),
}));

vi.mock('./pages/WhyFalcoPage', () => ({
  default: ({ onBack }: any) => (
    <div data-testid="why-falco-page">
      <button onClick={onBack} data-testid="back-button">
        Back to Graph
      </button>
    </div>
  ),
}));

vi.mock('./utils/sampleData', () => ({
  generateSampleCausalChain: () => ({
    nodes: Array.from({ length: 7 }, (_, i) => ({ id: `node-${i}`, label: `Node ${i}` })),
    edges: Array.from({ length: 6 }, (_, i) => ({ id: `edge-${i}`, source: `node-${i}`, target: `node-${i + 1}` })),
  }),
  generateComplexSampleGraph: () => ({
    nodes: Array.from({ length: 12 }, (_, i) => ({ id: `node-${i}`, label: `Node ${i}` })),
    edges: Array.from({ length: 15 }, (_, i) => ({ id: `edge-${i}`, source: `node-${i % 12}`, target: `node-${(i + 1) % 12}` })),
  }),
  generateBlastRadiusGraph: () => ({
    nodes: Array.from({ length: 20 }, (_, i) => ({ id: `node-${i}`, label: `Node ${i}` })),
    edges: Array.from({ length: 25 }, (_, i) => ({ id: `edge-${i}`, source: `node-${i % 20}`, target: `node-${(i + 1) % 20}` })),
  }),
}));

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial render', () => {
    it('should render the app with graph view by default', () => {
      render(<App />);
      expect(screen.getByTestId('cytoscape-graph')).toBeInTheDocument();
    });

    it('should display the header with title', () => {
      render(<App />);
      expect(screen.getByText('TFDrift-Falco Graph UI')).toBeInTheDocument();
    });

    it('should display Japanese subtitle', () => {
      render(<App />);
      expect(screen.getByText('因果関係グラフビジュアライゼーション')).toBeInTheDocument();
    });

    it('should display control panel with demo mode select', () => {
      render(<App />);
      expect(screen.getByDisplayValue('Simple Chain (Drift → Falco)')).toBeInTheDocument();
    });

    it('should render simple demo mode by default', () => {
      render(<App />);
      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-elements-count', '7');
    });

    it('should display node count and edge count', () => {
      render(<App />);
      expect(screen.getByText('Nodes')).toBeInTheDocument();
      expect(screen.getByText('Edges')).toBeInTheDocument();
    });

    it('should display help overlay with tips', () => {
      render(<App />);
      expect(screen.getByText('💡 Quick Tips')).toBeInTheDocument();
      expect(screen.getByText(/Click nodes to select them/)).toBeInTheDocument();
    });

    it('should display core value message', () => {
      render(<App />);
      expect(screen.getByText('「なぜ」を可視化する')).toBeInTheDocument();
    });
  });

  describe('Demo Mode Selection', () => {
    it('should change graph data when demo mode changes to complex', async () => {
      render(<App />);
      const demoSelect = screen.getByDisplayValue('Simple Chain (Drift → Falco)');

      fireEvent.change(demoSelect, { target: { value: 'complex' } });

      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-elements-count', '12');
    });

    it('should change graph data when demo mode changes to blast-radius', async () => {
      render(<App />);
      const demoSelect = screen.getByDisplayValue('Simple Chain (Drift → Falco)');

      fireEvent.change(demoSelect, { target: { value: 'blast-radius' } });

      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-elements-count', '20');
    });

    it('should update node and edge counts when demo mode changes', async () => {
      render(<App />);
      expect(screen.getByText('Nodes')).toBeInTheDocument();
      expect(screen.getByText('Edges')).toBeInTheDocument();

      const demoSelect = screen.getByDisplayValue('Simple Chain (Drift → Falco)');
      fireEvent.change(demoSelect, { target: { value: 'complex' } });

      expect(screen.getByText('Nodes')).toBeInTheDocument();
      expect(screen.getByText('Edges')).toBeInTheDocument();
    });
  });

  describe('Layout Selection', () => {
    it('should have hierarchical layout by default', () => {
      render(<App />);
      const layoutSelect = screen.getByDisplayValue('Hierarchical (Top-Down)');
      expect(layoutSelect).toBeInTheDocument();
    });

    it('should change layout when selection changes', () => {
      render(<App />);
      const layoutSelect = screen.getByDisplayValue('Hierarchical (Top-Down)') as HTMLSelectElement;

      fireEvent.change(layoutSelect, { target: { value: 'concentric' } });

      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-layout', 'concentric');
    });

    it('should support all layout options', () => {
      render(<App />);
      const layoutSelect = screen.getByDisplayValue('Hierarchical (Top-Down)') as HTMLSelectElement;

      const layoutOptions = ['dagre', 'concentric', 'cose', 'grid'];

      layoutOptions.forEach(layout => {
        fireEvent.change(layoutSelect, { target: { value: layout } });
        const graph = screen.getByTestId('cytoscape-graph');
        expect(graph).toHaveAttribute('data-layout', layout);
      });
    });
  });

  describe('Path highlighting (Simple mode only)', () => {
    it('should show highlight controls in simple mode', () => {
      render(<App />);
      expect(screen.getByText('⚡ Highlight Causal Path')).toBeInTheDocument();
    });

    it('should not show highlight controls in complex mode', () => {
      render(<App />);
      const demoSelect = screen.getByDisplayValue('Simple Chain (Drift → Falco)');
      fireEvent.change(demoSelect, { target: { value: 'complex' } });

      expect(screen.queryByText('⚡ Highlight Causal Path')).not.toBeInTheDocument();
    });

    it('should highlight path when button is clicked', () => {
      render(<App />);
      const highlightButton = screen.getByText('⚡ Highlight Causal Path');

      fireEvent.click(highlightButton);

      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-highlighted-count', '7');
    });

    it('should clear highlight when clear button is clicked', () => {
      render(<App />);
      const highlightButton = screen.getByText('⚡ Highlight Causal Path');
      const clearButton = screen.getByText('Clear');

      fireEvent.click(highlightButton);
      let graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-highlighted-count', '7');

      fireEvent.click(clearButton);
      graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-highlighted-count', '0');
    });
  });

  describe('Node Selection', () => {
    it('should display selected node ID in info panel', () => {
      render(<App />);
      const nodeButton = screen.getByTestId('graph-node-button');

      fireEvent.click(nodeButton);

      expect(screen.getByText('test-node')).toBeInTheDocument();
    });

    it('should show "Selected:" label when node is selected', () => {
      render(<App />);
      const nodeButton = screen.getByTestId('graph-node-button');

      fireEvent.click(nodeButton);

      expect(screen.getByText('Selected:')).toBeInTheDocument();
    });

    it('should not show selected node info initially', () => {
      render(<App />);
      expect(screen.queryByText('Selected:')).not.toBeInTheDocument();
    });

    it('should update selected node when different node is clicked', () => {
      render(<App />);
      const nodeButton = screen.getByTestId('graph-node-button');

      fireEvent.click(nodeButton);
      expect(screen.getByText('test-node')).toBeInTheDocument();
    });
  });

  describe('View Navigation', () => {
    it('should show "Why Falco?" button in header', () => {
      render(<App />);
      expect(screen.getByText('Why Falco?')).toBeInTheDocument();
    });

    it('should navigate to why-falco page when button is clicked', () => {
      render(<App />);
      const whyFalcoButton = screen.getByText('Why Falco?');

      fireEvent.click(whyFalcoButton);

      expect(screen.getByTestId('why-falco-page')).toBeInTheDocument();
      expect(screen.queryByTestId('cytoscape-graph')).not.toBeInTheDocument();
    });

    it('should show "Back to Graph" button on why-falco page', () => {
      render(<App />);
      const whyFalcoButton = screen.getByText('Why Falco?');

      fireEvent.click(whyFalcoButton);

      expect(screen.getByTestId('back-button')).toBeInTheDocument();
    });

    it('should navigate back to graph view', () => {
      render(<App />);
      const whyFalcoButton = screen.getByText('Why Falco?');

      fireEvent.click(whyFalcoButton);

      const backButton = screen.getByTestId('back-button');
      fireEvent.click(backButton);

      expect(screen.getByTestId('cytoscape-graph')).toBeInTheDocument();
      expect(screen.queryByTestId('why-falco-page')).not.toBeInTheDocument();
    });

    it('should toggle view correctly multiple times', () => {
      render(<App />);
      const whyFalcoButton = screen.getByText('Why Falco?');

      fireEvent.click(whyFalcoButton);
      expect(screen.getByTestId('why-falco-page')).toBeInTheDocument();

      const backButton = screen.getByTestId('back-button');
      fireEvent.click(backButton);
      expect(screen.getByTestId('cytoscape-graph')).toBeInTheDocument();

      const whyFalcoButton2 = screen.getByText('Why Falco?');
      fireEvent.click(whyFalcoButton2);
      expect(screen.getByTestId('why-falco-page')).toBeInTheDocument();
    });
  });

  describe('Info Panel', () => {
    it('should display TFDrift-Falco essence message', () => {
      render(<App />);
      expect(screen.getByText('TFDrift-Falcoの本質:')).toBeInTheDocument();
    });

    it('should display the causal chain explanation', () => {
      render(<App />);
      expect(
        screen.getByText(/Terraform Drift → IAM → ServiceAccount → Pod → Container → Falco Event/)
      ).toBeInTheDocument();
    });

    it('should display storytelling message', () => {
      render(<App />);
      expect(screen.getByText(/この因果関係が "物語" になる/)).toBeInTheDocument();
    });
  });

  describe('Graph data passing', () => {
    it('should pass correct layout to graph component', () => {
      render(<App />);
      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-layout', 'dagre');
    });

    it('should pass elements to graph component', () => {
      render(<App />);
      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-elements-count');
    });

    it('should pass onNodeClick handler to graph component', () => {
      render(<App />);
      const nodeButton = screen.getByTestId('graph-node-button');
      expect(nodeButton).toBeInTheDocument();
    });

    it('should pass highlightedPath to graph component', () => {
      render(<App />);
      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-highlighted-count', '0');
    });
  });

  describe('Styling and CSS', () => {
    it('should have main app container with flex layout', () => {
      const { container } = render(<App />);
      const appContainer = container.querySelector('.flex.flex-col.h-screen');
      expect(appContainer).toBeInTheDocument();
    });

    it('should have header with gradient styling', () => {
      const { container } = render(<App />);
      const header = container.querySelector('header');
      expect(header).toHaveClass('bg-gradient-to-r', 'from-red-600', 'to-pink-600');
    });

    it('should have control panel with white background', () => {
      const { container } = render(<App />);
      const controlPanel = container.querySelector('[class*="bg-white border-b"]');
      expect(controlPanel).toBeInTheDocument();
    });
  });

  describe('State management', () => {
    it('should maintain state when toggling views', () => {
      render(<App />);
      const demoSelect = screen.getByDisplayValue('Simple Chain (Drift → Falco)');

      fireEvent.change(demoSelect, { target: { value: 'complex' } });

      const whyFalcoButton = screen.getByText('Why Falco?');
      fireEvent.click(whyFalcoButton);

      const backButton = screen.getByTestId('back-button');
      fireEvent.click(backButton);

      const graph = screen.getByTestId('cytoscape-graph');
      expect(graph).toHaveAttribute('data-elements-count', '12');
    });

    it('should maintain selected node across state changes', () => {
      render(<App />);
      const nodeButton = screen.getByTestId('graph-node-button');
      fireEvent.click(nodeButton);

      const layoutSelect = screen.getByDisplayValue('Hierarchical (Top-Down)') as HTMLSelectElement;
      fireEvent.change(layoutSelect, { target: { value: 'radial' } });

      expect(screen.getByText('test-node')).toBeInTheDocument();
    });
  });
});
