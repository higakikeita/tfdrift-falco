#!/usr/bin/env bash
# Cleanup: revoke the demo ingress rule from the demo SG.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION
SG="${DEMO_SG:-$(terraform -chdir="$HERE/tf" output -raw security_group_id 2>/dev/null || true)}"
PORT="${DEMO_PORT:-22}"
CIDR="${DEMO_CIDR:-0.0.0.0/0}"
[ -z "$SG" ] && { echo "ERROR: no SG id (DEMO_SG or terraform output)"; exit 2; }
echo "▶ Revoking demo ingress tcp/$PORT from $CIDR on $SG"
aws ec2 revoke-security-group-ingress --group-id "$SG" \
  --ip-permissions "IpProtocol=tcp,FromPort=$PORT,ToPort=$PORT,IpRanges=[{CidrIp=$CIDR}]" 2>&1 | tail -1 || true
echo "  done."
