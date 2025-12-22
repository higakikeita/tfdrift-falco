/**
 * Sample Drift Data Generator
 * デモ用のサンプルドリフトデータ生成
 */

import type { DriftEvent, DriftSeverity, ChangeType, Provider } from '../types/drift';

const changeTypes: ChangeType[] = ['created', 'modified', 'deleted'];
const providers: Provider[] = ['aws', 'gcp'];

const awsResourceTypes = [
  'aws_security_group',
  'aws_iam_policy',
  'aws_iam_role',
  'aws_instance',
  'aws_s3_bucket',
  'aws_rds_instance',
  'aws_lambda_function',
  'aws_dynamodb_table',
];

const gcpResourceTypes = [
  'google_compute_firewall',
  'google_iam_policy',
  'google_compute_instance',
  'google_storage_bucket',
  'google_sql_instance',
];

const awsUsers = [
  { userName: 'admin-user', arn: 'arn:aws:iam::123456789012:user/admin-user', accountId: '123456789012' },
  { userName: 'developer', arn: 'arn:aws:iam::123456789012:user/developer', accountId: '123456789012' },
  { userName: 'operator', arn: 'arn:aws:iam::123456789012:user/operator', accountId: '123456789012' },
];

const gcpUsers = [
  { userName: 'admin@example.com', principalId: 'user:admin@example.com' },
  { userName: 'developer@example.com', principalId: 'user:developer@example.com' },
];

const regions: Record<Provider, string[]> = {
  aws: ['us-east-1', 'us-west-2', 'ap-northeast-1', 'eu-west-1'],
  gcp: ['us-central1', 'asia-northeast1', 'europe-west1'],
  azure: ['eastus', 'westus2', 'japaneast', 'westeurope'],
};

const changeExamples = [
  {
    attribute: 'ingress_rules',
    oldValue: '{"cidr": "10.0.0.0/8", "port": 22}',
    newValue: '{"cidr": "0.0.0.0/0", "port": 22}',
  },
  {
    attribute: 'policy_document',
    oldValue: '{"Effect": "Allow", "Action": "s3:GetObject"}',
    newValue: '{"Effect": "Allow", "Action": "s3:*"}',
  },
  {
    attribute: 'instance_type',
    oldValue: 't3.micro',
    newValue: 'm5.xlarge',
  },
  {
    attribute: 'deletion_protection',
    oldValue: 'true',
    newValue: 'false',
  },
  {
    attribute: 'public_access',
    oldValue: 'false',
    newValue: 'true',
  },
];

export function generateSampleDrift(id: number): DriftEvent {
  const provider = providers[Math.floor(Math.random() * providers.length)];
  const resourceTypes = provider === 'aws' ? awsResourceTypes : gcpResourceTypes;
  const resourceType = resourceTypes[Math.floor(Math.random() * resourceTypes.length)];
  const users = provider === 'aws' ? awsUsers : gcpUsers;
  const user = users[Math.floor(Math.random() * users.length)];
  const regionList = regions[provider];
  const region = regionList[Math.floor(Math.random() * regionList.length)];
  const changeExample = changeExamples[Math.floor(Math.random() * changeExamples.length)];

  // Generate timestamp within last 7 days
  const now = Date.now();
  const daysAgo = Math.floor(Math.random() * 7);
  const hoursAgo = Math.floor(Math.random() * 24);
  const timestamp = new Date(now - (daysAgo * 24 * 60 * 60 * 1000) - (hoursAgo * 60 * 60 * 1000));

  // Severity based on change type
  let severity: DriftSeverity = 'medium';
  if (changeExample.attribute === 'ingress_rules' || changeExample.attribute === 'public_access') {
    severity = 'critical';
  } else if (changeExample.attribute === 'policy_document' || changeExample.attribute === 'deletion_protection') {
    severity = 'high';
  } else if (changeExample.attribute === 'instance_type') {
    severity = 'medium';
  }

  return {
    id: `drift-${id.toString().padStart(6, '0')}`,
    timestamp: timestamp.toISOString(),
    severity,
    provider,
    resourceType,
    resourceId: provider === 'aws'
      ? `${resourceType.split('_')[1]}-${Math.random().toString(36).substring(7)}`
      : `projects/my-project/zones/${region}/instances/${Math.random().toString(36).substring(7)}`,
    resourceName: `${resourceType.split('_').pop()}-prod-${Math.floor(Math.random() * 100)}`,
    changeType: changeTypes[Math.floor(Math.random() * changeTypes.length)],
    attribute: changeExample.attribute,
    oldValue: changeExample.oldValue,
    newValue: changeExample.newValue,
    userIdentity: {
      type: provider === 'aws' ? 'IAMUser' : 'ServiceAccount',
      userName: user.userName,
      arn: 'arn' in user ? user.arn : undefined,
      accountId: 'accountId' in user ? user.accountId : undefined,
      principalId: 'principalId' in user ? user.principalId : undefined,
    },
    region,
    cloudtrailEventId: `${Math.random().toString(36).substring(2, 15)}-${Math.random().toString(36).substring(2, 15)}`,
    cloudtrailEventName: provider === 'aws' ? 'ModifySecurityGroupRules' : 'compute.firewalls.patch',
    sourceIP: `203.0.113.${Math.floor(Math.random() * 255)}`,
    userAgent: 'aws-cli/2.13.0 Python/3.11.4',
    tags: {
      Environment: ['production', 'staging', 'development'][Math.floor(Math.random() * 3)],
      Team: ['platform', 'backend', 'frontend'][Math.floor(Math.random() * 3)],
    },
  };
}

export function generateSampleDrifts(count: number = 50): DriftEvent[] {
  return Array.from({ length: count }, (_, i) => generateSampleDrift(i + 1));
}

export function generateRecentDrifts(count: number = 10): DriftEvent[] {
  const drifts = generateSampleDrifts(count);
  // Sort by timestamp descending
  return drifts.sort((a, b) =>
    new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );
}
