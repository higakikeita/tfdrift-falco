/**
 * NodeContextMenu - Right-click context menu for graph nodes
 * Provides quick actions for node operations
 */

import React from 'react';
import { ExternalLink, Eye, GitBranch, Target, Copy, Info } from 'lucide-react';

interface NodeContextMenuProps {
  position: { x: number; y: number };
  nodeId: string;
  nodeData: {
    label: string;
    type: string;
    resource_type: string;
  };
  onClose: () => void;
  onViewDetails?: () => void;
  onFocusView?: () => void;
  onShowDependencies?: () => void;
  onShowImpact?: () => void;
  onCopyId?: () => void;
}

export const NodeContextMenu: React.FC<NodeContextMenuProps> = ({
  position,
  nodeId,
  nodeData,
  onClose,
  onViewDetails,
  onFocusView,
  onShowDependencies,
  onShowImpact,
  onCopyId,
}) => {
  // Close menu when clicking outside
  React.useEffect(() => {
    const handleClickOutside = () => onClose();
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  }, [onClose]);

  const handleAction = (action: () => void) => {
    action();
    onClose();
  };

  return (
    <div
      className="fixed z-50 bg-white dark:bg-gray-800 rounded-lg shadow-2xl border border-gray-200 dark:border-gray-700 py-1 min-w-[220px]"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`,
      }}
      onClick={(e) => e.stopPropagation()}
    >
      {/* Header */}
      <div className="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
        <div className="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">
          {nodeData.label}
        </div>
        <div className="text-xs text-gray-500 dark:text-gray-400 truncate">
          {nodeData.resource_type}
        </div>
      </div>

      {/* Actions */}
      <div className="py-1">
        {onViewDetails && (
          <button
            onClick={() => handleAction(onViewDetails)}
            className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-blue-50 dark:hover:bg-blue-900/20 flex items-center gap-3 transition-colors"
          >
            <Info className="w-4 h-4" />
            <span>詳細を表示</span>
          </button>
        )}

        {onFocusView && (
          <button
            onClick={() => handleAction(onFocusView)}
            className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-purple-50 dark:hover:bg-purple-900/20 flex items-center gap-3 transition-colors"
          >
            <Eye className="w-4 h-4" />
            <span>フォーカスビュー</span>
          </button>
        )}

        {onShowDependencies && (
          <button
            onClick={() => handleAction(onShowDependencies)}
            className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-green-50 dark:hover:bg-green-900/20 flex items-center gap-3 transition-colors"
          >
            <GitBranch className="w-4 h-4" />
            <span>依存関係を表示</span>
          </button>
        )}

        {onShowImpact && (
          <button
            onClick={() => handleAction(onShowImpact)}
            className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-red-50 dark:hover:bg-red-900/20 flex items-center gap-3 transition-colors"
          >
            <Target className="w-4 h-4" />
            <span>影響範囲を表示</span>
          </button>
        )}

        <div className="my-1 border-t border-gray-200 dark:border-gray-700" />

        {onCopyId && (
          <button
            onClick={() => handleAction(onCopyId)}
            className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-3 transition-colors"
          >
            <Copy className="w-4 h-4" />
            <span>IDをコピー</span>
          </button>
        )}

        <a
          href={`#/node/${nodeId}`}
          className="w-full px-4 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-3 transition-colors"
          onClick={onClose}
        >
          <ExternalLink className="w-4 h-4" />
          <span>新しいタブで開く</span>
        </a>
      </div>
    </div>
  );
};
