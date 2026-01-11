/* eslint-disable @typescript-eslint/no-explicit-any */
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

// Map resource types to official AWS React Icon components
const AWS_ICON_MAPPING: Record<string, React.ComponentType<any>> = {
  // AWS IAM
  'aws_iam_policy': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_role': ResourceAWSIdentityAccessManagementIAMRolesAnywhere,
  'aws_iam_user': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_group': ArchitectureServiceAWSIAMIdentityCenter,
  'aws_iam_access_analyzer': ResourceAWSIdentityAccessManagementIAMAccessAnalyzer,

  // AWS Compute
  'aws_lambda_function': ArchitectureServiceAWSLambda,
  'aws_lambda': ArchitectureServiceAWSLambda,
  'aws_instance': ArchitectureServiceAmazonEC2,
  'aws_ec2_instance': ArchitectureServiceAmazonEC2,
  'ec2_instance': ArchitectureServiceAmazonEC2,
  'aws_ecs_service': ArchitectureServiceAmazonElasticContainerService,
  'aws_eks_cluster': ArchitectureServiceAmazonEKSCloud,

  // AWS Storage
  'aws_s3_bucket': ResourceAmazonSimpleStorageServiceBucket,
  'aws_s3': ResourceAmazonSimpleStorageServiceBucket,
  'aws_ebs_volume': ArchitectureServiceAmazonElasticBlockStore,

  // AWS Networking
  'aws_vpc': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_subnet': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_security_group': ArchitectureGroupVirtualprivatecloudVPC,
  'aws_elb': ArchitectureServiceElasticLoadBalancing,
  'aws_lb': ArchitectureServiceElasticLoadBalancing,
  'aws_alb': ArchitectureServiceElasticLoadBalancing,
  'load_balancer': ArchitectureServiceElasticLoadBalancing,
  'aws_nat_gateway': ArchitectureServiceAmazonVPCLattice,
  'nat_gateway': ArchitectureServiceAmazonVPCLattice,

  // AWS Database
  'aws_db_instance': ArchitectureServiceAmazonRDS,
  'aws_rds_instance': ArchitectureServiceAmazonRDS,
  'rds_instance': ArchitectureServiceAmazonRDS,
  'aws_dynamodb_table': ArchitectureServiceAmazonDynamoDB,
};

// For GCP and K8s, we'll use custom SVG paths for now
const GCP_K8S_ICON_MAPPING: Record<string, { src: string; alt: string }> = {
  // GCP IAM
  'gcp_iam_policy': {
    src: '/icons/gcp/iam.svg',
    alt: 'GCP IAM'
  },
  'gcp_iam_role': {
    src: '/icons/gcp/iam.svg',
    alt: 'GCP IAM Role'
  },
  'gcp_service_account': {
    src: '/icons/gcp/service-account.svg',
    alt: 'GCP Service Account'
  },

  // GCP Compute
  'gcp_compute_instance': {
    src: '/icons/gcp/compute-engine.svg',
    alt: 'GCP Compute Engine'
  },
  'gcp_gke_cluster': {
    src: '/icons/gcp/gke.svg',
    alt: 'GCP GKE'
  },
  'gcp_cloud_function': {
    src: '/icons/gcp/cloud-functions.svg',
    alt: 'GCP Cloud Functions'
  },

  // GCP Storage
  'gcp_storage_bucket': {
    src: '/icons/gcp/cloud-storage.svg',
    alt: 'GCP Cloud Storage'
  },

  // GCP Networking
  'gcp_compute_network': {
    src: '/icons/gcp/vpc.svg',
    alt: 'GCP VPC'
  },
  'gcp_compute_firewall': {
    src: '/icons/gcp/firewall.svg',
    alt: 'GCP Firewall'
  },

  // Kubernetes - Official icons from kubernetes/community
  'kubernetes_pod': {
    src: '/icons/k8s/pod-official.svg',
    alt: 'Kubernetes Pod'
  },
  'kubernetes_service': {
    src: '/icons/k8s/service-official.svg',
    alt: 'Kubernetes Service'
  },
  'kubernetes_service_account': {
    src: '/icons/k8s/sa-official.svg',
    alt: 'Kubernetes Service Account'
  },
  'kubernetes_deployment': {
    src: '/icons/k8s/deploy-official.svg',
    alt: 'Kubernetes Deployment'
  },
  'kubernetes_namespace': {
    src: '/icons/k8s/ns-official.svg',
    alt: 'Kubernetes Namespace'
  },
  'kubernetes_role': {
    src: '/icons/k8s/role-official.svg',
    alt: 'Kubernetes Role'
  },
  'kubernetes_cluster_role': {
    src: '/icons/k8s/role-official.svg',
    alt: 'Kubernetes ClusterRole'
  },

  // Terraform
  'terraform_change': {
    src: '/icons/terraform.svg',
    alt: 'Terraform'
  },
  'terraform_resource': {
    src: '/icons/terraform.svg',
    alt: 'Terraform Resource'
  },
};

