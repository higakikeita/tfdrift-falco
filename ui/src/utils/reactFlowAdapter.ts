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
 * Apply force-directed layout
 */
const getForceLayout = (nodes: Node[], edges: Edge[]): ReactFlowData => {
  // Simple force-directed simulation
  const positions = new Map<string, { x: number; y: number }>();

  // Initialize random positions
  nodes.forEach((node) => {
    positions.set(node.id, {
      x: Math.random() * 1200 - 600,
      y: Math.random() * 800 - 400
    });
  });

  // Run simulation iterations
  for (let i = 0; i < 100; i++) {
    const forces = new Map<string, { x: number; y: number }>();
    nodes.forEach(node => forces.set(node.id, { x: 0, y: 0 }));

    // Repulsive forces between all nodes
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const pos1 = positions.get(nodes[i].id)!;
        const pos2 = positions.get(nodes[j].id)!;
        const dx = pos2.x - pos1.x;
        const dy = pos2.y - pos1.y;
        const dist = Math.sqrt(dx * dx + dy * dy) || 1;
        const force = 10000 / (dist * dist);

        const fx = (dx / dist) * force;
        const fy = (dy / dist) * force;

        const f1 = forces.get(nodes[i].id)!;
        const f2 = forces.get(nodes[j].id)!;
        f1.x -= fx;
        f1.y -= fy;
        f2.x += fx;
        f2.y += fy;
      }
    }

    // Attractive forces for edges
    edges.forEach(edge => {
      const pos1 = positions.get(edge.source)!;
      const pos2 = positions.get(edge.target)!;
      const dx = pos2.x - pos1.x;
      const dy = pos2.y - pos1.y;
      const dist = Math.sqrt(dx * dx + dy * dy) || 1;
      const force = dist * 0.01;

      const fx = (dx / dist) * force;
      const fy = (dy / dist) * force;

      const f1 = forces.get(edge.source)!;
      const f2 = forces.get(edge.target)!;
      f1.x += fx;
      f1.y += fy;
      f2.x -= fx;
      f2.y -= fy;
    });

    // Apply forces
    nodes.forEach(node => {
      const pos = positions.get(node.id)!;
      const force = forces.get(node.id)!;
      pos.x += force.x * 0.1;
      pos.y += force.y * 0.1;
    });
  }

  // Apply positions
  const layoutedNodes = nodes.map(node => ({
    ...node,
    position: positions.get(node.id)!
  }));

  return { nodes: layoutedNodes, edges };
};

/**
 * Apply radial/concentric layout
 */
const getRadialLayout = (nodes: Node[], edges: Edge[]): ReactFlowData => {
  // Find center node (most connected)
  const degreeMap = new Map<string, number>();
  nodes.forEach(node => degreeMap.set(node.id, 0));

  edges.forEach(edge => {
    degreeMap.set(edge.source, (degreeMap.get(edge.source) || 0) + 1);
    degreeMap.set(edge.target, (degreeMap.get(edge.target) || 0) + 1);
  });

  const sortedNodes = [...nodes].sort((a, b) =>
    (degreeMap.get(b.id) || 0) - (degreeMap.get(a.id) || 0)
  );

  const layoutedNodes = sortedNodes.map((node, index) => {
    const level = Math.floor(Math.sqrt(index));
    const itemsInLevel = Math.max(1, level * 6);
    const angleStep = (2 * Math.PI) / itemsInLevel;
    const radius = level * 250;
    const indexInLevel = index - (level * level);
    const angle = angleStep * indexInLevel;

    return {
      ...node,
      position: {
        x: Math.cos(angle) * radius,
        y: Math.sin(angle) * radius
      }
    };
  });

  return { nodes: layoutedNodes, edges };
};

/**
 * Apply grid layout
 */
const getGridLayout = (nodes: Node[], edges: Edge[]): ReactFlowData => {
  const cols = Math.ceil(Math.sqrt(nodes.length));
  const spacing = 300;

  const layoutedNodes = nodes.map((node, index) => ({
    ...node,
    position: {
      x: (index % cols) * spacing,
      y: Math.floor(index / cols) * spacing
    }
  }));

  return { nodes: layoutedNodes, edges };
};

/**
 * Apply hierarchical network diagram layout (AWS-style)
 * Uses React Flow parent-child nodes for grouping
 */
