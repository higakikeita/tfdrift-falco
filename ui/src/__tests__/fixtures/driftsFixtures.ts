/**
 * Drifts Mock Data Fixtures
 * Reusable mock data for drifts-related tests
 */

import type { DriftAlert } from '../../api/types';
import type { PaginatedResponse } from '../../api/client';

export const mockDriftAlert: DriftAlert = {
  id: 'drift-123',
  severity: 'high',
  resource_type: 'aws_iam_role',
  resource_name: 'test-role',
  resource_id: 'role-123',
  attribute: 'assume_role_policy',
  old_value: '{}',
  new_value: '{"Version": "2012-10-17"}',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDAI123',
    ARN: 'arn:aws:iam::123456789012:user/test',
    AccountID: '123456789012',
    UserName: 'test-user',
  },
  matched_rules: ['rule1'],
  timestamp: '2024-01-01T00:00:00Z',
  alert_type: 'drift',
};

export const mockPaginatedDrifts: PaginatedResponse<DriftAlert> = {
  data: [mockDriftAlert],
  page: 1,
  limit: 10,
  total: 50,
  total_pages: 5,
};

/**
 * Creates a mock drift alert with custom properties
 */
export const createMockDrift = (overrides: Partial<DriftAlert> = {}): DriftAlert => ({
  ...mockDriftAlert,
  ...overrides,
});

/**
 * Creates paginated drifts with custom data
 */
export const createMockPaginatedDrifts = (
  drifts: DriftAlert[] = [mockDriftAlert],
  page = 1,
  limit = 10,
  total = 50
): PaginatedResponse<DriftAlert> => ({
  data: drifts,
  page,
  limit,
  total,
  total_pages: Math.ceil(total / limit),
});
