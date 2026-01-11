/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * TFDrift-Falco Graph Types
 *
 * 因果関係グラフのノード・エッジ・グラフ全体の型定義
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
  NETWORK: 'network'
} as const;

export type NodeType = typeof NodeType[keyof typeof NodeType];

export const EdgeType = {
  CAUSED_BY: 'caused_by',        // Drift → IAM change
  GRANTS_ACCESS: 'grants_access', // IAM → ServiceAccount
  USED_BY: 'used_by',            // SA → Pod
  CONTAINS: 'contains',          // Pod → Container
  TRIGGERED: 'triggered'         // Container → Falco Event
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
    metadata: Record<string, any>;
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
    metadata: Record<string, any>;
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

// Cytoscape.js用の型
export interface CytoscapeNode {
  data: {
    id: string;
    label: string;
    type: NodeType;
    severity?: Severity;
    resource_type: string;
    resource_name: string;
    metadata: Record<string, any>;
  };
}

export interface CytoscapeEdge {
  data: {
    id: string;
    source: string;
    target: string;
    label: string;
    type: EdgeType;
    relationship: string;
  };
}

export interface CytoscapeElements {
  nodes: CytoscapeNode[];
  edges: CytoscapeEdge[];
}

// ビューモード
export const ViewMode = {
  ATTACK_VIEW: 'attack_view',   // 攻撃視点
  OPS_VIEW: 'ops_view'          // 運用視点
} as const;

export type ViewMode = typeof ViewMode[keyof typeof ViewMode];

// レイアウトタイプ
export const LayoutType = {
  HIERARCHICAL: 'dagre',
  RADIAL: 'concentric',
  FORCE: 'cose',
  GRID: 'grid',
  NETWORK_DIAGRAM: 'network-diagram'
} as const;

export type LayoutType = typeof LayoutType[keyof typeof LayoutType];
