/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * PatternSearchPanel - Neo4j-style pattern matching search UI
 */

import React, { useState } from 'react';
import { X, Search, Loader2 } from 'lucide-react';
import { usePatternMatch } from '../api/hooks';

interface PatternSearchPanelProps {
  onClose: () => void;
  onNodeSelect?: (nodeId: string) => void;
}

interface Node {
  id: string;
  labels: string[];
  properties: Record<string, any>;
}

const PatternSearchPanel = ({ onClose, onNodeSelect }: PatternSearchPanelProps) => {
  const [startLabels, setStartLabels] = useState('');
  const [relType, setRelType] = useState('');
  const [endLabels, setEndLabels] = useState('');
  const [endFilter, setEndFilter] = useState('');
  const [searchEnabled, setSearchEnabled] = useState(false);

  // Build pattern object
  const pattern = searchEnabled
    ? {
        start_labels: startLabels.split(',').map((s) => s.trim()).filter(Boolean),
        rel_type: relType.trim(),
        end_labels: endLabels.split(',').map((s) => s.trim()).filter(Boolean),
        end_filter: endFilter.trim()
          ? JSON.parse(endFilter)
          : {},
      }
    : null;

  const { data, isLoading, error } = usePatternMatch(pattern, searchEnabled);

  const matches: Array<Node[]> = (data as any)?.data?.matches || [];

  const handleSearch = () => {
    setSearchEnabled(true);
  };

  const handleClear = () => {
    setSearchEnabled(false);
    setStartLabels('');
    setRelType('');
    setEndLabels('');
    setEndFilter('');
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-3xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-indigo-600 text-white flex items-center justify-between rounded-t-lg">
          <div className="flex items-center gap-2">
            <Search className="w-5 h-5" />
            <h3 className="font-semibold text-lg">パターンマッチング検索</h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-white/20 rounded transition-colors"
            aria-label="閉じる"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Search Form */}
          <div className="space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                開始ノードラベル (カンマ区切り)
              </label>
              <input
                type="text"
                value={startLabels}
                onChange={(e) => setStartLabels(e.target.value)}
                placeholder="例: EC2, Compute"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                関係タイプ (空欄で全て)
              </label>
              <input
                type="text"
                value={relType}
                onChange={(e) => setRelType(e.target.value)}
                placeholder="例: DEPENDS_ON, PART_OF"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                終了ノードラベル (カンマ区切り)
              </label>
              <input
                type="text"
                value={endLabels}
                onChange={(e) => setEndLabels(e.target.value)}
                placeholder="例: Subnet, Network"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                終了ノードフィルタ (JSON形式)
              </label>
              <textarea
                value={endFilter}
                onChange={(e) => setEndFilter(e.target.value)}
                placeholder='例: {"id": "subnet-123"}'
                rows={2}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 font-mono text-sm"
              />
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleSearch}
                disabled={isLoading}
                className="flex-1 px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-purple-400 text-white rounded font-medium transition-colors flex items-center justify-center gap-2"
              >
                {isLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    検索中...
                  </>
                ) : (
                  <>
                    <Search className="w-4 h-4" />
                    検索
                  </>
                )}
              </button>
              <button
                onClick={handleClear}
                className="px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded font-medium transition-colors"
              >
                クリア
              </button>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/20 border border-red-300 dark:border-red-700 rounded text-red-800 dark:text-red-200 text-sm">
              エラー: {(error as Error).message}
            </div>
          )}

          {/* Results */}
          {searchEnabled && matches.length === 0 && !isLoading && (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400">
              マッチする結果が見つかりませんでした
            </div>
          )}

          {matches.length > 0 && (
            <div className="space-y-3">
              <h4 className="font-semibold text-gray-700 dark:text-gray-300">
                検索結果: {matches.length} 件
              </h4>
              {matches.map((match, idx) => (
                <div
                  key={idx}
                  className="p-4 bg-gray-50 dark:bg-gray-700 rounded border border-gray-200 dark:border-gray-600"
                >
                  <div className="flex items-center gap-3">
                    {match.map((node, nodeIdx) => (
                      <React.Fragment key={node.id}>
                        <button
                          onClick={() => onNodeSelect && onNodeSelect(node.id)}
                          className="flex-1 text-left px-3 py-2 bg-white dark:bg-gray-800 rounded border border-gray-300 dark:border-gray-600 hover:border-purple-500 dark:hover:border-purple-400 transition-colors"
                        >
                          <div className="font-medium text-gray-900 dark:text-gray-100 text-sm">
                            {node.properties.name || node.id}
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 font-mono">
                            {node.properties.type}
                          </div>
                        </button>
                        {nodeIdx < match.length - 1 && (
                          <div className="text-gray-400 dark:text-gray-500">→</div>
                        )}
                      </React.Fragment>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PatternSearchPanel;
