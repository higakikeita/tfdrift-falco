import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('lucide-react', () => ({
  X: () => <div data-testid="close-icon" />,
  Search: () => <div data-testid="search-icon" />,
  Loader2: () => <div data-testid="loader-icon" />,
}));

vi.mock('../api/hooks', () => ({
  usePatternMatch: (pattern: unknown, enabled: boolean) => {
    if (!enabled) {
      return { data: null, isLoading: false, error: null };
    }
    return {
      data: {
        data: {
          matches: [
            [
              { id: 'node-1', labels: ['EC2'], properties: { name: 'instance-1', type: 'compute' } },
              { id: 'node-2', labels: ['Subnet'], properties: { name: 'subnet-1', type: 'network' } },
            ],
          ],
        },
      },
      isLoading: false,
      error: null,
    };
  },
}));

import PatternSearchPanel from './PatternSearchPanel';

describe('PatternSearchPanel', () => {
  const mockOnClose = vi.fn();
  const mockOnNodeSelect = vi.fn();

  it('should render without crashing', () => {
    const { container } = render(
      <PatternSearchPanel onClose={mockOnClose} />
    );
    expect(container.firstChild).toBeTruthy();
  });

  it('should render search form inputs', () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} />
    );
    expect(screen.getByText(/パターンマッチング検索/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/EC2, Compute/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/DEPENDS_ON, PART_OF/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/Subnet, Network/i)).toBeInTheDocument();
  });

  it('should render search and clear buttons', () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} />
    );
    expect(screen.getByText('検索')).toBeInTheDocument();
    expect(screen.getByText('クリア')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} />
    );
    const closeButton = screen.getByLabelText('閉じる');
    fireEvent.click(closeButton);
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('should handle input changes in search form', () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} />
    );
    const startLabelsInput = screen.getByPlaceholderText(/EC2, Compute/i) as HTMLInputElement;
    fireEvent.change(startLabelsInput, { target: { value: 'EC2' } });
    expect(startLabelsInput.value).toBe('EC2');
  });

  it('should render search results when enabled', async () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} onNodeSelect={mockOnNodeSelect} />
    );
    const startLabelsInput = screen.getByPlaceholderText(/EC2, Compute/i) as HTMLInputElement;
    fireEvent.change(startLabelsInput, { target: { value: 'EC2' } });

    const searchButton = screen.getByText('検索') as HTMLButtonElement;
    fireEvent.click(searchButton);

    // Results should appear (mocked data includes instance-1)
    expect(screen.getByText('instance-1')).toBeInTheDocument();
  });

  it('should call onNodeSelect when result node is clicked', async () => {
    render(
      <PatternSearchPanel onClose={mockOnClose} onNodeSelect={mockOnNodeSelect} />
    );
    const startLabelsInput = screen.getByPlaceholderText(/EC2, Compute/i) as HTMLInputElement;
    fireEvent.change(startLabelsInput, { target: { value: 'EC2' } });

    const searchButton = screen.getByText('検索');
    fireEvent.click(searchButton);

    const nodeButton = screen.getByText('instance-1');
    fireEvent.click(nodeButton);
    expect(mockOnNodeSelect).toHaveBeenCalledWith('node-1');
  });
});
