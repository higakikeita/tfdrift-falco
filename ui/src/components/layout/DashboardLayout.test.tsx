import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';

vi.mock('./Sidebar', () => ({
  Sidebar: () => <div data-testid="sidebar">Sidebar</div>,
}));

vi.mock('./Header', () => ({
  Header: () => <div data-testid="header">Header</div>,
}));

vi.mock('../toast/ToastContainer', () => ({
  ToastContainer: () => <div data-testid="toast-container">Toast Container</div>,
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  };
});

import { DashboardLayout } from './DashboardLayout';

describe('DashboardLayout', () => {
  it('should render without crashing', () => {
    const { container } = render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should render sidebar component', () => {
    render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    expect(screen.getByTestId('sidebar')).toBeInTheDocument();
  });

  it('should render header component', () => {
    render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    expect(screen.getByTestId('header')).toBeInTheDocument();
  });

  it('should render outlet for page content', () => {
    render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    expect(screen.getByTestId('outlet')).toBeInTheDocument();
  });

  it('should render toast container', () => {
    render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    expect(screen.getByTestId('toast-container')).toBeInTheDocument();
  });

  it('should have flex layout structure', () => {
    const { container } = render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    const mainDiv = container.firstChild as HTMLElement;
    expect(mainDiv.className).toContain('flex');
    expect(mainDiv.className).toContain('h-screen');
  });

  it('should render children components in correct order', () => {
    const { container } = render(
      <BrowserRouter>
        <DashboardLayout />
      </BrowserRouter>
    );
    const elements = container.querySelectorAll('[data-testid]');
    const testIds = Array.from(elements).map(el => el.getAttribute('data-testid'));
    expect(testIds).toContain('sidebar');
    expect(testIds).toContain('header');
    expect(testIds).toContain('outlet');
    expect(testIds).toContain('toast-container');
  });
});
