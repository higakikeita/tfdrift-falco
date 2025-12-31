/**
 * KeyboardShortcutsGuide Component Tests
 * Tests for keyboard shortcuts guide modal
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders, userEvent } from '@/__tests__/utils/testUtils';
import { KeyboardShortcutsGuide } from './KeyboardShortcutsGuide';

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  X: () => <div data-testid="x-icon">X</div>,
  Keyboard: () => <div data-testid="keyboard-icon">Keyboard</div>,
}));

describe('KeyboardShortcutsGuide', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    mockOnClose.mockClear();
  });

  describe('Rendering', () => {
    it('should render keyboard shortcuts guide modal', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
    });

    it('should render keyboard icon', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByTestId('keyboard-icon')).toBeInTheDocument();
    });

    it('should render all categories', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('ナビゲーション')).toBeInTheDocument();
      expect(screen.getByText('選択・操作')).toBeInTheDocument();
      expect(screen.getByText('表示・エクスポート')).toBeInTheDocument();
      expect(screen.getByText('ヘルプ')).toBeInTheDocument();
    });
  });

  describe('Shortcuts Display', () => {
    it('should display navigation shortcuts', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('グラフ全体を画面にフィット')).toBeInTheDocument();
      expect(screen.getByText('F')).toBeInTheDocument();

      expect(screen.getByText('グラフを中央に配置')).toBeInTheDocument();
      expect(screen.getByText('C')).toBeInTheDocument();

      expect(screen.getByText('ズームイン')).toBeInTheDocument();
      expect(screen.getByText('+')).toBeInTheDocument();

      expect(screen.getByText('ズームアウト')).toBeInTheDocument();
      expect(screen.getByText('-')).toBeInTheDocument();
    });

    it('should display selection shortcuts', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('ノード詳細パネルを開く')).toBeInTheDocument();
      expect(screen.getByText('Click')).toBeInTheDocument();

      expect(screen.getByText('フォーカスビューでハイライト')).toBeInTheDocument();
      expect(screen.getByText('Double Click')).toBeInTheDocument();

      expect(screen.getByText('コンテキストメニューを表示')).toBeInTheDocument();
      expect(screen.getByText('Right Click')).toBeInTheDocument();

      expect(screen.getByText('詳細パネルを閉じる')).toBeInTheDocument();
      expect(screen.getByText('ESC')).toBeInTheDocument();
    });

    it('should display export shortcuts', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('グラフをPNG形式で保存')).toBeInTheDocument();
      expect(screen.getByText('Ctrl/Cmd + S')).toBeInTheDocument();

      expect(screen.getByText('グラフをSVG形式でエクスポート')).toBeInTheDocument();
      expect(screen.getByText('Ctrl/Cmd + E')).toBeInTheDocument();
    });

    it('should display help shortcuts', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText('このヘルプを表示')).toBeInTheDocument();
      // Use getAllByText since "?" appears in both shortcuts list and footer
      const questionMarks = screen.getAllByText('?');
      expect(questionMarks.length).toBeGreaterThan(0);

      expect(screen.getByText('クイックヘルプを表示/非表示')).toBeInTheDocument();
      expect(screen.getByText('H')).toBeInTheDocument();
    });

    it('should display shortcuts in kbd elements', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const kbdElements = container.querySelectorAll('kbd');
      expect(kbdElements.length).toBeGreaterThan(0);
    });
  });

  describe('Close Functionality', () => {
    it('should call onClose when header close button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when footer close button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const closeButtons = screen.getAllByRole('button', { name: '閉じる' });
      const footerCloseButton = closeButtons[closeButtons.length - 1];
      await user.click(footerCloseButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Layout and Styling', () => {
    it('should have backdrop with blur', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const backdrop = container.querySelector('.backdrop-blur-sm');
      expect(backdrop).toBeInTheDocument();
    });

    it('should have gradient header', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const header = container.querySelector('.bg-gradient-to-r.from-indigo-600.to-purple-600');
      expect(header).toBeInTheDocument();
    });

    it('should have scrollable content area', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const content = container.querySelector('.overflow-y-auto');
      expect(content).toBeInTheDocument();
    });

    it('should have max height constraint', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const modal = container.querySelector('.max-h-\\[80vh\\]');
      expect(modal).toBeInTheDocument();
    });

    it('should have animation classes', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const modal = container.querySelector('.animate-in.zoom-in-95');
      expect(modal).toBeInTheDocument();
    });
  });

  describe('Category Organization', () => {
    it('should group shortcuts by category', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      // Check that categories are properly structured
      const categories = container.querySelectorAll('h3.font-semibold');
      expect(categories.length).toBe(4); // 4 categories
    });

    it('should display shortcuts under correct category', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      // Navigation category should contain F key
      const navigation = screen.getByText('ナビゲーション');
      expect(navigation).toBeInTheDocument();

      // Find F key under navigation
      expect(screen.getByText('F')).toBeInTheDocument();
      expect(screen.getByText('グラフ全体を画面にフィット')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have aria-label on close button', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const closeButton = screen.getByLabelText('閉じる');
      expect(closeButton).toBeInTheDocument();
    });

    it('should have proper heading hierarchy', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const mainHeading = screen.getByRole('heading', { name: 'キーボードショートカット', level: 2 });
      expect(mainHeading).toBeInTheDocument();
    });

    it('should have semantic kbd elements for keyboard keys', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const kbdElements = container.querySelectorAll('kbd');
      expect(kbdElements.length).toBeGreaterThan(10); // Multiple keyboard shortcuts
    });
  });

  describe('Footer', () => {
    it('should display help tip in footer', () => {
      renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      expect(screen.getByText(/ヒント:/)).toBeInTheDocument();
      expect(screen.getByText(/キーでいつでもこのガイドを表示できます/)).toBeInTheDocument();
    });

    it('should render footer with border', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const footer = container.querySelector('.border-t.border-gray-200');
      expect(footer).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('should have hover effects on shortcut items', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      const shortcutItems = container.querySelectorAll('.hover\\:bg-gray-50');
      expect(shortcutItems.length).toBeGreaterThan(0);
    });
  });

  describe('Content Completeness', () => {
    it('should display all 14 shortcuts', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      // Count all kbd elements (including footer hint with "?")
      const allKbdElements = container.querySelectorAll('kbd');
      // 14 shortcuts + 1 in footer hint = 15 total
      expect(allKbdElements.length).toBe(15);
    });

    it('should not have duplicate shortcuts', () => {
      const { container } = renderWithProviders(<KeyboardShortcutsGuide onClose={mockOnClose} />);

      // Check that each shortcut appears only once in kbd elements
      const kbdElements = container.querySelectorAll('kbd');
      const kbdTexts = Array.from(kbdElements).map(el => el.textContent);

      // Check specific shortcuts aren't duplicated (excluding the footer hint which also has kbd)
      const shortcutKeys = kbdTexts.slice(0, -1); // Exclude last one (footer hint)
      const uniqueKeys = new Set(shortcutKeys);
      expect(shortcutKeys.length).toBe(uniqueKeys.size);
    });
  });
});
