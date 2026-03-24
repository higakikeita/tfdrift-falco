// API Type Definitions

import type { UserIdentityAPI } from '../types/common';
import type { CytoscapeNode, CytoscapeEdge, CytoscapeElements } from '../types/graph';

export type { CytoscapeNode, CytoscapeEdge, CytoscapeElements };

/** @deprecated Use UserIdentityAPI from types/common instead */
export type UserIdentity = UserIdentityAPI;

export interface DriftAlert {
  id: string;
  severity: string;
  resource_type: string;
  resource_name: string;
  resource_id: string;
  attribute: string;
  old_value: string;
  new_value: string;
  user_identity: UserIdentity;
  matched_rules: string[];
  timestamp: string;
  alert_type: string;
}

export type EventStatus = 'open' | 'acknowledged' | 'ignored' | 'resolved';

export interface FalcoEvent {
  id: string;
  provider: string;
  event_name: string;
  resource_type: string;
  resource_id: string;
  user_identity: UserIdentity;
  changes: Record<string, unknown>;
  region: string;
  project_id: string;
  service_name: string;
  timestamp: string;
  severity: string;
  status: EventStatus;
  status_reason: string;
  raw_event?: unknown;
  related_drifts?: RelatedDrift[];
}

export interface RelatedDrift {
  severity: string;
  attribute: string;
  old_value: unknown;
  new_value: unknown;
  matched_rules: string[];
  timestamp: string;
  alert_type: string;
}

export interface StateMetadata {
  version: number;
  terraform_version: string;
  serial: number;
  lineage: string;
  resource_count: number;
  outputs_count: number;
}

export interface TerraformResource {
  type: string;
  name: string;
  provider: string;
  mode: string;
  attributes: Record<string, unknown>;
}

export interface Stats {
  graph: {
    total_nodes: number;
    total_edges: number;
  };
  drifts: {
    total: number;
    severity_counts: Record<string, number>;
    resource_types: Record<string, number>;
  };
  events: {
    total: number;
  };
  unmanaged: {
    total: number;
  };
  severity_breakdown: Record<string, number>;
  top_resource_types: Array<{
    resource_type: string;
    count: number;
  }>;
}

// AWS Discovery Types
export interface DiscoveredResource {
  id: string;
  type: string;
  arn?: string;
  name: string;
  region: string;
  attributes: Record<string, unknown>;
  tags?: Record<string, string>;
}

export interface FieldDiff {
  field: string;
  terraform_value: unknown;
  actual_value: unknown;
}

export interface ResourceDiff {
  resource_id: string;
  resource_type: string;
  terraform_state: Record<string, unknown>;
  actual_state: Record<string, unknown>;
  differences: FieldDiff[];
}

export interface DriftResult {
  unmanaged_resources: DiscoveredResource[];  // Resources in AWS but not in Terraform (manually created)
  missing_resources: TerraformResource[];      // Resources in Terraform but not in AWS (manually deleted)
  modified_resources: ResourceDiff[];          // Resources with configuration differences
}

export interface DriftSummary {
  region: string;
  timestamp: string;
  counts: {
    terraform_resources: number;
    aws_resources: number;
    unmanaged: number;
    missing: number;
    modified: number;
  };
  breakdown: {
    unmanaged_by_type: Record<string, number>;
    missing_by_type: Record<string, number>;
    modified_by_type: Record<string, number>;
  };
}

export interface DriftDetectionResult {
  region: string;
  timestamp: string;
  summary: {
    terraform_resources: number;
    aws_resources: number;
    unmanaged_count: number;
    missing_count: number;
    modified_count: number;
  };
  drift: DriftResult;
}
