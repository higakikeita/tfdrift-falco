import { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import { useToastStore } from '../../stores/toastStore';
import type { DriftSummary, DriftDetectionResult } from '../types';

// Hook to get drift summary
export function useDriftSummary(region: string = 'us-east-1', options?: { enabled?: boolean }) {
  const addToast = useToastStore((state) => state.addToast);

  const query = useQuery({
    queryKey: ['discovery', 'summary', region],
    queryFn: () => apiClient.getDriftSummary(region) as Promise<DriftSummary>,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
    ...options,
  });

  useEffect(() => {
    if (query.error) {
      const message = query.error instanceof Error
        ? query.error.message
        : 'Failed to fetch drift summary';
      addToast({ type: 'error', title: 'Drift Summary Error', message });
    }
  }, [query.error, addToast]);

  return query;
}

// Mutation hook to trigger drift detection manually
export function useTriggerDriftDetection(region: string = 'us-east-1') {
  const queryClient = useQueryClient();
  const addToast = useToastStore((state) => state.addToast);

  return useMutation({
    mutationFn: () => apiClient.detectDrift(region) as Promise<DriftDetectionResult>,
    onSuccess: () => {
      // Invalidate and refetch relevant queries
      queryClient.invalidateQueries({ queryKey: ['discovery'] });
      addToast({ type: 'success', title: 'Drift Detection Started' });
    },
    onError: (error) => {
      const message = error instanceof Error ? error.message : 'Failed to trigger drift detection';
      addToast({ type: 'error', title: 'Detection Error', message });
    },
  });
}
