import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('lucide-react', () => ({
  Bell: () => <div data-testid="bell-icon" />,
  Menu: () => <div data-testid="menu-icon" />,
  Search: () => <div data-testid="search-icon" />,
  Sun: () => <div data-testid="sun-icon" />,
  Moon: () => <div data-testid="moon-icon" />,
  Shield: () => <div data-testid="shield-icon" />,
  ShieldAlert: () => <div data-testid="shield-alert-icon" />,
  X: () => <div data-testid="x-icon" />,
  AlertTriangle: () => <div data-testid="alert-icon" />,
  Info: () => <div data-testid="info-icon" />,
  CheckCircle: () => <div data-testid="check-icon" />,
  AlertCircle: () => <div data-testid="alert-circle-icon" />,
  ChevronDown: () => <div data-testid="chevron-down" />,
  ExternalLink: () => <div data-testid="external-link" />,
}));

vi.mock('../../api/sse', () => ({
  useSSE: () => ({
    isConnected: true,
    lastEvent: null,
    events: [],
    connect: vi.fn(),
    disconnect: vi.fn(),
  }),
  sseClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
  },
}));

vi.mock('../../hooks/useTheme', () => ({
  useTheme: () => ({ theme: 'light', setTheme: vi.fn(), toggleTheme: vi.fn() }),
}));

import { Header } from './Header';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false } },
});

const renderHeader = () => {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Header />
      </BrowserRouter>
    </QueryClientProvider>
  );
};

describe('Header', () => {
  it('should render without crashing', () => {
    renderHeader();
    expect(document.querySelector('header')).toBeTruthy();
  });

  it('should display the app name', () => {
    renderHeader();
    expect(screen.getByText(/TFDrift/i)).toBeTruthy();
  });
});
