/**
 * useEvents Hook Tests
 * Tests for Falco events React Query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useEvents, useEvent } from './useEvents';
import { apiClient } from '../client';
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';
import {
  mockFalcoEvent,
  mockPaginatedEvents,
  createMockEvent,
  createMockGCPEvent,
} from '../../__tests__/fixtures/eventsFixtures';

// Mock API client
vi.mock('../client', () => ({
  apiClient: {
    getEvents: vi.fn(),
    getEvent: vi.fn(),
  },
}));

describe('useEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch events successfully', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockPaginatedEvents);
      expect(apiClient.getEvents).toHaveBeenCalledTimes(1);
      expect(apiClient.getEvents).toHaveBeenCalledWith(undefined);
    });

    it('should fetch events with pagination params', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const params = { page: 2, limit: 20 };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
    });

    it('should fetch events with severity filter', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const params = { severity: 'high' };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
    });

    it('should fetch events with provider filter', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const params = { provider: 'aws' };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
    });

    it('should fetch events with all filters combined', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const params = {
        page: 2,
        limit: 20,
        severity: 'high',
        provider: 'aws',
      };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
    });

    it('should fetch events with GCP provider', async () => {
      const gcpEvents = {
        ...mockPaginatedEvents,
        data: [createMockGCPEvent()],
      };
      vi.mocked(apiClient.getEvents).mockResolvedValue(gcpEvents);

      const params = { provider: 'gcp' };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
      expect(result.current.data?.data[0].provider).toBe('gcp');
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('API request failed');
      vi.mocked(apiClient.getEvents).mockRejectedValue(error);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      vi.mocked(apiClient.getEvents).mockRejectedValue(error);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle timeout errors', async () => {
      const error = new Error('Request timeout');
      vi.mocked(apiClient.getEvents).mockRejectedValue(error);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key without params', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['events', undefined]
      expect(apiClient.getEvents).toHaveBeenCalledWith(undefined);
    });

    it('should use correct query key with params', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const params = { page: 1, severity: 'high' };
      const { result } = renderHook(() => useEvents(params), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['events', params]
      expect(apiClient.getEvents).toHaveBeenCalledWith(params);
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially', () => {
      vi.mocked(apiClient.getEvents).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getEvents).mockResolvedValue(mockPaginatedEvents);

      const { result } = renderHook(() => useEvents(), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockPaginatedEvents);
    });
  });

  describe('Pagination', () => {
    it('should handle multiple pages', async () => {
      const page1 = { ...mockPaginatedEvents, page: 1 };

      vi.mocked(apiClient.getEvents).mockResolvedValueOnce(page1);

      const { result, rerender } = renderHook(
        ({ page }: { page: number }) => useEvents({ page }),
        {
          wrapper: createQueryClientWrapper(),
          initialProps: { page: 1 },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.page).toBe(1);

      // Switch to page 2
      const page2 = {
        ...mockPaginatedEvents,
        page: 2,
        data: [createMockEvent({ id: 'event-456' })]
      };
      vi.mocked(apiClient.getEvents).mockResolvedValueOnce(page2);
      rerender({ page: 2 });

      await waitFor(() => {
        expect(result.current.data?.page).toBe(2);
      });
    });
  });
});

describe('useEvent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Fetching', () => {
    it('should fetch single event successfully', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockFalcoEvent);
      expect(apiClient.getEvent).toHaveBeenCalledTimes(1);
      expect(apiClient.getEvent).toHaveBeenCalledWith('event-123');
    });

    it('should not fetch when id is empty', () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent(''), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toBeUndefined();
      expect(apiClient.getEvent).not.toHaveBeenCalled();
    });

    it('should fetch when id changes from empty to valid', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result, rerender } = renderHook(
        ({ id }: { id: string }) => useEvent(id),
        {
          wrapper: createQueryClientWrapper(),
          initialProps: { id: '' },
        }
      );

      expect(apiClient.getEvent).not.toHaveBeenCalled();

      rerender({ id: 'event-123' });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvent).toHaveBeenCalledWith('event-123');
    });

    it('should fetch different event when id changes', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValueOnce(mockFalcoEvent);

      const { result, rerender } = renderHook(
        ({ id }: { id: string }) => useEvent(id),
        {
          wrapper: createQueryClientWrapper(),
          initialProps: { id: 'event-123' },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.id).toBe('event-123');

      // Switch to different event
      const anotherEvent = createMockEvent({ id: 'event-456' });
      vi.mocked(apiClient.getEvent).mockResolvedValueOnce(anotherEvent);
      rerender({ id: 'event-456' });

      await waitFor(() => {
        expect(result.current.data?.id).toBe('event-456');
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors', async () => {
      const error = new Error('Event not found');
      vi.mocked(apiClient.getEvent).mockRejectedValue(error);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle 404 errors', async () => {
      const error = new Error('HTTP error! status: 404');
      vi.mocked(apiClient.getEvent).mockRejectedValue(error);

      const { result } = renderHook(() => useEvent('nonexistent'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });

    it('should handle server errors', async () => {
      const error = new Error('HTTP error! status: 500');
      vi.mocked(apiClient.getEvent).mockRejectedValue(error);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(error);
    });
  });

  describe('Query Key', () => {
    it('should use correct query key', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Query key should be ['events', 'event-123']
      expect(apiClient.getEvent).toHaveBeenCalledWith('event-123');
    });
  });

  describe('Enabled State', () => {
    it('should be disabled when id is empty', () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent(''), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.fetchStatus).toBe('idle');
      expect(apiClient.getEvent).not.toHaveBeenCalled();
    });

    it('should be enabled when id is provided', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getEvent).toHaveBeenCalled();
    });
  });

  describe('Loading States', () => {
    it('should show loading state initially when enabled', () => {
      vi.mocked(apiClient.getEvent).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      );

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
    });

    it('should transition to success state', async () => {
      vi.mocked(apiClient.getEvent).mockResolvedValue(mockFalcoEvent);

      const { result } = renderHook(() => useEvent('event-123'), {
        wrapper: createQueryClientWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toEqual(mockFalcoEvent);
    });
  });
});
