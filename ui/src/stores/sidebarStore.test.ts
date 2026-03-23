import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useSidebarStore } from './sidebarStore';
import { act } from '@testing-library/react';


const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
    removeItem: vi.fn((key: string) => { delete store[key]; }),
    clear: vi.fn(() => { store = {}; }),
    get length() { return Object.keys(store).length; },
    key: vi.fn((i: number) => Object.keys(store)[i] ?? null),
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('sidebarStore', () => {
  beforeEach(() => {
    // Reset store state
    act(() => {
      useSidebarStore.setState({ isCollapsed: false });
    });
  });

  it('should initialize with isCollapsed false', () => {
    const state = useSidebarStore.getState();
    expect(state.isCollapsed).toBe(false);
  });

  it('should toggle collapsed state', () => {
    const store = useSidebarStore.getState();

    act(() => {
      store.toggle();
    });
    expect(useSidebarStore.getState().isCollapsed).toBe(true);

    act(() => {
      useSidebarStore.getState().toggle();
    });
    expect(useSidebarStore.getState().isCollapsed).toBe(false);
  });

  it('should set collapsed to specific value', () => {
    act(() => {
      useSidebarStore.getState().setCollapsed(true);
    });
    expect(useSidebarStore.getState().isCollapsed).toBe(true);

    act(() => {
      useSidebarStore.getState().setCollapsed(false);
    });
    expect(useSidebarStore.getState().isCollapsed).toBe(false);
  });

  it('should set collapsed to same value without error', () => {
    act(() => {
      useSidebarStore.getState().setCollapsed(false);
    });
    expect(useSidebarStore.getState().isCollapsed).toBe(false);
  });
});
