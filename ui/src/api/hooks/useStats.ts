import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import type { Stats } from '../types';

export const useStats = () => {
  return useQuery({
    queryKey: ['stats'],
    queryFn: async () => {
      const data = await apiClient.getStats();
      return data as Stats;
    },
    // Refetch every 30 seconds
    refetchInterval: 30000,
  });
};
