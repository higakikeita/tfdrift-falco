import type { Meta, StoryObj } from '@storybook/react';
import DriftDetailPanel from './DriftDetailPanel';
import type { DriftEvent } from '../types/drift';

const meta: Meta<typeof DriftDetailPanel> = {
  title: 'Dashboard/DriftDetailPanel',
  component: DriftDetailPanel,
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
      <div style={{ backgroundColor: '#111827', minHeight: '100vh', display: 'flex' }}>
        <div style={{ flex: 1, backgroundColor: '#1f2937', padding: '20px' }}>
          {/* Placeholder for left panel */}
        </div>
        <div style={{ width: '50%', backgroundColor: '#ffffff' }}>
          <Story />
        </div>
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof DriftDetailPanel>;

const baseDrift: DriftEvent = {
  id: 'drift-001',
  timestamp: new Date(Date.now() - 3600000).toISOString(),
  severity: 'medium',
  provider: 'aws',
  resourceType: 's3:bucket',
  resourceId: 'my-production-bucket',
  resourceName: 'my-production-bucket',
  changeType: 'modified',
  attribute: 'BucketPolicy',
  oldValue: '{"Version":"2012-10-17","Statement":[]}',
  newValue: '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"s3:GetObject"}]}',
  userIdentity: {
    userName: 'john.doe',
    type: 'IAMUser',
    arn: 'arn:aws:iam::123456789012:user/john.doe',
    accountId: '123456789012',
  },
  region: 'us-east-1',
  cloudtrailEventId: 'evt-123456789',
  cloudtrailEventName: 'PutBucketPolicy',
  sourceIP: '203.0.113.42',
  userAgent: 'aws-cli/2.13.0 Python/3.11.0',
  tags: {
    Environment: 'production',
    Application: 'data-lake',
  },
};

export const DriftSelected: Story = {
  args: {
    drift: baseDrift,
    onClose: () => console.log('Panel closed'),
  },
};

export const CriticalSeverity: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-critical-001',
      severity: 'critical',
      changeType: 'deleted',
      resourceName: 'critical-backup-bucket',
      attribute: 'BucketDeletion',
      oldValue: '{"bucket": "critical-backup-bucket", "versioning": "enabled"}',
      newValue: null,
      cloudtrailEventName: 'DeleteBucket',
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const HighSeverity: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-high-001',
      severity: 'high',
      resourceType: 'rds:instance',
      resourceName: 'production-postgres-db',
      changeType: 'modified',
      attribute: 'EngineVersion',
      oldValue: '13.7',
      newValue: '13.9',
      cloudtrailEventName: 'ModifyDBInstance',
      userIdentity: {
        userName: 'database-admin',
        type: 'IAMRole',
        arn: 'arn:aws:iam::123456789012:role/database-admin-role',
        accountId: '123456789012',
      },
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const LowSeverity: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-low-001',
      severity: 'low',
      resourceType: 'ec2:instance',
      resourceName: 'dev-test-instance',
      changeType: 'modified',
      attribute: 'Tags',
      oldValue: '{"Environment":"staging","Owner":"dev-team"}',
      newValue: '{"Environment":"staging","Owner":"dev-team","Project":"experimental"}',
      cloudtrailEventName: 'CreateTags',
      region: 'ap-southeast-1',
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const NoSelection: Story = {
  args: {
    drift: null,
    onClose: () => console.log('Panel closed'),
  },
};

export const CreatedResource: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-created-001',
      severity: 'high',
      changeType: 'created',
      resourceType: 'security_group',
      resourceName: 'new-security-group',
      attribute: 'SecurityGroupRules',
      oldValue: null,
      newValue: '{"GroupId":"sg-123456","Rules":[{"IpProtocol":"tcp","FromPort":443,"ToPort":443,"CidrIp":"0.0.0.0/0"}]}',
      cloudtrailEventName: 'CreateSecurityGroup',
      userIdentity: {
        userName: 'admin',
        type: 'IAMUser',
        arn: 'arn:aws:iam::123456789012:user/admin',
        accountId: '123456789012',
      },
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const ComplexConfiguration: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-complex-001',
      severity: 'high',
      resourceType: 'lambda:function',
      resourceName: 'data-processing-lambda',
      changeType: 'modified',
      attribute: 'FunctionConfiguration',
      oldValue: '{"Runtime":"python3.8","Memory":128,"Timeout":30,"EphemeralStorage":512,"Layers":["layer-1"]}',
      newValue: '{"Runtime":"python3.11","Memory":512,"Timeout":120,"EphemeralStorage":1024,"Layers":["layer-1","layer-2","layer-3"]}',
      cloudtrailEventName: 'UpdateFunctionConfiguration',
      tags: {
        Environment: 'production',
        Application: 'data-pipeline',
        CostCenter: 'engineering',
        Compliance: 'sox',
      },
      metadata: {
        changes_count: 5,
        impact_score: 8.5,
        affected_services: ['DataLake', 'Analytics', 'Reporting'],
      },
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const GcpResource: Story = {
  args: {
    drift: {
      ...baseDrift,
      id: 'drift-gcp-001',
      provider: 'gcp',
      severity: 'medium',
      resourceType: 'compute:instance',
      resourceName: 'gke-node-pool-instance',
      changeType: 'modified',
      attribute: 'MachineType',
      oldValue: 'n1-standard-1',
      newValue: 'n1-standard-2',
      region: 'us-central1',
      cloudtrailEventName: 'compute.instances.setMachineType',
      userIdentity: {
        userName: 'infrastructure-sa@project.iam.gserviceaccount.com',
        type: 'ServiceAccount',
        arn: 'arn:gcp:iam::project-id:serviceaccount/infrastructure-sa',
        accountId: 'project-id',
      },
      sourceIP: '10.0.0.1',
    },
    onClose: () => console.log('Panel closed'),
  },
};

export const WithoutCloudTrailData: Story = {
  args: {
    drift: {
      id: 'drift-no-audit-001',
      timestamp: new Date(Date.now() - 7200000).toISOString(),
      severity: 'medium',
      provider: 'aws',
      resourceType: 's3:bucket',
      resourceId: 'legacy-bucket',
      resourceName: 'legacy-bucket',
      changeType: 'modified',
      attribute: 'VersioningConfiguration',
      oldValue: 'false',
      newValue: 'true',
      userIdentity: {
        userName: 'unknown',
        type: 'Unknown',
      },
      region: 'us-west-2',
    },
    onClose: () => console.log('Panel closed'),
  },
};
