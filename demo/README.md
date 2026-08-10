# TFDrift-Falco — live demo runner (OSS Summit money-shot)

A turnkey, low-risk demo of the core value: **a change made outside Terraform is
caught by TFDrift in real time, with who / what / when** — plus a one-shot
state-vs-cloud reconcile (`tfdrift scan`).

## Why it's built this way

The real-time act is **driven by the exact JSON Falco 0.43 emits over
`http_output`** (ADR-006) rather than a live cloud change, so it is deterministic
and needs no AWS session, no Falco container, and no S3 delivery latency — the
things most likely to fail on stage. It still exercises the real path:
receiver → parser → detector → broadcaster → UI. The `scan` act is genuinely live
against AWS (a fresh `okta-aws-cli` session).

## Prerequisites

- Go + Node/npm (repo already builds).
- For Act 3 only: a valid AWS session — `okta-aws-cli web … --profile draios-dev-developer`.

## Run it

```bash
bash demo/run.sh                 # build + start backend (:8080) + UI (:5173)
# open http://localhost:5173  (keep the tab visible)
```

### Act 2 — real-time detection (no AWS)
```bash
bash demo/trigger-realtime.sh
```
Within ~5s a drift card appears on the Dashboard:
`aws_instance i-0demoweb0000001 — ModifyInstanceAttribute — alice@corp.example — medium`.
(The feed auto-refreshes even if the tab isn't focused.)

### Act 3 — live drift on real infra (live AWS)
```bash
bash demo/scan.sh          # baseline
bash demo/drift-sg.sh      # add a Terraform-unmanaged ingress rule to the lab SG
bash demo/scan.sh          # the new rule shows immediately as ingress drift
bash demo/undrift-sg.sh    # cleanup
```
`scan` queries the live security-group state directly, so an added rule is
reported instantly — no CloudTrail delivery latency. (Real-time streaming to the
UI over live CloudTrail lags minutes, so the live act uses `scan`, not the UI
stream; Act 2 covers the real-time visual.)

### Teardown
```bash
bash demo/undrift-sg.sh   # revoke the demo SG rule
bash demo/teardown.sh    # stop backend + UI
```

## Suggested 3-act narration (~5 min)

1. **The problem** (30s): someone changes a prod resource in the Console — a
   `terraform plan` won't notice until the next run.
2. **Real-time** (90s): run `trigger-realtime.sh`, switch to the UI, the card is
   already there — point at *who / what / when*.
3. **Reconcile** (60s): run `scan.sh` — the full state-vs-cloud diff, including
   security-group rule drift.

## Backup (record before the talk)

Record a screencast of the full **real cloud → Falco → TFDrift → UI** flow while
your AWS session is fresh, and keep it ready in case anything fails live.

## Honest framing

Continuous delivery uses Falco `http_output` → the TFDrift receiver. Production
continuous ingestion (periodic CloudTrail pickup) is tracked in #360; the demo
path and the AWS-free CI E2E (#365) both prove the pipeline.
