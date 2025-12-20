import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import type { CytoscapeElements } from '../types';

export const useGraph = () => {
  return useQuery({
    queryKey: ['graph'],
    queryFn: async () => {
      const data = await apiClient.getGraph();
      return data as CytoscapeElements;
    },
    // Refetch every 30 seconds for real-time updates
    refetchInterval: 30000,
  });
};

export const useNodes = (params?: { page?: number; limit?: number }) => {
  return useQuery({
    queryKey: ['graph', 'nodes', params],
    queryFn: async () => {
      return await apiClient.getNodes(params);
    },
  });
};

export const useEdges = (params?: { page?: number; limit?: number }) => {
  return useQuery({
    queryKey: ['graph', 'edges', params],
    queryFn: async () => {
      return await apiClient.getEdges(params);
    },
  });
};
