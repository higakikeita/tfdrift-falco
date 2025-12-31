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
 * AWS Official Orange: #FF9900
 */
export const RegionGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-8 rounded-xl border-[4px] shadow-xl"
      style={{
        minWidth: '1950px',
        minHeight: '1300px',
        borderColor: '#FF9900',
        backgroundColor: '#FFF8F0'
      }}
    >
      <div className="flex items-center gap-3 mb-6 px-4 py-2 rounded-lg border-2 inline-block"
        style={{
          backgroundColor: 'rgba(255, 255, 255, 0.9)',
          borderColor: '#FF9900'
        }}
      >
        <div className="text-2xl">🌎</div>
        <div className="text-xl font-bold" style={{ color: '#FF9900' }}>
          {data.label}
        </div>
      </div>
    </div>
  );
});

RegionGroupNode.displayName = 'RegionGroupNode';

/**
 * VPC Group Node (Level 2)
 * AWS Official VPC Blue: #147EBA
 */
export const VPCGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-6 rounded-lg border-[4px] shadow-lg"
      style={{
        minWidth: '1820px',
        minHeight: '1120px',
        borderColor: '#147EBA',
        backgroundColor: '#E6F2F8'
      }}
    >
      <div className="flex items-center justify-between mb-4 px-4 py-2 rounded-md border-2"
        style={{
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#147EBA'
        }}
      >
        <div className="flex items-center gap-2">
          <div className="text-xl">☁️</div>
          <div className="text-lg font-bold" style={{ color: '#147EBA' }}>
            {data.label}
          </div>
        </div>
        {data.metadata?.cidr && (
          <div className="text-sm font-mono px-3 py-1.5 rounded border"
            style={{
              color: '#147EBA',
              backgroundColor: '#F0F8FF',
              borderColor: '#147EBA'
            }}
          >
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
 * AWS Official AZ Green: #759C3E
 */
export const AZGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  return (
    <div
      className="w-full h-full p-5 rounded-lg border-[3px] border-dashed shadow-md"
      style={{
        minWidth: '850px',
        minHeight: '1000px',
        borderColor: '#759C3E',
        backgroundColor: '#F2F7ED'
      }}
    >
      <div className="flex items-center gap-2 mb-4 px-3 py-1.5 rounded border-2 inline-block"
        style={{
          backgroundColor: 'rgba(255, 255, 255, 0.9)',
          borderColor: '#759C3E'
        }}
      >
        <div className="text-base">🏢</div>
        <div className="text-base font-bold" style={{ color: '#759C3E' }}>
          {data.label}
        </div>
      </div>
    </div>
  );
});

AZGroupNode.displayName = 'AZGroupNode';

/**
 * Subnet Group Node (Level 4)
 * Public: Green border, Private: Blue border (AWS Official)
 */
export const SubnetGroupNode = memo(({ data }: NodeProps<GroupNodeData>) => {
  const isPublic = data.metadata?.subnet_type === 'public';

  // AWS Official Subnet Colors - very subtle
  const borderColor = isPublic ? '#00A300' : '#0073BB';  // Green for public, Blue for private
  const bgColor = '#FFFFFF';  // White background like official diagrams
  const textColor = isPublic ? '#00A300' : '#0073BB';

  const icon = isPublic ? '🌐' : '🔒';

  return (
    <div
      className="w-full h-full p-4 rounded-md border-[2px] shadow-sm"
      style={{
        minWidth: '790px',
        minHeight: '420px',
        borderColor: borderColor,
        backgroundColor: bgColor
      }}
    >
      <div className="flex items-center justify-between mb-3">
        <div className="text-sm font-bold flex items-center gap-2 px-2 py-1"
          style={{
            color: textColor
          }}
        >
          <span className="text-base">{icon}</span>
          <span>{data.label}</span>
        </div>
        {data.metadata?.cidr && (
          <div className="text-xs font-mono px-2 py-0.5 border rounded"
            style={{
              color: textColor,
              borderColor: borderColor,
              backgroundColor: 'transparent'
            }}
          >
            {data.metadata.cidr}
          </div>
        )}
      </div>
    </div>
  );
});

SubnetGroupNode.displayName = 'SubnetGroupNode';
