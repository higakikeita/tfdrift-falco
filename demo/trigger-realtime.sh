#!/usr/bin/env bash
# Act 2 — real-time detection.
# Simulates an out-of-band change: someone modifies a *Terraform-managed* prod
# EC2 instance's type via the AWS Console. This POSTs the EXACT JSON that Falco
# 0.43 emits over http_output for that CloudTrail event (ADR-006), so the demo
# exercises the real receiver → parser → detector → UI path — no AWS needed.
set -euo pipefail
API="${TFDRIFT_API:-http://127.0.0.1:8080}"

echo "▶ Scenario: alice@corp modifies prod EC2 instance-type via the AWS Console (outside Terraform)"
curl -s -o /dev/null -w "  → tfdrift receiver: HTTP %{http_code}\n" \
  -X POST "${API}/api/v1/falco/events" \
  -H 'Content-Type: application/json' \
  -d '{
    "priority":"Warning",
    "rule":"Terraform Managed Resource Modified",
    "source":"aws_cloudtrail",
    "time":"2026-08-10T09:30:00.000000000Z",
    "output_fields":{
      "ct.name":"ModifyInstanceAttribute",
      "ct.region":"ap-northeast-1",
      "ct.request":"{\"instanceId\":\"i-0demoweb0000001\",\"instanceType\":{\"value\":\"t3.2xlarge\"}}",
      "ct.user":"alice@corp.example",
      "ct.user.arn":"arn:aws:sts::230446364776:assumed-role/Admin/alice@corp.example",
      "ct.user.accountid":"230446364776"
    }
  }'
echo "  → Now watch the Dashboard: a drift card appears within ~5s (who=alice, what=ModifyInstanceAttribute)."
