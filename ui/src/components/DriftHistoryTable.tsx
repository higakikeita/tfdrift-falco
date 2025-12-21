/**
 * Drift History Table Component
 * ドリフト履歴テーブル - 時系列でドリフトイベントを表示
 */

import { useState, useMemo } from 'react';
import type { DriftEvent, DriftSeverity, ChangeType, Provider } from '../types/drift';
import { SiAmazon, SiGooglecloud } from 'react-icons/si';

interface DriftHistoryTableProps {
  drifts: DriftEvent[];
  onSelectDrift?: (drift: DriftEvent) => void;
}

const severityColors: Record<DriftSeverity, string> = {
  critical: 'bg-red-100 text-red-800 border-red-200',
  high: 'bg-orange-100 text-orange-800 border-orange-200',
  medium: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  low: 'bg-blue-100 text-blue-800 border-blue-200',
};

const severityIcons: Record<DriftSeverity, string> = {
  critical: '🚨',
  high: '⚠️',
  medium: '⚡',
  low: 'ℹ️',
};

const changeTypeColors: Record<ChangeType, string> = {
  created: 'text-green-600',
  modified: 'text-blue-600',
  deleted: 'text-red-600',
};

const changeTypeLabels: Record<ChangeType, string> = {
  created: '作成',
  modified: '変更',
  deleted: '削除',
};

