/**
 * Graph Clustering Utilities
 * Groups nodes by type, provider, severity, or custom criteria
 */

import type { Node, Edge } from 'reactflow';

export interface ClusterNode<T = any> extends Node<T> {
  type: 'cluster';
  data: T & {
    clusterType: string;
    clusterLabel: string;
    childNodeIds: string[];
    isExpanded: boolean;
    childCount: number;
    severityCounts?: Record<string, number>;
  };
}

export interface ClusterOptions {
  groupBy: 'type' | 'provider' | 'severity' | 'custom';
  customGroupFn?: (node: Node) => string;
  minClusterSize?: number; // Minimum nodes to form a cluster
  maxClusterSize?: number; // Maximum nodes per cluster (split if exceeded)
}

/**
 * Groups nodes into clusters based on specified criteria
 */
export function clusterNodes<T = any>(
  nodes: Node<T>[],
  options: ClusterOptions
): { nodes: (Node<T> | ClusterNode<T>)[]; clusterMap: Map<string, string[]> } {
  const {
    groupBy,
    customGroupFn,
    minClusterSize = 5,
    maxClusterSize = 50,
  } = options;

  // Group nodes by specified criteria
  const groups = new Map<string, Node<T>[]>();

  nodes.forEach(node => {
    let groupKey: string;
    const data = node.data as any;

    switch (groupBy) {
      case 'type':
        groupKey = data?.resource_type || data?.type || 'unknown';
        break;
      case 'provider':
        groupKey = extractProvider(data?.resource_type || data?.type || '');
        break;
      case 'severity':
        groupKey = data?.severity || 'unknown';
        break;
      case 'custom':
        groupKey = customGroupFn ? customGroupFn(node) : 'default';
        break;
      default:
        groupKey = 'default';
    }

    if (!groups.has(groupKey)) {
      groups.set(groupKey, []);
    }
    groups.get(groupKey)!.push(node);
  });

  // Create cluster nodes
  const clusterMap = new Map<string, string[]>();
  const resultNodes: (Node<T> | ClusterNode<T>)[] = [];

  groups.forEach((groupNodes, groupKey) => {
    // Don't cluster if group is too small
    if (groupNodes.length < minClusterSize) {
      resultNodes.push(...groupNodes);
      return;
    }

    // Split large groups into multiple clusters
    if (groupNodes.length > maxClusterSize) {
      const subClusters = chunkNodesForClustering(groupNodes, maxClusterSize);
      subClusters.forEach((subCluster, index) => {
        const clusterId = `cluster-${groupKey}-${index}`;
        const clusterNode = createClusterNode(
          clusterId,
          `${groupKey} (${index + 1})`,
          subCluster,
          groupKey
        );
        resultNodes.push(clusterNode);
        clusterMap.set(clusterId, subCluster.map(n => n.id));
      });
    } else {
      const clusterId = `cluster-${groupKey}`;
      const clusterNode = createClusterNode(
        clusterId,
        groupKey,
        groupNodes,
        groupKey
      );
      resultNodes.push(clusterNode);
      clusterMap.set(clusterId, groupNodes.map(n => n.id));
    }
  });

  return { nodes: resultNodes, clusterMap };
}

/**
 * Creates a cluster node from a group of nodes
 */
function createClusterNode<T = any>(
  clusterId: string,
  label: string,
  childNodes: Node<T>[],
  clusterType: string
): ClusterNode<T> {
  // Calculate cluster position (average of child positions)
  const avgPosition = calculateAveragePosition(childNodes);

  // Calculate severity counts
  const severityCounts: Record<string, number> = {};
  childNodes.forEach(node => {
    const data = node.data as any;
    const severity = data?.severity || 'unknown';
    severityCounts[severity] = (severityCounts[severity] || 0) + 1;
  });

  return {
    id: clusterId,
    type: 'cluster',
    position: avgPosition,
    data: {
      clusterType,
      clusterLabel: label,
      childNodeIds: childNodes.map(n => n.id),
      isExpanded: false,
      childCount: childNodes.length,
      severityCounts,
      label, // For compatibility
      type: 'cluster',
      resource_type: clusterType,
    } as any,
  };
}

/**
 * Expands a cluster to show its child nodes
 */
export function expandCluster<T = any>(
  clusterId: string,
  clusterMap: Map<string, string[]>,
  allNodes: Node<T>[],
  currentNodes: (Node<T> | ClusterNode<T>)[]
): (Node<T> | ClusterNode<T>)[] {
  const childNodeIds = clusterMap.get(clusterId);
  if (!childNodeIds) return currentNodes;

  // Find the cluster node
  const clusterIndex = currentNodes.findIndex(n => n.id === clusterId);
  if (clusterIndex === -1) return currentNodes;

  const clusterNode = currentNodes[clusterIndex] as ClusterNode<T>;

  // Get child nodes
  const childNodes = allNodes.filter(n => childNodeIds.includes(n.id));

  // Position child nodes in a circle around cluster position
  const positionedChildren = positionNodesInCircle(
    childNodes,
    clusterNode.position,
    150 // radius
  );

  // Mark cluster as expanded
  const expandedCluster: ClusterNode<T> = {
    ...clusterNode,
    data: {
      ...clusterNode.data,
      isExpanded: true,
    },
  };

  // Replace cluster with expanded cluster and child nodes
  return [
    ...currentNodes.slice(0, clusterIndex),
    expandedCluster,
    ...positionedChildren,
    ...currentNodes.slice(clusterIndex + 1),
  ];
}

/**
 * Collapses a cluster to hide its child nodes
 */
