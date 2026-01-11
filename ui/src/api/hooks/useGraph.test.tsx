/**
 * useGraph Hook Tests
 * Tests for graph data React Query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useGraph, useNodes, useEdges } from './useGraph';
import { apiClient } from '../client';
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';
import {
  mockEdge,
  mockGraphData,
  mockPaginatedNodes,
  mockPaginatedEdges,
  createMockNode,
  createMockEdge,
  createLargeGraphData,
  createEmptyGraphData,
} from '../../__tests__/fixtures/graphFixtures';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getGraph: vi.fn(),
    getNodes: vi.fn(),
    getEdges: vi.fn(),
  },
}));

describe('useGraph', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch graph data successfully', async () => {
      vi.mocked(apiClient.getGraph).mockResolvedValue(mockGraphData);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockGraphData);
      expect(apiClient.getGraph).toHaveBeenCalledTimes(1);
    });

    it('should fetch graph with multiple nodes and edges', async () => {
      const largeGraph = createLargeGraphData();

      vi.mocked(apiClient.getGraph).mockResolvedValue(largeGraph);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.nodes).toHaveLength(3);
      expect(result.current.data?.edges).toHaveLength(2);
    });

    it('should fetch empty graph', async () => {
      const emptyGraph = createEmptyGraphData();

      vi.mocked(apiClient.getGraph).mockResolvedValue(emptyGraph);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.nodes).toHaveLength(0);
      expect(result.current.data?.edges).toHaveLength(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('API request failed');
      vi.mocked(apiClient.getGraph).mockRejectedValue(error);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getGraph).mockRejectedValue(error);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle server errors', async () => {
      const error = new Error('HTTP error! status: 500');
      vi.mocked(apiClient.getGraph).mockRejectedValue(error);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key', async () => {
      vi.mocked(apiClient.getGraph).mockResolvedValue(mockGraphData);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['graph']
      expect(apiClient.getGraph).toHaveBeenCalledTimes(1);
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getGraph).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getGraph).mockResolvedValue(mockGraphData);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockGraphData);
    });
  });
});

describe('useNodes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch nodes successfully', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockPaginatedNodes);
      expect(apiClient.getNodes).toHaveBeenCalledTimes(1);
      expect(apiClient.getNodes).toHaveBeenCalledWith(undefined);
    });

    it('should fetch nodes with pagination params', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const params = { page: 2, limit: 20 };
      const { result } = renderHook(() => useNodes(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getNodes).toHaveBeenCalledWith(params);
    });

    it('should fetch nodes with page param only', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const params = { page: 3 };
      const { result } = renderHook(() => useNodes(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getNodes).toHaveBeenCalledWith(params);
    });

    it('should fetch nodes with limit param only', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const params = { limit: 50 };
      const { result } = renderHook(() => useNodes(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getNodes).toHaveBeenCalledWith(params);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('API request failed');
      vi.mocked(apiClient.getNodes).mockRejectedValue(error);

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getNodes).mockRejectedValue(error);

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key without params', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['graph', 'nodes', undefined]
      expect(apiClient.getNodes).toHaveBeenCalledWith(undefined);
    });

    it('should use correct query key with params', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const params = { page: 1, limit: 10 };
      const { result } = renderHook(() => useNodes(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['graph', 'nodes', params]
      expect(apiClient.getNodes).toHaveBeenCalledWith(params);
    });
  });

  describe('Pagination', () => {
    it('should handle multiple pages', async () => {
      const page1 = { ...mockPaginatedNodes, page: 1 };
      const page2 = {
        ...mockPaginatedNodes,
        page: 2,
        data: [createMockNode({ id: 'node2' })],
      };

      vi.mocked(apiClient.getNodes).mockResolvedValueOnce(page1);

      const { result, rerender } = renderHook(
        ({ page }: { page: number }) => useNodes({ page }),
        {
          wrapper: createQueryClientWrapper(),
          initialProps: { page: 1 },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect((result.current.data as typeof page1)?.page).toBe(1);

      // Switch to page 2
      vi.mocked(apiClient.getNodes).mockResolvedValueOnce(page2);
      rerender({ page: 2 });

      await waitFor(() => {
        expect((result.current.data as typeof page2)?.page).toBe(2);
      });
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getNodes).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getNodes).mockResolvedValue(mockPaginatedNodes);

      const { result } = renderHook(() => useNodes(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockPaginatedNodes);
    });
  });
});

describe('useEdges', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch edges successfully', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockPaginatedEdges);
      expect(apiClient.getEdges).toHaveBeenCalledTimes(1);
      expect(apiClient.getEdges).toHaveBeenCalledWith(undefined);
    });

    it('should fetch edges with pagination params', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const params = { page: 2, limit: 20 };
      const { result } = renderHook(() => useEdges(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEdges).toHaveBeenCalledWith(params);
    });

    it('should fetch edges with page param only', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const params = { page: 3 };
      const { result } = renderHook(() => useEdges(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEdges).toHaveBeenCalledWith(params);
    });

    it('should fetch edges with limit param only', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const params = { limit: 50 };
      const { result } = renderHook(() => useEdges(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEdges).toHaveBeenCalledWith(params);
    });

    it('should fetch edges with different relationship types', async () => {
      const edgesWithTypes = {
        ...mockPaginatedEdges,
        data: [
          mockEdge,
          createMockEdge({ id: 'edge2', type: 'contains' }),
          createMockEdge({ id: 'edge3', type: 'references' }),
        ],
      };

      vi.mocked(apiClient.getEdges).mockResolvedValue(edgesWithTypes);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect((result.current.data as typeof mockPaginatedEdges)?.data).toHaveLength(3);
      expect((result.current.data as typeof mockPaginatedEdges)?.data[0].data.type).toBe('depends_on');
      expect((result.current.data as typeof mockPaginatedEdges)?.data[1].data.type).toBe('contains');
      expect((result.current.data as typeof mockPaginatedEdges)?.data[2].data.type).toBe('references');
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('API request failed');
      vi.mocked(apiClient.getEdges).mockRejectedValue(error);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getEdges).mockRejectedValue(error);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key without params', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['graph', 'edges', undefined]
      expect(apiClient.getEdges).toHaveBeenCalledWith(undefined);
    });

    it('should use correct query key with params', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const params = { page: 1, limit: 10 };
      const { result } = renderHook(() => useEdges(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['graph', 'edges', params]
      expect(apiClient.getEdges).toHaveBeenCalledWith(params);
    });
  });

  describe('Pagination', () => {
    it('should handle multiple pages', async () => {
      const page1 = { ...mockPaginatedEdges, page: 1 };
      const page2 = {
        ...mockPaginatedEdges,
        page: 2,
        data: [createMockEdge({ id: 'edge2' })],
      };

      vi.mocked(apiClient.getEdges).mockResolvedValueOnce(page1);

      const { result, rerender } = renderHook(
        ({ page }: { page: number }) => useEdges({ page }),
        {
          wrapper: createQueryClientWrapper(),
          initialProps: { page: 1 },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect((result.current.data as typeof page1)?.page).toBe(1);

      // Switch to page 2
      vi.mocked(apiClient.getEdges).mockResolvedValueOnce(page2);
      rerender({ page: 2 });

      await waitFor(() => {
        expect((result.current.data as typeof page2)?.page).toBe(2);
      });
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getEdges).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getEdges).mockResolvedValue(mockPaginatedEdges);

      const { result } = renderHook(() => useEdges(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockPaginatedEdges);
    });
  });
});
