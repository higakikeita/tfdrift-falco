import type { Meta, StoryObj } from '@storybook/react';
import DriftHistoryTable from './DriftHistoryTable';
import type { DriftEvent } from '../types/drift';

const meta: Meta<typeof DriftHistoryTable> = {
  title: 'Dashboard/DriftHistoryTable',
  component: DriftHistoryTable,
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
type Story = StoryObj<typeof DriftHistoryTable>;

const generateDriftEvents = (count: number): DriftEvent[] => {
  const providers: Array<'aws' | 'gcp' | 'azure'> = ['aws', 'gcp', 'azure'];
  const resourceTypes = ['s3:bucket', 'rds:instance', 'ec2:instance', 'security_group', 'iam:role', 'lambda:function', 'vpc', 'eks:cluster'];
  const changeTypes: Array<'created' | 'modified' | 'deleted'> = ['created', 'modified', 'deleted'];
  const severities: Array<'critical' | 'high' | 'medium' | 'low'> = ['critical', 'high', 'medium', 'low'];

  return Array.from({ length: count }, (_, i) => ({
    id: `drift-${i + 1}`,
    timestamp: new Date(Date.now() - i * 300000).toISOString(),
    severity: severities[Math.floor(Math.random() * severities.length)],
    provider: providers[Math.floor(Math.random() * providers.length)],
    resourceType: resourceTypes[Math.floor(Math.random() * resourceTypes.length)],
    resourceId: `resource-${Math.random().toString(36).substring(7)}`,
    resourceName: `resource-name-${i + 1}`,
    changeType: changeTypes[Math.floor(Math.random() * changeTypes.length)],
    attribute: ['Configuration', 'Policy', 'Tags', 'Settings', 'Permissions'][Math.floor(Math.random() * 5)],
    oldValue: i % 3 === 0 ? null : `{"old": "value${i}"}`,
    newValue: `{"new": "value${i + 1}"}`,
    userIdentity: {
      userName: ['john.doe', 'jane.smith', 'admin', 'ci-pipeline', 'terraform-bot'][Math.floor(Math.random() * 5)],
      type: ['IAMUser', 'IAMRole', 'ServiceAccount'][Math.floor(Math.random() * 3)],
      arn: `arn:aws:iam::123456789012:user/user${i}`,
      accountId: '123456789012',
    },
    region: ['us-east-1', 'us-west-2', 'eu-west-1', 'ap-southeast-1'][Math.floor(Math.random() * 4)],
  }));
};

export const PopulatedData: Story = {
  args: {
    drifts: generateDriftEvents(25),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const SmallDataset: Story = {
  args: {
    drifts: generateDriftEvents(5),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const LargeDataset: Story = {
  args: {
    drifts: generateDriftEvents(100),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const EmptyState: Story = {
  args: {
    drifts: [],
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const OnlyHighSeverity: Story = {
  args: {
    drifts: generateDriftEvents(20).filter(d => d.severity === 'high' || d.severity === 'critical'),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const OnlyAwsResources: Story = {
  args: {
    drifts: generateDriftEvents(30).filter(d => d.provider === 'aws'),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const OnlyModifiedResources: Story = {
  args: {
    drifts: generateDriftEvents(20).filter(d => d.changeType === 'modified'),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const RecentDrifts: Story = {
  args: {
    drifts: Array.from({ length: 10 }, (_, i) => ({
      id: `drift-recent-${i + 1}`,
      timestamp: new Date(Date.now() - i * 60000).toISOString(),
      severity: (['critical', 'high', 'medium', 'low'] as const)[i % 4],
      provider: (['aws', 'gcp'] as const)[i % 2],
      resourceType: ['s3:bucket', 'rds:instance', 'ec2:instance'][i % 3],
      resourceId: `recent-resource-${i + 1}`,
      resourceName: `Recent Resource ${i + 1}`,
      changeType: (['created', 'modified', 'deleted'] as const)[i % 3],
      attribute: 'Configuration',
      oldValue: null,
      newValue: '{}',
      userIdentity: {
        userName: 'automation-user',
        type: 'ServiceAccount',
        arn: 'arn:aws:iam::123456789012:role/automation',
        accountId: '123456789012',
      },
      region: 'us-east-1',
    })),
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const MixedSeverities: Story = {
  args: {
    drifts: [
      ...generateDriftEvents(5).map((d) => ({ ...d, severity: 'critical' as const })),
      ...generateDriftEvents(8).map((d) => ({ ...d, severity: 'high' as const })),
      ...generateDriftEvents(10).map((d) => ({ ...d, severity: 'medium' as const })),
      ...generateDriftEvents(7).map((d) => ({ ...d, severity: 'low' as const })),
    ],
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const WithoutCallback: Story = {
  args: {
    drifts: generateDriftEvents(15),
  },
};

export const SingleDrift: Story = {
  args: {
    drifts: [
      {
        id: 'drift-single',
        timestamp: new Date().toISOString(),
        severity: 'high',
        provider: 'aws',
        resourceType: 's3:bucket',
        resourceId: 'single-bucket',
        resourceName: 'Single Drift Event',
        changeType: 'modified',
        attribute: 'BucketPolicy',
        oldValue: '{"Version":"2012-10-17"}',
        newValue: '{"Version":"2012-10-17","Statement":[]}',
        userIdentity: {
          userName: 'admin',
          type: 'IAMUser',
          arn: 'arn:aws:iam::123456789012:user/admin',
          accountId: '123456789012',
        },
        region: 'us-east-1',
      },
    ],
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};

export const ComplexResourceTypes: Story = {
  args: {
    drifts: [
      ...generateDriftEvents(3).map((d, i) => ({ ...d, resourceType: 'lambda:function', resourceName: `Lambda Function ${i + 1}` })),
      ...generateDriftEvents(3).map((d, i) => ({ ...d, resourceType: 'rds:cluster', resourceName: `RDS Cluster ${i + 1}` })),
      ...generateDriftEvents(3).map((d, i) => ({ ...d, resourceType: 'eks:cluster', resourceName: `EKS Cluster ${i + 1}` })),
      ...generateDriftEvents(3).map((d, i) => ({ ...d, resourceType: 'api_gateway', resourceName: `API Gateway ${i + 1}` })),
      ...generateDriftEvents(3).map((d, i) => ({ ...d, resourceType: 'dynamodb:table', resourceName: `DynamoDB Table ${i + 1}` })),
    ],
    onSelectDrift: (drift) => console.log('Selected drift:', drift),
  },
};
