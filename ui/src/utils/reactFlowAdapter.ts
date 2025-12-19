/**
 * Adapter to convert Cytoscape data format to React Flow format
 */

import dagre from '@dagrejs/dagre';
import type { Node, Edge } from 'reactflow';
import type { CytoscapeElements } from '../types/graph';

export interface ReactFlowData {
  nodes: Node[];
  edges: Edge[];
}

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 200;
const nodeHeight = 180;

/**
 * Apply Dagre layout to nodes
 */
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'TB') => {
  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 100,
    ranksep: 150,
    marginx: 50,
    marginy: 50
  });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

/**
 * Convert Cytoscape elements to React Flow format with Dagre layout
 */
export const convertToReactFlow = (cytoscapeElements: CytoscapeElements): ReactFlowData => {
  // Convert nodes
  const nodes: Node[] = cytoscapeElements.nodes.map((node) => ({
    id: node.data.id,
    type: 'custom',
    position: { x: 0, y: 0 }, // Will be set by Dagre
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
    animated: false,
    style: {
      stroke: '#64748b',
      strokeWidth: 2.5
    },
    labelStyle: {
      fontSize: 13,
      fontWeight: 600,
      fill: '#475569'
    },
    labelBgStyle: {
      fill: '#ffffff',
      fillOpacity: 0.95,
      rx: 4,
      ry: 4
    },
    markerEnd: {
      type: 'arrowclosed',
      color: '#64748b',
      width: 20,
      height: 20
    }
  }));

  // Apply Dagre layout
  return getLayoutedElements(nodes, edges, 'TB');
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
