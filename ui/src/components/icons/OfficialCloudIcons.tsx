import React from 'react';

// Official AWS React Icons from aws-react-icons package
// Based on AWS Architecture Icons | Version Q1 2025
import {
  // AWS IAM Icons
  ArchitectureServiceAWSIAMIdentityCenter,
  ResourceAWSIdentityAccessManagementIAMAccessAnalyzer,
  ResourceAWSIdentityAccessManagementIAMRolesAnywhere,

  // AWS Compute Icons
  ArchitectureServiceAWSLambda,
  ArchitectureServiceAmazonEC2,
  ArchitectureServiceAmazonElasticContainerService,
  ArchitectureServiceAmazonEKSCloud,

  // AWS Storage Icons
  ResourceAmazonSimpleStorageServiceBucket,
  ArchitectureServiceAmazonElasticBlockStore,

  // AWS Database Icons
  ArchitectureServiceAmazonRDS,
  ArchitectureServiceAmazonDynamoDB,

  // AWS Networking Icons
  ArchitectureGroupVirtualprivatecloudVPC,
  ArchitectureServiceElasticLoadBalancing,
  ArchitectureServiceAmazonVPCLattice,
} from 'aws-react-icons';

interface OfficialIconProps {
  type: string;
  size?: number;
  className?: string;
}

interface InlineIconProps {
  size?: number;
  className?: string;
}

type AWSIconComponent = React.ComponentType<{ size?: number; className?: string }>;

// ============================================================
// AWS Icon Mapping - Official aws-react-icons components
// ============================================================
const AWS_ICON_MAPPING: Record<string, AWSIconComponent> = {
  // IAM
  'aws_iam_policy': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_role': ResourceAWSIdentityAccessManagementIAMRolesAnywhere,
  'aws_iam_user': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_group': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_access_analyzer': ResourceAWSIdentityAccessManagementIAMAccessAnalyzer,
  'aws_iam_instance_profile': ResourceAWSIdentityAccessManagementIAMRolesAnywhere,
  'aws_iam_policy_attachment': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_role_policy': ResourceAWSIdentityAccessManagementIAMRolesAnywhere,
  'aws_iam_role_policy_attachment': ResourceAWSIdentityAccessManagementIAMRolesAnywhere,

  // Compute
  'aws_lambda_function': ArchitectureServiceAWSLambda,
  'aws_lambda': ArchitectureServiceAWSLambda,
  'aws_lambda_permission': ArchitectureServiceAWSLambda,
  'aws_lambda_layer_version': ArchitectureServiceAWSLambda,
  'aws_instance': ArchitectureServiceAmazonEC2,
  'aws_ec2_instance': ArchitectureServiceAmazonEC2,
  'ec2_instance': ArchitectureServiceAmazonEC2,
  'aws_launch_template': ArchitectureServiceAmazonEC2,
  'aws_autoscaling_group': ArchitectureServiceAmazonEC2,
  'aws_ecs_service': ArchitectureServiceAmazonElasticContainerService,
  'aws_ecs_cluster': ArchitectureServiceAmazonElasticContainerService,
  'aws_ecs_task_definition': ArchitectureServiceAmazonElasticContainerService,
  'aws_eks_cluster': ArchitectureServiceAmazonEKSCloud,
  'aws_eks_node_group': ArchitectureServiceAmazonEKSCloud,

  // Storage
  'aws_s3_bucket': ResourceAmazonSimpleStorageServiceBucket,
  'aws_s3': ResourceAmazonSimpleStorageServiceBucket,
  'aws_s3_bucket_policy': ResourceAmazonSimpleStorageServiceBucket,
  'aws_s3_object': ResourceAmazonSimpleStorageServiceBucket,
  'aws_ebs_volume': ArchitectureServiceAmazonElasticBlockStore,

  // Networking
  'aws_vpc': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_subnet': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_security_group': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_security_group_rule': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_network_acl': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_internet_gateway': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_route_table': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_route': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_elb': ArchitectureServiceElasticLoadBalancing,
  'aws_lb': ArchitectureServiceElasticLoadBalancing,
  'aws_alb': ArchitectureServiceElasticLoadBalancing,
  'aws_lb_target_group': ArchitectureServiceElasticLoadBalancing,
  'aws_lb_listener': ArchitectureServiceElasticLoadBalancing,
  'load_balancer': ArchitectureServiceElasticLoadBalancing,
  'aws_nat_gateway': ArchitectureServiceAmazonVPCLattice,
  'nat_gateway': ArchitectureServiceAmazonVPCLattice,

  // Database
  'aws_db_instance': ArchitectureServiceAmazonRDS,
  'aws_rds_instance': ArchitectureServiceAmazonRDS,
  'rds_instance': ArchitectureServiceAmazonRDS,
  'aws_rds_cluster': ArchitectureServiceAmazonRDS,
  'aws_db_subnet_group': ArchitectureServiceAmazonRDS,
  'aws_dynamodb_table': ArchitectureServiceAmazonDynamoDB,
  'aws_dynamodb': ArchitectureServiceAmazonDynamoDB,
};

