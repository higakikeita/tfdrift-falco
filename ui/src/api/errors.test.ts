/**
 * Tests for error types and utilities
 */

import { describe, it, expect } from 'vitest';
import {
  AppError,
  NetworkError,
  ApiError,
  NotFoundError,
  ValidationError,
  isAppError,
  isNetworkError,
  isApiError,
  isNotFoundError,
  isValidationError,
  toAppError,
} from './errors';

describe('Error Types', () => {
  describe('AppError', () => {
    it('creates an AppError with message', () => {
      const error = new AppError('Test error');
      expect(error.message).toBe('Test error');
      expect(error.name).toBe('AppError');
      expect(error).toBeInstanceOf(Error);
    });
  });

  describe('NetworkError', () => {
    it('creates a NetworkError with message and optional original error', () => {
      const originalError = new Error('Connection failed');
      const error = new NetworkError('Network error', originalError);

      expect(error.message).toBe('Network error');
      expect(error.name).toBe('NetworkError');
      expect(error.originalError).toBe(originalError);
      expect(error).toBeInstanceOf(AppError);
    });

    it('can be created without original error', () => {
      const error = new NetworkError('Network error');
      expect(error.originalError).toBeUndefined();
    });
  });

  describe('ApiError', () => {
    it('creates an ApiError with message and status code', () => {
      const error = new ApiError('API failed', 500);

      expect(error.message).toBe('API failed');
      expect(error.statusCode).toBe(500);
      expect(error.name).toBe('ApiError');
      expect(error).toBeInstanceOf(AppError);
    });

    it('can include response data', () => {
      const response = { error: 'Server error' };
      const error = new ApiError('API failed', 500, response);

      expect(error.response).toBe(response);
    });
  });

  describe('NotFoundError', () => {
    it('creates a NotFoundError with 404 status code', () => {
      const error = new NotFoundError('Resource not found');

      expect(error.message).toBe('Resource not found');
      expect(error.statusCode).toBe(404);
      expect(error.name).toBe('NotFoundError');
      expect(error).toBeInstanceOf(ApiError);
    });
  });

  describe('ValidationError', () => {
    it('creates a ValidationError with message', () => {
      const error = new ValidationError('Invalid data');

      expect(error.message).toBe('Invalid data');
      expect(error.name).toBe('ValidationError');
      expect(error).toBeInstanceOf(AppError);
    });

    it('can include validation details', () => {
      const details = { field: 'email', reason: 'Invalid format' };
      const error = new ValidationError('Invalid data', details);

      expect(error.validationDetails).toBe(details);
    });
  });
});

describe('Type Guards', () => {
  describe('isAppError', () => {
    it('returns true for AppError instances', () => {
      const error = new AppError('Test');
      expect(isAppError(error)).toBe(true);
    });

    it('returns true for subclasses of AppError', () => {
      const networkError = new NetworkError('Test');
      const apiError = new ApiError('Test', 500);
      const notFoundError = new NotFoundError('Test');
      const validationError = new ValidationError('Test');

      expect(isAppError(networkError)).toBe(true);
      expect(isAppError(apiError)).toBe(true);
      expect(isAppError(notFoundError)).toBe(true);
      expect(isAppError(validationError)).toBe(true);
    });

    it('returns false for regular Error', () => {
      const error = new Error('Test');
      expect(isAppError(error)).toBe(false);
    });
  });

  describe('isNetworkError', () => {
    it('returns true for NetworkError instances', () => {
      const error = new NetworkError('Test');
      expect(isNetworkError(error)).toBe(true);
    });

    it('returns false for other error types', () => {
      const appError = new AppError('Test');
      const apiError = new ApiError('Test', 500);

      expect(isNetworkError(appError)).toBe(false);
      expect(isNetworkError(apiError)).toBe(false);
    });
  });

  describe('isApiError', () => {
    it('returns true for ApiError instances', () => {
      const error = new ApiError('Test', 500);
      expect(isApiError(error)).toBe(true);
    });

    it('returns true for NotFoundError', () => {
      const error = new NotFoundError('Test');
      expect(isApiError(error)).toBe(true);
    });

    it('returns false for other error types', () => {
      const appError = new AppError('Test');
      const networkError = new NetworkError('Test');

      expect(isApiError(appError)).toBe(false);
      expect(isApiError(networkError)).toBe(false);
    });
  });

  describe('isNotFoundError', () => {
    it('returns true for NotFoundError instances', () => {
      const error = new NotFoundError('Test');
      expect(isNotFoundError(error)).toBe(true);
    });

    it('returns false for other error types', () => {
      const apiError = new ApiError('Test', 500);
      const appError = new AppError('Test');

      expect(isNotFoundError(apiError)).toBe(false);
      expect(isNotFoundError(appError)).toBe(false);
    });
  });

  describe('isValidationError', () => {
    it('returns true for ValidationError instances', () => {
      const error = new ValidationError('Test');
      expect(isValidationError(error)).toBe(true);
    });

    it('returns false for other error types', () => {
      const appError = new AppError('Test');
      const apiError = new ApiError('Test', 500);

      expect(isValidationError(appError)).toBe(false);
      expect(isValidationError(apiError)).toBe(false);
    });
  });
});

describe('toAppError', () => {
  it('returns the same AppError if already an AppError', () => {
    const error = new AppError('Test');
    expect(toAppError(error)).toBe(error);
  });

  it('wraps regular Error in AppError', () => {
    const error = new Error('Test message');
    const appError = toAppError(error);

    expect(appError).toBeInstanceOf(AppError);
    expect(appError.message).toBe('Test message');
  });

  it('converts string to AppError', () => {
    const appError = toAppError('Test error message');

    expect(appError).toBeInstanceOf(AppError);
    expect(appError.message).toBe('Test error message');
  });

  it('handles unknown error types', () => {
    const appError = toAppError(123);

    expect(appError).toBeInstanceOf(AppError);
    expect(appError.message).toBe('123');
  });
});
