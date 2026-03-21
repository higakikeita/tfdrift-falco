import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import type { FalcoEvent, EventStatus } from '../types';

interface EventsParams {
  page?: number;
  limit?: number;
  severity?: string;
  provider?: string;
  status?: string;
  search?: string;
  from?: string;
  to?: string;
  sort?: string;
  order?: string;
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

export const useUpdateEventStatus = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      id,
      status,
      reason,
    }: {
      id: string;
      status: EventStatus;
      reason?: string;
    }) => {
      return apiClient.updateEventStatus(id, status, reason);
    },
    onSuccess: () => {
      // Invalidate events queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });
};
