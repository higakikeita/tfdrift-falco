/**
 * Tests for PageErrorBoundary component
 */

import React from 'react';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PageErrorBoundary } from './PageErrorBoundary';

// Mock the logger
vi.mock('../utils/logger', () => ({
  logger: {
    error: vi.fn(),
    warn: vi.fn(),
    info: vi.fn(),
  },
}));

// Component that throws an error
const ThrowError: React.FC<{ shouldThrow?: boolean }> = ({ shouldThrow = true }) => {
  if (shouldThrow) {
    throw new Error('Test error in page');
  }
  return <div>Page content</div>;
};

describe('PageErrorBoundary', () => {
  // Suppress console.error during tests since we're intentionally triggering errors
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders children when no error', () => {
    render(
      <PageErrorBoundary>
        <div>Test page content</div>
      </PageErrorBoundary>
    );

    expect(screen.getByText('Test page content')).toBeInTheDocument();
  });

  it('displays error UI when child throws', () => {
    render(
      <PageErrorBoundary>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    expect(screen.getByText('Something Went Wrong')).toBeInTheDocument();
  });

  it('retry button exists and can be clicked', async () => {
    const user = userEvent.setup();
    render(
      <PageErrorBoundary>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    // Should show error UI initially
    expect(screen.getByText('Something Went Wrong')).toBeInTheDocument();

    // Verify retry button exists and is clickable
    const retryButton = screen.getByRole('button', { name: /Try Again/i });
    expect(retryButton).toBeInTheDocument();

    // Click should be possible
    await user.click(retryButton);
    expect(retryButton).not.toBeInTheDocument();
  });

  it('shows pageName in error message', () => {
    render(
      <PageErrorBoundary pageName="Dashboard">
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    expect(screen.getByText(/We encountered an error in Dashboard/)).toBeInTheDocument();
  });

  it('calls onError callback when provided', () => {
    const onError = vi.fn();

    render(
      <PageErrorBoundary onError={onError}>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    expect(onError).toHaveBeenCalled();
    expect(onError.mock.calls[0][0]).toBeInstanceOf(Error);
    expect(onError.mock.calls[0][0].message).toBe('Test error in page');
  });

  it('shows error message in development mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'development';

    render(
      <PageErrorBoundary>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    expect(screen.getByText('Test error in page')).toBeInTheDocument();

    process.env.NODE_ENV = originalEnv;
  });

  it('does not show error message in production mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'production';

    render(
      <PageErrorBoundary>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    // Should show generic message, not the specific error
    expect(screen.queryByText('Test error in page')).not.toBeInTheDocument();
    expect(screen.getByText('Something Went Wrong')).toBeInTheDocument();

    process.env.NODE_ENV = originalEnv;
  });

  it('uses default pageName when not provided', () => {
    render(
      <PageErrorBoundary>
        <ThrowError shouldThrow={true} />
      </PageErrorBoundary>
    );

    expect(screen.getByText(/We encountered an error in this section/)).toBeInTheDocument();
  });

  it('renders multiple error boundaries independently', () => {
    render(
      <>
        <PageErrorBoundary pageName="Page1">
          <ThrowError shouldThrow={true} />
        </PageErrorBoundary>
        <PageErrorBoundary pageName="Page2">
          <div>Page2 content</div>
        </PageErrorBoundary>
      </>
    );

    expect(screen.getByText(/We encountered an error in Page1/)).toBeInTheDocument();
    expect(screen.getByText('Page2 content')).toBeInTheDocument();
  });
});
