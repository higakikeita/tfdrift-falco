/**
 * Convert API graph data to Cytoscape format
 */

import type { CytoscapeElements, CytoscapeNode, CytoscapeEdge } from '../types/graph';

interface APINode {
  id: string;
  labels: string[];
  properties: {
    id: string;
    type: string;
    name: string;
    has_drift?: boolean;
    [key: string]: unknown;
  };
}

interface APIEdge {
  id: string;
  type: string;
  start_node: string;
  end_node: string;
  properties?: {
    [key: string]: unknown;
  };
}

interface APIGraphData {
  nodes: APINode[];
  edges: APIEdge[];
}

export function convertAPIGraphToCytoscape(apiData: APIGraphData, driftedResources?: Set<string>): CytoscapeElements {
  const nodes: CytoscapeNode[] = apiData.nodes.map(node => {
    const props = node.properties || {};
    const hasDrift = driftedResources?.has(node.id) || props.has_drift || false;

    return {
      data: {
        id: node.id,
        label: props.name || node.id,
        type: mapResourceTypeToNodeType(props.type || 'unknown') as any,
        severity: hasDrift ? ('high' as const) : undefined,
        resource_type: props.type || 'unknown',
        resource_name: props.name || node.id,
        metadata: {
          ...props,
          has_drift: hasDrift,
          labels: node.labels || [],
        },
      },
    } as CytoscapeNode;
  });

  // Create a set of valid node IDs for edge validation
  const nodeIds = new Set(nodes.map(n => n.data.id));

  // Filter edges to only include those with valid source and target nodes
  const edges: CytoscapeEdge[] = apiData.edges
    .filter(edge => {
      const hasSource = edge.start_node && nodeIds.has(edge.start_node);
      const hasTarget = edge.end_node && nodeIds.has(edge.end_node);
      if (!hasSource || !hasTarget) {
        console.warn(`Skipping edge ${edge.id}: invalid source (${edge.start_node}) or target (${edge.end_node})`);
        return false;
      }
      return true;
    })
    .map(edge => ({
      data: {
        id: edge.id,
        source: edge.start_node,
        target: edge.end_node,
        label: formatEdgeLabel(edge.type || 'unknown'),
        type: mapEdgeType(edge.type || 'unknown') as any,
        relationship: edge.type || 'unknown',
      },
    } as CytoscapeEdge));

  return { nodes, edges };
}

function mapResourceTypeToNodeType(resourceType: string): string {
  // Map AWS resource types to node types
  const typeMap: Record<string, string> = {
    'aws_security_group': 'security_group',
    'aws_iam_role': 'iam_role',
    'aws_iam_policy': 'iam_policy',
    'aws_instance': 'terraform_change',
    'aws_subnet': 'network',
    'aws_vpc': 'network',
    'aws_eks_cluster': 'terraform_change',
    'aws_db_instance': 'terraform_change',
    'aws_lb': 'network',
  };

  return typeMap[resourceType] || 'terraform_change';
}

function mapEdgeType(edgeType: string): string {
  // Map relationship types to edge types
  const typeMap: Record<string, string> = {
    'DEPENDS_ON': 'caused_by',
    'PART_OF': 'contains',
    'SECURES': 'grants_access',
    'RUNS_IN': 'contains',
    'REGISTERS_TO': 'used_by',
  };

  return typeMap[edgeType] || 'caused_by';
}

function formatEdgeLabel(edgeType: string): string {
  return edgeType.toLowerCase().replace(/_/g, ' ');
}
