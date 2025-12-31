/**
 * Keyboard Shortcuts Guide
 * Displays available keyboard shortcuts in a modal
 */

import { X, Keyboard } from 'lucide-react';

interface KeyboardShortcutsGuideProps {
  onClose: () => void;
}

interface Shortcut {
  key: string;
  description: string;
  category: string;
}

const shortcuts: Shortcut[] = [
  // Navigation
  { key: 'F', description: 'グラフ全体を画面にフィット', category: 'ナビゲーション' },
  { key: 'C', description: 'グラフを中央に配置', category: 'ナビゲーション' },
  { key: '+', description: 'ズームイン', category: 'ナビゲーション' },
  { key: '-', description: 'ズームアウト', category: 'ナビゲーション' },
  { key: '↑ ↓ ← →', description: 'グラフをパン移動', category: 'ナビゲーション' },

  // Selection & Interaction
  { key: 'Click', description: 'ノード詳細パネルを開く', category: '選択・操作' },
  { key: 'Double Click', description: 'フォーカスビューでハイライト', category: '選択・操作' },
  { key: 'Right Click', description: 'コンテキストメニューを表示', category: '選択・操作' },
  { key: 'ESC', description: '詳細パネルを閉じる', category: '選択・操作' },

  // View & Display
  { key: 'Ctrl/Cmd + S', description: 'グラフをPNG形式で保存', category: '表示・エクスポート' },
  { key: 'Ctrl/Cmd + E', description: 'グラフをSVG形式でエクスポート', category: '表示・エクスポート' },
  { key: 'Ctrl/Cmd + F', description: '検索ボックスにフォーカス', category: '表示・エクスポート' },

  // Help
  { key: '?', description: 'このヘルプを表示', category: 'ヘルプ' },
  { key: 'H', description: 'クイックヘルプを表示/非表示', category: 'ヘルプ' },
];

export const KeyboardShortcutsGuide: React.FC<KeyboardShortcutsGuideProps> = ({ onClose }) => {
  const categories = Array.from(new Set(shortcuts.map(s => s.category)));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-3xl w-full mx-4 max-h-[80vh] overflow-hidden animate-in zoom-in-95 duration-300">
        {/* Header */}
        <div className="bg-gradient-to-r from-indigo-600 to-purple-600 px-6 py-4 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Keyboard className="w-6 h-6" />
              <h2 className="text-2xl font-bold">キーボードショートカット</h2>
            </div>
            <button
              onClick={onClose}
              className="p-2 hover:bg-white/20 rounded-lg transition-colors"
              aria-label="閉じる"
            >
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(80vh-120px)]">
          {categories.map((category) => (
            <div key={category} className="mb-6 last:mb-0">
              <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-3 pb-2 border-b border-gray-200 dark:border-gray-700">
                {category}
              </h3>
              <div className="space-y-2">
                {shortcuts
                  .filter(s => s.category === category)
                  .map((shortcut, idx) => (
                    <div
                      key={idx}
                      className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                    >
                      <span className="text-gray-700 dark:text-gray-300 text-sm">
                        {shortcut.description}
                      </span>
                      <kbd className="px-3 py-1.5 bg-gray-100 dark:bg-gray-900 border border-gray-300 dark:border-gray-600 rounded-md text-xs font-mono font-semibold text-gray-800 dark:text-gray-200 shadow-sm">
                        {shortcut.key}
                      </kbd>
                    </div>
                  ))}
              </div>
            </div>
          ))}
        </div>

        {/* Footer */}
        <div className="bg-gray-50 dark:bg-gray-900 px-6 py-4 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              ヒント: <kbd className="px-2 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs">?</kbd> キーでいつでもこのガイドを表示できます
            </p>
            <button
              onClick={onClose}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-medium text-sm transition-colors"
            >
              閉じる
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
