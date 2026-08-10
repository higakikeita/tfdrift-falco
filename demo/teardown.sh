#!/usr/bin/env bash
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_PORT="${API_PORT:-8080}"
UI_PORT="${UI_PORT:-5173}"
for p in .backend.pid .ui.pid; do [ -f "$HERE/$p" ] && kill "$(cat "$HERE/$p")" 2>/dev/null && rm -f "$HERE/$p"; done
# Also sweep the ports: npm spawns a child Vite process, so killing the recorded npm
# PID can leave the real listener behind — which is what silently breaks the next run.
for port in "$API_PORT" "$UI_PORT"; do
  pids="$(lsof -ti:"$port" 2>/dev/null || true)"
  # shellcheck disable=SC2086
  [ -n "$pids" ] && kill -9 $pids 2>/dev/null || true
done
rm -f "$HERE/.config.run.yaml" "$HERE/.scan.run.yaml"
echo "demo stopped. ports :$API_PORT and :$UI_PORT are free."
