/**
 * CytoscapeGraph Component
 *
 * TFDrift-Falco因果関係グラフのメインビジュアライゼーション
 * This component is a thin wrapper that composes the split cytoscape modules
 */

import { memo, useRef, useState, useCallback } from 'react';
import type { Core } from 'cytoscape';
import { CytoscapeRenderer, CytoscapeToolbar, type LayoutType } from './cytoscape';
import type { CytoscapeElements } from '../types/graph';

interface CytoscapeGraphProps {
  elements: CytoscapeElements;
  layout?: LayoutType;
  onNodeClick?: (nodeId: string, nodeData: unknown) => void;
  onEdgeClick?: (edgeId: string, edgeData: unknown) => void;
  highlightedPath?: string[];
  className?: string;
}

const CytoscapeGraphComponent = ({
  elements,
  layout = 'dagre',
  onNodeClick,
  onEdgeClick,
  highlightedPath = [],
  className = '',
}: CytoscapeGraphProps) => {
  const cyRef = useRef<Core | null>(null);
  const [currentLayout, setCurrentLayout] = useState<LayoutType>(layout);
  const [nodeScale, setNodeScale] = useState<number>(1.0);
  const [showLegend, setShowLegend] = useState(true);
  const [filterMode, setFilterMode] = useState<'all' | 'drift-only' | 'vpc-only'>('all');
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [cy, setCy] = useState<Core | null>(null);

  const handleNodeClick = (nodeId: string, nodeData: unknown) => {
    setSelectedNode(nodeId);
    if (onNodeClick) {
      onNodeClick(nodeId, nodeData);
    }
  };

  const handleCyInitialized = useCallback(() => {
    setCy(cyRef.current);
  }, []);

  return (
    <div className={`relative w-full h-full ${className}`}>
      <CytoscapeRenderer
        elements={elements}
        layout={currentLayout}
        cyRef={cyRef}
        onNodeClick={handleNodeClick}
        onEdgeClick={onEdgeClick}
        highlightedPath={highlightedPath}
        nodeScale={nodeScale}
        filterMode={filterMode}
        onInitialized={handleCyInitialized}
        className="w-full h-full"
      />

      <CytoscapeToolbar
        cy={cy}
        currentLayout={currentLayout}
        onLayoutChange={setCurrentLayout}
        nodeScale={nodeScale}
        onNodeScaleChange={setNodeScale}
        filterMode={filterMode}
        onFilterModeChange={setFilterMode}
        showLegend={showLegend}
        onShowLegendChange={setShowLegend}
      />

      {/* Selected Node Info */}
      {selectedNode && (
        <div className="absolute top-4 left-4 bg-white rounded-lg shadow-lg p-4 max-w-md">
          <h3 className="text-sm font-bold mb-2">Selected Node</h3>
          <p className="text-xs">ID: {selectedNode}</p>
          <button
            onClick={() => setSelectedNode(null)}
            className="mt-2 text-xs text-blue-600 hover:underline"
          >
            Clear selection
          </button>
        </div>
      )}
    </div>
  );
};

export const CytoscapeGraph = memo(CytoscapeGraphComponent);
CytoscapeGraph.displayName = 'CytoscapeGraph';

export default CytoscapeGraph;
