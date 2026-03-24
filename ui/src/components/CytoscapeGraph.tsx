/* eslint-disable @typescript-eslint/no-explicit-any */
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
  onNodeClick?: (nodeId: string, nodeData: any) => void;
  onEdgeClick?: (edgeId: string, edgeData: any) => void;
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
    console.log('🎨 Initializing Cytoscape graph with AWS official icons...', {
      nodesCount: elements.nodes.length,
      edgesCount: elements.edges.length,
      layout
    });

    if (!containerRef.current) {
      console.error('❌ Container ref is null!');
      return;
    }

    // Debug: Inspect node data structure BEFORE creating Cytoscape
    const vpcNodes = elements.nodes.filter(n => n.data.resource_type === 'aws_vpc');
    const subnetNodes = elements.nodes.filter(n => n.data.resource_type === 'aws_subnet');
    const nodesWithParent = elements.nodes.filter(n => n.data.parent);

    console.log('🔍 Pre-Cytoscape data inspection:', {
      totalNodes: elements.nodes.length,
      vpcCount: vpcNodes.length,
      subnetCount: subnetNodes.length,
      nodesWithParentCount: nodesWithParent.length,
      sampleVPC: vpcNodes[0] ? {
        id: vpcNodes[0].data.id,
        resource_type: vpcNodes[0].data.resource_type,
        hasParent: !!vpcNodes[0].data.parent,
        parent: vpcNodes[0].data.parent
      } : null,
      sampleSubnets: subnetNodes.slice(0, 2).map(n => ({
        id: n.data.id,
        resource_type: n.data.resource_type,
        parent: n.data.parent,
        vpcId: n.data.metadata?.attributes?.vpc_id
      })),
      sampleChildResources: nodesWithParent.slice(0, 3).map(n => ({
        id: n.data.id,
        resource_type: n.data.resource_type,
        parent: n.data.parent
      }))
    });

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

      console.log('✅ Cytoscape instance created successfully');

      // Debug: Check compound nodes
      const compoundNodes = cy.nodes().filter(node => node.isParent());
      const childNodes = cy.nodes().filter(node => node.isChild());
      console.log('🔍 Compound nodes check:', {
        totalNodes: cy.nodes().length,
        compoundNodesCount: compoundNodes.length,
        childNodesCount: childNodes.length,
        compoundNodeIds: compoundNodes.map(n => n.id()),
        sampleChildNodes: childNodes.slice(0, 3).map(n => ({
          id: n.id(),
          parent: n.data('parent')
        }))
      });

      // Apply initial layout
      const layoutConfig = (layoutConfigs as any)[layout];
      cy.layout(layoutConfig).run();

      console.log('✅ Layout applied successfully');

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

      console.log('✅ Event handlers registered');
    } catch (error) {
      console.error('❌ Error initializing Cytoscape:', error);
      console.error('Error details:', {
        message: error instanceof Error ? error.message : String(error),
        stack: error instanceof Error ? error.stack : undefined
      });
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

    // Expose useful methods
    (window as any).cytoscapeInstance = {
      fit: () => cyRef.current?.fit(),
      center: () => cyRef.current?.center(),
      zoom: (level: number) => cyRef.current?.zoom(level),
      highlightNode: (nodeId: string) => {
        const node = cyRef.current?.$id(nodeId);
        node?.addClass('highlighted');
      },
      clearHighlights: () => {
        cyRef.current?.elements().removeClass('highlighted');
      },
      exportPNG: () => {
        return cyRef.current?.png({ full: true, scale: 2 });
      },
      exportJSON: () => {
        return cyRef.current?.json();
      }
    };
  }, []);

  // Handle layout changes
  useEffect(() => {
    if (!cyRef.current) return;

    const layoutConfig = (layoutConfigs as any)[layout];
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
      cy.edges().forEach((edge: any) => {
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
      cy.nodes().forEach((node: any) => {
        const resourceType = node.data('resource_type');
        if (!networkTypes.includes(resourceType)) {
          node.style('display', 'none');
        }
      });
      // Hide edges connected to hidden nodes
      cy.edges().forEach((edge: any) => {
        const source = edge.source();
        const target = edge.target();
        if (source.style('display') === 'none' || target.style('display') === 'none') {
          edge.style('display', 'none');
        }
      });
    }

    console.log(`🔍 Filter mode changed to: ${filterMode}`);
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

    console.log(`📏 Node scale adjusted to ${nodeScale}x (zoom: ${cy.zoom()})`);
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
              onChange={(e) => setFilterMode(e.target.value as any)}
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
                      const layoutConfig = (layoutConfigs as any)[layoutType];
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

          {/* Debug Info */}
          <div className="pt-3 border-t border-gray-200">
            <button
              onClick={() => {
                console.log('📊 Current graph state:', {
                  nodes: cyRef.current?.nodes().length,
                  edges: cyRef.current?.edges().length,
                  zoom: cyRef.current?.zoom(),
                  pan: cyRef.current?.pan()
                });
              }}
              className="text-xs text-blue-600 hover:underline"
            >
              Show Debug Info
            </button>
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
