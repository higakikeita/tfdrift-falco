import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SaveButton } from './SaveButton';
import { Save, Check } from 'lucide-react';

describe('SaveButton component', () => {
  it('should render button with default label', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button', { name: /save/i });
    expect(button).toBeInTheDocument();
  });

  it('should render button with custom label', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} label="Apply" />);

    const button = screen.getByRole('button', { name: /apply/i });
    expect(button).toBeInTheDocument();
  });

  it('should render with default icon', () => {
    const handleClick = vi.fn();
    const { container } = render(<SaveButton onClick={handleClick} />);

    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should render with custom icon', () => {
    const handleClick = vi.fn();
    const { container } = render(<SaveButton onClick={handleClick} icon={Check} />);

    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should call onClick handler when clicked', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    await user.click(button);

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('should be enabled by default', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    expect(button).not.toBeDisabled();
  });

  it('should be disabled when disabled prop is true', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} disabled={true} />);

    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
  });

  it('should not trigger onClick when disabled', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} disabled={true} />);

    const button = screen.getByRole('button');
    await user.click(button);

    expect(handleClick).not.toHaveBeenCalled();
  });

  it('should have correct styling classes', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    expect(button).toHaveClass('inline-flex');
    expect(button).toHaveClass('items-center');
    expect(button).toHaveClass('gap-1.5');
    expect(button).toHaveClass('px-4');
    expect(button).toHaveClass('py-2');
    expect(button).toHaveClass('text-sm');
    expect(button).toHaveClass('font-medium');
    expect(button).toHaveClass('text-white');
    expect(button).toHaveClass('bg-indigo-600');
    expect(button).toHaveClass('rounded-lg');
    expect(button).toHaveClass('transition-colors');
  });

  it('should have hover state', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    expect(button).toHaveClass('hover:bg-indigo-700');
  });

  it('should have disabled opacity style', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} disabled={true} />);

    const button = screen.getByRole('button');
    expect(button).toHaveClass('disabled:opacity-50');
    expect(button).toHaveClass('disabled:cursor-not-allowed');
  });

  it('should render icon with correct size', () => {
    const handleClick = vi.fn();
    const { container } = render(<SaveButton onClick={handleClick} />);

    const svg = container.querySelector('svg');
    expect(svg).toHaveClass('h-4');
    expect(svg).toHaveClass('w-4');
  });

  it('should render label and icon together', () => {
    const handleClick = vi.fn();
    const { container } = render(<SaveButton onClick={handleClick} label="Save Changes" />);

    const button = screen.getByRole('button');
    const svg = container.querySelector('svg');

    expect(button).toHaveTextContent('Save Changes');
    expect(svg).toBeInTheDocument();
  });

  it('should handle multiple clicks', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    await user.click(button);
    await user.click(button);
    await user.click(button);

    expect(handleClick).toHaveBeenCalledTimes(3);
  });

  it('should work with different icon props', () => {
    const handleClick = vi.fn();
    const { container: container1 } = render(
      <SaveButton onClick={handleClick} icon={Save} />
    );
    const { container: container2 } = render(
      <SaveButton onClick={handleClick} icon={Check} />
    );

    expect(container1.querySelector('svg')).toBeInTheDocument();
    expect(container2.querySelector('svg')).toBeInTheDocument();
  });

  it('should support custom label styles through icon', () => {
    const handleClick = vi.fn();
    render(
      <SaveButton
        onClick={handleClick}
        label="Delete"
        icon={Check}
      />
    );

    const button = screen.getByRole('button', { name: /delete/i });
    expect(button).toBeInTheDocument();
  });

  it('should render as inline-flex for proper alignment', () => {
    const handleClick = vi.fn();
    render(<SaveButton onClick={handleClick} />);

    const button = screen.getByRole('button');
    expect(button).toHaveClass('inline-flex');
  });
});
