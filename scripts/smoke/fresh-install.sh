#!/usr/bin/env bash
# Fresh SQLite install: empty data dir → migration → admin seed → health + login + /auth/me menus.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

ensure_dirs
DATA_DIR="$SMOKE_TMP/fresh-data"
CFG="$SMOKE_TMP/fresh-config.yaml"
LOG="$SMOKE_TMP/fresh-server.log"
PORT="${SMOKE_PORT:-18080}"
BASE="http://127.0.0.1:${PORT}"

rm -rf "$DATA_DIR"
mkdir -p "$DATA_DIR"
write_smoke_config "$CFG" "$DATA_DIR" sqlite "$PORT"

BIN="$(build_server_bin "$SMOKE_TMP/bedrock-fresh")"
echo "==> starting server (fresh SQLite) on :$PORT"
"$BIN" --config "$CFG" >"$LOG" 2>&1 &
PID=$!
cleanup() {
  kill "$PID" 2>/dev/null || true
  wait "$PID" 2>/dev/null || true
}
trap cleanup EXIT

wait_http "$BASE/api/v1/health" 80
HEALTH="$(curl -fsS "$BASE/api/v1/health")"
echo "health: $HEALTH"
STATUS="$(json_get "$HEALTH" "o.get('data',{}).get('status','')")"
[[ "$STATUS" == "ok" ]] || { echo "unexpected health: $HEALTH" >&2; exit 1; }

TOKEN="$(api_login "$BASE")"
ME="$(curl -fsS "$BASE/api/v1/auth/me" -H "Authorization: Bearer $TOKEN")"
MENUS="$(json_get "$ME" "len(o.get('data',{}).get('menus') or [])")"
PERMS="$(json_get "$ME" "len(o.get('data',{}).get('permissions') or [])")"
echo "auth/me menus=$MENUS permissions=$PERMS"
[[ "$MENUS" != "0" ]] || { echo "expected non-empty menus from server" >&2; exit 1; }

# Embed injects encryption key into index.html
INDEX="$(curl -fsS "$BASE/")"
echo "$INDEX" | grep -q '__BEDROCK_ENCRYPTION_KEY__' || {
  echo "WARN: index.html missing __BEDROCK_ENCRYPTION_KEY__ injection (placeholder dist?)" >&2
}

# Deep-link style SPA fallback (bookmark path)
CODE="$(curl -sS -o /dev/null -w '%{http_code}' "$BASE/cicd/build-runs/1")"
[[ "$CODE" == "200" ]] || { echo "SPA deep link returned $CODE" >&2; exit 1; }

echo "PASS: fresh-install smoke (SQLite)"
