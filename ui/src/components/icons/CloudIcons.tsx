/**
 * Cloud Provider Icons Component
 *
 * AWS and GCP official icon mappings for TFDrift-Falco resources
 */

import {
  SiAmazon,
  SiKubernetes,
  SiTerraform
} from 'react-icons/si';
import {
  FaShieldAlt,
  FaKey,
  FaUsers,
  FaNetworkWired,
  FaServer,
  FaDatabase,
  FaCube,
  FaExclamationTriangle
} from 'react-icons/fa';

export type ResourceType =
  | 'terraform_change'
  | 'aws_iam_policy'
  | 'aws_iam_role'
  | 'aws_security_group'
  | 'aws_lambda'
  | 'aws_s3'
  | 'aws_dynamodb'
  | 'aws_ec2'
  | 'aws_eks'
  | 'gcp_iam_policy'
  | 'gcp_iam_role'
  | 'gcp_compute'
  | 'gcp_gke'
  | 'kubernetes_service_account'
  | 'kubernetes_pod'
  | 'kubernetes_container'
  | 'falco_event';

interface CloudIconProps {
  resourceType: ResourceType;
  size?: number;
  className?: string;
}

/**
 * AWS Resource Icons
 */
const AWSIcons: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  'aws_iam_policy': FaKey,
  'aws_iam_role': FaUsers,
  'aws_security_group': FaShieldAlt,
  'aws_lambda': SiAmazon,
  'aws_s3': SiAmazon,
  'aws_dynamodb': FaDatabase,
  'aws_ec2': FaServer,
  'aws_eks': SiKubernetes,
};

/**
 * GCP Resource Icons
 */
const GCPIcons: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  'gcp_iam_policy': FaKey,
  'gcp_iam_role': FaUsers,
  'gcp_compute': FaServer,
  'gcp_gke': SiKubernetes,
};

/**
 * Kubernetes Resource Icons
 */
const K8sIcons: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  'kubernetes_service_account': SiKubernetes,
  'kubernetes_pod': FaCube,
  'kubernetes_container': FaCube,
};

/**
 * Special Resource Icons
 */
const SpecialIcons: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
  'terraform_change': SiTerraform,
  'falco_event': FaExclamationTriangle,
};

/**
 * CloudIcon Component
 *
 * Renders the appropriate icon for a given cloud resource type
 */
export const CloudIcon: React.FC<CloudIconProps> = ({
  resourceType,
  size = 24,
  className = ''
}) => {
  // Combine all icon mappings
  const iconMap = {
    ...AWSIcons,
    ...GCPIcons,
    ...K8sIcons,
    ...SpecialIcons
  };

  const IconComponent = iconMap[resourceType] || FaNetworkWired;

  return <IconComponent size={size} className={className} />;
};

/**
 * Inline SVG Icons for Cytoscape.js nodes
 * These are base64 encoded SVGs for use in Cytoscape node backgrounds
 */
// eslint-disable-next-line react-refresh/only-export-components
export const CloudIconSVGs = {
  aws_iam: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI4IiBmaWxsPSIjRkY5OTAwIi8+CiAgPHBhdGggZD0iTTI0IDEyTDM2IDI0TDI0IDM2TDEyIDI0TDI0IDEyWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+',

  gcp_iam: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI4IiBmaWxsPSIjNDI4NUY0Ii8+CiAgPHBhdGggZD0iTTI0IDEyTDM2IDI0TDI0IDM2TDEyIDI0TDI0IDEyWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+',

  kubernetes: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8Y2lyY2xlIGN4PSIyNCIgY3k9IjI0IiByPSIyMCIgZmlsbD0iIzMyNkNFNSIvPgogIDxwYXRoIGQ9Ik0yNCAxMkwzMiAyNEgyNEwxNiAyNEwyNCAxMloiIGZpbGw9IndoaXRlIi8+CiAgPHBhdGggZD0iTTI0IDM2TDMyIDI0SDI0TDE2IDI0TDI0IDM2WiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+',

  terraform: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHJ4PSI4IiBmaWxsPSIjNjIzQ0U0Ii8+CiAgPHBhdGggZD0iTTE2IDEySDE2LjhWMzZIMTZWMTJaIiBmaWxsPSJ3aGl0ZSIvPgogIDxwYXRoIGQ9Ik0yMy4yIDEySDI0VjI0SDE2VjIzLjJMMjMuMiAxMloiIGZpbGw9IndoaXRlIi8+CiAgPHBhdGggZD0iTTMxLjIgMTJIMzJWMzZIMjRWMjRMMzEuMiAxMloiIGZpbGw9IndoaXRlIi8+Cjwvc3ZnPg==',

  falco: 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNDgiIGhlaWdodD0iNDgiIHZpZXdCb3g9IjAgMCA0OCA0OCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8Y2lyY2xlIGN4PSIyNCIgY3k9IjI0IiByPSIyMCIgZmlsbD0iI0VDNEMzOSIvPgogIDxwYXRoIGQ9Ik0yNCAxNkwyOCAyNEgyNEwyMCAyNEwyNCAxNloiIGZpbGw9IndoaXRlIi8+CiAgPHBhdGggZD0iTTI0IDMyTDIwIDI0SDI0TDI4IDI0TDI0IDMyWiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+'
};

export default CloudIcon;
