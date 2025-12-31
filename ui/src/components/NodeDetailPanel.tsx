/**
 * NodeDetailPanel - Display detailed node information and relationships
 */

import React from 'react';
import { X, Network, ArrowRight, ArrowLeft, Loader2, Target } from 'lucide-react';
import { useDependencies, useDependents, useNodeNeighbors, useNode, useImpactRadius } from '../api/hooks';

interface NodeDetailPanelProps {
  nodeId: string;
  onClose: () => void;
  onNodeSelect?: (nodeId: string) => void;
  onShowImpactRadius?: (nodeIds: string[], depth: number) => void;
}

interface Node {
  id: string;
  labels: string[];
  properties: Record<string, any>;
}

type TabType = 'overview' | 'relationships' | 'impact';

const NodeDetailPanel = ({ nodeId, onClose, onNodeSelect, onShowImpactRadius }: NodeDetailPanelProps) => {
  const [activeTab, setActiveTab] = React.useState<TabType>('overview');
  const [impactDepth, setImpactDepth] = React.useState(2);
  const [showImpact, setShowImpact] = React.useState(false);

  // Fetch node data
  const { data: nodeData, isLoading: nodeLoading } = useNode(nodeId);
  const { data: dependenciesData, isLoading: depsLoading } = useDependencies(nodeId, 2);
  const { data: dependentsData, isLoading: deptsLoading } = useDependents(nodeId, 2);
  const { data: neighborsData, isLoading: neighborsLoading } = useNodeNeighbors(nodeId);
  const { data: impactData, isLoading: impactLoading } = useImpactRadius(nodeId, impactDepth, showImpact);

  const node: Node | null = (nodeData as any)?.data?.node || null;
  const dependencies: Node[] = (dependenciesData as any)?.data?.dependencies || [];
  const dependents: Node[] = (dependentsData as any)?.data?.dependents || [];
  const neighbors: Node[] = (neighborsData as any)?.data?.neighbors || [];
  const impactNodes: Node[] = (impactData as any)?.data?.nodes || [];

  const handleNodeClick = (clickedNodeId: string) => {
    if (onNodeSelect) {
      onNodeSelect(clickedNodeId);
    }
  };

  const handleShowImpact = () => {
    setShowImpact(true);
  };

  const handleHideImpact = () => {
    setShowImpact(false);
    if (onShowImpactRadius) {
      onShowImpactRadius([], 0);
    }
  };

  // Update impact radius visualization when data changes
  React.useEffect(() => {
    if (showImpact && impactNodes.length > 0 && onShowImpactRadius) {
      const nodeIds = impactNodes.map(n => n.id);
      onShowImpactRadius(nodeIds, impactDepth);
    }
  }, [showImpact, impactNodes, impactDepth, onShowImpactRadius]);

  return (
    <div className="h-full bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex flex-col animate-in slide-in-from-right duration-300 w-full md:w-96 lg:w-[28rem] xl:w-[32rem]">
      {/* Header */}
      <div className="px-3 sm:px-4 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white flex items-center justify-between flex-shrink-0">
        <div className="flex items-center gap-2 min-w-0 flex-1">
          <Network className="w-4 h-4 sm:w-5 sm:h-5 flex-shrink-0" />
          <h3 className="font-semibold text-sm sm:text-base truncate">ノード詳細</h3>
        </div>
        <button
          onClick={onClose}
          className="p-1 hover:bg-white/20 rounded transition-colors flex-shrink-0"
          aria-label="閉じる"
        >
          <X className="w-4 h-4 sm:w-5 sm:h-5" />
        </button>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex-shrink-0">
        <nav className="flex overflow-x-auto">
          <button
            onClick={() => setActiveTab('overview')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'overview'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            概要
          </button>
          <button
            onClick={() => setActiveTab('relationships')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'relationships'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            関係性
          </button>
          <button
            onClick={() => setActiveTab('impact')}
            className={`flex-1 px-3 sm:px-4 py-2 sm:py-3 text-xs sm:text-sm font-medium border-b-2 transition-all duration-200 whitespace-nowrap ${
              activeTab === 'impact'
                ? 'border-blue-600 text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20'
                : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
          >
            影響範囲
          </button>
        </nav>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3 sm:p-4 space-y-4">
        {/* Node Basic Info */}
        <div
          className="transition-opacity duration-200"
          style={{
            display: activeTab === 'overview' ? 'block' : 'none',
            opacity: activeTab === 'overview' ? 1 : 0
          }}
        >
        {nodeLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="w-6 h-6 animate-spin text-blue-600 dark:text-blue-400" />
          </div>
        ) : node ? (
          <div className="space-y-2">
            <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1">
              基本情報
            </h4>
            <div className="space-y-1 text-sm">
              <div>
                <span className="font-medium text-gray-600 dark:text-gray-400">ID:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100 font-mono text-xs">{node.id}</span>
              </div>
              <div>
                <span className="font-medium text-gray-600 dark:text-gray-400">Type:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{node.properties.type}</span>
              </div>
              <div>
                <span className="font-medium text-gray-600 dark:text-gray-400">Name:</span>{' '}
                <span className="text-gray-900 dark:text-gray-100">{node.properties.name}</span>
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
        ) : (
          <div className="text-sm text-gray-500 dark:text-gray-400">ノード情報が見つかりません</div>
        )}
        </div>

        {/* Dependencies */}
        <div
          className="space-y-4 transition-opacity duration-200"
          style={{
            display: activeTab === 'relationships' ? 'block' : 'none',
            opacity: activeTab === 'relationships' ? 1 : 0
          }}
        >
        <div className="space-y-2">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
            <ArrowRight className="w-4 h-4 text-green-600 dark:text-green-400" />
            依存先 (Dependencies)
          </h4>
          {depsLoading ? (
            <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
              <Loader2 className="w-4 h-4 animate-spin" />
              読み込み中...
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
                    {dep.properties.name}
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                    {dep.properties.type}
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="text-sm text-gray-500 dark:text-gray-400">依存先なし</div>
          )}
        </div>

        {/* Dependents */}
        <div className="space-y-2">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
            <ArrowLeft className="w-4 h-4 text-orange-600 dark:text-orange-400" />
            依存元 (Dependents)
          </h4>
          {deptsLoading ? (
            <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
              <Loader2 className="w-4 h-4 animate-spin" />
              読み込み中...
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
                    {dept.properties.name}
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                    {dept.properties.type}
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="text-sm text-gray-500 dark:text-gray-400">依存元なし</div>
          )}
        </div>

        {/* Neighbors */}
        <div className="space-y-2">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
            <Network className="w-4 h-4 text-purple-600 dark:text-purple-400" />
            隣接ノード (Neighbors)
          </h4>
          {neighborsLoading ? (
            <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
              <Loader2 className="w-4 h-4 animate-spin" />
              読み込み中...
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
                    {neighbor.properties.name}
                  </div>
                  <div className="text-xs text-gray-500 dark:text-gray-400 truncate font-mono">
                    {neighbor.properties.type}
                  </div>
                </button>
              ))}
              {neighbors.length > 10 && (
                <div className="text-xs text-gray-500 dark:text-gray-400 text-center py-1">
                  ... 他 {neighbors.length - 10} 件
                </div>
              )}
            </div>
          ) : (
            <div className="text-sm text-gray-500 dark:text-gray-400">隣接ノードなし</div>
          )}
        </div>
        </div>

        {/* Impact Radius */}
        <div
          className="transition-opacity duration-200"
          style={{
            display: activeTab === 'impact' ? 'block' : 'none',
            opacity: activeTab === 'impact' ? 1 : 0
          }}
        >
        <div className="space-y-2">
          <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300 border-b border-gray-200 dark:border-gray-600 pb-1 flex items-center gap-2">
            <Target className="w-4 h-4 text-red-600 dark:text-red-400" />
            インパクト半径 (Impact Radius)
          </h4>
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <label className="text-xs text-gray-600 dark:text-gray-400">深さ:</label>
              <select
                value={impactDepth}
                onChange={(e) => setImpactDepth(Number(e.target.value))}
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
                onClick={handleShowImpact}
                className="w-full px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded text-sm font-medium transition-colors"
              >
                インパクト半径を表示
              </button>
            ) : (
              <>
                <button
                  onClick={handleHideImpact}
                  className="w-full px-3 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded text-sm font-medium transition-colors"
                >
                  ハイライトを解除
                </button>
                {impactLoading ? (
                  <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    読み込み中...
                  </div>
                ) : (
                  <div className="text-sm text-gray-600 dark:text-gray-400">
                    影響範囲: <span className="font-bold text-red-600 dark:text-red-400">{impactNodes.length}</span> ノード
                  </div>
                )}
              </>
            )}
          </div>
        </div>
        </div>
      </div>
    </div>
  );
};

export default NodeDetailPanel;
