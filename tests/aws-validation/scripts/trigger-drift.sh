#!/usr/bin/env bash
# Trigger intentional drift on Phase 1 resources
# This simulates real-world drift that tfdrift-falco should detect
set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-draios-dev-developer}"
AWS_REGION="${AWS_REGION:-ap-northeast-1}"

# Get resource IDs from Terraform state
cd "$(dirname "$0")/../phase1-resources"
SG_ID=$(terraform output -raw security_group_id)
INSTANCE_ID=$(terraform output -raw instance_id)
BUCKET_NAME=$(terraform output -raw s3_bucket_name)

echo "=== TFDrift-Falco Drift Test ==="
echo "Security Group: ${SG_ID}"
echo "Instance:       ${INSTANCE_ID}"
echo "S3 Bucket:      ${BUCKET_NAME}"
echo ""

# --- Test 1: Security Group drift (add unauthorized ingress rule) ---
echo "[Test 1] Adding unauthorized ingress rule to security group..."
aws ec2 authorize-security-group-ingress \
  --group-id "${SG_ID}" \
  --protocol tcp \
  --port 8443 \
  --cidr "10.0.0.0/8" \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" \
  --output text 2>/dev/null && echo "  -> Added port 8443 ingress (DRIFT)" || echo "  -> Rule may already exist"

echo ""

# --- Test 2: EC2 instance attribute drift (enable monitoring) ---
echo "[Test 2] Enabling detailed monitoring on EC2 instance..."
aws ec2 monitor-instances \
  --instance-ids "${INSTANCE_ID}" \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" \
  --output text 2>/dev/null && echo "  -> Enabled monitoring (DRIFT: TF has monitoring=false)" || echo "  -> Failed"

echo ""

# --- Test 3: EC2 tag drift ---
echo "[Test 3] Adding unexpected tag to EC2 instance..."
aws ec2 create-tags \
  --resources "${INSTANCE_ID}" \
  --tags Key=ManualChange,Value=drift-test-$(date +%s) \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" \
  --output text 2>/dev/null && echo "  -> Added ManualChange tag (DRIFT)" || echo "  -> Failed"

echo ""

# --- Test 4: S3 bucket versioning drift ---
echo "[Test 4] Suspending S3 bucket versioning..."
aws s3api put-bucket-versioning \
  --bucket "${BUCKET_NAME}" \
  --versioning-configuration Status=Suspended \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" 2>/dev/null && echo "  -> Suspended versioning (DRIFT: TF has Enabled)" || echo "  -> Failed"

echo ""
echo "=== Drift Triggered ==="
echo ""
echo "Expected detections:"
echo "  1. SecurityGroup ${SG_ID}: unauthorized ingress on port 8443"
echo "  2. Instance ${INSTANCE_ID}: monitoring enabled (should be disabled)"
echo "  3. Instance ${INSTANCE_ID}: unexpected ManualChange tag"
echo "  4. S3 ${BUCKET_NAME}: versioning Suspended (should be Enabled)"
echo ""
echo "Wait 5-15 minutes for CloudTrail events to propagate, then check:"
echo "  kubectl logs -n tfdrift -l app=tfdrift-falco --tail=50"
echo "  kubectl logs -n falco -l app.kubernetes.io/name=falco --tail=50"