const getHierarchicalNetworkLayout = (nodes: Node[], edges: Edge[]): ReactFlowData => {
  // Separate nodes by hierarchical level
  const regionNodes: Node[] = [];
  const vpcNodes: Node[] = [];
  const azNodes: Node[] = [];
  const subnetNodes: Node[] = [];
  const resourceNodes: Node[] = [];

  nodes.forEach(node => {
    const level = (node.data as any).metadata?.hierarchical_level;
    switch (level) {
      case 'region': regionNodes.push(node); break;
      case 'vpc': vpcNodes.push(node); break;
      case 'az': azNodes.push(node); break;
      case 'subnet': subnetNodes.push(node); break;
      case 'resource': resourceNodes.push(node); break;
      default: resourceNodes.push(node); break;
    }
  });

  const layoutedNodes: Node[] = [];

  // Region nodes (Level 1)
  regionNodes.forEach((node) => {
    layoutedNodes.push({
      ...node,
      type: 'region-group',
      position: { x: 50, y: 50 },
      style: {
        width: 1950,
        height: 1300,
        zIndex: 1
      },
      data: {
        ...node.data,
        level: 'region'
      }
    });
  });

  // VPC nodes (Level 2) - children of regions
  vpcNodes.forEach((node) => {
    const parent = (node.data as any).metadata?.parent;
    layoutedNodes.push({
      ...node,
      type: 'vpc-group',
      parentNode: parent,
      extent: 'parent' as any,
      position: { x: 60, y: 90 },
      style: {
        width: 1820,
        height: 1120,
        zIndex: 2
      },
      data: {
        ...node.data,
        level: 'vpc'
      }
    });
  });

  // AZ nodes (Level 3) - children of VPCs
  azNodes.forEach((node, idx) => {
    const parent = (node.data as any).metadata?.parent;
    const xOffset = idx * 900; // Increased spacing between AZs
    layoutedNodes.push({
      ...node,
      type: 'az-group',
      parentNode: parent,
      extent: 'parent' as any,
      position: { x: 30 + xOffset, y: 70 },
      style: {
        width: 850,
        height: 1000,
        zIndex: 3
      },
      data: {
        ...node.data,
        level: 'az'
      }
    });
  });

  // Subnet nodes (Level 4) - children of AZs
  const subnetsByParent = new Map<string, Node[]>();
  subnetNodes.forEach(node => {
    const parent = (node.data as any).metadata?.parent;
    if (!subnetsByParent.has(parent)) {
      subnetsByParent.set(parent, []);
    }
    subnetsByParent.get(parent)!.push(node);
  });

  subnetsByParent.forEach((subnets, parentId) => {
    subnets.forEach((node, idx) => {
      const subnetType = (node.data as any).metadata?.subnet_type;
      const yOffset = idx * 490; // Increased spacing between subnets
      layoutedNodes.push({
        ...node,
        type: subnetType === 'public' ? 'subnet-group-public' : 'subnet-group-private',
        parentNode: parentId,
        extent: 'parent' as any,
        position: { x: 25, y: 60 + yOffset },
        style: {
          width: 790,
          height: 420,
          zIndex: 4
        },
        data: {
          ...node.data,
          level: 'subnet'
        }
      });
    });
  });

  // Resource nodes (Level 5) - children of subnets
  const resourcesByParent = new Map<string, Node[]>();
  resourceNodes.forEach(node => {
    const parent = (node.data as any).metadata?.parent;
    if (!resourcesByParent.has(parent)) {
      resourcesByParent.set(parent, []);
    }
    resourcesByParent.get(parent)!.push(node);
  });

  resourcesByParent.forEach((resources, parentId) => {
    resources.forEach((node, idx) => {
      const col = idx % 3;
      const row = Math.floor(idx / 3);
      layoutedNodes.push({
        ...node,
        type: 'custom',
        parentNode: parentId,
        extent: 'parent' as any,
        position: { x: 30 + col * 250, y: 60 + row * 140 }, // More spacing between resources
        style: {
          width: 220,
          height: 110,
          zIndex: 5
        },
        data: {
          ...node.data
        }
      });
    });
  });

  return { nodes: layoutedNodes, edges };
};

/**
 * Convert Cytoscape elements to React Flow format with specified layout
 */
export const convertToReactFlow = (
  cytoscapeElements: CytoscapeElements,
  layout: 'dagre' | 'cose' | 'concentric' | 'grid' | 'network-diagram' = 'dagre'
): ReactFlowData => {
  // Convert nodes - all as custom nodes (flat graph)
  const nodes: Node[] = cytoscapeElements.nodes.map((node) => ({
    id: node.data.id,
    type: 'custom',
    position: { x: 0, y: 0 }, // Will be set by layout algorithm
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
      type: MarkerType.ArrowClosed,
      color: '#64748b'
    }
  }));

  // Apply selected layout algorithm
  switch (layout) {
    case 'network-diagram':
      return getHierarchicalNetworkLayout(nodes, edges);
    case 'cose':
      return getForceLayout(nodes, edges);
    case 'concentric':
      return getRadialLayout(nodes, edges);
    case 'grid':
      return getGridLayout(nodes, edges);
    case 'dagre':
    default:
      return getLayoutedElements(nodes, edges, 'TB');
  }
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
