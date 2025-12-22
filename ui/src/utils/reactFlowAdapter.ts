/**
 * Adapter to convert Cytoscape data format to React Flow format
 */

import dagre from '@dagrejs/dagre';
import { MarkerType } from 'reactflow';
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
 * Layout child nodes within their parent containers
 */
const layoutChildNodes = (childNodes: Node[]): Node[] => {
  // Group children by parent
  const childrenByParent: Record<string, Node[]> = {};
  childNodes.forEach(node => {
    if (node.parentNode) {
      if (!childrenByParent[node.parentNode]) {
        childrenByParent[node.parentNode] = [];
      }
      childrenByParent[node.parentNode].push(node);
    }
  });

  // Position children within each parent
  const layoutedChildren: Node[] = [];

  Object.entries(childrenByParent).forEach(([_parentId, children]) => {
    const level = children[0]?.data?.level;

    if (level === 'vpc') {
      // VPCs are positioned at top-left of region
      children.forEach((child, index) => {
        layoutedChildren.push({
          ...child,
          position: { x: 20, y: 60 + index * 520 }
        });
      });
    } else if (level === 'az') {
      // AZs are positioned horizontally within VPC
      children.forEach((child, index) => {
        layoutedChildren.push({
          ...child,
          position: { x: 20 + index * 320, y: 60 }
        });
      });
    } else if (level === 'subnet') {
      // Subnets are positioned vertically within AZ
      children.forEach((child, index) => {
        layoutedChildren.push({
          ...child,
          position: { x: 10, y: 40 + index * 220 }
        });
      });
    } else {
      // Regular resources within subnets
      children.forEach((child, index) => {
        const row = Math.floor(index / 2);
        const col = index % 2;
        layoutedChildren.push({
          ...child,
          position: { x: 10 + col * 110, y: 40 + row * 100 }
        });
      });
    }
  });

  return layoutedChildren;
};

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
  const nodes: Node[] = cytoscapeElements.nodes.map((node) => {
    const parentNode = node.data.metadata?.parent_node;
    const level = node.data.metadata?.level;

    // Determine node type based on level or existing type
    let nodeType = 'custom';
    const dataType = String(node.data.type || '');
    if (level === 'region' || dataType === 'region-group') {
      nodeType = 'region-group';
    } else if (level === 'vpc' || dataType === 'vpc-group') {
      nodeType = 'vpc-group';
    } else if (level === 'az' || dataType === 'az-group') {
      nodeType = 'az-group';
    } else if (level === 'subnet' || dataType.startsWith('subnet-group')) {
      nodeType = dataType; // Use the specific subnet type (public/private)
    }

    const reactFlowNode: Node = {
      id: node.data.id,
      type: nodeType,
      position: { x: 0, y: 0 }, // Will be set by layout or manually
      data: {
        label: node.data.label,
        type: node.data.type,
        resource_type: node.data.resource_type,
        severity: node.data.severity,
        resource_name: node.data.resource_name,
        metadata: node.data.metadata,
        level: level
      }
    };

    // Set parent node and extent for child nodes
    if (parentNode) {
      reactFlowNode.parentNode = parentNode;
      reactFlowNode.extent = 'parent';
    }

    return reactFlowNode;
  });

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
      type: MarkerType.ArrowClosed,
      color: '#64748b'
    }
  }));

  // Separate parent nodes from child nodes for layout
  const parentNodes = nodes.filter(n => !n.parentNode);
  const childNodes = nodes.filter(n => n.parentNode);

  // Apply Dagre layout only to parent nodes and non-hierarchical nodes
  const layouted = getLayoutedElements(parentNodes, edges, 'TB');

  // Apply manual layout for hierarchical nodes
  const layoutedChildNodes = layoutChildNodes(childNodes);

  return {
    nodes: [...layouted.nodes, ...layoutedChildNodes],
    edges: layouted.edges
  };
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
