/**
 * Events Mock Data Fixtures
 * Reusable mock data for events-related tests
 */

import type { FalcoEvent } from '../../api/types';
import type { PaginatedResponse } from '../../api/client';

export const mockFalcoEvent: FalcoEvent = {
  id: 'event-123',
  provider: 'aws',
  event_name: 'CreateRole',
  resource_type: 'aws_iam_role',
  resource_id: 'role-123',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDAI123',
    ARN: 'arn:aws:iam::123456789012:user/test',
    AccountID: '123456789012',
    UserName: 'test-user',
  },
  changes: { name: 'test-role' },
  region: 'us-east-1',
  project_id: 'project-123',
  service_name: 'iam',
};

export const mockPaginatedEvents: PaginatedResponse<FalcoEvent> = {
  data: [mockFalcoEvent],
  page: 1,
  limit: 10,
  total: 200,
  total_pages: 20,
};

/**
 * Creates a mock event with custom properties
 */
export const createMockEvent = (overrides: Partial<FalcoEvent> = {}): FalcoEvent => ({
  ...mockFalcoEvent,
  ...overrides,
});

/**
 * Creates a mock GCP event
 */
export const createMockGCPEvent = (overrides: Partial<FalcoEvent> = {}): FalcoEvent => ({
  ...mockFalcoEvent,
  provider: 'gcp',
  ...overrides,
});

/**
 * Creates paginated events with custom data
 */
export const createMockPaginatedEvents = (
  events: FalcoEvent[] = [mockFalcoEvent],
  page = 1,
  limit = 10,
  total = 200
): PaginatedResponse<FalcoEvent> => ({
  data: events,
  page,
  limit,
  total,
  total_pages: Math.ceil(total / limit),
});
