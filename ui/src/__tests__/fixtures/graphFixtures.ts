/**
 * Graph Mock Data Fixtures
 * Reusable mock data for graph-related tests
 */

import type { CytoscapeElements, CytoscapeNode, CytoscapeEdge } from '../../api/types';
import type { PaginatedResponse } from '../../api/client';

export const mockNode: CytoscapeNode = {
  data: {
    id: 'node1',
    label: 'Test Node',
    type: 'aws_iam_role',
    severity: 'high',
    metadata: {
      arn: 'arn:aws:iam::123456789012:role/test',
    },
  },
};

export const mockEdge: CytoscapeEdge = {
  data: {
    id: 'edge1',
    source: 'node1',
    target: 'node2',
    type: 'depends_on',
    label: 'depends on',
  },
};

export const mockGraphData: CytoscapeElements = {
  nodes: [mockNode],
  edges: [mockEdge],
};

export const mockPaginatedNodes: PaginatedResponse<CytoscapeNode> = {
  data: [mockNode],
  page: 1,
  limit: 10,
  total: 100,
  total_pages: 10,
};

export const mockPaginatedEdges: PaginatedResponse<CytoscapeEdge> = {
  data: [mockEdge],
  page: 1,
  limit: 10,
  total: 150,
  total_pages: 15,
};

/**
 * Creates a mock node with custom properties
 */
export const createMockNode = (overrides: Partial<CytoscapeNode['data']> = {}): CytoscapeNode => ({
  data: {
    ...mockNode.data,
    ...overrides,
  },
});

/**
 * Creates a mock edge with custom properties
 */
export const createMockEdge = (overrides: Partial<CytoscapeEdge['data']> = {}): CytoscapeEdge => ({
  data: {
    ...mockEdge.data,
    ...overrides,
  },
});

/**
 * Creates a large graph for testing with multiple nodes and edges
 */
export const createLargeGraphData = (): CytoscapeElements => ({
  nodes: [
    mockNode,
    createMockNode({ id: 'node2', label: 'Node 2' }),
    createMockNode({ id: 'node3', label: 'Node 3' }),
  ],
  edges: [
    mockEdge,
    createMockEdge({ id: 'edge2', source: 'node2', target: 'node3' }),
  ],
});

/**
 * Creates an empty graph for testing
 */
export const createEmptyGraphData = (): CytoscapeElements => ({
  nodes: [],
  edges: [],
});
