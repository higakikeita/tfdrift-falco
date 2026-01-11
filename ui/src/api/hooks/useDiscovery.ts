import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import type { DriftSummary, DriftDetectionResult } from '../types';

// Query keys
export const discoveryKeys = {
  all: ['discovery'] as const,
  summary: (region: string) => [...discoveryKeys.all, 'summary', region] as const,
  drift: (region: string) => [...discoveryKeys.all, 'drift', region] as const,
  scan: (region: string) => [...discoveryKeys.all, 'scan', region] as const,
};

// Hook to get drift summary
export function useDriftSummary(region: string = 'us-east-1', options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: discoveryKeys.summary(region),
    queryFn: () => apiClient.getDriftSummary(region) as Promise<DriftSummary>,
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
    ...options,
  });
}

// Hook to detect full drift
export function useDriftDetection(region: string = 'us-east-1', options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: discoveryKeys.drift(region),
    queryFn: () => apiClient.detectDrift(region) as Promise<DriftDetectionResult>,
    staleTime: 5 * 60 * 1000,
    refetchInterval: 5 * 60 * 1000,
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
      queryClient.invalidateQueries({ queryKey: discoveryKeys.summary(region) });
      queryClient.invalidateQueries({ queryKey: discoveryKeys.drift(region) });
    },
  });
}

// Mutation hook to scan AWS resources
export function useScanAWSResources(region: string = 'us-east-1') {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.scanAWSResources(region),
    onSuccess: () => {
      // Invalidate discovery queries
      queryClient.invalidateQueries({ queryKey: discoveryKeys.all });
    },
  });
}
