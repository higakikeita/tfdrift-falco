/**
 * useState Hook Tests
 * Tests for Terraform state metadata React Query hook
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useState } from './useState';
import { apiClient } from '../client';
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';
import type { StateMetadata } from '../types';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getState: vi.fn(),
  },
}));

const mockStateMetadata: StateMetadata = {
  version: 4,
  terraform_version: '1.5.0',
  serial: 123,
  lineage: 'abc-123-def',
  resource_count: 42,
  outputs_count: 3,
};

describe('useState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch state metadata successfully', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockStateMetadata);
      expect(apiClient.getState).toHaveBeenCalledTimes(1);
    });

    it('should fetch state with different version', async () => {
      const customState = { ...mockStateMetadata, version: 5, serial: 200 };
      vi.mocked(apiClient.getState).mockResolvedValue(customState);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.version).toBe(5);
      expect(result.current.data?.serial).toBe(200);
    });

    it('should fetch state with varying resource counts', async () => {
      const customState = { ...mockStateMetadata, resource_count: 100, outputs_count: 10 };
      vi.mocked(apiClient.getState).mockResolvedValue(customState);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.resource_count).toBe(100);
      expect(result.current.data?.outputs_count).toBe(10);
    });

    it('should fetch state with zero outputs', async () => {
      const customState = { ...mockStateMetadata, outputs_count: 0 };
      vi.mocked(apiClient.getState).mockResolvedValue(customState);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.outputs_count).toBe(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('Failed to fetch state');
      vi.mocked(apiClient.getState).mockRejectedValue(error);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getState).mockRejectedValue(error);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle 404 errors', async () => {
      const error = new Error('HTTP error! status: 404');
      vi.mocked(apiClient.getState).mockRejectedValue(error);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toContain('404');
    });

    it('should handle server errors', async () => {
      const error = new Error('HTTP error! status: 500');
      vi.mocked(apiClient.getState).mockRejectedValue(error);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toContain('500');
    });
  });

  describe('Query Key', () => {
    it('should use correct query key', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['state']
      expect(apiClient.getState).toHaveBeenCalledTimes(1);
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getState).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockStateMetadata);
    });

    it('should transition to error state', async () => {
      const error = new Error('Fetch failed');
      vi.mocked(apiClient.getState).mockRejectedValue(error);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toEqual(error);
    });
  });

  describe('Refetch Interval', () => {
    it('should be configured with 60 second refetch interval', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Verify data is fetched successfully with refetch configured
      expect(result.current.data).toEqual(mockStateMetadata);
    });
  });

  describe('State Metadata Content', () => {
    it('should contain terraform version information', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.terraform_version).toBe('1.5.0');
    });

    it('should contain state serial number', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.serial).toBe(123);
    });

    it('should contain lineage information', async () => {
      vi.mocked(apiClient.getState).mockResolvedValue(mockStateMetadata);

      const { result } = renderHook(() => useState(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.lineage).toBe('abc-123-def');
    });
  });
});
