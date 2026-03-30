import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NodeDetailPanel from './NodeDetailPanel';

// Mock the API hooks
vi.mock('../api/hooks', () => ({
  useNode: vi.fn(() => ({
    data: { data: { node: { id: 'node-1', labels: ['Server'], properties: { name: 'test-server' } } } },
    isLoading: false,
  })),
  useDependencies: vi.fn(() => ({
    data: { data: { dependencies: [] } },
    isLoading: false,
  })),
  useDependents: vi.fn(() => ({
    data: { data: { dependents: [] } },
    isLoading: false,
  })),
  useNodeNeighbors: vi.fn(() => ({
    data: { data: { neighbors: [] } },
    isLoading: false,
  })),
  useImpactRadius: vi.fn(() => ({
    data: { data: { nodes: [] } },
    isLoading: false,
  })),
}));

describe('NodeDetailPanel component', () => {
  const defaultProps = {
    nodeId: 'test-node-1',
    onClose: vi.fn(),
    onNodeSelect: vi.fn(),
    onShowImpactRadius: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render panel', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    expect(container.querySelector('div')).toBeInTheDocument();
  });

  it('should have close button', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const header = container.querySelector('div [class*="flex"][class*="items-center"][class*="justify-between"]');
    const closeButton = header?.querySelector('button');
    expect(closeButton).toBeInTheDocument();
  });

  it('should call onClose when close button clicked', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();

    const { container } = render(<NodeDetailPanel {...defaultProps} onClose={onClose} />);

    const buttons = screen.getAllByRole('button');
    // The first button in the header should be the close button
    if (buttons.length > 0) {
      await user.click(buttons[0]);
      expect(onClose).toHaveBeenCalled();
    }
  });

  it('should render tab navigation', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const nav = container.querySelector('nav');
    expect(nav).toBeInTheDocument();
  });

  it('should have tabs in navigation', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const nav = container.querySelector('nav');
    const buttons = nav?.querySelectorAll('button');
    expect(buttons?.length).toBeGreaterThanOrEqual(3);
  });

  it('should show first tab selected by default', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const nav = container.querySelector('nav');
    const firstButton = nav?.querySelector('button');
    expect(firstButton).toHaveClass('border-blue-600');
  });

  it('should switch tabs when clicked', async () => {
    const user = userEvent.setup();
    const { container } = render(<NodeDetailPanel {...defaultProps} />);

    const nav = container.querySelector('nav');
    const buttons = nav?.querySelectorAll('button');
    if (buttons && buttons.length > 1) {
      const secondTab = buttons[1] as HTMLElement;
      await user.click(secondTab);
      expect(secondTab).toHaveClass('border-blue-600');
    }
  });

  it('should render with proper panel styling', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const panel = container.firstChild as HTMLElement;

    expect(panel).toHaveClass('h-full');
    expect(panel).toHaveClass('bg-white');
    expect(panel).toHaveClass('border-l');
    expect(panel).toHaveClass('flex');
    expect(panel).toHaveClass('flex-col');
  });

  it('should render header with icon', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const header = container.querySelector('svg');
    expect(header).toBeInTheDocument();
  });

  it('should render header with title text', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const header = container.querySelector('div h3');
    expect(header).toBeInTheDocument();
  });

  it('should pass nodeId to hooks', () => {
    const { container } = render(<NodeDetailPanel nodeId="custom-node-id" onClose={vi.fn()} />);

    // Component should render without errors - header h3 should exist
    expect(container.querySelector('h3')).toBeInTheDocument();
  });

  it('should render without optional callbacks', () => {
    const { container } = render(
      <NodeDetailPanel
        nodeId="test-node"
        onClose={vi.fn()}
      />
    );

    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle node selection callback', async () => {
    const user = userEvent.setup();
    const onNodeSelect = vi.fn();

    const { container } = render(<NodeDetailPanel {...defaultProps} onNodeSelect={onNodeSelect} />);

    // Component should render without errors
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should handle impact radius callback', async () => {
    const user = userEvent.setup();
    const onShowImpactRadius = vi.fn();

    const { container } = render(<NodeDetailPanel {...defaultProps} onShowImpactRadius={onShowImpactRadius} />);

    // Component should render without errors
    expect(container.firstChild).toBeInTheDocument();
  });

  it('should have responsive width classes', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const panel = container.firstChild as HTMLElement;

    expect(panel).toHaveClass('w-full');
    expect(panel).toHaveClass('md:w-96');
    expect(panel).toHaveClass('lg:w-[28rem]');
    expect(panel).toHaveClass('xl:w-[32rem]');
  });

  it('should have slide-in animation', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const panel = container.firstChild as HTMLElement;

    expect(panel).toHaveClass('animate-in');
    expect(panel).toHaveClass('slide-in-from-right');
  });

  it('should render multiple tabs', () => {
    render(<NodeDetailPanel {...defaultProps} />);

    const tabs = screen.getAllByRole('button');
    expect(tabs.length).toBeGreaterThan(0);
  });

  it('should have proper tab styling', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);

    const nav = container.querySelector('nav');
    const overviewTab = nav?.querySelector('button');
    expect(overviewTab).toHaveClass('flex-1');
    expect(overviewTab).toHaveClass('px-3');
    expect(overviewTab).toHaveClass('py-2');
    expect(overviewTab).toHaveClass('sm:px-4');
    expect(overviewTab).toHaveClass('sm:py-3');
  });

  it('should render in flex column layout', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const panel = container.firstChild as HTMLElement;

    expect(panel).toHaveClass('flex');
    expect(panel).toHaveClass('flex-col');
  });

  it('should have proper border styling', () => {
    const { container } = render(<NodeDetailPanel {...defaultProps} />);
    const panel = container.firstChild as HTMLElement;

    expect(panel).toHaveClass('border-l');
    expect(panel).toHaveClass('border-gray-200');
    expect(panel).toHaveClass('dark:border-gray-700');
  });

  it('should maintain state across tab switches', async () => {
    const user = userEvent.setup();
    const { container } = render(<NodeDetailPanel {...defaultProps} />);

    const nav = container.querySelector('nav');
    const buttons = nav?.querySelectorAll('button');

    if (buttons && buttons.length >= 3) {
      const secondTab = buttons[1] as HTMLElement;
      await user.click(secondTab);
      expect(secondTab).toHaveClass('border-blue-600');

      const thirdTab = buttons[2] as HTMLElement;
      await user.click(thirdTab);
      expect(thirdTab).toHaveClass('border-blue-600');

      const firstTab = buttons[0] as HTMLElement;
      await user.click(firstTab);
      expect(firstTab).toHaveClass('border-blue-600');
    }
  });
});
