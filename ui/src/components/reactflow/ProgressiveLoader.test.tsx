import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ProgressiveLoader } from './ProgressiveLoader';

describe('ProgressiveLoader component', () => {
  it('should not render when not loading', () => {
    const { container } = render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={false}
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it('should render when loading', () => {
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/loading graph/i)).toBeInTheDocument();
  });

  it('should display current batch and total batches', () => {
    render(
      <ProgressiveLoader
        progress={30}
        currentBatch={3}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText('3 / 10')).toBeInTheDocument();
  });

  it('should display progress percentage', () => {
    render(
      <ProgressiveLoader
        progress={75}
        currentBatch={7}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/75% complete/i)).toBeInTheDocument();
  });

  it('should calculate remaining batches correctly', () => {
    render(
      <ProgressiveLoader
        progress={40}
        currentBatch={4}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/6 batches remaining/i)).toBeInTheDocument();
  });

  it('should render skip button when onSkip callback provided', () => {
    const handleSkip = vi.fn();
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
        onSkip={handleSkip}
      />
    );
    const skipButton = screen.getByRole('button', { name: /load all now/i });
    expect(skipButton).toBeInTheDocument();
  });

  it('should render cancel button when onCancel callback provided', () => {
    const handleCancel = vi.fn();
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
        onCancel={handleCancel}
      />
    );
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    expect(cancelButton).toBeInTheDocument();
  });

  it('should call onSkip when skip button clicked', async () => {
    const user = userEvent.setup();
    const handleSkip = vi.fn();
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
        onSkip={handleSkip}
      />
    );
    const skipButton = screen.getByRole('button', { name: /load all now/i });
    await user.click(skipButton);
    expect(handleSkip).toHaveBeenCalledTimes(1);
  });

  it('should call onCancel when cancel button clicked', async () => {
    const user = userEvent.setup();
    const handleCancel = vi.fn();
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
        onCancel={handleCancel}
      />
    );
    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await user.click(cancelButton);
    expect(handleCancel).toHaveBeenCalledTimes(1);
  });

  it('should not render skip button when onSkip not provided', () => {
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.queryByRole('button', { name: /load all now/i })).not.toBeInTheDocument();
  });

  it('should not render cancel button when onCancel not provided', () => {
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
  });

  it('should display loading icon', () => {
    const { container } = render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
    expect(svg).toHaveClass('animate-spin');
  });

  it('should have correct container styling', () => {
    const { container } = render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    const loaderDiv = container.querySelector('div');
    expect(loaderDiv).toHaveClass('fixed');
    expect(loaderDiv).toHaveClass('top-4');
    expect(loaderDiv).toHaveClass('left-1/2');
    expect(loaderDiv).toHaveClass('z-50');
    expect(loaderDiv).toHaveClass('bg-white');
    expect(loaderDiv).toHaveClass('rounded-xl');
    expect(loaderDiv).toHaveClass('shadow-2xl');
  });

  it('should render progress bar', () => {
    const { container } = render(
      <ProgressiveLoader
        progress={60}
        currentBatch={6}
        totalBatches={10}
        isLoading={true}
      />
    );
    const progressBars = container.querySelectorAll('div');
    expect(progressBars.length).toBeGreaterThan(0);
  });

  it('should display helpful message', () => {
    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/loading in batches for optimal performance/i)).toBeInTheDocument();
  });

  it('should handle progress at 0%', () => {
    render(
      <ProgressiveLoader
        progress={0}
        currentBatch={0}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/0% complete/i)).toBeInTheDocument();
  });

  it('should handle progress at 100%', () => {
    render(
      <ProgressiveLoader
        progress={100}
        currentBatch={10}
        totalBatches={10}
        isLoading={true}
      />
    );
    expect(screen.getByText(/100% complete/i)).toBeInTheDocument();
  });

  it('should update progress when props change', () => {
    const { rerender } = render(
      <ProgressiveLoader
        progress={25}
        currentBatch={2}
        totalBatches={10}
        isLoading={true}
      />
    );

    expect(screen.getByText(/25% complete/i)).toBeInTheDocument();

    rerender(
      <ProgressiveLoader
        progress={75}
        currentBatch={7}
        totalBatches={10}
        isLoading={true}
      />
    );

    expect(screen.getByText(/75% complete/i)).toBeInTheDocument();
  });

  it('should render both buttons when both callbacks provided', () => {
    const handleSkip = vi.fn();
    const handleCancel = vi.fn();

    render(
      <ProgressiveLoader
        progress={50}
        currentBatch={5}
        totalBatches={10}
        isLoading={true}
        onSkip={handleSkip}
        onCancel={handleCancel}
      />
    );

    expect(screen.getByRole('button', { name: /load all now/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });
});
