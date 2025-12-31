/**
 * Help Overlay - Contextual help and quick tips
 * Floats above the graph with helpful information
 */

import { useState } from 'react';
import { HelpCircle, X, ChevronDown, ChevronUp, Lightbulb, Zap, Target } from 'lucide-react';

interface HelpOverlayProps {
  onOpenShortcuts?: () => void;
  onOpenWelcome?: () => void;
}

export const HelpOverlay: React.FC<HelpOverlayProps> = ({
  onOpenShortcuts,
  onOpenWelcome
}) => {
  const [isExpanded, setIsExpanded] = useState(true);
  const [isVisible, setIsVisible] = useState(true);

  if (!isVisible) {
    return (
      <button
        onClick={() => setIsVisible(true)}
        className="fixed bottom-6 right-6 p-3 bg-blue-600 hover:bg-blue-700 text-white rounded-full shadow-lg transition-all hover:scale-110 z-40"
        aria-label="ヘルプを表示"
      >
        <HelpCircle className="w-6 h-6" />
      </button>
    );
  }

  return (
    <div className="fixed bottom-6 right-6 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 max-w-sm z-40 animate-in slide-in-from-bottom duration-300">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-4 py-3 rounded-t-xl text-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Lightbulb className="w-5 h-5" />
            <h3 className="font-semibold text-sm">クイックヘルプ</h3>
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="p-1 hover:bg-white/20 rounded transition-colors"
              aria-label={isExpanded ? '折りたたむ' : '展開する'}
            >
              {isExpanded ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronUp className="w-4 h-4" />
              )}
            </button>
            <button
              onClick={() => setIsVisible(false)}
              className="p-1 hover:bg-white/20 rounded transition-colors"
              aria-label="閉じる"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      {isExpanded && (
        <div className="p-4 space-y-4">
          {/* Quick Tips */}
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide flex items-center gap-2">
              <Zap className="w-3 h-3" />
              クイックヒント
            </h4>
            <ul className="space-y-2 text-xs text-gray-700 dark:text-gray-300">
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-600 mt-1.5 flex-shrink-0" />
                <span>ノードをクリックで詳細を表示</span>
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-600 mt-1.5 flex-shrink-0" />
                <span>ダブルクリックでフォーカスビュー</span>
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-600 mt-1.5 flex-shrink-0" />
                <span>右クリックで依存関係を表示</span>
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-600 mt-1.5 flex-shrink-0" />
                <span>マウスホイールでズーム操作</span>
              </li>
              <li className="flex items-start gap-2">
                <div className="w-1.5 h-1.5 rounded-full bg-blue-600 mt-1.5 flex-shrink-0" />
                <span>ドラッグでグラフを移動</span>
              </li>
            </ul>
          </div>

          {/* Key Features */}
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide flex items-center gap-2">
              <Target className="w-3 h-3" />
              主な機能
            </h4>
            <div className="space-y-1.5">
              <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg px-3 py-2">
                <p className="text-xs font-medium text-blue-900 dark:text-blue-300">影響範囲分析</p>
                <p className="text-xs text-blue-700 dark:text-blue-400 mt-0.5">
                  詳細パネルの「影響範囲」タブで確認
                </p>
              </div>
              <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg px-3 py-2">
                <p className="text-xs font-medium text-green-900 dark:text-green-300">依存関係追跡</p>
                <p className="text-xs text-green-700 dark:text-green-400 mt-0.5">
                  「関係性」タブで依存先・依存元を表示
                </p>
              </div>
              <div className="bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800 rounded-lg px-3 py-2">
                <p className="text-xs font-medium text-purple-900 dark:text-purple-300">検索・フィルター</p>
                <p className="text-xs text-purple-700 dark:text-purple-400 mt-0.5">
                  左サイドバーで深刻度・タイプで絞り込み
                </p>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="space-y-2 pt-2 border-t border-gray-200 dark:border-gray-700">
            {onOpenShortcuts && (
              <button
                onClick={onOpenShortcuts}
                className="w-full px-3 py-2 text-xs font-medium text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 hover:bg-blue-100 dark:hover:bg-blue-900/50 rounded-lg transition-colors"
              >
                ⌨️ キーボードショートカット
              </button>
            )}
            {onOpenWelcome && (
              <button
                onClick={onOpenWelcome}
                className="w-full px-3 py-2 text-xs font-medium text-indigo-700 dark:text-indigo-300 bg-indigo-50 dark:bg-indigo-900/30 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 rounded-lg transition-colors"
              >
                🎯 チュートリアルを再表示
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
