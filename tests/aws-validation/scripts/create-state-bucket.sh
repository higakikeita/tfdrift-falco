#!/usr/bin/env bash
# Create the S3 bucket for Terraform state (run once before Phase 1)
set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-draios-dev-developer}"
AWS_REGION="${AWS_REGION:-ap-northeast-1}"
BUCKET_NAME="tfdrift-validation-state"

echo "Creating TF state bucket: ${BUCKET_NAME} in ${AWS_REGION}..."

aws s3api create-bucket \
  --bucket "${BUCKET_NAME}" \
  --region "${AWS_REGION}" \
  --create-bucket-configuration LocationConstraint="${AWS_REGION}" \
  --profile "${AWS_PROFILE}" 2>/dev/null || echo "Bucket already exists or creation failed"

aws s3api put-bucket-versioning \
  --bucket "${BUCKET_NAME}" \
  --versioning-configuration Status=Enabled \
  --profile "${AWS_PROFILE}"

aws s3api put-bucket-encryption \
  --bucket "${BUCKET_NAME}" \
  --server-side-encryption-configuration '{
    "Rules": [{"ApplyServerSideEncryptionByDefault": {"SSEAlgorithm": "AES256"}}]
  }' \
  --profile "${AWS_PROFILE}"

aws s3api put-public-access-block \
  --bucket "${BUCKET_NAME}" \
  --public-access-block-configuration '{
    "BlockPublicAcls": true,
    "IgnorePublicAcls": true,
    "BlockPublicPolicy": true,
    "RestrictPublicBuckets": true
  }' \
  --profile "${AWS_PROFILE}"

echo "State bucket ready: s3://${BUCKET_NAME}"
