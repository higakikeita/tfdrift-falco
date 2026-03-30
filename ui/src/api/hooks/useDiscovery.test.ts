/**
 * useDiscovery Hook Tests
 * Tests for drift discovery React Query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useDriftSummary, useTriggerDriftDetection } from './useDiscovery';
import { apiClient } from '../client';
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';
import type { DriftSummary, DriftDetectionResult } from '../types';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getDriftSummary: vi.fn(),
    detectDrift: vi.fn(),
  },
}));

const mockDriftSummary: DriftSummary = {
  region: 'us-east-1',
  timestamp: '2024-01-01T00:00:00Z',
  counts: {
    terraform_resources: 50,
    aws_resources: 60,
    unmanaged: 10,
    missing: 2,
    modified: 3,
  },
  breakdown: {
    unmanaged_by_type: {
      'aws_iam_role': 5,
      'aws_s3_bucket': 3,
      'aws_lambda_function': 2,
    },
    missing_by_type: {
      'aws_ec2_instance': 2,
    },
    modified_by_type: {
      'aws_security_group': 2,
      'aws_iam_policy': 1,
    },
  },
};

const mockDriftDetectionResult: DriftDetectionResult = {
  region: 'us-east-1',
  timestamp: '2024-01-01T00:00:00Z',
  summary: {
    terraform_resources: 50,
    aws_resources: 60,
    unmanaged: 10,
    missing: 2,
    modified: 3,
  },
  results: {
    unmanaged_resources: [],
    missing_resources: [],
    modified_resources: [],
  },
};

describe('useDriftSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch drift summary successfully', async () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockDriftSummary);
      expect(apiClient.getDriftSummary).toHaveBeenCalledTimes(1);
      expect(apiClient.getDriftSummary).toHaveBeenCalledWith('us-east-1');
    });

    it('should fetch drift summary with custom region', async () => {
      const customSummary = { ...mockDriftSummary, region: 'us-west-2' };
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(customSummary);

      const { result } = renderHook(() => useDriftSummary('us-west-2'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.region).toBe('us-west-2');
      expect(apiClient.getDriftSummary).toHaveBeenCalledWith('us-west-2');
    });

    it('should fetch drift summary with different unmanaged resources', async () => {
      const customSummary = {
        ...mockDriftSummary,
        counts: {
          ...mockDriftSummary.counts,
          unmanaged: 25,
        },
        breakdown: {
          ...mockDriftSummary.breakdown,
          unmanaged_by_type: {
            'aws_iam_role': 15,
            's3_bucket': 10,
          },
        },
      };
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(customSummary);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.counts.unmanaged).toBe(25);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('Failed to fetch drift summary');
      vi.mocked(apiClient.getDriftSummary).mockRejectedValue(error);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getDriftSummary).mockRejectedValue(error);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key with default region', async () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDriftSummary).toHaveBeenCalledWith('us-east-1');
    });

    it('should use correct query key with custom region', async () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary('eu-west-1'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDriftSummary).toHaveBeenCalledWith('eu-west-1');
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getDriftSummary).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockDriftSummary);
    });
  });

  describe('Options', () => {
    it('should respect enabled option', () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary('us-east-1', { enabled: false }), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toBeUndefined();
      expect(apiClient.getDriftSummary).not.toHaveBeenCalled();
    });

    it('should enable query when enabled option is true', async () => {
      vi.mocked(apiClient.getDriftSummary).mockResolvedValue(mockDriftSummary);

      const { result } = renderHook(() => useDriftSummary('us-east-1', { enabled: true }), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDriftSummary).toHaveBeenCalled();
    });
  });
});

describe('useTriggerDriftDetection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Mutation', () => {
    it('should trigger drift detection successfully', async () => {
      vi.mocked(apiClient.detectDrift).mockResolvedValue(mockDriftDetectionResult);

      const { result } = renderHook(() => useTriggerDriftDetection(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isPending).toBe(false);

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockDriftDetectionResult);
      expect(apiClient.detectDrift).toHaveBeenCalledTimes(1);
      expect(apiClient.detectDrift).toHaveBeenCalledWith('us-east-1');
    });

    it('should trigger drift detection with custom region', async () => {
      vi.mocked(apiClient.detectDrift).mockResolvedValue(mockDriftDetectionResult);

      const { result } = renderHook(() => useTriggerDriftDetection('us-west-2'), {
        wrapper: createQueryClientWrapper(),
      });

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.detectDrift).toHaveBeenCalledWith('us-west-2');
    });

    it('should show pending state during mutation', async () => {
      vi.mocked(apiClient.detectDrift).mockImplementation(
        () => new Promise((resolve) => {
          setTimeout(() => resolve(mockDriftDetectionResult), 100);
        })
      );

      const { result } = renderHook(() => useTriggerDriftDetection(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isPending).toBe(false);

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isPending).toBe(true);
      });

      await waitFor(() => {
        expect(result.current.isPending).toBe(false);
      });

      expect(result.current.isSuccess).toBe(true);
    });
  });

  describe('Error Handling', () => {
    it('should handle drift detection errors', async () => {
      const error = new Error('Drift detection failed');
      vi.mocked(apiClient.detectDrift).mockRejectedValue(error);

      const { result } = renderHook(() => useTriggerDriftDetection(), {
        wrapper: createQueryClientWrapper(),
      });

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle timeout errors', async () => {
      const error = new Error('Request timeout');
      vi.mocked(apiClient.detectDrift).mockRejectedValue(error);

      const { result } = renderHook(() => useTriggerDriftDetection(), {
        wrapper: createQueryClientWrapper(),
      });

      result.current.mutate();

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });
});