export function collapseCluster<T = any>(
  clusterId: string,
  clusterMap: Map<string, string[]>,
  currentNodes: (Node<T> | ClusterNode<T>)[]
): (Node<T> | ClusterNode<T>)[] {
  const childNodeIds = clusterMap.get(clusterId);
  if (!childNodeIds) return currentNodes;

  // Remove child nodes
  const filteredNodes = currentNodes.filter(n => !childNodeIds.includes(n.id));

  // Mark cluster as collapsed
  return filteredNodes.map(n => {
    if (n.id === clusterId) {
      return {
        ...n,
        data: {
          ...(n as ClusterNode<T>).data,
          isExpanded: false,
        },
      } as ClusterNode<T>;
    }
    return n;
  });
}

/**
 * Filters edges to handle cluster expansion/collapse
 */
export function filterEdgesForClusters(
  edges: Edge[],
  clusterMap: Map<string, string[]>,
  visibleNodeIds: Set<string>
): Edge[] {
  return edges
    .map(edge => {
      let source = edge.source;
      let target = edge.target;

      // If source is collapsed in a cluster, redirect to cluster
      if (!visibleNodeIds.has(source)) {
        const clusterId = findNodeCluster(source, clusterMap, visibleNodeIds);
        if (clusterId) source = clusterId;
      }

      // If target is collapsed in a cluster, redirect to cluster
      if (!visibleNodeIds.has(target)) {
        const clusterId = findNodeCluster(target, clusterMap, visibleNodeIds);
        if (clusterId) target = clusterId;
      }

      // Don't show edges within the same cluster
      if (source === target) return null;

      return source !== edge.source || target !== edge.target
        ? { ...edge, id: `${source}-${target}`, source, target }
        : edge;
    })
    .filter((edge): edge is Edge => edge !== null);
}

/**
 * Finds which cluster contains a node
 */
function findNodeCluster(
  nodeId: string,
  clusterMap: Map<string, string[]>,
  visibleClusterIds: Set<string>
): string | null {
  for (const [clusterId, childIds] of clusterMap.entries()) {
    if (childIds.includes(nodeId) && visibleClusterIds.has(clusterId)) {
      return clusterId;
    }
  }
  return null;
}

/**
 * Extracts provider from resource type (e.g., 'aws_s3_bucket' -> 'aws')
 */
function extractProvider(resourceType: string): string {
  const match = resourceType.match(/^(aws|gcp|azure|kubernetes|k8s)_/);
  return match ? match[1] : 'other';
}

/**
 * Calculates average position of nodes
 */
function calculateAveragePosition(nodes: Node[]): { x: number; y: number } {
  if (nodes.length === 0) return { x: 0, y: 0 };

  const sum = nodes.reduce(
    (acc, node) => ({
      x: acc.x + node.position.x,
      y: acc.y + node.position.y,
    }),
    { x: 0, y: 0 }
  );

  return {
    x: sum.x / nodes.length,
    y: sum.y / nodes.length,
  };
}

/**
 * Positions nodes in a circle around a center point
 */
function positionNodesInCircle<T = any>(
  nodes: Node<T>[],
  center: { x: number; y: number },
  radius: number
): Node<T>[] {
  const angleStep = (2 * Math.PI) / nodes.length;

  return nodes.map((node, index) => {
    const angle = index * angleStep;
    return {
      ...node,
      position: {
        x: center.x + radius * Math.cos(angle),
        y: center.y + radius * Math.sin(angle),
      },
    };
  });
}

/**
 * Chunks nodes for clustering
 */
function chunkNodesForClustering<T = any>(
  nodes: Node<T>[],
  chunkSize: number
): Node<T>[][] {
  const chunks: Node<T>[][] = [];
  for (let i = 0; i < nodes.length; i += chunkSize) {
    chunks.push(nodes.slice(i, i + chunkSize));
  }
  return chunks;
}

/**
 * Hook for managing cluster state
 */
import { useState, useCallback, useMemo } from 'react';

export function useGraphClustering<T = any>(
  allNodes: Node<T>[],
  allEdges: Edge[],
  options: ClusterOptions
) {
  const [expandedClusters, setExpandedClusters] = useState<Set<string>>(new Set());

  // Initial clustering
  const { nodes: clusteredNodes, clusterMap } = useMemo(() => {
    return clusterNodes(allNodes, options);
  }, [allNodes, options]);

  // Get visible nodes (including expanded cluster children)
  const visibleNodes = useMemo(() => {
    let nodes = [...clusteredNodes];

    expandedClusters.forEach(clusterId => {
      nodes = expandCluster(clusterId, clusterMap, allNodes, nodes);
    });

    return nodes;
  }, [clusteredNodes, clusterMap, allNodes, expandedClusters]);

  // Get visible node IDs
  const visibleNodeIds = useMemo(() => {
    return new Set(visibleNodes.map(n => n.id));
  }, [visibleNodes]);

  // Filter edges for visible nodes
  const visibleEdges = useMemo(() => {
    return filterEdgesForClusters(allEdges, clusterMap, visibleNodeIds);
  }, [allEdges, clusterMap, visibleNodeIds]);

  // Toggle cluster expansion
  const toggleCluster = useCallback((clusterId: string) => {
    setExpandedClusters(prev => {
      const next = new Set(prev);
      if (next.has(clusterId)) {
        next.delete(clusterId);
      } else {
        next.add(clusterId);
      }
      return next;
    });
  }, []);

  // Expand all clusters
  const expandAll = useCallback(() => {
    const allClusterIds = Array.from(clusterMap.keys());
    setExpandedClusters(new Set(allClusterIds));
  }, [clusterMap]);

  // Collapse all clusters
  const collapseAll = useCallback(() => {
    setExpandedClusters(new Set());
  }, []);

  return {
    visibleNodes,
    visibleEdges,
    clusterMap,
    expandedClusters,
    toggleCluster,
    expandAll,
    collapseAll,
  };
}
