import type { Meta, StoryObj } from '@storybook/react';
import { EventDetailPanel } from './EventDetailPanel';
import type { FalcoEvent } from '../../api/types';

const meta: Meta<typeof EventDetailPanel> = {
  title: 'Events/EventDetailPanel',
  component: EventDetailPanel,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0f172a' },
        { name: 'light', value: '#ffffff' },
      ],
    },
  },
  decorators: [
    (Story) => (
      <div style={{ backgroundColor: '#0f172a', minHeight: '100vh', padding: '20px' }}>
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof EventDetailPanel>;

const baseEvent: FalcoEvent = {
  id: 'event-001',
  provider: 'aws',
  event_name: 'PutBucketPolicy',
  resource_type: 's3:bucket',
  resource_id: 'my-bucket-12345',
  user_identity: {
    Type: 'IAMUser',
    PrincipalID: 'AIDACKCEVSQ6C2EXAMPLE',
    UserName: 'john.doe',
    ARN: 'arn:aws:iam::123456789012:user/john.doe',
    AccountID: '123456789012',
  },
  changes: {
    BucketPolicy: {
      Statement: [
        {
          Effect: 'Allow',
          Principal: '*',
          Action: 's3:GetObject',
        },
      ],
    },
  },
  region: 'us-east-1',
  project_id: 'project-1',
  service_name: 's3',
  timestamp: new Date(Date.now() - 3600000).toISOString(),
  severity: 'medium',
  status: 'open',
  status_reason: '',
  related_drifts: [
    {
      severity: 'medium',
      attribute: 'BucketPolicy',
      old_value: { Statement: [] },
      new_value: { Statement: [{ Effect: 'Allow', Principal: '*', Action: 's3:GetObject' }] },
      matched_rules: ['S3_BUCKET_POLICY_CHANGED'],
      timestamp: new Date(Date.now() - 3600000).toISOString(),
      alert_type: 'CONFIGURATION_CHANGE',
    },
  ],
};

export const DefaultState: Story = {
  args: {
    event: baseEvent,
    onClose: () => console.log('Closed'),
    hasPrev: true,
    hasNext: true,
    currentIndex: 5,
    totalCount: 12,
  },
};

export const CriticalSeverity: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-critical',
      severity: 'critical',
      event_name: 'DeleteBucket',
      related_drifts: [
        {
          ...(baseEvent.related_drifts?.[0] ?? {
            severity: 'critical',
            attribute: 'BucketDeletion',
            old_value: null,
            new_value: null,
            matched_rules: [],
            timestamp: new Date().toISOString(),
            alert_type: 'CONFIGURATION_CHANGE',
          }),
          severity: 'critical',
          attribute: 'BucketDeletion',
        },
      ],
    },
    onClose: () => console.log('Closed'),
    hasPrev: false,
    hasNext: true,
    currentIndex: 0,
    totalCount: 3,
  },
};

export const HighSeverity: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-high',
      severity: 'high',
      event_name: 'ModifyDbInstance',
      resource_type: 'rds:instance',
      resource_id: 'prod-db-instance',
      status: 'acknowledged',
      status_reason: 'Monitoring for impact',
    },
    onClose: () => console.log('Closed'),
    hasPrev: true,
    hasNext: false,
    currentIndex: 10,
    totalCount: 12,
  },
};

export const LongAttributeData: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-long',
      changes: {
        SecurityGroupRules: Array.from({ length: 5 }, (_, i) => ({
          IpProtocol: 'tcp',
          FromPort: 8000 + i * 100,
          ToPort: 8100 + i * 100,
          CidrIp: '0.0.0.0/0',
        })),
      },
      related_drifts: Array.from({ length: 3 }, (_, i) => ({
        severity: 'high',
        attribute: `SecurityGroupRule${i + 1}`,
        old_value: null,
        new_value: {
          IpProtocol: 'tcp',
          FromPort: 8000 + i * 100,
          ToPort: 8100 + i * 100,
          CidrIp: '0.0.0.0/0',
          Description: `Ingress rule ${i + 1} - allows traffic on port ${8000 + i * 100} from anywhere`,
        },
        matched_rules: ['SECURITY_GROUP_OVERLY_PERMISSIVE'],
        timestamp: new Date().toISOString(),
        alert_type: 'SECURITY_CONFIGURATION_CHANGE',
      })),
    },
    onClose: () => console.log('Closed'),
    hasPrev: true,
    hasNext: true,
    currentIndex: 5,
    totalCount: 15,
  },
};

export const NoDataState: Story = {
  args: {
    event: null,
    onClose: () => console.log('Closed'),
  },
};

export const AcknowledgedStatus: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-ack',
      status: 'acknowledged',
      status_reason: 'Waiting for approval from security team',
      severity: 'high',
    },
    onClose: () => console.log('Closed'),
    hasPrev: true,
    hasNext: true,
    currentIndex: 3,
    totalCount: 8,
  },
};

export const IgnoredStatus: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-ignored',
      status: 'ignored',
      status_reason: 'Test environment - expected behavior',
      severity: 'low',
    },
    onClose: () => console.log('Closed'),
    hasPrev: false,
    hasNext: true,
    currentIndex: 0,
    totalCount: 5,
  },
};

export const MultipleRelatedDrifts: Story = {
  args: {
    event: {
      ...baseEvent,
      id: 'event-multi',
      related_drifts: Array.from({ length: 5 }, (_, i) => ({
        severity: ['critical', 'high', 'medium', 'low', 'medium'][i],
        attribute: `Config.${['EnableLogging', 'EnableEncryption', 'SetLifecycle', 'SetCors', 'ModifyVersioning'][i]}`,
        old_value: ['false', 'false', null, null, 'false'][i],
        new_value: ['true', 'true', { days: 90 }, { AllowedMethods: ['GET'] }, 'true'][i],
        matched_rules: [
          ['LOGGING_NOT_ENABLED'],
          ['ENCRYPTION_NOT_ENABLED'],
          ['LIFECYCLE_NOT_CONFIGURED'],
          ['CORS_ENABLED'],
          ['VERSIONING_ENABLED'],
        ][i],
        timestamp: new Date(Date.now() - (5 - i) * 600000).toISOString(),
        alert_type: 'CONFIGURATION_CHANGE',
      })),
    },
    onClose: () => console.log('Closed'),
    hasPrev: true,
    hasNext: true,
    currentIndex: 4,
    totalCount: 10,
  },
};
