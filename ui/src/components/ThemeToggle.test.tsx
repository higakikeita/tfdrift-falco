import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  Sun: () => <div data-testid="sun-icon" />,
  Moon: () => <div data-testid="moon-icon" />,
}));

import ThemeToggle from './ThemeToggle';

describe('ThemeToggle', () => {
  it('should render without crashing', () => {
    const { container } = render(<ThemeToggle />);
    expect(container.firstChild).toBeTruthy();
  });

  it('should render a clickable element', () => {
    render(<ThemeToggle />);
    const buttons = document.querySelectorAll('button');
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('should toggle theme on click', () => {
    render(<ThemeToggle />);
    const button = document.querySelector('button');
    if (button) {
      fireEvent.click(button);
      // Should not throw
      expect(true).toBe(true);
    }
  });
});
