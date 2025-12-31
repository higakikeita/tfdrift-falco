/**
 * Test Utilities
 * Common testing helpers and mock data generators
 */

import { render, RenderOptions } from '@testing-library/react';
import { ReactElement, ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactFlowProvider } from 'reactflow';
import type { Drift, DriftChange } from '@/types/drift';
import type { FalcoEvent } from '@/types/falco';
import type { CytoscapeNode, CytoscapeEdge } from '@/types/graph';
import type { Severity } from '@/types/common';

// ==================== Providers ====================

/**
 * Create a new QueryClient for testing
 * Disables retries and logging for cleaner test output
 */
export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
    logger: {
      log: () => {},
      warn: () => {},
      error: () => {},
    },
  });
}

interface AllProvidersProps {
  children: ReactNode;
  queryClient?: QueryClient;
}

/**
 * Wrapper component with all necessary providers
 */
function AllProviders({ children, queryClient }: AllProvidersProps) {
  const client = queryClient || createTestQueryClient();

  return (
    <QueryClientProvider client={client}>
      <ReactFlowProvider>
        {children}
      </ReactFlowProvider>
    </QueryClientProvider>
  );
}

/**
 * Custom render function that wraps components with providers
 */
export function renderWithProviders(
  ui: ReactElement,
  {
    queryClient,
    ...renderOptions
  }: RenderOptions & { queryClient?: QueryClient } = {}
) {
  return render(ui, {
    wrapper: ({ children }) => (
      <AllProviders queryClient={queryClient}>{children}</AllProviders>
    ),
    ...renderOptions,
  });
}

// ==================== Mock Data Generators ====================

/**
 * Generate mock Drift data
 */
export function createMockDrift(overrides?: Partial<Drift>): Drift {
  return {
    id: 'drift-123',
    resource_type: 'aws_iam_role',
    resource_name: 'test-role',
    resource_id: 'arn:aws:iam::123456789012:role/test-role',
    provider: 'aws',
    region: 'us-east-1',
    change_type: 'modified',
    severity: 'high',
    detected_at: '2025-01-01T00:00:00Z',
    changes: [],
    metadata: {},
    ...overrides,
  };
}

/**
 * Generate mock DriftChange data
 */
export function createMockDriftChange(overrides?: Partial<DriftChange>): DriftChange {
  return {
    path: 'AssumeRolePolicyDocument',
    before: { Version: '2012-10-17' },
    after: { Version: '2012-10-17', Statement: [] },
    change_type: 'modified',
    ...overrides,
  };
}

/**
 * Generate mock FalcoEvent data
 */
export function createMockFalcoEvent(overrides?: Partial<FalcoEvent>): FalcoEvent {
  return {
    id: 'falco-456',
    timestamp: '2025-01-01T00:01:00Z',
    priority: 'Warning',
    rule: 'Terminal shell in container',
    output: 'A shell was spawned in a container',
    severity: 'medium',
    source: 'syscall',
    tags: ['container', 'shell'],
    fields: {
      'container.id': 'abc123',
      'container.name': 'test-container',
      'proc.cmdline': '/bin/bash',
    },
    ...overrides,
  };
}

/**
 * Generate mock CytoscapeNode data
 */
export function createMockCytoscapeNode(overrides?: Partial<CytoscapeNode>): CytoscapeNode {
  return {
    data: {
      id: 'node-1',
      label: 'Test Node',
      type: 'aws_iam_role',
      severity: 'high',
      provider: 'aws',
      metadata: {},
      ...overrides?.data,
    },
  };
}

/**
 * Generate mock CytoscapeEdge data
 */
export function createMockCytoscapeEdge(overrides?: Partial<CytoscapeEdge>): CytoscapeEdge {
  return {
    data: {
      id: 'edge-1',
      source: 'node-1',
      target: 'node-2',
      label: 'depends_on',
      ...overrides?.data,
    },
  };
}

/**
 * Generate multiple mock nodes
 */
export function createMockNodes(count: number): CytoscapeNode[] {
  return Array.from({ length: count }, (_, i) =>
    createMockCytoscapeNode({
      data: {
        id: `node-${i + 1}`,
        label: `Node ${i + 1}`,
        type: i % 2 === 0 ? 'aws_iam_role' : 'aws_ec2_instance',
        severity: (['critical', 'high', 'medium', 'low'] as Severity[])[i % 4],
      },
    })
  );
}

/**
 * Generate multiple mock edges connecting nodes
 */
export function createMockEdges(nodeCount: number): CytoscapeEdge[] {
  const edges: CytoscapeEdge[] = [];
  for (let i = 1; i < nodeCount; i++) {
    edges.push(
      createMockCytoscapeEdge({
        data: {
          id: `edge-${i}`,
          source: `node-${i}`,
          target: `node-${i + 1}`,
          label: 'depends_on',
        },
      })
    );
  }
  return edges;
}

// ==================== Test Helpers ====================

/**
 * Wait for async operations to complete
 */
export function waitForAsync() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

/**
 * Mock window.matchMedia for responsive tests
 */
export function mockMatchMedia(matches: boolean = false) {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => true,
    }),
  });
}

/**
 * Get element by test ID
 */
export function getByTestId(container: HTMLElement, testId: string) {
  return container.querySelector(`[data-testid="${testId}"]`);
}

// ==================== React Flow Helpers ====================

/**
 * Mock React Flow instance for testing
 */
export function createMockReactFlowInstance() {
  return {
    getNodes: () => [],
    getEdges: () => [],
    setNodes: () => {},
    setEdges: () => {},
    addNodes: () => {},
    addEdges: () => {},
    fitView: () => {},
    zoomIn: () => {},
    zoomOut: () => {},
    setCenter: () => {},
    getZoom: () => 1,
    setViewport: () => {},
    getViewport: () => ({ x: 0, y: 0, zoom: 1 }),
    screenToFlowPosition: (pos: { x: number; y: number }) => pos,
    flowToScreenPosition: (pos: { x: number; y: number }) => pos,
  };
}

// Re-export testing library utilities
export * from '@testing-library/react';
export { default as userEvent } from '@testing-library/user-event';
