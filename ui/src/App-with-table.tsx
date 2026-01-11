/**
 * TFDrift-Falco UI - Complete Application
 * グラフビュー + ドリフトテーブルビュー統合版
 */

/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState } from 'react';
import { ReactFlowProvider } from 'reactflow';
import { ReactFlowGraph } from './components/reactflow/ReactFlowGraph';
import DriftHistoryTable from './components/DriftHistoryTable';
import DriftDetailPanel from './components/DriftDetailPanel';
import NodeDetailPanel from './components/NodeDetailPanel';
// import PatternSearchPanel from './components/PatternSearchPanel'; // TODO: Integrate pattern search
import { ThemeToggle } from './components/ThemeToggle';
import {
  generateSampleCausalChain,
  generateComplexSampleGraph,
  generateBlastRadiusGraph,
  generateNetworkDiagram
} from './utils/sampleData';
import { generateSampleDrifts } from './utils/sampleDrifts';
import { LayoutType } from './types/graph';
import type { LayoutType as LayoutTypeType } from './types/graph';
import type { DriftEvent } from './types/drift';
import { useGraph, useDrifts, useCriticalNodes } from './api/hooks';
import { Loader2, AlertCircle } from 'lucide-react';
import type { DriftAlert } from './api/types';
import './App.css';

type DemoMode = 'api' | 'simple' | 'complex' | 'blast-radius' | 'network-diagram';
type ViewMode = 'graph' | 'table' | 'split';

// Convert API DriftAlert to UI DriftEvent
const convertDriftAlertToEvent = (alert: DriftAlert): DriftEvent => ({
  id: alert.id,
  timestamp: alert.timestamp,
  severity: alert.severity as any,
  provider: 'aws' as any, // TODO: Get from alert
  resourceType: alert.resource_type,
  resourceId: alert.resource_id,
  resourceName: alert.resource_name,
  changeType: 'modified' as any, // TODO: Determine from alert
  attribute: alert.attribute,
  oldValue: alert.old_value,
  newValue: alert.new_value,
  userIdentity: {
    type: alert.user_identity.Type,
    userName: alert.user_identity.UserName,
    arn: alert.user_identity.ARN,
    accountId: alert.user_identity.AccountID,
    principalId: alert.user_identity.PrincipalID,
  },
  region: '', // TODO: Get from alert
});

