/**
 * TopologyPage Tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TopologyPage } from './TopologyPage';
import { useGraph } from '../api/hooks/useGraph';

vi.mock('../components/graph/GraphExportButton', () => ({
  GraphExportButton: () => <button data-testid="export-button">Export</button>,
}));

// The real Cytoscape renderer needs a canvas; stub it to a marker.
vi.mock('../components/CytoscapeGraph', () => ({
  CytoscapeGraph: () => <div data-testid="cytoscape-graph" />,
}));

vi.mock('../api/hooks/useGraph', () => ({
  useGraph: vi.fn(),
}));

const mockUseGraph = vi.mocked(useGraph);

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <TopologyPage />
    </QueryClientProvider>
  );
}

describe('TopologyPage', () => {
  beforeEach(() => {
    // Default: loaded with a non-empty graph
    mockUseGraph.mockReturnValue({
      data: { nodes: [{ data: { id: 'n1' } }], edges: [] },
      isLoading: false,
      isError: false,
    } as never);
  });

  describe('Rendering', () => {
    it('should render the page title', () => {
      renderPage();
      const title = screen.getByText('Topology');
      expect(title.className).toContain('font-bold');
    });

    it('should render the export button and description', () => {
      renderPage();
      expect(screen.getByTestId('export-button')).toBeTruthy();
      expect(screen.getByText('Infrastructure topology view')).toBeTruthy();
    });
  });

  describe('Graph wiring', () => {
    it('renders the CytoscapeGraph when the API returns nodes', () => {
      renderPage();
      expect(screen.getByTestId('cytoscape-graph')).toBeTruthy();
    });

    it('shows a loading state while fetching', () => {
      mockUseGraph.mockReturnValue({ data: undefined, isLoading: true, isError: false } as never);
      renderPage();
      expect(screen.getByText(/Loading topology/)).toBeTruthy();
    });

    it('shows an empty state when there are no nodes', () => {
      mockUseGraph.mockReturnValue({
        data: { nodes: [], edges: [] },
        isLoading: false,
        isError: false,
      } as never);
      renderPage();
      expect(screen.getByText(/No topology yet/)).toBeTruthy();
      expect(screen.queryByTestId('cytoscape-graph')).toBeNull();
    });

    it('shows an error state when the query fails', () => {
      mockUseGraph.mockReturnValue({ data: undefined, isLoading: false, isError: true } as never);
      renderPage();
      expect(screen.getByText(/Failed to load/)).toBeTruthy();
    });
  });

  describe('Styling', () => {
    it('applies flex column layout and dark-mode classes', () => {
      const { container } = renderPage();
      expect(container.innerHTML).toContain('flex-col');
      expect(container.innerHTML).toContain('dark:');
      expect(container.innerHTML).toContain('rounded-xl');
    });
  });
});
