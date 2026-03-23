import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div data-testid="responsive-container">{children}</div>,
  LineChart: ({ children }: { children: React.ReactNode }) => <div data-testid="line-chart">{children}</div>,
  Line: () => <div data-testid="line" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="grid" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
  AreaChart: ({ children }: { children: React.ReactNode }) => <div data-testid="area-chart">{children}</div>,
  Area: () => <div data-testid="area" />,
}));

import { TimelineChart } from './TimelineChart';

const mockData = [
  { date: '03/20', aws: 5, gcp: 3 },
  { date: '03/21', aws: 8, gcp: 2 },
  { date: '03/22', aws: 3, gcp: 6 },
];

describe('TimelineChart', () => {
  it('should render without crashing', () => {
    const { container } = render(<TimelineChart data={mockData} />);
    expect(container.firstChild).toBeTruthy();
  });
});
