import { useQuery } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import type { DriftAlert } from '../types';

interface DriftsParams {
  page?: number;
  limit?: number;
  severity?: string;
  resource_type?: string;
}

export const useDrifts = (params?: DriftsParams) => {
  return useQuery({
    queryKey: ['drifts', params],
    queryFn: async () => {
      const data = await apiClient.getDrifts(params);
      return data as PaginatedResponse<DriftAlert>;
    },
    // Refetch every 30 seconds
    refetchInterval: 30000,
  });
};

export const useDrift = (id: string) => {
  return useQuery({
    queryKey: ['drifts', id],
    queryFn: async () => {
      const data = await apiClient.getDrift(id);
      return data as DriftAlert;
    },
    enabled: !!id,
  });
};
