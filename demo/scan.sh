#!/usr/bin/env bash
# Act 3 — `tfdrift scan`: one-shot reconcile of the *applied* Terraform state
# (demo/tf) against live AWS. Needs a valid AWS session (see aws-session.sh).
# A rule added by hand in the console shows up as a MODIFIED aws_security_group.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
: "${AWS_PROFILE:=draios-dev-developer}"; export AWS_PROFILE
: "${AWS_REGION:=ap-northeast-1}"; export AWS_REGION

STATE="$HERE/tf/terraform.tfstate"
if [ ! -f "$STATE" ]; then
  echo "ERROR: $STATE not found. Run 'terraform -chdir=$HERE/tf apply' first." >&2
  exit 2
fi
sed "s#REPLACED_BY_SCAN_SH#$STATE#" "$HERE/scan-config.yaml" > "$HERE/.scan.run.yaml"
"$HERE/.tfdrift" scan --config "$HERE/.scan.run.yaml" --output human
