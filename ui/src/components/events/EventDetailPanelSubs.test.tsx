/**
 * Tests for EventDetailPanelSubs sub-components
 * Tests EventMetadata, EventChanges, EventUserInfo, and StatusActions components
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { EventMetadata, EventChanges, EventUserInfo, StatusActions } from './EventDetailPanelSubs';
import type { FalcoEvent, RelatedDrift } from '../../api/types';

// Mock JsonDiff component
vi.mock('./JsonDiff', () => ({
  JsonDiff: ({ attribute, oldValue, newValue }: Record<string, unknown>) =>
    `JsonDiff: ${attribute} ${oldValue} -> ${newValue}`,
}));

const mockFalcoEvent = (overrides: Partial<FalcoEvent> = {}): FalcoEvent => ({
  id: 'event-123',
  event_name: 'PutBucketPolicy',
  provider: 'aws',
  timestamp: '2024-01-15T10:30:45Z',
  resource_id: 's3-bucket-prod',
  status: 'open',
  status_reason: undefined,
  changes: undefined,
  related_drifts: [],
  user_identity: {
    UserName: 'alice.smith',
    ARN: 'arn:aws:iam::123456789012:user/alice.smith',
    AccountID: '123456789012',
  },
  region: 'us-east-1',
  ...overrides,
});

describe('EventMetadata', () => {
  it('renders event info section', () => {
    const event = mockFalcoEvent();
    render(<EventMetadata event={event} />);

    expect(screen.getByText('Event Info')).toBeInTheDocument();
  });

  it('displays event name', () => {
    const event = mockFalcoEvent({ event_name: 'PutBucketPolicy' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('PutBucketPolicy')).toBeInTheDocument();
  });

  it('displays provider in uppercase', () => {
    const event = mockFalcoEvent({ provider: 'aws' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('aws')).toBeInTheDocument();
  });

  it('formats and displays timestamp', () => {
    const event = mockFalcoEvent({ timestamp: '2024-01-15T10:30:45Z' });
    render(<EventMetadata event={event} />);

    const timestamp = new Date('2024-01-15T10:30:45Z').toLocaleString();
    expect(screen.getByText(timestamp)).toBeInTheDocument();
  });

  it('displays resource ID', () => {
    const event = mockFalcoEvent({ resource_id: 's3-bucket-prod' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('s3-bucket-prod')).toBeInTheDocument();
  });

  it('handles missing timestamp gracefully', () => {
    const event = mockFalcoEvent({ timestamp: undefined });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('-')).toBeInTheDocument();
  });

  it('handles missing event_name', () => {
    const event = mockFalcoEvent({ event_name: '' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('Event Info')).toBeInTheDocument();
  });

  it('handles GCP provider', () => {
    const event = mockFalcoEvent({ provider: 'gcp' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('gcp')).toBeInTheDocument();
  });

  it('handles Azure provider', () => {
    const event = mockFalcoEvent({ provider: 'azure' });
    render(<EventMetadata event={event} />);

    expect(screen.getByText('azure')).toBeInTheDocument();
  });
});

describe('EventChanges', () => {
  it('renders nothing when no drifts or changes', () => {
    const event = mockFalcoEvent({
      related_drifts: [],
      changes: undefined,
    });
    const { container } = render(<EventChanges event={event} />);

    expect(container.innerHTML).toBe('');
  });

  it('renders related drifts section when available', () => {
    const drift: RelatedDrift = {
      severity: 'high',
      attribute: 'bucket_policy',
      old_value: '{}',
      new_value: '{"Principal": "*"}',
      matched_rules: ['rule-1'],
    };
    const event = mockFalcoEvent({ related_drifts: [drift] });
    render(<EventChanges event={event} />);

    expect(screen.getByText(/Resource Diff/)).toBeInTheDocument();
  });

  it('displays drift count in section title', () => {
    const drifts: RelatedDrift[] = [
      {
        severity: 'high',
        attribute: 'policy',
        old_value: '{}',
        new_value: '{}',
        matched_rules: [],
      },
      {
        severity: 'medium',
        attribute: 'tags',
        old_value: '{}',
        new_value: '{}',
        matched_rules: [],
      },
    ];
    const event = mockFalcoEvent({ related_drifts: drifts });
    render(<EventChanges event={event} />);

    expect(screen.getByText(/Resource Diff \(2\)/)).toBeInTheDocument();
  });

  it('displays severity badge for each drift', () => {
    const drifts: RelatedDrift[] = [
      {
        severity: 'critical',
        attribute: 'attr1',
        old_value: 'old',
        new_value: 'new',
        matched_rules: [],
      },
      {
        severity: 'low',
        attribute: 'attr2',
        old_value: 'old',
        new_value: 'new',
        matched_rules: [],
      },
    ];
    const event = mockFalcoEvent({ related_drifts: drifts });
    render(<EventChanges event={event} />);

    expect(screen.getByText(/critical/i)).toBeInTheDocument();
    expect(screen.getByText(/low/i)).toBeInTheDocument();
  });

  it('displays matched rules when available', () => {
    const drift: RelatedDrift = {
      severity: 'high',
      attribute: 'bucket_policy',
      old_value: '{}',
      new_value: '{}',
      matched_rules: ['rule-1', 'rule-2'],
    };
    const event = mockFalcoEvent({ related_drifts: [drift] });
    render(<EventChanges event={event} />);

    expect(screen.getByText(/rule-1, rule-2/)).toBeInTheDocument();
  });

  it('renders Changes section when raw changes available', () => {
    const event = mockFalcoEvent({
      changes: { key1: 'value1', key2: 'value2' },
      related_drifts: [],
    });
    render(<EventChanges event={event} />);

    expect(screen.getByText('Changes')).toBeInTheDocument();
    expect(screen.getByText(/key1/)).toBeInTheDocument();
  });

  it('renders nothing for empty changes object', () => {
    const event = mockFalcoEvent({ changes: {}, related_drifts: [] });
    const { container } = render(<EventChanges event={event} />);

    expect(container.innerHTML).toBe('');
  });

  it('handles multiple drifts correctly', () => {
    const drifts: RelatedDrift[] = Array.from({ length: 3 }, (_, i) => ({
      severity: i % 2 === 0 ? 'high' : 'medium',
      attribute: `attribute_${i}`,
      old_value: `old_${i}`,
      new_value: `new_${i}`,
      matched_rules: [`rule_${i}`],
    }));
    const event = mockFalcoEvent({ related_drifts: drifts });
    render(<EventChanges event={event} />);

    expect(screen.getByText(/Resource Diff \(3\)/)).toBeInTheDocument();
  });
});

describe('EventUserInfo', () => {
  it('renders User Identity section', () => {
    const event = mockFalcoEvent();
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('User Identity')).toBeInTheDocument();
  });

  it('displays username when available', () => {
    const event = mockFalcoEvent({
      user_identity: { UserName: 'alice.smith' },
    });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('alice.smith')).toBeInTheDocument();
  });

  it('hides username when not available', () => {
    const event = mockFalcoEvent({
      user_identity: { UserName: undefined },
    });
    render(<EventUserInfo event={event} />);

    expect(screen.queryByText('alice.smith')).not.toBeInTheDocument();
  });

  it('displays ARN when available', () => {
    const arn = 'arn:aws:iam::123456789012:user/alice.smith';
    const event = mockFalcoEvent({
      user_identity: { ARN: arn },
    });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText(arn)).toBeInTheDocument();
  });

  it('hides ARN section when not available', () => {
    const event = mockFalcoEvent({
      user_identity: { ARN: undefined },
    });
    render(<EventUserInfo event={event} />);

    expect(screen.queryByText(/arn:aws/i)).not.toBeInTheDocument();
  });

  it('displays AccountID when available', () => {
    const event = mockFalcoEvent({
      user_identity: { AccountID: '123456789012' },
    });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('123456789012')).toBeInTheDocument();
  });

  it('hides AccountID when not available', () => {
    const event = mockFalcoEvent({
      user_identity: { AccountID: undefined },
    });
    render(<EventUserInfo event={event} />);

    const accountElements = screen.queryAllByText('Account');
    expect(accountElements.length).toBe(0);
  });

  it('renders Region section when region available', () => {
    const event = mockFalcoEvent({ region: 'us-east-1' });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('Region')).toBeInTheDocument();
    expect(screen.getByText('us-east-1')).toBeInTheDocument();
  });

  it('hides Region section when region not available', () => {
    const event = mockFalcoEvent({ region: undefined });
    render(<EventUserInfo event={event} />);

    expect(screen.queryByText('Region')).not.toBeInTheDocument();
  });

  it('handles empty user_identity object', () => {
    const event = mockFalcoEvent({ user_identity: {} });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('User Identity')).toBeInTheDocument();
  });

  it('displays all available user info together', () => {
    const event = mockFalcoEvent({
      user_identity: {
        UserName: 'bob.jones',
        ARN: 'arn:aws:iam::999888777666:user/bob.jones',
        AccountID: '999888777666',
      },
      region: 'eu-west-1',
    });
    render(<EventUserInfo event={event} />);

    expect(screen.getByText('bob.jones')).toBeInTheDocument();
    expect(screen.getByText('arn:aws:iam::999888777666:user/bob.jones')).toBeInTheDocument();
    expect(screen.getByText('999888777666')).toBeInTheDocument();
    expect(screen.getByText('eu-west-1')).toBeInTheDocument();
  });
});

describe('StatusActions', () => {
  it('displays current status badge', () => {
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    expect(screen.getByText('Open')).toBeInTheDocument();
  });

  it('displays status reason when available', () => {
    const event = mockFalcoEvent({
      status: 'acknowledged',
      status_reason: 'Investigating',
    });
    render(<StatusActions event={event} />);

    expect(screen.getByText(/Investigating/)).toBeInTheDocument();
  });

  it('shows acknowledge and ignore buttons for open status', () => {
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    expect(screen.getByText('Acknowledge')).toBeInTheDocument();
    expect(screen.getByText('Ignore')).toBeInTheDocument();
  });

  it('calls onStatusChange with acknowledge status', async () => {
    const user = userEvent.setup();
    const onStatusChange = vi.fn();
    const event = mockFalcoEvent({
      status: 'open',
      resource_id: 'resource-123',
    });
    render(<StatusActions event={event} onStatusChange={onStatusChange} />);

    const acknowledgeButton = screen.getByText('Acknowledge');
    await user.click(acknowledgeButton);

    expect(onStatusChange).toHaveBeenCalledWith('resource-123', 'acknowledged');
  });

  it('shows reopen button for acknowledged status', () => {
    const event = mockFalcoEvent({ status: 'acknowledged' });
    render(<StatusActions event={event} />);

    expect(screen.getByText('Reopen')).toBeInTheDocument();
  });

  it('shows reopen button for ignored status', () => {
    const event = mockFalcoEvent({ status: 'ignored' });
    render(<StatusActions event={event} />);

    expect(screen.getByText('Reopen')).toBeInTheDocument();
  });

  it('calls onStatusChange with reopen status', async () => {
    const user = userEvent.setup();
    const onStatusChange = vi.fn();
    const event = mockFalcoEvent({
      status: 'acknowledged',
      resource_id: 'resource-123',
    });
    render(<StatusActions event={event} onStatusChange={onStatusChange} />);

    const reopenButton = screen.getByText('Reopen');
    await user.click(reopenButton);

    expect(onStatusChange).toHaveBeenCalledWith('resource-123', 'open');
  });

  it('toggles ignore form visibility', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    expect(screen.getByText(/Reason for ignoring/)).toBeInTheDocument();
  });

  it('clears ignore form and reason on cancel', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const textarea = screen.getByPlaceholderText(/Known test environment change/);
    await user.type(textarea, 'Test reason');

    const cancelButton = screen.getByText('Cancel');
    await user.click(cancelButton);

    expect(screen.queryByText(/Reason for ignoring/)).not.toBeInTheDocument();
  });

  it('disables confirm ignore button when reason is empty', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const confirmButton = screen.getByText('Confirm Ignore');
    expect(confirmButton).toBeDisabled();
  });

  it('enables confirm ignore button when reason is entered', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const textarea = screen.getByPlaceholderText(/Known test environment change/);
    await user.type(textarea, 'Known issue');

    const confirmButton = screen.getByText('Confirm Ignore');
    expect(confirmButton).not.toBeDisabled();
  });

  it('calls onStatusChange with ignore status and reason', async () => {
    const user = userEvent.setup();
    const onStatusChange = vi.fn();
    const event = mockFalcoEvent({
      status: 'open',
      resource_id: 'resource-123',
    });
    render(<StatusActions event={event} onStatusChange={onStatusChange} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const textarea = screen.getByPlaceholderText(/Known test environment change/);
    await user.type(textarea, 'Test environment update');

    const confirmButton = screen.getByText('Confirm Ignore');
    await user.click(confirmButton);

    expect(onStatusChange).toHaveBeenCalledWith(
      'resource-123',
      'ignored',
      'Test environment update'
    );
  });

  it('clears form after successful ignore', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({
      status: 'open',
      resource_id: 'resource-123',
    });
    render(<StatusActions event={event} onStatusChange={vi.fn()} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const textarea = screen.getByPlaceholderText(/Known test environment change/);
    await user.type(textarea, 'Clear test');

    const confirmButton = screen.getByText('Confirm Ignore');
    await user.click(confirmButton);

    expect(screen.queryByText(/Reason for ignoring/)).not.toBeInTheDocument();
  });

  it('trims whitespace from ignore reason', async () => {
    const user = userEvent.setup();
    const onStatusChange = vi.fn();
    const event = mockFalcoEvent({
      status: 'open',
      resource_id: 'resource-123',
    });
    render(<StatusActions event={event} onStatusChange={onStatusChange} />);

    const ignoreButton = screen.getByText('Ignore');
    await user.click(ignoreButton);

    const textarea = screen.getByPlaceholderText(/Known test environment change/);
    await user.type(textarea, '  Trimmed reason  ');

    const confirmButton = screen.getByText('Confirm Ignore');
    await user.click(confirmButton);

    expect(onStatusChange).toHaveBeenCalledWith(
      'resource-123',
      'ignored',
      'Trimmed reason'
    );
  });

  it('does not call onStatusChange if handler is not provided', async () => {
    const user = userEvent.setup();
    const event = mockFalcoEvent({ status: 'open' });
    render(<StatusActions event={event} />);

    const acknowledgeButton = screen.getByText('Acknowledge');
    await user.click(acknowledgeButton);
  });

  it('handles resolved status', () => {
    const event = mockFalcoEvent({ status: 'resolved' });
    render(<StatusActions event={event} />);

    expect(screen.getByText('Resolved')).toBeInTheDocument();
  });

  it('displays all status variants correctly', () => {
    const statuses: Array<'open' | 'acknowledged' | 'ignored' | 'resolved'> = [
      'open',
      'acknowledged',
      'ignored',
      'resolved',
    ];

    statuses.forEach((status) => {
      const { unmount } = render(
        <StatusActions event={mockFalcoEvent({ status })} />
      );
      expect(screen.getByText(new RegExp(status, 'i'))).toBeInTheDocument();
      unmount();
    });
  });
});
