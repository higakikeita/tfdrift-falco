/**
 * APIClient Tests
 * Tests for API client fetch-based implementation
 */

import { describe, it, expect, beforeAll, afterEach, afterAll } from 'vitest';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { apiClient, type APIResponse } from './client';
import type { CytoscapeElements, DriftAlert, FalcoEvent, Stats } from './types';

// Setup MSW server
const server = setupServer();

// Mock data
const mockGraphData: CytoscapeElements = {
  nodes: [
    {
      data: {
        id: 'node1',
        label: 'Test Node',
        type: 'aws_iam_role',
        severity: 'high',
      },
    },
  ],
  edges: [
    {
      data: {
        id: 'edge1',
        source: 'node1',
        target: 'node2',
        type: 'depends_on',
      },
    },
  ],
};

const mockDriftAlert: DriftAlert = {
  id: 'drift-123',
  severity: 'high',
  resource_type: 'aws_iam_role',
  resource_name: 'test-role',
  resource_id: 'role-123',
  attribute: 'assume_role_policy',
  old_value: '{}',
  new_value: '{"Version": "2012-10-17"}',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDAI123',
    ARN: 'arn:aws:iam::123456789012:user/test',
    AccountID: '123456789012',
    UserName: 'test-user',
  },
  matched_rules: ['rule1'],
  timestamp: '2024-01-01T00:00:00Z',
  alert_type: 'drift',
};

const mockFalcoEvent: FalcoEvent = {
  id: 'event-123',
  provider: 'aws',
  event_name: 'CreateRole',
  resource_type: 'aws_iam_role',
  resource_id: 'role-123',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDAI123',
    ARN: 'arn:aws:iam::123456789012:user/test',
    AccountID: '123456789012',
    UserName: 'test-user',
  },
  changes: { name: 'test-role' },
  region: 'us-east-1',
  project_id: 'project-123',
  service_name: 'iam',
};

const mockStats: Stats = {
  graph: {
    total_nodes: 100,
    total_edges: 150,
  },
  drifts: {
    total: 50,
    severity_counts: { high: 10, medium: 20, low: 20 },
    resource_types: { aws_iam_role: 15, aws_s3_bucket: 35 },
  },
  events: {
    total: 200,
  },
  unmanaged: {
    total: 10,
  },
  severity_breakdown: { critical: 5, high: 15, medium: 20, low: 10 },
  top_resource_types: [
    { resource_type: 'aws_s3_bucket', count: 35 },
    { resource_type: 'aws_iam_role', count: 15 },
  ],
};

const API_BASE_URL = 'http://localhost:8080/api/v1';

// Helper to create successful API response
const createSuccessResponse = <T>(data: T): APIResponse<T> => ({
  success: true,
  data,
});

