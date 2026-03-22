import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => new Proxy({}, {
  get: (_, name) => () => <div data-testid={`icon-${String(name)}`} />,
}));

import GraphExportButton from './GraphExportButton';

describe('GraphExportButton', () => {
  it('should render without crashing', () => {
    const { container } = render(<GraphExportButton />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render a button element', () => {
    render(<GraphExportButton />);
    const buttons = document.querySelectorAll('button');
    expect(buttons.length).toBeGreaterThan(0);
  });
});
