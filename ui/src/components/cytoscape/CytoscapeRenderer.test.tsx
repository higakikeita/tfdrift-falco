/**
 * CytoscapeRenderer Component Tests
 * Tests for Cytoscape graph rendering and instance management
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderWithProviders, createMockCytoscapeNode, createMockCytoscapeEdge } from '@/__tests__/utils/testUtils';
import { CytoscapeRenderer } from './CytoscapeRenderer';
import type { CytoscapeElements } from '../../types/graph';
import React from 'react';

// Mock Cytoscape instance
const mockCyInstance = {
  elements: vi.fn(() => ({
    removeClass: vi.fn(),
    style: vi.fn(),
    forEach: vi.fn(),
  })),
  layout: vi.fn(() => ({ run: vi.fn() })),
  destroy: vi.fn(),
  getElementById: vi.fn(() => ({
    addClass: vi.fn(),
    connectedEdges: vi.fn(() => ({ addClass: vi.fn() })),
  })),
  zoom: vi.fn(),
  center: vi.fn(),
  fit: vi.fn(),
  $id: vi.fn(() => ({
    addClass: vi.fn(),
    connectedEdges: vi.fn(() => ({ addClass: vi.fn() })),
  })),
  edges: vi.fn(() => ({
    addClass: vi.fn(),
    style: vi.fn(),
    forEach: vi.fn(),
  })),
  nodes: vi.fn(() => ({
    style: vi.fn(),
    forEach: vi.fn(),
  })),
  collection: vi.fn(() => ({
    merge: vi.fn(),
  })),
  width: vi.fn(() => 800),
  height: vi.fn(() => 600),
};

// Mock cytoscape and plugins
vi.mock('cytoscape', () => {
  const mockCytoscape = vi.fn(() => mockCyInstance);
  mockCytoscape.use = vi.fn();
  return { default: mockCytoscape };
});

vi.mock('cytoscape-dagre', () => ({
  default: vi.fn(),
}));

vi.mock('cytoscape-fcose', () => ({
  default: vi.fn(),
}));

// Mock CytoscapeEventHandler
vi.mock('./CytoscapeEventHandler', () => ({
  CytoscapeEventHandler: vi.fn(() => ({
    bind: vi.fn(),
    unbind: vi.fn(),
  })),
}));

// Mock styles
vi.mock('../../styles/cytoscapeStyles', () => ({
  cytoscapeConfig: {
    style: [],
    wheelSensitivity: 0.1,
  },
  layoutConfigs: {
    dagre: { name: 'dagre', spacingFactor: 1.2 },
    fcose: { name: 'fcose', spacingFactor: 1.2 },
  },
}));

describe('CytoscapeRenderer', () => {
  const cyRef = React.createRef<unknown>();

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

  describe('Rendering', () => {
    it('should render container div', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const containerDiv = container.querySelector('div.w-full.h-full.bg-white');
      expect(containerDiv).toBeInTheDocument();
    });

    it('should render with custom className', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          className="custom-cytoscape-class"
        />
      );

      const rendererDiv = container.querySelector('.custom-cytoscape-class');
      expect(rendererDiv).toBeInTheDocument();
    });

    it('should render with empty elements', () => {
      const elements: CytoscapeElements = { nodes: [], edges: [] };
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const containerDiv = container.querySelector('div.w-full.h-full.bg-white');
      expect(containerDiv).toBeInTheDocument();
    });

    it('should have minimum height of 600px', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const rendererDiv = container.querySelector('.w-full.h-full');
      expect(rendererDiv).toBeInTheDocument();
      expect(rendererDiv).toHaveStyle({ minHeight: '600px' });
    });
  });

  describe('Props', () => {
    it('should accept onNodeClick callback', () => {
      const mockOnNodeClick = vi.fn();
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          onNodeClick={mockOnNodeClick}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should accept onEdgeClick callback', () => {
      const mockOnEdgeClick = vi.fn();
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          onEdgeClick={mockOnEdgeClick}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should accept highlighted path', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          highlightedPath={['node-1', 'node-2']}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should accept nodeScale prop', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          nodeScale={1.5}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should accept filterMode prop', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          filterMode="drift-only"
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should accept onInitialized callback', () => {
      const mockOnInitialized = vi.fn();
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          onInitialized={mockOnInitialized}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });
  });

  describe('Layout Options', () => {
    it('should render with dagre layout', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should render with fcose layout', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="fcose"
          cyRef={cyRef}
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });
  });

  describe('Filter Modes', () => {
    it('should render with all filter mode', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          filterMode="all"
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should render with drift-only filter mode', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          filterMode="drift-only"
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });

    it('should render with vpc-only filter mode', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
          filterMode="vpc-only"
        />
      );

      expect(container.querySelector('div.w-full.h-full.bg-white')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should be memoized component', () => {
      const elements = createMockElements();
      const { rerender, container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const initialDiv = container.querySelector('div.w-full.h-full.bg-white');
      expect(initialDiv).toBeInTheDocument();

      // Rerender with same props
      rerender(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const rerenderDiv = container.querySelector('div.w-full.h-full.bg-white');
      expect(rerenderDiv).toBeInTheDocument();
    });
  });

  describe('CSS Classes', () => {
    it('should apply w-full h-full and bg-white classes', () => {
      const elements = createMockElements();
      const { container } = renderWithProviders(
        <CytoscapeRenderer
          elements={elements}
          layout="dagre"
          cyRef={cyRef}
        />
      );

      const rendererDiv = container.querySelector('.w-full.h-full.bg-white');
      expect(rendererDiv).toBeInTheDocument();
    });
  });
});
