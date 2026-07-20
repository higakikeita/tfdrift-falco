/**
 * AnalyticsPage Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AnalyticsPage } from './AnalyticsPage';
import { useStats } from '../api/hooks/useStats';

vi.mock('../components/analytics/SeverityChart', () => ({
  SeverityChart: ({ data }: { data: unknown[] }) => (
    <div data-testid="severity-chart">{data.length} severities</div>
  ),
}));

vi.mock('../components/analytics/ServiceBreakdown', () => ({
  ServiceBreakdown: ({ data }: { data: unknown[] }) => (
    <div data-testid="service-breakdown">{data.length} services</div>
  ),
}));

vi.mock('../api/hooks/useStats', () => ({
  useStats: vi.fn(),
}));

const mockUseStats = vi.mocked(useStats);

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <AnalyticsPage />
    </QueryClientProvider>
  );
}

describe('AnalyticsPage', () => {
  beforeEach(() => {
    mockUseStats.mockReturnValue({
      data: {
        graph: { total_nodes: 42, total_edges: 10 },
        drifts: { total: 7, severity_counts: {}, resource_types: {} },
        events: { total: 30 },
        unmanaged: { total: 0 },
        severity_breakdown: { critical: 2, high: 3, medium: 1, low: 1 },
        top_resource_types: [
          { resource_type: 'aws_instance', count: 4 },
          { resource_type: 'aws_s3_bucket', count: 2 },
        ],
      },
      isLoading: false,
    } as never);
  });

  it('renders live KPI values from /stats (no hardcoded mocks)', async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('7')).toBeInTheDocument(); // total drifts
      expect(screen.getByText('30')).toBeInTheDocument(); // cloudtrail events
      expect(screen.getByText('42')).toBeInTheDocument(); // tracked resources
    });
  });

  it('feeds severity and service charts from real stats', async () => {
    renderPage();
    await waitFor(() => {
      expect(screen.getByTestId('severity-chart')).toHaveTextContent('4 severities');
      expect(screen.getByTestId('service-breakdown')).toHaveTextContent('2 services');
    });
  });

  it('shows empty panels when there is no data', async () => {
    mockUseStats.mockReturnValue({
      data: {
        graph: { total_nodes: 0, total_edges: 0 },
        drifts: { total: 0, severity_counts: {}, resource_types: {} },
        events: { total: 0 },
        unmanaged: { total: 0 },
        severity_breakdown: { critical: 0, high: 0, medium: 0, low: 0 },
        top_resource_types: [],
      },
      isLoading: false,
    } as never);
    renderPage();
    await waitFor(() => {
      expect(screen.getByText('Severity Breakdown')).toBeInTheDocument();
      expect(screen.getByText('Top Resource Types')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('severity-chart')).toBeNull();
  });
});
