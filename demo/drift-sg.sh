#!/usr/bin/env bash
# Demo drift: add a Terraform-unmanaged ingress rule to the lab SG so `tfdrift
# scan` reports it. Idempotent-ish (ignores "already exists").
set -euo pipefail
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION
SG="${DEMO_SG:-sg-056b65b187cd8ece3}"
PORT="${DEMO_PORT:-8443}"
CIDR="${DEMO_CIDR:-203.0.113.7/32}"
echo "▶ Adding out-of-band ingress rule to $SG: tcp/$PORT from $CIDR"
aws ec2 authorize-security-group-ingress --group-id "$SG" \
  --ip-permissions "IpProtocol=tcp,FromPort=$PORT,ToPort=$PORT,IpRanges=[{CidrIp=$CIDR,Description=demo-drift}]" \
  --query 'SecurityGroupRules[0].SecurityGroupRuleId' --output text 2>&1 | tail -1 || true
echo "  → now run: bash demo/scan.sh   (the new rule shows as ingress drift)"
