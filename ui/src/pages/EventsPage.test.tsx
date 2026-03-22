import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

vi.mock('lucide-react', () => ({
  ArrowUpDown: () => <div data-testid="icon-ArrowUpDown" />,
  ChevronLeft: () => <div data-testid="icon-ChevronLeft" />,
  ChevronRight: () => <div data-testid="icon-ChevronRight" />,
  X: () => <div data-testid="icon-X" />,
  Search: () => <div data-testid="icon-Search" />,
  Clock: () => <div data-testid="icon-Clock" />,
  AlertCircle: () => <div data-testid="icon-AlertCircle" />,
  CheckCircle: () => <div data-testid="icon-CheckCircle" />,
  CheckCircle2: () => <div data-testid="icon-CheckCircle2" />,
  Circle: () => <div data-testid="icon-Circle" />,
  EyeOff: () => <div data-testid="icon-EyeOff" />,
  Shield: () => <div data-testid="icon-Shield" />,
  User: () => <div data-testid="icon-User" />,
  MapPin: () => <div data-testid="icon-MapPin" />,
  Activity: () => <div data-testid="icon-Activity" />,
}));

vi.mock('../api/client', () => ({
  apiClient: {
    getEvents: vi.fn().mockResolvedValue({
      success: true,
      data: { data: [], page: 1, limit: 20, total: 0, total_pages: 0 },
    }),
    getEvent: vi.fn().mockResolvedValue({ success: true, data: null }),
    updateEventStatus: vi.fn().mockResolvedValue({ success: true }),
  },
}));

vi.mock('../api/sse', () => ({
  sseClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

import { EventsPage } from './EventsPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, gcTime: 0 } },
});

describe('EventsPage', () => {
  it('should render without crashing', () => {
    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <EventsPage />
        </BrowserRouter>
      </QueryClientProvider>
    );
    expect(container.firstChild).toBeTruthy();
  });
});
