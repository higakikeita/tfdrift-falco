import { describe, it, expect, beforeEach } from 'vitest';
import { useToastStore, toast } from './toastStore';
import { act } from '@testing-library/react';

describe('toastStore', () => {
  beforeEach(() => {
    act(() => {
      useToastStore.getState().clearAll();
    });
  });

  it('should initialize with empty toasts', () => {
    expect(useToastStore.getState().toasts).toEqual([]);
  });

  it('should add a toast', () => {
    act(() => {
      useToastStore.getState().addToast({
        type: 'success',
        title: 'Test Toast',
        message: 'Hello',
      });
    });

    const toasts = useToastStore.getState().toasts;
    expect(toasts).toHaveLength(1);
    expect(toasts[0].type).toBe('success');
    expect(toasts[0].title).toBe('Test Toast');
    expect(toasts[0].message).toBe('Hello');
    expect(toasts[0].id).toBeDefined();
  });

  it('should remove a toast by id', () => {
    let id: string;
    act(() => {
      id = useToastStore.getState().addToast({
        type: 'info',
        title: 'Removable',
      });
    });

    expect(useToastStore.getState().toasts).toHaveLength(1);

    act(() => {
      useToastStore.getState().removeToast(id);
    });

    expect(useToastStore.getState().toasts).toHaveLength(0);
  });

  it('should clear all toasts', () => {
    act(() => {
      useToastStore.getState().addToast({ type: 'success', title: 'A' });
      useToastStore.getState().addToast({ type: 'error', title: 'B' });
      useToastStore.getState().addToast({ type: 'warning', title: 'C' });
    });

    expect(useToastStore.getState().toasts).toHaveLength(3);

    act(() => {
      useToastStore.getState().clearAll();
    });

    expect(useToastStore.getState().toasts).toHaveLength(0);
  });

  it('should generate unique IDs for each toast', () => {
    act(() => {
      useToastStore.getState().addToast({ type: 'info', title: 'A' });
      useToastStore.getState().addToast({ type: 'info', title: 'B' });
    });

    const toasts = useToastStore.getState().toasts;
    expect(toasts[0].id).not.toBe(toasts[1].id);
  });

  describe('toast helper functions', () => {
    it('should create success toast', () => {
      act(() => {
        toast.success('Operation completed');
      });
      const t = useToastStore.getState().toasts[0];
      expect(t.type).toBe('success');
      expect(t.title).toBe('Operation completed');
    });

    it('should create error toast', () => {
      act(() => {
        toast.error('Something failed');
      });
      const t = useToastStore.getState().toasts[0];
      expect(t.type).toBe('error');
      expect(t.title).toBe('Something failed');
    });

    it('should create warning toast', () => {
      act(() => {
        toast.warning('Be careful');
      });
      const t = useToastStore.getState().toasts[0];
      expect(t.type).toBe('warning');
    });

    it('should create info toast', () => {
      act(() => {
        toast.info('FYI');
      });
      const t = useToastStore.getState().toasts[0];
      expect(t.type).toBe('info');
    });
  });
});
