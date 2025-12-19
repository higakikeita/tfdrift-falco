/**
 * CytoscapeGraph Component
 *
 * TFDrift-Falco因果関係グラフのメインビジュアライゼーション
 */

import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import type { Core } from 'cytoscape';
// @ts-ignore
import dagre from 'cytoscape-dagre';
import { cytoscapeConfig, layoutConfigs } from '../styles/cytoscapeStyles';
import type { CytoscapeElements } from '../types/graph';

type LayoutType = 'dagre' | 'concentric' | 'cose' | 'grid';

// Register dagre layout
cytoscape.use(dagre);

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

  // Initialize Cytoscape
  useEffect(() => {
    console.log('🎨 Initializing Cytoscape graph...', {
      nodesCount: elements.nodes.length,
      edgesCount: elements.edges.length,
      layout
    });

    if (!containerRef.current) {
      console.error('❌ Container ref is null!');
      return;
    }

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
        node.style('cursor', 'pointer');

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

  return (
    <div className={`relative w-full h-full ${className}`}>
      <div
        ref={containerRef}
        className="w-full h-full bg-gray-50"
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

      {/* Legend */}
      <div className="absolute bottom-4 left-4 bg-white rounded-lg shadow-lg p-4">
        <h3 className="text-sm font-bold mb-2">Legend</h3>
        <div className="space-y-1 text-xs">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-red-500 rounded" />
            <span>Terraform Change</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-blue-400 rounded" />
            <span>IAM Policy/Role</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-green-400 rounded" />
            <span>ServiceAccount</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-yellow-400 rounded" />
            <span>Pod</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 bg-pink-400 rounded" />
            <span>Falco Event</span>
          </div>
        </div>
      </div>

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
