#!/bin/bash
set -e

# TFDrift-Falco Quick Start Setup
# This script sets up everything needed to run TFDrift-Falco in 5 minutes

VERSION="0.5.0"
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BOLD}🚀 TFDrift-Falco Quick Start (v${VERSION})${NC}"
echo ""
echo "This script will set up TFDrift-Falco with Falco in Docker."
echo ""

# Check prerequisites
echo -e "${BOLD}📋 Checking prerequisites...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker: https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✅ Docker installed${NC}"

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi
echo -e "${GREEN}✅ Docker Compose installed${NC}"

# Check AWS credentials
if [ ! -d "${HOME}/.aws" ] || [ ! -f "${HOME}/.aws/credentials" ]; then
    echo -e "${YELLOW}⚠️  AWS credentials not found at ~/.aws/credentials${NC}"
    echo ""
    read -p "Do you want to configure AWS credentials now? (y/n): " configure_aws
    if [[ "$configure_aws" == "y" || "$configure_aws" == "Y" ]]; then
        echo ""
        echo -e "${BOLD}Enter your AWS credentials:${NC}"
        read -p "AWS Access Key ID: " aws_access_key_id
        read -sp "AWS Secret Access Key: " aws_secret_access_key
        echo ""
        read -p "AWS Region (default: us-east-1): " aws_region
        aws_region=${aws_region:-us-east-1}

        mkdir -p "${HOME}/.aws"
        cat > "${HOME}/.aws/credentials" <<EOF
[default]
aws_access_key_id = ${aws_access_key_id}
aws_secret_access_key = ${aws_secret_access_key}
EOF

        cat > "${HOME}/.aws/config" <<EOF
[default]
region = ${aws_region}
output = json
EOF

        echo -e "${GREEN}✅ AWS credentials configured${NC}"
    else
        echo -e "${RED}❌ AWS credentials are required to run TFDrift-Falco${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ AWS credentials found${NC}"
fi

echo ""
echo -e "${BOLD}📂 Setting up configuration files...${NC}"

# Create directories
mkdir -p deployments/falco
mkdir -p rules
mkdir -p examples/terraform

# Create Falco configuration
echo -e "${YELLOW}Creating Falco configuration...${NC}"
cat > deployments/falco/falco.yaml <<'EOF'
# Falco configuration for TFDrift-Falco
watch_config_files: true
time_format_iso_8601: true

# Rules
rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
  - /etc/falco/rules.d

# gRPC output (required for TFDrift-Falco)
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8

grpc_output:
  enabled: true

# JSON output for CloudTrail events
json_output: true
json_include_output_property: true
json_include_tags_property: true

# Logging
log_stderr: true
log_syslog: false
log_level: info

# CloudTrail plugin configuration
plugins:
  - name: cloudtrail
    library_path: libcloudtrail.so
    init_config:
      s3DownloadConcurrency: 64
      sqsDelete: false
      useAsync: true
    open_params: ""

# Load CloudTrail plugin rules
load_plugins: [cloudtrail]
EOF

echo -e "${GREEN}✅ Falco configuration created${NC}"

# Create TFDrift Falco rules
echo -e "${YELLOW}Creating TFDrift Falco rules...${NC}"
cat > rules/terraform_drift.yaml <<'EOF'
# TFDrift-Falco Rules
# These rules detect CloudTrail events that may indicate Terraform drift

- rule: Terraform Managed Resource Modified
  desc: Detect modifications to resources that should be managed by Terraform
  condition: >
    evt.type = aws_api_call and
    ct.name in (
      ModifyInstanceAttribute, ModifyDBInstance, ModifySecurityGroupRules,
      PutBucketPolicy, PutBucketEncryption, UpdateFunctionConfiguration,
      PutRolePolicy, UpdateAssumeRolePolicy, AttachRolePolicy,
      AuthorizeSecurityGroupIngress, RevokeSecurityGroupIngress,
      CreateStack, UpdateStack, DeleteStack
    )
  output: >
    Potential Terraform drift detected
    (user=%ct.user
     event=%ct.name
     resource=%ct.request.instanceid
     region=%ct.region
     source_ip=%ct.srcip
     aws_account=%ct.account)
  priority: WARNING
  tags: [terraform, drift, iac]
  source: aws_cloudtrail

- rule: Critical Infrastructure Change
  desc: Detect critical infrastructure changes that bypass IaC workflows
  condition: >
    evt.type = aws_api_call and
    ct.name in (
      TerminateInstances, DeleteDBInstance, DeleteBucket,
      DeleteSecurityGroup, ScheduleKeyDeletion, DeleteRole,
      DeleteStack, DeleteStateMachine
    )
  output: >
    Critical infrastructure deletion detected
    (user=%ct.user
     event=%ct.name
     resource=%ct.request.instanceid
     region=%ct.region
     aws_account=%ct.account)
  priority: CRITICAL
  tags: [terraform, drift, deletion, critical]
  source: aws_cloudtrail

