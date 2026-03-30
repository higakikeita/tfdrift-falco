/**
 * Root-level Error Boundary
 * Catches rendering errors throughout the application
 */

import { Component, type ReactNode, type ErrorInfo } from 'react';
import { logger } from '../utils/logger';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(): Partial<State> {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    logger.error('ErrorBoundary caught an error:', error);
    logger.error('Error details:', errorInfo.componentStack);

    this.setState({
      error,
      errorInfo,
    });
  }

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      const isDevelopment = process.env.NODE_ENV === 'development';

      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-100 px-4">
          <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
            <div className="flex items-center justify-center mb-4">
              <div className="text-4xl">⚠️</div>
            </div>

            <h1 className="text-2xl font-bold text-gray-900 mb-2 text-center">
              Something Went Wrong
            </h1>

            <p className="text-gray-600 text-sm mb-4 text-center">
              {isDevelopment
                ? 'An error occurred in the application. Check the console for more details.'
                : 'We encountered an unexpected error. Please try refreshing the page.'}
            </p>

            {isDevelopment && this.state.error && (
              <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-xs">
                <p className="font-mono text-red-800 break-words">
                  {this.state.error.message}
                </p>
                {this.state.errorInfo?.componentStack && (
                  <details className="mt-2">
                    <summary className="cursor-pointer text-red-700 font-semibold mb-1">
                      Stack Trace
                    </summary>
                    <pre className="text-red-700 text-xs whitespace-pre-wrap overflow-auto max-h-48">
                      {this.state.errorInfo.componentStack}
                    </pre>
                  </details>
                )}
              </div>
            )}

            <button
              onClick={this.handleReload}
              className="w-full bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded transition duration-200"
            >
              Reload Page
            </button>

            <p className="text-xs text-gray-500 text-center mt-4">
              If this problem persists, please contact support.
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
