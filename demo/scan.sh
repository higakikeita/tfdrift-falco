#!/usr/bin/env bash
# Act 3 — `driftwire scan`: one-shot Terraform-state-vs-live-cloud reconcile.
# Needs a valid AWS session (okta-aws-cli). Uses the lab bastion SG fixture so a
# real security-group rule drift shows up.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION
sed "s#REPLACED_BY_SCAN_SH#$HERE/scan-state.tfstate#" "$HERE/scan-config.yaml" > "$HERE/.scan.run.yaml"
"$HERE/.driftwire" scan --config "$HERE/.scan.run.yaml" --output human
