import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ToastContainer } from './ToastContainer';
import { useToastStore } from '../../stores/toastStore';
import { act } from '@testing-library/react';

describe('ToastContainer component', () => {
  beforeEach(() => {
    act(() => {
      useToastStore.getState().clearAll();
    });
  });

  it('should render nothing when no toasts', () => {
    const { container } = render(<ToastContainer />);
    expect(container.firstChild).toBeNull();
  });

  it('should render toast container when toasts exist', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success!',
      });
    });

    const { container } = render(<ToastContainer />);
    const toastContainer = container.querySelector('div');
    expect(toastContainer).toHaveClass('fixed');
    expect(toastContainer).toHaveClass('bottom-4');
    expect(toastContainer).toHaveClass('right-4');
    expect(toastContainer).toHaveClass('z-[100]');
  });

  it('should render success toast with correct styles', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Operation completed',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');
    expect(toast).toHaveClass('border-green-200');
    expect(toast).toHaveClass('bg-green-50');
  });

  it('should render error toast with correct styles', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'error',
        title: 'Something went wrong',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');
    expect(toast).toHaveClass('border-red-200');
    expect(toast).toHaveClass('bg-red-50');
  });

  it('should render warning toast with correct styles', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'warning',
        title: 'Warning message',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');
    expect(toast).toHaveClass('border-amber-200');
    expect(toast).toHaveClass('bg-amber-50');
  });

  it('should render info toast with correct styles', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'info',
        title: 'Information',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');
    expect(toast).toHaveClass('border-blue-200');
    expect(toast).toHaveClass('bg-blue-50');
  });

  it('should display toast title', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success title',
      });
    });

    render(<ToastContainer />);
    expect(screen.getByText('Success title')).toBeInTheDocument();
  });

  it('should display toast message', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
        message: 'Operation completed successfully',
      });
    });

    render(<ToastContainer />);
    expect(screen.getByText('Operation completed successfully')).toBeInTheDocument();
  });

  it('should not display message when not provided', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
      });
    });

    render(<ToastContainer />);
    expect(screen.getByText('Success')).toBeInTheDocument();
  });

  it('should render close button', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
      });
    });

    render(<ToastContainer />);
    const closeButton = screen.getByRole('button');
    expect(closeButton).toBeInTheDocument();
  });

  it('should remove toast when close button clicked', async () => {
    const user = userEvent.setup();

    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
      });
    });

    const { rerender } = render(<ToastContainer />);
    expect(screen.getByRole('alert')).toBeInTheDocument();

    const closeButton = screen.getByRole('button');
    await user.click(closeButton);

    rerender(<ToastContainer />);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('should render multiple toasts', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success toast',
      });
      useToastStore.getState().addToast({
        type: 'error',
        title: 'Error toast',
      });
    });

    render(<ToastContainer />);
    expect(screen.getByText('Success toast')).toBeInTheDocument();
    expect(screen.getByText('Error toast')).toBeInTheDocument();
  });

  it('should render correct icon for success', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
      });
    });

    const { container } = render(<ToastContainer />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should render correct icon for error', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'error',
        title: 'Error',
      });
    });

    const { container } = render(<ToastContainer />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should render correct icon for warning', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'warning',
        title: 'Warning',
      });
    });

    const { container } = render(<ToastContainer />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should render correct icon for info', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'info',
        title: 'Info',
      });
    });

    const { container } = render(<ToastContainer />);
    const svg = container.querySelector('svg');
    expect(svg).toBeInTheDocument();
  });

  it('should have proper toast styling classes', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Test',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');

    expect(toast).toHaveClass('flex');
    expect(toast).toHaveClass('items-start');
    expect(toast).toHaveClass('gap-3');
    expect(toast).toHaveClass('px-4');
    expect(toast).toHaveClass('py-3');
    expect(toast).toHaveClass('rounded-lg');
    expect(toast).toHaveClass('border');
    expect(toast).toHaveClass('shadow-lg');
  });

  it('should render icon with proper color', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Success',
      });
    });

    render(<ToastContainer />);
    const icon = screen.getByRole('alert').querySelector('svg');
    expect(icon).toHaveClass('text-green-600');
  });

  it('should handle rapid toast additions', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Toast 1',
      });
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Toast 2',
      });
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Toast 3',
      });
    });

    render(<ToastContainer />);
    expect(screen.getByText('Toast 1')).toBeInTheDocument();
    expect(screen.getByText('Toast 2')).toBeInTheDocument();
    expect(screen.getByText('Toast 3')).toBeInTheDocument();
  });

  it('should render with animation class', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Animated',
      });
    });

    render(<ToastContainer />);
    const toast = screen.getByRole('alert');
    expect(toast).toHaveClass('animate-slide-in-right');
  });

  it('should position container fixed at bottom right', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Test',
      });
    });

    const { container } = render(<ToastContainer />);
    const toastContainer = container.querySelector('div');

    expect(toastContainer).toHaveClass('fixed');
    expect(toastContainer).toHaveClass('bottom-4');
    expect(toastContainer).toHaveClass('right-4');
    expect(toastContainer).toHaveClass('flex');
    expect(toastContainer).toHaveClass('flex-col');
    expect(toastContainer).toHaveClass('gap-2');
    expect(toastContainer).toHaveClass('max-w-sm');
  });

  it('should render toast with title and message formatting', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Operation Successful',
        message: 'Your changes have been saved to the database.',
      });
    });

    render(<ToastContainer />);
    const title = screen.getByText('Operation Successful');
    const message = screen.getByText('Your changes have been saved to the database.');

    expect(title).toHaveClass('text-sm');
    expect(title).toHaveClass('font-medium');
    expect(message).toHaveClass('text-xs');
  });
});
