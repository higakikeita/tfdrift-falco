import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import { useToastStore } from '../../stores/toastStore';
import type { Stats } from '../types';

// Hook to fetch aggregate dashboard statistics (drifts, events, severity breakdown).
export const useStats = () => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['stats'],
    queryFn: async () => (await apiClient.getStats()) as Stats,
    // Poll in step with the drift feed (useDrifts) so the KPI counters and the
    // Recent Drift Events list update together as drifts arrive.
    refetchInterval: 5000,
  });

  useEffect(() => {
    if (query.error) {
      const message =
        query.error instanceof Error ? query.error.message : 'Failed to fetch stats';
      addToast({ type: 'error', title: 'Stats Error', message });
    }
  }, [query.error, addToast]);

  return query;
};
