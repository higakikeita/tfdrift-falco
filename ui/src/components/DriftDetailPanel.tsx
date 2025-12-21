/**
 * Drift Detail Panel Component
 * 選択されたドリフトイベントの詳細情報を表示
 */

import type { DriftEvent } from '../types/drift';
import { SiAmazon, SiGooglecloud } from 'react-icons/si';

interface DriftDetailPanelProps {
  drift: DriftEvent | null;
  onClose?: () => void;
}

const severityColors = {
  critical: 'bg-red-100 text-red-800 border-red-200',
  high: 'bg-orange-100 text-orange-800 border-orange-200',
  medium: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  low: 'bg-blue-100 text-blue-800 border-blue-200',
};

export default function DriftDetailPanel({ drift, onClose }: DriftDetailPanelProps) {
  if (!drift) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">
        <div className="text-center">
          <p className="text-lg font-medium mb-2">ドリフトイベントを選択してください</p>
          <p className="text-sm">詳細情報を表示します</p>
        </div>
      </div>
    );
  }

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const formatValue = (value: string | null) => {
    if (value === null) return <span className="text-gray-400 dark:text-gray-500 italic">null</span>;
    try {
      const parsed = JSON.parse(value);
      return (
        <pre className="text-xs bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 p-2 rounded overflow-auto max-h-40">
          {JSON.stringify(parsed, null, 2)}
        </pre>
      );
    } catch {
      return <code className="text-sm text-gray-900 dark:text-gray-100">{value}</code>;
    }
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900 border-l border-gray-200 dark:border-gray-700">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2 mb-2">
              {drift.provider === 'aws' && <SiAmazon size={24} className="text-orange-500" />}
              {drift.provider === 'gcp' && <SiGooglecloud size={24} className="text-blue-500" />}
              <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">{drift.resourceName || drift.resourceId}</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-300 font-mono">{drift.resourceType}</p>
          </div>
          {onClose && (
            <button
              onClick={onClose}
              className="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
            >
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto px-6 py-4 space-y-6">
        {/* Severity & Status */}
        <div>
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">ステータス</h4>
          <div className="flex items-center gap-3">
            <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${severityColors[drift.severity]}`}>
              {drift.severity.toUpperCase()}
            </span>
            <span className="text-sm text-gray-600 dark:text-gray-300">
              {drift.changeType === 'created' && '🆕 作成'}
              {drift.changeType === 'modified' && '✏️ 変更'}
              {drift.changeType === 'deleted' && '🗑️ 削除'}
            </span>
          </div>
        </div>

        {/* Timestamp */}
        <div>
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">検知時刻</h4>
          <div className="text-sm text-gray-900 dark:text-gray-100">{formatTimestamp(drift.timestamp)}</div>
        </div>

        {/* Resource Info */}
        <div>
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">リソース情報</h4>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">ID:</span>
              <code className="text-xs bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-2 py-1 rounded">{drift.resourceId}</code>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">リージョン:</span>
              <span className="font-mono text-gray-900 dark:text-gray-100">{drift.region}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">プロバイダー:</span>
              <span className="font-medium text-gray-900 dark:text-gray-100">{drift.provider.toUpperCase()}</span>
            </div>
          </div>
        </div>

        {/* Change Details */}
        <div>
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">変更内容</h4>
          <div className="space-y-3">
            <div>
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">属性</div>
              <code className="text-sm bg-blue-50 dark:bg-blue-900/20 text-blue-900 dark:text-blue-200 px-2 py-1 rounded border border-blue-200 dark:border-blue-800">
                {drift.attribute}
              </code>
            </div>
            <div>
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">変更前 (Old Value)</div>
              <div className="bg-red-50 dark:bg-red-900/20 p-3 rounded border border-red-200 dark:border-red-800">
                {formatValue(drift.oldValue)}
              </div>
            </div>
            <div>
              <div className="text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">変更後 (New Value)</div>
              <div className="bg-green-50 dark:bg-green-900/20 p-3 rounded border border-green-200 dark:border-green-800">
                {formatValue(drift.newValue)}
              </div>
            </div>
          </div>
        </div>

        {/* User Identity */}
        <div>
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">実行ユーザー</h4>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">ユーザー名:</span>
              <span className="font-medium text-gray-900 dark:text-gray-100">{drift.userIdentity.userName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600 dark:text-gray-400">タイプ:</span>
              <span className="text-gray-900 dark:text-gray-100">{drift.userIdentity.type}</span>
            </div>
            {drift.userIdentity.arn && (
              <div>
                <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">ARN:</div>
                <code className="text-xs bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-2 py-1 rounded block overflow-auto">
                  {drift.userIdentity.arn}
                </code>
              </div>
            )}
            {drift.userIdentity.accountId && (
              <div className="flex justify-between">
                <span className="text-gray-600 dark:text-gray-400">Account ID:</span>
                <code className="text-xs text-gray-900 dark:text-gray-100">{drift.userIdentity.accountId}</code>
              </div>
            )}
          </div>
        </div>

        {/* CloudTrail Info */}
        {drift.cloudtrailEventId && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">CloudTrail / Audit Log</h4>
            <div className="space-y-2 text-sm">
              <div>
                <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">Event ID:</div>
                <code className="text-xs bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-2 py-1 rounded block overflow-auto">
                  {drift.cloudtrailEventId}
                </code>
              </div>
              {drift.cloudtrailEventName && (
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Event Name:</span>
                  <code className="text-xs text-gray-900 dark:text-gray-100">{drift.cloudtrailEventName}</code>
                </div>
              )}
              {drift.sourceIP && (
                <div className="flex justify-between">
                  <span className="text-gray-600 dark:text-gray-400">Source IP:</span>
                  <code className="text-xs text-gray-900 dark:text-gray-100">{drift.sourceIP}</code>
                </div>
              )}
              {drift.userAgent && (
                <div>
                  <div className="text-xs text-gray-600 dark:text-gray-400 mb-1">User Agent:</div>
                  <code className="text-xs bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-2 py-1 rounded block overflow-auto">
                    {drift.userAgent}
                  </code>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Tags */}
        {drift.tags && Object.keys(drift.tags).length > 0 && (
          <div>
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">タグ</h4>
            <div className="flex flex-wrap gap-2">
              {Object.entries(drift.tags).map(([key, value]) => (
                <span
                  key={key}
                  className="inline-flex items-center px-2 py-1 rounded-full text-xs bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300"
                >
                  <span className="font-medium">{key}:</span>
                  <span className="ml-1">{value}</span>
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">推奨アクション</h4>
          <div className="space-y-2 text-sm text-gray-700 dark:text-gray-300">
            <div className="flex items-start gap-2">
              <span className="text-blue-600 dark:text-blue-400 mt-1">1.</span>
              <span>変更内容をユーザー ({drift.userIdentity.userName}) に確認する</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-blue-600 dark:text-blue-400 mt-1">2.</span>
              <span>意図的な変更の場合、Terraformコードを更新する</span>
            </div>
            <div className="flex items-start gap-2">
              <span className="text-blue-600 dark:text-blue-400 mt-1">3.</span>
              <span>不正な変更の場合、<code className="bg-gray-100 dark:bg-gray-800 px-1 rounded">terraform plan</code> でStateを同期する</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
