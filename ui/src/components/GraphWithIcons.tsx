/**
 * Graph with HTML Icon Overlays
 *
 * Renders cloud provider icons as HTML elements positioned over Cytoscape nodes
 */

import { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import type { Core, NodeSingular } from 'cytoscape';
// @ts-ignore
import dagre from 'cytoscape-dagre';
import { cytoscapeConfig, layoutConfigs } from '../styles/cytoscapeStyles';
import type { CytoscapeElements } from '../types/graph';
import { OfficialCloudIcon } from './icons/OfficialCloudIcons';

cytoscape.use(dagre);

type LayoutType = 'dagre' | 'concentric' | 'cose' | 'grid';

interface GraphWithIconsProps {
  elements: CytoscapeElements;
  layout?: LayoutType;
  onNodeClick?: (nodeId: string, nodeData: any) => void;
  highlightedPath?: string[];
  className?: string;
}

interface IconPosition {
  id: string;
  x: number;
  y: number;
  type: string;
  resourceType: string;
  visible: boolean;
}

export const GraphWithIcons: React.FC<GraphWithIconsProps> = ({
  elements,
  layout = 'dagre',
  onNodeClick,
  highlightedPath = [],
  className = ''
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const [iconPositions, setIconPositions] = useState<IconPosition[]>([]);

  // Update icon positions based on Cytoscape node positions
  const updateIconPositions = (cy: Core) => {
    const positions: IconPosition[] = [];

    cy.nodes().forEach((node: NodeSingular) => {
      const pos = node.renderedPosition();
      const bb = node.renderedBoundingBox();

      positions.push({
        id: node.id(),
        x: pos.x,
        y: pos.y - bb.h / 2 - 20, // Position above node
        type: node.data('type'),
        resourceType: node.data('resource_type') || node.data('type'),
        visible: true
      });
    });

    setIconPositions(positions);
  };

  // Initialize Cytoscape
  useEffect(() => {
    if (!containerRef.current) return;

    const cy = cytoscape({
      container: containerRef.current,
      elements: {
        nodes: elements.nodes,
        edges: elements.edges
      },
      ...cytoscapeConfig
    });

    cyRef.current = cy;

    // Apply layout
    const layoutConfig = (layoutConfigs as any)[layout];
    const layoutInstance = cy.layout(layoutConfig);

    layoutInstance.on('layoutstop', () => {
      updateIconPositions(cy);
    });

    layoutInstance.run();

    // Update positions on pan/zoom
    cy.on('pan zoom', () => {
      updateIconPositions(cy);
    });

    // Node click handler
    cy.on('tap', 'node', (evt) => {
      const node = evt.target;
      if (onNodeClick) {
        onNodeClick(node.id(), node.data());
      }
    });

    // Initial position update
    setTimeout(() => updateIconPositions(cy), 100);

    return () => {
      cy.destroy();
    };
  }, [elements, layout]);

  // Update highlighted path
  useEffect(() => {
    if (!cyRef.current) return;
    const cy = cyRef.current;

    cy.elements().removeClass('highlighted');

    if (highlightedPath.length > 0) {
      highlightedPath.forEach(nodeId => {
        cy.$id(nodeId).addClass('highlighted');
      });

      for (let i = 0; i < highlightedPath.length - 1; i++) {
        const edge = cy.edges(`[source="${highlightedPath[i]}"][target="${highlightedPath[i + 1]}"]`);
        edge.addClass('highlighted');
      }
    }
  }, [highlightedPath]);

  // Re-layout when layout type changes
  useEffect(() => {
    if (!cyRef.current) return;
    const layoutConfig = (layoutConfigs as any)[layout];
    const layoutInstance = cyRef.current.layout(layoutConfig);

    layoutInstance.on('layoutstop', () => {
      updateIconPositions(cyRef.current!);
    });

    layoutInstance.run();
  }, [layout]);

  return (
    <div className={`relative w-full h-full ${className}`}>
      {/* Cytoscape Container */}
      <div
        ref={containerRef}
        className="w-full h-full"
        style={{ background: '#f8fafc' }}
      />

      {/* HTML Icon Overlays - Official Cloud Provider Icons */}
      <div className="absolute inset-0 pointer-events-none">
        {iconPositions.map((pos) => {
          if (!pos.visible) return null;

          return (
            <div
              key={pos.id}
              className="absolute transition-all duration-200"
              style={{
                left: `${pos.x}px`,
                top: `${pos.y}px`,
                transform: 'translate(-50%, -50%)',
                zIndex: 1000
              }}
            >
              <div className="bg-white rounded-xl shadow-xl p-2.5 border border-gray-300 hover:border-blue-500 hover:shadow-2xl transition-all duration-200">
                <OfficialCloudIcon type={pos.resourceType} size={48} />
              </div>
            </div>
          );
        })}
      </div>

      {/* Graph Controls */}
      <div className="absolute top-4 right-4 flex flex-col gap-2 pointer-events-auto">
        <button
          onClick={() => cyRef.current?.fit()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium border border-gray-200"
        >
          Fit
        </button>
        <button
          onClick={() => cyRef.current?.center()}
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium border border-gray-200"
        >
          Center
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
          className="px-3 py-2 bg-white rounded-lg shadow-md hover:bg-gray-100 text-sm font-medium border border-gray-200"
        >
          Export
        </button>
      </div>
    </div>
  );
};

export default GraphWithIcons;
