/**
 * Common Types - Shared across the application
 * Single source of truth for commonly used types
 */

// ==================== Severity ====================

export type Severity = 'critical' | 'high' | 'medium' | 'low';

// ==================== User Identity ====================

/**
 * Unified user identity across AWS/GCP/Azure
 * Backend APIから返される形式（PascalCase）
 */
export interface UserIdentityAPI {
  Type: string;
  PrincipalID: string;
  ARN: string;
  AccountID: string;
  UserName: string;
}

/**
 * Frontend用のユーザーアイデンティティ（camelCase）
 */
export interface UserIdentity {
  type: string;
  userName: string;
  arn?: string;
  accountId?: string;
  principalId?: string;
}

/**
 * API形式からフロントエンド形式への変換
 */
export function convertUserIdentity(api: UserIdentityAPI): UserIdentity {
  return {
    type: api.Type,
    userName: api.UserName,
    arn: api.ARN,
    accountId: api.AccountID,
    principalId: api.PrincipalID
  };
}

// ==================== Provider ====================

export type Provider = 'aws' | 'gcp' | 'azure';

// ==================== Change Type ====================

export type ChangeType = 'created' | 'modified' | 'deleted';

// ==================== Metadata ====================

/**
 * Generic metadata container
 */
export type Metadata = Record<string, unknown>;

// ==================== Timestamp ====================

/**
 * ISO 8601 timestamp string
 */
export type Timestamp = string;

// ==================== Resource Info ====================

/**
 * Common resource information
 */
export interface ResourceInfo {
  type: string;
  id: string;
  name?: string;
  provider?: Provider;
  region?: string;
}

// ==================== Filters ====================

/**
 * Date range filter
 */
export interface DateRangeFilter {
  start: string;
  end: string;
}

/**
 * Common filter interface
 */
export interface BaseFilters {
  search?: string;
  dateRange?: DateRangeFilter;
}
