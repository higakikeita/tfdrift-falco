import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from './card';

describe('Card component', () => {
  it('should render Card', () => {
    const { container } = render(<Card>Test Card</Card>);
    const card = container.querySelector('div');
    expect(card).toHaveClass('rounded-lg');
    expect(card).toHaveClass('border');
    expect(card).toHaveClass('bg-card');
    expect(card).toHaveClass('shadow-sm');
  });

  it('should accept custom className', () => {
    const { container } = render(<Card className="custom-class">Card</Card>);
    const card = container.firstChild as HTMLElement;
    expect(card).toHaveClass('custom-class');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    const { container } = render(<Card ref={(el) => { ref = el; }}>Card</Card>);
    expect(ref).toBeInstanceOf(HTMLDivElement);
    expect(ref).toBe(container.firstChild);
  });

  it('should render children', () => {
    render(
      <Card>
        <div>Child content</div>
      </Card>
    );
    expect(screen.getByText('Child content')).toBeInTheDocument();
  });
});

describe('CardHeader component', () => {
  it('should render CardHeader', () => {
    const { container } = render(<CardHeader>Header</CardHeader>);
    const header = container.querySelector('div');
    expect(header).toHaveClass('flex');
    expect(header).toHaveClass('flex-col');
    expect(header).toHaveClass('space-y-1.5');
    expect(header).toHaveClass('p-6');
  });

  it('should accept custom className', () => {
    const { container } = render(<CardHeader className="custom-padding">Header</CardHeader>);
    const header = container.firstChild as HTMLElement;
    expect(header).toHaveClass('custom-padding');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    render(<CardHeader ref={(el) => { ref = el; }}>Header</CardHeader>);
    expect(ref).toBeInstanceOf(HTMLDivElement);
  });

  it('should render children', () => {
    render(<CardHeader>Header content</CardHeader>);
    expect(screen.getByText('Header content')).toBeInTheDocument();
  });
});

describe('CardTitle component', () => {
  it('should render CardTitle as h3', () => {
    const { container } = render(<CardTitle>Title</CardTitle>);
    const title = container.querySelector('h3');
    expect(title).toBeInTheDocument();
    expect(title).toHaveClass('text-2xl');
    expect(title).toHaveClass('font-semibold');
  });

  it('should accept custom className', () => {
    const { container } = render(<CardTitle className="custom-size">Title</CardTitle>);
    const title = container.firstChild as HTMLElement;
    expect(title).toHaveClass('custom-size');
  });

  it('should forward ref', () => {
    let ref: HTMLElement | null = null;
    render(<CardTitle ref={(el) => { ref = el; }}>Title</CardTitle>);
    expect(ref).toBeInstanceOf(HTMLElement);
    expect(ref?.tagName).toBe('H3');
  });

  it('should render children', () => {
    render(<CardTitle>My Card Title</CardTitle>);
    expect(screen.getByText('My Card Title')).toBeInTheDocument();
  });

  it('should have tracking-tight class', () => {
    const { container } = render(<CardTitle>Title</CardTitle>);
    const title = container.querySelector('h3');
    expect(title).toHaveClass('tracking-tight');
  });
});

describe('CardDescription component', () => {
  it('should render CardDescription as p', () => {
    const { container } = render(<CardDescription>Description</CardDescription>);
    const description = container.querySelector('p');
    expect(description).toBeInTheDocument();
    expect(description).toHaveClass('text-sm');
    expect(description).toHaveClass('text-muted-foreground');
  });

  it('should accept custom className', () => {
    const { container } = render(<CardDescription className="custom-text">Desc</CardDescription>);
    const description = container.firstChild as HTMLElement;
    expect(description).toHaveClass('custom-text');
  });

  it('should forward ref', () => {
    let ref: HTMLParagraphElement | null = null;
    render(<CardDescription ref={(el) => { ref = el; }}>Description</CardDescription>);
    expect(ref).toBeInstanceOf(HTMLParagraphElement);
  });

  it('should render children', () => {
    render(<CardDescription>Card description text</CardDescription>);
    expect(screen.getByText('Card description text')).toBeInTheDocument();
  });
});

describe('CardContent component', () => {
  it('should render CardContent', () => {
    const { container } = render(<CardContent>Content</CardContent>);
    const content = container.querySelector('div');
    expect(content).toHaveClass('p-6');
    expect(content).toHaveClass('pt-0');
  });

  it('should accept custom className', () => {
    const { container } = render(<CardContent className="custom-content">Content</CardContent>);
    const content = container.firstChild as HTMLElement;
    expect(content).toHaveClass('custom-content');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    render(<CardContent ref={(el) => { ref = el; }}>Content</CardContent>);
    expect(ref).toBeInstanceOf(HTMLDivElement);
  });

  it('should render children', () => {
    render(<CardContent>Main content</CardContent>);
    expect(screen.getByText('Main content')).toBeInTheDocument();
  });
});

describe('CardFooter component', () => {
  it('should render CardFooter', () => {
    const { container } = render(<CardFooter>Footer</CardFooter>);
    const footer = container.querySelector('div');
    expect(footer).toHaveClass('flex');
    expect(footer).toHaveClass('items-center');
    expect(footer).toHaveClass('p-6');
    expect(footer).toHaveClass('pt-0');
  });

  it('should accept custom className', () => {
    const { container } = render(<CardFooter className="custom-footer">Footer</CardFooter>);
    const footer = container.firstChild as HTMLElement;
    expect(footer).toHaveClass('custom-footer');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    render(<CardFooter ref={(el) => { ref = el; }}>Footer</CardFooter>);
    expect(ref).toBeInstanceOf(HTMLDivElement);
  });

  it('should render children', () => {
    render(<CardFooter>Footer content</CardFooter>);
    expect(screen.getByText('Footer content')).toBeInTheDocument();
  });
});

describe('Card composition', () => {
  it('should compose all card components together', () => {
    render(
      <Card>
        <CardHeader>
          <CardTitle>Card Title</CardTitle>
          <CardDescription>Card description goes here</CardDescription>
        </CardHeader>
        <CardContent>Main content area</CardContent>
        <CardFooter>Footer area</CardFooter>
      </Card>
    );

    expect(screen.getByText('Card Title')).toBeInTheDocument();
    expect(screen.getByText('Card description goes here')).toBeInTheDocument();
    expect(screen.getByText('Main content area')).toBeInTheDocument();
    expect(screen.getByText('Footer area')).toBeInTheDocument();
  });

  it('should work with multiple cards', () => {
    render(
      <>
        <Card>
          <CardHeader>
            <CardTitle>Card 1</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Card 2</CardTitle>
          </CardHeader>
        </Card>
      </>
    );

    expect(screen.getByText('Card 1')).toBeInTheDocument();
    expect(screen.getByText('Card 2')).toBeInTheDocument();
  });
});
