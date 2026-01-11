import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';

/**
 * GraphDB Statistics Hook
 */
export const useGraphStats = () => {
  return useQuery({
    queryKey: ['graphdb', 'stats'],
    queryFn: async () => {
      return await apiClient.getGraphStats();
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  });
};

/**
 * Get Node by ID Hook
 */
export const useNode = (nodeId: string, enabled: boolean = true) => {
  return useQuery({
    queryKey: ['graphdb', 'node', nodeId],
    queryFn: async () => {
      return await apiClient.getNodeById(nodeId);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Node Neighbors Hook
 */
export const useNodeNeighbors = (nodeId: string, enabled: boolean = true) => {
  return useQuery({
    queryKey: ['graphdb', 'neighbors', nodeId],
    queryFn: async () => {
      return await apiClient.getNodeNeighbors(nodeId);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Node Relationships Hook
 */
export const useNodeRelationships = (
  nodeId: string,
  direction?: 'incoming' | 'outgoing' | 'all',
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ['graphdb', 'relationships', nodeId, direction],
    queryFn: async () => {
      return await apiClient.getNodeRelationships(nodeId, direction);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Impact Radius Hook
 */
export const useImpactRadius = (
  nodeId: string,
  maxDepth: number = 2,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ['graphdb', 'impact', nodeId, maxDepth],
    queryFn: async () => {
      return await apiClient.getImpactRadius(nodeId, maxDepth);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Node Dependencies Hook
 */
export const useDependencies = (
  nodeId: string,
  depth: number = 5,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ['graphdb', 'dependencies', nodeId, depth],
    queryFn: async () => {
      return await apiClient.getDependencies(nodeId, depth);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Node Dependents Hook
 */
export const useDependents = (
  nodeId: string,
  depth: number = 5,
  enabled: boolean = true
) => {
  return useQuery({
    queryKey: ['graphdb', 'dependents', nodeId, depth],
    queryFn: async () => {
      return await apiClient.getDependents(nodeId, depth);
    },
    enabled: enabled && !!nodeId,
  });
};

/**
 * Get Critical Nodes Hook
 */
export const useCriticalNodes = (min: number = 3) => {
  return useQuery({
    queryKey: ['graphdb', 'critical', min],
    queryFn: async () => {
      return await apiClient.getCriticalNodes(min);
    },
    refetchInterval: 30000, // Refetch every 30 seconds
  });
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
  return useQuery({
    queryKey: ['graphdb', 'match', pattern],
    queryFn: async () => {
      if (!pattern) throw new Error('Pattern is required');
      return await apiClient.matchPattern(pattern);
    },
    enabled: enabled && !!pattern,
  });
};
