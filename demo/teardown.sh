#!/usr/bin/env bash
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
for p in .backend.pid .ui.pid; do [ -f "$HERE/$p" ] && kill "$(cat "$HERE/$p")" 2>/dev/null && rm -f "$HERE/$p"; done
rm -f "$HERE/.config.run.yaml" "$HERE/.scan.run.yaml"
echo "demo stopped."
