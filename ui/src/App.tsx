/**
 * TFDrift-Falco Graph UI
 *
 * Grafana/Datadog-style dark theme with animations
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

const DEMO_OPTIONS: { value: DemoMode; label: string }[] = [
  { value: 'simple', label: 'Simple Chain' },
  { value: 'complex', label: 'Complex Graph' },
  { value: 'blast-radius', label: 'Blast Radius' },
];

const LAYOUT_OPTIONS: { value: LayoutTypeType; label: string }[] = [
  { value: LayoutType.HIERARCHICAL, label: 'Hierarchical' },
  { value: LayoutType.RADIAL, label: 'Radial' },
  { value: LayoutType.FORCE, label: 'Force-Directed' },
  { value: LayoutType.GRID, label: 'Grid' },
];

function App() {
  const [demoMode, setDemoMode] = useState<DemoMode>('simple');
  const [layout, setLayout] = useState<LayoutTypeType>(LayoutType.HIERARCHICAL);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [highlightedPath, setHighlightedPath] = useState<string[]>([]);

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

  const handleNodeClick = (nodeId: string, nodeData: unknown) => {
    setSelectedNodeId(nodeId);
    console.log('Node clicked:', nodeId, nodeData);
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

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh', background: '#0f1117' }}>
      {/* Header */}
      <header className="header-dark">
        <div className="logo">
          <span className="logo-icon">🦅</span>
          <span className="logo-text">TFDrift<span className="logo-accent">-Falco</span></span>
        </div>
        <div className="live-badge">● LIVE</div>
        <div className="provider-badges">
          <span className="badge-aws">AWS</span>
          <span className="badge-gcp">GCP</span>
        </div>
        <div className="version">v0.5.0</div>
        <div className="header-actions">
          <button className="header-btn" title="Settings">⚙</button>
          <button className="header-btn" title="Help">?</button>
        </div>
      </header>

      {/* Main Content */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Sidebar */}
        <aside className="sidebar">
          {/* Demo Mode */}
          <div className="sidebar-section">
            <div className="sidebar-section-title">Demo</div>
            {DEMO_OPTIONS.map((opt) => (
              <div
                key={opt.value}
                className={`demo-option ${demoMode === opt.value ? 'active' : ''}`}
                onClick={() => setDemoMode(opt.value)}
              >
                <div className="demo-radio" />
                {opt.label}
              </div>
            ))}
          </div>

          {/* Layout Selection */}
          <div className="sidebar-section">
            <div className="sidebar-section-title">Layout</div>
            {LAYOUT_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                className={`layout-btn ${layout === opt.value ? 'active' : ''}`}
                onClick={() => setLayout(opt.value)}
              >
                {opt.label}
              </button>
            ))}
          </div>

          {/* Status Panel */}
          <div className="status-panel">
            <div className="sidebar-section-title">Status</div>
            <div className="status-item">
              <div className="status-dot green" />
              <span>Falco: Connected</span>
            </div>
            <div className="status-item">
              <div className="status-dot green" />
              <span>AWS: Active</span>
            </div>
            <div className="status-item">
              <div className="status-dot neutral" />
              <span>GCP: Standby</span>
            </div>
          </div>
        </aside>

        {/* Graph Area */}
        <div className="graph-area">
          <CytoscapeGraph
            elements={graphData}
            layout={layout}
            onNodeClick={handleNodeClick}
            highlightedPath={highlightedPath}
            className="w-full h-full"
          />
        </div>
      </div>

      {/* Bottom Bar */}
      <div className="bottom-bar">
        <div className="stat-item">
          <span className="stat-label">Nodes:</span>
          <span className="stat-value">{graphData.nodes.length}</span>
        </div>
        <div className="stat-item">
          <span className="stat-label">Edges:</span>
          <span className="stat-value">{graphData.edges.length}</span>
        </div>

        {demoMode === 'simple' && (
          <>
            <button className="highlight-btn" onClick={handleHighlightPath}>
              ⚡ Highlight Path
            </button>
            {highlightedPath.length > 0 && (
              <button className="clear-btn" onClick={handleClearHighlight}>
                Clear
              </button>
            )}
          </>
        )}

        {selectedNodeId && (
          <div className="selected-node">
            <span>Selected:</span>
            <span className="selected-node-id">{selectedNodeId}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
