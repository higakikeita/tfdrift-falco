import { useQuery } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import type { FalcoEvent } from '../types';

interface EventsParams {
  page?: number;
  limit?: number;
  severity?: string;
  provider?: string;
}

export const useEvents = (params?: EventsParams) => {
  return useQuery({
    queryKey: ['events', params],
    queryFn: async () => {
      const data = await apiClient.getEvents(params);
      return data as PaginatedResponse<FalcoEvent>;
    },
    // Refetch every 30 seconds
    refetchInterval: 30000,
  });
};

export const useEvent = (id: string) => {
  return useQuery({
    queryKey: ['events', id],
    queryFn: async () => {
      const data = await apiClient.getEvent(id);
      return data as FalcoEvent;
    },
    enabled: !!id,
  });
};
