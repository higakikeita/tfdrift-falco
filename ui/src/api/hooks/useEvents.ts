import { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import { useToastStore } from '../../stores/toastStore';
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
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['events', params],
    queryFn: async () => {
      const data = await apiClient.getEvents(params);
      return data as PaginatedResponse<FalcoEvent>;
    },
    // Refetch every 30 seconds
    refetchInterval: 30000,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch events';
      addToast({ type: 'error', title: 'Events Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

export const useEvent = (id: string) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['events', id],
    queryFn: async () => {
      const data = await apiClient.getEvent(id);
      return data as FalcoEvent;
    },
    enabled: !!id,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch event';
      addToast({ type: 'error', title: 'Event Error', message });
    }
  }, [query.error, addToast]);

  return query;
};

export const useUpdateEventStatus = () => {
  const queryClient = useQueryClient();
  const addToast = useToastStore((state) => state.addToast);

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
      addToast({ type: 'success', title: 'Event Status Updated' });
    },
    onError: (error) => {
      const message = error instanceof Error ? error.message : 'Failed to update event status';
      addToast({ type: 'error', title: 'Status Update Failed', message });
    },
  });
};
