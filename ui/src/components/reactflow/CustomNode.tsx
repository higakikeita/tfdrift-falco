/**
 * Custom Node Component for React Flow
 * High-quality node rendering with official cloud provider icons
 */

import { memo, useState, useRef } from 'react';
import { Handle, Position } from 'reactflow';
import type { NodeProps } from 'reactflow';
import { OfficialCloudIcon } from '../icons/OfficialCloudIcons';
import { NodeTooltip } from '../graph/NodeTooltip';
import { NodeContextMenu } from '../graph/NodeContextMenu';

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

export const CustomNode = memo(({ data, selected, id }: NodeProps<CustomNodeData>) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });
  const [showContextMenu, setShowContextMenu] = useState(false);
  const [contextMenuPosition, setContextMenuPosition] = useState({ x: 0, y: 0 });
  const nodeRef = useRef<HTMLDivElement>(null);
  const severityColor = getSeverityColor(data.severity);
  const badgeColor = getSeverityBadgeColor(data.severity);

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
    // Emit custom event for focus view
    window.dispatchEvent(new CustomEvent('node-focus', { detail: { nodeId: id, data } }));
  };

  const handleClick = () => {
    // Emit custom event for detail panel
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
          relative px-5 py-4 rounded-2xl border-2 shadow-xl
          transition-all duration-300 min-w-[200px] cursor-pointer
          ${severityColor}
          ${selected ? 'ring-4 ring-blue-500 shadow-2xl scale-110 border-blue-500' : 'hover:shadow-2xl hover:scale-105 hover:border-gray-400'}
        `}
      >
      {/* Input Handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 !bg-blue-500 !border-2 !border-white"
      />

      {/* Icon */}
      <div className="flex justify-center mb-3">
        <div className="p-3 bg-white rounded-xl shadow-lg ring-1 ring-gray-200 transform transition-transform hover:scale-110">
          <OfficialCloudIcon
            type={data.resource_type || data.type}
            size={64}
          />
        </div>
      </div>

      {/* Label */}
      <div className="text-center">
        <div className="font-bold text-sm text-gray-900 whitespace-pre-line leading-tight">
          {data.label}
        </div>
        {data.resource_name && (
          <div className="text-xs text-gray-500 mt-1.5 truncate font-medium">
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
