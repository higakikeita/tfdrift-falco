#!/bin/bash
#
# CloudTrail Setup Script for TFDrift-Falco
# This script creates CloudTrail and S3 bucket for drift detection
#

set -e

# Configuration
REGION="${AWS_REGION:-us-east-1}"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
TRAIL_NAME="tfdrift-falco-trail"
BUCKET_NAME="tfdrift-cloudtrail-${ACCOUNT_ID}-${REGION}"

echo "=========================================="
echo "CloudTrail Setup for TFDrift-Falco"
echo "=========================================="
echo "Region: $REGION"
echo "Account ID: $ACCOUNT_ID"
echo "Trail Name: $TRAIL_NAME"
echo "S3 Bucket: $BUCKET_NAME"
echo "=========================================="
echo ""

# Step 1: Create S3 bucket for CloudTrail logs
echo "Step 1: Creating S3 bucket..."
if aws s3 ls "s3://${BUCKET_NAME}" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3api create-bucket \
        --bucket "${BUCKET_NAME}" \
        --region "${REGION}" \
        $(if [ "$REGION" != "us-east-1" ]; then echo "--create-bucket-configuration LocationConstraint=${REGION}"; fi)
    echo "✓ S3 bucket created: ${BUCKET_NAME}"
else
    echo "✓ S3 bucket already exists: ${BUCKET_NAME}"
fi

# Step 2: Apply bucket policy for CloudTrail
echo ""
echo "Step 2: Applying bucket policy..."
cat > /tmp/cloudtrail-bucket-policy.json <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailAclCheck",
            "Effect": "Allow",
            "Principal": {
                "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "arn:aws:s3:::${BUCKET_NAME}"
        },
        {
            "Sid": "AWSCloudTrailWrite",
            "Effect": "Allow",
            "Principal": {
                "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::${BUCKET_NAME}/AWSLogs/${ACCOUNT_ID}/*",
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
    ]
}
EOF

aws s3api put-bucket-policy \
    --bucket "${BUCKET_NAME}" \
    --policy file:///tmp/cloudtrail-bucket-policy.json

echo "✓ Bucket policy applied"

# Step 3: Create CloudTrail
echo ""
echo "Step 3: Creating CloudTrail..."
if aws cloudtrail describe-trails --region "${REGION}" --query "trailList[?Name=='${TRAIL_NAME}']" --output text | grep -q "${TRAIL_NAME}"; then
    echo "✓ CloudTrail already exists: ${TRAIL_NAME}"
else
    aws cloudtrail create-trail \
        --name "${TRAIL_NAME}" \
        --s3-bucket-name "${BUCKET_NAME}" \
        --is-multi-region-trail \
        --region "${REGION}"
    echo "✓ CloudTrail created: ${TRAIL_NAME}"
fi

# Step 4: Start logging
echo ""
echo "Step 4: Starting CloudTrail logging..."
aws cloudtrail start-logging \
    --name "${TRAIL_NAME}" \
    --region "${REGION}"
echo "✓ CloudTrail logging started"

# Step 5: Verify CloudTrail status
echo ""
echo "Step 5: Verifying CloudTrail status..."
aws cloudtrail get-trail-status \
    --name "${TRAIL_NAME}" \
    --region "${REGION}"

echo ""
echo "=========================================="
echo "✓ CloudTrail setup completed!"
echo "=========================================="
echo ""
echo "Configuration for Falco CloudTrail plugin:"
echo "  S3 Bucket: ${BUCKET_NAME}"
echo "  Region: ${REGION}"
echo ""
echo "Add this to your Falco configuration:"
echo ""
echo "plugins:"
echo "  - name: cloudtrail"
echo "    library_path: libcloudtrail.so"
echo "    init_config: \"\""
echo "    open_params: \"s3Bucket=${BUCKET_NAME}\""
echo ""
echo "Note: It may take 5-15 minutes for the first CloudTrail logs to appear."
echo ""
