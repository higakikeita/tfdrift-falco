/**
 * CytoscapeGraph Component
 *
 * TFDrift-Falco因果関係グラフのメインビジュアライゼーション
 */

import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import type { Core } from 'cytoscape';
// @ts-expect-error - cytoscape-dagre lacks type definitions
import dagre from 'cytoscape-dagre';
// @ts-expect-error - cytoscape-fcose lacks type definitions
import fcose from 'cytoscape-fcose';
import { cytoscapeConfig, layoutConfigs } from '../styles/cytoscapeStyles';
import { AWS_SERVICE_LEGEND, DRIFT_STATUS_LEGEND } from '../constants/colors';
import type { CytoscapeElements } from '../types/graph';

type LayoutType = 'fcose' | 'dagre' | 'concentric' | 'cose' | 'grid';

// Register layouts
cytoscape.use(dagre);
cytoscape.use(fcose);

interface CytoscapeGraphProps {
  elements: CytoscapeElements;
  layout?: LayoutType;
  onNodeClick?: (nodeId: string, nodeData: unknown) => void;
  onEdgeClick?: (edgeId: string, edgeData: unknown) => void;
  highlightedPath?: string[];
  className?: string;
}

export const CytoscapeGraph: React.FC<CytoscapeGraphProps> = ({
  elements,
  layout = 'dagre',
  onNodeClick,
  onEdgeClick,
  highlightedPath = [],
  className = ''
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [showOptionsPanel, setShowOptionsPanel] = useState(false);
  const [showLegend, setShowLegend] = useState(true);
  const [filterMode, setFilterMode] = useState<'all' | 'drift-only' | 'vpc-only'>('all');
  const [currentLayout, setCurrentLayout] = useState<LayoutType>(layout);
  const [nodeScale, setNodeScale] = useState<number>(1.0); // Scale multiplier for node sizes

  // Draggable panel state
  const [panelPosition, setPanelPosition] = useState({ x: 0, y: 60 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const panelRef = useRef<HTMLDivElement>(null);

  // Initialize Cytoscape
  useEffect(() => {
    if (!containerRef.current) return;

    let cy;
    try {
      // Create Cytoscape instance
      cy = cytoscape({
        container: containerRef.current,
        elements: {
          nodes: elements.nodes,
          edges: elements.edges
        },
        ...cytoscapeConfig
      });

      cyRef.current = cy;

      // Apply initial layout
      const layoutConfig = (layoutConfigs as Record<string, object>)[layout];
      cy.layout(layoutConfig).run();

      // Event handlers
      cy.on('tap', 'node', (evt) => {
        const node = evt.target;
        const nodeId = node.id();
        setSelectedNode(nodeId);

        if (onNodeClick) {
          onNodeClick(nodeId, node.data());
        }
      });

      cy.on('tap', 'edge', (evt) => {
        const edge = evt.target;
        const edgeId = edge.id();

        if (onEdgeClick) {
          onEdgeClick(edgeId, edge.data());
        }
      });

      // Hover effects
      cy.on('mouseover', 'node', (evt) => {
        const node = evt.target;
        // Show connected edges
        const connectedEdges = node.connectedEdges();
        connectedEdges.addClass('hover-highlight');
      });

      cy.on('mouseout', 'node', (evt) => {
        const node = evt.target;
        const connectedEdges = node.connectedEdges();
        connectedEdges.removeClass('hover-highlight');
      });

    } catch {
      return;
    }

    // Cleanup
    return () => {
      if (cy) {
        cy.destroy();
      }
    };
  }, [elements, layout, onNodeClick, onEdgeClick]);

  // Update highlighted path
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cyRef.current;

    // Remove previous highlights
    cy.elements().removeClass('highlighted');

    if (highlightedPath.length > 0) {
      // Highlight nodes in path
      highlightedPath.forEach(nodeId => {
        const node = cy.$id(nodeId);
        node.addClass('highlighted');
      });

      // Highlight edges in path
      for (let i = 0; i < highlightedPath.length - 1; i++) {
        const edge = cy.edges(`[source="${highlightedPath[i]}"][target="${highlightedPath[i + 1]}"]`);
        edge.addClass('highlighted');
      }

      // Center view on path
      const pathElements = cy.collection();
      highlightedPath.forEach(nodeId => {
        pathElements.merge(cy.$id(nodeId));
      });
      cy.fit(pathElements, 100);
    }
  }, [highlightedPath]);

  // Public methods accessible via ref
  useEffect(() => {
    if (!cyRef.current) return;

  }, []);

  // Handle layout changes
  useEffect(() => {
    if (!cyRef.current) return;

    const layoutConfig = (layoutConfigs as Record<string, object>)[layout];
    cyRef.current.layout(layoutConfig).run();
  }, [layout]);

  // Handle filter mode changes
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cyRef.current;

    // Show all elements first
    cy.elements().style('display', 'element');

    if (filterMode === 'drift-only') {
      // Show only nodes with severity
      cy.nodes('[!severity]').style('display', 'none');
      // Hide edges connected to hidden nodes
      cy.edges().forEach((edge) => {
        const source = edge.source();
        const target = edge.target();
        if (source.style('display') === 'none' || target.style('display') === 'none') {
          edge.style('display', 'none');
        }
      });
    } else if (filterMode === 'vpc-only') {
      // Show only VPC, Subnet, and network-related resources
      const networkTypes = [
        'aws_vpc', 'aws_subnet', 'aws_internet_gateway', 'aws_nat_gateway',
        'aws_route_table', 'aws_route', 'aws_security_group', 'aws_lb'
      ];
      cy.nodes().forEach((node) => {
        const resourceType = node.data('resource_type') as string;
        if (!networkTypes.includes(resourceType)) {
          node.style('display', 'none');
        }
      });
      // Hide edges connected to hidden nodes
      cy.edges().forEach((edge) => {
        const source = edge.source();
        const target = edge.target();
        if (source.style('display') === 'none' || target.style('display') === 'none') {
          edge.style('display', 'none');
        }
      });
    }
  }, [filterMode]);

  // Handle node scale changes (using Cytoscape zoom)
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cyRef.current;

    // Use Cytoscape's zoom feature as scale
    cy.zoom({
      level: nodeScale,
      renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 }
    });

  }, [nodeScale]);

  // Draggable panel handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    if (!panelRef.current) return;
    setIsDragging(true);
    const rect = panelRef.current.getBoundingClientRect();
    setDragOffset({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top
    });
  };

  useEffect(() => {
    if (isDragging) {
      const handleMouseMove = (e: MouseEvent) => {
        setPanelPosition({
          x: e.clientX - dragOffset.x,
          y: e.clientY - dragOffset.y
        });
      };
      const handleMouseUp = () => {
        setIsDragging(false);
      };
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, dragOffset]);

  return (
    <div className={`relative w-full h-full ${className}`}>
      <div
        ref={containerRef}
        className="w-full h-full bg-white"
        style={{ minHeight: '600px' }}
      />

      {/* Controls Overlay */}
      <div className="absolute top-4 right-4 flex flex-col gap-2">
        <button
          onClick={() => cyRef.current?.fit()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Fit to view"
        >
          🔍 Fit
        </button>
        <button
          onClick={() => cyRef.current?.center()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Center view"
        >
          🎯 Center
        </button>
        <button
          onClick={() => setShowOptionsPanel(!showOptionsPanel)}
          className={`px-3 py-2 rounded-lg shadow-md text-sm font-medium ${
            showOptionsPanel ? 'bg-blue-500 text-white' : 'bg-white hover:bg-gray-100'
          }`}
          title="Show/Hide Options"
        >
          ⚙️ Options
        </button>
        <button
          onClick={() => {
            const png = cyRef.current?.png({ full: true, scale: 2 });
            if (png) {
              const link = document.createElement('a');
              link.href = png;
              link.download = 'tfdrift-graph.png';
              link.click();
            }
          }}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium"
          title="Export as PNG"
        >
          📸 Export
        </button>
      </div>

      {/* Options Panel - Draggable */}
      {showOptionsPanel && (
        <div
          ref={panelRef}
          className="fixed bg-white rounded-lg shadow-xl p-4 w-64 z-50 border-2 border-gray-200"
          style={{
            left: `${panelPosition.x}px`,
            top: `${panelPosition.y}px`,
            cursor: isDragging ? 'grabbing' : 'grab'
          }}
        >
          <div
            className="flex items-center justify-between mb-3 pb-2 border-b border-gray-200"
            onMouseDown={handleMouseDown}
          >
            <h3 className="text-sm font-bold text-gray-900 select-none">⚙️ Display Options</h3>
            <button
              onClick={() => setShowOptionsPanel(false)}
              className="text-gray-400 hover:text-gray-600 text-lg leading-none"
              title="Close"
            >
              ×
            </button>
          </div>

          {/* Filter Mode */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">Filter</label>
            <select
              value={filterMode}
              onChange={(e) => setFilterMode(e.target.value as 'all' | 'drift-only' | 'vpc-only')}
              className="w-full px-2 py-1 text-sm border border-gray-300 rounded"
            >
              <option value="all">All Resources</option>
              <option value="drift-only">Drift Only</option>
              <option value="vpc-only">VPC/Network Only</option>
            </select>
          </div>

          {/* Node Scale Slider */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">
              Node Scale: {nodeScale.toFixed(1)}x
            </label>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">小</span>
              <input
                type="range"
                min="0.5"
                max="2.0"
                step="0.1"
                value={nodeScale}
                onChange={(e) => setNodeScale(parseFloat(e.target.value))}
                className="flex-1"
              />
              <span className="text-xs text-gray-500">大</span>
            </div>
            <div className="flex gap-1 mt-2">
              <button
                onClick={() => setNodeScale(0.7)}
                className={`flex-1 px-2 py-1 text-xs rounded ${nodeScale === 0.7 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
              >
                小
              </button>
              <button
                onClick={() => setNodeScale(1.0)}
                className={`flex-1 px-2 py-1 text-xs rounded ${nodeScale === 1.0 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
              >
                標準
              </button>
              <button
                onClick={() => setNodeScale(1.3)}
                className={`flex-1 px-2 py-1 text-xs rounded ${nodeScale === 1.3 ? 'bg-blue-500 text-white' : 'bg-gray-100 hover:bg-gray-200'}`}
              >
                大
              </button>
            </div>
          </div>

          {/* Legend Toggle */}
          <div className="mb-4">
            <label className="flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={showLegend}
                onChange={(e) => setShowLegend(e.target.checked)}
                className="mr-2"
              />
              <span className="text-sm text-gray-700">Show Legend</span>
            </label>
          </div>

          {/* Layout Selection */}
          <div className="mb-4">
            <label className="text-xs font-semibold text-gray-700 mb-2 block">Layout</label>
            <div className="space-y-1">
              {(['fcose', 'dagre', 'concentric', 'cose', 'grid'] as LayoutType[]).map((layoutType) => (
                <label key={layoutType} className="flex items-center cursor-pointer">
                  <input
                    type="radio"
                    name="layout"
                    value={layoutType}
                    checked={currentLayout === layoutType}
                    onChange={() => {
                      setCurrentLayout(layoutType);
                      const layoutConfig = (layoutConfigs as Record<string, object>)[layoutType];
                      cyRef.current?.layout(layoutConfig).run();
                    }}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700 capitalize">
                    {layoutType}
                    {layoutType === 'fcose' && ' (推奨)'}
                  </span>
                </label>
              ))}
            </div>
          </div>

        </div>
      )}

      {/* Legend */}
      {showLegend && (
        <div className="absolute bottom-4 left-4 bg-white rounded-lg shadow-lg p-4 max-h-96 overflow-y-auto w-64">
        <h3 className="text-sm font-bold mb-2">AWS Services</h3>
        <div className="grid grid-cols-2 gap-1 text-xs">
          {AWS_SERVICE_LEGEND.flatMap((category) =>
            category.items.map((item) => (
              <div key={item.label} className="flex items-center gap-1">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: item.color }} />
                <span className="text-[10px]">{item.label}</span>
              </div>
            ))
          )}
        </div>
        <hr className="my-2" />
        <h3 className="text-sm font-bold mb-2">Drift Status</h3>
        <div className="space-y-1 text-xs">
          {DRIFT_STATUS_LEGEND.map((status) => (
            <div key={status.label} className="flex items-center gap-2">
              <div className={`w-4 h-4 rounded border-4 ${status.borderClass}`} />
              <span>{status.label} ({status.description})</span>
            </div>
          ))}
        </div>
        </div>
      )}

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

export default CytoscapeGraph;
