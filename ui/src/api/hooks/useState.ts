import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../client';
import type { StateMetadata } from '../types';

export const useState = () => {
  return useQuery({
    queryKey: ['state'],
    queryFn: async () => {
      const data = await apiClient.getState();
      return data as StateMetadata;
    },
    // Refetch every 60 seconds (state changes less frequently)
    refetchInterval: 60000,
  });
};
