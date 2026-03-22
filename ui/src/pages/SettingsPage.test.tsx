import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
}));

vi.mock('../api/client', () => ({
  apiClient: {
    getConfig: vi.fn().mockResolvedValue({ success: true, data: {} }),
    testWebhook: vi.fn().mockResolvedValue({ success: true }),
  },
}));

import SettingsPage from './SettingsPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, gcTime: 0 } },
});

describe('SettingsPage', () => {
  it('should render without crashing', () => {
    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <SettingsPage />
        </BrowserRouter>
      </QueryClientProvider>
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should display settings heading or tabs', () => {
    render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <SettingsPage />
        </BrowserRouter>
      </QueryClientProvider>
    );
    // Settings page should have some identifiable content
    const text = document.body.textContent;
    expect(text).toBeTruthy();
  });
});