// Fallback icons
const FALLBACK_ICONS: Record<string, { src: string; alt: string }> = {
  'aws': {
    src: '/icons/aws/aws.svg',
    alt: 'AWS'
  },
  'gcp': {
    src: '/icons/gcp/gcp.svg',
    alt: 'Google Cloud'
  },
  'kubernetes': {
    src: '/icons/k8s/kubernetes-official.svg',
    alt: 'Kubernetes'
  },
  'terraform': {
    src: '/icons/terraform.svg',
    alt: 'Terraform'
  },
  'default': {
    src: '/icons/cloud.svg',
    alt: 'Cloud Resource'
  }
};

export const OfficialCloudIcon: React.FC<OfficialIconProps> = ({
  type,
  size = 32,
  className = ''
}) => {
  // Check if it's an AWS icon with official React component
  const AwsIconComponent = AWS_ICON_MAPPING[type];
  if (AwsIconComponent) {
    return <AwsIconComponent size={size} className={className} />;
  }

  // Check if it's a GCP or K8s icon (using SVG)
  const iconData = GCP_K8S_ICON_MAPPING[type];
  if (iconData) {
    return (
      <img
        src={iconData.src}
        alt={iconData.alt}
        width={size}
        height={size}
        className={`inline-block ${className}`}
        style={{ objectFit: 'contain' }}
      />
    );
  }

  // Determine fallback
  let fallbackIcon = FALLBACK_ICONS['default'];
  if (type.startsWith('aws_')) {
    fallbackIcon = FALLBACK_ICONS['aws'];
  } else if (type.startsWith('gcp_')) {
    fallbackIcon = FALLBACK_ICONS['gcp'];
  } else if (type.startsWith('kubernetes_')) {
    fallbackIcon = FALLBACK_ICONS['kubernetes'];
  } else if (type.startsWith('terraform_')) {
    fallbackIcon = FALLBACK_ICONS['terraform'];
  }

  return (
    <img
      src={fallbackIcon.src}
      alt={fallbackIcon.alt}
      width={size}
      height={size}
      className={`inline-block ${className}`}
      style={{ objectFit: 'contain' }}
      onError={(e) => {
        // Fallback to default if image fails to load
        const target = e.target as HTMLImageElement;
        if (target.src !== FALLBACK_ICONS['default'].src) {
          target.src = FALLBACK_ICONS['default'].src;
        }
      }}
    />
  );
};

// Helper function to get icon path for a resource type
// eslint-disable-next-line react-refresh/only-export-components
export const getOfficialIconPath = (type: string): string => {
  const iconData = GCP_K8S_ICON_MAPPING[type];
  if (iconData) {
    return iconData.src;
  }

  // Determine fallback
  if (type.startsWith('aws_')) {
    return FALLBACK_ICONS['aws'].src;
  } else if (type.startsWith('gcp_')) {
    return FALLBACK_ICONS['gcp'].src;
  } else if (type.startsWith('kubernetes_')) {
    return FALLBACK_ICONS['kubernetes'].src;
  } else if (type.startsWith('terraform_')) {
    return FALLBACK_ICONS['terraform'].src;
  }

  return FALLBACK_ICONS['default'].src;
};
