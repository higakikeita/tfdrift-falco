import { describe, it, expect } from 'vitest';
import { chunkArray, WeakCache, getMemoryUsage } from './memoryOptimization';

describe('memoryOptimization', () => {
  describe('chunkArray', () => {
    it('should split array into chunks of specified size', () => {
      const arr = [1, 2, 3, 4, 5, 6];
      const result = chunkArray(arr, 2);
      expect(result).toEqual([[1, 2], [3, 4], [5, 6]]);
    });

    it('should handle last chunk being smaller than chunkSize', () => {
      const arr = [1, 2, 3, 4, 5];
      const result = chunkArray(arr, 3);
      expect(result).toEqual([[1, 2, 3], [4, 5]]);
    });

    it('should return single chunk when array is smaller than chunkSize', () => {
      const arr = [1, 2];
      const result = chunkArray(arr, 5);
      expect(result).toEqual([[1, 2]]);
    });

    it('should return empty array for empty input', () => {
      const result = chunkArray([], 3);
      expect(result).toEqual([]);
    });

    it('should handle chunkSize of 1', () => {
      const arr = ['a', 'b', 'c'];
      const result = chunkArray(arr, 1);
      expect(result).toEqual([['a'], ['b'], ['c']]);
    });

    it('should handle array with single element', () => {
      const result = chunkArray([42], 3);
      expect(result).toEqual([[42]]);
    });
  });

  describe('WeakCache', () => {
    it('should store and retrieve values', () => {
      const cache = new WeakCache<object, string>();
      const key = { id: 1 };
      cache.set(key, 'value1');
      expect(cache.get(key)).toBe('value1');
    });

    it('should return undefined for missing keys', () => {
      const cache = new WeakCache<object, string>();
      const key = { id: 1 };
      expect(cache.get(key)).toBeUndefined();
    });

    it('should report has correctly', () => {
      const cache = new WeakCache<object, number>();
      const key = { id: 1 };
      expect(cache.has(key)).toBe(false);
      cache.set(key, 42);
      expect(cache.has(key)).toBe(true);
    });

    it('should overwrite existing values', () => {
      const cache = new WeakCache<object, string>();
      const key = { id: 1 };
      cache.set(key, 'old');
      cache.set(key, 'new');
      expect(cache.get(key)).toBe('new');
    });

    it('should handle multiple keys independently', () => {
      const cache = new WeakCache<object, number>();
      const key1 = { id: 1 };
      const key2 = { id: 2 };
      cache.set(key1, 100);
      cache.set(key2, 200);
      expect(cache.get(key1)).toBe(100);
      expect(cache.get(key2)).toBe(200);
    });
  });

  describe('getMemoryUsage', () => {
    it('should return null when performance.memory is not available', () => {
      const result = getMemoryUsage();
      // In test environment, performance.memory is typically not available
      expect(result).toBeNull();
    });

    it('should return memory info when performance.memory is available', () => {
      const mockMemory = {
        usedJSHeapSize: 50 * 1048576,
        totalJSHeapSize: 100 * 1048576,
        jsHeapSizeLimit: 200 * 1048576,
      };
      const originalMemory = (performance as unknown as Record<string, unknown>).memory;
      Object.defineProperty(performance, 'memory', {
        value: mockMemory,
        configurable: true,
        writable: true,
      });
      const result = getMemoryUsage();
      expect(result).not.toBeNull();
      expect(result!.used).toBe(50);
      expect(result!.total).toBe(100);
      expect(result!.percentage).toBe(25);
      if (originalMemory !== undefined) {
        Object.defineProperty(performance, 'memory', {
          value: originalMemory,
          configurable: true,
          writable: true,
        });
      } else {
        delete (performance as unknown as Record<string, unknown>).memory;
      }
    });
  });
});
