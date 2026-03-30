/**
 * LoadingSpinner Component Tests
 */

import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LoadingSpinner from './LoadingSpinner';

describe('LoadingSpinner', () => {
  it('renders spinner with md size by default', () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
    expect(spinner).toHaveClass('w-8', 'h-8');
  });

  it('renders spinner with custom size', () => {
    const { container } = render(<LoadingSpinner size="lg" />);
    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
    expect(spinner).toHaveClass('w-12', 'h-12');
  });

  it('renders with text when provided', () => {
    render(<LoadingSpinner text="Loading..." />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('does not render text when not provided', () => {
    render(<LoadingSpinner />);
    const text = screen.queryByText(/^Loading/);
    expect(text).not.toBeInTheDocument();
  });

  it('renders small spinner', () => {
    const { container } = render(<LoadingSpinner size="sm" />);
    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toHaveClass('w-4', 'h-4');
  });

  it('applies correct border colors', () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector('.animate-spin');
    expect(spinner).toHaveClass('border-gray-300', 'border-t-blue-500');
  });
});
