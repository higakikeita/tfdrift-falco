/**
 * WelcomeModal Component Tests
 * Tests for welcome/onboarding modal functionality
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { renderWithProviders, userEvent } from '@/__tests__/utils/testUtils';
import { WelcomeModal, shouldShowWelcome, resetWelcome } from './WelcomeModal';

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  X: () => <div data-testid="x-icon">X</div>,
  Zap: () => <div data-testid="zap-icon">Zap</div>,
  Network: () => <div data-testid="network-icon">Network</div>,
  Search: () => <div data-testid="search-icon">Search</div>,
  Target: () => <div data-testid="target-icon">Target</div>,
  FileImage: () => <div data-testid="fileimage-icon">FileImage</div>,
  Keyboard: () => <div data-testid="keyboard-icon">Keyboard</div>,
}));

describe('WelcomeModal', () => {
  const mockOnClose = vi.fn();
  const STORAGE_KEY = 'tfdrift-welcome-seen';

  beforeEach(() => {
    mockOnClose.mockClear();
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  describe('Rendering', () => {
    it('should render welcome modal with first step', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      expect(screen.getByText('TFDrift-Falcoへようこそ')).toBeInTheDocument();
      expect(screen.getByText('ステップ 1 / 6')).toBeInTheDocument();
    });

    it('should display step description', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Use partial text match as the full text might be broken across elements
      expect(screen.getByText(/クラウドインフラのセキュリティとドリフト分析/)).toBeInTheDocument();
    });

    it('should display step details as bullet points', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      expect(screen.getByText(/Terraform Drift → IAM → Kubernetes → Falcoの因果関係を追跡/)).toBeInTheDocument();
    });

    it('should display progress dots', () => {
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const progressDots = container.querySelectorAll('.rounded-full.w-2\\.5.h-2\\.5');
      expect(progressDots).toHaveLength(6); // 6 steps
    });

    it('should highlight current step dot', () => {
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const activeDot = container.querySelector('.bg-blue-600.w-8');
      expect(activeDot).toBeInTheDocument();
    });
  });

  describe('Navigation', () => {
    it('should navigate to next step when "次へ" button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const nextButton = screen.getByRole('button', { name: '次へ' });
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText('グラフの操作方法')).toBeInTheDocument();
        expect(screen.getByText('ステップ 2 / 6')).toBeInTheDocument();
      });
    });

    it('should navigate to previous step when "戻る" button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Navigate to step 2 first
      const nextButton = screen.getByRole('button', { name: '次へ' });
      await user.click(nextButton);

      // Then go back
      const prevButton = await screen.findByRole('button', { name: '戻る' });
      await user.click(prevButton);

      await waitFor(() => {
        expect(screen.getByText('TFDrift-Falcoへようこそ')).toBeInTheDocument();
        expect(screen.getByText('ステップ 1 / 6')).toBeInTheDocument();
      });
    });

    it('should not show "戻る" button on first step', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const prevButton = screen.queryByRole('button', { name: '戻る' });
      expect(prevButton).not.toBeInTheDocument();
    });

    it('should show "始める" button on last step', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Navigate to last step (6 steps total)
      const nextButton = screen.getByRole('button', { name: '次へ' });
      for (let i = 0; i < 5; i++) {
        await user.click(nextButton);
      }

      await waitFor(() => {
        expect(screen.getByRole('button', { name: '始める' })).toBeInTheDocument();
      });
    });

    it('should navigate directly to step by clicking progress dot', async () => {
      const user = userEvent.setup();
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const progressDots = container.querySelectorAll('button[aria-label*="ステップ"]');
      await user.click(progressDots[2]); // Click step 3

      await waitFor(() => {
        expect(screen.getByText('依存関係の可視化')).toBeInTheDocument();
        expect(screen.getByText('ステップ 3 / 6')).toBeInTheDocument();
      });
    });
  });

  describe('Step Content', () => {
    it('should show correct content for each step', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const expectedSteps = [
        'TFDrift-Falcoへようこそ',
        'グラフの操作方法',
        '依存関係の可視化',
        '検索とフィルタリング',
        'エクスポートと共有',
        'キーボードショートカット',
      ];

      for (let i = 0; i < expectedSteps.length; i++) {
        expect(screen.getByText(expectedSteps[i])).toBeInTheDocument();

        if (i < expectedSteps.length - 1) {
          const nextButton = screen.getByRole('button', { name: '次へ' });
          await user.click(nextButton);
          await waitFor(() => {
            expect(screen.getByText(`ステップ ${i + 2} / 6`)).toBeInTheDocument();
          });
        }
      }
    });
  });

  describe('LocalStorage Integration', () => {
    it('should save to localStorage and close when "始める" is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Navigate to last step
      const nextButton = screen.getByRole('button', { name: '次へ' });
      for (let i = 0; i < 5; i++) {
        await user.click(nextButton);
      }

      // Click finish button
      const finishButton = await screen.findByRole('button', { name: '始める' });
      await user.click(finishButton);

      await waitFor(() => {
        expect(localStorage.getItem(STORAGE_KEY)).toBe('true');
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should save to localStorage and close when "スキップ" is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const skipButton = screen.getAllByRole('button', { name: 'スキップ' })[0];
      await user.click(skipButton);

      await waitFor(() => {
        expect(localStorage.getItem(STORAGE_KEY)).toBe('true');
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should close when X button is clicked', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      await waitFor(() => {
        expect(localStorage.getItem(STORAGE_KEY)).toBe('true');
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });
  });

  describe('Helper Functions', () => {
    it('shouldShowWelcome should return true when localStorage is empty', () => {
      expect(shouldShowWelcome()).toBe(true);
    });

    it('shouldShowWelcome should return false when localStorage has the key', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      expect(shouldShowWelcome()).toBe(false);
    });

    it('resetWelcome should remove localStorage key', () => {
      localStorage.setItem(STORAGE_KEY, 'true');
      expect(shouldShowWelcome()).toBe(false);

      resetWelcome();
      expect(shouldShowWelcome()).toBe(true);
    });
  });

  describe('Accessibility', () => {
    it('should have aria-label on close button', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const closeButton = screen.getByLabelText('閉じる');
      expect(closeButton).toBeInTheDocument();
    });

    it('should have aria-labels on progress dots', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      for (let i = 1; i <= 6; i++) {
        const dot = screen.getByLabelText(`ステップ ${i}に移動`);
        expect(dot).toBeInTheDocument();
      }
    });

    it('should have proper heading hierarchy', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const heading = screen.getByRole('heading', { name: 'TFDrift-Falcoへようこそ' });
      expect(heading).toBeInTheDocument();
    });
  });

  describe('Styling and Animations', () => {
    it('should have backdrop with blur', () => {
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const backdrop = container.querySelector('.backdrop-blur-sm');
      expect(backdrop).toBeInTheDocument();
    });

    it('should have gradient header', () => {
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const header = container.querySelector('.bg-gradient-to-r.from-blue-600.to-indigo-600');
      expect(header).toBeInTheDocument();
    });

    it('should have animation classes', () => {
      const { container } = renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const modal = container.querySelector('.animate-in.zoom-in-95');
      expect(modal).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid navigation', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const nextButton = screen.getByRole('button', { name: '次へ' });

      // Rapidly click next multiple times
      await user.click(nextButton);
      await user.click(nextButton);
      await user.click(nextButton);

      await waitFor(() => {
        expect(screen.getByText('検索とフィルタリング')).toBeInTheDocument();
      });
    });

    it('should not navigate beyond last step', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Navigate to last step
      const nextButton = screen.getByRole('button', { name: '次へ' });
      for (let i = 0; i < 5; i++) {
        await user.click(nextButton);
      }

      // Try to go further (button should be "始める" now, not "次へ")
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: '始める' })).toBeInTheDocument();
      });
    });

    it('should not navigate before first step', () => {
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // Should not have "戻る" button on first step
      expect(screen.queryByRole('button', { name: '戻る' })).not.toBeInTheDocument();
    });
  });
});
