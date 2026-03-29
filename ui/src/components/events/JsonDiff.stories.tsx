import type { Meta, StoryObj } from '@storybook/react';
import { JsonDiff } from './JsonDiff';

const meta: Meta<typeof JsonDiff> = {
  title: 'Events/JsonDiff',
  component: JsonDiff,
  tags: ['autodocs'],
  parameters: {
    layout: 'padded',
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
        <div style={{ backgroundColor: '#1e293b', padding: '20px', borderRadius: '8px', maxWidth: '800px' }}>
          <Story />
        </div>
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof JsonDiff>;

export const SimpleDiff: Story = {
  args: {
    oldValue: { name: 'John', age: 30 },
    newValue: { name: 'John', age: 31 },
    attribute: 'user.age',
  },
};

export const StringDiff: Story = {
  args: {
    oldValue: 'Hello World',
    newValue: 'Hello Universe',
    attribute: 'description',
  },
};

export const ComplexNestedDiff: Story = {
  args: {
    oldValue: {
      database: {
        host: 'db-old.example.com',
        port: 5432,
        credentials: {
          username: 'admin',
          password: 'secret123',
        },
        settings: {
          max_connections: 100,
          backup_enabled: false,
          backup_schedule: null,
        },
      },
    },
    newValue: {
      database: {
        host: 'db-new.example.com',
        port: 5432,
        credentials: {
          username: 'admin_prod',
          password: 'newsecret456',
        },
        settings: {
          max_connections: 200,
          backup_enabled: true,
          backup_schedule: 'daily',
        },
      },
    },
    attribute: 'database.configuration',
  },
};

export const NoChanges: Story = {
  args: {
    oldValue: { status: 'active', region: 'us-east-1', tags: ['prod', 'monitoring'] },
    newValue: { status: 'active', region: 'us-east-1', tags: ['prod', 'monitoring'] },
    attribute: 'resource.state',
  },
};

export const LargeDiff: Story = {
  args: {
    oldValue: {
      name: 'Lambda Function v1',
      runtime: 'python3.8',
      memory: 128,
      timeout: 30,
      handler: 'index.handler',
      environment: {
        LOG_LEVEL: 'INFO',
        ENVIRONMENT: 'staging',
      },
      layers: ['layer-1', 'layer-2'],
      vpc_config: {
        subnet_ids: ['subnet-123', 'subnet-456'],
        security_group_ids: ['sg-123', 'sg-456'],
      },
      dead_letter_queue: null,
      ephemeral_storage: 512,
      code_sha: 'abc123def456',
      version: '1',
      last_modified: '2023-01-15T10:30:00Z',
    },
    newValue: {
      name: 'Lambda Function v2',
      runtime: 'python3.11',
      memory: 256,
      timeout: 60,
      handler: 'index.handler',
      environment: {
        LOG_LEVEL: 'DEBUG',
        ENVIRONMENT: 'production',
        API_KEY: 'new-api-key',
      },
      layers: ['layer-1', 'layer-2', 'layer-3', 'layer-4'],
      vpc_config: {
        subnet_ids: ['subnet-789', 'subnet-012'],
        security_group_ids: ['sg-789', 'sg-012', 'sg-345'],
      },
      dead_letter_queue: 'arn:aws:sqs:us-east-1:123456789012:dlq',
      ephemeral_storage: 1024,
      code_sha: 'xyz789uvw012',
      version: '5',
      last_modified: '2024-02-20T14:45:00Z',
      enable_cors: true,
      authentication: {
        type: 'JWT',
        issuer: 'https://auth.example.com',
      },
    },
    attribute: 'lambda.function.config',
  },
};

export const JsonStringDiff: Story = {
  args: {
    oldValue: '{"rules":[{"id":1,"action":"ALLOW"}]}',
    newValue: '{"rules":[{"id":1,"action":"DENY"},{"id":2,"action":"ALLOW"}]}',
    attribute: 'policy.rules',
  },
};

export const NullToValueDiff: Story = {
  args: {
    oldValue: null,
    newValue: {
      enabled: true,
      options: {
        retry_count: 3,
        timeout_ms: 5000,
      },
    },
    attribute: 'feature.config',
  },
};

export const ValueToNullDiff: Story = {
  args: {
    oldValue: {
      backup_schedule: 'daily',
      backup_retention_days: 30,
      backup_location: 's3://backups',
    },
    newValue: null,
    attribute: 'backup.settings',
  },
};

export const BooleanDiff: Story = {
  args: {
    oldValue: false,
    newValue: true,
    attribute: 'enable_encryption',
  },
};

export const NumericDiff: Story = {
  args: {
    oldValue: 100,
    newValue: 500,
    attribute: 'max_retries',
  },
};

export const ArrayDiff: Story = {
  args: {
    oldValue: ['rule-1', 'rule-2', 'rule-3'],
    newValue: ['rule-1', 'rule-3', 'rule-4', 'rule-5'],
    attribute: 'security_rules',
  },
};
