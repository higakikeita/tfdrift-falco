import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { useSidebarStore } from '../../stores/sidebarStore';
import { act } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  LayoutDashboard: () => <div data-testid="dashboard-icon" />,
  AlertTriangle: () => <div data-testid="alert-icon" />,
  Network: () => <div data-testid="network-icon" />,
  Settings: () => <div data-testid="settings-icon" />,
  BarChart3: () => <div data-testid="chart-icon" />,
  ChevronLeft: () => <div data-testid="chevron-left" />,
  ChevronRight: () => <div data-testid="chevron-right" />,
  GitCompare: () => <div data-testid="git-compare" />,
  Activity: () => <div data-testid="activity-icon" />,
  Shield: () => <div data-testid="shield-icon" />,
}));

import { Sidebar } from './Sidebar';

const renderSidebar = () => {
  return render(
    <BrowserRouter>
      <Sidebar />
    </BrowserRouter>
  );
};

describe('Sidebar', () => {
  beforeEach(() => {
    act(() => {
      useSidebarStore.setState({ isCollapsed: false });
    });
  });

  it('should render without crashing', () => {
    renderSidebar();
    expect(document.querySelector('aside, nav, [role="navigation"]')).toBeTruthy();
  });

  it('should render navigation links', () => {
    renderSidebar();
    const links = document.querySelectorAll('a');
    expect(links.length).toBeGreaterThan(0);
  });

  it('should have dashboard link', () => {
    renderSidebar();
    const dashboardLink = document.querySelector('a[href="/"]') || document.querySelector('a[href="/dashboard"]');
    expect(dashboardLink).toBeTruthy();
  });
});
