import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Button, buttonVariants } from './button';

describe('Button component', () => {
  it('should render a button', () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
  });

  it('should apply default variant and size', () => {
    render(<Button>Default</Button>);
    const button = screen.getByRole('button', { name: /default/i });
    expect(button).toHaveClass('bg-primary');
    expect(button).toHaveClass('h-10');
  });

  it('should apply variant prop', () => {
    render(<Button variant="destructive">Delete</Button>);
    const button = screen.getByRole('button', { name: /delete/i });
    expect(button).toHaveClass('bg-destructive');
  });

  it('should apply destructive variant', () => {
    render(<Button variant="destructive">Delete</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('text-destructive-foreground');
  });

  it('should apply outline variant', () => {
    render(<Button variant="outline">Outline</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('border');
    expect(button).toHaveClass('bg-background');
  });

  it('should apply secondary variant', () => {
    render(<Button variant="secondary">Secondary</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('bg-secondary');
  });

  it('should apply ghost variant', () => {
    render(<Button variant="ghost">Ghost</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('hover:bg-accent');
  });

  it('should apply link variant', () => {
    render(<Button variant="link">Link</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('text-primary');
    expect(button).toHaveClass('underline-offset-4');
  });

  it('should apply small size', () => {
    render(<Button size="sm">Small</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('h-9');
    expect(button).toHaveClass('px-3');
  });

  it('should apply large size', () => {
    render(<Button size="lg">Large</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('h-11');
    expect(button).toHaveClass('px-8');
  });

  it('should apply icon size', () => {
    render(<Button size="icon">🔍</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('h-10');
    expect(button).toHaveClass('w-10');
  });

  it('should be disabled when disabled prop is true', () => {
    render(<Button disabled>Disabled</Button>);
    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
    expect(button).toHaveClass('disabled:opacity-50');
  });

  it('should accept custom className', () => {
    render(<Button className="custom-class">Custom</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('custom-class');
  });

  it('should forward ref', () => {
    let ref: HTMLButtonElement | null = null;
    render(<Button ref={(el) => { ref = el; }}>Ref</Button>);
    expect(ref).toBeInstanceOf(HTMLButtonElement);
  });

  it('should render as a different component with asChild', () => {
    render(
      <Button asChild>
        <a href="/link">Link Button</a>
      </Button>
    );
    const link = screen.getByRole('link', { name: /link button/i });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', '/link');
  });

  it('should have focus visible styles', () => {
    render(<Button>Focus</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('focus-visible:ring-2');
  });

  it('should support onClick handler', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click</Button>);
    const button = screen.getByRole('button');
    button.click();
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('should display icon with text', () => {
    render(<Button>📁 Open</Button>);
    const button = screen.getByRole('button', { name: /open/i });
    expect(button).toHaveTextContent('Open');
  });
});

describe('buttonVariants', () => {
  it('should export buttonVariants CVA', () => {
    expect(buttonVariants).toBeDefined();
  });

  it('should generate correct classes for variant and size combinations', () => {
    const classes = buttonVariants({ variant: 'destructive', size: 'lg' });
    expect(classes).toContain('bg-destructive');
    expect(classes).toContain('h-11');
  });

  it('should apply default variant and size', () => {
    const classes = buttonVariants({});
    expect(classes).toContain('bg-primary');
    expect(classes).toContain('h-10');
  });
});
