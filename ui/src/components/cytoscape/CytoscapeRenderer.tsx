/**
 * Cytoscape Renderer Component
 * Core rendering logic and Cytoscape instance management
 */

import { memo, useEffect, useRef } from 'react';
import cytoscape from 'cytoscape';
import type { Core, LayoutOptions } from 'cytoscape';
import dagre from 'cytoscape-dagre';
import fcose from 'cytoscape-fcose';
import { cytoscapeConfig, layoutConfigs } from '../../styles/cytoscapeStyles';
import type { CytoscapeElements } from '../../types/graph';
import { CytoscapeEventHandler, type EventHandlerCallbacks } from './CytoscapeEventHandler';
import type { LayoutType } from './CytoscapeToolbar';

// Register layouts
cytoscape.use(dagre);
cytoscape.use(fcose);

export interface CytoscapeRendererProps {
  elements: CytoscapeElements;
  layout: LayoutType;
  cyRef: React.MutableRefObject<Core | null>;
  onNodeClick?: (nodeId: string, nodeData: unknown) => void;
  onEdgeClick?: (edgeId: string, edgeData: unknown) => void;
  highlightedPath?: string[];
  nodeScale?: number;
  filterMode?: 'all' | 'drift-only' | 'vpc-only';
  onInitialized?: () => void;
  className?: string;
}

const CytoscapeRendererComponent = ({
  elements,
  layout,
  cyRef,
  onNodeClick,
  onEdgeClick,
  highlightedPath = [],
  nodeScale = 1.0,
  filterMode = 'all',
  onInitialized,
  className = '',
}: CytoscapeRendererProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const eventHandlerRef = useRef<CytoscapeEventHandler | null>(null);

  // Initialize Cytoscape
  useEffect(() => {
    if (!containerRef.current) return;

    let cy: Core;
    try {
      cy = cytoscape({
        container: containerRef.current,
        elements: {
          nodes: elements.nodes,
          edges: elements.edges,
        },
        ...cytoscapeConfig,
      });

      cyRef.current = cy;

      // Initialize event handler
      const callbacks: EventHandlerCallbacks = {
        onNodeClick,
        onEdgeClick,
      };
      eventHandlerRef.current = new CytoscapeEventHandler(cy, callbacks);
      eventHandlerRef.current.bind();

      // Apply initial layout
      const layoutConfig = (layoutConfigs as Record<string, LayoutOptions>)[layout];
      cy.layout(layoutConfig).run();

      // Signal that cy is initialized
      if (onInitialized) {
        onInitialized();
      }
    } catch {
      return;
    }

    // Cleanup
    return () => {
      if (eventHandlerRef.current) {
        eventHandlerRef.current.unbind();
      }
      if (cy) {
        cy.destroy();
      }
    };
  }, [elements, layout, onNodeClick, onEdgeClick, cyRef, onInitialized]);

  // Update highlighted path
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cyRef.current;

    // Remove previous highlights
    cy.elements().removeClass('highlighted');

    if (highlightedPath.length > 0) {
      // Highlight nodes in path
      highlightedPath.forEach((nodeId) => {
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
      highlightedPath.forEach((nodeId) => {
        pathElements.merge(cy.$id(nodeId));
      });
      cy.fit(pathElements, 100);
    }
  }, [highlightedPath, cyRef]);

  // Handle layout changes
  useEffect(() => {
    if (!cyRef.current) return;

    const layoutConfig = (layoutConfigs as Record<string, LayoutOptions>)[layout];
    cyRef.current.layout(layoutConfig).run();
  }, [layout, cyRef]);

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
        'aws_vpc',
        'aws_subnet',
        'aws_internet_gateway',
        'aws_nat_gateway',
        'aws_route_table',
        'aws_route',
        'aws_security_group',
        'aws_lb',
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
  }, [filterMode, cyRef]);

  // Handle node scale changes (using Cytoscape zoom)
  useEffect(() => {
    if (!cyRef.current) return;

    const cy = cyRef.current;

    // Use Cytoscape's zoom feature as scale
    cy.zoom({
      level: nodeScale,
      renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 },
    });
  }, [nodeScale, cyRef]);

  return (
    <div
      ref={containerRef}
      className={`w-full h-full bg-white ${className}`}
      style={{ minHeight: '600px' }}
    />
  );
};

export const CytoscapeRenderer = memo(CytoscapeRendererComponent);
CytoscapeRenderer.displayName = 'CytoscapeRenderer';
