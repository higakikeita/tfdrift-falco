import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Tabs, TabsList, TabsTrigger, TabsContent } from './tabs';

describe('Tabs component', () => {
  it('should render tabs with multiple triggers and content', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    expect(screen.getByText('Tab 1')).toBeInTheDocument();
    expect(screen.getByText('Tab 2')).toBeInTheDocument();
    expect(screen.getByText('Content 1')).toBeInTheDocument();
  });

  it('should show default tab content', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    expect(screen.getByText('Content 1')).toBeVisible();
  });

  it('should switch tabs on click', async () => {
    const user = userEvent.setup();
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    const tab2Button = screen.getByRole('tab', { name: /tab 2/i });
    await user.click(tab2Button);

    expect(screen.getByText('Content 2')).toBeVisible();
  });

  it('should set active state on selected tab', async () => {
    const user = userEvent.setup();
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    const tab2 = screen.getByRole('tab', { name: /tab 2/i });
    await user.click(tab2);

    expect(tab2).toHaveAttribute('data-state', 'active');
  });

  it('should handle keyboard navigation', async () => {
    const user = userEvent.setup();
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    const tab1 = screen.getByRole('tab', { name: /tab 1/i });
    tab1.focus();
    await user.keyboard('{ArrowRight}');

    const tab2 = screen.getByRole('tab', { name: /tab 2/i });
    expect(tab2).toHaveFocus();
  });
});

describe('TabsList component', () => {
  it('should render TabsList with correct styles', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    const list = container.querySelector('[role="tablist"]');
    expect(list).toHaveClass('inline-flex');
    expect(list).toHaveClass('h-10');
    expect(list).toHaveClass('bg-muted');
    expect(list).toHaveClass('rounded-md');
  });

  it('should accept custom className', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList className="custom-class">
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    const list = container.querySelector('[role="tablist"]');
    expect(list).toHaveClass('custom-class');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    render(
      <Tabs defaultValue="tab1">
        <TabsList ref={(el) => { ref = el; }}>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    expect(ref).toBeInstanceOf(HTMLDivElement);
  });
});

describe('TabsTrigger component', () => {
  it('should render TabsTrigger as button', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    const trigger = screen.getByRole('tab', { name: /tab 1/i });
    expect(trigger).toBeInTheDocument();
    expect(trigger.tagName).toBe('BUTTON');
  });

  it('should have focus-visible styles', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    const trigger = container.querySelector('[role="tab"]');
    expect(trigger).toHaveClass('focus-visible:ring-2');
  });

  it('should accept custom className', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1" className="custom-trigger">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    const trigger = container.querySelector('[role="tab"]');
    expect(trigger).toHaveClass('custom-trigger');
  });

  it('should forward ref', () => {
    let ref: HTMLButtonElement | null = null;
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger ref={(el) => { ref = el; }} value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>
    );

    expect(ref).toBeInstanceOf(HTMLButtonElement);
  });

  it('should be disabled when disabled prop is true', async () => {
    const user = userEvent.setup();
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2" disabled>Tab 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
      </Tabs>
    );

    const disabledTab = screen.getByRole('tab', { name: /tab 2/i });
    expect(disabledTab).toBeDisabled();
  });
});

describe('TabsContent component', () => {
  it('should render TabsContent', () => {
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content text</TabsContent>
      </Tabs>
    );

    expect(screen.getByText('Content text')).toBeInTheDocument();
  });

  it('should have focus-visible styles', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content</TabsContent>
      </Tabs>
    );

    const content = container.querySelector('[role="tabpanel"]');
    expect(content).toHaveClass('focus-visible:ring-2');
  });

  it('should accept custom className', () => {
    const { container } = render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1" className="custom-content">Content</TabsContent>
      </Tabs>
    );

    const content = container.querySelector('[role="tabpanel"]');
    expect(content).toHaveClass('custom-content');
  });

  it('should forward ref', () => {
    let ref: HTMLDivElement | null = null;
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
        <TabsContent ref={(el) => { ref = el; }} value="tab1">Content</TabsContent>
      </Tabs>
    );

    expect(ref).toBeInstanceOf(HTMLDivElement);
  });
});

describe('Tabs with many tabs', () => {
  it('should handle multiple tabs', async () => {
    const user = userEvent.setup();
    render(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          <TabsTrigger value="tab4">Tab 4</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
        <TabsContent value="tab2">Content 2</TabsContent>
        <TabsContent value="tab3">Content 3</TabsContent>
        <TabsContent value="tab4">Content 4</TabsContent>
      </Tabs>
    );

    await user.click(screen.getByRole('tab', { name: /tab 3/i }));
    expect(screen.getByText('Content 3')).toBeVisible();

    await user.click(screen.getByRole('tab', { name: /tab 4/i }));
    expect(screen.getByText('Content 4')).toBeVisible();
  });
});
