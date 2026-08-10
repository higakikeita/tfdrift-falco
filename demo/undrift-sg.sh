#!/usr/bin/env bash
# Cleanup: revoke the demo ingress rule.
set -euo pipefail
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION
SG="${DEMO_SG:-sg-056b65b187cd8ece3}"
PORT="${DEMO_PORT:-8443}"
CIDR="${DEMO_CIDR:-203.0.113.7/32}"
echo "▶ Revoking demo ingress rule tcp/$PORT from $CIDR on $SG"
aws ec2 revoke-security-group-ingress --group-id "$SG" \
  --ip-permissions "IpProtocol=tcp,FromPort=$PORT,ToPort=$PORT,IpRanges=[{CidrIp=$CIDR}]" 2>&1 | tail -1 || true
echo "  done."
