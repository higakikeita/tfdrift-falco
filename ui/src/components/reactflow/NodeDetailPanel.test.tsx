/**
 * NodeDetailPanel Component Tests
 * Tests for node detail panel display and interactions
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders, userEvent } from '@/__tests__/utils/testUtils';
import { NodeDetailPanel } from './NodeDetailPanel';

// Mock OfficialCloudIcon component
vi.mock('../icons/OfficialCloudIcons', () => ({
  OfficialCloudIcon: ({ type, size }: { type: string; size: number }) => (
    <div data-testid="cloud-icon" data-type={type} data-size={size}>
      {type}
    </div>
  ),
}));

describe('NodeDetailPanel', () => {
  const mockOnClose = vi.fn();

  const mockNode = {
    id: 'node-123',
    data: {
      label: 'Test IAM Role',
      type: 'aws_iam_role',
      resource_type: 'aws_iam_role',
      severity: 'high',
      resource_name: 'test-role-name',
      metadata: {
        arn: 'arn:aws:iam::123456789012:role/test-role',
        created_date: '2024-01-01',
        tags: { Environment: 'production' },
      },
    },
  };

  beforeEach(() => {
    mockOnClose.mockClear();
  });

  describe('Rendering', () => {
    it('should render nothing when node is null', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={null} onClose={mockOnClose} />
      );

      expect(container.firstChild).toBeNull();
    });

    it('should render node detail panel when node is provided', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Test IAM Role')).toBeInTheDocument();
      expect(screen.getByText('Resource Details')).toBeInTheDocument();
    });

    it('should render cloud icon with correct type', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      const icon = screen.getByTestId('cloud-icon');
      expect(icon).toHaveAttribute('data-type', 'aws_iam_role');
      expect(icon).toHaveAttribute('data-size', '48');
    });

    it('should display node label', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Test IAM Role')).toBeInTheDocument();
    });

    it('should have slide-in animation class', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const panel = container.querySelector('.animate-slide-in');
      expect(panel).toBeInTheDocument();
    });
  });

  describe('Severity Display', () => {
    it('should display critical severity badge', () => {
      const criticalNode = {
        ...mockNode,
        data: { ...mockNode.data, severity: 'critical' },
      };

      renderWithProviders(<NodeDetailPanel node={criticalNode} onClose={mockOnClose} />);

      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
      const badge = screen.getByText('CRITICAL');
      expect(badge).toHaveClass('bg-red-100', 'text-red-800');
    });

    it('should display high severity badge', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('HIGH')).toBeInTheDocument();
      const badge = screen.getByText('HIGH');
      expect(badge).toHaveClass('bg-orange-100', 'text-orange-800');
    });

    it('should display medium severity badge', () => {
      const mediumNode = {
        ...mockNode,
        data: { ...mockNode.data, severity: 'medium' },
      };

      renderWithProviders(<NodeDetailPanel node={mediumNode} onClose={mockOnClose} />);

      expect(screen.getByText('MEDIUM')).toBeInTheDocument();
      const badge = screen.getByText('MEDIUM');
      expect(badge).toHaveClass('bg-yellow-100', 'text-yellow-800');
    });

    it('should display low severity badge', () => {
      const lowNode = {
        ...mockNode,
        data: { ...mockNode.data, severity: 'low' },
      };

      renderWithProviders(<NodeDetailPanel node={lowNode} onClose={mockOnClose} />);

      expect(screen.getByText('LOW')).toBeInTheDocument();
      const badge = screen.getByText('LOW');
      expect(badge).toHaveClass('bg-blue-100', 'text-blue-800');
    });

    it('should not display severity section when severity is undefined', () => {
      const noSeverityNode = {
        ...mockNode,
        data: { ...mockNode.data, severity: undefined },
      };

      renderWithProviders(<NodeDetailPanel node={noSeverityNode} onClose={mockOnClose} />);

      expect(screen.queryByText('Severity')).not.toBeInTheDocument();
    });
  });

  describe('Resource Information', () => {
    it('should display resource type', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Resource Type')).toBeInTheDocument();
      // Use getAllByText since the type appears in both icon and code element
      const resourceTypes = screen.getAllByText('aws_iam_role');
      expect(resourceTypes.length).toBeGreaterThan(0);
    });

    it('should display resource name when provided', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Resource Name')).toBeInTheDocument();
      expect(screen.getByText('test-role-name')).toBeInTheDocument();
    });

    it('should not display resource name section when not provided', () => {
      const noNameNode = {
        ...mockNode,
        data: { ...mockNode.data, resource_name: undefined },
      };

      renderWithProviders(<NodeDetailPanel node={noNameNode} onClose={mockOnClose} />);

      expect(screen.queryByText('Resource Name')).not.toBeInTheDocument();
    });

    it('should display node ID', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Node ID')).toBeInTheDocument();
      expect(screen.getByText('node-123')).toBeInTheDocument();
    });
  });

  describe('Metadata Display', () => {
    it('should display metadata section when metadata exists', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('Metadata')).toBeInTheDocument();
    });

    it('should display string metadata values', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('arn')).toBeInTheDocument();
      expect(screen.getByText('arn:aws:iam::123456789012:role/test-role')).toBeInTheDocument();
      expect(screen.getByText('created_date')).toBeInTheDocument();
      expect(screen.getByText('2024-01-01')).toBeInTheDocument();
    });

    it('should display object metadata as JSON', () => {
      const { container } = renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      expect(screen.getByText('tags')).toBeInTheDocument();
      // JSON is displayed in a pre element
      const preElements = container.querySelectorAll('pre');
      expect(preElements.length).toBeGreaterThan(0);
      // Check that JSON content contains expected keys
      const jsonContent = preElements[0].textContent || '';
      expect(jsonContent).toContain('Environment');
      expect(jsonContent).toContain('production');
    });

    it('should not display metadata section when metadata is empty', () => {
      const noMetadataNode = {
        ...mockNode,
        data: { ...mockNode.data, metadata: {} },
      };

      renderWithProviders(<NodeDetailPanel node={noMetadataNode} onClose={mockOnClose} />);

      expect(screen.queryByText('Metadata')).not.toBeInTheDocument();
    });

    it('should not display metadata section when metadata is undefined', () => {
      const noMetadataNode = {
        ...mockNode,
        data: { ...mockNode.data, metadata: undefined },
      };

      renderWithProviders(<NodeDetailPanel node={noMetadataNode} onClose={mockOnClose} />);

      expect(screen.queryByText('Metadata')).not.toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should call onClose when header close button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      const headerCloseButton = screen.getByLabelText('Close');
      await user.click(headerCloseButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when footer close button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      // There are two close buttons (header with aria-label and footer with text)
      const closeButtons = screen.getAllByRole('button', { name: /close/i });
      const footerCloseButton = closeButtons[closeButtons.length - 1]; // Get the last one (footer)
      await user.click(footerCloseButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Layout and Styling', () => {
    it('should have correct positioning classes', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const panel = container.querySelector('.absolute.right-6.top-24');
      expect(panel).toBeInTheDocument();
    });

    it('should have correct width class', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const panel = container.querySelector('.w-96');
      expect(panel).toBeInTheDocument();
    });

    it('should have gradient header', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const header = container.querySelector('.bg-gradient-to-r.from-blue-600.to-blue-700');
      expect(header).toBeInTheDocument();
    });

    it('should have scrollable content area', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const content = container.querySelector('.overflow-y-auto');
      expect(content).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have aria-label on close button', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      const closeButton = screen.getByLabelText('Close');
      expect(closeButton).toBeInTheDocument();
    });

    it('should have proper semantic structure with headings', () => {
      renderWithProviders(<NodeDetailPanel node={mockNode} onClose={mockOnClose} />);

      const heading = screen.getByRole('heading', { name: 'Test IAM Role' });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle node with minimal data', () => {
      const minimalNode = {
        id: 'minimal-node',
        data: {
          label: 'Minimal',
          type: 'test',
          resource_type: 'test_resource',
        },
      };

      renderWithProviders(<NodeDetailPanel node={minimalNode} onClose={mockOnClose} />);

      expect(screen.getByText('Minimal')).toBeInTheDocument();
      // test_resource appears in both icon and code element
      const resourceTypes = screen.getAllByText('test_resource');
      expect(resourceTypes.length).toBeGreaterThan(0);
    });

    it('should handle very long metadata values', () => {
      const longValueNode = {
        ...mockNode,
        data: {
          ...mockNode.data,
          metadata: {
            long_value: 'a'.repeat(1000),
          },
        },
      };

      renderWithProviders(<NodeDetailPanel node={longValueNode} onClose={mockOnClose} />);

      expect(screen.getByText('long_value')).toBeInTheDocument();
    });

    it('should handle special characters in metadata', () => {
      const specialCharsNode = {
        ...mockNode,
        data: {
          ...mockNode.data,
          metadata: {
            special: '<script>alert("xss")</script>',
          },
        },
      };

      renderWithProviders(<NodeDetailPanel node={specialCharsNode} onClose={mockOnClose} />);

      // Should display the text safely without executing
      expect(screen.getByText('special')).toBeInTheDocument();
    });
  });
});
