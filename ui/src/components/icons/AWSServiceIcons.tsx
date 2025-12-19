/**
 * AWS Service Icons (SVG Components)
 *
 * Simplified versions of AWS official service icons for graph visualization
 */

import React from 'react';

interface IconProps {
  size?: number;
  className?: string;
}

/**
 * AWS IAM Icon
 */
export const AWSIAMIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#DD344C" />
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
 * AWS Lambda Icon
 */
export const AWSLambdaIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#FF9900" />
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
 * AWS S3 Icon
 */
export const AWSS3Icon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#569A31" />
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
 * AWS EC2 Icon
 */
export const AWSEC2Icon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#FF9900" />
    <rect x="14" y="16" width="20" height="16" rx="2" fill="white" />
    <rect x="16" y="18" width="4" height="4" fill="#FF9900" />
    <rect x="22" y="18" width="4" height="4" fill="#FF9900" />
    <rect x="28" y="18" width="4" height="4" fill="#FF9900" />
    <rect x="16" y="24" width="16" height="2" fill="#FF9900" />
  </svg>
);

/**
 * AWS Security Group Icon
 */
export const AWSSecurityGroupIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#DD344C" />
    <path
      d="M24 14L30 17V26C30 29 27 32 24 34C21 32 18 29 18 26V17L24 14Z"
      fill="white"
    />
    <path
      d="M24 20V28M20 24H28"
      stroke="#DD344C"
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
);

/**
 * AWS EKS Icon
 */
export const AWSEKSIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#FF9900" />
    <circle cx="24" cy="24" r="8" stroke="white" strokeWidth="2" fill="none" />
    <circle cx="24" cy="16" r="2" fill="white" />
    <circle cx="24" cy="32" r="2" fill="white" />
    <circle cx="16" cy="24" r="2" fill="white" />
    <circle cx="32" cy="24" r="2" fill="white" />
  </svg>
);

/**
 * AWS DynamoDB Icon
 */
export const AWSDynamoDBIcon: React.FC<IconProps> = ({ size = 48, className = '' }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 48 48"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className={className}
  >
    <rect width="48" height="48" rx="4" fill="#4053D6" />
    <ellipse cx="24" cy="20" rx="10" ry="4" fill="white" />
    <path
      d="M14 20V28C14 30.2091 18.4772 32 24 32C29.5228 32 34 30.2091 34 28V20"
      stroke="white"
      strokeWidth="2"
      fill="none"
    />
  </svg>
);

export const AWSServiceIcons = {
  'aws_iam_policy': AWSIAMIcon,
  'aws_iam_role': AWSIAMIcon,
  'aws_lambda': AWSLambdaIcon,
  'aws_s3': AWSS3Icon,
  'aws_ec2': AWSEC2Icon,
  'aws_security_group': AWSSecurityGroupIcon,
  'aws_eks': AWSEKSIcon,
  'aws_dynamodb': AWSDynamoDBIcon,
};
