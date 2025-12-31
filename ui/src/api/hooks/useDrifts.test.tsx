/**
 * useDrifts Hook Tests
 * Tests for drift data React Query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useDrifts, useDrift } from './useDrifts';
import { apiClient } from '../client';
import type { DriftAlert, PaginatedResponse } from '../types';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getDrifts: vi.fn(),
    getDrift: vi.fn(),
  },
}));

// Mock data
const mockDriftAlert: DriftAlert = {
  id: 'drift-123',
  severity: 'high',
  resource_type: 'aws_iam_role',
  resource_name: 'test-role',
  resource_id: 'role-123',
  attribute: 'assume_role_policy',
  old_value: '{}',
  new_value: '{"Version": "2012-10-17"}',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDAI123',
    ARN: 'arn:aws:iam::123456789012:user/test',
    AccountID: '123456789012',
    UserName: 'test-user',
  },
  matched_rules: ['rule1'],
  timestamp: '2024-01-01T00:00:00Z',
  alert_type: 'drift',
};

const mockPaginatedDrifts: PaginatedResponse<DriftAlert> = {
  data: [mockDriftAlert],
  page: 1,
  limit: 10,
  total: 50,
  total_pages: 5,
};

// Helper to create React Query wrapper
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // Disable retries for tests
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useDrifts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch drifts successfully', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockPaginatedDrifts);
      expect(apiClient.getDrifts).toHaveBeenCalledTimes(1);
      expect(apiClient.getDrifts).toHaveBeenCalledWith(undefined);
    });

    it('should fetch drifts with pagination params', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = { page: 2, limit: 20 };
      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
    });

    it('should fetch drifts with severity filter', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = { severity: 'high' };
      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
    });

    it('should fetch drifts with resource_type filter', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = { resource_type: 'aws_iam_role' };
      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
    });

    it('should fetch drifts with all filters combined', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = {
        page: 2,
        limit: 20,
        severity: 'high',
        resource_type: 'aws_iam_role',
      };
      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('API request failed');
      vi.mocked(apiClient.getDrifts).mockRejectedValue(error);

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getDrifts).mockRejectedValue(error);

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key without params', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['drifts', undefined]
      expect(apiClient.getDrifts).toHaveBeenCalledWith(undefined);
    });

    it('should use correct query key with params', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = { page: 1, severity: 'high' };
      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['drifts', params]
      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getDrifts).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const { result } = renderHook(() => useDrifts(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockPaginatedDrifts);
    });
  });
});

describe('useDrift', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch single drift successfully', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockDriftAlert);
      expect(apiClient.getDrift).toHaveBeenCalledTimes(1);
      expect(apiClient.getDrift).toHaveBeenCalledWith('drift-123');
    });

    it('should not fetch when id is empty', () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift(''), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toBeUndefined();
      expect(apiClient.getDrift).not.toHaveBeenCalled();
    });

    it('should fetch when id changes from empty to valid', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result, rerender } = renderHook(
        ({ id }: { id: string }) => useDrift(id),
        {
          wrapper: createWrapper(),
          initialProps: { id: '' },
        }
      );

      expect(apiClient.getDrift).not.toHaveBeenCalled();

      rerender({ id: 'drift-123' });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrift).toHaveBeenCalledWith('drift-123');
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('Drift not found');
      vi.mocked(apiClient.getDrift).mockRejectedValue(error);

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle 404 errors', async () => {
      const error = new Error('HTTP error! status: 404');
      vi.mocked(apiClient.getDrift).mockRejectedValue(error);

      const { result } = renderHook(() => useDrift('nonexistent'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['drifts', 'drift-123']
      expect(apiClient.getDrift).toHaveBeenCalledWith('drift-123');
    });
  });

  describe('Enabled State', () => {
    it('should be disabled when id is empty', () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift(''), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.fetchStatus).toBe('idle');
      expect(apiClient.getDrift).not.toHaveBeenCalled();
    });

    it('should be enabled when id is provided', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrift).toHaveBeenCalled();
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially when enabled', () => {
      vi.mocked(apiClient.getDrift).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result } = renderHook(() => useDrift('drift-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockDriftAlert);
    });
  });
});
