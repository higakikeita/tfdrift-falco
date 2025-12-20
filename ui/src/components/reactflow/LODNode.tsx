/**
 * LOD (Level-of-Detail) Node Component
 * Dynamically adjusts rendering detail based on zoom level for performance optimization
 *
 * Zoom levels:
 * - < 0.2: Minimal (4x4px point)
 * - 0.2-0.5: Medium (icon + label)
 * - > 0.5: Full detail (complete CustomNode)
 */

import { memo } from 'react';
import { Handle, Position, useStore } from 'reactflow';
import type { NodeProps } from 'reactflow';
import { OfficialCloudIcon } from '../icons/OfficialCloudIcons';
import { CustomNode } from './CustomNode';

interface LODNodeData {
  label: string;
  type: string;
  resource_type: string;
  severity?: 'critical' | 'high' | 'medium' | 'low';
  resource_name?: string;
  metadata?: Record<string, any>;
}

const getSeverityColor = (severity?: string): string => {
  switch (severity) {
    case 'critical':
      return '#ef4444'; // red-500
    case 'high':
      return '#f97316'; // orange-500
    case 'medium':
      return '#eab308'; // yellow-500
    case 'low':
      return '#3b82f6'; // blue-500
    default:
      return '#6b7280'; // gray-500
  }
};

const getSeverityBorderColor = (severity?: string): string => {
  switch (severity) {
    case 'critical':
      return 'border-red-500';
    case 'high':
      return 'border-orange-500';
    case 'medium':
      return 'border-yellow-500';
    case 'low':
      return 'border-blue-500';
    default:
      return 'border-gray-300';
  }
};

// Minimal node: Just a colored point (zoom < 0.2)
const MinimalNode = memo(({ data }: { data: LODNodeData }) => {
  const color = getSeverityColor(data.severity);

  return (
    <div
      className="w-1 h-1 rounded-full"
      style={{ backgroundColor: color }}
      title={data.label}
    />
  );
});

MinimalNode.displayName = 'MinimalNode';

// Medium detail node: Icon + label (zoom 0.2-0.5)
const MediumNode = memo(({ data, selected }: { data: LODNodeData; selected?: boolean }) => {
  const borderColor = getSeverityBorderColor(data.severity);

  return (
    <div
      className={`
        relative px-2 py-2 rounded-lg border-2 bg-white shadow-md
        transition-all duration-200 min-w-[80px]
        ${borderColor}
        ${selected ? 'ring-2 ring-blue-500 scale-105' : 'hover:shadow-lg'}
      `}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="w-2 h-2 !bg-blue-500 !border-white"
      />

      <div className="flex flex-col items-center gap-1">
        <OfficialCloudIcon
          type={data.resource_type || data.type}
          size={24}
        />
        <div className="text-[10px] font-medium text-gray-700 text-center truncate max-w-[76px]">
          {data.label}
        </div>
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="w-2 h-2 !bg-green-500 !border-white"
      />
    </div>
  );
});

MediumNode.displayName = 'MediumNode';

// Main LOD Node component with zoom-based switching
export const LODNode = memo(({ data, id, selected }: NodeProps<LODNodeData>) => {
  // Get current zoom level from ReactFlow store
  const zoom = useStore((state) => state.transform[2]);

  // Minimal detail: zoom < 0.2
  if (zoom < 0.2) {
    return <MinimalNode data={data} />;
  }

  // Medium detail: zoom 0.2-0.5
  if (zoom < 0.5) {
    return <MediumNode data={data} selected={selected} />;
  }

  // Full detail: zoom >= 0.5
  return <CustomNode data={data} id={id} selected={selected} />;
});

LODNode.displayName = 'LODNode';

// Utility function to determine if LOD should be used based on node count
export const shouldUseLOD = (nodeCount: number): boolean => {
  return nodeCount > 100; // Use LOD for graphs with more than 100 nodes
};

// Utility function to get recommended zoom thresholds based on node count
export const getLODThresholds = (nodeCount: number) => {
  if (nodeCount < 100) {
    return { minimal: 0, medium: 0, full: 0 }; // Always full detail
  } else if (nodeCount < 500) {
    return { minimal: 0.2, medium: 0.5, full: 1.0 };
  } else if (nodeCount < 1000) {
    return { minimal: 0.3, medium: 0.6, full: 1.0 };
  } else {
    return { minimal: 0.4, medium: 0.7, full: 1.0 };
  }
};
