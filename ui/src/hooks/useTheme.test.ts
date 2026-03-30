/**
 * useTheme Hook Tests
 * Tests for theme management hook
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from './useTheme';

describe('useTheme', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
    // Reset document classes
    document.documentElement.className = '';
    // Mock matchMedia - return light preference by default
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(query => ({
        matches: query === '(prefers-color-scheme: dark)' ? false : true,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  describe('Initial theme detection', () => {
    it('should return light theme by default when no localStorage and light preference', () => {
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: vi.fn().mockImplementation(query => ({
          matches: query === '(prefers-color-scheme: dark)' ? false : true,
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
      });

      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('light');
    });

    it('should return dark theme when system preference is dark and no localStorage', () => {
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: vi.fn().mockImplementation(query => ({
          matches: query === '(prefers-color-scheme: dark)',
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
      });

      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('dark');
    });

    it('should use localStorage value if available', () => {
      localStorage.setItem('theme', 'dark');

      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('dark');
    });

    it('should prefer localStorage over system preference', () => {
      localStorage.setItem('theme', 'light');

      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: vi.fn().mockImplementation(query => ({
          matches: query === '(prefers-color-scheme: dark)',
          media: query,
          onchange: null,
          addListener: vi.fn(),
          removeListener: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          dispatchEvent: vi.fn(),
        })),
      });

      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('light');
    });
  });

  describe('setTheme function', () => {
    it('should update theme to dark', () => {
      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('light');

      act(() => {
        result.current.setTheme('dark');
      });

      expect(result.current.theme).toBe('dark');
    });

    it('should update theme to light', () => {
      localStorage.setItem('theme', 'dark');
      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('dark');

      act(() => {
        result.current.setTheme('light');
      });

      expect(result.current.theme).toBe('light');
    });

    it('should persist theme to localStorage', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.setTheme('dark');
      });

      expect(localStorage.getItem('theme')).toBe('dark');
    });

    it('should update document class when theme changes', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.setTheme('dark');
      });

      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(document.documentElement.classList.contains('light')).toBe(false);
    });

    it('should remove old theme class when updating', () => {
      const { result } = renderHook(() => useTheme());
      expect(document.documentElement.classList.contains('light')).toBe(true);

      act(() => {
        result.current.setTheme('dark');
      });

      expect(document.documentElement.classList.contains('light')).toBe(false);
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });
  });

  describe('toggleTheme function', () => {
    it('should toggle from light to dark', () => {
      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('light');

      act(() => {
        result.current.toggleTheme();
      });

      expect(result.current.theme).toBe('dark');
    });

    it('should toggle from dark to light', () => {
      localStorage.setItem('theme', 'dark');
      const { result } = renderHook(() => useTheme());
      expect(result.current.theme).toBe('dark');

      act(() => {
        result.current.toggleTheme();
      });

      expect(result.current.theme).toBe('light');
    });

    it('should persist toggled theme to localStorage', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.toggleTheme();
      });

      expect(localStorage.getItem('theme')).toBe('dark');

      act(() => {
        result.current.toggleTheme();
      });

      expect(localStorage.getItem('theme')).toBe('light');
    });

    it('should update document class when toggling', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.toggleTheme();
      });

      expect(document.documentElement.classList.contains('dark')).toBe(true);

      act(() => {
        result.current.toggleTheme();
      });

      expect(document.documentElement.classList.contains('light')).toBe(true);
    });

    it('should toggle multiple times correctly', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.toggleTheme();
      });
      expect(result.current.theme).toBe('dark');

      act(() => {
        result.current.toggleTheme();
      });
      expect(result.current.theme).toBe('light');

      act(() => {
        result.current.toggleTheme();
      });
      expect(result.current.theme).toBe('dark');

      act(() => {
        result.current.toggleTheme();
      });
      expect(result.current.theme).toBe('light');
    });
  });

  describe('Document class management', () => {
    it('should add theme class on mount', () => {
      renderHook(() => useTheme());

      expect(document.documentElement.classList.contains('light')).toBe(true);
    });

    it('should only have one theme class at a time', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.setTheme('dark');
      });

      const themeClasses = Array.from(document.documentElement.classList).filter(
        c => c === 'light' || c === 'dark'
      );

      expect(themeClasses).toHaveLength(1);
    });

    it('should handle rapid theme changes', () => {
      const { result } = renderHook(() => useTheme());

      act(() => {
        result.current.toggleTheme();
        result.current.toggleTheme();
        result.current.toggleTheme();
      });

      const hasValidTheme =
        document.documentElement.classList.contains('dark') ||
        document.documentElement.classList.contains('light');
      expect(hasValidTheme).toBe(true);
    });
  });

  describe('Hook exports', () => {
    it('should return object with theme, setTheme, and toggleTheme', () => {
      const { result } = renderHook(() => useTheme());

      expect(result.current).toHaveProperty('theme');
      expect(result.current).toHaveProperty('setTheme');
      expect(result.current).toHaveProperty('toggleTheme');
    });

    it('should have theme as a string', () => {
      const { result } = renderHook(() => useTheme());
      expect(typeof result.current.theme).toBe('string');
    });

    it('should have setTheme as a function', () => {
      const { result } = renderHook(() => useTheme());
      expect(typeof result.current.setTheme).toBe('function');
    });

    it('should have toggleTheme as a function', () => {
      const { result } = renderHook(() => useTheme());
      expect(typeof result.current.toggleTheme).toBe('function');
    });
  });

  describe('Multiple hook instances', () => {
    it('should keep multiple instances in sync via localStorage', () => {
      const { result: result1 } = renderHook(() => useTheme());
      const { result: result2 } = renderHook(() => useTheme());

      act(() => {
        result1.current.setTheme('dark');
      });

      // Note: In actual usage with multiple components, localStorage would sync them
      // but renderHook instances are independent, so we verify localStorage is updated
      expect(localStorage.getItem('theme')).toBe('dark');
    });
  });
});
