/**
 * Hierarchical Group Nodes for AWS Architecture Diagram
 * Following AWS standard architecture diagram conventions
 */

import { memo } from 'react';
import type { NodeProps } from 'reactflow';

interface GroupNodeData {
  label: string;
  type: string;
  level: 'region' | 'vpc' | 'az' | 'subnet';
  metadata?: {
    cidr?: string;
    subnet_type?: string;
    [key: string]: any;
  };
}

/**
 * Region Group Node (Level 1)
 * Background: Orange (#FFF7ED), Border: #F59E0B
 */
export const RegionGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-8 rounded-2xl border-[3px] border-[#F59E0B] bg-[#FFF7ED] dark:bg-[#7C2D12] dark:border-[#EA580C]"
      style={{ minWidth: '1800px', minHeight: '1200px' }}
    >
      <div className="flex items-center gap-2 mb-6">
        <div className="text-2xl font-bold text-[#EA580C] dark:text-[#FB923C]">
          📍 {data.label}
        </div>
      </div>
    </div>
  );
});

RegionGroupNode.displayName = 'RegionGroupNode';

/**
 * VPC Group Node (Level 2)
 * Background: Blue (#EFF6FF), Border: #3B82F6
 */
export const VPCGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-6 rounded-xl border-2 border-[#3B82F6] bg-[#EFF6FF] dark:bg-[#1E3A8A] dark:border-[#60A5FA]"
      style={{ minWidth: '1600px', minHeight: '1000px' }}
    >
      <div className="flex items-center justify-between mb-4">
        <div className="text-xl font-bold text-[#1E40AF] dark:text-[#93C5FD]">
          🔷 {data.label}
        </div>
        {data.metadata?.cidr && (
          <div className="text-sm font-mono text-[#3B82F6] dark:text-[#60A5FA] bg-white/80 dark:bg-gray-800/80 px-3 py-1.5 rounded">
            {data.metadata.cidr}
          </div>
        )}
      </div>
    </div>
  );
});

VPCGroupNode.displayName = 'VPCGroupNode';

/**
 * Availability Zone Group Node (Level 3)
 * Background: Green (#F0FDF4), Border: Dashed #10B981
 */
export const AZGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-5 rounded-lg border-2 border-dashed border-[#10B981] bg-[#F0FDF4] dark:bg-[#064E3B] dark:border-[#34D399]"
      style={{ minWidth: '700px', minHeight: '800px' }}
    >
      <div className="flex items-center gap-2 mb-4">
        <div className="text-lg font-bold text-[#047857] dark:text-[#6EE7B7]">
          📦 {data.label}
        </div>
      </div>
    </div>
  );
});

AZGroupNode.displayName = 'AZGroupNode';

/**
 * Subnet Group Node (Level 4)
 * Public: Blue background, Private: Purple background
 */
export const SubnetGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  const isPublic = data.metadata?.subnet_type === 'public';

  const bgColor = isPublic
    ? 'bg-[#DBEAFE] dark:bg-[#1E3A8A]'
    : 'bg-[#E0E7FF] dark:bg-[#312E81]';

  const borderColor = isPublic
    ? 'border-[#0EA5E9] dark:border-[#38BDF8]'
    : 'border-[#6366F1] dark:border-[#818CF8]';

  const textColor = isPublic
    ? 'text-[#0369A1] dark:text-[#7DD3FC]'
    : 'text-[#4338CA] dark:text-[#A5B4FC]';

  const icon = isPublic ? '🌐' : '🔒';

  return (
    <div
      className={`w-full h-full p-4 rounded-md border ${borderColor} ${bgColor}`}
      style={{ minWidth: '600px', minHeight: '350px' }}
    >
      <div className="flex items-center justify-between mb-3">
        <div className={`text-sm font-bold ${textColor} flex items-center gap-2`}>
          <span className="text-lg">{icon}</span>
          <span>{data.label}</span>
        </div>
        {data.metadata?.cidr && (
          <div className={`text-xs font-mono ${textColor} opacity-80`}>
            {data.metadata.cidr}
          </div>
        )}
      </div>
    </div>
  );
});

SubnetGroupNode.displayName = 'SubnetGroupNode';
