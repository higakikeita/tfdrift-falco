/**
 * Drift Event Types
 * TFDrift-Falcoで検知されたドリフトイベントの型定義
 */

export type DriftSeverity = 'critical' | 'high' | 'medium' | 'low';

export type ChangeType = 'created' | 'modified' | 'deleted';

export type Provider = 'aws' | 'gcp' | 'azure';

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
