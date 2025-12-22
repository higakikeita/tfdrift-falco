#!/bin/bash
# TFDrift-Falco Test Environment Cleanup Script
set -e

echo "🧹 TFDrift-Falco Test Environment Cleanup"
echo "=========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Warning
echo ""
echo -e "${YELLOW}⚠️  WARNING: This will destroy ALL resources created by this test environment!${NC}"
echo ""
echo "Resources to be destroyed:"
echo "  - S3 Test Bucket"
echo "  - Security Group"
echo "  - IAM Policy and Role"
echo "  - CloudWatch Log Group"
echo "  - SNS Topic"
echo ""

read -p "Are you sure you want to continue? (type 'yes' to confirm): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Cleanup cancelled."
    exit 0
fi

# Load configuration if exists
if [ -f "../config-test-env.sh" ]; then
    echo ""
    echo "📋 Loading configuration..."
    source ../config-test-env.sh
fi

# Step 1: Destroy Terraform resources
echo ""
echo "🔥 Step 1: Destroying Terraform resources..."
terraform destroy -auto-approve

echo -e "${GREEN}✅ Terraform resources destroyed${NC}"

# Step 2: Remove state bucket (optional)
echo ""
read -p "Do you want to also delete the state bucket? (yes/no): " DELETE_STATE

if [ "$DELETE_STATE" = "yes" ]; then
    if [ -n "$STATE_BUCKET" ]; then
        echo "Deleting state bucket: $STATE_BUCKET"

        # Empty bucket first
        aws s3 rm s3://$STATE_BUCKET --recursive || true

        # Delete bucket
        aws s3 rb s3://$STATE_BUCKET --force || true

        echo -e "${GREEN}✅ State bucket deleted${NC}"
    else
        echo -e "${YELLOW}⚠️  STATE_BUCKET not set, skipping${NC}"
    fi
else
    echo "State bucket preserved. Delete manually if needed:"
    echo "  aws s3 rb s3://\$STATE_BUCKET --force"
fi

# Step 3: Clean up local files
echo ""
echo "🗑️  Step 3: Cleaning up local files..."

# Remove Terraform files
rm -f terraform.tfvars
rm -f tfplan
rm -rf .terraform/
rm -f .terraform.lock.hcl
rm -f terraform.tfstate*

# Remove generated config
rm -f ../config-test-env.sh

echo -e "${GREEN}✅ Local files cleaned up${NC}"

# Step 4: Verify cleanup
echo ""
echo "🔍 Step 4: Verifying cleanup..."

# Check if any resources remain
echo "Checking for remaining resources..."

# Check S3 buckets
if [ -n "$TEST_BUCKET" ]; then
    if aws s3 ls "s3://$TEST_BUCKET" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  Test bucket still exists: $TEST_BUCKET${NC}"
        echo "   Delete manually: aws s3 rb s3://$TEST_BUCKET --force"
    else
        echo -e "${GREEN}✅ Test bucket removed${NC}"
    fi
fi

# Check Security Group
if [ -n "$SECURITY_GROUP_ID" ]; then
    if aws ec2 describe-security-groups --group-ids "$SECURITY_GROUP_ID" 2>/dev/null >/dev/null; then
        echo -e "${YELLOW}⚠️  Security group still exists: $SECURITY_GROUP_ID${NC}"
        echo "   Delete manually: aws ec2 delete-security-group --group-id $SECURITY_GROUP_ID"
    else
        echo -e "${GREEN}✅ Security group removed${NC}"
    fi
fi

echo ""
echo -e "${GREEN}✅ Cleanup complete!${NC}"
echo ""
echo "Summary:"
echo "  - Terraform resources: Destroyed"
echo "  - State bucket: $([ "$DELETE_STATE" = "yes" ] && echo "Deleted" || echo "Preserved")"
echo "  - Local files: Cleaned up"
echo ""
echo "To run the test again, execute: ./setup.sh"
