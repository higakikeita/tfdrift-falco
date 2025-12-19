/**
 * Adapter to convert Cytoscape data format to React Flow format
 */

import type { Node, Edge } from 'reactflow';
import type { CytoscapeElements } from '../types/graph';

export interface ReactFlowData {
  nodes: Node[];
  edges: Edge[];
}

/**
 * Convert Cytoscape elements to React Flow format
 */
export const convertToReactFlow = (cytoscapeElements: CytoscapeElements): ReactFlowData => {
  // Convert nodes
  const nodes: Node[] = cytoscapeElements.nodes.map((node, index) => ({
    id: node.data.id,
    type: 'custom',
    position: {
      x: index * 250, // Initial positioning, will be auto-laid out
      y: Math.floor(index / 4) * 200
    },
    data: {
      label: node.data.label,
      type: node.data.type,
      resource_type: node.data.resource_type,
      severity: node.data.severity,
      resource_name: node.data.resource_name,
      metadata: node.data.metadata
    }
  }));

  // Convert edges
  const edges: Edge[] = cytoscapeElements.edges.map((edge) => ({
    id: edge.data.id,
    source: edge.data.source,
    target: edge.data.target,
    label: edge.data.label,
    type: 'smoothstep',
    animated: true,
    style: {
      stroke: '#64748b',
      strokeWidth: 2
    },
    labelStyle: {
      fontSize: 12,
      fontWeight: 500,
      fill: '#475569'
    },
    labelBgStyle: {
      fill: '#ffffff',
      fillOpacity: 0.9
    }
  }));

  return { nodes, edges };
};

/**
 * Highlight a path in the graph
 */
export const highlightPath = (
  nodes: Node[],
  edges: Edge[],
  path: string[]
): { nodes: Node[]; edges: Edge[] } => {
  const pathSet = new Set(path);

  const highlightedNodes = nodes.map(node => ({
    ...node,
    className: pathSet.has(node.id) ? 'highlighted' : '',
    style: pathSet.has(node.id) ? {
      boxShadow: '0 0 20px rgba(59, 130, 246, 0.6)',
    } : undefined
  }));

  const highlightedEdges = edges.map(edge => {
    const sourceIndex = path.indexOf(edge.source);
    const targetIndex = path.indexOf(edge.target);
    const isInPath = sourceIndex >= 0 && targetIndex === sourceIndex + 1;

    return {
      ...edge,
      animated: isInPath,
      style: {
        ...edge.style,
        stroke: isInPath ? '#3b82f6' : '#64748b',
        strokeWidth: isInPath ? 3 : 2
      }
    };
  });

  return { nodes: highlightedNodes, edges: highlightedEdges };
};
