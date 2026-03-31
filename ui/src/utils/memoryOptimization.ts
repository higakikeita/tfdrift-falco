/**
 * Memory Optimization Utilities
 * Performance helpers for large-scale graph rendering
 */

import * as React from 'react';
import { useMemo, useCallback, useRef, useEffect } from 'react';

/**
 * Debounce hook for search and filter inputs
 * Delays execution until user stops typing
 */
export function useDebounce<T>(value: T, delay: number = 300): T {
  const [debouncedValue, setDebouncedValue] = React.useState<T>(value);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    timeoutRef.current = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [value, delay]);

  return debouncedValue;
}

/**
 * Throttle hook for scroll and resize events
 * Limits execution frequency
 */
export function useThrottle<T>(value: T, limit: number = 100): T {
  const [throttledValue, setThrottledValue] = React.useState<T>(value);
  const lastRan = useRef<number | null>(null);

  React.useEffect(() => {
    if (lastRan.current === null) {
      lastRan.current = performance.now();
    }

    const handler = setTimeout(() => {
      const now = performance.now();
      const elapsed = now - (lastRan.current ?? 0);
      if (elapsed >= limit) {
        setThrottledValue(value);
        lastRan.current = now;
      }
    }, Math.max(0, limit - ((lastRan.current ? performance.now() - lastRan.current : 0))));

    return () => {
      clearTimeout(handler);
    };
  }, [value, limit]);

  return throttledValue;
}

/**
 * Memoized filter function for large datasets
 * Prevents unnecessary recalculations
 */
export function useMemoizedFilter<T>(
  items: T[],
  filterFn: (item: T) => boolean,
  dependencies: readonly unknown[] = []
): T[] {
  const memoizedFilterFn = useCallback(
    (item: T) => filterFn(item),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [filterFn, ...dependencies]
  );

  return useMemo(() => {
    return items.filter(memoizedFilterFn);
  }, [items, memoizedFilterFn]);
}

/**
 * Memoized sort function for large datasets
 */
export function useMemoizedSort<T>(
  items: T[],
  compareFn: (a: T, b: T) => number,
  dependencies: readonly unknown[] = []
): T[] {
  const memoizedCompareFn = useCallback(
    (a: T, b: T) => compareFn(a, b),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [compareFn, ...dependencies]
  );

  return useMemo(() => {
    return [...items].sort(memoizedCompareFn);
  }, [items, memoizedCompareFn]);
}

/**
 * Stable callback hook
 * Prevents unnecessary re-renders when passing callbacks to child components
 */
export function useStableCallback<T extends (...args: unknown[]) => unknown>(
  callback: T
): T {
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  return useCallback((...args: unknown[]) => {
    return callbackRef.current(...args);
  }, []) as T;
}

/**
 * Virtual list hook for rendering large lists efficiently
 * Only renders visible items
 */
export function useVirtualList<T>(
  items: T[],
  itemHeight: number,
  containerHeight: number,
  overscan: number = 5
) {
  const [scrollTop, setScrollTop] = React.useState(0);

  const visibleRange = useMemo(() => {
    const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
    const endIndex = Math.min(
      items.length - 1,
      Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
    );

    return { startIndex, endIndex };
  }, [scrollTop, itemHeight, containerHeight, items.length, overscan]);

  const visibleItems = useMemo(() => {
    return items.slice(visibleRange.startIndex, visibleRange.endIndex + 1);
  }, [items, visibleRange]);

  const totalHeight = items.length * itemHeight;
  const offsetY = visibleRange.startIndex * itemHeight;

  return {
    visibleItems,
    totalHeight,
    offsetY,
    startIndex: visibleRange.startIndex,
    endIndex: visibleRange.endIndex,
    onScroll: (e: React.UIEvent<HTMLDivElement>) => {
      setScrollTop(e.currentTarget.scrollTop);
    },
  };
}

/**
 * Chunk array into smaller batches for progressive rendering
 */
export function chunkArray<T>(array: T[], chunkSize: number): T[][] {
  const chunks: T[][] = [];
  for (let i = 0; i < array.length; i += chunkSize) {
    chunks.push(array.slice(i, i + chunkSize));
  }
  return chunks;
}

/**
 * Batch update hook using requestAnimationFrame
 * Schedules updates for next animation frame
 */
export function useAnimationFrame(callback: () => void, dependencies: readonly unknown[]) {
  const callbackRef = useRef(callback);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  useEffect(() => {
    const frameId = requestAnimationFrame(() => callbackRef.current());
    return () => cancelAnimationFrame(frameId);
  }, [...dependencies]);
}

/**
 * Performance monitor hook
 * Tracks render count and execution time
 */
export function usePerformanceMonitor() {
  const renderCountRef = useRef(0);
  const startTimeRef = useRef<number | null>(null);
  const [renderCount, setRenderCount] = React.useState(0);

  if (startTimeRef.current === null) {
    // eslint-disable-next-line react-hooks/purity -- performance.now() in ref init is safe, only runs once
    startTimeRef.current = performance.now();
  }
  renderCountRef.current++;

  useEffect(() => {
    const endTime = performance.now();
    startTimeRef.current = endTime;
    setRenderCount(renderCountRef.current);
  }, []);

  return {
    renderCount,
  };
}

/**
 * Memory-efficient WeakMap cache
 * Automatically garbage collects unused entries
 */
export class WeakCache<K extends object, V> {
  private cache = new WeakMap<K, V>();

  get(key: K): V | undefined {
    return this.cache.get(key);
  }

  set(key: K, value: V): void {
    this.cache.set(key, value);
  }

  has(key: K): boolean {
    return this.cache.has(key);
  }
}

/**
 * Calculate memory usage (development only)
 */
interface PerformanceMemory {
  usedJSHeapSize: number;
  totalJSHeapSize: number;
  jsHeapSizeLimit: number;
}

export function getMemoryUsage(): { used: number; total: number; percentage: number } | null {
  const performanceWithMemory = performance as unknown as { memory?: PerformanceMemory };
  if ('memory' in performance && performanceWithMemory.memory) {
    const memory = performanceWithMemory.memory;
    return {
      used: Math.round(memory.usedJSHeapSize / 1048576), // MB
      total: Math.round(memory.totalJSHeapSize / 1048576), // MB
      percentage: Math.round((memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100),
    };
  }
  return null;
}

/**
 * Lazy load component with retry logic
 */
export function lazyWithRetry<T extends React.ComponentType<Record<string, unknown>>>(
  componentImport: () => Promise<{ default: T }>,
  retries: number = 3
): React.LazyExoticComponent<T> {
  return React.lazy(async () => {
    for (let i = 0; i < retries; i++) {
      try {
        return await componentImport();
      } catch (error) {
        if (i === retries - 1) throw error;
        // Exponential backoff
        await new Promise(resolve => setTimeout(resolve, 1000 * Math.pow(2, i)));
      }
    }
    throw new Error('Failed to load component');
  });
}

// React is imported at the top of the file
