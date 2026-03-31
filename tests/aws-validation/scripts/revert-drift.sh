#!/usr/bin/env bash
# Revert intentional drift back to Terraform-managed state
set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-draios-dev-developer}"
AWS_REGION="${AWS_REGION:-ap-northeast-1}"

cd "$(dirname "$0")/../phase1-resources"
SG_ID=$(terraform output -raw security_group_id)
INSTANCE_ID=$(terraform output -raw instance_id)
BUCKET_NAME=$(terraform output -raw s3_bucket_name)

echo "=== Reverting Drift ==="

echo "[1] Removing unauthorized SG rule (port 8443)..."
aws ec2 revoke-security-group-ingress \
  --group-id "${SG_ID}" \
  --protocol tcp \
  --port 8443 \
  --cidr "10.0.0.0/8" \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" 2>/dev/null && echo "  -> Removed" || echo "  -> Not found"

echo "[2] Disabling detailed monitoring..."
aws ec2 unmonitor-instances \
  --instance-ids "${INSTANCE_ID}" \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" 2>/dev/null && echo "  -> Disabled" || echo "  -> Failed"

echo "[3] Removing ManualChange tag..."
aws ec2 delete-tags \
  --resources "${INSTANCE_ID}" \
  --tags Key=ManualChange \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" 2>/dev/null && echo "  -> Removed" || echo "  -> Not found"

echo "[4] Re-enabling S3 versioning..."
aws s3api put-bucket-versioning \
  --bucket "${BUCKET_NAME}" \
  --versioning-configuration Status=Enabled \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" 2>/dev/null && echo "  -> Enabled" || echo "  -> Failed"

echo ""
echo "=== Drift Reverted ==="
echo "Or use: cd ../phase1-resources && terraform apply"