export default function DriftHistoryTable({ drifts, onSelectDrift }: DriftHistoryTableProps) {
  const [selectedDrift, setSelectedDrift] = useState<DriftEvent | null>(null);
  const [severityFilter, setSeverityFilter] = useState<Set<DriftSeverity>>(new Set());
  const [providerFilter, setProviderFilter] = useState<Set<Provider>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<'timestamp' | 'severity'>('timestamp');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  // Filter and sort drifts
  const filteredDrifts = useMemo(() => {
    let filtered = [...drifts];

    // Filter by severity
    if (severityFilter.size > 0) {
      filtered = filtered.filter(d => severityFilter.has(d.severity));
    }

    // Filter by provider
    if (providerFilter.size > 0) {
      filtered = filtered.filter(d => providerFilter.has(d.provider));
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(d =>
        d.resourceType.toLowerCase().includes(query) ||
        d.resourceName?.toLowerCase().includes(query) ||
        d.userIdentity.userName.toLowerCase().includes(query) ||
        d.attribute.toLowerCase().includes(query)
      );
    }

    // Sort
    filtered.sort((a, b) => {
      if (sortBy === 'timestamp') {
        const timeA = new Date(a.timestamp).getTime();
        const timeB = new Date(b.timestamp).getTime();
        return sortOrder === 'asc' ? timeA - timeB : timeB - timeA;
      } else {
        const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 };
        const orderA = severityOrder[a.severity];
        const orderB = severityOrder[b.severity];
        return sortOrder === 'asc' ? orderA - orderB : orderB - orderA;
      }
    });

    return filtered;
  }, [drifts, severityFilter, providerFilter, searchQuery, sortBy, sortOrder]);

  const handleRowClick = (drift: DriftEvent) => {
    setSelectedDrift(drift);
    onSelectDrift?.(drift);
  };

  const toggleSeverityFilter = (severity: DriftSeverity) => {
    const newFilter = new Set(severityFilter);
    if (newFilter.has(severity)) {
      newFilter.delete(severity);
    } else {
      newFilter.add(severity);
    }
    setSeverityFilter(newFilter);
  };

  const toggleProviderFilter = (provider: Provider) => {
    const newFilter = new Set(providerFilter);
    if (newFilter.has(provider)) {
      newFilter.delete(provider);
    } else {
      newFilter.add(provider);
    }
    setProviderFilter(newFilter);
  };

  const formatTimestamp = (timestamp: string) => {
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
    return date.toLocaleDateString('ja-JP', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const stats = useMemo(() => {
    const total = filteredDrifts.length;
    const critical = filteredDrifts.filter(d => d.severity === 'critical').length;
    const high = filteredDrifts.filter(d => d.severity === 'high').length;
    const medium = filteredDrifts.filter(d => d.severity === 'medium').length;
    const low = filteredDrifts.filter(d => d.severity === 'low').length;
    return { total, critical, high, medium, low };
  }, [filteredDrifts]);

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900 rounded-lg shadow-lg">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">ドリフト履歴</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400">実環境での変更履歴を時系列で表示</p>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-gray-600 dark:text-gray-400">合計:</span>
            <span className="font-bold text-lg dark:text-gray-100">{stats.total}</span>
          </div>
        </div>

        {/* Stats Bar */}
        <div className="flex items-center gap-4 text-sm">
          <button
            onClick={() => toggleSeverityFilter('critical')}
            className={`px-3 py-1 rounded-full border ${
              severityFilter.has('critical') ? severityColors.critical : 'bg-gray-100 text-gray-600 border-gray-200'
            } transition-colors`}
          >
            🚨 Critical: {stats.critical}
          </button>
          <button
            onClick={() => toggleSeverityFilter('high')}
            className={`px-3 py-1 rounded-full border ${
              severityFilter.has('high') ? severityColors.high : 'bg-gray-100 text-gray-600 border-gray-200'
            } transition-colors`}
          >
            ⚠️ High: {stats.high}
          </button>
          <button
            onClick={() => toggleSeverityFilter('medium')}
            className={`px-3 py-1 rounded-full border ${
              severityFilter.has('medium') ? severityColors.medium : 'bg-gray-100 text-gray-600 border-gray-200'
            } transition-colors`}
          >
            ⚡ Medium: {stats.medium}
          </button>
          <button
            onClick={() => toggleSeverityFilter('low')}
            className={`px-3 py-1 rounded-full border ${
              severityFilter.has('low') ? severityColors.low : 'bg-gray-100 text-gray-600 border-gray-200'
            } transition-colors`}
          >
            ℹ️ Low: {stats.low}
          </button>
        </div>
      </div>

      {/* Filters & Search */}
      <div className="px-6 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        <div className="flex items-center gap-4">
          {/* Search */}
          <div className="flex-1">
            <input
              type="text"
              placeholder="リソース名、ユーザー、属性で検索..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          {/* Provider Filter */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">Provider:</span>
            <button
              onClick={() => toggleProviderFilter('aws')}
              className={`px-3 py-1 rounded border text-sm ${
                providerFilter.has('aws') ? 'bg-orange-100 text-orange-800 border-orange-200 dark:bg-orange-900 dark:text-orange-200 dark:border-orange-700' : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-300 dark:border-gray-600'
              }`}
            >
              AWS
            </button>
            <button
              onClick={() => toggleProviderFilter('gcp')}
              className={`px-3 py-1 rounded border text-sm ${
                providerFilter.has('gcp') ? 'bg-blue-100 text-blue-800 border-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:border-blue-700' : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border-gray-300 dark:border-gray-600'
              }`}
            >
              GCP
            </button>
          </div>

          {/* Sort */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">並び替え:</span>
            <select
              value={`${sortBy}-${sortOrder}`}
              onChange={(e) => {
                const [by, order] = e.target.value.split('-');
                setSortBy(by as 'timestamp' | 'severity');
                setSortOrder(order as 'asc' | 'desc');
              }}
              className="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500"
            >
              <option value="timestamp-desc">最新順</option>
              <option value="timestamp-asc">古い順</option>
              <option value="severity-desc">重大度: 高→低</option>
              <option value="severity-asc">重大度: 低→高</option>
            </select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full">
          <thead className="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
            <tr className="text-xs font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wider">
              <th className="px-4 py-3 text-left">時刻</th>
              <th className="px-4 py-3 text-left">重大度</th>
              <th className="px-4 py-3 text-left">プロバイダー</th>
              <th className="px-4 py-3 text-left">リソース</th>
              <th className="px-4 py-3 text-left">変更</th>
              <th className="px-4 py-3 text-left">属性</th>
              <th className="px-4 py-3 text-left">ユーザー</th>
              <th className="px-4 py-3 text-left">リージョン</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
            {filteredDrifts.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                  該当するドリフトイベントが見つかりません
                </td>
              </tr>
            ) : (
              filteredDrifts.map((drift) => (
                <tr
                  key={drift.id}
                  onClick={() => handleRowClick(drift)}
                  className={`hover:bg-blue-50 dark:hover:bg-blue-900/20 cursor-pointer transition-colors ${
                    selectedDrift?.id === drift.id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  }`}
                >
                  <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300 whitespace-nowrap">
                    {formatTimestamp(drift.timestamp)}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${severityColors[drift.severity]}`}>
                      {severityIcons[drift.severity]} {drift.severity.toUpperCase()}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      {drift.provider === 'aws' && <SiAmazon size={20} className="text-orange-500" />}
                      {drift.provider === 'gcp' && <SiGooglecloud size={20} className="text-blue-500" />}
                      <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
                        {drift.provider.toUpperCase()}
                      </span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div>
                      <div className="text-sm font-medium text-gray-900 dark:text-gray-100">{drift.resourceName || drift.resourceId}</div>
                      <div className="text-xs text-gray-500 dark:text-gray-400 font-mono">{drift.resourceType}</div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-sm font-medium ${changeTypeColors[drift.changeType]}`}>
                      {changeTypeLabels[drift.changeType]}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <code className="text-xs bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-2 py-1 rounded">{drift.attribute}</code>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-200">{drift.userIdentity.userName}</td>
                  <td className="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">{drift.region}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Footer */}
      <div className="px-6 py-3 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        <div className="flex items-center justify-between text-sm text-gray-600 dark:text-gray-300">
          <div>
            表示中: {filteredDrifts.length} / {drifts.length} イベント
          </div>
          {severityFilter.size > 0 || providerFilter.size > 0 || searchQuery ? (
            <button
              onClick={() => {
                setSeverityFilter(new Set());
                setProviderFilter(new Set());
                setSearchQuery('');
              }}
              className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
            >
              フィルターをクリア
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
