/**
 * Custom Node Component for React Flow
 * Enhanced with provider color bands, modern card design, and hover animations
 */

import { memo, useState, useRef, useMemo } from 'react';
import { Handle, Position } from 'reactflow';
import type { NodeProps } from 'reactflow';
import { OfficialCloudIcon, getProviderFromType, getProviderColor } from '../icons/OfficialCloudIcons';
import { NodeTooltip } from '../graph/NodeTooltip';
import { NodeContextMenu } from '../graph/NodeContextMenu';

interface CustomNodeData {
  label: string;
  type: string;
  resource_type: string;
  severity?: 'critical' | 'high' | 'medium' | 'low';
  resource_name?: string;
  metadata?: Record<string, unknown>;
}

const getSeverityColor = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return 'border-red-600 dark:border-red-500 bg-red-50 dark:bg-red-950/20';
    case 'high':
      return 'border-orange-600 dark:border-orange-500 bg-orange-50 dark:bg-orange-950/20';
    case 'medium':
      return 'border-yellow-600 dark:border-yellow-500 bg-yellow-50 dark:bg-yellow-950/20';
    case 'low':
      return 'border-blue-600 dark:border-blue-500 bg-blue-50 dark:bg-blue-950/20';
    default:
      return 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800';
  }
};

const getSeverityBorderWidth = (severity?: string) => {
  switch (severity) {
    case 'critical': return 'border-[3px]';
    case 'high': return 'border-[3px]';
    case 'medium': return 'border-2';
    default: return 'border';
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

export const CustomNode = memo(({ data, selected, id }: NodeProps<CustomNodeData>) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const [showContextMenu, setShowContextMenu] = useState(false);
  const [contextMenuPosition, setContextMenuPosition] = useState({ x: 0, y: 0 });
  const nodeRef = useRef<HTMLDivElement>(null);
  const severityColor = getSeverityColor(data.severity);
  const severityBorder = getSeverityBorderWidth(data.severity);
  const badgeColor = getSeverityBadgeColor(data.severity);

  // Provider-based color band
  const provider = useMemo(() => getProviderFromType(data.resource_type || data.type), [data.resource_type, data.type]);
  const providerColor = useMemo(() => getProviderColor(provider), [provider]);

  const handleMouseEnter = () => {
    if (nodeRef.current) {
      const rect = nodeRef.current.getBoundingClientRect();
      setTooltipPosition({
        x: rect.left + rect.width / 2,
        y: rect.top,
      });
      setShowTooltip(true);
    }
  };

  const handleMouseLeave = () => {
    setShowTooltip(false);
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenuPosition({ x: e.clientX, y: e.clientY });
    setShowContextMenu(true);
    setShowTooltip(false);
  };

  const handleDoubleClick = () => {
    window.dispatchEvent(new CustomEvent('node-focus', { detail: { nodeId: id, data } }));
  };

  const handleClick = () => {
    window.dispatchEvent(new CustomEvent('node-detail', { detail: { nodeId: id, data } }));
  };

  const handleViewDetails = () => {
    window.dispatchEvent(new CustomEvent('node-detail', { detail: { nodeId: id, data } }));
  };

  const handleFocusView = () => {
    window.dispatchEvent(new CustomEvent('node-focus', { detail: { nodeId: id, data } }));
  };

  const handleShowDependencies = () => {
    window.dispatchEvent(new CustomEvent('node-dependencies', { detail: { nodeId: id, data } }));
  };

  const handleShowImpact = () => {
    window.dispatchEvent(new CustomEvent('node-impact', { detail: { nodeId: id, data } }));
  };

  const handleCopyId = () => {
    navigator.clipboard.writeText(id || '');
  };

  return (
    <>
      <div
        ref={nodeRef}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        onContextMenu={handleContextMenu}
        onDoubleClick={handleDoubleClick}
        onClick={handleClick}
        className={`
          relative rounded-2xl shadow-lg overflow-hidden
          transition-all duration-300 ease-out min-w-[220px] cursor-pointer
          ${severityBorder} ${severityColor}
          ${selected
            ? 'ring-4 ring-blue-500/60 shadow-2xl scale-105'
            : 'hover:shadow-xl hover:scale-[1.03] hover:-translate-y-0.5'
          }
        `}
      >
        {/* Provider Color Band (top accent) */}
        <div
          className="h-1.5 w-full"
          style={{ backgroundColor: providerColor }}
        />

        {/* Input Handle */}
        <Handle
          type="target"
          position={Position.Top}
          className="w-3 h-3 !bg-blue-500 !border-2 !border-white dark:!border-gray-900"
        />

        <div className="px-5 py-4">
          {/* Icon */}
          <div className="flex justify-center mb-3">
            <div className="p-3 bg-white dark:bg-gray-700 rounded-xl shadow-md ring-1 ring-gray-100 dark:ring-gray-600">
              <OfficialCloudIcon
                type={data.resource_type || data.type}
                size={56}
              />
            </div>
          </div>

          {/* Label */}
          <div className="text-center">
            <div className="font-semibold text-sm text-gray-900 dark:text-gray-100 whitespace-pre-line leading-tight">
              {data.label}
            </div>
            {data.resource_name && (
              <div className="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate font-medium max-w-[200px] mx-auto">
                {data.resource_name}
              </div>
            )}
            {/* Provider tag */}
            <div className="mt-2 flex justify-center">
              <span
                className="text-[10px] font-semibold px-2 py-0.5 rounded-full text-white"
                style={{ backgroundColor: providerColor }}
              >
                {provider.toUpperCase()}
              </span>
            </div>
          </div>
        </div>

        {/* Severity Badge */}
        {data.severity && (
          <div className="absolute -top-1 -right-1">
            <span className={`
              px-2 py-1 text-xs font-bold rounded-full shadow-lg
              ring-2 ring-white dark:ring-gray-900
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
          className="w-3 h-3 !bg-green-500 !border-2 !border-white dark:!border-gray-900"
        />
      </div>

      {/* Tooltip */}
      {showTooltip && !showContextMenu && (
        <NodeTooltip
          data={{
            id: id || '',
            label: data.label,
            type: data.type,
            resourceType: data.resource_type,
            resourceName: data.resource_name || data.label,
            severity: (data.severity || 'low') as 'low' | 'medium' | 'high' | 'critical',
            metadata: data.metadata,
          }}
          position={tooltipPosition}
        />
      )}

      {/* Context Menu */}
      {showContextMenu && (
        <NodeContextMenu
          position={contextMenuPosition}
          nodeId={id || ''}
          nodeData={{
            label: data.label,
            type: data.type,
            resource_type: data.resource_type,
          }}
          onClose={() => setShowContextMenu(false)}
          onViewDetails={handleViewDetails}
          onFocusView={handleFocusView}
          onShowDependencies={handleShowDependencies}
          onShowImpact={handleShowImpact}
          onCopyId={handleCopyId}
        />
      )}
    </>
  );
});

CustomNode.displayName = 'CustomNode';
