#!/bin/bash
# TFDrift-Falco Test Environment Setup Script
set -e

echo "🚀 TFDrift-Falco Test Environment Setup"
echo "========================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo ""
echo "📋 Checking prerequisites..."

if ! command -v aws &> /dev/null; then
    echo -e "${RED}❌ AWS CLI is not installed${NC}"
    echo "Install: https://aws.amazon.com/cli/"
    exit 1
fi
echo -e "${GREEN}✅ AWS CLI found${NC}"

if ! command -v terraform &> /dev/null; then
    echo -e "${RED}❌ Terraform is not installed${NC}"
    echo "Install: https://www.terraform.io/downloads"
    exit 1
fi
echo -e "${GREEN}✅ Terraform found${NC}"

if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq is not installed (optional but recommended)${NC}"
    echo "Install: brew install jq"
fi

# Check AWS credentials
echo ""
echo "🔐 Checking AWS credentials..."
if ! aws sts get-caller-identity &> /dev/null; then
    echo -e "${RED}❌ AWS credentials are not configured${NC}"
    echo "Run: aws configure"
    exit 1
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_USER=$(aws sts get-caller-identity --query Arn --output text)
echo -e "${GREEN}✅ Authenticated as: $AWS_USER${NC}"
echo -e "${GREEN}   Account ID: $ACCOUNT_ID${NC}"

# Set AWS region
read -p "Enter AWS region [us-east-1]: " AWS_REGION
AWS_REGION=${AWS_REGION:-us-east-1}
export AWS_DEFAULT_REGION=$AWS_REGION

# Step 1: Create S3 backend bucket
echo ""
echo "📦 Step 1: Creating S3 backend bucket..."
STATE_BUCKET="tfdrift-test-state-$(date +%Y%m%d)-${ACCOUNT_ID:0:8}"
echo "Bucket name: $STATE_BUCKET"

if aws s3 ls "s3://$STATE_BUCKET" 2>/dev/null; then
    echo -e "${YELLOW}⚠️  Bucket already exists, skipping creation${NC}"
else
    aws s3api create-bucket \
        --bucket $STATE_BUCKET \
        --region $AWS_REGION \
        $([ "$AWS_REGION" != "us-east-1" ] && echo "--create-bucket-configuration LocationConstraint=$AWS_REGION")

    echo "Enabling versioning..."
    aws s3api put-bucket-versioning \
        --bucket $STATE_BUCKET \
        --versioning-configuration Status=Enabled

    echo "Enabling encryption..."
    aws s3api put-bucket-encryption \
        --bucket $STATE_BUCKET \
        --server-side-encryption-configuration '{
            "Rules": [{
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                }
            }]
        }'

    echo -e "${GREEN}✅ S3 backend bucket created${NC}"
fi

# Step 2: Update backend.tf
echo ""
echo "⚙️  Step 2: Updating backend.tf..."
cat > backend.tf <<EOF
# Terraform Backend Configuration
# S3 backend for storing state - TFDrift will monitor this state

terraform {
  backend "s3" {
    bucket = "$STATE_BUCKET"
    key    = "test-environment/terraform.tfstate"
    region = "$AWS_REGION"
  }
}
EOF
echo -e "${GREEN}✅ backend.tf updated${NC}"

# Step 3: Get default VPC
echo ""
echo "🌐 Step 3: Finding default VPC..."
DEFAULT_VPC=$(aws ec2 describe-vpcs \
    --filters "Name=isDefault,Values=true" \
    --query "Vpcs[0].VpcId" \
    --output text \
    --region $AWS_REGION)

if [ "$DEFAULT_VPC" = "None" ] || [ -z "$DEFAULT_VPC" ]; then
    echo -e "${YELLOW}⚠️  No default VPC found. Creating one...${NC}"
    aws ec2 create-default-vpc --region $AWS_REGION
    DEFAULT_VPC=$(aws ec2 describe-vpcs \
        --filters "Name=isDefault,Values=true" \
        --query "Vpcs[0].VpcId" \
        --output text \
        --region $AWS_REGION)
fi
echo -e "${GREEN}✅ Default VPC: $DEFAULT_VPC${NC}"

# Step 4: Create terraform.tfvars
echo ""
echo "📝 Step 4: Creating terraform.tfvars..."

# Generate unique bucket name
TEST_BUCKET_NAME="tfdrift-test-$(date +%Y%m%d)-$(openssl rand -hex 4)"

cat > terraform.tfvars <<EOF
# TFDrift Test Environment Configuration

aws_region       = "$AWS_REGION"
environment      = "test"
test_bucket_name = "$TEST_BUCKET_NAME"
vpc_id           = "$DEFAULT_VPC"

# Optional: Email for drift alerts
alert_email      = ""
EOF

echo -e "${GREEN}✅ terraform.tfvars created${NC}"
cat terraform.tfvars

# Step 5: Terraform init
echo ""
echo "🔧 Step 5: Initializing Terraform..."
terraform init

echo -e "${GREEN}✅ Terraform initialized${NC}"

# Step 6: Review plan
echo ""
echo "📊 Step 6: Reviewing Terraform plan..."
terraform plan -out=tfplan

# Step 7: Apply
echo ""
read -p "Do you want to apply this plan? (yes/no): " APPLY
if [ "$APPLY" = "yes" ]; then
    echo ""
    echo "🚀 Applying Terraform configuration..."
    terraform apply tfplan
    rm tfplan

    echo ""
    echo -e "${GREEN}✅ Resources created successfully!${NC}"
    echo ""
    echo "📋 Resource Summary:"
    terraform output -json | jq -r '.resources_summary.value | to_entries[] | "  - \(.key): \(.value)"'

    # Save important values
    echo ""
    echo "💾 Saving configuration..."
    cat > ../config-test-env.sh <<EOF
#!/bin/bash
# Auto-generated configuration for TFDrift test environment

export AWS_REGION="$AWS_REGION"
export STATE_BUCKET="$STATE_BUCKET"
export STATE_KEY="test-environment/terraform.tfstate"
export TEST_BUCKET="$TEST_BUCKET_NAME"
export VPC_ID="$DEFAULT_VPC"

# Resource IDs
export S3_BUCKET=\$(cd terraform/test-environment && terraform output -raw s3_bucket_name)
export SECURITY_GROUP_ID=\$(cd terraform/test-environment && terraform output -raw security_group_id)
export IAM_POLICY_ARN=\$(cd terraform/test-environment && terraform output -raw iam_policy_arn)
EOF
    chmod +x ../config-test-env.sh

    echo -e "${GREEN}✅ Configuration saved to ../config-test-env.sh${NC}"
    echo ""
    echo "🎉 Setup complete!"
    echo ""
    echo "Next steps:"
    echo "1. Source the config: source ../config-test-env.sh"
    echo "2. Update ../../config.yaml with the state location"
    echo "3. Run TFDrift: cd ../.. && ./tfdrift --config config.yaml"
    echo "4. Try drift scenarios in README.md"
else
    echo "Skipping apply. Run 'terraform apply' manually when ready."
fi
