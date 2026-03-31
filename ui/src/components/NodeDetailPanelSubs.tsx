/**
 * Sub-components for NodeDetailPanel
 * Extracted for better code organization and reusability
 */

import React, { useMemo } from 'react';
import { ArrowRight, ArrowLeft, Loader2, Target, Network } from 'lucide-react';
import { useDependencies, useDependents, useNodeNeighbors, useNode, useImpactRadius } from '../api/hooks';
import type { Node, NodeResponse, DependenciesResponse, DependentsResponse, NodeNeighborsResponse, ImpactRadiusResponse } from '../types/api';

interface NodePropertiesProps {
  nodeId: string;
}

export function NodeProperties({ nodeId }: NodePropertiesProps) {
  const { data: nodeData, isLoading: nodeLoading } = useNode(nodeId) as unknown as {
    data?: NodeResponse;
    isLoading: boolean;
  };

  const node: Node | null = nodeData?.data?.node || null;

  if (nodeLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 animate-spin text-blue-600 dark:text-blue-400" />
      </div>
    );
  }

  if (!node) {
    return <div className="text-sm text-gray-500 dark:text-gray-400">ãã¼ãæå ±ãè¦ã¤ããã¾ãã</div>;
  }

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1">
        åºæ¬æå ±
      </h4>
      <div className="space-y-1 text-sm">
        <div>
          <span className="font-medium text-gray-600 dark:text-gray-400">ID:</span>{' '}
          <span className="text-gray-900 dark:text-gray-100 font-mono text-xs">{node.id}</span>
        </div>
        <div>
          <span className="font-medium text-gray-600 dark:text-gray-400">Type:</span>{' '}
          <span className="text-gray-900 dark:text-gray-100">{String(node.properties.type ?? '')}</span>
        </div>
        <div>
          <span className="font-medium text-gray-600 dark:text-gray-400">Name:</span>{' '}
          <span className="text-gray-900 dark:text-gray-100">{String(node.properties.name ?? '')}</span>
        </div>
        <div>
          <span className="font-medium text-gray-600 dark:text-gray-400">Labels:</span>{' '}
          <div className="flex flex-wrap gap-1 mt-1">
            {node.labels.map((label) => (
              <span
                key={label}
                className="px-2 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded text-xs"
              >
                {label}
              </span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

interface NodeRelationshipsProps {
  nodeId: string;
  onNodeSelect?: (nodeId: string) => void;
}

export function NodeRelationships({ nodeId, onNodeSelect }: NodeRelationshipsProps) {
  const { data: dependenciesData, isLoading: depsLoading } = useDependencies(nodeId, 2) as unknown as {
    data?: DependenciesResponse;
    isLoading: boolean;
  };
  const { data: dependentsData, isLoading: deptsLoading } = useDependents(nodeId, 2) as unknown as {
    data?: DependentsResponse;
    isLoading: boolean;
  };
  const { data: neighborsData, isLoading: neighborsLoading } = useNodeNeighbors(nodeId) as unknown as {
    data?: NodeNeighborsResponse;
    isLoading: boolean;
  };

  const dependencies: Node[] = dependenciesData?.data?.dependencies || [];
  const dependents: Node[] = dependentsData?.data?.dependents || [];
  const neighbors: Node[] = neighborsData?.data?.neighbors || [];

  const handleNodeClick = (clickedNodeId: string) => {
    if (onNodeSelect) {
      onNodeSelect(clickedNodeId);
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
          <ArrowRight className="w-4 h-4 text-green-600 dark:text-green-400" />
          ä¾å­å (Dependencies)
        </h4>
        {depsLoading ? (
          <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Loader2 className="w-4 h-4 animate-spin" />
            èª­ã¿è¾¼ã¿ä¸­...
          </div>
        ) : dependencies.length > 0 ? (
          <div className="space-y-1">
            {dependencies.map((dep) => (
              <button
                key={dep.id}
                onClick={() => handleNodeClick(dep.id)}
                className="w-full text-left px-2 py-1.5 rounded bg-gray-50 dark:bg-gray-700 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors text-sm"
              >
                <div className="font-medium text-gray-900 dark:text-gray-100 truncate">
                  {String(dep.properties.name ?? '')}
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                  {String(dep.properties.type ?? '')}
                </div>
              </button>
            ))}
          </div>
        ) : (
          <div className="text-sm text-gray-500 dark:text-gray-400">ä¾å­åãªã</div>
        )}
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
          <ArrowLeft className="w-4 h-4 text-orange-600 dark:text-orange-400" />
          ä¾å­å (Dependents)
        </h4>
        {deptsLoading ? (
          <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Loader2 className="w-4 h-4 animate-spin" />
            èª­ã¿è¾¼ã¿ä¸­...
          </div>
        ) : dependents.length > 0 ? (
          <div className="space-y-1">
            {dependents.map((dept) => (
              <button
                key={dept.id}
                onClick={() => handleNodeClick(dept.id)}
                className="w-full text-left px-2 py-1.5 rounded bg-gray-50 dark:bg-gray-700 hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-colors text-sm"
              >
                <div className="font-medium text-gray-900 dark:text-gray-100 truncate">
                  {String(dept.properties.name ?? '')}
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                  {String(dept.properties.type ?? '')}
                </div>
              </button>
            ))}
          </div>
        ) : (
          <div className="text-sm text-gray-500 dark:text-gray-400">ä¾å­åãªã</div>
        )}
      </div>

      <div className="space-y-2">
        <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
          <Network className="w-4 h-4 text-purple-600 dark:text-purple-400" />
          é£æ¥ãã¼ã (Neighbors)
        </h4>
        {neighborsLoading ? (
          <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Loader2 className="w-4 h-4 animate-spin" />
            èª­ã¿è¾¼ã¿ä¸­...
          </div>
        ) : neighbors.length > 0 ? (
          <div className="space-y-1">
            {neighbors.slice(0, 10).map((neighbor) => (
              <button
                key={neighbor.id}
                onClick={() => handleNodeClick(neighbor.id)}
                className="w-full text-left px-2 py-1.5 rounded bg-gray-50 dark:bg-gray-700 hover:bg-purple-50 dark:hover:bg-purple-900/20 transition-colors text-sm"
              >
                <div className="font-medium text-gray-900 dark:text-gray-100 truncate">
                  {String(neighbor.properties.name ?? '')}
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                  {String(neighbor.properties.type ?? '')}
                </div>
              </button>
            ))}
            {neighbors.length > 10 && (
              <div className="text-xs text-gray-500 dark:text-gray-400 text-center py-1">
                ... ä» {neighbors.length - 10} ä»¶
              </div>
            )}
          </div>
        ) : (
          <div className="text-sm text-gray-500 dark:text-gray-400">é£æ¥ãã¼ããªã</div>
        )}
      </div>
    </div>
  );
}

interface NodeMetadataProps {
  nodeId: string;
  impactDepth: number;
  onImpactDepthChange: (depth: number) => void;
  showImpact: boolean;
  onShowImpact: () => void;
  onHideImpact: () => void;
  onShowImpactRadius?: (nodeIds: string[], depth: number) => void;
}

export function NodeMetadata({
  nodeId,
  impactDepth,
  onImpactDepthChange,
  showImpact,
  onShowImpact,
  onHideImpact,
  onShowImpactRadius,
}: NodeMetadataProps) {
  const { data: impactData, isLoading: impactLoading } = useImpactRadius(
    nodeId,
    impactDepth,
    showImpact
  ) as unknown as { data?: ImpactRadiusResponse; isLoading: boolean };

  const impactNodes = useMemo(() => {
    return impactData?.data?.nodes || [];
  }, [impactData]);

  React.useEffect(() => {
    if (showImpact && impactNodes.length > 0 && onShowImpactRadius) {
      const nodeIds = impactNodes.map((n: Node) => n.id);
      onShowImpactRadius(nodeIds, impactDepth);
    }
  }, [showImpact, impactNodes, impactDepth, onShowImpactRadius]);

  return (
    <div className="space-y-2">
      <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
        <Target className="w-4 h-4 text-red-600 dark:text-red-400" />
        ã¤ã³ãã¯ãåå¾ (Impact Radius)
      </h4>
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <label className="text-xs text-gray-600 dark:text-gray-400">æ·±ã:</label>
          <select
            value={impactDepth}
            onChange={(e) => onImpactDepthChange(Number(e.target.value))}
            className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
          >
            <option value={1}>1 hop</option>
            <option value={2}>2 hops</option>
            <option value={3}>3 hops</option>
            <option value={4}>4 hops</option>
            <option value={5}>5 hops</option>
          </select>
        </div>
        {!showImpact ? (
          <button
            onClick={onShowImpact}
            className="w-full px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded text-sm font-medium transition-colors"
          >
            ã¤ã³ãã¯ãåå¾ãè¡¨ç¤º
          </button>
        ) : (
          <>
            <button
              onClick={onHideImpact}
              className="w-full px-3 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded text-sm font-medium transition-colors"
            >
              ãã¤ã©ã¤ããè§£é¤
            </button>
            {impactLoading ? (
              <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                <Loader2 className="w-4 h-4 animate-spin" />
                èª­ã¿è¾¼ã¿ä¸­...
              </div>
            ) : (
              <div className="text-sm text-gray-600 dark:text-gray-400">
                å½±é¿ç¯å²: <span className="font-bold text-red-600 dark:text-red-400">{impactNodes.length}</span> ãã¼ã
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
