import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

vi.mock('lucide-react', () => ({
  Activity: () => <div data-testid="icon-Activity" />,
  AlertTriangle: () => <div data-testid="icon-AlertTriangle" />,
  CheckCircle: () => <div data-testid="icon-CheckCircle" />,
  Cloud: () => <div data-testid="icon-Cloud" />,
}));

vi.mock('../api/client', () => ({
  apiClient: {
    getStats: vi.fn().mockResolvedValue({ success: true, data: {
      graph: { total_nodes: 10, total_edges: 5 },
      drifts: { total: 3, severity_counts: { critical: 1, high: 1, medium: 1 } },
      events: { total: 15 },
    }}),
    getGraph: vi.fn().mockResolvedValue({ success: true, data: { nodes: [], edges: [] } }),
  },
}));

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
  it('should render without crashing', () => {
    const { container } = renderPage();
    expect(container.firstChild).toBeTruthy();
  });
});
