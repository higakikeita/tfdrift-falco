/**
 * GCP Service Icons (SVG Components)
 *
 * Simplified versions of GCP official service icons for graph visualization
 */

import React from 'react';

interface IconProps {
  size?: number;
  className?: string;
}

/**
 * GCP IAM Icon
 */
export const GCPIAMIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#4285F4" />
    <path
      d="M24 16C27.3137 16 30 18.6863 30 22C30 25.3137 27.3137 28 24 28C20.6863 28 18 25.3137 18 22C18 18.6863 20.6863 16 24 16Z"
      fill="white"
    />
    <path
      d="M31 30H17C15.8954 30 15 30.8954 15 32V33C15 33.5523 15.4477 34 16 34H32C32.5523 34 33 33.5523 33 33V32C33 30.8954 32.1046 30 31 30Z"
      fill="white"
    />
  </svg>
);

/**
 * GCP Compute Engine Icon
 */
export const GCPComputeIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#4285F4" />
    <rect x="14" y="16" width="20" height="16" rx="2" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="18" cy="20" r="1.5" fill="white" />
    <circle cx="22" cy="20" r="1.5" fill="white" />
    <circle cx="26" cy="20" r="1.5" fill="white" />
    <rect x="16" y="24" width="16" height="2" rx="1" fill="white" />
    <rect x="16" y="28" width="12" height="2" rx="1" fill="white" />
  </svg>
);

/**
 * GCP Cloud Storage Icon
 */
export const GCPStorageIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#FBBC04" />
    <path
      d="M24 14L32 18V30L24 34L16 30V18L24 14Z"
      stroke="white"
      strokeWidth="2"
      fill="none"
    />
    <path
      d="M16 18L24 22L32 18"
      stroke="white"
      strokeWidth="2"
    />
    <path
      d="M24 22V34"
      stroke="white"
      strokeWidth="2"
    />
  </svg>
);

/**
 * GCP GKE Icon
 */
export const GCPGKEIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#4285F4" />
    <circle cx="24" cy="24" r="10" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="24" cy="16" r="2" fill="white" />
    <circle cx="24" cy="32" r="2" fill="white" />
    <circle cx="16" cy="24" r="2" fill="white" />
    <circle cx="32" cy="24" r="2" fill="white" />
    <circle cx="19" cy="19" r="1.5" fill="white" />
    <circle cx="29" cy="19" r="1.5" fill="white" />
    <circle cx="19" cy="29" r="1.5" fill="white" />
    <circle cx="29" cy="29" r="1.5" fill="white" />
  </svg>
);

/**
 * GCP Cloud Functions Icon
 */
export const GCPCloudFunctionsIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#4285F4" />
    <path
      d="M20 14L28 24L20 34V14Z"
      fill="white"
    />
    <path
      d="M28 14L36 24L28 34V14Z"
      fill="white"
      opacity="0.7"
    />
  </svg>
);

/**
 * GCP VPC Icon
 */
export const GCPVPCIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#34A853" />
    <rect x="14" y="18" width="8" height="12" rx="2" stroke="white" strokeWidth="2" fill="none" />
    <rect x="26" y="18" width="8" height="12" rx="2" stroke="white" strokeWidth="2" fill="none" />
    <path d="M22 24H26" stroke="white" strokeWidth="2" />
  </svg>
);

// eslint-disable-next-line react-refresh/only-export-components
export const GCPServiceIcons = {
  'gcp_iam_policy': GCPIAMIcon,
  'gcp_iam_role': GCPIAMIcon,
  'gcp_compute': GCPComputeIcon,
  'gcp_storage': GCPStorageIcon,
  'gcp_gke': GCPGKEIcon,
  'gcp_cloud_functions': GCPCloudFunctionsIcon,
  'gcp_vpc': GCPVPCIcon,
};
