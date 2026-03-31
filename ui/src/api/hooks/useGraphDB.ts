import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import { useToastStore } from '../../stores/toastStore';

/**
 * Get Node by ID Hook
 */
export const useNode = (nodeId: string, enabled: boolean = true) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'node', nodeId],
    queryFn: async () => {
      return await apiClient.getNodeById(nodeId);
    },
    enabled: enabled && !!nodeId,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch node';
      addToast({ type: 'error', title: 'Node Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

/**
 * Get Node Neighbors Hook
 */
export const useNodeNeighbors = (nodeId: string, enabled: boolean = true) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'neighbors', nodeId],
    queryFn: async () => {
      return await apiClient.getNodeNeighbors(nodeId);
    },
    enabled: enabled && !!nodeId,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch node neighbors';
      addToast({ type: 'error', title: 'Neighbors Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

/**
 * Get Impact Radius Hook
 */
export const useImpactRadius = (
  nodeId: string,
  maxDepth: number = 2,
  enabled: boolean = true
) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'impact', nodeId, maxDepth],
    queryFn: async () => {
      return await apiClient.getImpactRadius(nodeId, maxDepth);
    },
    enabled: enabled && !!nodeId,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch impact radius';
      addToast({ type: 'error', title: 'Impact Radius Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

/**
 * Get Node Dependencies Hook
 */
export const useDependencies = (
  nodeId: string,
  depth: number = 5,
  enabled: boolean = true
) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'dependencies', nodeId, depth],
    queryFn: async () => {
      return await apiClient.getDependencies(nodeId, depth);
    },
    enabled: enabled && !!nodeId,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch dependencies';
      addToast({ type: 'error', title: 'Dependencies Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

/**
 * Get Node Dependents Hook
 */
export const useDependents = (
  nodeId: string,
  depth: number = 5,
  enabled: boolean = true
) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'dependents', nodeId, depth],
    queryFn: async () => {
      return await apiClient.getDependents(nodeId, depth);
    },
    enabled: enabled && !!nodeId,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch dependents';
      addToast({ type: 'error', title: 'Dependents Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

/**
 * Pattern Matching Hook
 * Note: This is a POST request, so we use enabled pattern
 */
export const usePatternMatch = (
  pattern: {
    start_labels: string[];
    rel_type: string;
    end_labels: string[];
    end_filter: Record<string, unknown>;
  } | null,
  enabled: boolean = true
) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graphdb', 'match', pattern],
    queryFn: async () => {
      if (!pattern) throw new Error('Pattern is required');
      return await apiClient.matchPattern(pattern);
    },
    enabled: enabled && !!pattern,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to match pattern';
      addToast({ type: 'error', title: 'Pattern Match Error', message });
    }
  }, [query.error, addToast]);

  return query;
};
