import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
}));

vi.mock('../api/sse', () => ({
  sseClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

vi.mock('../api/websocket', () => ({
  wsClient: { connect: vi.fn(), disconnect: vi.fn(), on: vi.fn(), off: vi.fn() },
}));

import ConnectionStatus from './ConnectionStatus';

describe('ConnectionStatus', () => {
  it('should render without crashing', () => {
    const { container } = render(<ConnectionStatus />);
    expect(container.firstChild).toBeTruthy();
  });
});
