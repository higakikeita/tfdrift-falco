/**
 * Progressive Graph Loading Hook
 * Loads and renders large graphs in batches for better performance
 */

import { useState, useEffect, useCallback, useMemo } from 'react';
import { Node, Edge } from 'reactflow';
import { chunkArray } from '../utils/memoryOptimization';

interface ProgressiveGraphOptions {
  batchSize?: number; // Number of nodes to load per batch
  batchDelay?: number; // Delay between batches in ms
  priorityNodeIds?: string[]; // Load these nodes first
}

interface ProgressiveGraphState<T = any> {
  visibleNodes: Node<T>[];
  visibleEdges: Edge[];
  isLoading: boolean;
  progress: number; // 0-100
  currentBatch: number;
  totalBatches: number;
}

export function useProgressiveGraph<T = any>(
  allNodes: Node<T>[],
  allEdges: Edge[],
  options: ProgressiveGraphOptions = {}
): ProgressiveGraphState<T> {
  const {
    batchSize = 100,
    batchDelay = 50,
    priorityNodeIds = [],
  } = options;

  const [currentBatch, setCurrentBatch] = useState(0);
  const [isLoading, setIsLoading] = useState(false);

  // Calculate node batches with priority handling
  const nodeBatches = useMemo(() => {
    if (allNodes.length === 0) return [];

    // Separate priority nodes from regular nodes
    const priorityNodes = allNodes.filter(node =>
      priorityNodeIds.includes(node.id)
    );
    const regularNodes = allNodes.filter(node =>
      !priorityNodeIds.includes(node.id)
    );

    // Chunk regular nodes
    const regularBatches = chunkArray(regularNodes, batchSize);

    // If we have priority nodes, make them the first batch
    if (priorityNodes.length > 0) {
      return [priorityNodes, ...regularBatches];
    }

    return regularBatches;
  }, [allNodes, batchSize, priorityNodeIds]);

  const totalBatches = nodeBatches.length;

  // Get visible nodes up to current batch
  const visibleNodes = useMemo(() => {
    if (currentBatch === 0) return [];

    return nodeBatches
      .slice(0, currentBatch)
      .flat();
  }, [nodeBatches, currentBatch]);

  // Get visible node IDs for edge filtering
  const visibleNodeIds = useMemo(() => {
    return new Set(visibleNodes.map(node => node.id));
  }, [visibleNodes]);

  // Filter edges to only show those connecting visible nodes
  const visibleEdges = useMemo(() => {
    return allEdges.filter(edge =>
      visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target)
    );
  }, [allEdges, visibleNodeIds]);

  // Calculate progress percentage
  const progress = useMemo(() => {
    if (totalBatches === 0) return 100;
    return Math.round((currentBatch / totalBatches) * 100);
  }, [currentBatch, totalBatches]);

  // Progressive loading effect
  useEffect(() => {
    if (allNodes.length === 0) {
      setCurrentBatch(0);
      setIsLoading(false);
      return;
    }

    // Start loading if not already started
    if (currentBatch === 0) {
      setIsLoading(true);
      setCurrentBatch(1);
      return;
    }

    // Continue loading next batches
    if (currentBatch < totalBatches) {
      const timer = setTimeout(() => {
        requestAnimationFrame(() => {
          setCurrentBatch(prev => prev + 1);
        });
      }, batchDelay);

      return () => clearTimeout(timer);
    }

    // Finished loading
    if (currentBatch === totalBatches) {
      setIsLoading(false);
    }
  }, [allNodes.length, currentBatch, totalBatches, batchDelay]);

  // Reset when nodes change
  useEffect(() => {
    setCurrentBatch(0);
    setIsLoading(false);
  }, [allNodes]);

  const reset = useCallback(() => {
    setCurrentBatch(0);
    setIsLoading(false);
  }, []);

  const skipToEnd = useCallback(() => {
    setCurrentBatch(totalBatches);
    setIsLoading(false);
  }, [totalBatches]);

  return {
    visibleNodes,
    visibleEdges,
    isLoading,
    progress,
    currentBatch,
    totalBatches,
    reset,
    skipToEnd,
  } as ProgressiveGraphState<T> & { reset: () => void; skipToEnd: () => void };
}

/**
 * Hook for progressive edge loading
 * Useful when edges need to be loaded separately from nodes
 */
export function useProgressiveEdges(
  allEdges: Edge[],
  visibleNodeIds: Set<string>,
  batchSize: number = 200
): Edge[] {
  const [loadedCount, setLoadedCount] = useState(0);

  // Filter relevant edges
  const relevantEdges = useMemo(() => {
    return allEdges.filter(edge =>
      visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target)
    );
  }, [allEdges, visibleNodeIds]);

  // Progressive loading
  useEffect(() => {
    if (loadedCount < relevantEdges.length) {
      const timer = setTimeout(() => {
        requestAnimationFrame(() => {
          setLoadedCount(prev => Math.min(prev + batchSize, relevantEdges.length));
        });
      }, 16); // ~60fps

      return () => clearTimeout(timer);
    }
  }, [loadedCount, relevantEdges.length, batchSize]);

  // Reset when relevant edges change
  useEffect(() => {
    setLoadedCount(0);
  }, [relevantEdges]);

  return useMemo(() => {
    return relevantEdges.slice(0, loadedCount);
  }, [relevantEdges, loadedCount]);
}

/**
 * Hook for intelligent viewport-based loading
 * Only loads nodes within or near the current viewport
 */
export function useViewportBasedLoading<T = any>(
  allNodes: Node<T>[],
  allEdges: Edge[],
  viewport: { x: number; y: number; zoom: number },
  viewportDimensions: { width: number; height: number }
): { visibleNodes: Node<T>[]; visibleEdges: Edge[] } {
  // Calculate viewport bounds with padding
  const viewportBounds = useMemo(() => {
    const padding = 500; // Extra padding for smooth scrolling
    return {
      minX: -viewport.x / viewport.zoom - padding,
      maxX: (-viewport.x + viewportDimensions.width) / viewport.zoom + padding,
      minY: -viewport.y / viewport.zoom - padding,
      maxY: (-viewport.y + viewportDimensions.height) / viewport.zoom + padding,
    };
  }, [viewport, viewportDimensions]);

  // Filter nodes within viewport
  const visibleNodes = useMemo(() => {
    return allNodes.filter(node => {
      const x = node.position.x;
      const y = node.position.y;
      return (
        x >= viewportBounds.minX &&
        x <= viewportBounds.maxX &&
        y >= viewportBounds.minY &&
        y <= viewportBounds.maxY
      );
    });
  }, [allNodes, viewportBounds]);

  // Get visible node IDs
  const visibleNodeIds = useMemo(() => {
    return new Set(visibleNodes.map(node => node.id));
  }, [visibleNodes]);

  // Filter edges connecting visible nodes
  const visibleEdges = useMemo(() => {
    return allEdges.filter(edge =>
      visibleNodeIds.has(edge.source) && visibleNodeIds.has(edge.target)
    );
  }, [allEdges, visibleNodeIds]);

  return { visibleNodes, visibleEdges };
}