function App() {
  console.log('🚀 TFDrift-Falco App is loading...');

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>('split');
  const [demoMode, setDemoMode] = useState<DemoMode>('api');
  const [layout, setLayout] = useState<LayoutTypeType>(LayoutType.HIERARCHICAL);

  // Graph state
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);
  const [highlightedNodes, setHighlightedNodes] = useState<string[]>([]);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [showCriticalNodes, setShowCriticalNodes] = useState(false);
  const [criticalNodeMin, setCriticalNodeMin] = useState(3);

  // Drift table state
  const [selectedDrift, setSelectedDrift] = useState<DriftEvent | null>(null);

  // Fetch data from API
  const { data: apiGraphData, isLoading: graphLoading, error: graphError } = useGraph();
  const { data: apiDriftsData, isLoading: driftsLoading, error: driftsError } = useDrifts({ limit: 100 });
  const { data: criticalNodesData } = useCriticalNodes(showCriticalNodes ? criticalNodeMin : 999);

  // Convert API drifts to UI format or use sample data
  const drifts: DriftEvent[] = demoMode === 'api' && apiDriftsData?.data
    ? apiDriftsData.data.map(convertDriftAlertToEvent)
    : generateSampleDrifts(100);

  // Get critical node IDs
  const criticalNodeIds: string[] = showCriticalNodes && criticalNodesData
    ? ((criticalNodesData as any)?.data?.critical_nodes || []).map((node: any) => node.id)
    : [];

  // Get graph data based on demo mode
  const getGraphData = () => {
    if (demoMode === 'api') {
      if (apiGraphData) {
        // Convert API types to UI types
        const convertedNodes = (apiGraphData.nodes || []).map((node: any) => ({
          data: {
            id: node.data.id,
            label: node.data.label,
            type: node.data.type,
            severity: node.data.severity,
            resource_type: node.data.resourceType || node.data.resource_type || node.data.type,
            resource_name: node.data.label,
            metadata: node.data.metadata || {},
          }
        }));
        const convertedEdges = (apiGraphData.edges || []).map((edge: any) => ({
          data: {
            id: edge.data.id,
            source: edge.data.source,
            target: edge.data.target,
            label: edge.data.label || '',
            type: edge.data.type || 'default',
            relationship: edge.data.relationship || edge.data.type || 'related',
          }
        }));
        return {
          nodes: convertedNodes,
          edges: convertedEdges,
        };
      }
      return { nodes: [], edges: [] };
    }

    switch (demoMode) {
      case 'simple':
        return generateSampleCausalChain();
      case 'complex':
        return generateComplexSampleGraph();
      case 'blast-radius':
        return generateBlastRadiusGraph();
      case 'network-diagram':
        return generateNetworkDiagram();
      default:
        return generateSampleCausalChain();
    }
  };

  const graphData = getGraphData();

  const handleNodeClick = (nodeId: string, nodeData: any) => {
    console.log('Node clicked:', nodeId, nodeData);
    setSelectedNode(nodeId);
  };

  const handleHighlightPath = () => {
    if (demoMode === 'simple') {
      setHighlightedPath([
        'drift-001',
        'iam-policy-001',
        'iam-role-001',
        'sa-001',
        'pod-001',
        'container-001',
        'falco-001'
      ]);
    }
  };

  const handleClearHighlight = () => {
    setHighlightedPath([]);
  };

  const handleSelectDrift = (drift: DriftEvent) => {
    setSelectedDrift(drift);
  };

  const handleShowImpactRadius = (nodeIds: string[], depth: number) => {
    console.log('Impact Radius:', nodeIds.length, 'nodes at depth', depth);
    setHighlightedNodes(nodeIds);
  };

  return (
    <div className="flex flex-col h-screen bg-gray-100 dark:bg-gray-900">
      {/* Header */}
      <header className="bg-gradient-to-r from-red-600 to-pink-600 text-white px-6 py-4 shadow-lg">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">TFDrift-Falco Web UI</h1>
            <p className="text-sm opacity-90">リアルタイム Terraform Drift 検知 & 可視化</p>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-right text-sm">
              <p className="font-semibold">Core Value</p>
              <p className="opacity-90">「なぜ」を可視化する</p>
            </div>
            <ThemeToggle />
          </div>
        </div>
      </header>

      {/* Control Panel */}
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-3">
        <div className="flex items-center gap-4 flex-wrap">
          {/* View Mode Tabs */}
          <div className="flex items-center gap-2 border-r border-gray-300 dark:border-gray-600 pr-4">
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">表示:</label>
            <div className="flex rounded-lg border border-gray-300 dark:border-gray-600 overflow-hidden">
              <button
                onClick={() => setViewMode('graph')}
                className={`px-4 py-1.5 text-sm font-medium transition-colors ${
                  viewMode === 'graph'
                    ? 'bg-red-600 text-white'
                    : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                }`}
              >
                📊 グラフ
              </button>
              <button
                onClick={() => setViewMode('table')}
                className={`px-4 py-1.5 text-sm font-medium border-l border-gray-300 dark:border-gray-600 transition-colors ${
                  viewMode === 'table'
                    ? 'bg-red-600 text-white'
                    : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                }`}
              >
                📋 テーブル
              </button>
              <button
                onClick={() => setViewMode('split')}
                className={`px-4 py-1.5 text-sm font-medium border-l border-gray-300 dark:border-gray-600 transition-colors ${
                  viewMode === 'split'
                    ? 'bg-red-600 text-white'
                    : 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600'
                }`}
              >
                ⚡ 分割
              </button>
            </div>
          </div>

          {/* Demo Mode (for Graph) */}
          {(viewMode === 'graph' || viewMode === 'split') && (
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">データソース:</label>
              <select
                value={demoMode}
                onChange={(e) => setDemoMode(e.target.value as DemoMode)}
                className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-red-500"
              >
                <option value="api">🔴 Live API (本番)</option>
                <option value="network-diagram">🟦 Network Diagram (AWS構成図)</option>
                <option value="simple">🟢 Simple Chain (デモ)</option>
                <option value="complex">🟢 Complex Graph (デモ)</option>
                <option value="blast-radius">🟢 Blast Radius (デモ)</option>
              </select>
              {demoMode === 'api' && graphLoading && (
                <Loader2 className="w-4 h-4 animate-spin text-red-600" />
              )}
              {demoMode === 'api' && graphError && (
                <span title="API接続エラー">
                  <AlertCircle className="w-4 h-4 text-red-600" />
                </span>
              )}
            </div>
          )}

          {/* Layout (for Graph) */}
          {(viewMode === 'graph' || viewMode === 'split') && (
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">レイアウト:</label>
              <select
                value={layout}
                onChange={(e) => setLayout(e.target.value as LayoutTypeType)}
                className="px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-red-500"
              >
                <option value={LayoutType.NETWORK_DIAGRAM}>🟦 ネットワーク構成図 (AWS-Style)</option>
                <option value={LayoutType.HIERARCHICAL}>階層 (Hierarchical/Dagre)</option>
                <option value={LayoutType.FORCE}>力学配置 (Force)</option>
                <option value={LayoutType.RADIAL}>放射状 (Radial)</option>
                <option value={LayoutType.GRID}>グリッド (Grid)</option>
              </select>
            </div>
          )}

          {/* Critical Nodes Control */}
          {(viewMode === 'graph' || viewMode === 'split') && (
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">クリティカル:</label>
              <input
                type="checkbox"
                checked={showCriticalNodes}
                onChange={(e) => setShowCriticalNodes(e.target.checked)}
                className="w-4 h-4 text-red-600 rounded focus:ring-red-500"
              />
              {showCriticalNodes && (
                <select
                  value={criticalNodeMin}
                  onChange={(e) => setCriticalNodeMin(Number(e.target.value))}
                  className="px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-xs bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                >
                  <option value={2}>2+ 依存</option>
                  <option value={3}>3+ 依存</option>
                  <option value={4}>4+ 依存</option>
                  <option value={5}>5+ 依存</option>
                </select>
              )}
            </div>
          )}

          {/* Graph Controls */}
          {(viewMode === 'graph' || viewMode === 'split') && (
            <div className="flex items-center gap-2 ml-auto">
              <button
                onClick={handleHighlightPath}
                className="px-3 py-1.5 bg-yellow-100 text-yellow-800 rounded-md text-sm font-medium hover:bg-yellow-200 transition-colors"
              >
                ⚡ パスをハイライト
              </button>
              <button
                onClick={handleClearHighlight}
                className="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded-md text-sm font-medium hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
              >
                ✕ クリア
              </button>
            </div>
          )}

          {/* Stats */}
          <div className="ml-auto text-sm text-gray-600 dark:text-gray-400 flex items-center gap-2">
            <span className="font-medium">総ドリフト数:</span>{' '}
            <span className="text-lg font-bold text-red-600 dark:text-red-400">{drifts.length}</span>
            {demoMode === 'api' && driftsLoading && (
              <Loader2 className="w-4 h-4 animate-spin text-gray-400 dark:text-gray-500" />
            )}
            {demoMode === 'api' && driftsError && (
              <span title="ドリフト読み込みエラー">
                <AlertCircle className="w-4 h-4 text-orange-500" />
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Graph View */}
        {viewMode === 'graph' && (
          <div className="flex-1 flex overflow-hidden">
            <div className="flex-1 p-4">
              <div className="h-full bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
                <ReactFlowProvider>
                  <ReactFlowGraph
                    elements={graphData}
                    layout={layout}
                    onNodeClick={handleNodeClick}
                    highlightedPath={highlightedPath}
                    highlightedNodes={highlightedNodes}
                    criticalNodes={criticalNodeIds}
                  />
                </ReactFlowProvider>
              </div>
            </div>
            {selectedNode && (
              <div className="w-96">
                <NodeDetailPanel
                  nodeId={selectedNode}
                  onClose={() => setSelectedNode(null)}
                  onNodeSelect={(nodeId) => setSelectedNode(nodeId)}
                  onShowImpactRadius={handleShowImpactRadius}
                />
              </div>
            )}
          </div>
        )}

        {/* Table View */}
        {viewMode === 'table' && (
          <div className="flex-1 flex overflow-hidden">
            <div className="flex-1 p-4">
              <DriftHistoryTable
                drifts={drifts}
                onSelectDrift={handleSelectDrift}
              />
            </div>
            {selectedDrift && (
              <div className="w-96">
                <DriftDetailPanel
                  drift={selectedDrift}
                  onClose={() => setSelectedDrift(null)}
                />
              </div>
            )}
          </div>
        )}

        {/* Split View */}
        {viewMode === 'split' && (
          <>
            {/* Left: Graph */}
            <div className="flex-1 p-4 pr-2">
              <div className="h-full bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
                <div className="px-4 py-2 bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600">
                  <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">因果関係グラフ</h3>
                </div>
                <div className="h-[calc(100%-40px)]">
                  <ReactFlowProvider>
                    <ReactFlowGraph
                      elements={graphData}
                      layout={layout}
                      onNodeClick={handleNodeClick}
                      highlightedPath={highlightedPath}
                      highlightedNodes={highlightedNodes}
                      criticalNodes={criticalNodeIds}
                    />
                  </ReactFlowProvider>
                </div>
              </div>
            </div>

            {/* Right: Table + Detail Panels */}
            <div className="flex-1 flex p-4 pl-2 overflow-hidden">
              <div className="flex-1 mr-2">
                <div className="h-full">
                  <DriftHistoryTable
                    drifts={drifts}
                    onSelectDrift={handleSelectDrift}
                  />
                </div>
              </div>
              {selectedNode && (
                <div className="w-96">
                  <NodeDetailPanel
                    nodeId={selectedNode}
                    onClose={() => setSelectedNode(null)}
                    onNodeSelect={(nodeId) => setSelectedNode(nodeId)}
                    onShowImpactRadius={handleShowImpactRadius}
                  />
                </div>
              )}
              {!selectedNode && selectedDrift && (
                <div className="w-80">
                  <div className="h-full bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
                    <DriftDetailPanel
                      drift={selectedDrift}
                      onClose={() => setSelectedDrift(null)}
                    />
                  </div>
                </div>
              )}
            </div>
          </>
        )}
      </div>

      {/* Footer / Status Bar */}
      <div className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 px-6 py-2">
        <div className="flex items-center justify-between text-xs text-gray-600 dark:text-gray-400">
          <div className="flex items-center gap-4">
            <span>🟢 接続: Falco gRPC</span>
            <span>📊 グラフノード: {graphData.nodes?.length || 0}</span>
            <span>🔗 エッジ: {graphData.edges?.length || 0}</span>
          </div>
          <div>
            <span className="font-mono">v0.5.0 | TFDrift-Falco Web UI</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
