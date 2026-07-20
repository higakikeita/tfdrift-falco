import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

vi.mock('lucide-react', () => ({
  Activity: () => <div data-testid="icon-Activity" />,
  AlertTriangle: () => <div data-testid="icon-AlertTriangle" />,
  Database: () => <div data-testid="icon-Database" />,
  Cloud: () => <div data-testid="icon-Cloud" />,
}));

vi.mock('../api/client', () => ({
  apiClient: {
    getStats: vi.fn(),
    getEvents: vi.fn(),
  },
}));

import { apiClient } from '../api/client';

vi.mock('../api/sse', () => ({
  sseClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

import { DashboardPage } from './DashboardPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, gcTime: 0 } },
});

const renderPage = () => {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <DashboardPage />
      </BrowserRouter>
    </QueryClientProvider>
  );
};

describe('DashboardPage', () => {
  beforeEach(() => {
    // apiClient methods return the unwrapped `data` (request() strips the envelope)
    vi.mocked(apiClient.getStats).mockResolvedValue({
      graph: { total_nodes: 10, total_edges: 5 },
      drifts: { total: 3, severity_counts: { critical: 1, high: 1, medium: 1 } },
      events: { total: 15 },
    } as never);
    vi.mocked(apiClient.getEvents).mockResolvedValue({
      data: [
        {
          id: 'e1',
          event_name: 'ModifyInstanceAttribute',
          resource_type: 'aws_instance',
          resource_id: 'i-123',
          user_identity: { UserName: 'alice' },
          severity: 'critical',
          timestamp: '2026-07-20T00:00:00Z',
        },
      ],
      page: 1,
      limit: 8,
      total: 1,
      total_pages: 1,
    } as never);
  });

  it('should render without crashing', () => {
    const { container } = renderPage();
    expect(container.firstChild).toBeTruthy();
  });

  it('renders live stats from the API instead of hardcoded values', async () => {
    renderPage();
    // drifts.total = 3, events.total = 15, graph.total_nodes = 10, critical = 1
    await waitFor(() => {
      expect(screen.getByText('3')).toBeInTheDocument();
      expect(screen.getByText('15')).toBeInTheDocument();
      expect(screen.getByText('10')).toBeInTheDocument();
    });
    // The old placeholder must be gone
    expect(screen.queryByText(/Coming Soon/i)).not.toBeInTheDocument();
  });

  it('renders a recent drift event with who/what', async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('ModifyInstanceAttribute')).toBeInTheDocument();
      expect(screen.getByText('alice')).toBeInTheDocument();
    });
  });
});
