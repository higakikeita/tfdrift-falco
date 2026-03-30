/**
 * Tests for DriftHistoryTable component
 */

import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import DriftHistoryTable from './DriftHistoryTable';
import type { DriftEvent } from '../types/drift';

// Mock ProviderIcon component
vi.mock('./icons/ProviderIcons', () => ({
  ProviderIcon: ({ provider }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': `provider-icon-${provider}` }, `Icon: ${provider}`),
}));

// Mock formatTimestamp utility
vi.mock('../utils/formatTimestamp', () => ({
  formatTimestamp: (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toISOString().split('T')[0];
  },
}));

describe('DriftHistoryTable', () => {
  const mockDriftEvent = (overrides: Partial<DriftEvent> = {}): DriftEvent => ({
    id: 'drift-1',
    timestamp: '2024-01-15T10:00:00Z',
    severity: 'high',
    provider: 'aws',
    resourceType: 'aws_instance',
    resourceId: 'i-1234567890abcdef0',
    resourceName: 'production-server',
    changeType: 'modified',
    attribute: 'instance_type',
    oldValue: 't2.micro',
    newValue: 't2.small',
    userIdentity: {
      userName: 'john.doe',
      type: 'IAMUser',
      arn: 'arn:aws:iam::123456789012:user/john.doe',
      accountId: '123456789012',
    },
    region: 'us-east-1',
    cloudtrailEventId: 'evt-123',
    cloudtrailEventName: 'ModifyInstanceAttribute',
    sourceIP: '203.0.113.42',
    userAgent: 'aws-cli/2.0.0',
    tags: { Environment: 'production' },
    ...overrides,
  });

  const mockDrifts: DriftEvent[] = [
    mockDriftEvent({
      id: 'drift-1',
      severity: 'critical',
      timestamp: '2024-01-15T10:00:00Z',
      resourceName: 'prod-critical',
    }),
    mockDriftEvent({
      id: 'drift-2',
      severity: 'high',
      timestamp: '2024-01-15T09:00:00Z',
      resourceName: 'prod-high',
    }),
    mockDriftEvent({
      id: 'drift-3',
      severity: 'medium',
      timestamp: '2024-01-15T08:00:00Z',
      resourceName: 'prod-medium',
    }),
    mockDriftEvent({
      id: 'drift-4',
      severity: 'low',
      timestamp: '2024-01-15T07:00:00Z',
      provider: 'gcp',
      resourceType: 'gcp_instance',
      resourceName: 'gcp-low',
    }),
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic Rendering', () => {
    it('renders table with header and footer', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.getByText('ドリフト履歴')).toBeInTheDocument();
      expect(screen.getByText('実環境での変更履歴を時系列で表示')).toBeInTheDocument();
    });

    it('displays all drift events in table rows', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.getByText('prod-critical')).toBeInTheDocument();
      expect(screen.getAllByRole('row')).toHaveLength(5); // 1 header + 4 data rows
    });

    it('renders table columns correctly', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const headers = screen.getAllByRole('columnheader');
      expect(headers.some(h => h.textContent?.includes('時刻'))).toBe(true);
      expect(headers.some(h => h.textContent?.includes('重大度'))).toBe(true);
      expect(headers.some(h => h.textContent?.includes('プロバイダー'))).toBe(true);
      expect(headers.some(h => h.textContent?.includes('リソース'))).toBe(true);
    });

    it('displays total count in header', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.getByText('4')).toBeInTheDocument();
    });
  });

  describe('Empty State', () => {
    it('renders empty state message when no drifts', () => {
      render(<DriftHistoryTable drifts={[]} />);

      expect(screen.getByText('該当するドリフトイベントが見つかりません')).toBeInTheDocument();
    });

    it('shows total count as 0 for empty drifts', () => {
      render(<DriftHistoryTable drifts={[]} />);

      expect(screen.getByText(/表示中: 0 \/ 0 イベント/)).toBeInTheDocument();
    });

    it('hides clear filters button when no filters active', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.queryByText('フィルターをクリア')).not.toBeInTheDocument();
    });
  });

  describe('Severity Filtering', () => {
    it('filters drifts by critical severity', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      await user.click(criticalButton);

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 critical
      expect(screen.getByText('prod-critical')).toBeInTheDocument();
    });

    it('filters drifts by high severity', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const highButton = screen.getByRole('button', { name: /High: 1/ });
      await user.click(highButton);

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 high
      expect(screen.getByText('prod-high')).toBeInTheDocument();
    });

    it('combines multiple severity filters', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      const highButton = screen.getByRole('button', { name: /High: 1/ });

      await user.click(criticalButton);
      await user.click(highButton);

      expect(screen.getAllByRole('row')).toHaveLength(3); // 1 header + 2 drifts
    });

    it('toggles severity filter on and off', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });

      await user.click(criticalButton);
      expect(screen.getAllByRole('row')).toHaveLength(2); // filtered

      await user.click(criticalButton);
      expect(screen.getAllByRole('row')).toHaveLength(5); // unfiltered
    });
  });

  describe('Provider Filtering', () => {
    it('filters drifts by AWS provider', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const awsButton = screen.getByRole('button', { name: 'AWS' });
      await user.click(awsButton);

      expect(screen.getAllByRole('row')).toHaveLength(4); // 1 header + 3 AWS drifts
    });

    it('filters drifts by GCP provider', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const gcpButton = screen.getByRole('button', { name: 'GCP' });
      await user.click(gcpButton);

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 GCP drift
      expect(screen.getByText('gcp-low')).toBeInTheDocument();
    });

    it('combines severity and provider filters', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      const awsButton = screen.getByRole('button', { name: 'AWS' });

      await user.click(criticalButton);
      await user.click(awsButton);

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 drift
      expect(screen.getByText('prod-critical')).toBeInTheDocument();
    });
  });

  describe('Search Filtering', () => {
    it('filters by resource name', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'prod-critical');

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 match
    });

    it('filters by resource type', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'gcp_instance');

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 gcp_instance
    });

    it('filters by user name', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'john.doe');

      expect(screen.getAllByRole('row')).toHaveLength(5); // 1 header + 4 with same user
    });

    it('filters by attribute', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'instance_type');

      expect(screen.getAllByRole('row')).toHaveLength(5); // 1 header + 4 with same attribute
    });

    it('returns empty results for non-matching search', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'nonexistent-resource');

      expect(screen.getByText('該当するドリフトイベントが見つかりません')).toBeInTheDocument();
    });

    it('is case-insensitive', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);
      await user.type(searchInput, 'PROD-CRITICAL');

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 match
    });
  });

  describe('Sorting', () => {
    it('sorts by timestamp descending by default', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const rows = screen.getAllByRole('row');
      // First data row should be from 10:00:00 (most recent)
      expect(within(rows[1]).getByText('prod-critical')).toBeInTheDocument();
    });

    it('sorts by timestamp ascending', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const sortSelect = screen.getByDisplayValue('最新順');
      await user.selectOptions(sortSelect, 'timestamp-asc');

      expect(screen.getAllByRole('row')).toHaveLength(5);
    });

    it('sorts by severity descending', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const sortSelect = screen.getByDisplayValue('最新順');
      await user.selectOptions(sortSelect, 'severity-desc');

      expect(screen.getAllByRole('row')).toHaveLength(5);
    });

    it('sorts by severity ascending', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const sortSelect = screen.getByDisplayValue('最新順');
      await user.selectOptions(sortSelect, 'severity-asc');

      expect(screen.getAllByRole('row')).toHaveLength(5);
    });
  });

  describe('Row Selection', () => {
    it('calls onSelectDrift when row is clicked', async () => {
      const user = userEvent.setup();
      const onSelectDrift = vi.fn();
      render(<DriftHistoryTable drifts={mockDrifts} onSelectDrift={onSelectDrift} />);

      const rows = screen.getAllByRole('row');
      await user.click(rows[1]); // Click first data row

      expect(onSelectDrift).toHaveBeenCalledWith(expect.objectContaining({ id: 'drift-1' }));
    });

    it('highlights selected row', async () => {
      const user = userEvent.setup();
      const { container } = render(<DriftHistoryTable drifts={mockDrifts} />);

      const rows = screen.getAllByRole('row');
      await user.click(rows[1]);

      // Check that row has selected styling applied
      expect(container.querySelector('.bg-blue-50')).toBeInTheDocument();
    });

    it('handles multiple selections', async () => {
      const user = userEvent.setup();
      const onSelectDrift = vi.fn();
      render(<DriftHistoryTable drifts={mockDrifts} onSelectDrift={onSelectDrift} />);

      const rows = screen.getAllByRole('row');
      await user.click(rows[1]);
      await user.click(rows[2]);

      expect(onSelectDrift).toHaveBeenCalledTimes(2);
    });
  });

  describe('Statistics Bar', () => {
    it('displays correct severity counts', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.getByText(/Critical: 1/)).toBeInTheDocument();
      expect(screen.getByText(/High: 1/)).toBeInTheDocument();
      expect(screen.getByText(/Medium: 1/)).toBeInTheDocument();
      expect(screen.getByText(/Low: 1/)).toBeInTheDocument();
    });

    it('updates stats after filtering', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      await user.click(criticalButton);

      // Counts should not change, only displayed results
      expect(screen.getByText(/Critical: 1/)).toBeInTheDocument();
    });
  });

  describe('Clear Filters Button', () => {
    it('appears when filters are active', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      await user.click(criticalButton);

      expect(screen.getByText('フィルターをクリア')).toBeInTheDocument();
    });

    it('clears all filters when clicked', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      // Apply multiple filters
      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      const searchInput = screen.getByPlaceholderText(/リソース名、ユーザー、属性で検索/);

      await user.click(criticalButton);
      await user.type(searchInput, 'test');

      // Clear filters
      const clearButton = screen.getByText('フィルターをクリア');
      await user.click(clearButton);

      // All drifts should be visible again
      expect(screen.getAllByRole('row')).toHaveLength(5); // 1 header + 4 drifts
      expect((searchInput as HTMLInputElement).value).toBe('');
    });

    it('disappears after clearing filters', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      await user.click(criticalButton);

      const clearButton = screen.getByText('フィルターをクリア');
      await user.click(clearButton);

      expect(screen.queryByText('フィルターをクリア')).not.toBeInTheDocument();
    });
  });

  describe('Footer Information', () => {
    it('displays filtered vs total count', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(screen.getByText(/表示中: 4 \/ 4 イベント/)).toBeInTheDocument();
    });

    it('updates count after filtering', async () => {
      const user = userEvent.setup();
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const criticalButton = screen.getByRole('button', { name: /Critical: 1/ });
      await user.click(criticalButton);

      expect(screen.getByText(/表示中: 1 \/ 4 イベント/)).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('renders with proper table structure', () => {
      const { container } = render(<DriftHistoryTable drifts={mockDrifts} />);

      expect(container.querySelector('table')).toBeInTheDocument();
      expect(container.querySelector('thead')).toBeInTheDocument();
      expect(container.querySelector('tbody')).toBeInTheDocument();
    });

    it('has proper heading hierarchy', () => {
      render(<DriftHistoryTable drifts={mockDrifts} />);

      const heading = screen.getByRole('heading', { level: 2 });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('handles single drift event', () => {
      render(<DriftHistoryTable drifts={[mockDriftEvent()]} />);

      expect(screen.getAllByRole('row')).toHaveLength(2); // 1 header + 1 data row
      expect(screen.getByText('1')).toBeInTheDocument();
    });

    it('handles drifts with null resourceName', () => {
      const driftsWithoutName = [
        mockDriftEvent({ id: 'drift-x', resourceName: undefined }),
      ];
      render(<DriftHistoryTable drifts={driftsWithoutName} />);

      expect(screen.getByText('i-1234567890abcdef0')).toBeInTheDocument();
    });

    it('handles drifts with optional fields missing', () => {
      const minimalDrift = mockDriftEvent({
        id: 'drift-y',
        tags: undefined,
        cloudtrailEventId: undefined,
      });
      render(<DriftHistoryTable drifts={[minimalDrift]} />);

      expect(screen.getByText('production-server')).toBeInTheDocument();
    });
  });
});
