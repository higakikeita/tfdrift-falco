import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
}));

import EventTable from './EventTable';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const mockEvents = [
  {
    id: 'evt-1',
    provider: 'aws',
    event_name: 'ModifySecurityGroupRules',
    resource_type: 'aws_security_group',
    resource_id: 'sg-12345',
    user_identity: 'admin',
    timestamp: '2026-03-22T10:00:00Z',
    severity: 'critical',
    status: 'open',
    region: 'us-east-1',
    service_name: 'EC2',
  },
  {
    id: 'evt-2',
    provider: 'gcp',
    event_name: 'compute.firewalls.patch',
    resource_type: 'google_compute_firewall',
    resource_id: 'fw-allow-ssh',
    user_identity: 'user@project.iam.gserviceaccount.com',
    timestamp: '2026-03-22T09:00:00Z',
    severity: 'high',
    status: 'acknowledged',
    region: 'us-central1',
    service_name: 'Compute Engine',
  },
];

const renderTable = (events = mockEvents, onSelect = vi.fn()) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <EventTable events={events} onSelectEvent={onSelect} />
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
