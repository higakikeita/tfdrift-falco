#!/usr/bin/env bash
# One-command demo bring-up for the OSS Summit money-shot: builds tfdrift, starts
# the API backend (Falco http transport) + the React UI, then returns. Run it
# once before the talk (first build downloads deps and can take ~1min).
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/.." && pwd)"
API_PORT="${API_PORT:-8080}"
UI_PORT="${UI_PORT:-5173}"

# Free the ports before we start. teardown.sh only kills the PIDs it recorded, so a
# leftover process from an earlier run (rehearsal, crashed shell, second run.sh) keeps
# holding the port. Vite then fails with "Port 5173 is already in use" — but only in
# .ui.log, while this script still prints "Demo up" and the browser shows a white page.
# That is a silent stage failure, so reclaim the ports up front.
reclaim() {
  local port="$1" name="$2" pids
  pids="$(lsof -ti:"$port" 2>/dev/null || true)"
  if [ -n "$pids" ]; then
    echo "▶ port :$port still held ($name) — reclaiming pid(s): $pids"
    # shellcheck disable=SC2086
    kill -9 $pids 2>/dev/null || true
    sleep 1
  fi
}
reclaim "$API_PORT" backend
reclaim "$UI_PORT" ui
rm -f "$HERE/.backend.pid" "$HERE/.ui.pid"

sed "s#REPLACED_BY_RUN_SH#$HERE/state.tfstate#" "$HERE/config.yaml" > "$HERE/.config.run.yaml"

echo "▶ building tfdrift…"
( cd "$ROOT" && go build -o "$HERE/.tfdrift" ./cmd/tfdrift )

echo "▶ starting backend on :$API_PORT …"
nohup "$HERE/.tfdrift" --server --api-port "$API_PORT" --config "$HERE/.config.run.yaml" > "$HERE/.backend.log" 2>&1 &
echo $! > "$HERE/.backend.pid"; disown || true
for _ in 1 2 3 4 5 6 7 8; do curl -sf "http://127.0.0.1:$API_PORT/health" >/dev/null 2>&1 && break; sleep 1; done
curl -sf "http://127.0.0.1:$API_PORT/health" >/dev/null 2>&1 && echo "  backend: healthy" || { echo "  backend FAILED — see $HERE/.backend.log"; exit 1; }

echo "▶ starting UI (Vite) on :$UI_PORT …"
( cd "$ROOT/ui" && nohup npm run dev -- --port "$UI_PORT" --strictPort > "$HERE/.ui.log" 2>&1 & echo $! > "$HERE/.ui.pid"; disown || true )

# Wait for Vite to actually serve. Without this the script cheerfully reports success
# while the dev server is dead, which is exactly how you end up staring at a white
# page on stage.
UI_OK=0
for _ in 1 2 3 4 5 6 7 8 9 10 11 12; do
  if curl -sf "http://127.0.0.1:$UI_PORT/" >/dev/null 2>&1; then UI_OK=1; break; fi
  sleep 1
done
if [ "$UI_OK" -ne 1 ]; then
  echo "  UI FAILED — see $HERE/.ui.log"
  tail -20 "$HERE/.ui.log" || true
  exit 1
fi
echo "  UI: serving"

cat <<MSG

✅ Demo up. Open the UI and keep the tab visible:
     UI:   http://localhost:$UI_PORT
     API:  http://127.0.0.1:$API_PORT

   Act 2 (real-time, no AWS):  bash "$HERE/trigger-realtime.sh"
   Act 3 (scan, needs AWS):    bash "$HERE/scan.sh"
   Teardown:                   bash "$HERE/teardown.sh"
MSG
