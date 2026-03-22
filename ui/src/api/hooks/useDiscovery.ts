import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import type { DriftSummary, DriftDetectionResult } from '../types';

// Hook to get drift summary
export function useDriftSummary(region: string = 'us-east-1', options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['discovery', 'summary', region],
    queryFn: () => apiClient.getDriftSummary(region) as Promise<DriftSummary>,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
    ...options,
  });
}

// Mutation hook to trigger drift detection manually
export function useTriggerDriftDetection(region: string = 'us-east-1') {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.detectDrift(region) as Promise<DriftDetectionResult>,
    onSuccess: () => {
      // Invalidate and refetch relevant queries
      queryClient.invalidateQueries({ queryKey: ['discovery'] });
    },
  });
}
