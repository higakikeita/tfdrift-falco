/**
 * WhyFalcoPage Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import WhyFalcoPage from './WhyFalcoPage';

describe('WhyFalcoPage', () => {
  const mockOnBack = vi.fn();

  describe('Rendering', () => {
    it('should render the page title', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('Why Falco?')).toBeTruthy();
    });

    it('should render page with proper structure', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.querySelector('div')).toBeTruthy();
    });

    it('should have dark mode background', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('from-gray-900');
      expect(container.innerHTML).toContain('to-gray-800');
    });
  });

  describe('Header', () => {
    it('should render header with title', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('Why Falco?')).toBeTruthy();
    });

    it('should render back button', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('Back to Graph')).toBeTruthy();
    });

    it('should call onBack when back button is clicked', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const backButton = screen.getByText('Back to Graph');
      fireEvent.click(backButton);
      expect(mockOnBack).toHaveBeenCalledTimes(1);
    });

    it('should have sticky header', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const header = container.querySelector('header');
      expect(header?.className).toContain('sticky');
    });
  });

  describe('Content Sections', () => {
    it('should render epigraph section', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const epigraph = screen.getByText(/Terraform tells us/);
      expect(epigraph).toBeTruthy();
    });

    it('should render blueprint section', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('The Perfect Blueprint')).toBeTruthy();
    });

    it('should render witness section', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('Enter Falco: The Witness')).toBeTruthy();
    });

    it('should have blueprint emoji', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const blueprintSection = screen.getByText('The Perfect Blueprint').parentElement;
      expect(blueprintSection).toBeTruthy();
    });

    it('should have witness emoji', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const witnessSection = screen.getByText('Enter Falco: The Witness').parentElement;
      expect(witnessSection).toBeTruthy();
    });
  });

  describe('Key Concepts', () => {
    it('should explain Terraform as "what should exist"', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('what should exist');
    });

    it('should explain Falco as "what actually happened"', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('what actually happened');
    });

    it('should mention blueprint and architecture metaphor', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('blueprint');
    });

    it('should describe Falco as witness', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('witness');
    });
  });

  describe('Witness Information', () => {
    it('should explain what Falco observes', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText(/observe the exact moment/)).toBeTruthy();
    });

    it('should display witness capabilities', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText(/Who touched the gate/)).toBeTruthy();
      expect(screen.getByText(/When they did it/)).toBeTruthy();
      expect(screen.getByText(/Which gate it was/)).toBeTruthy();
      expect(screen.getByText(/What their intent was/)).toBeTruthy();
    });
  });

  describe('Styling', () => {
    it('should have proper typography', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('text-3xl');
      expect(container.innerHTML).toContain('font-bold');
    });

    it('should have color accents', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('text-blue-400');
      expect(container.innerHTML).toContain('text-red-400');
      expect(container.innerHTML).toContain('text-purple-400');
    });

    it('should have proper spacing', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('space-y');
      expect(container.innerHTML).toContain('py-');
      expect(container.innerHTML).toContain('px-');
    });

    it('should have gradient background', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('gradient-to-b');
    });

    it('should have responsive max-width container', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('max-w-3xl');
      expect(container.innerHTML).toContain('mx-auto');
    });
  });

  describe('Interactive Elements', () => {
    it('should render back button as interactive element', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const button = screen.getByText('Back to Graph') as HTMLButtonElement;
      expect(button.tagName).toBe('BUTTON');
    });

    it('should handle multiple back button clicks', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const backButton = screen.getByText('Back to Graph');
      fireEvent.click(backButton);
      fireEvent.click(backButton);
      fireEvent.click(backButton);
      expect(mockOnBack).toHaveBeenCalledTimes(3);
    });

    it('should have hover effects on back button', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const button = container.querySelector('button');
      expect(button?.className).toContain('hover:');
    });
  });

  describe('Layout Structure', () => {
    it('should render full-height layout', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('min-h-screen');
    });

    it('should have proper main content area', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const main = container.querySelector('main');
      expect(main).toBeTruthy();
    });

    it('should organize content into sections', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const sections = container.querySelectorAll('section');
      expect(sections.length).toBeGreaterThan(0);
    });

    it('should use grid layout for witness cards', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('grid');
      expect(container.innerHTML).toContain('grid-cols-2');
    });
  });

  describe('Semantic HTML', () => {
    it('should use semantic header element', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const header = container.querySelector('header');
      expect(header).toBeTruthy();
    });

    it('should use semantic main element', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const main = container.querySelector('main');
      expect(main).toBeTruthy();
    });

    it('should use semantic section elements', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const sections = container.querySelectorAll('section');
      expect(sections.length).toBeGreaterThan(0);
    });

    it('should use blockquote for epigraph', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      const blockquotes = container.querySelectorAll('blockquote');
      expect(blockquotes.length).toBeGreaterThan(0);
    });
  });

  describe('Integration', () => {
    it('should render complete page without errors', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container).toBeTruthy();
      expect(container.querySelector('div')).toBeTruthy();
    });

    it('should render all major sections', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(screen.getByText('Why Falco?')).toBeTruthy();
      expect(screen.getByText('The Perfect Blueprint')).toBeTruthy();
      expect(screen.getByText('Enter Falco: The Witness')).toBeTruthy();
    });

    it('should be functional with onBack prop', () => {
      render(<WhyFalcoPage onBack={mockOnBack} />);
      const backButton = screen.getByText('Back to Graph');
      fireEvent.click(backButton);
      expect(mockOnBack).toHaveBeenCalled();
    });
  });

  describe('Content Completeness', () => {
    it('should contain narrative about Terraform', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('blueprint');
      expect(container.innerHTML).toContain('Terraform');
    });

    it('should contain narrative about Falco', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('Falco');
      expect(container.innerHTML).toContain('witness');
    });

    it('should explain the complementary relationship', () => {
      const { container } = render(<WhyFalcoPage onBack={mockOnBack} />);
      expect(container.innerHTML).toContain('what should exist');
      expect(container.innerHTML).toContain('what actually happened');
    });
  });
});
