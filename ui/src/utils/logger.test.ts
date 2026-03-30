/**
 * Tests for logger utility
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { logger } from './logger';

describe('Logger', () => {
  let consoleDebugSpy: ReturnType<typeof vi.spyOn>;
  let consoleInfoSpy: ReturnType<typeof vi.spyOn>;
  let consoleWarnSpy: ReturnType<typeof vi.spyOn>;
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
  const originalEnv = process.env.NODE_ENV;

  beforeEach(() => {
    consoleDebugSpy = vi.spyOn(console, 'debug').mockImplementation(() => {});
    consoleInfoSpy = vi.spyOn(console, 'info').mockImplementation(() => {});
    consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleDebugSpy.mockRestore();
    consoleInfoSpy.mockRestore();
    consoleWarnSpy.mockRestore();
    consoleErrorSpy.mockRestore();
    process.env.NODE_ENV = originalEnv;
  });

  describe('in development mode', () => {
    beforeEach(() => {
      process.env.NODE_ENV = 'development';
    });

    it('logs debug messages to console', () => {
      logger.debug('Debug message', { data: 'test' });

      expect(consoleDebugSpy).toHaveBeenCalledWith('[DEBUG] Debug message', { data: 'test' });
    });

    it('logs info messages to console', () => {
      logger.info('Info message', { data: 'test' });

      expect(consoleInfoSpy).toHaveBeenCalledWith('[INFO] Info message', { data: 'test' });
    });

    it('logs warn messages to console', () => {
      logger.warn('Warn message', { data: 'test' });

      expect(consoleWarnSpy).toHaveBeenCalledWith('[WARN] Warn message', { data: 'test' });
    });

    it('logs error messages to console', () => {
      logger.error('Error message', { data: 'test' });

      expect(consoleErrorSpy).toHaveBeenCalledWith('[ERROR] Error message', { data: 'test' });
    });

    it('logs without data parameter', () => {
      logger.info('Info message');

      expect(consoleInfoSpy).toHaveBeenCalledWith('[INFO] Info message', undefined);
    });
  });

  describe('in production mode', () => {
    beforeEach(() => {
      process.env.NODE_ENV = 'production';
    });

    it('suppresses debug messages', () => {
      logger.debug('Debug message');

      expect(consoleDebugSpy).not.toHaveBeenCalled();
    });

    it('suppresses info messages', () => {
      logger.info('Info message');

      expect(consoleInfoSpy).not.toHaveBeenCalled();
    });

    it('suppresses warn messages', () => {
      logger.warn('Warn message');

      expect(consoleWarnSpy).not.toHaveBeenCalled();
    });

    it('suppresses error messages from console', () => {
      logger.error('Error message');

      expect(consoleErrorSpy).not.toHaveBeenCalled();
    });
  });

  describe('message formatting', () => {
    beforeEach(() => {
      process.env.NODE_ENV = 'development';
    });

    it('prefixes messages with log level', () => {
      logger.debug('test');
      expect(consoleDebugSpy).toHaveBeenCalledWith('[DEBUG] test', undefined);

      logger.info('test');
      expect(consoleInfoSpy).toHaveBeenCalledWith('[INFO] test', undefined);

      logger.warn('test');
      expect(consoleWarnSpy).toHaveBeenCalledWith('[WARN] test', undefined);

      logger.error('test');
      expect(consoleErrorSpy).toHaveBeenCalledWith('[ERROR] test', undefined);
    });

    it('includes data parameter when provided', () => {
      const data = { userId: 123, action: 'login' };
      logger.info('User action', data);

      expect(consoleInfoSpy).toHaveBeenCalledWith('[INFO] User action', data);
    });
  });
});
