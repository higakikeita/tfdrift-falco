import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  ExternalLink: () => <div data-testid="external-link-icon" />,
  Eye: () => <div data-testid="eye-icon" />,
  GitBranch: () => <div data-testid="git-branch-icon" />,
  Target: () => <div data-testid="target-icon" />,
  Copy: () => <div data-testid="copy-icon" />,
  Info: () => <div data-testid="info-icon" />,
}));

import { NodeContextMenu } from './NodeContextMenu';

describe('NodeContextMenu', () => {
  const mockPosition = { x: 100, y: 100 };
  const mockNodeData = {
    label: 'test-instance',
    type: 'node',
    resource_type: 'aws_instance',
  };
  const mockOnClose = vi.fn();
  const mockOnViewDetails = vi.fn();
  const mockOnFocusView = vi.fn();
  const mockOnShowDependencies = vi.fn();
  const mockOnShowImpact = vi.fn();
  const mockOnCopyId = vi.fn();

  it('should render without crashing', () => {
    const { container } = render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
      />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should display node header information', () => {
    render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
      />
    );
    expect(screen.getByText('test-instance')).toBeInTheDocument();
    expect(screen.getByText('aws_instance')).toBeInTheDocument();
  });

  it('should render menu items when callbacks are provided', () => {
    render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
        onViewDetails={mockOnViewDetails}
        onFocusView={mockOnFocusView}
        onShowDependencies={mockOnShowDependencies}
        onShowImpact={mockOnShowImpact}
        onCopyId={mockOnCopyId}
      />
    );
    expect(screen.getByText('詳細を表示')).toBeInTheDocument();
    expect(screen.getByText('フォーカスビュー')).toBeInTheDocument();
    expect(screen.getByText('依存関係を表示')).toBeInTheDocument();
    expect(screen.getByText('影響範囲を表示')).toBeInTheDocument();
    expect(screen.getByText('IDをコピー')).toBeInTheDocument();
  });

  it('should call callback when menu item is clicked', () => {
    render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
        onViewDetails={mockOnViewDetails}
      />
    );
    const detailsButton = screen.getByText('詳細を表示');
    fireEvent.click(detailsButton);
    expect(mockOnViewDetails).toHaveBeenCalled();
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should render external link menu item', () => {
    render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
      />
    );
    expect(screen.getByText('新しいタブで開く')).toBeInTheDocument();
  });

  it('should close menu when external link is clicked', () => {
    render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
      />
    );
    const linkButton = screen.getByText('新しいタブで開く');
    fireEvent.click(linkButton);
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should position menu at correct coordinates', () => {
    const { container } = render(
      <NodeContextMenu
        position={mockPosition}
        nodeId="node-1"
        nodeData={mockNodeData}
        onClose={mockOnClose}
      />
    );
    const menu = container.querySelector('div[style*="100px"]');
    expect(menu).toBeTruthy();
  });
});
