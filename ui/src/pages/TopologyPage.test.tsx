/**
 * TopologyPage Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TopologyPage } from './TopologyPage';

// Mock graph components
vi.mock('../components/graph/GraphExportButton', () => ({
  GraphExportButton: () => <button data-testid="export-button">Export</button>,
}));

describe('TopologyPage', () => {
  describe('Rendering', () => {
    it('should render the page title', () => {
      render(<TopologyPage />);
      const title = screen.getByText('Topology');
      expect(title).toBeTruthy();
      expect(title.className).toContain('font-bold');
    });

    it('should render page with proper structure', () => {
      const { container } = render(<TopologyPage />);
      expect(container.querySelector('div')).toBeTruthy();
    });

    it('should render heading with correct styling', () => {
      render(<TopologyPage />);
      const title = screen.getByText('Topology');
      expect(title.className).toContain('text-2xl');
      expect(title.className).toContain('font-bold');
    });
  });

  describe('Header Section', () => {
    it('should render export button', () => {
      render(<TopologyPage />);
      expect(screen.getByTestId('export-button')).toBeTruthy();
    });

    it('should render description text', () => {
      render(<TopologyPage />);
      expect(screen.getByText('Infrastructure topology view')).toBeTruthy();
    });

    it('should have proper header layout', () => {
      const { container } = render(<TopologyPage />);
      const headerDiv = container.querySelector('[class*="flex"][class*="justify-between"]');
      expect(headerDiv).toBeTruthy();
    });
  });

  describe('Content Area', () => {
    it('should render topology graph placeholder', () => {
      render(<TopologyPage />);
      expect(screen.getByText(/Infrastructure Topology Graph/)).toBeTruthy();
    });

    it('should have minimum height for graph container', () => {
      const { container } = render(<TopologyPage />);
      const graphContainer = container.querySelector('[class*="min-h"]');
      expect(graphContainer).toBeTruthy();
    });

    it('should display placeholder message', () => {
      render(<TopologyPage />);
      const placeholder = screen.getByText(/Existing CytoscapeGraph component/);
      expect(placeholder).toBeTruthy();
    });
  });

  describe('Styling', () => {
    it('should apply flex layout', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('flex');
      expect(container.innerHTML).toContain('flex-col');
    });

    it('should have dark mode support', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('dark:');
    });

    it('should have proper spacing', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('space-y');
      expect(container.innerHTML).toContain('gap');
    });

    it('should have rounded border styling', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('rounded-xl');
      expect(container.innerHTML).toContain('border');
    });

    it('should apply light and dark background colors', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('bg-white');
      expect(container.innerHTML).toContain('dark:bg-slate-900');
    });
  });

  describe('Layout Structure', () => {
    it('should render full-height layout', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('h-full');
    });

    it('should render flex container with column direction', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('flex');
      expect(container.innerHTML).toContain('flex-col');
    });

    it('should have expandable content area', () => {
      const { container } = render(<TopologyPage />);
      expect(container.innerHTML).toContain('flex-1');
    });

    it('should have proper header and content separation', () => {
      render(<TopologyPage />);
      const title = screen.getByText('Topology');
      const description = screen.getByText('Infrastructure topology view');
      expect(title).toBeTruthy();
      expect(description).toBeTruthy();
    });
  });

  describe('Component Integration', () => {
    it('should include GraphExportButton', () => {
      render(<TopologyPage />);
      expect(screen.getByTestId('export-button')).toBeTruthy();
    });

    it('should render all required elements', () => {
      render(<TopologyPage />);
      expect(screen.getByText('Topology')).toBeTruthy();
      expect(screen.getByTestId('export-button')).toBeTruthy();
      expect(screen.getByText('Infrastructure topology view')).toBeTruthy();
    });
  });

  describe('Accessibility', () => {
    it('should have semantic heading', () => {
      const { container } = render(<TopologyPage />);
      const heading = container.querySelector('h1');
      expect(heading).toBeTruthy();
    });

    it('should have descriptive text', () => {
      render(<TopologyPage />);
      expect(screen.getByText('Infrastructure topology view')).toBeTruthy();
    });
  });

  describe('Integration', () => {
    it('should render complete page without errors', () => {
      const { container } = render(<TopologyPage />);
      expect(container).toBeTruthy();
      expect(container.querySelector('div')).toBeTruthy();
    });

    it('should render all visible content', () => {
      render(<TopologyPage />);
      expect(screen.getByText('Topology')).toBeTruthy();
      expect(screen.getByText('Infrastructure topology view')).toBeTruthy();
      expect(screen.getByText(/Infrastructure Topology Graph/)).toBeTruthy();
    });
  });
});
