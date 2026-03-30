import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  ExternalLink: () => <div data-testid="external-link-icon" />,
  Clock: () => <div data-testid="clock-icon" />,
  User: () => <div data-testid="user-icon" />,
  AlertTriangle: () => <div data-testid="alert-triangle-icon" />,
  CheckCircle: () => <div data-testid="check-circle-icon" />,
}));

import { NodeTooltip } from './NodeTooltip';

describe('NodeTooltip', () => {
  const mockData = {
    id: 'node-1',
    label: 'instance-1',
    type: 'resource',
    resourceType: 'aws_instance',
    resourceName: 'web-server',
    severity: 'high' as const,
    metadata: {
      mode: 'managed',
      provider: 'aws',
      tf_name: 'aws_instance.web_server',
      has_drift: true,
      last_modified: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
      user: 'alice',
      drift_count: 2,
    },
  };

  const mockPosition = { x: 100, y: 100 };

  it('should render without crashing', () => {
    const { container } = render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should display resource name and type', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('web-server')).toBeInTheDocument();
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });

  it('should display severity status', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('ステータス:')).toBeInTheDocument();
    expect(screen.getByText('ドリフト検出')).toBeInTheDocument();
  });

  it('should display drift information when available', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('ドリフト:')).toBeInTheDocument();
    expect(screen.getByText(/2件の変更を検出/i)).toBeInTheDocument();
  });

  it('should display last modified timestamp', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('最終更新:')).toBeInTheDocument();
  });

  it('should display user information', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('変更者:')).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
  });

  it('should display provider information', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText('プロバイダー:')).toBeInTheDocument();
    expect(screen.getByText('aws')).toBeInTheDocument();
  });

  it('should display resource ID', () => {
    render(
      <NodeTooltip data={mockData} position={mockPosition} />
    );
    expect(screen.getByText(/node-1/i)).toBeInTheDocument();
  });

  it('should handle data without metadata', () => {
    const minimalData = {
      id: 'node-2',
      label: 'resource-2',
      type: 'resource',
      resourceType: 'aws_s3_bucket',
      resourceName: 'my-bucket',
      severity: 'low' as const,
    };
    const { container } = render(
      <NodeTooltip data={minimalData} position={mockPosition} />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should display different severity levels', () => {
    const criticalData = {
      ...mockData,
      severity: 'critical' as const,
    };
    render(
      <NodeTooltip data={criticalData} position={mockPosition} />
    );
    expect(screen.getByText('重大なドリフト')).toBeInTheDocument();
  });

  it('should position tooltip at correct coordinates', () => {
    const { container } = render(
      <NodeTooltip data={mockData} position={{ x: 200, y: 300 }} />
    );
    const tooltip = container.querySelector('div');
    expect(tooltip?.style.left).toBe('220px');
    expect(tooltip?.style.top).toBe('300px');
  });
});
