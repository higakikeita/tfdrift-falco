/**
 * Tests for DriftDashboard component
 */

import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DriftDashboard } from './DriftDashboard';
import type { DriftSummary } from '../api/types';

// Mock the hooks
vi.mock('../api/hooks/useDiscovery', () => ({
  useDriftSummary: vi.fn(),
  useTriggerDriftDetection: vi.fn(),
}));

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  AlertCircle: ({ className }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'alert-circle-icon', className }, 'Alert'),
  CheckCircle: ({ className }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'check-circle-icon', className }, 'Check'),
  XCircle: ({ className }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'x-circle-icon', className }, 'Error'),
  RefreshCw: ({ className }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'refresh-icon', className }, 'Refresh'),
  AlertTriangle: ({ className }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'alert-triangle-icon', className }, 'Warning'),
}));

import { useDriftSummary, useTriggerDriftDetection } from '../api/hooks/useDiscovery';

describe('DriftDashboard', () => {
  const mockDriftSummary = (overrides: Partial<DriftSummary> = {}): DriftSummary => ({
    region: 'us-east-1',
    timestamp: '2024-01-15T10:00:00Z',
    counts: {
      terraform_resources: 42,
      aws_resources: 50,
      unmanaged: 3,
      missing: 2,
      modified: 1,
    },
    breakdown: {
      unmanaged_by_type: {
        'aws_instance': 2,
        'aws_security_group': 1,
      },
      missing_by_type: {
        'aws_s3_bucket': 1,
        'aws_rds_instance': 1,
      },
      modified_by_type: {
        'aws_instance': 1,
      },
    },
    ...overrides,
  });

  const mockUseDriftSummary = vi.fn();
  const mockTriggerDetection = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    mockUseDriftSummary.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    mockTriggerDetection.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    });

    (useDriftSummary as unknown as typeof vi.fn).mockImplementation(mockUseDriftSummary);
    (useTriggerDriftDetection as unknown as typeof vi.fn).mockImplementation(mockTriggerDetection);
  });

  describe('Loading State', () => {
    it('displays loading spinner when data is loading', () => {
      mockUseDriftSummary.mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('Scanning AWS resources...')).toBeInTheDocument();
    });

    it('shows loading spinner with animation', () => {
      mockUseDriftSummary.mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      });

      const { container } = render(<DriftDashboard />);

      const spinner = container.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('displays error message when error occurs', () => {
      const errorMessage = 'Failed to fetch data';
      mockUseDriftSummary.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error(errorMessage),
      });

      render(<DriftDashboard />);

      expect(screen.getByText(/Failed to load drift information/)).toBeInTheDocument();
    });

    it('shows error icon', () => {
      mockUseDriftSummary.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Test error'),
      });

      render(<DriftDashboard />);

      expect(screen.getByTestId('x-circle-icon')).toBeInTheDocument();
    });

    it('displays error in red styling', () => {
      mockUseDriftSummary.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Test error'),
      });

      const { container } = render(<DriftDashboard />);

      expect(container.querySelector('.bg-red-50')).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('displays empty state message when summary is null', () => {
      mockUseDriftSummary.mockReturnValue({
        data: null,
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('No drift data available')).toBeInTheDocument();
    });
  });

  describe('Successful Data Display', () => {
    beforeEach(() => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });
    });

    it('renders header with region information', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('AWS Drift Detection')).toBeInTheDocument();
      expect(screen.getByText(/Region: us-east-1/)).toBeInTheDocument();
    });

    it('displays last updated timestamp', () => {
      render(<DriftDashboard />);

      expect(screen.getByText(/Last updated:/)).toBeInTheDocument();
    });

    it('renders auto-refresh checkbox', () => {
      render(<DriftDashboard />);

      const checkbox = screen.getByRole('checkbox', { name: /Auto-refresh/ });
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).toBeChecked();
    });

    it('renders scan now button', () => {
      render(<DriftDashboard />);

      expect(screen.getByRole('button', { name: /Scan Now/ })).toBeInTheDocument();
    });
  });

  describe('Drift Status Display', () => {
    it('shows "No Drift Detected" when resources match Terraform state', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('No Drift Detected')).toBeInTheDocument();
      expect(screen.getByText('All AWS resources match Terraform state')).toBeInTheDocument();
    });

    it('shows "Drift Detected" when there are unmanaged resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 50,
            unmanaged: 3,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('Drift Detected')).toBeInTheDocument();
    });

    it('shows "Drift Detected" when there are missing resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 40,
            unmanaged: 0,
            missing: 2,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('Drift Detected')).toBeInTheDocument();
    });

    it('shows "Drift Detected" when there are modified resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 0,
            modified: 1,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('Drift Detected')).toBeInTheDocument();
    });

    it('uses green styling for "No Drift Detected"', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      const { container } = render(<DriftDashboard />);

      expect(container.querySelector('.bg-green-50')).toBeInTheDocument();
    });

    it('uses yellow styling for "Drift Detected"', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });

      const { container } = render(<DriftDashboard />);

      expect(container.querySelector('.bg-yellow-50')).toBeInTheDocument();
    });
  });

  describe('Resource Count Cards', () => {
    beforeEach(() => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });
    });

    it('displays Terraform resource count', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('Terraform Resources')).toBeInTheDocument();
      expect(screen.getByText('42')).toBeInTheDocument();
    });

    it('displays unmanaged resource count', () => {
      render(<DriftDashboard />);

      const unmanagedCards = screen.getAllByText(/Unmanaged/);
      expect(unmanagedCards.length).toBeGreaterThan(0);
      expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('displays missing resource count', () => {
      render(<DriftDashboard />);

      const missingCards = screen.getAllByText(/Missing/);
      expect(missingCards.length).toBeGreaterThan(0);
      const allTexts = screen.getAllByText('2');
      expect(allTexts.length).toBeGreaterThan(0);
    });

    it('displays modified resource count', () => {
      render(<DriftDashboard />);

      const modifiedCards = screen.getAllByText(/Modified/);
      expect(modifiedCards.length).toBeGreaterThan(0);
      const allTexts = screen.getAllByText('1');
      expect(allTexts.length).toBeGreaterThan(0);
    });

    it('highlights unmanaged card when there are unmanaged resources', () => {
      const { container } = render(<DriftDashboard />);

      const unmanagedCard = container.querySelector('.bg-orange-50');
      expect(unmanagedCard).toBeInTheDocument();
    });

    it('highlights missing card when there are missing resources', () => {
      const { container } = render(<DriftDashboard />);

      const missingCard = container.querySelector('.bg-red-50');
      expect(missingCard).toBeInTheDocument();
    });

    it('highlights modified card when there are modified resources', () => {
      const { container } = render(<DriftDashboard />);

      const modifiedCards = container.querySelectorAll('.bg-yellow-50');
      expect(modifiedCards.length).toBeGreaterThan(0);
    });
  });

  describe('Resource Type Breakdown', () => {
    beforeEach(() => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });
    });

    it('shows breakdown section when there is drift', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('Resource Type Breakdown')).toBeInTheDocument();
    });

    it('displays unmanaged resources by type', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('Unmanaged Resources')).toBeInTheDocument();
      const awsInstances = screen.getAllByText('aws_instance');
      expect(awsInstances.length).toBeGreaterThan(0);
      expect(screen.getByText('aws_security_group')).toBeInTheDocument();
    });

    it('displays missing resources by type', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('Missing Resources')).toBeInTheDocument();
      expect(screen.getByText('aws_s3_bucket')).toBeInTheDocument();
      expect(screen.getByText('aws_rds_instance')).toBeInTheDocument();
    });

    it('displays modified resources by type', () => {
      render(<DriftDashboard />);

      expect(screen.getByText('Modified Resources')).toBeInTheDocument();
    });

    it('hides breakdown section when no drift', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.queryByText('Resource Type Breakdown')).not.toBeInTheDocument();
    });

    it('hides unmanaged section when no unmanaged resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 2,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.queryByText('Unmanaged Resources')).not.toBeInTheDocument();
    });

    it('hides missing section when no missing resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 50,
            unmanaged: 3,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.queryByText('Missing Resources')).not.toBeInTheDocument();
    });

    it('hides modified section when no modified resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 50,
            unmanaged: 3,
            missing: 2,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.queryByText('Modified Resources')).not.toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    beforeEach(() => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });
    });

    it('toggles auto-refresh checkbox', async () => {
      const user = userEvent.setup();
      render(<DriftDashboard />);

      const checkbox = screen.getByRole('checkbox', { name: /Auto-refresh/ });
      expect(checkbox).toBeChecked();

      await user.click(checkbox);
      expect(checkbox).not.toBeChecked();

      await user.click(checkbox);
      expect(checkbox).toBeChecked();
    });

    it('calls triggerDetection.mutate when Scan Now is clicked', async () => {
      const user = userEvent.setup();
      const mutateFn = vi.fn();
      mockTriggerDetection.mockReturnValue({
        mutate: mutateFn,
        isPending: false,
      });

      render(<DriftDashboard />);

      const scanButton = screen.getByRole('button', { name: /Scan Now/ });
      await user.click(scanButton);

      expect(mutateFn).toHaveBeenCalled();
    });

    it('disables Scan Now button when mutation is pending', () => {
      mockTriggerDetection.mockReturnValue({
        mutate: vi.fn(),
        isPending: true,
      });

      render(<DriftDashboard />);

      const scanButton = screen.getByRole('button', { name: /Scan Now/ });
      expect(scanButton).toBeDisabled();
    });

    it('shows spinner on Scan Now button when pending', () => {
      mockTriggerDetection.mockReturnValue({
        mutate: vi.fn(),
        isPending: true,
      });

      const { container } = render(<DriftDashboard />);

      const spinner = container.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('Region Support', () => {
    it('uses default region when not provided', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary(),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(mockUseDriftSummary).toHaveBeenCalledWith('us-east-1', expect.any(Object));
    });

    it('uses provided region', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({ region: 'eu-west-1' }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard region="eu-west-1" />);

      expect(mockUseDriftSummary).toHaveBeenCalledWith('eu-west-1', expect.any(Object));
    });

    it('displays correct region from props', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({ region: 'ap-southeast-1' }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard region="ap-southeast-1" />);

      expect(screen.getByText(/Region: ap-southeast-1/)).toBeInTheDocument();
    });
  });

  describe('Drift Summary Calculation', () => {
    it('correctly counts total drift resources', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 50,
            unmanaged: 3,
            missing: 2,
            modified: 1,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      // 3 + 2 + 1 = 6 total drifts
      expect(screen.getByText(/6 resource\(s\) differ/)).toBeInTheDocument();
    });

    it('shows zero total when no drift', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 42,
            aws_resources: 42,
            unmanaged: 0,
            missing: 0,
            modified: 0,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.queryByText(/differ/)).not.toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('handles empty resource type breakdowns', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          breakdown: {
            unmanaged_by_type: {},
            missing_by_type: {},
            modified_by_type: {},
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('Resource Type Breakdown')).toBeInTheDocument();
    });

    it('handles large counts gracefully', () => {
      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({
          counts: {
            terraform_resources: 9999,
            aws_resources: 10000,
            unmanaged: 500,
            missing: 300,
            modified: 200,
          },
        }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText('9999')).toBeInTheDocument();
      expect(screen.getByText('500')).toBeInTheDocument();
    });

    it('handles future dates in timestamps', () => {
      const futureDate = new Date();
      futureDate.setDate(futureDate.getDate() + 1);

      mockUseDriftSummary.mockReturnValue({
        data: mockDriftSummary({ timestamp: futureDate.toISOString() }),
        isLoading: false,
        error: null,
      });

      render(<DriftDashboard />);

      expect(screen.getByText(/Last updated:/)).toBeInTheDocument();
    });
  });
});
