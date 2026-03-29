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

import { TopUsersChart } from './TopUsersChart';

describe('TopUsersChart', () => {
  const mockData = [
    { user: 'alice', events: 42 },
    { user: 'bob', events: 28 },
    { user: 'charlie', events: 15 },
  ];

  it('should render without crashing', () => {
    const { container } = render(<TopUsersChart data={mockData} />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render chart title', () => {
    const { getByText } = render(<TopUsersChart data={mockData} />);
    expect(getByText('Top Drift-Causing Users')).toBeInTheDocument();
  });

  it('should render chart components', () => {
    const { getByTestId } = render(<TopUsersChart data={mockData} />);
    expect(getByTestId('responsive-container')).toBeInTheDocument();
    expect(getByTestId('bar-chart')).toBeInTheDocument();
    expect(getByTestId('bar')).toBeInTheDocument();
    expect(getByTestId('x-axis')).toBeInTheDocument();
    expect(getByTestId('y-axis')).toBeInTheDocument();
  });

  it('should handle empty data', () => {
    const { container } = render(<TopUsersChart data={[]} />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render with single user', () => {
    const singleUser = [{ user: 'alice', events: 100 }];
    const { container } = render(<TopUsersChart data={singleUser} />);
    expect(container.firstChild).toBeTruthy();
  });
});
