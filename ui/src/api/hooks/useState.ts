import { useQuery } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import type { StateMetadata, TerraformResource } from '../types';

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

export const useStateResources = (params?: {
  page?: number;
  limit?: number;
}) => {
  return useQuery({
    queryKey: ['state', 'resources', params],
    queryFn: async () => {
      const data = await apiClient.getStateResources(params);
      return data as PaginatedResponse<TerraformResource>;
    },
  });
};
