/**
 * LoadingSpinner Component
 * Simple spinning indicator with Tailwind classes
 */

import { memo } from 'react';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  text?: string;
}

const sizeClasses = {
  sm: 'w-4 h-4',
  md: 'w-8 h-8',
  lg: 'w-12 h-12',
};

const spinnerBorderClasses = {
  sm: 'border-2',
  md: 'border-2',
  lg: 'border-4',
};

function LoadingSpinnerComponent({ size = 'md', text }: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-3">
      <div className="flex items-center justify-center">
        <div
          className={`
            ${sizeClasses[size]}
            ${spinnerBorderClasses[size]}
            border-gray-300
            border-t-blue-500
            rounded-full
            animate-spin
          `}
        />
      </div>
      {text && (
        <p className="text-sm text-gray-600 dark:text-gray-400">{text}</p>
      )}
    </div>
  );
}

export const LoadingSpinner = memo(LoadingSpinnerComponent);
LoadingSpinner.displayName = 'LoadingSpinner';

export default LoadingSpinner;
