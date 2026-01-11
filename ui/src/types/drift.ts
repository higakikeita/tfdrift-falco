/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * Drift Event Types
 * TFDrift-Falcoで検知されたドリフトイベントの型定義
 */

export type DriftSeverity = 'critical' | 'high' | 'medium' | 'low';

export type ChangeType = 'created' | 'modified' | 'deleted';

export type Provider = 'aws' | 'gcp' | 'azure';

/**
 * Drift Change - Individual change within a drift
 */
export interface DriftChange {
  path: string;
  before: unknown;
  after: unknown;
  change_type: string;
}

/**
 * Drift - Drift detection result
 */
export interface Drift {
  id: string;
  resource_type: string;
  resource_name: string;
  resource_id: string;
  provider: string;
  region: string;
  change_type: string;
  severity: string;
  detected_at: string;
  changes: DriftChange[];
  metadata: Record<string, unknown>;
}

export interface UserIdentity {
  type: string;
  userName: string;
  arn?: string;
  accountId?: string;
  principalId?: string;
}

export interface DriftEvent {
  id: string;
  timestamp: string;
  severity: DriftSeverity;
  provider: Provider;
  resourceType: string;
  resourceId: string;
  resourceName?: string;
  changeType: ChangeType;
  attribute: string;
  oldValue: string | null;
  newValue: string | null;
  userIdentity: UserIdentity;
  region: string;
  cloudtrailEventId?: string;
  cloudtrailEventName?: string;
  sourceIP?: string;
  userAgent?: string;
  tags?: Record<string, string>;
  metadata?: Record<string, any>;
}

export interface DriftStats {
  total: number;
  critical: number;
  high: number;
  medium: number;
  low: number;
  byProvider: Record<Provider, number>;
  byResourceType: Record<string, number>;
  recentCount: number; // Last 24h
}

export interface DriftFilters {
  severity?: DriftSeverity[];
  provider?: Provider[];
  resourceType?: string[];
  changeType?: ChangeType[];
  dateRange?: {
    start: string;
    end: string;
  };
  search?: string;
}
