/**
 * Progressive Loader Component
 * Shows loading progress for large graph rendering
 */

import { memo } from 'react';
import { Loader2 } from 'lucide-react';

interface ProgressiveLoaderProps {
  progress: number; // 0-100
  currentBatch: number;
  totalBatches: number;
  isLoading: boolean;
  onSkip?: () => void;
  onCancel?: () => void;
}

export const ProgressiveLoader = memo(({
  progress,
  currentBatch,
  totalBatches,
  isLoading,
  onSkip,
  onCancel,
}: ProgressiveLoaderProps) => {
  if (!isLoading) return null;

  return (
    <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-white rounded-xl shadow-2xl px-6 py-4 min-w-[320px] border border-gray-200">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Loader2 className="w-5 h-5 text-blue-500 animate-spin" />
          <span className="text-sm font-semibold text-gray-900">
            Loading Graph
          </span>
        </div>
        <span className="text-xs font-medium text-gray-500">
          {currentBatch} / {totalBatches}
        </span>
      </div>

      {/* Progress Bar */}
      <div className="mb-3">
        <div className="w-full h-3 bg-gray-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-blue-500 to-blue-600 transition-all duration-300 ease-out"
            style={{ width: `${progress}%` }}
          />
        </div>
        <div className="flex justify-between mt-1.5">
          <span className="text-xs text-gray-500">
            {progress}% complete
          </span>
          <span className="text-xs font-medium text-blue-600">
            {totalBatches - currentBatch} batches remaining
          </span>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        {onSkip && (
          <button
            onClick={onSkip}
            className="flex-1 px-3 py-1.5 text-xs font-medium bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            Load All Now
          </button>
        )}
        {onCancel && (
          <button
            onClick={onCancel}
            className="px-3 py-1.5 text-xs font-medium bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
          >
            Cancel
          </button>
        )}
      </div>

      {/* Info */}
      <div className="mt-3 pt-3 border-t border-gray-100">
        <p className="text-[10px] text-gray-400 text-center">
          Loading in batches for optimal performance
        </p>
      </div>
    </div>
  );
});

ProgressiveLoader.displayName = 'ProgressiveLoader';

/**
 * Compact version for inline display
 */
export const CompactProgressiveLoader = memo(({
  progress,
  onSkip,
}: {
  progress: number;
  onSkip?: () => void;
}) => {
  return (
    <div className="inline-flex items-center gap-2 bg-blue-50 rounded-lg px-3 py-1.5">
      <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />
      <span className="text-xs font-medium text-blue-700">
        Loading {progress}%
      </span>
      {onSkip && (
        <button
          onClick={onSkip}
          className="text-xs text-blue-600 hover:text-blue-800 font-medium underline"
        >
          Skip
        </button>
      )}
    </div>
  );
});

CompactProgressiveLoader.displayName = 'CompactProgressiveLoader';

/**
 * Skeleton loader for graph placeholder
 */
export const GraphSkeleton = memo(() => {
  return (
    <div className="w-full h-full bg-gray-50 animate-pulse flex items-center justify-center">
      <div className="text-center">
        <Loader2 className="w-12 h-12 text-gray-400 animate-spin mx-auto mb-4" />
        <div className="text-sm font-medium text-gray-500">
          Preparing graph...
        </div>
        <div className="text-xs text-gray-400 mt-1">
          This may take a moment for large graphs
        </div>
      </div>
    </div>
  );
});

GraphSkeleton.displayName = 'GraphSkeleton';