// ============================================================
// GCP Inline SVG Icons (replaces <img> tags that caused 404/mystery squares)
// ============================================================
const GCPIAMIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#4285F4"/>
    <circle cx="12" cy="9" r="3" fill="white"/>
    <path d="M17 18H7c-.55 0-1-.45-1-1v-.5c0-1.66 2.24-3 5-3h2c2.76 0 5 1.34 5 3v.5c0 .55-.45 1-1 1z" fill="white"/>
  </svg>
);

const GCPComputeIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#4285F4"/>
    <rect x="5" y="7" width="14" height="10" rx="1.5" stroke="white" strokeWidth="1.5" fill="none"/>
    <circle cx="8" cy="10" r="1" fill="white"/>
    <circle cx="11" cy="10" r="1" fill="white"/>
    <rect x="7" y="13" width="10" height="1.5" rx=".75" fill="white"/>
  </svg>
);

const GCPStorageIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#FBBC04"/>
    <path d="M12 5l6 3v8l-6 3-6-3V8l6-3z" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M6 8l6 3 6-3M12 11v8" stroke="white" strokeWidth="1.5"/>
  </svg>
);

const GCPGKEIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#4285F4"/>
    <circle cx="12" cy="12" r="7" stroke="white" strokeWidth="1.5" fill="none"/>
    <circle cx="12" cy="6.5" r="1.5" fill="white"/>
    <circle cx="12" cy="17.5" r="1.5" fill="white"/>
    <circle cx="6.5" cy="12" r="1.5" fill="white"/>
    <circle cx="17.5" cy="12" r="1.5" fill="white"/>
  </svg>
);

const GCPCloudFunctionsIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#4285F4"/>
    <path d="M8 5l6 7-6 7" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
    <path d="M13 5l6 7-6 7" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round" opacity=".6"/>
  </svg>
);

const GCPVPCIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#34A853"/>
    <rect x="4" y="8" width="6" height="8" rx="1.5" stroke="white" strokeWidth="1.5" fill="none"/>
    <rect x="14" y="8" width="6" height="8" rx="1.5" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M10 12h4" stroke="white" strokeWidth="1.5"/>
  </svg>
);

const GCPFirewallIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#EA4335"/>
    <path d="M12 4l7 4v6c0 3.5-3 6.5-7 8-4-1.5-7-4.5-7-8V8l7-4z" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M12 8v4m0 2v1" stroke="white" strokeWidth="1.5" strokeLinecap="round"/>
  </svg>
);

// ============================================================
// Kubernetes Inline SVG Icons
// ============================================================
const K8sPodIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="12" cy="12" r="3" fill="white"/>
  </svg>
);

const K8sServiceIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" stroke="white" strokeWidth="1.2" fill="none"/>
    <path d="M8 10h8M8 12h8M8 14h8" stroke="white" strokeWidth="1" strokeLinecap="round"/>
  </svg>
);

const K8sServiceAccountIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <circle cx="12" cy="9" r="3" fill="white"/>
    <path d="M7 18c0-2.76 2.24-5 5-5s5 2.24 5 5" stroke="white" strokeWidth="1.5" fill="none" strokeLinecap="round"/>
  </svg>
);

const K8sDeploymentIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="9" cy="11" r="2" fill="white"/>
    <circle cx="15" cy="11" r="2" fill="white"/>
    <circle cx="12" cy="15" r="2" fill="white"/>
  </svg>
);

