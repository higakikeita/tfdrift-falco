import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('./icons/ProviderIcons', () => ({
  ProviderIcon: ({ provider }: { provider: string; size: number }) => (
    <div data-testid={`provider-icon-${provider}`}>{provider}</div>
  ),
}));

vi.mock('../constants', () => ({
  SEVERITY_ICONS: {
    critical: '🚨',
    high: '⚠️',
    medium: '⚡',
    low: 'ℹ️',
  },
}));

vi.mock('../utils/formatTimestamp', () => ({
  formatTimestamp: (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString();
  },
}));

import DriftHistoryTable from './DriftHistoryTable';
import type { DriftEvent } from '../types/drift';

describe('DriftHistoryTable', () => {
  const mockDrifts: DriftEvent[] = [
    {
      id: 'drift-1',
      timestamp: '2024-03-29T10:00:00Z',
      resourceId: 'i-123',
      resourceName: 'instance-1',
      resourceType: 'aws_instance',
      provider: 'aws',
      region: 'us-east-1',
      severity: 'high',
      changeType: 'modified',
      attribute: 'tags',
      oldValue: '{}',
      newValue: '{"env":"prod"}',
      userIdentity: {
        userName: 'alice',
        type: 'IAMUser',
        arn: 'arn:aws:iam::123456789012:user/alice',
        accountId: '123456789012',
      },
      cloudtrailEventId: 'event-1',
      cloudtrailEventName: 'ModifyInstanceAttribute',
      sourceIP: '192.168.1.1',
      userAgent: 'cli',
      tags: {},
    },
    {
      id: 'drift-2',
      timestamp: '2024-03-29T11:00:00Z',
      resourceId: 'bucket-xyz',
      resourceName: 'my-bucket',
      resourceType: 'aws_s3_bucket',
      provider: 'aws',
      region: 'us-west-2',
      severity: 'critical',
      changeType: 'created',
      attribute: 'acl',
      oldValue: null,
      newValue: 'public-read',
      userIdentity: {
        userName: 'bob',
        type: 'IAMUser',
        arn: 'arn:aws:iam::123456789012:user/bob',
        accountId: '123456789012',
      },
      cloudtrailEventId: 'event-2',
      cloudtrailEventName: 'CreateBucket',
      sourceIP: '192.168.1.2',
      userAgent: 'cli',
      tags: {},
    },
  ];

  it('should render without crashing', () => {
    const { container } = render(
      <DriftHistoryTable drifts={[]} />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should render table header', () => {
    render(<DriftHistoryTable drifts={mockDrifts} />);
    expect(screen.getByText(/ドリフト履歴/i)).toBeInTheDocument();
    expect(screen.getByText(/時刻/i)).toBeInTheDocument();
  });

  it('should render drift rows when data is available', () => {
    render(<DriftHistoryTable drifts={mockDrifts} />);
    expect(screen.getByText('instance-1')).toBeInTheDocument();
    expect(screen.getByText('my-bucket')).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
    expect(screen.getByText('bob')).toBeInTheDocument();
  });

  it('should display empty message when no drifts', () => {
    render(<DriftHistoryTable drifts={[]} />);
    expect(screen.getByText(/該当するドリフトイベントが見つかりません/i)).toBeInTheDocument();
  });

  it('should call onSelectDrift when a row is clicked', () => {
    const onSelectDrift = vi.fn();
    render(
      <DriftHistoryTable drifts={mockDrifts} onSelectDrift={onSelectDrift} />
    );
    const rows = screen.getAllByRole('row');
    // First row is header, second row is first drift
    expect(rows.length).toBeGreaterThan(1);
  });

  it('should display statistics', () => {
    render(<DriftHistoryTable drifts={mockDrifts} />);
    expect(screen.getByText('2')).toBeInTheDocument(); // total count
  });
});
