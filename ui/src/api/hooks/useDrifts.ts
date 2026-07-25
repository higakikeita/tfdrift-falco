import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiClient, type PaginatedResponse } from '../client';
import { useToastStore } from '../../stores/toastStore';
import type { DriftAlert } from '../types';

interface DriftsParams {
  page?: number;
  limit?: number;
  severity?: string;
  provider?: string;
  search?: string;
  sort?: string;
  order?: string;
}

// useDrifts fetches detected drifts from /api/v1/drifts. The drift list views
// (Dashboard "Recent Drift Events", Events page) previously read /events, which
// only holds raw CloudTrail events and is empty — so drifts never appeared even
// though the KPIs (from /stats) counted them (#364). Polls frequently so a new
// drift shows up in the feed within a few seconds of detection.
export const useDrifts = (params?: DriftsParams) => {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['drifts', params],
    queryFn: async () => {
      const data = await apiClient.getDrifts(params);
      return data as PaginatedResponse<DriftAlert>;
    },
    refetchInterval: 5000,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error ? query.error.message : 'Failed to fetch drifts';
      addToast({ type: 'error', title: 'Drifts Error', message });
    }
  }, [query.error, addToast]);

  return query;
};
