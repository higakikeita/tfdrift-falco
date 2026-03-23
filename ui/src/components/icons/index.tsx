/**
 * TFDrift-Falco Icon System
 *
 * Unified icon system - ALL icons go through OfficialCloudIcons
 * Legacy icon files (AWSServiceIcons, GCPServiceIcons, K8sAndSpecialIcons, CloudIcons)
 * are kept for backward compatibility but OfficialCloudIcon is the single source of truth.
 */

/* eslint-disable react-refresh/only-export-components */

// Primary icon system - USE THIS
export {
  OfficialCloudIcon,
  getOfficialIconPath,
  getProviderFromType,
  getProviderColor,
  ALL_ICON_MAPPINGS,
} from './OfficialCloudIcons';

// Legacy exports (kept for backward compatibility)
export * from './CloudIcons';
export * from './AWSServiceIcons';
export * from './GCPServiceIcons';
export * from './K8sAndSpecialIcons';

import { AWSServiceIcons } from './AWSServiceIcons';
import { GCPServiceIcons } from './GCPServiceIcons';
import { K8sAndSpecialIcons } from './K8sAndSpecialIcons';

/**
 * Combined icon registry for all resource types (legacy)
 * @deprecated Use OfficialCloudIcon component directly instead
 */
export const ResourceIcons = {
  ...AWSServiceIcons,
  ...GCPServiceIcons,
  ...K8sAndSpecialIcons,
};

/**
 * Get icon component for a resource type (legacy)
 * @deprecated Use OfficialCloudIcon component directly instead
 */
export function getResourceIcon(resourceType: string): React.ComponentType<{ size?: number; className?: string }> | null {
  const normalizedType = resourceType
    .toLowerCase()
    .replace(/^(aws|gcp|kubernetes|k8s)_/, (_match, prefix) => {
      if (prefix === 'k8s') return 'kubernetes_';
      return `${prefix}_`;
    });

  return ResourceIcons[normalizedType as keyof typeof ResourceIcons] || null;
}

/**
 * Resource type categories
 */
export const ResourceCategories = {
  AWS: [
    'aws_iam_policy', 'aws_iam_role', 'aws_iam_user',
    'aws_lambda_function', 'aws_lambda',
    'aws_s3_bucket', 'aws_s3',
    'aws_instance', 'aws_ec2_instance',
    'aws_security_group', 'aws_vpc', 'aws_subnet',
    'aws_eks_cluster', 'aws_ecs_service',
    'aws_dynamodb_table', 'aws_db_instance',
    'aws_lb', 'aws_elb', 'aws_nat_gateway',
    'aws_ebs_volume', 'aws_route_table',
  ],
  GCP: [
    'gcp_iam_policy', 'gcp_iam_role', 'gcp_service_account',
    'gcp_compute_instance', 'gcp_compute',
    'gcp_storage_bucket', 'gcp_storage',
    'gcp_gke_cluster', 'gcp_gke',
    'gcp_cloud_function', 'gcp_cloud_functions',
    'gcp_compute_network', 'gcp_vpc',
    'gcp_compute_firewall',
  ],
  Kubernetes: [
    'kubernetes_pod', 'kubernetes_service', 'kubernetes_service_account',
    'kubernetes_deployment', 'kubernetes_namespace',
    'kubernetes_role', 'kubernetes_cluster_role',
    'kubernetes_container', 'container',
    'kubernetes_config_map', 'kubernetes_secret',
    'kubernetes_ingress', 'kubernetes_stateful_set',
  ],
  Special: [
    'terraform_change', 'terraform', 'terraform_resource',
    'falco_event', 'falco',
    'drift', 'drift_detected',
    'network', 'security_group',
  ],
};

/**
 * Get color for resource category
 */
export function getCategoryColor(resourceType: string): string {
  if (resourceType.startsWith('aws_')) return '#FF9900';
  if (resourceType.startsWith('gcp_')) return '#4285F4';
  if (resourceType.startsWith('kubernetes_') || resourceType.startsWith('k8s_')) return '#326CE5';
  if (resourceType.startsWith('terraform_')) return '#7B42BC';
  if (resourceType.startsWith('falco')) return '#00AEC7';
  if (resourceType === 'drift' || resourceType === 'drift_detected') return '#FF5722';
  return '#607D8B';
}
