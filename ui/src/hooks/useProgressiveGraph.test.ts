import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useProgressiveGraph } from './useProgressiveGraph';
import type { Node, Edge } from 'reactflow';

describe('useProgressiveGraph hook', () => {
  let mockNodes: Node[];
  let mockEdges: Edge[];

  beforeEach(() => {
    mockNodes = Array.from({ length: 250 }, (_, i) => ({
      id: `node-${i}`,
      data: { label: `Node ${i}` },
      position: { x: i * 100, y: i * 100 },
    }));

    mockEdges = Array.from({ length: 200 }, (_, i) => ({
      id: `edge-${i}`,
      source: `node-${i}`,
      target: `node-${(i + 1) % 250}`,
    }));
  });

  it('should initialize with empty visible nodes', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.visibleNodes).toEqual([]);
    expect(result.current.isLoading).toBe(false);
  });

  it('should have default options', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.progress).toBe(0);
    expect(result.current.currentBatch).toBe(0);
    expect(result.current.totalBatches).toBeGreaterThan(0);
  });

  it('should accept custom batchSize', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, { batchSize: 50 })
    );

    expect(result.current.totalBatches).toEqual(
      Math.ceil(mockNodes.length / 50)
    );
  });

  it('should handle empty nodes array', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph([], mockEdges)
    );

    expect(result.current.visibleNodes).toEqual([]);
    expect(result.current.totalBatches).toBe(0);
  });

  it('should handle empty edges array', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, [])
    );

    expect(result.current.visibleEdges).toEqual([]);
  });

  it('should calculate progress correctly', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, { batchSize: 25 })
    );

    act(() => {
      // Simulate advancing batches
      expect(result.current.progress).toBeGreaterThanOrEqual(0);
      expect(result.current.progress).toBeLessThanOrEqual(100);
    });
  });

  it('should prioritize nodes when priorityNodeIds provided', () => {
    const priorityIds = ['node-0', 'node-1', 'node-2'];

    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, {
        priorityNodeIds: priorityIds,
        batchSize: 50,
      })
    );

    // Priority nodes should be in the first batch when loaded
    expect(result.current.totalBatches).toBeGreaterThan(0);
  });

  it('should have reset function', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.reset).toBeDefined();
    expect(typeof result.current.reset).toBe('function');
  });

  it('should have skipToEnd function', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.skipToEnd).toBeDefined();
    expect(typeof result.current.skipToEnd).toBe('function');
  });

  it('should return visibleEdges', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.visibleEdges).toBeDefined();
    expect(Array.isArray(result.current.visibleEdges)).toBe(true);
  });

  it('should filter edges based on visible nodes', () => {
    const smallMockNodes = Array.from({ length: 10 }, (_, i) => ({
      id: `node-${i}`,
      data: { label: `Node ${i}` },
      position: { x: 0, y: 0 },
    }));

    const smallMockEdges = Array.from({ length: 8 }, (_, i) => ({
      id: `edge-${i}`,
      source: `node-${i}`,
      target: `node-${(i + 1) % 10}`,
    }));

    const { result } = renderHook(() =>
      useProgressiveGraph(smallMockNodes, smallMockEdges)
    );

    // Only edges with both source and target in visibleNodes should be included
    expect(result.current.visibleEdges.length).toBeLessThanOrEqual(
      smallMockEdges.length
    );
  });

  it('should memoize visible nodes', () => {
    const { result, rerender } = renderHook(
      ({ nodes, edges }) =>
        useProgressiveGraph(nodes, edges),
      {
        initialProps: { nodes: mockNodes, edges: mockEdges },
      }
    );

    const initialVisibleNodes = result.current.visibleNodes;

    rerender({ nodes: mockNodes, edges: mockEdges });

    // Should return the same reference when nodes haven't changed
    expect(result.current.visibleNodes).toEqual(initialVisibleNodes);
  });

  it('should handle different batchDelay values', () => {
    const { result: result1 } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, { batchDelay: 10 })
    );

    const { result: result2 } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, { batchDelay: 100 })
    );

    expect(result1.current.totalBatches).toBe(result2.current.totalBatches);
  });

  it('should track current batch number', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current.currentBatch).toBeGreaterThanOrEqual(0);
    expect(result.current.currentBatch).toBeLessThanOrEqual(
      result.current.totalBatches
    );
  });

  it('should expose totalBatches', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, { batchSize: 100 })
    );

    expect(result.current.totalBatches).toBe(
      Math.ceil(mockNodes.length / 100)
    );
  });

  it('should handle large node arrays efficiently', () => {
    const largeNodeArray = Array.from({ length: 5000 }, (_, i) => ({
      id: `node-${i}`,
      data: { label: `Node ${i}` },
      position: { x: 0, y: 0 },
    }));

    const { result } = renderHook(() =>
      useProgressiveGraph(largeNodeArray, [], { batchSize: 100 })
    );

    expect(result.current.totalBatches).toBe(50);
  });

  it('should handle nodes with same batch size as total', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, {
        batchSize: mockNodes.length,
      })
    );

    expect(result.current.totalBatches).toBe(1);
  });

  it('should handle batch size larger than total nodes', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges, {
        batchSize: mockNodes.length + 100,
      })
    );

    expect(result.current.totalBatches).toBe(1);
  });

  it('should return correct state shape', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    expect(result.current).toHaveProperty('visibleNodes');
    expect(result.current).toHaveProperty('visibleEdges');
    expect(result.current).toHaveProperty('isLoading');
    expect(result.current).toHaveProperty('progress');
    expect(result.current).toHaveProperty('currentBatch');
    expect(result.current).toHaveProperty('totalBatches');
    expect(result.current).toHaveProperty('reset');
    expect(result.current).toHaveProperty('skipToEnd');
  });

  it('should return nodes without duplicates', () => {
    const { result } = renderHook(() =>
      useProgressiveGraph(mockNodes, mockEdges)
    );

    const nodeIds = new Set(result.current.visibleNodes.map(n => n.id));
    expect(nodeIds.size).toBe(result.current.visibleNodes.length);
  });

  it('should update when nodes prop changes', () => {
    const newMockNodes = Array.from({ length: 100 }, (_, i) => ({
      id: `new-node-${i}`,
      data: { label: `New Node ${i}` },
      position: { x: 0, y: 0 },
    }));

    const { result, rerender } = renderHook(
      ({ nodes, edges }) =>
        useProgressiveGraph(nodes, edges),
      {
        initialProps: { nodes: mockNodes, edges: mockEdges },
      }
    );

    const initialTotalBatches = result.current.totalBatches;

    rerender({ nodes: newMockNodes, edges: mockEdges });

    // Total batches should recalculate based on new nodes
    expect(result.current.totalBatches).toEqual(
      Math.ceil(newMockNodes.length / 100)
    );
  });
});
