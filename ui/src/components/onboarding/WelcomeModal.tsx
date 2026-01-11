/**
 * Welcome Modal - First-time user onboarding
 * Introduces key features and workflows
 */

import { useState } from 'react';
import { X, Zap, Network, Search, Target, FileImage, Keyboard } from 'lucide-react';

interface WelcomeModalProps {
  onClose: () => void;
}

const STORAGE_KEY = 'tfdrift-welcome-seen';

export const WelcomeModal: React.FC<WelcomeModalProps> = ({ onClose }) => {
  const [currentStep, setCurrentStep] = useState(0);

  const steps = [
    {
      title: 'TFDrift-Falcoへようこそ',
      icon: <Network className="w-16 h-16 text-blue-600" />,
      description: 'クラウドインフラのセキュリティとドリフト分析を可視化します',
      details: [
        'Terraform Drift → IAM → Kubernetes → Falcoの因果関係を追跡',
        '「なぜ」を可視化して、セキュリティインシデントの根本原因を特定',
        'インタラクティブなグラフで複雑な依存関係を理解'
      ]
    },
    {
      title: 'グラフの操作方法',
      icon: <Zap className="w-16 h-16 text-yellow-600" />,
      description: 'ノードをクリック、ダブルクリック、右クリックで操作',
      details: [
        '左クリック: ノード詳細パネルを開く',
        'ダブルクリック: フォーカスビューで選択ノードをハイライト',
        '右クリック: コンテキストメニューで依存関係・影響範囲を表示',
        'マウスホイール: ズームイン/アウト',
        'ドラッグ: グラフをパン移動'
      ]
    },
    {
      title: '依存関係の可視化',
      icon: <Target className="w-16 h-16 text-red-600" />,
      description: 'ノードの依存関係と影響範囲を追跡',
      details: [
        'ノード詳細パネルの「関係性」タブで依存先・依存元を確認',
        '「影響範囲」タブでインパクト半径を計算',
        '深さを指定して影響範囲を調整可能（1〜5ホップ）',
        '影響を受けるノードがグラフ上でハイライト表示'
      ]
    },
    {
      title: '検索とフィルタリング',
      icon: <Search className="w-16 h-16 text-green-600" />,
      description: '大規模なグラフから必要な情報を素早く抽出',
      details: [
        '左サイドバーの検索ボックスでノード名・タイプを検索',
        '深刻度フィルター: Critical、High、Medium、Lowで絞り込み',
        'リソースタイプフィルター: AWS、GCP、Kubernetesリソースで分類',
        'フィルター適用後もグラフの接続関係を維持'
      ]
    },
    {
      title: 'エクスポートと共有',
      icon: <FileImage className="w-16 h-16 text-purple-600" />,
      description: 'グラフを高品質な画像として保存',
      details: [
        '右上の「PNG」ボタン: 高解像度PNG画像（2400x1600）をエクスポート',
        '右上の「SVG」ボタン: スケーラブルなSVG画像をエクスポート',
        'レポートやドキュメントに挿入可能',
        '公式クラウドアイコンも含めてエクスポート'
      ]
    },
    {
      title: 'キーボードショートカット',
      icon: <Keyboard className="w-16 h-16 text-indigo-600" />,
      description: '効率的な操作のためのショートカット',
      details: [
        'F: グラフ全体を画面にフィット',
        'C: グラフを中央に配置',
        'ESC: 詳細パネルを閉じる',
        '+/-: ズームイン/アウト',
        '矢印キー: グラフをパン移動'
      ]
    }
  ];

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleFinish = () => {
    localStorage.setItem(STORAGE_KEY, 'true');
    onClose();
  };

  const handleSkip = () => {
    localStorage.setItem(STORAGE_KEY, 'true');
    onClose();
  };

  const step = steps[currentStep];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-2xl w-full mx-4 overflow-hidden animate-in zoom-in-95 duration-300">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-4 text-white">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold">{step.title}</h2>
              <p className="text-blue-100 text-sm mt-1">ステップ {currentStep + 1} / {steps.length}</p>
            </div>
            <button
              onClick={handleSkip}
              className="p-2 hover:bg-white/20 rounded-lg transition-colors"
              aria-label="閉じる"
            >
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="p-8">
          {/* Icon */}
          <div className="flex justify-center mb-6">
            <div className="p-4 bg-gray-50 dark:bg-gray-700 rounded-2xl">
              {step.icon}
            </div>
          </div>

          {/* Description */}
          <p className="text-center text-lg text-gray-700 dark:text-gray-300 mb-6">
            {step.description}
          </p>

          {/* Details */}
          <div className="bg-gray-50 dark:bg-gray-900 rounded-xl p-6 mb-6">
            <ul className="space-y-3">
              {step.details.map((detail, idx) => (
                <li key={idx} className="flex items-start gap-3 text-gray-700 dark:text-gray-300">
                  <div className="w-2 h-2 rounded-full bg-blue-600 mt-2 flex-shrink-0" />
                  <span className="text-sm leading-relaxed">{detail}</span>
                </li>
              ))}
            </ul>
          </div>

          {/* Progress Dots */}
          <div className="flex justify-center gap-2 mb-6">
            {steps.map((_, idx) => (
              <button
                key={idx}
                onClick={() => setCurrentStep(idx)}
                className={`w-2.5 h-2.5 rounded-full transition-all ${
                  idx === currentStep
                    ? 'bg-blue-600 w-8'
                    : idx < currentStep
                    ? 'bg-blue-300'
                    : 'bg-gray-300 dark:bg-gray-600'
                }`}
                aria-label={`ステップ ${idx + 1}に移動`}
              />
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="bg-gray-50 dark:bg-gray-900 px-8 py-4 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
          <button
            onClick={handleSkip}
            className="px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 font-medium transition-colors"
          >
            スキップ
          </button>
          <div className="flex gap-3">
            {currentStep > 0 && (
              <button
                onClick={handlePrev}
                className="px-6 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-lg font-medium transition-colors"
              >
                戻る
              </button>
            )}
            {currentStep < steps.length - 1 ? (
              <button
                onClick={handleNext}
                className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
              >
                次へ
              </button>
            ) : (
              <button
                onClick={handleFinish}
                className="px-6 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors"
              >
                始める
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Check if user should see welcome modal
 */
// eslint-disable-next-line react-refresh/only-export-components
export const shouldShowWelcome = (): boolean => {
  return !localStorage.getItem(STORAGE_KEY);
};

/**
 * Reset welcome modal (for testing)
 */
// eslint-disable-next-line react-refresh/only-export-components
export const resetWelcome = () => {
  localStorage.removeItem(STORAGE_KEY);
};