- rule: IAM Permission Escalation
  desc: Detect potential IAM permission escalation
  condition: >
    evt.type = aws_api_call and
    ct.name in (
      AttachUserPolicy, AttachRolePolicy, PutUserPolicy, PutRolePolicy
    ) and
    (ct.request.policyarn contains "AdministratorAccess" or
     ct.request.policydocument contains "\"Effect\":\"Allow\"" and
     ct.request.policydocument contains "\"Action\":\"*\"")
  output: >
    Potential IAM privilege escalation detected
    (user=%ct.user
     event=%ct.name
     target_user=%ct.request.username
     target_role=%ct.request.rolename
     policy=%ct.request.policyarn
     region=%ct.region)
  priority: ALERT
  tags: [terraform, drift, iam, security, privilege-escalation]
  source: aws_cloudtrail
EOF

echo -e "${GREEN}✅ TFDrift rules created${NC}"

# Prompt for configuration
echo ""
echo -e "${BOLD}🔧 TFDrift-Falco Configuration${NC}"
echo ""

# AWS Region
read -p "AWS Region to monitor (default: us-east-1): " aws_region
aws_region=${aws_region:-us-east-1}

# Terraform State Backend
echo ""
echo "Terraform State Backend:"
echo "  1) S3 (recommended for production)"
echo "  2) Local file (for testing)"
read -p "Select backend (1-2, default: 2): " backend_choice
backend_choice=${backend_choice:-2}

if [ "$backend_choice" == "1" ]; then
    read -p "S3 bucket name: " s3_bucket
    read -p "S3 key (e.g., prod/terraform.tfstate): " s3_key

    state_config=$(cat <<EOF
    state:
      backend: "s3"
      s3_bucket: "${s3_bucket}"
      s3_key: "${s3_key}"
EOF
    )
else
    # Create example Terraform state
    mkdir -p examples/terraform
    echo '{"version": 4, "terraform_version": "1.0.0", "resources": []}' > examples/terraform/terraform.tfstate

    state_config=$(cat <<EOF
    state:
      backend: "local"
      local_path: "/terraform/terraform.tfstate"
EOF
    )
fi

# Slack Webhook (optional)
echo ""
read -p "Slack webhook URL (optional, press Enter to skip): " slack_webhook

if [ -n "$slack_webhook" ]; then
    slack_config=$(cat <<EOF
  slack:
    enabled: true
    webhook_url: "${slack_webhook}"
    channel: "#tfdrift-alerts"
EOF
    )
else
    slack_config=$(cat <<EOF
  slack:
    enabled: false
EOF
    )
fi

# Create TFDrift-Falco configuration
echo ""
echo -e "${YELLOW}Creating TFDrift-Falco configuration...${NC}"
cat > config.yaml <<EOF
# TFDrift-Falco Configuration (Auto-generated by quick-start.sh)

# Cloud Provider Configuration
providers:
  aws:
    enabled: true
    regions:
      - ${aws_region}
${state_config}

# Falco Integration
falco:
  enabled: true
  hostname: "falco"
  port: 5060

# Drift Detection Rules
drift_rules:
  - name: "EC2 Instance Modification"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
      - "disable_api_termination"
      - "security_groups"
    severity: "high"

  - name: "Security Group Changes"
    resource_types:
      - "aws_security_group"
      - "aws_security_group_rule"
    watched_attributes:
      - "ingress"
      - "egress"
      - "cidr_blocks"
    severity: "critical"

  - name: "IAM Policy Changes"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
      - "aws_iam_user"
    watched_attributes:
      - "policy"
      - "assume_role_policy"
      - "inline_policy"
    severity: "critical"

  - name: "S3 Bucket Security"
    resource_types:
      - "aws_s3_bucket"
      - "aws_s3_bucket_public_access_block"
    watched_attributes:
      - "block_public_acls"
      - "block_public_policy"
      - "ignore_public_acls"
      - "restrict_public_buckets"
    severity: "critical"

# Notification Channels
notifications:
${slack_config}

  falco_output:
    enabled: true
    priority: "warning"

# Logging
logging:
  level: "info"
  format: "json"
EOF

echo -e "${GREEN}✅ TFDrift-Falco configuration created${NC}"

# Create .env file
echo ""
echo -e "${YELLOW}Creating .env file...${NC}"
cat > .env <<EOF
# TFDrift-Falco Environment Variables
AWS_REGION=${aws_region}
TZ=UTC
EOF

echo -e "${GREEN}✅ .env file created${NC}"

# Summary
echo ""
echo -e "${BOLD}${GREEN}✅ Setup Complete!${NC}"
echo ""
echo -e "${BOLD}🚀 Start TFDrift-Falco:${NC}"
echo ""
echo -e "  ${BOLD}docker compose up -d${NC}"
echo ""
echo -e "${BOLD}📊 View logs:${NC}"
echo ""
echo "  docker compose logs -f tfdrift"
echo "  docker compose logs -f falco"
echo ""
echo -e "${BOLD}🛑 Stop services:${NC}"
echo ""
echo "  docker compose down"
echo ""
echo -e "${BOLD}📚 Next Steps:${NC}"
echo ""
echo "  1. Review configuration: vim config.yaml"
echo "  2. Customize drift rules for your infrastructure"
echo "  3. Set up Slack/webhook notifications"
echo "  4. Read full documentation: https://github.com/higakikeita/tfdrift-falco"
echo ""
echo -e "${YELLOW}⚡ Pro tip: Run 'make logs' or 'make restart' for common operations${NC}"
echo ""
