import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('lucide-react', () => ({
  AlertCircle: () => <div data-testid="alert-circle-icon" />,
  CheckCircle: () => <div data-testid="check-circle-icon" />,
  XCircle: () => <div data-testid="x-circle-icon" />,
  RefreshCw: () => <div data-testid="refresh-icon" />,
  AlertTriangle: () => <div data-testid="alert-triangle-icon" />,
}));

vi.mock('../api/hooks/useDiscovery', () => ({
  useDriftSummary: (region: string, opts: { enabled: boolean }) => {
    if (!opts.enabled) {
      return { data: null, isLoading: false, error: null };
    }
    return {
      data: {
        region,
        timestamp: new Date().toISOString(),
        counts: {
          terraform_resources: 42,
          unmanaged: 5,
          missing: 2,
          modified: 3,
        },
        breakdown: {
          unmanaged_by_type: { 'aws_instance': 3, 'aws_s3_bucket': 2 },
          missing_by_type: { 'aws_security_group': 2 },
          modified_by_type: { 'aws_iam_role': 2, 'aws_subnet': 1 },
        },
      },
      isLoading: false,
      error: null,
    };
  },
  useTriggerDriftDetection: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

import { DriftDashboard } from './DriftDashboard';

const queryClient = new QueryClient();

describe('DriftDashboard', () => {
  const renderComponent = (region?: string) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <DriftDashboard region={region} />
      </QueryClientProvider>
    );
  };

  it('should render without crashing', () => {
    const { container } = renderComponent();
    expect(container.firstChild).toBeTruthy();
  });

  it('should display the dashboard with drift data', () => {
    renderComponent('us-east-1');
    expect(screen.getByText(/AWS Drift Detection/i)).toBeInTheDocument();
    expect(screen.getByText(/us-east-1/i)).toBeInTheDocument();
  });

  it('should display summary cards with counts', () => {
    renderComponent();
    expect(screen.getByText('42')).toBeInTheDocument(); // terraform_resources
    expect(screen.getByText('5')).toBeInTheDocument(); // unmanaged
    // "2" and "3" appear multiple times in cards and breakdowns
    // Just verify they exist at all (in the summary cards section)
    const allTwos = screen.getAllByText('2');
    expect(allTwos.length).toBeGreaterThanOrEqual(1); // At least the Missing card
    const allThrees = screen.getAllByText('3');
    expect(allThrees.length).toBeGreaterThanOrEqual(1); // At least the Modified card
  });

  it('should display resource type breakdown when drift is detected', () => {
    renderComponent();
    expect(screen.getByText(/Resource Type Breakdown/i)).toBeInTheDocument();
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });
});
