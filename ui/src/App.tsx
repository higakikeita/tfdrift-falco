/**
 * TFDrift-Falco Graph UI
 *
 * 因果関係グラフビジュアライゼーション - メインアプリケーション
 */

import { useState } from 'react';
import CytoscapeGraph from './components/CytoscapeGraph';
import {
  generateSampleCausalChain,
  generateComplexSampleGraph,
  generateBlastRadiusGraph
} from './utils/sampleData';
import { LayoutType } from './types/graph';
import type { LayoutType as LayoutTypeType } from './types/graph';
import './App.css';

type DemoMode = 'simple' | 'complex' | 'blast-radius';

function App() {
  console.log('🚀 TFDrift-Falco App is loading...');

  const [demoMode, setDemoMode] = useState<DemoMode>('simple');
  const [layout, setLayout] = useState<LayoutTypeType>(LayoutType.HIERARCHICAL);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);

  // Get graph data based on demo mode
  const getGraphData = () => {
    switch (demoMode) {
      case 'simple':
        return generateSampleCausalChain();
      case 'complex':
        return generateComplexSampleGraph();
      case 'blast-radius':
        return generateBlastRadiusGraph();
      default:
        return generateSampleCausalChain();
    }
  };

  const graphData = getGraphData();

  const handleNodeClick = (nodeId: string, nodeData: any) => {
    setSelectedNodeId(nodeId);
    console.log('Node clicked:', nodeId, nodeData);
  };

  const handleHighlightPath = () => {
    if (demoMode === 'simple') {
      // Highlight the full causal chain
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

  return (
    <div className="flex flex-col h-screen bg-gray-100">
      {/* Header */}
      <header className="bg-gradient-to-r from-red-600 to-pink-600 text-white px-6 py-4 shadow-lg">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">TFDrift-Falco Graph UI</h1>
            <p className="text-sm opacity-90">因果関係グラフビジュアライゼーション</p>
          </div>
          <div className="text-right text-sm">
            <p className="font-semibold">Core Value</p>
            <p className="opacity-90">「なぜ」を可視化する</p>
          </div>
        </div>
      </header>

      {/* Control Panel */}
      <div className="bg-white border-b border-gray-200 px-6 py-3 flex items-center gap-4 flex-wrap">
        {/* Demo Mode Selection */}
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Demo Mode:</label>
          <select
            value={demoMode}
            onChange={(e) => setDemoMode(e.target.value as DemoMode)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-red-500"
          >
            <option value="simple">Simple Chain (Drift → Falco)</option>
            <option value="complex">Complex Graph (Multiple Paths)</option>
            <option value="blast-radius">Blast Radius Demo</option>
          </select>
        </div>

        {/* Layout Selection */}
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium text-gray-700">Layout:</label>
          <select
            value={layout}
            onChange={(e) => setLayout(e.target.value as LayoutTypeType)}
            className="px-3 py-1 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-red-500"
          >
            <option value={LayoutType.HIERARCHICAL}>Hierarchical (Top-Down)</option>
            <option value={LayoutType.RADIAL}>Radial (Blast Radius)</option>
            <option value={LayoutType.FORCE}>Force-Directed</option>
            <option value={LayoutType.GRID}>Grid</option>
          </select>
        </div>

        {/* Path Highlight Controls */}
        {demoMode === 'simple' && (
          <div className="flex items-center gap-2">
            <button
              onClick={handleHighlightPath}
              className="px-3 py-1 bg-yellow-400 text-gray-900 rounded-md text-sm font-medium hover:bg-yellow-500"
            >
              ⚡ Highlight Causal Path
            </button>
            <button
              onClick={handleClearHighlight}
              className="px-3 py-1 bg-gray-200 text-gray-700 rounded-md text-sm hover:bg-gray-300"
            >
              Clear
            </button>
          </div>
        )}

        {/* Stats */}
        <div className="ml-auto flex items-center gap-4 text-sm text-gray-600">
          <span>
            <strong>{graphData.nodes.length}</strong> Nodes
          </span>
          <span>
            <strong>{graphData.edges.length}</strong> Edges
          </span>
        </div>
      </div>

      {/* Main Graph View */}
      <div className="flex-1 relative">
        <CytoscapeGraph
          elements={graphData}
          layout={layout}
          onNodeClick={handleNodeClick}
          highlightedPath={highlightedPath}
          className="w-full h-full"
        />
      </div>

      {/* Info Panel (Bottom) */}
      <div className="bg-white border-t border-gray-200 px-6 py-3">
        <div className="flex items-center justify-between">
          <div className="text-sm text-gray-600">
            <p className="font-semibold">TFDrift-Falcoの本質:</p>
            <p className="text-xs">
              Terraform Drift → IAM → ServiceAccount → Pod → Container → Falco Event
              <span className="ml-2 text-red-600">この因果関係が "物語" になる</span>
            </p>
          </div>

          {selectedNodeId && (
            <div className="text-sm">
              <span className="text-gray-600">Selected:</span>
              <span className="ml-2 font-mono text-xs bg-gray-100 px-2 py-1 rounded">
                {selectedNodeId}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Help Overlay */}
      <div className="fixed bottom-20 right-6 bg-blue-50 border border-blue-200 rounded-lg p-4 max-w-xs shadow-lg">
        <h3 className="text-sm font-bold text-blue-900 mb-2">💡 Quick Tips</h3>
        <ul className="text-xs text-blue-800 space-y-1">
          <li>• Click nodes to select them</li>
          <li>• Use mouse wheel to zoom</li>
          <li>• Drag to pan the view</li>
          <li>• Try "Highlight Causal Path" in Simple mode</li>
          <li>• Switch layouts to see different perspectives</li>
        </ul>
      </div>
    </div>
  );
}

export default App;