// Helper to create error API response
const createErrorResponse = (code: number, message: string): APIResponse<never> => ({
  success: false,
  error: { code, message },
});

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('APIClient', () => {
  describe('Graph API', () => {
    it('should fetch graph data', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph`, () => {
          return HttpResponse.json(createSuccessResponse(mockGraphData));
        })
      );

      const result = await apiClient.getGraph();
      expect(result).toEqual(mockGraphData);
    });

    it('should fetch nodes with pagination', async () => {
      const mockPaginatedNodes = {
        data: mockGraphData.nodes,
        page: 1,
        limit: 10,
        total: 100,
        total_pages: 10,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/nodes`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('1');
          expect(url.searchParams.get('limit')).toBe('10');
          return HttpResponse.json(createSuccessResponse(mockPaginatedNodes));
        })
      );

      const result = await apiClient.getNodes({ page: 1, limit: 10 });
      expect(result).toEqual(mockPaginatedNodes);
    });

    it('should fetch nodes without pagination params', async () => {
      const mockPaginatedNodes = {
        data: mockGraphData.nodes,
        page: 1,
        limit: 50,
        total: 100,
        total_pages: 2,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/nodes`, () => {
          return HttpResponse.json(createSuccessResponse(mockPaginatedNodes));
        })
      );

      const result = await apiClient.getNodes();
      expect(result).toEqual(mockPaginatedNodes);
    });

    it('should fetch edges with pagination', async () => {
      const mockPaginatedEdges = {
        data: mockGraphData.edges,
        page: 2,
        limit: 20,
        total: 150,
        total_pages: 8,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/edges`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('2');
          expect(url.searchParams.get('limit')).toBe('20');
          return HttpResponse.json(createSuccessResponse(mockPaginatedEdges));
        })
      );

      const result = await apiClient.getEdges({ page: 2, limit: 20 });
      expect(result).toEqual(mockPaginatedEdges);
    });
  });

  describe('State API', () => {
    it('should fetch state', async () => {
      const mockState = {
        version: 1,
        terraform_version: '1.5.0',
        serial: 1,
        lineage: 'abc123',
        resource_count: 50,
        outputs_count: 5,
      };

      server.use(
        http.get(`${API_BASE_URL}/state`, () => {
          return HttpResponse.json(createSuccessResponse(mockState));
        })
      );

      const result = await apiClient.getState();
      expect(result).toEqual(mockState);
    });

    it('should fetch state resources with pagination', async () => {
      const mockResources = {
        data: [
          {
            type: 'aws_iam_role',
            name: 'test-role',
            provider: 'aws',
            mode: 'managed',
            attributes: {},
          },
        ],
        page: 1,
        limit: 10,
        total: 50,
        total_pages: 5,
      };

      server.use(
        http.get(`${API_BASE_URL}/state/resources`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('1');
          expect(url.searchParams.get('limit')).toBe('10');
          return HttpResponse.json(createSuccessResponse(mockResources));
        })
      );

      const result = await apiClient.getStateResources({ page: 1, limit: 10 });
      expect(result).toEqual(mockResources);
    });
  });

  describe('Events API', () => {
    it('should fetch events with all filters', async () => {
      const mockPaginatedEvents = {
        data: [mockFalcoEvent],
        page: 1,
        limit: 10,
        total: 200,
        total_pages: 20,
      };

      server.use(
        http.get(`${API_BASE_URL}/events`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('1');
          expect(url.searchParams.get('limit')).toBe('10');
          expect(url.searchParams.get('severity')).toBe('high');
          expect(url.searchParams.get('provider')).toBe('aws');
          return HttpResponse.json(createSuccessResponse(mockPaginatedEvents));
        })
      );

      const result = await apiClient.getEvents({
        page: 1,
        limit: 10,
        severity: 'high',
        provider: 'aws',
      });
      expect(result).toEqual(mockPaginatedEvents);
    });

    it('should fetch events without filters', async () => {
      const mockPaginatedEvents = {
        data: [mockFalcoEvent],
        page: 1,
        limit: 50,
        total: 200,
        total_pages: 4,
      };

      server.use(
        http.get(`${API_BASE_URL}/events`, () => {
          return HttpResponse.json(createSuccessResponse(mockPaginatedEvents));
        })
      );

      const result = await apiClient.getEvents();
      expect(result).toEqual(mockPaginatedEvents);
    });

    it('should fetch single event by id', async () => {
      server.use(
        http.get(`${API_BASE_URL}/events/event-123`, () => {
          return HttpResponse.json(createSuccessResponse(mockFalcoEvent));
        })
      );

      const result = await apiClient.getEvent('event-123');
      expect(result).toEqual(mockFalcoEvent);
    });
  });

  describe('Drifts API', () => {
    it('should fetch drifts with all filters', async () => {
      const mockPaginatedDrifts = {
        data: [mockDriftAlert],
        page: 1,
        limit: 10,
        total: 50,
        total_pages: 5,
      };

      server.use(
        http.get(`${API_BASE_URL}/drifts`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('1');
          expect(url.searchParams.get('limit')).toBe('10');
          expect(url.searchParams.get('severity')).toBe('high');
          expect(url.searchParams.get('resource_type')).toBe('aws_iam_role');
          return HttpResponse.json(createSuccessResponse(mockPaginatedDrifts));
        })
      );

      const result = await apiClient.getDrifts({
        page: 1,
        limit: 10,
        severity: 'high',
        resource_type: 'aws_iam_role',
      });
      expect(result).toEqual(mockPaginatedDrifts);
    });

    it('should fetch drifts without filters', async () => {
      const mockPaginatedDrifts = {
        data: [mockDriftAlert],
        page: 1,
        limit: 50,
        total: 50,
        total_pages: 1,
      };

      server.use(
        http.get(`${API_BASE_URL}/drifts`, () => {
          return HttpResponse.json(createSuccessResponse(mockPaginatedDrifts));
        })
      );

      const result = await apiClient.getDrifts();
      expect(result).toEqual(mockPaginatedDrifts);
    });

    it('should fetch single drift by id', async () => {
      server.use(
        http.get(`${API_BASE_URL}/drifts/drift-123`, () => {
          return HttpResponse.json(createSuccessResponse(mockDriftAlert));
        })
      );

      const result = await apiClient.getDrift('drift-123');
      expect(result).toEqual(mockDriftAlert);
    });
  });

  describe('Stats API', () => {
    it('should fetch stats', async () => {
      server.use(
        http.get(`${API_BASE_URL}/stats`, () => {
          return HttpResponse.json(createSuccessResponse(mockStats));
        })
      );

      const result = await apiClient.getStats();
      expect(result).toEqual(mockStats);
    });
  });

  describe('GraphDB Query API', () => {
    it('should fetch graph stats', async () => {
      const mockGraphStats = {
        node_count: 100,
        edge_count: 150,
        node_types: { IamRole: 25, S3Bucket: 75 },
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/stats`, () => {
          return HttpResponse.json(createSuccessResponse(mockGraphStats));
        })
      );

      const result = await apiClient.getGraphStats();
      expect(result).toEqual(mockGraphStats);
    });

    it('should fetch node by id', async () => {
      const mockNode = mockGraphData.nodes[0];

      server.use(
        http.get(`${API_BASE_URL}/graph/nodes/node1`, () => {
          return HttpResponse.json(createSuccessResponse(mockNode));
        })
      );

      const result = await apiClient.getNodeById('node1');
      expect(result).toEqual(mockNode);
    });

    it('should fetch node neighbors', async () => {
      const mockNeighbors = {
        neighbors: [mockGraphData.nodes[0]],
        count: 1,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/neighbors/node1`, () => {
          return HttpResponse.json(createSuccessResponse(mockNeighbors));
        })
      );

      const result = await apiClient.getNodeNeighbors('node1');
      expect(result).toEqual(mockNeighbors);
    });

    it('should fetch node relationships with direction', async () => {
      const mockRelationships = {
        relationships: [
          {
            source: 'node1',
            target: 'node2',
            type: 'depends_on',
          },
        ],
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/relationships/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('direction')).toBe('outgoing');
          return HttpResponse.json(createSuccessResponse(mockRelationships));
        })
      );

      const result = await apiClient.getNodeRelationships('node1', 'outgoing');
      expect(result).toEqual(mockRelationships);
    });

    it('should fetch node relationships without direction', async () => {
      const mockRelationships = {
        relationships: [
          {
            source: 'node1',
            target: 'node2',
            type: 'depends_on',
          },
        ],
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/relationships/node1`, () => {
          return HttpResponse.json(createSuccessResponse(mockRelationships));
        })
      );

      const result = await apiClient.getNodeRelationships('node1');
      expect(result).toEqual(mockRelationships);
    });

    it('should fetch impact radius', async () => {
      const mockImpact = {
        nodes: [mockGraphData.nodes[0]],
        edges: [mockGraphData.edges[0]],
        depth_levels: { 1: 5, 2: 10, 3: 15 },
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/impact/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('max_depth')).toBe('3');
          return HttpResponse.json(createSuccessResponse(mockImpact));
        })
      );

      const result = await apiClient.getImpactRadius('node1', 3);
      expect(result).toEqual(mockImpact);
    });

    it('should fetch dependencies with custom depth', async () => {
      const mockDependencies = {
        dependencies: [mockGraphData.nodes[0]],
        depth: 3,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/dependencies/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('depth')).toBe('3');
          return HttpResponse.json(createSuccessResponse(mockDependencies));
        })
      );

      const result = await apiClient.getDependencies('node1', 3);
      expect(result).toEqual(mockDependencies);
    });

    it('should fetch dependencies with default depth', async () => {
      const mockDependencies = {
        dependencies: [mockGraphData.nodes[0]],
        depth: 5,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/dependencies/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('depth')).toBe('5');
          return HttpResponse.json(createSuccessResponse(mockDependencies));
        })
      );

      const result = await apiClient.getDependencies('node1');
      expect(result).toEqual(mockDependencies);
    });

    it('should fetch dependents with custom depth', async () => {
      const mockDependents = {
        dependents: [mockGraphData.nodes[0]],
        depth: 3,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/dependents/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('depth')).toBe('3');
          return HttpResponse.json(createSuccessResponse(mockDependents));
        })
      );

      const result = await apiClient.getDependents('node1', 3);
      expect(result).toEqual(mockDependents);
    });

    it('should fetch dependents with default depth', async () => {
      const mockDependents = {
        dependents: [mockGraphData.nodes[0]],
        depth: 5,
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/dependents/node1`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('depth')).toBe('5');
          return HttpResponse.json(createSuccessResponse(mockDependents));
        })
      );

      const result = await apiClient.getDependents('node1');
      expect(result).toEqual(mockDependents);
    });

    it('should fetch critical nodes with custom min', async () => {
      const mockCriticalNodes = {
        critical_nodes: [
          {
            node_id: 'node1',
            connection_count: 10,
            risk_score: 0.8,
          },
        ],
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/critical`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('min')).toBe('5');
          return HttpResponse.json(createSuccessResponse(mockCriticalNodes));
        })
      );

      const result = await apiClient.getCriticalNodes(5);
      expect(result).toEqual(mockCriticalNodes);
    });

    it('should fetch critical nodes with default min', async () => {
      const mockCriticalNodes = {
        critical_nodes: [
          {
            node_id: 'node1',
            connection_count: 10,
            risk_score: 0.8,
          },
        ],
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/critical`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('min')).toBe('3');
          return HttpResponse.json(createSuccessResponse(mockCriticalNodes));
        })
      );

      const result = await apiClient.getCriticalNodes();
      expect(result).toEqual(mockCriticalNodes);
    });

    it('should match pattern with POST request', async () => {
      const pattern = {
        start_labels: ['IamRole'],
        rel_type: 'ASSUMES',
        end_labels: ['Service'],
        end_filter: { name: 'lambda' },
      };

      const mockMatches = {
        matches: [
          {
            start_node: mockGraphData.nodes[0],
            relationship: mockGraphData.edges[0],
            end_node: mockGraphData.nodes[0],
          },
        ],
      };

      server.use(
        http.post(`${API_BASE_URL}/graph/match`, async ({ request }) => {
          const body = await request.json();
          expect(body).toEqual(pattern);
          return HttpResponse.json(createSuccessResponse(mockMatches));
        })
      );

      const result = await apiClient.matchPattern(pattern);
      expect(result).toEqual(mockMatches);
    });
  });

  describe('Error Handling', () => {
    it('should handle HTTP 404 error', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph/nodes/nonexistent`, () => {
          return new HttpResponse(null, { status: 404 });
        })
      );

      await expect(apiClient.getNodeById('nonexistent')).rejects.toThrow(
        'HTTP error! status: 404'
      );
    });

    it('should handle HTTP 500 error', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph`, () => {
          return new HttpResponse(null, { status: 500 });
        })
      );

      await expect(apiClient.getGraph()).rejects.toThrow('HTTP error! status: 500');
    });

    it('should handle API error response', async () => {
      server.use(
        http.get(`${API_BASE_URL}/drifts/invalid`, () => {
          return HttpResponse.json(createErrorResponse(400, 'Invalid drift ID'));
        })
      );

      await expect(apiClient.getDrift('invalid')).rejects.toThrow('Invalid drift ID');
    });

    it('should handle network error', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph`, () => {
          return HttpResponse.error();
        })
      );

      await expect(apiClient.getGraph()).rejects.toThrow();
    });

    it('should handle malformed JSON response', async () => {
      server.use(
        http.get(`${API_BASE_URL}/stats`, () => {
          return new HttpResponse('invalid json', {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          });
        })
      );

      await expect(apiClient.getStats()).rejects.toThrow();
    });

    it('should handle API response with success=false and no error message', async () => {
      server.use(
        http.get(`${API_BASE_URL}/events/bad`, () => {
          return HttpResponse.json({ success: false });
        })
      );

      await expect(apiClient.getEvent('bad')).rejects.toThrow('API request failed');
    });
  });

  describe('Request Headers', () => {
    it('should include Content-Type header', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph`, ({ request }) => {
          expect(request.headers.get('Content-Type')).toBe('application/json');
          return HttpResponse.json(createSuccessResponse(mockGraphData));
        })
      );

      await apiClient.getGraph();
    });

    it('should include Content-Type header in POST request', async () => {
      const pattern = {
        start_labels: ['IamRole'],
        rel_type: 'ASSUMES',
        end_labels: ['Service'],
        end_filter: {},
      };

      server.use(
        http.post(`${API_BASE_URL}/graph/match`, ({ request }) => {
          expect(request.headers.get('Content-Type')).toBe('application/json');
          return HttpResponse.json(createSuccessResponse({ matches: [] }));
        })
      );

      await apiClient.matchPattern(pattern);
    });
  });
});
