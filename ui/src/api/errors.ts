/**
 * Custom error types for API interactions
 */

/**
 * Base application error class
 */
export class AppError extends Error {
  declare name: string;

  constructor(message: string) {
    super(message);
    this.name = 'AppError';
    Object.setPrototypeOf(this, AppError.prototype);
  }
}

/**
 * Network errors (connection failures, timeouts, etc.)
 */
export class NetworkError extends AppError {
  declare name: string;
  declare originalError: Error | undefined;

  constructor(message: string, originalError?: Error) {
    super(message);
    this.name = 'NetworkError';
    this.originalError = originalError;
    Object.setPrototypeOf(this, NetworkError.prototype);
  }
}

/**
 * API errors with HTTP status codes
 */
export class ApiError extends AppError {
  declare name: string;
  declare statusCode: number;
  declare response: unknown;

  constructor(
    message: string,
    statusCode: number,
    response?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
    this.statusCode = statusCode;
    this.response = response;
    Object.setPrototypeOf(this, ApiError.prototype);
  }
}

/**
 * 404 Not Found errors
 */
export class NotFoundError extends ApiError {
  declare name: string;

  constructor(message: string, response?: unknown) {
    super(message, 404, response);
    this.name = 'NotFoundError';
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * Validation errors (invalid response format, missing fields, etc.)
 */
export class ValidationError extends AppError {
  declare name: string;
  declare validationDetails: Record<string, unknown> | undefined;

  constructor(
    message: string,
    validationDetails?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'ValidationError';
    this.validationDetails = validationDetails;
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
}

/**
 * Type guard to check if an error is an AppError
 */
export function isAppError(error: unknown): error is AppError {
  return error instanceof AppError;
}

/**
 * Type guard to check if an error is a NetworkError
 */
export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof NetworkError;
}

/**
 * Type guard to check if an error is an ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

/**
 * Type guard to check if an error is a NotFoundError
 */
export function isNotFoundError(error: unknown): error is NotFoundError {
  return error instanceof NotFoundError;
}

/**
 * Type guard to check if an error is a ValidationError
 */
export function isValidationError(error: unknown): error is ValidationError {
  return error instanceof ValidationError;
}

/**
 * Convert a generic error to an AppError-based error
 */
export function toAppError(error: unknown): AppError {
  if (error instanceof AppError) {
    return error;
  }

  if (error instanceof Error) {
    return new AppError(error.message);
  }

  return new AppError(String(error));
}
