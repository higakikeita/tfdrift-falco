#!/usr/bin/env bash
# Get (or refresh) the AWS session that Act 3 needs, then prove it worked.
#
# Run this at ~16:20 on the day — NOT in the morning. okta-aws-cli issues a
# 1-hour session, and Act 3 runs about 18 minutes into the talk (~16:53), so a
# session taken at breakfast is already dead. Taken at 16:20 it expires 17:20,
# which leaves a comfortable margin.
set -euo pipefail

PROFILE="${AWS_PROFILE:-draios-dev-developer}"
ORG="${OKTA_ORG:-sysdig.okta.com}"
CLIENT_ID="${OKTA_CLIENT_ID:-0oaqz1xwwzHyCiV7A357}"
FED_APP_ID="${OKTA_FED_APP_ID:-0oaltf406vJIPOhrY356}"
IDP_ARN="${AWS_IAM_IDP:-arn:aws:iam::230446364776:saml-provider/OKTA}"
ROLE_ARN="${AWS_IAM_ROLE:-arn:aws:iam::230446364776:role/ADMIN_OKTA_TEST}"
DURATION="${SESSION_DURATION:-3600}"

echo "▶ requesting an AWS session (profile: $PROFILE, ${DURATION}s)…"
okta-aws-cli web \
  --org-domain "$ORG" \
  --oidc-client-id "$CLIENT_ID" \
  --aws-acct-fed-app-id "$FED_APP_ID" \
  --profile "$PROFILE" \
  --aws-iam-idp  "$IDP_ARN" \
  --aws-iam-role "$ROLE_ARN" \
  --aws-session-duration "$DURATION" \
  --write-aws-credentials \
  --open-browser

echo
echo "▶ verifying…"
if ARN=$(aws sts get-caller-identity --profile "$PROFILE" --query Arn --output text 2>&1); then
  echo "  ✅ $ARN"
  # The session name at the tail of an assumed-role ARN is the human. That is the
  # same field TFDrift pulls out of CloudTrail for the actor — worth a glance,
  # because it is exactly what Act 2 claims to be able to do.
  echo "  actor (session name): ${ARN##*/}"
  EXPIRES=$(date -v+"${DURATION}"S '+%H:%M' 2>/dev/null || date -d "+${DURATION} seconds" '+%H:%M')
  echo "  expires ~$EXPIRES  — Act 3 must run before then"
else
  echo "  ❌ session not usable:"
  echo "     $ARN"
  echo "     Act 3 (scan.sh) will fail. Use fallback-act3.mp4 and move on."
  exit 1
fi
