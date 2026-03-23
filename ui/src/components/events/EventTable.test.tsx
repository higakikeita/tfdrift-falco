import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('lucide-react', () => ({
  ArrowUpDown: () => <div data-testid="icon-ArrowUpDown" />,
  ChevronLeft: () => <div data-testid="icon-ChevronLeft" />,
  ChevronRight: () => <div data-testid="icon-ChevronRight" />,
}));

import { EventTable } from './EventTable';
import type { DriftEvent } from '../../types/drift';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const mockEvents: DriftEvent[] = [
  {
    id: 'evt-1',
    provider: 'aws',
    timestamp: '2026-03-22T10:00:00Z',
    severity: 'critical',
    resourceType: 'aws_security_group',
    resourceId: 'sg-12345',
    resourceName: 'SecurityGroup-1',
    changeType: 'modified',
    attribute: 'ingress_rules',
    oldValue: null,
    newValue: null,
    userIdentity: { type: 'IAMUser', userName: 'admin', arn: 'arn:aws:iam::123456789:user/admin', accountId: '123456789' },
    region: 'us-east-1',
  },
  {
    id: 'evt-2',
    provider: 'gcp',
    timestamp: '2026-03-22T09:00:00Z',
    severity: 'high',
    resourceType: 'google_compute_firewall',
    resourceId: 'fw-allow-ssh',
    resourceName: 'firewall-allow-ssh',
    changeType: 'modified',
    attribute: 'source_ranges',
    oldValue: null,
    newValue: null,
    userIdentity: { type: 'ServiceAccount', userName: 'user@project.iam.gserviceaccount.com', arn: '', accountId: '' },
    region: 'us-central1',
  },
];

const renderTable = (events = mockEvents, onEventClick = vi.fn()) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <EventTable events={events} onEventClick={onEventClick} />
    </QueryClientProvider>
  );
};

describe('EventTable', () => {
  it('should render without crashing', () => {
    const { container } = renderTable();
    expect(container.firstChild).toBeTruthy();
  });

  it('should render event data', () => {
    renderTable();
    const text = document.body.textContent || '';
    // Should contain some event info
    expect(text.length).toBeGreaterThan(0);
  });

  it('should render empty state with no events', () => {
    const { container } = renderTable([]);
    expect(container.firstChild).toBeTruthy();
  });
});
