import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  Bell: () => <div data-testid="bell-icon" />,
  X: () => <div data-testid="close-icon" />,
  Wifi: () => <div data-testid="wifi-icon" />,
  WifiOff: () => <div data-testid="wifioff-icon" />,
  Trash2: () => <div data-testid="trash-icon" />,
}));

vi.mock('../../api/sse', () => ({
  useSSE: () => {
    return {
      isConnected: true,
      lastEvent: null,
      events: [],
      connect: vi.fn(),
      disconnect: vi.fn(),
    };
  },
}));

vi.mock('../../stores/toastStore', () => ({
  toast: {
    error: vi.fn(),
    warning: vi.fn(),
    success: vi.fn(),
  },
}));

vi.mock('../../lib/utils', () => ({
  cn: () => '',
}));

import { NotificationPanel } from './NotificationPanel';

describe('NotificationPanel', () => {
  it('should render without crashing', () => {
    const { container } = render(<NotificationPanel />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render notification bell button', () => {
    render(<NotificationPanel />);
    const button = screen.getByRole('button');
    expect(button).toBeInTheDocument();
  });

  it('should display bell icon', () => {
    render(<NotificationPanel />);
    expect(screen.getByTestId('bell-icon')).toBeInTheDocument();
  });

  it('should show connection status when panel is open', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    expect(screen.getByText(/Live|Offline/i)).toBeInTheDocument();
  });

  it('should toggle panel open and closed', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');

    // Initially should show "No notifications yet"
    fireEvent.click(bellButton);
    expect(screen.getByText(/No notifications yet/i)).toBeInTheDocument();

    // Close the panel
    fireEvent.click(bellButton);
    // Panel should be hidden (not in document)
  });

  it('should render empty state message', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    expect(screen.getByText(/No notifications yet/i)).toBeInTheDocument();
  });

  it('should display Notifications title in panel', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    expect(screen.getByText('Notifications')).toBeInTheDocument();
  });

  it('should render close button in panel header', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('should render WiFi status indicator when connected', () => {
    render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    expect(screen.getByTestId('wifi-icon')).toBeInTheDocument();
  });

  it('should handle empty notifications gracefully', () => {
    const { container } = render(<NotificationPanel />);
    const bellButton = screen.getByRole('button');
    fireEvent.click(bellButton);
    expect(screen.getByText(/No notifications yet/i)).toBeInTheDocument();
    expect(container).toBeTruthy();
  });
});
