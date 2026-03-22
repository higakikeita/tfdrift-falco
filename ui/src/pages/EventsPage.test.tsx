import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
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

import EventsPage from './EventsPage';

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
