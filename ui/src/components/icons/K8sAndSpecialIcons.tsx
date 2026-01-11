/**
 * Kubernetes and Special Icons (Terraform, Falco)
 *
 * Icons for Kubernetes resources and TFDrift-Falco specific resources
 */

import React from 'react';

interface IconProps {
  size?: number;
  className?: string;
}

/**
 * Kubernetes Pod Icon
 */
export const K8sPodIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#326CE5" />
    <circle cx="24" cy="24" r="12" stroke="white" strokeWidth="2" fill="none" />
    <path
      d="M24 14L28 24L24 34L20 24L24 14Z"
      fill="white"
    />
    <circle cx="24" cy="24" r="3" fill="white" />
  </svg>
);

/**
 * Kubernetes Service Account Icon
 */
export const K8sServiceAccountIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#326CE5" />
    <circle cx="24" cy="20" r="5" fill="white" />
    <path
      d="M31 30H17C15.8954 30 15 30.8954 15 32V33C15 33.5523 15.4477 34 16 34H32C32.5523 34 33 33.5523 33 33V32C33 30.8954 32.1046 30 31 30Z"
      fill="white"
    />
    <path
      d="M18 32L24 28L30 32"
      stroke="#326CE5"
      strokeWidth="2"
    />
  </svg>
);

/**
 * Container Icon
 */
export const ContainerIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#0DB7ED" />
    <rect x="12" y="16" width="24" height="16" rx="2" stroke="white" strokeWidth="2" fill="none" />
    <path d="M12 22H36" stroke="white" strokeWidth="2" />
    <path d="M12 26H36" stroke="white" strokeWidth="2" />
    <circle cx="16" cy="19" r="1" fill="white" />
    <circle cx="20" cy="19" r="1" fill="white" />
    <circle cx="24" cy="19" r="1" fill="white" />
  </svg>
);

/**
 * Terraform Change Icon
 */
export const TerraformIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#623CE4" />
    <path
      d="M16 14L20 16V32L16 30V14Z"
      fill="white"
    />
    <path
      d="M22 16L26 18V34L22 32V16Z"
      fill="white"
    />
    <path
      d="M28 16L32 18V34L28 32V16Z"
      fill="white"
    />
    <path
      d="M16 14L26 20"
      stroke="#623CE4"
      strokeWidth="0.5"
    />
  </svg>
);

/**
 * Falco Event Icon
 */
export const FalcoIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#EC4C39" />
    <circle cx="24" cy="24" r="14" stroke="white" strokeWidth="2" fill="none" />
    <path
      d="M24 16L28 24L24 32L20 24L24 16Z"
      fill="white"
    />
    <circle cx="24" cy="24" r="3" fill="#EC4C39" />
    <path
      d="M24 16V12M24 36V32M16 24H12M36 24H32"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
);

/**
 * Network/Security Group Icon
 */
export const NetworkIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#9C27B0" />
    <circle cx="18" cy="18" r="4" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="30" cy="18" r="4" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="18" cy="30" r="4" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="30" cy="30" r="4" stroke="white" strokeWidth="2" fill="none" />
    <path
      d="M22 18H26M18 22V26M30 22V26M22 30H26"
      stroke="white"
      strokeWidth="2"
    />
  </svg>
);

/**
 * Drift Detection Icon
 */
export const DriftIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#FF5722" />
    <path
      d="M14 24H22"
      stroke="white"
      strokeWidth="3"
      strokeLinecap="round"
    />
    <path
      d="M26 24H34"
      stroke="white"
      strokeWidth="3"
      strokeLinecap="round"
      strokeDasharray="4 2"
    />
    <circle cx="24" cy="24" r="3" fill="white" />
    <path
      d="M20 16L24 12L28 16"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M28 32L24 36L20 32"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

// eslint-disable-next-line react-refresh/only-export-components
export const K8sAndSpecialIcons = {
  'kubernetes_pod': K8sPodIcon,
  'kubernetes_service_account': K8sServiceAccountIcon,
  'kubernetes_container': ContainerIcon,
  'container': ContainerIcon,
  'terraform_change': TerraformIcon,
  'terraform': TerraformIcon,
  'falco_event': FalcoIcon,
  'falco': FalcoIcon,
  'network': NetworkIcon,
  'security_group': NetworkIcon,
  'drift': DriftIcon,
};
