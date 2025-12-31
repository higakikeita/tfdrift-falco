/**
 * HelpOverlay Component Tests
 * Tests for contextual help overlay functionality
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders, userEvent } from '@/__tests__/utils/testUtils';
import { HelpOverlay } from './HelpOverlay';

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  HelpCircle: () => <div data-testid="help-circle-icon">HelpCircle</div>,
  X: () => <div data-testid="x-icon">X</div>,
  ChevronDown: () => <div data-testid="chevron-down-icon">ChevronDown</div>,
  ChevronUp: () => <div data-testid="chevron-up-icon">ChevronUp</div>,
  Lightbulb: () => <div data-testid="lightbulb-icon">Lightbulb</div>,
  Zap: () => <div data-testid="zap-icon">Zap</div>,
  Target: () => <div data-testid="target-icon">Target</div>,
}));

describe('HelpOverlay', () => {
  const mockOnOpenShortcuts = vi.fn();
  const mockOnOpenWelcome = vi.fn();

  beforeEach(() => {
    mockOnOpenShortcuts.mockClear();
    mockOnOpenWelcome.mockClear();
  });

  describe('Rendering - Expanded State', () => {
    it('should render help overlay in expanded state by default', () => {
      renderWithProviders(
        <HelpOverlay
          onOpenShortcuts={mockOnOpenShortcuts}
          onOpenWelcome={mockOnOpenWelcome}
        />
      );

      expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();
    });

    it('should display lightbulb icon in header', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByTestId('lightbulb-icon')).toBeInTheDocument();
    });

    it('should display quick tips section', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('クイックヒント')).toBeInTheDocument();
    });

    it('should display key features section', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('主な機能')).toBeInTheDocument();
    });
  });

  describe('Quick Tips', () => {
    it('should display all quick tips', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('ノードをクリックで詳細を表示')).toBeInTheDocument();
      expect(screen.getByText('ダブルクリックでフォーカスビュー')).toBeInTheDocument();
      expect(screen.getByText('右クリックで依存関係を表示')).toBeInTheDocument();
      expect(screen.getByText('マウスホイールでズーム操作')).toBeInTheDocument();
      expect(screen.getByText('ドラッグでグラフを移動')).toBeInTheDocument();
    });
  });

  describe('Key Features', () => {
    it('should display impact analysis feature', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('影響範囲分析')).toBeInTheDocument();
      expect(screen.getByText('詳細パネルの「影響範囲」タブで確認')).toBeInTheDocument();
    });

    it('should display dependency tracking feature', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('依存関係追跡')).toBeInTheDocument();
      expect(screen.getByText('「関係性」タブで依存先・依存元を表示')).toBeInTheDocument();
    });

    it('should display search filter feature', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('検索・フィルター')).toBeInTheDocument();
      expect(screen.getByText('左サイドバーで深刻度・タイプで絞り込み')).toBeInTheDocument();
    });
  });

  describe('Action Buttons', () => {
    it('should display keyboard shortcuts button when onOpenShortcuts is provided', () => {
      renderWithProviders(
        <HelpOverlay
          onOpenShortcuts={mockOnOpenShortcuts}
          onOpenWelcome={mockOnOpenWelcome}
        />
      );

      expect(screen.getByText('⌨️ キーボードショートカット')).toBeInTheDocument();
    });

    it('should display tutorial button when onOpenWelcome is provided', () => {
      renderWithProviders(
        <HelpOverlay
          onOpenShortcuts={mockOnOpenShortcuts}
          onOpenWelcome={mockOnOpenWelcome}
        />
      );

      expect(screen.getByText('🎯 チュートリアルを再表示')).toBeInTheDocument();
    });

    it('should not display shortcuts button when onOpenShortcuts is not provided', () => {
      renderWithProviders(<HelpOverlay onOpenWelcome={mockOnOpenWelcome} />);

      expect(screen.queryByText('⌨️ キーボードショートカット')).not.toBeInTheDocument();
    });

    it('should not display tutorial button when onOpenWelcome is not provided', () => {
      renderWithProviders(<HelpOverlay onOpenShortcuts={mockOnOpenShortcuts} />);

      expect(screen.queryByText('🎯 チュートリアルを再表示')).not.toBeInTheDocument();
    });

    it('should call onOpenShortcuts when shortcuts button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(
        <HelpOverlay
          onOpenShortcuts={mockOnOpenShortcuts}
          onOpenWelcome={mockOnOpenWelcome}
        />
      );

      const button = screen.getByText('⌨️ キーボードショートカット');
      await user.click(button);

      expect(mockOnOpenShortcuts).toHaveBeenCalledTimes(1);
    });

    it('should call onOpenWelcome when tutorial button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(
        <HelpOverlay
          onOpenShortcuts={mockOnOpenShortcuts}
          onOpenWelcome={mockOnOpenWelcome}
        />
      );

      const button = screen.getByText('🎯 チュートリアルを再表示');
      await user.click(button);

      expect(mockOnOpenWelcome).toHaveBeenCalledTimes(1);
    });
  });

  describe('Expand/Collapse Functionality', () => {
    it('should collapse content when collapse button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // Initially expanded - quick tips should be visible
      expect(screen.getByText('クイックヒント')).toBeInTheDocument();

      // Click collapse button
      const collapseButton = screen.getByLabelText('折りたたむ');
      await user.click(collapseButton);

      // Content should be hidden
      await waitFor(() => {
        expect(screen.queryByText('クイックヒント')).not.toBeInTheDocument();
      });
    });

    it('should expand content when expand button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // Collapse first
      const collapseButton = screen.getByLabelText('折りたたむ');
      await user.click(collapseButton);

      await waitFor(() => {
        expect(screen.queryByText('クイックヒント')).not.toBeInTheDocument();
      });

      // Then expand
      const expandButton = screen.getByLabelText('展開する');
      await user.click(expandButton);

      await waitFor(() => {
        expect(screen.getByText('クイックヒント')).toBeInTheDocument();
      });
    });
  });

  describe('Hide/Show Functionality', () => {
    it('should hide overlay and show floating button when close button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // Initially visible
      expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();

      // Click close button
      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      // Overlay should be hidden, floating button should appear
      await waitFor(() => {
        expect(screen.queryByText('クイックヘルプ')).not.toBeInTheDocument();
        expect(screen.getByLabelText('ヘルプを表示')).toBeInTheDocument();
      });
    });

    it('should show overlay when floating help button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // Hide overlay first
      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('クイックヘルプ')).not.toBeInTheDocument();
      });

      // Show overlay again
      const showButton = screen.getByLabelText('ヘルプを表示');
      await user.click(showButton);

      await waitFor(() => {
        expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();
      });
    });

    it('should display HelpCircle icon in floating button', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // Hide overlay
      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.getByTestId('help-circle-icon')).toBeInTheDocument();
      });
    });
  });

  describe('Layout and Styling', () => {
    it('should have fixed positioning at bottom right', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const overlay = container.querySelector('.fixed.bottom-6.right-6');
      expect(overlay).toBeInTheDocument();
    });

    it('should have gradient header', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const header = container.querySelector('.bg-gradient-to-r.from-blue-600.to-indigo-600');
      expect(header).toBeInTheDocument();
    });

    it('should have rounded corners', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const overlay = container.querySelector('.rounded-xl');
      expect(overlay).toBeInTheDocument();
    });

    it('should have shadow and border', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const overlay = container.querySelector('.shadow-2xl.border');
      expect(overlay).toBeInTheDocument();
    });

    it('should have animation classes', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const overlay = container.querySelector('.animate-in.slide-in-from-bottom');
      expect(overlay).toBeInTheDocument();
    });

    it('should have max width constraint', () => {
      const { container } = renderWithProviders(<HelpOverlay />);

      const overlay = container.querySelector('.max-w-sm');
      expect(overlay).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have aria-labels on control buttons', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByLabelText('折りたたむ')).toBeInTheDocument();
      expect(screen.getByLabelText('閉じる')).toBeInTheDocument();
    });

    it('should have proper heading hierarchy', () => {
      renderWithProviders(<HelpOverlay />);

      const heading = screen.getByRole('heading', { name: 'クイックヘルプ', level: 3 });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should render without any props', () => {
      renderWithProviders(<HelpOverlay />);

      expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();
    });

    it('should handle rapid expand/collapse', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      const collapseButton = screen.getByLabelText('折りたたむ');

      // Rapid clicks
      await user.click(collapseButton);
      await user.click(screen.getByLabelText('展開する'));
      await user.click(screen.getByLabelText('折りたたむ'));

      // Should end in collapsed state
      await waitFor(() => {
        expect(screen.queryByText('クイックヒント')).not.toBeInTheDocument();
      });
    });

    it('should handle rapid hide/show', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      const closeButton = screen.getByLabelText('閉じる');

      // Hide
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.getByLabelText('ヘルプを表示')).toBeInTheDocument();
      });

      // Show again
      await user.click(screen.getByLabelText('ヘルプを表示'));

      await waitFor(() => {
        expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();
      });
    });
  });
});
