import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div data-testid="responsive-container">{children}</div>,
  PieChart: ({ children }: { children: React.ReactNode }) => <div data-testid="pie-chart">{children}</div>,
  Pie: () => <div data-testid="pie" />,
  Cell: () => <div data-testid="cell" />,
  Tooltip: () => <div data-testid="tooltip" />,
  Legend: () => <div data-testid="legend" />,
  BarChart: ({ children }: { children: React.ReactNode }) => <div data-testid="bar-chart">{children}</div>,
  Bar: () => <div data-testid="bar" />,
  XAxis: () => <div data-testid="x-axis" />,
  YAxis: () => <div data-testid="y-axis" />,
  CartesianGrid: () => <div data-testid="grid" />,
}));

import { SeverityChart } from './SeverityChart';

const mockData = [
  { name: 'Critical', value: 5, fill: '#ef4444' },
  { name: 'High', value: 10, fill: '#f97316' },
  { name: 'Medium', value: 15, fill: '#eab308' },
  { name: 'Low', value: 20, fill: '#3b82f6' },
];

describe('SeverityChart', () => {
  it('should render without crashing', () => {
    const { container } = render(<SeverityChart data={mockData} />);
    expect(container.firstChild).toBeTruthy();
  });
});
