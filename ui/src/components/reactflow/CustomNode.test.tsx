/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Custom Node Component Tests
 * Tests for React Flow custom node rendering and interactions
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders, userEvent } from '@/__tests__/utils/testUtils';
import { CustomNode } from './CustomNode';
import type { NodeProps } from 'reactflow';

// Mock child components
vi.mock('../icons/OfficialCloudIcons', () => ({
  OfficialCloudIcon: ({ type, size }: { type: string; size: number }) => (
    <div data-testid="cloud-icon" data-type={type} data-size={size}>
      {type}
    </div>
  ),
}));

vi.mock('../graph/NodeTooltip', () => ({
  NodeTooltip: ({ data, position }: any) => (
    <div data-testid="node-tooltip" data-position={JSON.stringify(position)}>
      <div>{data.label}</div>
      <div>{data.type}</div>
      <div>{data.severity}</div>
    </div>
  ),
}));

vi.mock('../graph/NodeContextMenu', () => ({
  NodeContextMenu: ({
    nodeId,
    onClose,
    onViewDetails,
    onFocusView,
    onShowDependencies,
    onShowImpact,
    onCopyId,
  }: any) => (
    <div data-testid="context-menu" data-node-id={nodeId}>
      <button onClick={onViewDetails}>View Details</button>
      <button onClick={onFocusView}>Focus View</button>
      <button onClick={onShowDependencies}>Show Dependencies</button>
      <button onClick={onShowImpact}>Show Impact</button>
      <button onClick={onCopyId}>Copy ID</button>
      <button onClick={onClose}>Close</button>
    </div>
  ),
}));

