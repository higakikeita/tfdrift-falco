/**
 * NodeTooltip - リッチなホバーツールチップコンポーネント
 * ノードにホバーした時に詳細情報を表示
 */

import { memo } from 'react';
import { ExternalLink, Clock, User, AlertTriangle, CheckCircle } from 'lucide-react';

export interface NodeTooltipProps {
  data: {
    id: string;
    label: string;
    type: string;
    resourceType: string;
    resourceName: string;
    severity: 'low' | 'medium' | 'high' | 'critical';
    metadata?: {
      mode?: string;
      provider?: string;
      tf_name?: string;
      has_drift?: boolean;
      last_modified?: string;
      user?: string;
      drift_count?: number;
    };
  };
  position: { x: number; y: number };
}

const severityConfig = {
  low: {
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-50 dark:bg-green-900/20',
    borderColor: 'border-green-200 dark:border-green-800',
    icon: CheckCircle,
    label: '正常',
  },
  medium: {
    color: 'text-yellow-600 dark:text-yellow-400',
    bgColor: 'bg-yellow-50 dark:bg-yellow-900/20',
    borderColor: 'border-yellow-200 dark:border-yellow-800',
    icon: AlertTriangle,
    label: '警告',
  },
  high: {
    color: 'text-orange-600 dark:text-orange-400',
    bgColor: 'bg-orange-50 dark:bg-orange-900/20',
    borderColor: 'border-orange-200 dark:border-orange-800',
    icon: AlertTriangle,
    label: 'ドリフト検出',
  },
  critical: {
    color: 'text-red-600 dark:text-red-400',
    bgColor: 'bg-red-50 dark:bg-red-900/20',
    borderColor: 'border-red-200 dark:border-red-800',
    icon: AlertTriangle,
    label: '重大なドリフト',
  },
};

export const NodeTooltip = memo(({ data, position }: NodeTooltipProps) => {
  const config = severityConfig[data.severity] || severityConfig.low;
  const StatusIcon = config.icon;
  const hasDrift = data.metadata?.has_drift;
  const driftCount = data.metadata?.drift_count || 0;

  return (
    <div
      className="absolute z-50 pointer-events-none"
      style={{
        left: `${position.x + 20}px`,
        top: `${position.y}px`,
        maxWidth: '360px',
      }}
    >
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden animate-in fade-in slide-in-from-left-2 duration-200">
        {/* Header */}
        <div className={`px-4 py-3 ${config.bgColor} border-b ${config.borderColor}`}>
          <div className="flex items-start justify-between gap-2">
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-bold text-gray-900 dark:text-gray-100 truncate">
                {data.resourceName || data.label}
              </h3>
              <p className="text-xs text-gray-600 dark:text-gray-400 font-mono mt-0.5">
                {data.resourceType}
              </p>
            </div>
            <StatusIcon className={`w-5 h-5 ${config.color} flex-shrink-0`} />
          </div>
        </div>

        {/* Content */}
        <div className="px-4 py-3 space-y-2.5">
          {/* Status */}
          <div className="flex items-center justify-between text-xs">
            <span className="text-gray-600 dark:text-gray-400">ステータス:</span>
            <span className={`font-medium ${config.color}`}>{config.label}</span>
          </div>

          {/* Drift Info */}
          {hasDrift && (
            <div className="flex items-center justify-between text-xs">
              <span className="text-gray-600 dark:text-gray-400">ドリフト:</span>
              <span className="font-medium text-orange-600 dark:text-orange-400">
                {driftCount}件の変更を検出
              </span>
            </div>
          )}

          {/* Last Modified */}
          {data.metadata?.last_modified && (
            <div className="flex items-center gap-2 text-xs">
              <Clock className="w-3.5 h-3.5 text-gray-400" />
              <span className="text-gray-600 dark:text-gray-400">最終更新:</span>
              <span className="text-gray-900 dark:text-gray-100">
                {formatTimestamp(data.metadata.last_modified)}
              </span>
            </div>
          )}

          {/* User */}
          {data.metadata?.user && (
            <div className="flex items-center gap-2 text-xs">
              <User className="w-3.5 h-3.5 text-gray-400" />
              <span className="text-gray-600 dark:text-gray-400">変更者:</span>
              <span className="text-gray-900 dark:text-gray-100">{data.metadata.user}</span>
            </div>
          )}

          {/* Provider */}
          {data.metadata?.provider && (
            <div className="flex items-center justify-between text-xs">
              <span className="text-gray-600 dark:text-gray-400">プロバイダー:</span>
              <span className="text-gray-900 dark:text-gray-100 font-mono text-[11px]">
                {data.metadata.provider}
              </span>
            </div>
          )}

          {/* Resource ID */}
          <div className="pt-2 border-t border-gray-100 dark:border-gray-700">
            <div className="text-[10px] text-gray-500 dark:text-gray-500 font-mono break-all">
              ID: {data.id}
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="px-4 py-2.5 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-100 dark:border-gray-700">
          <div className="flex items-center gap-2 text-xs">
            <span className="text-gray-500 dark:text-gray-400">
              クリックで詳細を表示
            </span>
            <ExternalLink className="w-3 h-3 text-gray-400" />
          </div>
        </div>
      </div>
    </div>
  );
});

NodeTooltip.displayName = 'NodeTooltip';

// Helper function to format timestamp
function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'たった今';
    if (minutes < 60) return `${minutes}分前`;
    if (hours < 24) return `${hours}時間前`;
    if (days < 7) return `${days}日前`;

    return date.toLocaleDateString('ja-JP', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch {
    return timestamp;
  }
}
