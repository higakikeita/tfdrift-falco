/**
 * Utility functions and data for icon handling
 * Separated from component file for react-refresh compliance
 */

// Need to import icon mappings from the component file to generate ALL_ICON_MAPPINGS
// Defined inline here to avoid circular dependency
const AWS_ICON_MAPPING_KEYS = [
  'aws_iam_policy', 'aws_iam_role', 'aws_iam_user', 'aws_iam_group', 'aws_iam_access_analyzer',
  'aws_iam_instance_profile', 'aws_iam_policy_attachment', 'aws_iam_role_policy',
  'aws_iam_role_policy_attachment', 'aws_lambda_function', 'aws_lambda', 'aws_lambda_permission',
  'aws_lambda_layer_version', 'aws_instance', 'aws_ec2_instance', 'ec2_instance',
  'aws_launch_template', 'aws_autoscaling_group', 'aws_ecs_service', 'aws_ecs_cluster',
  'aws_ecs_task_definition', 'aws_eks_cluster', 'aws_eks_node_group', 'aws_s3_bucket',
  'aws_s3', 'aws_s3_bucket_policy', 'aws_s3_object', 'aws_ebs_volume', 'aws_vpc',
  'aws_subnet', 'aws_security_group', 'aws_security_group_rule', 'aws_network_acl',
  'aws_internet_gateway', 'aws_route_table', 'aws_route', 'aws_elb', 'aws_lb',
  'aws_alb', 'aws_lb_target_group', 'aws_lb_listener', 'load_balancer', 'aws_nat_gateway',
  'nat_gateway', 'aws_db_instance', 'aws_rds_instance', 'rds_instance', 'aws_rds_cluster',
  'aws_db_subnet_group', 'aws_dynamodb_table', 'aws_dynamodb',
];

const GCP_ICON_MAPPING_KEYS = [
  'gcp_iam_policy', 'gcp_iam_role', 'gcp_iam_member', 'gcp_iam_binding',
  'gcp_service_account', 'gcp_compute_instance', 'gcp_compute', 'gcp_compute_instance_template',
  'gcp_storage_bucket', 'gcp_storage', 'gcp_gke_cluster', 'gcp_gke', 'gcp_container_cluster',
  'gcp_cloud_function', 'gcp_cloud_functions', 'gcp_compute_network', 'gcp_vpc',
  'gcp_compute_subnetwork', 'gcp_compute_firewall', 'gcp_firewall',
];

const K8S_ICON_MAPPING_KEYS = [
  'kubernetes_pod', 'kubernetes_service', 'kubernetes_service_account', 'kubernetes_deployment',
  'kubernetes_namespace', 'kubernetes_role', 'kubernetes_cluster_role', 'kubernetes_cluster_role_binding',
  'kubernetes_role_binding', 'kubernetes_config_map', 'kubernetes_secret', 'kubernetes_ingress',
  'kubernetes_stateful_set', 'kubernetes_daemon_set', 'kubernetes_replica_set', 'kubernetes_job',
  'kubernetes_cron_job', 'kubernetes_container', 'container',
];

const SPECIAL_ICON_MAPPING_KEYS = [
  'terraform_change', 'terraform', 'terraform_resource', 'falco_event', 'falco', 'falco_alert',
  'drift', 'drift_detected', 'network', 'security_group',
];

export const ALL_ICON_MAPPINGS: Record<string, string> = {
  ...Object.fromEntries(AWS_ICON_MAPPING_KEYS.map(k => [k, 'aws'])),
  ...Object.fromEntries(GCP_ICON_MAPPING_KEYS.map(k => [k, 'gcp'])),
  ...Object.fromEntries(K8S_ICON_MAPPING_KEYS.map(k => [k, 'kubernetes'])),
  ...Object.fromEntries(SPECIAL_ICON_MAPPING_KEYS.map(k => [k, 'special'])),
};

const ICON_DATA_URI = {
  aws: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23232F3E"/><text x="12" y="14" text-anchor="middle" fill="%23FF9900" font-size="6" font-weight="bold">AWS</text></svg>'),
  gcp: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%234285F4"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="5" font-weight="bold">GCP</text></svg>'),
  kubernetes: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23326CE5"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="5" font-weight="bold">K8s</text></svg>'),
  terraform: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%237B42BC"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="6" font-weight="bold">TF</text></svg>'),
  default: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23607D8B"/><circle cx="12" cy="12" r="4" stroke="white" stroke-width="1.5" fill="none"/></svg>'),
} as const;

const PROVIDER_COLORS = {
  aws: '#FF9900',
  gcp: '#4285F4',
  kubernetes: '#326CE5',
  terraform: '#7B42BC',
  falco: '#00AEC7',
  unknown: '#607D8B',
} as const;

const PROVIDER_PREFIXES = [
  { prefix: 'aws_', provider: 'aws' },
  { prefix: 'gcp_', provider: 'gcp' },
  { prefix: 'google_', provider: 'gcp' },
  { prefix: 'kubernetes_', provider: 'kubernetes' },
  { prefix: 'k8s_', provider: 'kubernetes' },
  { prefix: 'terraform_', provider: 'terraform' },
  { prefix: 'tf_', provider: 'terraform' },
  { prefix: 'falco', provider: 'falco' },
] as const;

export function getProviderFromType(type: string): string {
  const t = type.toLowerCase();

  for (const { prefix, provider } of PROVIDER_PREFIXES) {
    if (t.startsWith(prefix)) {
      return provider;
    }
  }

  return 'unknown';
}

export function getProviderColor(provider: string): string {
  return PROVIDER_COLORS[provider as keyof typeof PROVIDER_COLORS] || PROVIDER_COLORS.unknown;
}

export function getOfficialIconPath(type: string): string {
  const normalizedType = type.toLowerCase().replace(/^k8s_/, 'kubernetes_');
  const provider = getProviderFromType(normalizedType);

  return ICON_DATA_URI[provider as keyof typeof ICON_DATA_URI] || ICON_DATA_URI.default;
}
