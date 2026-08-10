#!/usr/bin/env bash
# Demo drift: add a Terraform-unmanaged ingress rule to the *demo* SG (the one
# demo/tf manages) so `tfdrift scan` reports it as MODIFIED. On stage this is
# done in the AWS console; this script is the CLI equivalent / rehearsal aid.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION

# Default SG = whatever demo/tf applied. Override with DEMO_SG=sg-….
SG="${DEMO_SG:-$(terraform -chdir="$HERE/tf" output -raw security_group_id 2>/dev/null || true)}"
PORT="${DEMO_PORT:-22}"
CIDR="${DEMO_CIDR:-0.0.0.0/0}"
if [ -z "$SG" ]; then
  echo "ERROR: no SG id. Set DEMO_SG or run terraform apply in demo/tf." >&2
  exit 2
fi
echo "▶ Adding out-of-band ingress to $SG: tcp/$PORT from $CIDR"
aws ec2 authorize-security-group-ingress --group-id "$SG" \
  --ip-permissions "IpProtocol=tcp,FromPort=$PORT,ToPort=$PORT,IpRanges=[{CidrIp=$CIDR,Description=demo-drift}]" \
  --query 'SecurityGroupRules[0].SecurityGroupRuleId' --output text 2>&1 | tail -1 || true
echo "  → now run: bash demo/scan.sh   (the new rule appears as MODIFIED ingress)"
