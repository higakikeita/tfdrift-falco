import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import { useToastStore } from '../../stores/toastStore';
import type { CytoscapeElements } from '../types';

export const useGraph = () => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graph'],
    queryFn: async () => {
      const data = await apiClient.getGraph();
      return data as CytoscapeElements;
    },
    // Refetch every 30 seconds for real-time updates
    refetchInterval: 30000,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch graph data';
      addToast({ type: 'error', title: 'Graph Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

export const useNodes = (params?: { page?: number; limit?: number }) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graph', 'nodes', params],
    queryFn: async () => {
      return await apiClient.getNodes(params);
    },
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch nodes';
      addToast({ type: 'error', title: 'Nodes Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

export const useEdges = (params?: { page?: number; limit?: number }) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['graph', 'edges', params],
    queryFn: async () => {
      return await apiClient.getEdges(params);
    },
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch edges';
      addToast({ type: 'error', title: 'Edges Error', message });
    }
  }, [query.error, addToast]);

  return query;
};
