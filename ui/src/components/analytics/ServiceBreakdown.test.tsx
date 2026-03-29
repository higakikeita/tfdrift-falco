import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  BarChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="bar-chart">{children}</div>
  ),
  Bar: () => <div data-testid="bar" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
}));

import { ServiceBreakdown } from './ServiceBreakdown';

describe('ServiceBreakdown', () => {
  const mockData = [
    { service: 'EC2', count: 45 },
    { service: 'S3', count: 28 },
    { service: 'RDS', count: 15 },
  ];

  it('should render without crashing', () => {
    const { container } = render(<ServiceBreakdown data={mockData} />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render chart title', () => {
    const { getByText } = render(<ServiceBreakdown data={mockData} />);
    expect(getByText('Top Affected Services')).toBeInTheDocument();
  });

  it('should render chart components', () => {
    const { getByTestId } = render(<ServiceBreakdown data={mockData} />);
    expect(getByTestId('responsive-container')).toBeInTheDocument();
    expect(getByTestId('bar-chart')).toBeInTheDocument();
    expect(getByTestId('bar')).toBeInTheDocument();
  });

  it('should handle empty data', () => {
    const { container } = render(<ServiceBreakdown data={[]} />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render with various data sizes', () => {
    const largeData = Array.from({ length: 10 }, (_, i) => ({
      service: `Service${i}`,
      count: Math.floor(Math.random() * 100),
    }));
    const { container } = render(<ServiceBreakdown data={largeData} />);
    expect(container.firstChild).toBeTruthy();
  });
});
