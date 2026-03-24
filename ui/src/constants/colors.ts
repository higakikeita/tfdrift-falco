/**
 * Shared Color Constants
 *
 * AWSサービスカラー、ドリフトステータスカラー、その他のビジュアルカラーの定義。
 * CytoscapeGraph、cytoscapeStyles、DriftDashboard などで共通利用。
 *
 * AWS公式カラーパレットに準拠:
 * https://aws.amazon.com/architecture/icons/
 */

// ==================== AWS Service Colors ====================

/** AWS Compute サービスカラー */
export const AWS_COMPUTE_COLORS = {
  lambda: '#FF9900',
  eks: '#ED7100',
  ecsFargate: '#FF9900',
  ec2: '#FF9900',
} as const;

/** AWS Database サービスカラー */
export const AWS_DATABASE_COLORS = {
  rds: '#3B48CC',
  aurora: '#3B48CC',
  dynamodb: '#3B48CC',
  elasticache: '#5A6EDB',
  neptune: '#3B48CC',
} as const;

/** AWS Storage サービスカラー */
export const AWS_STORAGE_COLORS = {
  s3: '#569A31',
} as const;

/** AWS Network サービスカラー */
export const AWS_NETWORK_COLORS = {
  vpc: '#2E73B8',
  subnet: '#5294CF',
  albNlb: '#8C4FFF',
  cloudfront: '#8C4FFF',
  igwNat: '#4A90E2',
} as const;

/** AWS Security サービスカラー */
export const AWS_SECURITY_COLORS = {
  securityGroup: '#DD344C',
  iam: '#DD344C',
  kms: '#759C3E',
} as const;

/** AWS Integration サービスカラー */
export const AWS_INTEGRATION_COLORS = {
  apiGateway: '#945DF2',
  snsSqs: '#D13212',
  stepFunctions: '#D13212',
  eventBridge: '#D13212',
} as const;

/** 全AWSサービスカラーの統合オブジェクト */
export const AWS_SERVICE_COLORS = {
  ...AWS_COMPUTE_COLORS,
  ...AWS_DATABASE_COLORS,
  ...AWS_STORAGE_COLORS,
  ...AWS_NETWORK_COLORS,
  ...AWS_SECURITY_COLORS,
  ...AWS_INTEGRATION_COLORS,
} as const;

// ==================== Legend Data ====================

/** レジェンド用のAWSサービスエントリ */
export interface LegendEntry {
  label: string;
  color: string;
}

/** レジェンドカテゴリ定義 */
export interface LegendCategory {
  name: string;
  items: LegendEntry[];
}

/** AWSサービスレジェンドのカテゴリ別定義 */
export const AWS_SERVICE_LEGEND: LegendCategory[] = [
  {
    name: 'Compute',
    items: [
      { label: 'Lambda', color: AWS_COMPUTE_COLORS.lambda },
      { label: 'EKS', color: AWS_COMPUTE_COLORS.eks },
      { label: 'ECS/Fargate', color: AWS_COMPUTE_COLORS.ecsFargate },
    ],
  },
  {
    name: 'Database',
    items: [
      { label: 'RDS', color: AWS_DATABASE_COLORS.rds },
      { label: 'Aurora', color: AWS_DATABASE_COLORS.aurora },
      { label: 'DynamoDB', color: AWS_DATABASE_COLORS.dynamodb },
      { label: 'ElastiCache', color: AWS_DATABASE_COLORS.elasticache },
      { label: 'Neptune', color: AWS_DATABASE_COLORS.neptune },
    ],
  },
  {
    name: 'Storage',
    items: [
      { label: 'S3', color: AWS_STORAGE_COLORS.s3 },
    ],
  },
  {
    name: 'Network',
    items: [
      { label: 'VPC', color: AWS_NETWORK_COLORS.vpc },
      { label: 'Subnet', color: AWS_NETWORK_COLORS.subnet },
      { label: 'ALB/NLB', color: AWS_NETWORK_COLORS.albNlb },
      { label: 'CloudFront', color: AWS_NETWORK_COLORS.cloudfront },
      { label: 'IGW/NAT', color: AWS_NETWORK_COLORS.igwNat },
    ],
  },
  {
    name: 'Security',
    items: [
      { label: 'SG', color: AWS_SECURITY_COLORS.securityGroup },
      { label: 'IAM', color: AWS_SECURITY_COLORS.iam },
      { label: 'KMS', color: AWS_SECURITY_COLORS.kms },
    ],
  },
  {
    name: 'Integration',
    items: [
      { label: 'API Gateway', color: AWS_INTEGRATION_COLORS.apiGateway },
      { label: 'SNS/SQS', color: AWS_INTEGRATION_COLORS.snsSqs },
      { label: 'Step Functions', color: AWS_INTEGRATION_COLORS.stepFunctions },
      { label: 'EventBridge', color: AWS_INTEGRATION_COLORS.eventBridge },
    ],
  },
];

// ==================== Drift Status Colors ====================

/** ドリフトステータスのカラー定義（Hex値） */
export const DRIFT_STATUS_COLORS = {
  critical: '#dc2626', // red-600
  high: '#f97316',     // orange-500
  medium: '#eab308',   // yellow-500
  low: '#adb5bd',      // gray
} as const;

/** ドリフトステータスのTailwindボーダークラス */
export const DRIFT_STATUS_BORDER_CLASSES = {
  critical: 'border-red-600',
  high: 'border-orange-500',
  medium: 'border-yellow-500',
  low: 'border-gray-400',
} as const;

/** ドリフトステータスレジェンド定義 */
export interface DriftLegendEntry {
  label: string;
  description: string;
  borderClass: string;
}

export const DRIFT_STATUS_LEGEND: DriftLegendEntry[] = [
  { label: 'Critical', description: 'Missing', borderClass: DRIFT_STATUS_BORDER_CLASSES.critical },
  { label: 'High', description: 'Modified', borderClass: DRIFT_STATUS_BORDER_CLASSES.high },
  { label: 'Medium', description: 'Unmanaged', borderClass: DRIFT_STATUS_BORDER_CLASSES.medium },
];

// ==================== Cytoscape Style Colors ====================

/** Cytoscapeスタイルで使用するドリフト重大度の色（hex） */
export const CYTOSCAPE_DRIFT_COLORS = {
  critical: '#c92a2a',
  high: '#e8590c',
  medium: '#fab005',
  low: '#adb5bd',
} as const;

// ==================== Node Type Colors ====================

/** ノードタイプ別のカラー定義（Cytoscapeスタイル用） */
export const NODE_TYPE_COLORS = {
  terraformChange: '#623CE4',
  falcoEvent: '#EC4C39',
  iamRole: '#1c7ed6',
  serviceAccount: '#51cf66',
  pod: '#ffd43b',
  container: '#0DB7ED',
  securityGroup: '#9775fa',
  network: '#22b8cf',
  kubernetes: '#326CE5',
} as const;

// ==================== Provider Colors ====================

/** クラウドプロバイダーカラー */
export const PROVIDER_COLORS = {
  aws: '#f59e0b',     // amber-500
  gcp: '#3b82f6',     // blue-500
  azure: '#0078d4',
} as const;
