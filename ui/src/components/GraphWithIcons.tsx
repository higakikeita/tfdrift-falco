/**
 * Graph with HTML Icon Overlays
 *
 * Renders cloud provider icons as HTML elements positioned over Cytoscape nodes.
 * Enhanced with html-to-image export, dark mode, and network segment visualization.
 */

import { useEffect, useRef, useState, useCallback } from 'react';
import cytoscape from 'cytoscape';
import type { Core, NodeSingular } from 'cytoscape';
// @ts-expect-error - cytoscape-dagre lacks type definitions
import dagre from 'cytoscape-dagre';
import { toPng } from 'html-to-image';
import { cytoscapeConfig, layoutConfigs } from '../styles/cytoscapeStyles';
import type { CytoscapeElements } from '../types/graph';
import { OfficialCloudIcon, getProviderFromType, getProviderColor } from './icons/OfficialCloudIcons';

cytoscape.use(dagre);

type LayoutType = 'dagre' | 'concentric' | 'cose' | 'grid';

interface GraphWithIconsProps {
  elements: CytoscapeElements;
  layout?: LayoutType;
  onNodeClick?: (nodeId: string, nodeData: Record<string, unknown>) => void;
  highlightedPath?: string[];
  className?: string;
  darkMode?: boolean;
}

interface IconPosition {
  id: string;
  x: number;
  y: number;
  type: string;
  resourceType: string;
  visible: boolean;
  provider: string;
  providerColor: string;
}

export const GraphWithIcons: React.FC<GraphWithIconsProps> = ({
  elements,
  layout = 'dagre',
  onNodeClick,
  highlightedPath = [],
  className = '',
  darkMode = false
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);
  const [iconPositions, setIconPositions] = useState<IconPosition[]>([]);

  // Update icon positions based on Cytoscape node positions
  const updateIconPositions = useCallback((cy: Core) => {
    const positions: IconPosition[] = [];

    cy.nodes().forEach((node: NodeSingular) => {
      const pos = node.renderedPosition();
      const bb = node.renderedBoundingBox();
      const resourceType = node.data('resource_type') || node.data('type');
      const provider = getProviderFromType(resourceType);

      positions.push({
        id: node.id(),
        x: pos.x,
        y: pos.y - bb.h / 2 - 20,
        type: node.data('type'),
        resourceType,
        visible: true,
        provider,
        providerColor: getProviderColor(provider),
      });
    });

    setIconPositions(positions);
  }, []);

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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const layoutConfig = (layoutConfigs as any)[layout];
    const layoutInstance = cyRef.current.layout(layoutConfig);

    layoutInstance.on('layoutstop', () => {
      updateIconPositions(cyRef.current!);
    });

    layoutInstance.run();
  }, [layout, updateIconPositions]);

  // High-quality export using html-to-image (captures HTML overlays)
  const handleExport = useCallback(async () => {
    if (!wrapperRef.current) return;

    try {
      const dataUrl = await toPng(wrapperRef.current, {
        quality: 1.0,
        pixelRatio: 2,
        backgroundColor: darkMode ? '#1a202c' : '#f8fafc',
      });
      const link = document.createElement('a');
      link.href = dataUrl;
      link.download = 'tfdrift-graph.png';
      link.click();
    } catch (err) {
      // Fallback to Cytoscape PNG export
      console.warn('html-to-image export failed, using Cytoscape fallback:', err);
      const png = cyRef.current?.png({ full: true, scale: 2 });
      if (png) {
        const link = document.createElement('a');
        link.href = png;
        link.download = 'tfdrift-graph.png';
        link.click();
      }
    }
  }, [darkMode]);

  const bgColor = darkMode ? '#1a202c' : '#f8fafc';
  const cardBg = darkMode ? 'bg-gray-800 border-gray-600' : 'bg-white border-gray-200';
  const btnStyle = darkMode
    ? 'bg-gray-700 hover:bg-gray-600 text-gray-200 border-gray-600'
    : 'bg-white hover:bg-gray-50 text-gray-700 border-gray-200';

  return (
    <div ref={wrapperRef} className={`relative w-full h-full ${className}`}>
      {/* Cytoscape Container */}
      <div
        ref={containerRef}
        className="w-full h-full"
        style={{ background: bgColor }}
      />

      {/* HTML Icon Overlays */}
      <div className="absolute inset-0 pointer-events-none">
        {iconPositions.map((pos) => {
          if (!pos.visible) return null;

          return (
            <div
              key={pos.id}
              className="absolute transition-all duration-150"
              style={{
                left: `${pos.x}px`,
                top: `${pos.y}px`,
                transform: 'translate(-50%, -50%)',
                zIndex: 1000
              }}
            >
              <div className={`rounded-xl shadow-lg p-2 border ${cardBg} hover:shadow-xl transition-shadow duration-200`}>
                {/* Provider color accent */}
                <div
                  className="h-0.5 -mx-2 -mt-2 mb-1.5 rounded-t-xl"
                  style={{ backgroundColor: pos.providerColor }}
                />
                <OfficialCloudIcon type={pos.resourceType} size={40} />
              </div>
            </div>
          );
        })}
      </div>

      {/* Graph Controls */}
      <div className="absolute top-4 right-4 flex flex-col gap-2 pointer-events-auto">
        <button
          onClick={() => cyRef.current?.fit()}
          className={`px-3 py-2 rounded-lg shadow-md text-sm font-medium border ${btnStyle}`}
        >
          Fit
        </button>
        <button
          onClick={() => cyRef.current?.center()}
          className={`px-3 py-2 rounded-lg shadow-md text-sm font-medium border ${btnStyle}`}
        >
          Center
        </button>
        <button
          onClick={handleExport}
          className={`px-3 py-2 rounded-lg shadow-md text-sm font-medium border ${btnStyle}`}
        >
          Export
        </button>
      </div>
    </div>
  );
};

export default GraphWithIcons;
