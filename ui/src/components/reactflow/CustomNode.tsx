/**
 * Custom Node Component for React Flow
 * High-quality node rendering with official cloud provider icons
 */

import { memo } from 'react';
import { Handle, Position } from 'reactflow';
import type { NodeProps } from 'reactflow';
import { OfficialCloudIcon } from '../icons/OfficialCloudIcons';

interface CustomNodeData {
  label: string;
  type: string;
  resource_type: string;
  severity?: 'critical' | 'high' | 'medium' | 'low';
  resource_name?: string;
  metadata?: Record<string, any>;
}

const getSeverityColor = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return 'border-red-500 bg-red-50';
    case 'high':
      return 'border-orange-500 bg-orange-50';
    case 'medium':
      return 'border-yellow-500 bg-yellow-50';
    case 'low':
      return 'border-blue-500 bg-blue-50';
    default:
      return 'border-gray-300 bg-white';
  }
};

const getSeverityBadgeColor = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return 'bg-red-500 text-white';
    case 'high':
      return 'bg-orange-500 text-white';
    case 'medium':
      return 'bg-yellow-500 text-white';
    case 'low':
      return 'bg-blue-500 text-white';
    default:
      return 'bg-gray-500 text-white';
  }
};

export const CustomNode = memo(({ data, selected }: NodeProps<CustomNodeData>) => {
  const severityColor = getSeverityColor(data.severity);
  const badgeColor = getSeverityBadgeColor(data.severity);

  return (
    <div
      className={`
        relative px-4 py-3 rounded-xl border-2 shadow-lg
        transition-all duration-200 min-w-[180px]
        ${severityColor}
        ${selected ? 'ring-4 ring-blue-400 shadow-2xl scale-105' : 'hover:shadow-xl hover:scale-102'}
      `}
    >
      {/* Input Handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 !bg-blue-500 !border-2 !border-white"
      />

      {/* Icon */}
      <div className="flex justify-center mb-2">
        <div className="p-2 bg-white rounded-lg shadow-md">
          <OfficialCloudIcon
            type={data.resource_type || data.type}
            size={56}
          />
        </div>
      </div>

      {/* Label */}
      <div className="text-center">
        <div className="font-semibold text-sm text-gray-900 whitespace-pre-line">
          {data.label}
        </div>
        {data.resource_name && (
          <div className="text-xs text-gray-600 mt-1 truncate">
            {data.resource_name}
          </div>
        )}
      </div>

      {/* Severity Badge */}
      {data.severity && (
        <div className="absolute -top-2 -right-2">
          <span className={`
            px-2 py-1 text-xs font-bold rounded-full shadow-md
            ${badgeColor}
          `}>
            {data.severity.toUpperCase()}
          </span>
        </div>
      )}

      {/* Output Handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 !bg-green-500 !border-2 !border-white"
      />
    </div>
  );
});

CustomNode.displayName = 'CustomNode';