const K8sNamespaceIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <rect x="4" y="4" width="16" height="16" rx="2" stroke="white" strokeWidth="1.5" fill="none" strokeDasharray="3 2"/>
    <text x="12" y="14" textAnchor="middle" fill="white" fontSize="7" fontWeight="bold">NS</text>
  </svg>
);

const K8sRoleIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <path d="M12 4l6 3.5v5c0 3-2.5 5.5-6 7-3.5-1.5-6-4-6-7v-5L12 4z" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M10 12l2 2 4-4" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
);

const ContainerIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#0DB7ED"/>
    <rect x="4" y="7" width="16" height="10" rx="1.5" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M4 11h16M4 14h16" stroke="white" strokeWidth="1"/>
    <circle cx="7" cy="9" r=".8" fill="white"/>
    <circle cx="9.5" cy="9" r=".8" fill="white"/>
    <circle cx="12" cy="9" r=".8" fill="white"/>
  </svg>
);

// ============================================================
// Special Icons (Terraform, Falco, Drift, Network)
// ============================================================
const TerraformIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#7B42BC"/>
    <path d="M5 5.5l4 2.3v7.4L5 13V5.5z" fill="white"/>
    <path d="M10 8l4 2.3v7.4L10 15.4V8z" fill="white"/>
    <path d="M15 8l4 2.3v7.4l-4-2.3V8z" fill="white"/>
    <path d="M10 .5l4 2.3v5L10 5.5V.5z" fill="white" opacity=".7"/>
  </svg>
);

const FalcoIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#00AEC7"/>
    <circle cx="12" cy="12" r="8" stroke="white" strokeWidth="1.5" fill="none"/>
    <path d="M12 6v6l4 4" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
    <circle cx="12" cy="12" r="2" fill="white"/>
  </svg>
);

const DriftIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#FF5722"/>
    <path d="M5 12h5" stroke="white" strokeWidth="2" strokeLinecap="round"/>
    <path d="M14 12h5" stroke="white" strokeWidth="2" strokeLinecap="round" strokeDasharray="3 2"/>
    <circle cx="12" cy="12" r="2" fill="white"/>
    <path d="M9 7l3-3 3 3M15 17l-3 3-3-3" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
);

const NetworkIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#9C27B0"/>
    <circle cx="8" cy="8" r="2.5" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="16" cy="8" r="2.5" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="8" cy="16" r="2.5" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="16" cy="16" r="2.5" stroke="white" strokeWidth="1.2" fill="none"/>
    <path d="M10.5 8h3M8 10.5v3M16 10.5v3M10.5 16h3" stroke="white" strokeWidth="1.2"/>
  </svg>
);

// ============================================================
// Provider-Specific Fallback Icons (no more generic blue squares!)
// ============================================================
const AWSFallbackIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#232F3E"/>
    <path d="M6 14.5c2.67 1.33 5.33 2 8 2s4-.5 5-1" stroke="#FF9900" strokeWidth="1.5" strokeLinecap="round"/>
    <path d="M18 14.5l1.5 1.5-2.5.5" stroke="#FF9900" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
    <text x="12" y="11" textAnchor="middle" fill="#FF9900" fontSize="6" fontWeight="bold">AWS</text>
  </svg>
);

const GCPFallbackIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#4285F4"/>
    <path d="M12 6l5.2 3v6L12 18l-5.2-3V9L12 6z" stroke="white" strokeWidth="1.5" fill="none"/>
    <text x="12" y="13.5" textAnchor="middle" fill="white" fontSize="5" fontWeight="bold">GCP</text>
  </svg>
);

const K8sFallbackIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#326CE5"/>
    <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z" stroke="white" strokeWidth="1.2" fill="none"/>
    <circle cx="12" cy="12" r="2" fill="white"/>
    <path d="M12 6v4M12 14v4M7.5 9l3.5 2M13 11l3.5-2M7.5 15l3.5-2M13 13l3.5 2" stroke="white" strokeWidth=".8"/>
  </svg>
);

const TerraformFallbackIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#7B42BC"/>
    <text x="12" y="14" textAnchor="middle" fill="white" fontSize="6" fontWeight="bold">TF</text>
  </svg>
);

