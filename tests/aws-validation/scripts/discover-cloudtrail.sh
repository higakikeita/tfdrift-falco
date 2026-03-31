#!/usr/bin/env bash
# Discover existing CloudTrail configuration in the AWS account
set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-draios-dev-developer}"
AWS_REGION="${AWS_REGION:-ap-northeast-1}"

echo "=== CloudTrail Discovery ==="
echo "Account: $(aws sts get-caller-identity --profile ${AWS_PROFILE} --query 'Account' --output text)"
echo ""

echo "--- Active Trails ---"
aws cloudtrail describe-trails \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" \
  --query 'trailList[*].{Name:Name, S3Bucket:S3BucketName, IsOrgTrail:IsOrganizationTrail, IsMultiRegion:IsMultiRegionTrail, HomeRegion:HomeRegion}' \
  --output table 2>/dev/null || echo "No trails found or access denied"

echo ""
echo "--- Trail Status ---"
for trail in $(aws cloudtrail describe-trails --profile "${AWS_PROFILE}" --region "${AWS_REGION}" --query 'trailList[*].TrailARN' --output text 2>/dev/null); do
  echo "Trail: ${trail}"
  aws cloudtrail get-trail-status \
    --name "${trail}" \
    --profile "${AWS_PROFILE}" \
    --region "${AWS_REGION}" \
    --query '{IsLogging:IsLogging, LatestDeliveryTime:LatestDeliveryTime}' \
    --output table 2>/dev/null || echo "  (access denied)"
  echo ""
done

echo "--- SQS Queues (CloudTrail-related) ---"
aws sqs list-queues \
  --profile "${AWS_PROFILE}" \
  --region "${AWS_REGION}" \
  --queue-name-prefix "cloudtrail" \
  --output text 2>/dev/null || echo "No CloudTrail SQS queues found"

echo ""
echo "Done. Use the S3 bucket name in phase2-eks/variables.tf (cloudtrail_bucket)."
