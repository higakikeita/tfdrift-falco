/**
 * driftwire Graph Types
 *
 * Node/edge data types and graph definition types
 */

export const NodeType = {
  TERRAFORM_CHANGE: 'terraform_change',
  IAM_POLICY: 'iam_policy',
  IAM_ROLE: 'iam_role',
  SERVICE_ACCOUNT: 'service_account',
  POD: 'pod',
  CONTAINER: 'container',
  FALCO_EVENT: 'falco_event',
  SECURITY_GROUP: 'security_group',
  NETWORK: 'network',
  // Infrastructure topology node types
  REGION: 'region',
  VPC: 'vpc',
  AVAILABILITY_ZONE: 'availability_zone',
  SUBNET: 'subnet',
  EC2_INSTANCE: 'ec2_instance',
  RDS_INSTANCE: 'rds_instance',
  LAMBDA_FUNCTION: 'lambda_function',
  LOAD_BALANCER: 'load_balancer',
  NAT_GATEWAY: 'nat_gateway',
} as const;

export type NodeType = typeof NodeType[keyof typeof NodeType];

export const EdgeType = {
  CAUSED_BY: 'caused_by',        // Drift â IAM change
  GRANTS_ACCESS: 'grants_access', // IAM â ServiceAccount
  USED_BY: 'used_by',            // SA â Pod
  CONTAINS: 'contains',          // Pod â Container
  TRIGGERED: 'triggered'         // Container â Falco Event
} as const;

export type EdgeType = typeof EdgeType[keyof typeof EdgeType];

export type Severity = 'critical' | 'high' | 'medium' | 'low';

export interface GraphNode {
  id: string;
  type: NodeType;
  label: string;
  data: {
    resource_type: string;
    resource_name: string;
    timestamp?: string;
    severity?: Severity;
    metadata: Record<string, unknown>;
  };
  style?: {
    color?: string;
    size?: number;
    shape?: string;
  };
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  type: EdgeType;
  label: string;
  data: {
    relationship: string;
    metadata: Record<string, unknown>;
  };
  style?: {
    color?: string;
    width?: number;
    line_style?: 'solid' | 'dashed' | 'dotted';
  };
}

export interface CausalGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
  metadata: {
    generated_at: string;
    root_cause?: string;
    blast_radius_size?: number;
  };
}

export interface PathResult {
  paths: Array<{
    length: number;
    nodes: string[];
    edges: GraphEdge[];
  }>;
}

export interface BlastRadiusResult {
  center: {
    id: string;
    type: NodeType;
  };
  affected_resources: Array<{
    id: string;
    type: NodeType;
    distance: number;
  }>;
  graph: CausalGraph;
}

// Cytoscape.jsç¨ã®å
export interface CytoscapeNode {
  data: {
    id: string;
    label: string;
    type?: NodeType;
    severity?: Severity;
    resource_type: string;
    resource_name?: string;
    metadata?: Record<string, unknown>;
    parent?: string;
    [key: string]: unknown;
  };
  position?: {
    x: number;
    y: number;
  };
}

export interface CytoscapeEdge {
  data: {
    id: string;
    source: string;
    target: string;
    label: string;
    type?: EdgeType;
    relationship?: string;
    [key: string]: unknown;
  };
}

export interface CytoscapeElements {
  nodes: CytoscapeNode[];
  edges: CytoscapeEdge[];
}

// ãã¥ã¼ã¢ã¼ã
export const ViewMode = {
  ATTACK_VIEW: 'attack_view',   // æ»æè¦ç¹
  OPS_VIEW: 'ops_view'          // éç¨è¦ç¹
} as const;

export type ViewMode = typeof ViewMode[keyof typeof ViewMode];

// ã¬ã¤ã¢ã¦ãã¿ã¤ã
export const LayoutType = {
  HIERARCHICAL: 'dagre',
  RADIAL: 'concentric',
  FORCE: 'cose',
  GRID: 'grid',
  NETWORK_DIAGRAM: 'network-diagram'
} as const;

export type LayoutType = typeof LayoutType[keyof typeof LayoutType];
