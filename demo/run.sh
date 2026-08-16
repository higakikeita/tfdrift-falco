#!/usr/bin/env bash
# One-command demo bring-up for the OSS Summit money-shot: builds driftwire, starts
# the API backend (Falco http transport) + the React UI, then returns. Run it
# once before the talk (first build downloads deps and can take ~1min).
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
API_PORT="${API_PORT:-8080}"

sed "s#REPLACED_BY_RUN_SH#$HERE/state.tfstate#" "$HERE/config.yaml" > "$HERE/.config.run.yaml"

echo "▶ building driftwire…"
( cd "$ROOT" && go build -o "$HERE/.driftwire" ./cmd/driftwire )

echo "▶ starting backend on :$API_PORT …"
nohup "$HERE/.driftwire" --server --api-port "$API_PORT" --config "$HERE/.config.run.yaml" > "$HERE/.backend.log" 2>&1 &
echo $! > "$HERE/.backend.pid"; disown || true
for _ in 1 2 3 4 5 6 7 8; do curl -sf "http://127.0.0.1:$API_PORT/health" >/dev/null 2>&1 && break; sleep 1; done
curl -sf "http://127.0.0.1:$API_PORT/health" >/dev/null 2>&1 && echo "  backend: healthy" || { echo "  backend FAILED — see $HERE/.backend.log"; exit 1; }

echo "▶ starting UI (Vite) on :5173 …"
( cd "$ROOT/ui" && nohup npm run dev -- --port 5173 --strictPort > "$HERE/.ui.log" 2>&1 & echo $! > "$HERE/.ui.pid"; disown || true )

cat <<MSG

✅ Demo up. Open the UI and keep the tab visible:
     UI:   http://localhost:5173
     API:  http://127.0.0.1:$API_PORT

   Act 2 (real-time, no AWS):  bash "$HERE/trigger-realtime.sh"
   Act 3 (scan, needs AWS):    bash "$HERE/scan.sh"
   Teardown:                   bash "$HERE/teardown.sh"
MSG
