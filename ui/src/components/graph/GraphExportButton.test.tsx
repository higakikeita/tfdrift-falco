import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  Download: () => <div data-testid="icon-Download" />,
  Image: () => <div data-testid="icon-Image" />,
  FileCode: () => <div data-testid="icon-FileCode" />,
  FileJson: () => <div data-testid="icon-FileJson" />,
  ChevronDown: () => <div data-testid="icon-ChevronDown" />,
}));

import { GraphExportButton } from './GraphExportButton';

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
