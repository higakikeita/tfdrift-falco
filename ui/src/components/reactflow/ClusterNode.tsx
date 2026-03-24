/**
 * Cluster Node Component
 * Renders grouped nodes as expandable/collapsible clusters
 */

import { memo } from 'react';
import { Handle, Position } from 'reactflow';
import type { NodeProps } from 'reactflow';
import { ChevronDown, ChevronRight, Package } from 'lucide-react';

interface ClusterNodeData {
  clusterType: string;
  clusterLabel: string;
  childNodeIds: string[];
  isExpanded: boolean;
  childCount: number;
  severityCounts?: Record<string, number>;
  label: string;
}

const getClusterColor = (clusterType: string): string => {
  if (clusterType.startsWith('aws')) return 'border-orange-500 bg-orange-50';
  if (clusterType.startsWith('gcp')) return 'border-blue-500 bg-blue-50';
  if (clusterType.startsWith('kubernetes') || clusterType.startsWith('k8s')) {
    return 'border-purple-500 bg-purple-50';
  }
  if (clusterType === 'critical') return 'border-red-500 bg-red-50';
  if (clusterType === 'high') return 'border-orange-500 bg-orange-50';
  if (clusterType === 'medium') return 'border-yellow-500 bg-yellow-50';
  if (clusterType === 'low') return 'border-blue-500 bg-blue-50';
  return 'border-gray-400 bg-gray-50';
};

const getSeverityBadgeColor = (severity: string): string => {
  switch (severity) {
    case 'critical':
      return 'bg-red-500';
    case 'high':
      return 'bg-orange-500';
    case 'medium':
      return 'bg-yellow-500';
    case 'low':
      return 'bg-blue-500';
    default:
      return 'bg-gray-500';
  }
};

export const ClusterNode = memo(({ data, selected }: NodeProps<ClusterNodeData>) => {
  const clusterColor = getClusterColor(data.clusterType);
  const { isExpanded, childCount, severityCounts = {} } = data;

  // Sort severities by priority
  const severityOrder = ['critical', 'high', 'medium', 'low'];
  const sortedSeverities = Object.entries(severityCounts)
    .filter(([_, count]) => count > 0)
    .sort(([a], [b]) => severityOrder.indexOf(a) - severityOrder.indexOf(b));

  return (
    <div
      className={`
        relative px-4 py-3 rounded-xl border-2 shadow-lg
        transition-all duration-300 min-w-[180px]
        ${clusterColor}
        ${selected ? 'ring-4 ring-blue-500 shadow-2xl scale-105' : 'hover:shadow-xl hover:scale-102'}
      `}
    >
      {/* Handles */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 !bg-blue-500 !border-2 !border-white"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 !bg-green-500 !border-2 !border-white"
      />

      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <div className="p-1.5 bg-white rounded-lg shadow-sm">
            <Package size={20} className="text-gray-700" />
          </div>
          {isExpanded ? (
            <ChevronDown size={16} className="text-gray-600" />
          ) : (
            <ChevronRight size={16} className="text-gray-600" />
          )}
        </div>
        <div className="px-2 py-0.5 bg-white rounded-full shadow-sm">
          <span className="text-xs font-bold text-gray-700">{childCount}</span>
        </div>
      </div>

      {/* Label */}
      <div className="text-center mb-2">
        <div className="font-bold text-sm text-gray-900 truncate">
          {data.clusterLabel}
        </div>
        <div className="text-xs text-gray-600 mt-0.5">
          {isExpanded ? 'Expanded' : 'Collapsed'}
        </div>
      </div>

      {/* Severity breakdown */}
      {sortedSeverities.length > 0 && (
        <div className="flex flex-wrap gap-1 justify-center mt-2">
          {sortedSeverities.map(([severity, count]) => (
            <div
              key={severity}
              className={`
                flex items-center gap-1 px-2 py-0.5 rounded-full text-white text-xs
                ${getSeverityBadgeColor(severity)}
              `}
            >
              <span className="font-medium">{severity}</span>
              <span className="font-bold">{count}</span>
            </div>
          ))}
        </div>
      )}

      {/* Expansion hint */}
      <div className="mt-2 text-center">
        <div className="text-[10px] text-gray-500 font-medium">
          {isExpanded ? 'Click to collapse' : 'Click to expand'}
        </div>
      </div>
    </div>
  );
});

ClusterNode.displayName = 'ClusterNode';

/**
 * Minimal cluster node for high zoom out
 */
export const MinimalClusterNode = memo(({ data }: { data: ClusterNodeData }) => {
  const clusterColor = getClusterColor(data.clusterType);

  return (
    <div
      className={`
        w-4 h-4 rounded-full border-2 shadow-md
        ${clusterColor}
      `}
      title={`${data.clusterLabel} (${data.childCount} nodes)`}
    >
      <div className="absolute -top-1 -right-1 w-3 h-3 bg-gray-700 text-white text-[8px] font-bold rounded-full flex items-center justify-center">
        {data.childCount}
      </div>
    </div>
  );
});

MinimalClusterNode.displayName = 'MinimalClusterNode';
