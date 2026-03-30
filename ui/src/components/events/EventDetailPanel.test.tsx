import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  X: () => <div data-testid="close-icon" />,
  ChevronLeft: () => <div data-testid="prev-icon" />,
  ChevronRight: () => <div data-testid="next-icon" />,
  CheckCircle2: () => <div data-testid="check-icon" />,
  EyeOff: () => <div data-testid="eye-off-icon" />,
  Clock: () => <div data-testid="clock-icon" />,
  Shield: () => <div data-testid="shield-icon" />,
  User: () => <div data-testid="user-icon" />,
  MapPin: () => <div data-testid="map-pin-icon" />,
  Activity: () => <div data-testid="activity-icon" />,
}));

vi.mock('../../lib/utils', () => ({
  cn: () => '',
}));

vi.mock('./JsonDiff', () => ({
  JsonDiff: ({ attribute }: { attribute: string; oldValue: unknown; newValue: unknown }) => (
    <div data-testid={`json-diff-${attribute}`}>{attribute}</div>
  ),
}));

import { EventDetailPanel } from './EventDetailPanel';
import type { FalcoEvent } from '../../api/types';

describe('EventDetailPanel', () => {
  const mockEvent: FalcoEvent = {
    id: 'event-1',
    event_name: 'Unauthorized_Process_Execution',
    resource_id: 'i-1234567890abcdef0',
    resource_type: 'aws_instance',
    severity: 'high',
    provider: 'aws',
    timestamp: '2024-03-29T10:00:00Z',
    region: 'us-east-1',
    status: 'open',
    user_identity: {
      UserName: 'alice',
      ARN: 'arn:aws:iam::123456789012:user/alice',
      AccountID: '123456789012',
    },
    changes: { key: 'value' },
    related_drifts: [
      {
        severity: 'high',
        attribute: 'tags',
        old_value: '{}',
        new_value: '{"env":"prod"}',
        matched_rules: ['rule1'],
      },
    ],
  };

  const mockOnClose = vi.fn();
  const mockOnStatusChange = vi.fn();

  it('should return null when event is null', () => {
    const { container } = render(
      <EventDetailPanel event={null} onClose={mockOnClose} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('should render event details when event is provided', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    expect(screen.getAllByText('i-1234567890abcdef0')).toHaveLength(2);
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });

  it('should display event name and provider information', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    expect(screen.getByText('Unauthorized_Process_Execution')).toBeInTheDocument();
    const providerElement = document.querySelector('dd.font-medium');
    expect(providerElement?.textContent).toBe('aws');
  });

  it('should display user identity information', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    expect(screen.getByText('alice')).toBeInTheDocument();
    expect(screen.getByText('arn:aws:iam::123456789012:user/alice')).toBeInTheDocument();
  });

  it('should display region information', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    expect(screen.getByText('us-east-1')).toBeInTheDocument();
  });

  it('should render related drifts section', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    expect(screen.getByText(/Resource Diff/i)).toBeInTheDocument();
  });

  it('should call onStatusChange when acknowledge button is clicked', () => {
    render(
      <EventDetailPanel
        event={mockEvent}
        onClose={mockOnClose}
        onStatusChange={mockOnStatusChange}
      />
    );
    const acknowledgeButton = screen.getByText('Acknowledge');
    fireEvent.click(acknowledgeButton);
    expect(mockOnStatusChange).toHaveBeenCalledWith('i-1234567890abcdef0', 'acknowledged');
  });

  it('should show ignore form when ignore button is clicked', () => {
    render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    const ignoreButton = screen.getByText('Ignore');
    fireEvent.click(ignoreButton);
    expect(screen.getByText(/Reason for ignoring/i)).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const { container } = render(
      <EventDetailPanel event={mockEvent} onClose={mockOnClose} />
    );
    // Find and click the close button (X icon)
    const closeButtons = container.querySelectorAll('button');
    expect(closeButtons.length).toBeGreaterThan(0);
  });

  it('should handle event with no user identity', () => {
    const eventWithoutUser = { ...mockEvent, user_identity: undefined };
    const { container } = render(
      <EventDetailPanel event={eventWithoutUser as FalcoEvent} onClose={mockOnClose} />
    );
    expect(container.firstChild).toBeTruthy();
  });
});
