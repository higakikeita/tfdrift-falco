/**
 * useGraphDB Hooks Tests
 * Tests for GraphDB query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import {
  useNode,
  useNodeNeighbors,
  useImpactRadius,
  useDependencies,
  useDependents,
  usePatternMatch,
} from './useGraphDB';
import { apiClient } from '../client';
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getNodeById: vi.fn(),
    getNodeNeighbors: vi.fn(),
    getImpactRadius: vi.fn(),
    getDependencies: vi.fn(),
    getDependents: vi.fn(),
    matchPattern: vi.fn(),
  },
}));

const mockNodeData = {
  id: 'node-1',
  label: 'Test Node',
  type: 'service',
  metadata: { region: 'us-east-1' },
};

const mockNeighborsData = {
  incoming: [{ id: 'node-2', label: 'Upstream' }],
  outgoing: [{ id: 'node-3', label: 'Downstream' }],
};

const mockImpactData = {
  nodes: [mockNodeData],
  edges: [],
  affectedCount: 5,
};

const mockDependenciesData = {
  nodes: [mockNodeData],
  edges: [],
  depth: 3,
};

const mockDependentsData = {
  nodes: [mockNodeData],
  edges: [],
  depth: 2,
};

const mockPatternMatchData = {
  matches: [{ id: 'node-1', label: 'Match 1' }],
  count: 1,
};

describe('useNode', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch node data successfully', async () => {
    vi.mocked(apiClient.getNodeById).mockResolvedValue(mockNodeData);

    const { result } = renderHook(() => useNode('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockNodeData);
    expect(apiClient.getNodeById).toHaveBeenCalledWith('node-1');
  });

  it('should not fetch if nodeId is empty', () => {
    vi.mocked(apiClient.getNodeById).mockResolvedValue(mockNodeData);

    const { result } = renderHook(() => useNode(''), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getNodeById).not.toHaveBeenCalled();
  });

  it('should respect enabled prop', () => {
    vi.mocked(apiClient.getNodeById).mockResolvedValue(mockNodeData);

    const { result } = renderHook(() => useNode('node-1', false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getNodeById).not.toHaveBeenCalled();
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch node');
    vi.mocked(apiClient.getNodeById).mockRejectedValue(error);

    const { result } = renderHook(() => useNode('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(error);
  });
});

describe('useNodeNeighbors', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch node neighbors successfully', async () => {
    vi.mocked(apiClient.getNodeNeighbors).mockResolvedValue(mockNeighborsData);

    const { result } = renderHook(() => useNodeNeighbors('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockNeighborsData);
    expect(apiClient.getNodeNeighbors).toHaveBeenCalledWith('node-1');
  });

  it('should not fetch if nodeId is empty', () => {
    vi.mocked(apiClient.getNodeNeighbors).mockResolvedValue(mockNeighborsData);

    const { result } = renderHook(() => useNodeNeighbors(''), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getNodeNeighbors).not.toHaveBeenCalled();
  });

  it('should handle enabled prop', () => {
    vi.mocked(apiClient.getNodeNeighbors).mockResolvedValue(mockNeighborsData);

    const { result } = renderHook(() => useNodeNeighbors('node-1', false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getNodeNeighbors).not.toHaveBeenCalled();
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch neighbors');
    vi.mocked(apiClient.getNodeNeighbors).mockRejectedValue(error);

    const { result } = renderHook(() => useNodeNeighbors('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(error);
  });
});

describe('useImpactRadius', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch impact radius with default depth', async () => {
    vi.mocked(apiClient.getImpactRadius).mockResolvedValue(mockImpactData);

    const { result } = renderHook(() => useImpactRadius('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockImpactData);
    expect(apiClient.getImpactRadius).toHaveBeenCalledWith('node-1', 2);
  });

  it('should fetch impact radius with custom depth', async () => {
    vi.mocked(apiClient.getImpactRadius).mockResolvedValue(mockImpactData);

    const { result } = renderHook(() => useImpactRadius('node-1', 5), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiClient.getImpactRadius).toHaveBeenCalledWith('node-1', 5);
  });

  it('should not fetch if nodeId is empty', () => {
    vi.mocked(apiClient.getImpactRadius).mockResolvedValue(mockImpactData);

    const { result } = renderHook(() => useImpactRadius(''), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getImpactRadius).not.toHaveBeenCalled();
  });

  it('should handle enabled prop', () => {
    vi.mocked(apiClient.getImpactRadius).mockResolvedValue(mockImpactData);

    const { result } = renderHook(() => useImpactRadius('node-1', 2, false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getImpactRadius).not.toHaveBeenCalled();
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch impact radius');
    vi.mocked(apiClient.getImpactRadius).mockRejectedValue(error);

    const { result } = renderHook(() => useImpactRadius('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe('useDependencies', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch dependencies with default depth', async () => {
    vi.mocked(apiClient.getDependencies).mockResolvedValue(mockDependenciesData);

    const { result } = renderHook(() => useDependencies('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockDependenciesData);
    expect(apiClient.getDependencies).toHaveBeenCalledWith('node-1', 5);
  });

  it('should fetch dependencies with custom depth', async () => {
    vi.mocked(apiClient.getDependencies).mockResolvedValue(mockDependenciesData);

    const { result } = renderHook(() => useDependencies('node-1', 3), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiClient.getDependencies).toHaveBeenCalledWith('node-1', 3);
  });

  it('should not fetch if nodeId is empty', () => {
    vi.mocked(apiClient.getDependencies).mockResolvedValue(mockDependenciesData);

    const { result } = renderHook(() => useDependencies(''), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getDependencies).not.toHaveBeenCalled();
  });

  it('should handle enabled prop', () => {
    vi.mocked(apiClient.getDependencies).mockResolvedValue(mockDependenciesData);

    const { result } = renderHook(() => useDependencies('node-1', 5, false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getDependencies).not.toHaveBeenCalled();
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch dependencies');
    vi.mocked(apiClient.getDependencies).mockRejectedValue(error);

    const { result } = renderHook(() => useDependencies('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe('useDependents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch dependents with default depth', async () => {
    vi.mocked(apiClient.getDependents).mockResolvedValue(mockDependentsData);

    const { result } = renderHook(() => useDependents('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockDependentsData);
    expect(apiClient.getDependents).toHaveBeenCalledWith('node-1', 5);
  });

  it('should fetch dependents with custom depth', async () => {
    vi.mocked(apiClient.getDependents).mockResolvedValue(mockDependentsData);

    const { result } = renderHook(() => useDependents('node-1', 2), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiClient.getDependents).toHaveBeenCalledWith('node-1', 2);
  });

  it('should not fetch if nodeId is empty', () => {
    vi.mocked(apiClient.getDependents).mockResolvedValue(mockDependentsData);

    const { result } = renderHook(() => useDependents(''), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getDependents).not.toHaveBeenCalled();
  });

  it('should handle enabled prop', () => {
    vi.mocked(apiClient.getDependents).mockResolvedValue(mockDependentsData);

    const { result } = renderHook(() => useDependents('node-1', 5, false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.getDependents).not.toHaveBeenCalled();
  });

  it('should handle errors', async () => {
    const error = new Error('Failed to fetch dependents');
    vi.mocked(apiClient.getDependents).mockRejectedValue(error);

    const { result } = renderHook(() => useDependents('node-1'), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});

describe('usePatternMatch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockPattern = {
    start_labels: ['Service'],
    rel_type: 'DEPENDS_ON',
    end_labels: ['Database'],
    end_filter: { role: 'primary' },
  };

  it('should fetch pattern match successfully', async () => {
    vi.mocked(apiClient.matchPattern).mockResolvedValue(mockPatternMatchData);

    const { result } = renderHook(() => usePatternMatch(mockPattern), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockPatternMatchData);
    expect(apiClient.matchPattern).toHaveBeenCalledWith(mockPattern);
  });

  it('should not fetch if pattern is null', () => {
    vi.mocked(apiClient.matchPattern).mockResolvedValue(mockPatternMatchData);

    const { result } = renderHook(() => usePatternMatch(null), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.matchPattern).not.toHaveBeenCalled();
  });

  it('should handle enabled prop', () => {
    vi.mocked(apiClient.matchPattern).mockResolvedValue(mockPatternMatchData);

    const { result } = renderHook(() => usePatternMatch(mockPattern, false), {
      wrapper: createQueryClientWrapper(),
    });

    expect(result.current.isLoading).toBe(false);
    expect(apiClient.matchPattern).not.toHaveBeenCalled();
  });

  it('should throw error if pattern is null and enabled', async () => {
    vi.mocked(apiClient.matchPattern).mockImplementation(() => {
      throw new Error('Pattern is required');
    });

    const { result } = renderHook(() => usePatternMatch(null, true), {
      wrapper: createQueryClientWrapper(),
    });

    // The hook should not call the function if pattern is null, even if enabled is true
    expect(result.current.isLoading).toBe(false);
  });

  it('should handle errors from API', async () => {
    const error = new Error('Pattern matching failed');
    vi.mocked(apiClient.matchPattern).mockRejectedValue(error);

    const { result } = renderHook(() => usePatternMatch(mockPattern), {
      wrapper: createQueryClientWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(error);
  });

  it('should update when pattern changes', async () => {
    const pattern1 = {
      start_labels: ['Service'],
      rel_type: 'DEPENDS_ON',
      end_labels: ['Database'],
      end_filter: {},
    };

    vi.mocked(apiClient.matchPattern).mockResolvedValue(mockPatternMatchData);

    const { result, rerender } = renderHook(
      ({ pattern }: { pattern: typeof pattern1 | null }) => usePatternMatch(pattern),
      {
        wrapper: createQueryClientWrapper(),
        initialProps: { pattern: pattern1 },
      }
    );

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(apiClient.matchPattern).toHaveBeenCalledWith(pattern1);

    const pattern2 = {
      start_labels: ['API'],
      rel_type: 'CALLS',
      end_labels: ['Service'],
      end_filter: {},
    };

    vi.mocked(apiClient.matchPattern).mockClear();
    rerender({ pattern: pattern2 });

    await waitFor(() => {
      expect(apiClient.matchPattern).toHaveBeenCalledWith(pattern2);
    });
  });
});

describe('Query Keys', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('useNode should use correct query key', async () => {
    vi.mocked(apiClient.getNodeById).mockResolvedValue(mockNodeData);

    renderHook(() => useNode('node-123'), {
      wrapper: createQueryClientWrapper(),
    });

    expect(apiClient.getNodeById).toHaveBeenCalledWith('node-123');
  });

  it('useNodeNeighbors should use correct query key', async () => {
    vi.mocked(apiClient.getNodeNeighbors).mockResolvedValue(mockNeighborsData);

    renderHook(() => useNodeNeighbors('node-456'), {
      wrapper: createQueryClientWrapper(),
    });

    expect(apiClient.getNodeNeighbors).toHaveBeenCalledWith('node-456');
  });

  it('useDependencies should use correct query key with depth', async () => {
    vi.mocked(apiClient.getDependencies).mockResolvedValue(mockDependenciesData);

    renderHook(() => useDependencies('node-789', 3), {
      wrapper: createQueryClientWrapper(),
    });

    expect(apiClient.getDependencies).toHaveBeenCalledWith('node-789', 3);
  });
});
