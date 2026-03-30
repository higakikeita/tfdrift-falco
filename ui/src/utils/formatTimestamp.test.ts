/**
 * formatTimestamp Utility Tests
 * Tests for timestamp formatting function
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { formatTimestamp } from './formatTimestamp';

describe('formatTimestamp', () => {
  let realNow: number;

  beforeEach(() => {
    // Mock Date.now() to have consistent test results
    realNow = new Date('2024-01-15T12:00:00Z').getTime();
    vi.useFakeTimers();
    vi.setSystemTime(new Date(realNow));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Recent timestamps (within minutes)', () => {
    it('should show "たった今" for timestamps less than 1 minute ago', () => {
      const timestamp = new Date(realNow - 30000).toISOString(); // 30 seconds ago
      const result = formatTimestamp(timestamp);
      expect(result).toBe('たった今');
    });

    it('should show "たった今" for timestamps at exactly 0 seconds ago', () => {
      const timestamp = new Date(realNow).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('たった今');
    });

    it('should show minute format for 1 minute ago', () => {
      const timestamp = new Date(realNow - 60000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('1分前');
    });

    it('should show minute format for 30 minutes ago', () => {
      const timestamp = new Date(realNow - 30 * 60000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('30分前');
    });

    it('should show minute format for 59 minutes ago', () => {
      const timestamp = new Date(realNow - 59 * 60000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('59分前');
    });
  });

  describe('Recent timestamps (within hours)', () => {
    it('should show hour format for 1 hour ago', () => {
      const timestamp = new Date(realNow - 3600000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('1時間前');
    });

    it('should show hour format for 5 hours ago', () => {
      const timestamp = new Date(realNow - 5 * 3600000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('5時間前');
    });

    it('should show hour format for 23 hours ago', () => {
      const timestamp = new Date(realNow - 23 * 3600000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('23時間前');
    });
  });

  describe('Recent timestamps (within days)', () => {
    it('should show day format for 1 day ago', () => {
      const timestamp = new Date(realNow - 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('1日前');
    });

    it('should show day format for 3 days ago', () => {
      const timestamp = new Date(realNow - 3 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('3日前');
    });

    it('should show day format for 6 days ago', () => {
      const timestamp = new Date(realNow - 6 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('6日前');
    });
  });

  describe('Older timestamps (7+ days)', () => {
    it('should show full date format for 7 days ago', () => {
      const timestamp = new Date(realNow - 7 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      // Should contain date and time components in Japanese locale format
      expect(result).toMatch(/\d+月\d+日/);
      expect(result).toMatch(/\d{2}:\d{2}/);
    });

    it('should show full date format for 30 days ago', () => {
      const timestamp = new Date(realNow - 30 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toMatch(/\d+月\d+日/);
      expect(result).toMatch(/\d{2}:\d{2}/);
    });

    it('should show full date format for 1 year ago', () => {
      const timestamp = new Date(realNow - 365 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toMatch(/\d+月\d+日/);
      expect(result).toMatch(/\d{2}:\d{2}/);
    });
  });

  describe('Edge cases', () => {
    it('should handle future timestamps (negative diff)', () => {
      const timestamp = new Date(realNow + 60000).toISOString();
      const result = formatTimestamp(timestamp);
      // Future timestamps will have negative diff, resulting in large negative numbers
      // This will likely go to the formatted date path
      expect(result).toBeDefined();
    });

    it('should handle timestamp at exactly 1 hour boundary', () => {
      const timestamp = new Date(realNow - 3600000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('1時間前');
    });

    it('should handle timestamp at exactly 1 day boundary', () => {
      const timestamp = new Date(realNow - 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBe('1日前');
    });

    it('should handle timestamp at exactly 7 day boundary', () => {
      const timestamp = new Date(realNow - 7 * 86400000).toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toMatch(/\d+月\d+日/);
    });

    it('should handle invalid ISO date strings gracefully', () => {
      // Invalid date will result in NaN
      const timestamp = 'invalid-date';
      const result = formatTimestamp(timestamp);
      expect(result).toBeDefined();
      expect(typeof result).toBe('string');
    });

    it('should handle very old dates', () => {
      const timestamp = new Date('1970-01-01T00:00:00Z').toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toBeDefined();
      expect(result).toMatch(/\d+月\d+日/);
    });
  });

  describe('Japanese locale formatting', () => {
    it('should format date with Japanese month name for dates older than 7 days', () => {
      // Jan 8 (12:00) - 7 days = Jan 1
      const timestamp = new Date('2024-01-08T12:00:00Z').toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toMatch(/\d+月/);
      expect(result).toMatch(/\d+日/);
    });

    it('should include time in 24-hour format for older dates', () => {
      // Use a date that's 8+ days before the mocked time (2024-01-15T12:00:00Z)
      const timestamp = new Date('2024-01-06T14:30:00Z').toISOString();
      const result = formatTimestamp(timestamp);
      expect(result).toMatch(/\d{2}:\d{2}/);
    });
  });

  describe('Boundary conditions', () => {
    it('should transition from minutes to hours at 60 minutes', () => {
      // Just before 60 minutes
      const timestamp59 = new Date(realNow - 59 * 60000 - 59000).toISOString();
      const result59 = formatTimestamp(timestamp59);
      expect(result59).toMatch(/分前/);

      // Just after 60 minutes
      const timestamp60 = new Date(realNow - 60 * 60000).toISOString();
      const result60 = formatTimestamp(timestamp60);
      expect(result60).toMatch(/時間前/);
    });

    it('should transition from hours to days at 24 hours', () => {
      // Just before 24 hours
      const timestamp23 = new Date(realNow - 23 * 3600000 - 3599000).toISOString();
      const result23 = formatTimestamp(timestamp23);
      expect(result23).toMatch(/時間前/);

      // Just after 24 hours
      const timestamp24 = new Date(realNow - 24 * 3600000).toISOString();
      const result24 = formatTimestamp(timestamp24);
      expect(result24).toMatch(/日前/);
    });

    it('should transition from days to date at 7 days', () => {
      // Just before 7 days
      const timestamp6 = new Date(realNow - 6 * 86400000 - 3600000).toISOString();
      const result6 = formatTimestamp(timestamp6);
      expect(result6).toMatch(/日前/);

      // Just after 7 days
      const timestamp7 = new Date(realNow - 7 * 86400000).toISOString();
      const result7 = formatTimestamp(timestamp7);
      expect(result7).toMatch(/\d+月\d+日/);
    });
  });
});
