import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('react-icons/si', () => ({
  SiGooglecloud: () => <div data-testid="gcp-icon" />,
}));

vi.mock('react-icons/fa', () => ({
  FaAws: () => <div data-testid="aws-icon" />,
}));

import DriftDetailPanel from './DriftDetailPanel';
import type { DriftEvent } from '../types/drift';

describe('DriftDetailPanel', () => {
  const mockDrift: DriftEvent = {
    id: 'drift-1',
    timestamp: '2024-03-29T10:00:00Z',
    resourceId: 'i-1234567890abcdef0',
    resourceName: 'web-server',
    resourceType: 'aws_instance',
    provider: 'aws',
    region: 'us-east-1',
    severity: 'high',
    changeType: 'modified',
    attribute: 'tags',
    oldValue: '{"env":"dev"}',
    newValue: '{"env":"prod"}',
    userIdentity: {
      userName: 'john.doe',
      type: 'IAMUser',
      arn: 'arn:aws:iam::123456789012:user/john.doe',
      accountId: '123456789012',
    },
    cloudtrailEventId: 'event-123',
    cloudtrailEventName: 'ModifyInstanceAttribute',
    sourceIP: '192.168.1.1',
    userAgent: 'aws-cli/2.0.0',
    tags: { environment: 'production', team: 'platform' },
  };

  it('should render null when no drift is selected', () => {
    const { container } = render(
      <DriftDetailPanel drift={null} />
    );
    expect(container.firstChild).toBeTruthy();
    expect(screen.getByText(/ドリフトイベントを選択してください/i)).toBeInTheDocument();
  });

  it('should render drift details when drift data is provided', () => {
    render(
      <DriftDetailPanel drift={mockDrift} />
    );
    expect(screen.getByText('web-server')).toBeInTheDocument();
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });

  it('should display resource information', () => {
    render(
      <DriftDetailPanel drift={mockDrift} />
    );
    expect(screen.getByText('i-1234567890abcdef0')).toBeInTheDocument();
    expect(screen.getByText('us-east-1')).toBeInTheDocument();
    expect(screen.getByText('AWS')).toBeInTheDocument();
  });

  it('should display user identity information', () => {
    render(
      <DriftDetailPanel drift={mockDrift} />
    );
    expect(screen.getByText('john.doe')).toBeInTheDocument();
    expect(screen.getByText('IAMUser')).toBeInTheDocument();
  });

  it('should display tags when present', () => {
    render(
      <DriftDetailPanel drift={mockDrift} />
    );
    expect(screen.getByText(/environment:/i)).toBeInTheDocument();
    expect(screen.getByText('production')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <DriftDetailPanel drift={mockDrift} onClose={onClose} />
    );
    const closeButton = document.querySelector('button svg');
    expect(closeButton).toBeInTheDocument();
  });
});
