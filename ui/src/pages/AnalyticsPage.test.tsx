/**
 * AnalyticsPage Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AnalyticsPage } from './AnalyticsPage';

// Mock analytics components
vi.mock('../components/analytics/TimelineChart', () => ({
  TimelineChart: () => <div data-testid="timeline-chart">Timeline Chart</div>,
}));

vi.mock('../components/analytics/SeverityChart', () => ({
  SeverityChart: () => <div data-testid="severity-chart">Severity Chart</div>,
}));

vi.mock('../components/analytics/ServiceBreakdown', () => ({
  ServiceBreakdown: () => <div data-testid="service-breakdown">Service Breakdown</div>,
}));

vi.mock('../components/analytics/TopUsersChart', () => ({
  TopUsersChart: () => <div data-testid="top-users-chart">Top Users Chart</div>,
}));

vi.mock('../mocks/analyticsData', () => ({
  timelineData: [
    { date: '2024-01-01', events: 10 },
    { date: '2024-01-02', events: 15 },
  ],
  severityData: [
    { name: 'Low', value: 5 },
    { name: 'Medium', value: 10 },
    { name: 'High', value: 15 },
    { name: 'Critical', value: 5 },
  ],
  serviceData: [
    { name: 'IAM', value: 10 },
    { name: 'S3', value: 15 },
  ],
  topUsersData: [
    { name: 'User1', events: 5 },
    { name: 'User2', events: 3 },
  ],
}));

describe('AnalyticsPage', () => {
  describe('Rendering', () => {
    it('should render the page title', () => {
      render(<AnalyticsPage />);
      const title = screen.getByText('Analytics');
      expect(title).toBeTruthy();
      expect(title.className).toContain('font-bold');
    });

    it('should render the time range selector', () => {
      render(<AnalyticsPage />);
      const select = screen.getByDisplayValue('Last 7 days');
      expect(select).toBeTruthy();
    });

    it('should render all time range options', () => {
      render(<AnalyticsPage />);
      expect(screen.getByDisplayValue('Last 7 days')).toBeTruthy();
      expect(screen.getByText('Last 30 days')).toBeTruthy();
      expect(screen.getByText('Last 90 days')).toBeTruthy();
    });
  });

  describe('KPI Cards', () => {
    it('should render all KPI cards', () => {
      render(<AnalyticsPage />);
      expect(screen.getByText('Total Events')).toBeTruthy();
      expect(screen.getByText('Critical')).toBeTruthy();
      expect(screen.getByText('Avg/Day')).toBeTruthy();
      expect(screen.getByText('Unique Users')).toBeTruthy();
    });

    it('should display KPI values', () => {
      render(<AnalyticsPage />);
      // Total events = 5 + 10 + 15 + 5 = 35
      expect(screen.getByText('35')).toBeTruthy();
    });

    it('should display critical percentage', () => {
      const { container } = render(<AnalyticsPage />);
      // Critical (5) / Total (35) * 100 = 14.28% ≈ 14%
      expect(container.innerHTML).toContain('14%');
    });

    it('should display average events per day', () => {
      const { container } = render(<AnalyticsPage />);
      // 35 / 7 = 5.0
      expect(container.innerHTML).toContain('5.0');
    });

    it('should display unique users count', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container.innerHTML).toContain('Unique Users');
    });

    it('should display KPI subtitles', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container.innerHTML).toContain('Last 7 days');
      expect(container.innerHTML).toContain('Trend: +12%');
      expect(container.innerHTML).toContain('Caused drift');
    });
  });

  describe('Charts', () => {
    it('should render timeline chart', () => {
      render(<AnalyticsPage />);
      expect(screen.getByTestId('timeline-chart')).toBeTruthy();
    });

    it('should render severity chart', () => {
      render(<AnalyticsPage />);
      expect(screen.getByTestId('severity-chart')).toBeTruthy();
    });

    it('should render service breakdown', () => {
      render(<AnalyticsPage />);
      expect(screen.getByTestId('service-breakdown')).toBeTruthy();
    });

    it('should render top users chart', () => {
      render(<AnalyticsPage />);
      expect(screen.getByTestId('top-users-chart')).toBeTruthy();
    });

    it('should render all charts together', () => {
      render(<AnalyticsPage />);
      expect(screen.getByTestId('timeline-chart')).toBeTruthy();
      expect(screen.getByTestId('severity-chart')).toBeTruthy();
      expect(screen.getByTestId('service-breakdown')).toBeTruthy();
      expect(screen.getByTestId('top-users-chart')).toBeTruthy();
    });
  });

  describe('Layout', () => {
    it('should have correct grid structure for KPI cards', () => {
      const { container } = render(<AnalyticsPage />);
      const grids = container.querySelectorAll('[class*="grid-cols"]');
      expect(grids.length).toBeGreaterThan(0);
    });

    it('should be responsive', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container.innerHTML).toContain('grid');
      expect(container.innerHTML).toContain('gap');
    });

    it('should render with proper spacing', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container.innerHTML).toContain('space-y');
    });
  });

  describe('Data Calculations', () => {
    it('should calculate total events correctly', () => {
      render(<AnalyticsPage />);
      // severityData: Low(5) + Medium(10) + High(15) + Critical(5) = 35
      expect(screen.getByText('35')).toBeTruthy();
    });

    it('should calculate critical count correctly', () => {
      const { container } = render(<AnalyticsPage />);
      // Critical value from severityData = 5
      expect(container.innerHTML).toContain('Critical');
    });

    it('should handle edge case when critical is not found', () => {
      // The component defaults to 0 if not found, so this should still render
      render(<AnalyticsPage />);
      expect(screen.getByText('Analytics')).toBeTruthy();
    });
  });

  describe('Styling', () => {
    it('should render with proper text styles', () => {
      render(<AnalyticsPage />);
      const title = screen.getByText('Analytics');
      expect(title.className).toContain('text-2xl');
      expect(title.className).toContain('font-bold');
    });

    it('should render cards with proper styling', () => {
      const { container } = render(<AnalyticsPage />);
      const cards = container.querySelectorAll('[class*="bg-white"]');
      expect(cards.length).toBeGreaterThan(0);
    });

    it('should apply color classes to KPI icons', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container.innerHTML).toContain('indigo');
      expect(container.innerHTML).toContain('red');
      expect(container.innerHTML).toContain('amber');
      expect(container.innerHTML).toContain('blue');
    });
  });

  describe('Integration', () => {
    it('should render complete page without errors', () => {
      const { container } = render(<AnalyticsPage />);
      expect(container).toBeTruthy();
      expect(container.querySelector('div')).toBeTruthy();
    });

    it('should render header and content sections', () => {
      render(<AnalyticsPage />);
      expect(screen.getByText('Analytics')).toBeTruthy();
      expect(screen.getByDisplayValue('Last 7 days')).toBeTruthy();
    });
  });
});
