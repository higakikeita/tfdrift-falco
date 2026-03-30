/**
 * Tests for DriftDetailPanel component
 */

import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import DriftDetailPanel from './DriftDetailPanel';
import type { DriftEvent } from '../types/drift';

// Mock icon libraries
vi.mock('react-icons/si', () => ({
  SiGooglecloud: ({ 'aria-label': ariaLabel }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'gcp-icon', 'aria-label': ariaLabel }, 'GCP Icon'),
}));

vi.mock('react-icons/fa', () => ({
  FaAws: ({ 'aria-label': ariaLabel }: Record<string, unknown>) =>
    React.createElement('div', { 'data-testid': 'aws-icon', 'aria-label': ariaLabel }, 'AWS Icon'),
}));

describe('DriftDetailPanel', () => {
  const mockDriftEvent = (overrides: Partial<DriftEvent> = {}): DriftEvent => ({
    id: 'drift-1',
    timestamp: '2024-01-15T10:30:45Z',
    severity: 'high',
    provider: 'aws',
    resourceType: 'aws_instance',
    resourceId: 'i-1234567890abcdef0',
    resourceName: 'production-webserver',
    changeType: 'modified',
    attribute: 'instance_type',
    oldValue: '"t2.micro"',
    newValue: '"t2.small"',
    userIdentity: {
      userName: 'alice.smith',
      type: 'IAMUser',
      arn: 'arn:aws:iam::123456789012:user/alice.smith',
      accountId: '123456789012',
    },
    region: 'us-east-1',
    cloudtrailEventId: 'evt-abc123',
    cloudtrailEventName: 'ModifyInstanceAttribute',
    sourceIP: '203.0.113.42',
    userAgent: 'aws-cli/2.0.0',
    tags: { Environment: 'production', Team: 'platform' },
    ...overrides,
  });

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Empty State', () => {
    it('renders placeholder when drift is null', () => {
      render(<DriftDetailPanel drift={null} />);

      expect(screen.getByText('ドリフトイベントを選択してください')).toBeInTheDocument();
      expect(screen.getByText('詳細情報を表示します')).toBeInTheDocument();
    });

    it('has appropriate role and aria attributes for empty state', () => {
      render(<DriftDetailPanel drift={null} />);

      const placeholder = screen.getByRole('status');
      expect(placeholder).toBeInTheDocument();
    });
  });

  describe('Basic Rendering', () => {
    it('renders drift details when drift is provided', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('production-webserver')).toBeInTheDocument();
      expect(screen.getByText('aws_instance')).toBeInTheDocument();
    });

    it('displays AWS icon for AWS provider', () => {
      const drift = mockDriftEvent({ provider: 'aws' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByTestId('aws-icon')).toBeInTheDocument();
    });

    it('displays GCP icon for GCP provider', () => {
      const drift = mockDriftEvent({ provider: 'gcp' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByTestId('gcp-icon')).toBeInTheDocument();
    });

    it('uses resourceId as fallback for resourceName', () => {
      const drift = mockDriftEvent({ id: 'drift-unnamed', resourceName: undefined });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getAllByText('i-1234567890abcdef0').length).toBeGreaterThan(0);
    });
  });

  describe('Severity Display', () => {
    it('displays severity badge for critical', () => {
      const drift = mockDriftEvent({ severity: 'critical' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    });

    it('displays severity badge for high', () => {
      const drift = mockDriftEvent({ severity: 'high' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('HIGH')).toBeInTheDocument();
    });

    it('displays severity badge for medium', () => {
      const drift = mockDriftEvent({ severity: 'medium' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('MEDIUM')).toBeInTheDocument();
    });

    it('displays severity badge for low', () => {
      const drift = mockDriftEvent({ severity: 'low' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('LOW')).toBeInTheDocument();
    });
  });

  describe('Change Type Display', () => {
    it('displays created indicator', () => {
      const drift = mockDriftEvent({ changeType: 'created' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('🆕 作成')).toBeInTheDocument();
    });

    it('displays modified indicator', () => {
      const drift = mockDriftEvent({ changeType: 'modified' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('✏️ 変更')).toBeInTheDocument();
    });

    it('displays deleted indicator', () => {
      const drift = mockDriftEvent({ changeType: 'deleted' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('🗑️ 削除')).toBeInTheDocument();
    });
  });

  describe('Timestamp Display', () => {
    it('formats and displays timestamp', () => {
      const drift = mockDriftEvent({ timestamp: '2024-01-15T10:30:45Z' });
      render(<DriftDetailPanel drift={drift} />);

      const detailPanel = screen.getByRole('complementary');
      expect(detailPanel.textContent).toMatch(/2024/);
    });
  });

  describe('Resource Information Section', () => {
    it('displays resource ID', () => {
      const drift = mockDriftEvent({ resourceId: 'i-0987654321fedcba0' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('i-0987654321fedcba0')).toBeInTheDocument();
    });

    it('displays region', () => {
      const drift = mockDriftEvent({ region: 'eu-west-1' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('eu-west-1')).toBeInTheDocument();
    });

    it('displays provider', () => {
      const drift = mockDriftEvent({ provider: 'aws' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('AWS')).toBeInTheDocument();
    });
  });

  describe('Change Details Section', () => {
    it('displays attribute', () => {
      const drift = mockDriftEvent({ attribute: 'instance_type' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('instance_type')).toBeInTheDocument();
    });

    it('displays old value as JSON when valid', () => {
      const drift = mockDriftEvent({ oldValue: '{"type": "t2.micro"}' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/"type"/)).toBeInTheDocument();
    });

    it('displays old value as plain text when not JSON', () => {
      const drift = mockDriftEvent({ oldValue: 't2.micro' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('t2.micro')).toBeInTheDocument();
    });

    it('displays new value as JSON when valid', () => {
      const drift = mockDriftEvent({ newValue: '{"type": "t2.small"}' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/"type"/)).toBeInTheDocument();
    });

    it('displays null values correctly', () => {
      const drift = mockDriftEvent({ oldValue: null });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('null')).toBeInTheDocument();
    });
  });

  describe('User Identity Section', () => {
    it('displays user name', () => {
      const drift = mockDriftEvent({
        id: 'drift-bob',
        userIdentity: { userName: 'bob.jones', type: 'IAMUser' },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('bob.jones')).toBeInTheDocument();
    });

    it('displays user type', () => {
      const drift = mockDriftEvent({
        id: 'drift-role',
        userIdentity: { userName: 'test-user', type: 'IAMRole' },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('IAMRole')).toBeInTheDocument();
    });

    it('displays ARN when provided', () => {
      const drift = mockDriftEvent({
        id: 'drift-arn',
        userIdentity: {
          userName: 'test',
          type: 'IAMUser',
          arn: 'arn:aws:iam::123456789012:user/test',
        },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/arn:aws:iam::/)).toBeInTheDocument();
    });

    it('hides ARN when not provided', () => {
      const drift = mockDriftEvent({
        id: 'drift-noarn',
        userIdentity: { userName: 'test', type: 'IAMUser' },
      });
      render(<DriftDetailPanel drift={drift} />);

      const arnElements = screen.queryAllByText(/ARN:/);
      expect(arnElements.length).toBe(0);
    });

    it('displays account ID when provided', () => {
      const drift = mockDriftEvent({
        id: 'drift-acct',
        userIdentity: { userName: 'test', type: 'IAMUser', accountId: '123456789012' },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('123456789012')).toBeInTheDocument();
    });

    it('hides account ID when not provided', () => {
      const drift = mockDriftEvent({
        id: 'drift-noacct',
        userIdentity: { userName: 'test', type: 'IAMUser' },
      });
      render(<DriftDetailPanel drift={drift} />);

      const accountIdElements = screen.queryAllByText(/Account ID:/);
      expect(accountIdElements.length).toBe(0);
    });
  });

  describe('CloudTrail/Audit Log Section', () => {
    it('displays CloudTrail section when cloudtrailEventId is provided', () => {
      const drift = mockDriftEvent({ cloudtrailEventId: 'evt-123' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/CloudTrail/)).toBeInTheDocument();
      expect(screen.getByText('evt-123')).toBeInTheDocument();
    });

    it('hides CloudTrail section when cloudtrailEventId is not provided', () => {
      const drift = mockDriftEvent({ cloudtrailEventId: undefined });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.queryByText(/CloudTrail/)).not.toBeInTheDocument();
    });

    it('displays event name when provided', () => {
      const drift = mockDriftEvent({
        cloudtrailEventId: 'evt-123',
        cloudtrailEventName: 'ModifyInstanceAttribute',
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('ModifyInstanceAttribute')).toBeInTheDocument();
    });

    it('displays source IP when provided', () => {
      const drift = mockDriftEvent({
        cloudtrailEventId: 'evt-123',
        sourceIP: '203.0.113.42',
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('203.0.113.42')).toBeInTheDocument();
    });

    it('displays user agent when provided', () => {
      const drift = mockDriftEvent({
        cloudtrailEventId: 'evt-123',
        userAgent: 'aws-cli/2.0.0',
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('aws-cli/2.0.0')).toBeInTheDocument();
    });
  });

  describe('Tags Section', () => {
    it('displays tags when provided', () => {
      const drift = mockDriftEvent({
        id: 'drift-tags',
        tags: { Environment: 'production', Team: 'platform' },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/Environment:/)).toBeInTheDocument();
      expect(screen.getByText('production')).toBeInTheDocument();
      expect(screen.getByText(/Team:/)).toBeInTheDocument();
      expect(screen.getByText('platform')).toBeInTheDocument();
    });

    it('hides tags section when no tags provided', () => {
      const drift = mockDriftEvent({ id: 'drift-notags', tags: {} });
      render(<DriftDetailPanel drift={drift} />);

      const tagHeaders = screen.queryAllByText('タグ');
      expect(tagHeaders.length).toBe(0);
    });

    it('hides tags section when tags is undefined', () => {
      const drift = mockDriftEvent({ id: 'drift-notags2', tags: undefined });
      render(<DriftDetailPanel drift={drift} />);

      const tagHeaders = screen.queryAllByText('タグ');
      expect(tagHeaders.length).toBe(0);
    });
  });

  describe('Recommended Actions Section', () => {
    it('displays recommended actions', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText(/推奨アクション/)).toBeInTheDocument();
      expect(screen.getByText(/変更内容をユーザー/)).toBeInTheDocument();
      expect(screen.getByText(/意図的な変更の場合/)).toBeInTheDocument();
      expect(screen.getByText(/不正な変更の場合/)).toBeInTheDocument();
    });

    it('includes username in recommended actions', () => {
      const drift = mockDriftEvent({
        id: 'drift-charlie',
        userIdentity: { userName: 'charlie.brown', type: 'IAMUser' },
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getAllByText(/charlie.brown/).length).toBeGreaterThan(0);
    });
  });

  describe('Close Button', () => {
    it('renders close button when onClose is provided', () => {
      const drift = mockDriftEvent();
      const onClose = vi.fn();
      render(<DriftDetailPanel drift={drift} onClose={onClose} />);

      const closeButton = screen.getByRole('button', { name: /Close drift details panel/ });
      expect(closeButton).toBeInTheDocument();
    });

    it('does not render close button when onClose is not provided', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      const closeButton = screen.queryByRole('button', { name: /Close drift details panel/ });
      expect(closeButton).not.toBeInTheDocument();
    });

    it('calls onClose when close button is clicked', async () => {
      const user = userEvent.setup();
      const drift = mockDriftEvent();
      const onClose = vi.fn();
      render(<DriftDetailPanel drift={drift} onClose={onClose} />);

      const closeButton = screen.getByRole('button', { name: /Close drift details panel/ });
      await user.click(closeButton);

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Keyboard Navigation', () => {
    it('renders properly with onClose handler', () => {
      const drift = mockDriftEvent();
      const onClose = vi.fn();
      render(<DriftDetailPanel drift={drift} onClose={onClose} />);

      expect(screen.getByRole('button', { name: /Close drift details panel/ })).toBeInTheDocument();
    });

    it('does not call onClose when onClose is not provided', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      // Should render without close button
      const closeButton = screen.queryByRole('button', { name: /Close drift details panel/ });
      expect(closeButton).not.toBeInTheDocument();
    });

    it('attaches escape key listener when onClose is provided', () => {
      const drift = mockDriftEvent();
      const onClose = vi.fn();
      const addEventListenerSpy = vi.spyOn(window, 'addEventListener');

      render(<DriftDetailPanel drift={drift} onClose={onClose} />);

      expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));

      addEventListenerSpy.mockRestore();
    });
  });

  describe('Accessibility', () => {
    it('has proper complementary role', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      const panel = screen.getByRole('complementary');
      expect(panel).toBeInTheDocument();
    });

    it('has aria-label with resource information', () => {
      const drift = mockDriftEvent({
        id: 'drift-aria',
        resourceName: 'my-resource',
        resourceId: 'res-123',
      });
      render(<DriftDetailPanel drift={drift} />);

      const panel = screen.getByRole('complementary');
      expect(panel).toHaveAttribute('aria-label', expect.stringContaining('my-resource'));
    });

    it('has aria-label with resourceId when resourceName is missing', () => {
      const drift = mockDriftEvent({
        id: 'drift-aria2',
        resourceName: undefined,
        resourceId: 'res-123',
      });
      render(<DriftDetailPanel drift={drift} />);

      const panel = screen.getByRole('complementary');
      expect(panel).toHaveAttribute('aria-label', expect.stringContaining('res-123'));
    });

    it('proper heading hierarchy', () => {
      const drift = mockDriftEvent();
      render(<DriftDetailPanel drift={drift} />);

      const heading = screen.getByRole('heading', { level: 2 });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('handles very long values gracefully', () => {
      const drift = mockDriftEvent({
        id: 'drift-long',
        oldValue: JSON.stringify({ key: 'a'.repeat(100) }),
      });
      const { container } = render(<DriftDetailPanel drift={drift} />);

      // Check that value is rendered (in pre tag for JSON)
      const preElement = container.querySelector('pre');
      expect(preElement).toBeInTheDocument();
    });

    it('handles malformed JSON gracefully', () => {
      const drift = mockDriftEvent({ id: 'drift-invalid', oldValue: '{invalid json}' });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('{invalid json}')).toBeInTheDocument();
    });

    it('handles change from null to value', () => {
      const drift = mockDriftEvent({
        id: 'drift-null-to-val',
        oldValue: null,
        newValue: 'new-value',
      });
      render(<DriftDetailPanel drift={drift} />);

      expect(screen.getByText('null')).toBeInTheDocument();
      expect(screen.getByText('new-value')).toBeInTheDocument();
    });
  });
});
