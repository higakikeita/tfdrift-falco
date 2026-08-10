#!/usr/bin/env bash
# Act 2 — real-time detection, narrated for a live audience.
#
# Simulates an out-of-band change: someone modifies a *Terraform-managed* prod
# EC2 instance's type via the AWS Console. This POSTs the EXACT JSON that Falco
# 0.43 emits over http_output for that CloudTrail event (ADR-006), so the demo
# exercises the real receiver → parser → detector → UI path — no AWS needed.
#
# The output deliberately walks the same four hops the slides just showed, one
# step at a time. A bare "HTTP 200" tells an audience nothing; they cannot tell a
# real pipeline from a magic trick. Set PACE=0 to skip the pauses when testing.
set -euo pipefail
API="${TFDRIFT_API:-http://127.0.0.1:8080}"
PACE="${PACE:-1.4}"
INSTANCE="i-0demoweb0000001"
ACTOR="alice@corp.example"

b() { printf '\033[1m%s\033[0m\n' "$1"; }        # bold
d() { printf '\033[2m%s\033[0m\n' "$1"; }        # dim
g() { printf '\033[1;32m%s\033[0m\n' "$1"; }     # bold green
hr() { d "────────────────────────────────────────────────────────────"; }
beat() { sleep "$PACE"; }

PAYLOAD=$(cat <<JSON
{
  "priority":"Warning",
  "rule":"Terraform Managed Resource Modified",
  "source":"aws_cloudtrail",
  "time":"$(date -u +%Y-%m-%dT%H:%M:%S).000000000Z",
  "output_fields":{
    "ct.name":"ModifyInstanceAttribute",
    "ct.region":"ap-northeast-1",
    "ct.request":"{\"instanceId\":\"$INSTANCE\",\"instanceType\":{\"value\":\"t3.2xlarge\"}}",
    "ct.user":"$ACTOR",
    "ct.user.arn":"arn:aws:sts::230446364776:assumed-role/Admin/$ACTOR",
    "ct.user.accountid":"230446364776"
  }
}
JSON
)

clear
echo
hr
b " HOP 1 — a human changes a Terraform-managed resource"
echo "         $ACTOR  ·  AWS Console  ·  outside Terraform"
hr
beat

echo
b " HOP 2 — CloudTrail records it, Falco matches its rule and emits:"
echo
echo "         rule        Terraform Managed Resource Modified"
echo "         source      aws_cloudtrail"
echo "         ct.name     ModifyInstanceAttribute"
echo "         ct.user     $ACTOR"
echo "         ct.request  {\"instanceId\":\"$INSTANCE\", \"instanceType\":\"t3.2xlarge\"}"
echo
d "         (this is byte-for-byte what Falco sends over http_output — ADR-006)"
beat

echo
b " HOP 3 — Falco POSTs it to TFDrift"
echo
CODE=$(curl -s -o /dev/null -w '%{http_code}' \
  -X POST "${API}/api/v1/falco/events" \
  -H 'Content-Type: application/json' \
  -d "$PAYLOAD")
echo "         POST /api/v1/falco/events   ->   HTTP $CODE"
if [ "$CODE" != "200" ] && [ "$CODE" != "202" ]; then
  echo
  echo "         receiver did not accept it. Is the backend up? (bash demo/run.sh)"
  exit 1
fi
beat

echo
b " HOP 4 — TFDrift parsed it, compared it against Terraform state, and filed:"
echo
sleep 1
FOUND=$(curl -s "${API}/api/v1/drifts" | python3 -c '
import json,sys
try: d=json.load(sys.stdin)
except Exception: print(""); sys.exit()
rows=d.get("drifts", d) if isinstance(d,dict) else d
if not rows: print(""); sys.exit()
r=rows[0] if isinstance(rows,list) else {}
t=r.get("resource_type") or r.get("type") or "aws_instance"
i=r.get("resource_id") or r.get("id") or ""
u=r.get("user") or r.get("actor") or ""
s=r.get("severity") or ""
print(f"{t}  {i}  by {u}  [{s}]".strip())
' 2>/dev/null || true)

if [ -n "$FOUND" ]; then
  g "         DRIFT   $FOUND"
else
  d "         (the API has not indexed it yet — the Dashboard will show it)"
fi

echo
hr
b " -> Now the Dashboard.  http://localhost:5173"
hr
echo