describe('CustomNode', () => {
  let eventListener: any;

  beforeEach(() => {
    // Mock clipboard API
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
      writable: true,
      configurable: true,
    });

    // Capture custom events
    eventListener = vi.fn();
    window.addEventListener('node-detail', eventListener);
    window.addEventListener('node-focus', eventListener);
    window.addEventListener('node-dependencies', eventListener);
    window.addEventListener('node-impact', eventListener);
  });

  afterEach(() => {
    window.removeEventListener('node-detail', eventListener);
    window.removeEventListener('node-focus', eventListener);
    window.removeEventListener('node-dependencies', eventListener);
    window.removeEventListener('node-impact', eventListener);
    vi.clearAllMocks();
  });

  const createNodeProps = (overrides?: Partial<NodeProps>): NodeProps => ({
    id: 'test-node-1',
    type: 'custom',
    data: {
      label: 'Test IAM Role',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      severity: 'high',
      resource_name: 'test-role',
      metadata: { arn: 'arn:aws:iam::123456789012:role/test' },
    },
    selected: false,
    isConnectable: true,
    xPos: 0,
    yPos: 0,
    dragging: false,
    zIndex: 0,
    ...overrides,
  });

  describe('Rendering', () => {
    it('should render node with label', () => {
      const props = createNodeProps();
      renderWithProviders(<CustomNode {...props} />);

      expect(screen.getByText('Test IAM Role')).toBeInTheDocument();
    });

    it('should render cloud icon with correct type', () => {
      const props = createNodeProps();
      renderWithProviders(<CustomNode {...props} />);

      const icon = screen.getByTestId('cloud-icon');
      expect(icon).toHaveAttribute('data-type', 'aws_iam_role');
      expect(icon).toHaveAttribute('data-size', '80');
    });

    it('should display resource name when provided', () => {
      const props = createNodeProps();
      renderWithProviders(<CustomNode {...props} />);

      expect(screen.getByText('test-role')).toBeInTheDocument();
    });

    it('should not display resource name when not provided', () => {
      const props = createNodeProps({
        data: {
          label: 'Test Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
        },
      });
      renderWithProviders(<CustomNode {...props} />);

      expect(screen.queryByText('test-role')).not.toBeInTheDocument();
    });
  });

  describe('Severity Styling', () => {
    it('should apply critical severity styling', () => {
      const props = createNodeProps({
        data: {
          label: 'Critical Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
          severity: 'critical',
        },
      });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.border-red-600');
      expect(node).toBeInTheDocument();
      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    });

    it('should apply high severity styling', () => {
      const props = createNodeProps({
        data: {
          label: 'High Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
          severity: 'high',
        },
      });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.border-orange-600');
      expect(node).toBeInTheDocument();
      expect(screen.getByText('HIGH')).toBeInTheDocument();
    });

    it('should apply medium severity styling', () => {
      const props = createNodeProps({
        data: {
          label: 'Medium Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
          severity: 'medium',
        },
      });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.border-yellow-600');
      expect(node).toBeInTheDocument();
      expect(screen.getByText('MEDIUM')).toBeInTheDocument();
    });

    it('should apply low severity styling', () => {
      const props = createNodeProps({
        data: {
          label: 'Low Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
          severity: 'low',
        },
      });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.border-blue-600');
      expect(node).toBeInTheDocument();
      expect(screen.getByText('LOW')).toBeInTheDocument();
    });

    it('should apply default styling when severity is not provided', () => {
      const props = createNodeProps({
        data: {
          label: 'Default Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
        },
      });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.border-gray-300');
      expect(node).toBeInTheDocument();
    });

    it('should not render severity badge when severity is not provided', () => {
      const props = createNodeProps({
        data: {
          label: 'No Severity Node',
          type: 'aws_iam_role',
          resource_type: 'aws_iam_role',
        },
      });
      renderWithProviders(<CustomNode {...props} />);

      expect(screen.queryByText(/CRITICAL|HIGH|MEDIUM|LOW/)).not.toBeInTheDocument();
    });
  });

  describe('Selected State', () => {
    it('should apply selected styling when node is selected', () => {
      const props = createNodeProps({ selected: true });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.ring-4.ring-blue-500');
      expect(node).toBeInTheDocument();
    });

    it('should not apply selected styling when node is not selected', () => {
      const props = createNodeProps({ selected: false });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.ring-4.ring-blue-500');
      expect(node).not.toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should dispatch node-detail event on click', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.click(node);

      await waitFor(() => {
        expect(eventListener).toHaveBeenCalled();
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-detail'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should dispatch node-focus event on double click', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.dblClick(node);

      await waitFor(() => {
        expect(eventListener).toHaveBeenCalled();
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-focus'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should show tooltip on mouse enter', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;

      // Mock getBoundingClientRect
      vi.spyOn(node, 'getBoundingClientRect').mockReturnValue({
        left: 100,
        top: 50,
        width: 240,
        height: 180,
        right: 340,
        bottom: 230,
        x: 100,
        y: 50,
        toJSON: () => {},
      });

      await user.hover(node);

      await waitFor(() => {
        expect(screen.getByTestId('node-tooltip')).toBeInTheDocument();
      });
    });

    it('should hide tooltip on mouse leave', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;

      // Mock getBoundingClientRect
      vi.spyOn(node, 'getBoundingClientRect').mockReturnValue({
        left: 100,
        top: 50,
        width: 240,
        height: 180,
        right: 340,
        bottom: 230,
        x: 100,
        y: 50,
        toJSON: () => {},
      });

      await user.hover(node);
      await waitFor(() => expect(screen.getByTestId('node-tooltip')).toBeInTheDocument());

      await user.unhover(node);
      await waitFor(() => expect(screen.queryByTestId('node-tooltip')).not.toBeInTheDocument());
    });

    it('should show context menu on right click', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      await waitFor(() => {
        expect(screen.getByTestId('context-menu')).toBeInTheDocument();
      });
    });

    it('should hide tooltip when context menu is shown', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;

      // Mock getBoundingClientRect
      vi.spyOn(node, 'getBoundingClientRect').mockReturnValue({
        left: 100,
        top: 50,
        width: 240,
        height: 180,
        right: 340,
        bottom: 230,
        x: 100,
        y: 50,
        toJSON: () => {},
      });

      // Show tooltip first
      await user.hover(node);
      await waitFor(() => expect(screen.getByTestId('node-tooltip')).toBeInTheDocument());

      // Right click to show context menu
      await user.pointer({ keys: '[MouseRight>]', target: node });

      await waitFor(() => {
        expect(screen.queryByTestId('node-tooltip')).not.toBeInTheDocument();
        expect(screen.getByTestId('context-menu')).toBeInTheDocument();
      });
    });
  });

  describe('Context Menu Actions', () => {
    it('should dispatch node-detail event when "View Details" is clicked', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      const viewDetailsButton = await screen.findByText('View Details');
      await user.click(viewDetailsButton);

      await waitFor(() => {
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-detail'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should dispatch node-focus event when "Focus View" is clicked', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      const focusButton = await screen.findByText('Focus View');
      await user.click(focusButton);

      await waitFor(() => {
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-focus'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should dispatch node-dependencies event when "Show Dependencies" is clicked', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      const depsButton = await screen.findByText('Show Dependencies');
      await user.click(depsButton);

      await waitFor(() => {
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-dependencies'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should dispatch node-impact event when "Show Impact" is clicked', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      const impactButton = await screen.findByText('Show Impact');
      await user.click(impactButton);

      await waitFor(() => {
        const event = eventListener.mock.calls.find((call: any) =>
          call[0].type === 'node-impact'
        );
        expect(event).toBeTruthy();
      });
    });

    it('should copy node ID to clipboard when "Copy ID" is clicked', async () => {
      const user = userEvent.setup();
      const writeTextSpy = vi.fn().mockResolvedValue(undefined);

      // Mock clipboard with spy
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: writeTextSpy,
        },
        writable: true,
        configurable: true,
      });

      const props = createNodeProps({ id: 'test-node-123' });
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      const copyButton = await screen.findByText('Copy ID');
      await user.click(copyButton);

      await waitFor(() => {
        expect(writeTextSpy).toHaveBeenCalledWith('test-node-123');
      });
    });

    it('should close context menu when "Close" is clicked', async () => {
      const user = userEvent.setup();
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      await user.pointer({ keys: '[MouseRight>]', target: node });

      await waitFor(() => expect(screen.getByTestId('context-menu')).toBeInTheDocument());

      const closeButton = screen.getByText('Close');
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByTestId('context-menu')).not.toBeInTheDocument();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper cursor pointer styling', () => {
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.cursor-pointer');
      expect(node).toBeInTheDocument();
    });

    it('should prevent default context menu behavior', async () => {
      const props = createNodeProps();
      const { container } = renderWithProviders(<CustomNode {...props} />);

      const node = container.querySelector('.rounded-2xl')!;
      const contextMenuEvent = new MouseEvent('contextmenu', {
        bubbles: true,
        cancelable: true,
      });

      const preventDefaultSpy = vi.spyOn(contextMenuEvent, 'preventDefault');
      node.dispatchEvent(contextMenuEvent);

      expect(preventDefaultSpy).toHaveBeenCalled();
    });
  });
});