const DefaultFallbackIcon: React.FC<InlineIconProps> = ({ size = 48, className = '' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" className={className}>
    <rect width="24" height="24" rx="2" fill="#607D8B"/>
    <path d="M19 11H5c-.55 0-1 .45-1 1s.45 1 1 1h14c.55 0 1-.45 1-1s-.45-1-1-1z" fill="white" opacity=".5"/>
    <circle cx="12" cy="12" r="4" stroke="white" strokeWidth="1.5" fill="none"/>
    <circle cx="12" cy="12" r="1.5" fill="white"/>
  </svg>
);

// ============================================================
// GCP Icon Mapping - Inline SVG components
// ============================================================
const GCP_ICON_MAPPING: Record<string, React.FC<InlineIconProps>> = {
  'gcp_iam_policy': GCPIAMIcon,
  'gcp_iam_role': GCPIAMIcon,
  'gcp_iam_member': GCPIAMIcon,
  'gcp_iam_binding': GCPIAMIcon,
  'gcp_service_account': GCPIAMIcon,
  'gcp_compute_instance': GCPComputeIcon,
  'gcp_compute': GCPComputeIcon,
  'gcp_compute_instance_template': GCPComputeIcon,
  'gcp_storage_bucket': GCPStorageIcon,
  'gcp_storage': GCPStorageIcon,
  'gcp_gke_cluster': GCPGKEIcon,
  'gcp_gke': GCPGKEIcon,
  'gcp_container_cluster': GCPGKEIcon,
  'gcp_cloud_function': GCPCloudFunctionsIcon,
  'gcp_cloud_functions': GCPCloudFunctionsIcon,
  'gcp_compute_network': GCPVPCIcon,
  'gcp_vpc': GCPVPCIcon,
  'gcp_compute_subnetwork': GCPVPCIcon,
  'gcp_compute_firewall': GCPFirewallIcon,
  'gcp_firewall': GCPFirewallIcon,
};

// ============================================================
// Kubernetes Icon Mapping - Inline SVG components
// ============================================================
const K8S_ICON_MAPPING: Record<string, React.FC<InlineIconProps>> = {
  'kubernetes_pod': K8sPodIcon,
  'kubernetes_service': K8sServiceIcon,
  'kubernetes_service_account': K8sServiceAccountIcon,
  'kubernetes_deployment': K8sDeploymentIcon,
  'kubernetes_namespace': K8sNamespaceIcon,
  'kubernetes_role': K8sRoleIcon,
  'kubernetes_cluster_role': K8sRoleIcon,
  'kubernetes_cluster_role_binding': K8sRoleIcon,
  'kubernetes_role_binding': K8sRoleIcon,
  'kubernetes_config_map': K8sServiceIcon,
  'kubernetes_secret': K8sRoleIcon,
  'kubernetes_ingress': K8sServiceIcon,
  'kubernetes_stateful_set': K8sDeploymentIcon,
  'kubernetes_daemon_set': K8sDeploymentIcon,
  'kubernetes_replica_set': K8sDeploymentIcon,
  'kubernetes_job': K8sPodIcon,
  'kubernetes_cron_job': K8sPodIcon,
  'kubernetes_container': ContainerIcon,
  'container': ContainerIcon,
};

// ============================================================
// Special Icon Mapping
// ============================================================
const SPECIAL_ICON_MAPPING: Record<string, React.FC<InlineIconProps>> = {
  'terraform_change': TerraformIcon,
  'terraform': TerraformIcon,
  'terraform_resource': TerraformIcon,
  'falco_event': FalcoIcon,
  'falco': FalcoIcon,
  'falco_alert': FalcoIcon,
  'drift': DriftIcon,
  'drift_detected': DriftIcon,
  'network': NetworkIcon,
  'security_group': NetworkIcon,
};

// ============================================================
// Main OfficialCloudIcon Component
// Single unified icon system - replaces all legacy icon systems
// ============================================================
export const OfficialCloudIcon: React.FC<OfficialIconProps> = ({
  type,
  size = 32,
  className = ''
}) => {
  const normalizedType = type.toLowerCase().replace(/^k8s_/, 'kubernetes_');

  // 1. Check AWS icons (official aws-react-icons)
  const AwsIconComponent = AWS_ICON_MAPPING[normalizedType];
  if (AwsIconComponent) {
    return <AwsIconComponent size={size} className={className} />;
  }

  // 2. Check GCP icons (inline SVG)
  const GcpIconComponent = GCP_ICON_MAPPING[normalizedType];
  if (GcpIconComponent) {
    return <GcpIconComponent size={size} className={className} />;
  }

  // 3. Check K8s icons (inline SVG)
  const K8sIconComponent = K8S_ICON_MAPPING[normalizedType];
  if (K8sIconComponent) {
    return <K8sIconComponent size={size} className={className} />;
  }

  // 4. Check special icons (inline SVG)
  const SpecialIconComponent = SPECIAL_ICON_MAPPING[normalizedType];
  if (SpecialIconComponent) {
    return <SpecialIconComponent size={size} className={className} />;
  }

  // 5. Provider-specific fallback (no more generic blue squares!)
  if (normalizedType.startsWith('aws_')) {
    return <AWSFallbackIcon size={size} className={className} />;
  }
  if (normalizedType.startsWith('gcp_') || normalizedType.startsWith('google_')) {
    return <GCPFallbackIcon size={size} className={className} />;
  }
  if (normalizedType.startsWith('kubernetes_') || normalizedType.startsWith('k8s_')) {
    return <K8sFallbackIcon size={size} className={className} />;
  }
  if (normalizedType.startsWith('terraform_') || normalizedType.startsWith('tf_')) {
    return <TerraformFallbackIcon size={size} className={className} />;
  }

  // 6. Generic fallback
  return <DefaultFallbackIcon size={size} className={className} />;
};

// ============================================================
// Helper: Get icon path (for Cytoscape background-image)
// Now returns data URI SVGs instead of file paths
// ============================================================
// eslint-disable-next-line react-refresh/only-export-components
export const getOfficialIconPath = (type: string): string => {
  // Return provider-specific placeholder SVG data URIs
  const normalizedType = type.toLowerCase().replace(/^k8s_/, 'kubernetes_');

  if (normalizedType.startsWith('aws_')) {
    return "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23232F3E"/><text x="12" y="14" text-anchor="middle" fill="%23FF9900" font-size="6" font-weight="bold">AWS</text></svg>');
  }
  if (normalizedType.startsWith('gcp_') || normalizedType.startsWith('google_')) {
    return "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%234285F4"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="5" font-weight="bold">GCP</text></svg>');
  }
  if (normalizedType.startsWith('kubernetes_') || normalizedType.startsWith('k8s_')) {
    return "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23326CE5"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="5" font-weight="bold">K8s</text></svg>');
  }
  if (normalizedType.startsWith('terraform_') || normalizedType.startsWith('tf_')) {
    return "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%237B42BC"/><text x="12" y="14" text-anchor="middle" fill="white" font-size="6" font-weight="bold">TF</text></svg>');
  }

  return "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect width="24" height="24" rx="2" fill="%23607D8B"/><circle cx="12" cy="12" r="4" stroke="white" stroke-width="1.5" fill="none"/></svg>');
};

// ============================================================
// Exported mappings for use in other components
// ============================================================
// eslint-disable-next-line react-refresh/only-export-components
export const ALL_ICON_MAPPINGS = {
  ...Object.fromEntries(Object.keys(AWS_ICON_MAPPING).map(k => [k, 'aws'])),
  ...Object.fromEntries(Object.keys(GCP_ICON_MAPPING).map(k => [k, 'gcp'])),
  ...Object.fromEntries(Object.keys(K8S_ICON_MAPPING).map(k => [k, 'kubernetes'])),
  ...Object.fromEntries(Object.keys(SPECIAL_ICON_MAPPING).map(k => [k, 'special'])),
};

// eslint-disable-next-line react-refresh/only-export-components
export const getProviderFromType = (type: string): string => {
  const t = type.toLowerCase();
  if (t.startsWith('aws_')) return 'aws';
  if (t.startsWith('gcp_') || t.startsWith('google_')) return 'gcp';
  if (t.startsWith('kubernetes_') || t.startsWith('k8s_')) return 'kubernetes';
  if (t.startsWith('terraform_') || t.startsWith('tf_')) return 'terraform';
  if (t.startsWith('falco')) return 'falco';
  return 'unknown';
};

// eslint-disable-next-line react-refresh/only-export-components
export const getProviderColor = (provider: string): string => {
  switch (provider) {
    case 'aws': return '#FF9900';
    case 'gcp': return '#4285F4';
    case 'kubernetes': return '#326CE5';
    case 'terraform': return '#7B42BC';
    case 'falco': return '#00AEC7';
    default: return '#607D8B';
  }
};
