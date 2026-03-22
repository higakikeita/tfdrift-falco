import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  Wifi: () => <div data-testid="wifi-icon" />,
  WifiOff: () => <div data-testid="wifioff-icon" />,
  Activity: () => <div data-testid="activity-icon" />,
  Radio: () => <div data-testid="radio-icon" />,
}));

vi.mock('../api/sse', () => ({
  useSSE: () => ({
    isConnected: true,
    lastEvent: null,
    events: [],
    connect: vi.fn(),
    disconnect: vi.fn(),
  }),
  sseClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

vi.mock('../api/websocket', () => ({
  useWebSocket: () => ({
    isConnected: false,
    lastMessage: null,
    messages: [],
    connect: vi.fn(),
    disconnect: vi.fn(),
    send: vi.fn(),
  }),
  wsClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

import { ConnectionStatus } from './ConnectionStatus';

describe('ConnectionStatus', () => {
  it('should render without crashing', () => {
    const { container } = render(<ConnectionStatus />);
    expect(container.firstChild).toBeTruthy();
  });
});
