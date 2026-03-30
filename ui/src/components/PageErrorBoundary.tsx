/**
 * Page-level Error Boundary
 * Wraps individual routes to isolate errors at the page level
 * Does not crash the entire application
 */

import { Component, type ReactNode, type ErrorInfo } from 'react';
import { logger } from '../utils/logger';

interface Props {
  children: ReactNode;
  pageName?: string;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class PageErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
    };
  }

  static getDerivedStateFromError(): Partial<State> {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const pageName = this.props.pageName || 'unknown page';
    logger.error(`Error in ${pageName}:`, error);
    logger.error('Component stack:', errorInfo.componentStack);

    this.setState({
      error,
    });

    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
    });
  };

  render() {
    if (this.state.hasError) {
      const pageName = this.props.pageName || 'this section';

      return (
        <div className="flex items-center justify-center min-h-96 bg-gray-50 border border-gray-200 rounded-lg p-6">
          <div className="text-center max-w-md">
            <div className="text-5xl mb-4">⚠️</div>

            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Something Went Wrong
            </h2>

            <p className="text-gray-600 text-sm mb-4">
              We encountered an error in {pageName}. The rest of the application
              is still working properly.
            </p>

            {process.env.NODE_ENV === 'development' && this.state.error && (
              <div className="mb-4 p-2 bg-red-50 border border-red-200 rounded text-xs text-left">
                <p className="font-mono text-red-800 break-words">
                  {this.state.error.message}
                </p>
              </div>
            )}

            <button
              onClick={this.handleRetry}
              className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition duration-200"
            >
              Try Again
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default PageErrorBoundary;
