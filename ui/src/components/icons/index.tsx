/**
 * TFDrift-Falco Icon System
 *
 * Unified export for all cloud provider and service icons
 */

/* eslint-disable react-refresh/only-export-components */
export * from './CloudIcons';
export * from './AWSServiceIcons';
export * from './GCPServiceIcons';
export * from './K8sAndSpecialIcons';
/* eslint-enable react-refresh/only-export-components */

import { AWSServiceIcons } from './AWSServiceIcons';
import { GCPServiceIcons } from './GCPServiceIcons';
import { K8sAndSpecialIcons } from './K8sAndSpecialIcons';

/**
 * Combined icon registry for all resource types
 */
export const ResourceIcons = {
  ...AWSServiceIcons,
  ...GCPServiceIcons,
  ...K8sAndSpecialIcons,
};

/**
 * Get icon component for a resource type
 */
export function getResourceIcon(resourceType: string): React.ComponentType<{ size?: number; className?: string }> | null {
  // Normalize resource type (handle prefixes and variations)
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
    'aws_iam_policy',
    'aws_iam_role',
    'aws_lambda',
    'aws_s3',
    'aws_ec2',
    'aws_security_group',
    'aws_eks',
    'aws_dynamodb',
  ],
  GCP: [
    'gcp_iam_policy',
    'gcp_iam_role',
    'gcp_compute',
    'gcp_storage',
    'gcp_gke',
    'gcp_cloud_functions',
    'gcp_vpc',
  ],
  Kubernetes: [
    'kubernetes_pod',
    'kubernetes_service_account',
    'kubernetes_container',
  ],
  Special: [
    'terraform_change',
    'falco_event',
    'drift',
    'network',
  ],
};

/**
 * Get color for resource category
 */
export function getCategoryColor(resourceType: string): string {
  if (resourceType.startsWith('aws_')) return '#FF9900';
  if (resourceType.startsWith('gcp_')) return '#4285F4';
  if (resourceType.startsWith('kubernetes_')) return '#326CE5';
  if (resourceType === 'terraform_change') return '#623CE4';
  if (resourceType === 'falco_event') return '#EC4C39';
  return '#9C27B0';
}
