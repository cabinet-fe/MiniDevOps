#!/usr/bin/env bash
# Shared helpers for smoke scripts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SMOKE_TMP="${SMOKE_TMP:-$ROOT/.tmp/smoke}"
ENCRYPTION_KEY="${BEDROCK_SMOKE_ENCRYPTION_KEY:-0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef}"
ADMIN_USER="${BEDROCK_SMOKE_USER:-admin}"
ADMIN_PASS="${BEDROCK_SMOKE_PASS:-admin123}"

ensure_dirs() {
  mkdir -p "$SMOKE_TMP"
}

json_get() {
  # json_get <json> <python expr on obj>
  local json="$1"
  local expr="$2"
  python3 -c "import json,sys; o=json.load(sys.stdin); print($expr)" <<<"$json"
}

# Golden AES-256-CBC hex(IV||ciphertext) for plaintext "admin123" with config.yaml encryption.key
# (see internal/pkg/crypto_golden_test.go). Prefer password_cipher; plaintext is debug-only fallback.
login_payload() {
  if [[ "$ADMIN_PASS" == "admin123" ]]; then
    local cipher="000102030405060708090a0b0c0d0e0f17f1b26aff75e950ec141048626a9ed8"
    printf '{"username":"%s","password_cipher":"%s"}' "$ADMIN_USER" "$cipher"
  else
    printf '{"username":"%s","password":"%s"}' "$ADMIN_USER" "$ADMIN_PASS"
  fi
}

wait_http() {
  local url="$1"
  local tries="${2:-60}"
  local i=0
  while (( i < tries )); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
    i=$((i + 1))
  done
  echo "timeout waiting for $url" >&2
  return 1
}

api_login() {
  local base="$1"
  local body
  body="$(login_payload)"
  local resp
  resp="$(curl -fsS -X POST "$base/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "$body")"
  local token
  token="$(json_get "$resp" "o.get('data',{}).get('access_token','')")"
  if [[ -z "$token" || "$token" == "None" ]]; then
    echo "login failed: $resp" >&2
    return 1
  fi
  printf '%s' "$token"
}

write_smoke_config() {
  local cfg="$1"
  local data_dir="$2"
  local driver="${3:-sqlite}"
  local port="${4:-${SMOKE_PORT:-18080}}"
  cat >"$cfg" <<EOF
server:
  port: ${port}
  host: "127.0.0.1"
database:
  driver: ${driver}
  path: "${data_dir}/bedrock.sqlite"
  host: "${BEDROCK_DB_HOST:-127.0.0.1}"
  port: ${BEDROCK_DB_PORT:-5432}
  name: "${BEDROCK_DB_NAME:-bedrock_smoke}"
  user: "${BEDROCK_DB_USER:-bedrock}"
  password: "${BEDROCK_DB_PASSWORD:-bedrock}"
  ssl_mode: disable
  max_open_conns: 10
  max_idle_conns: 2
  conn_max_lifetime: 1h
jwt:
  secret: "bedrock-smoke-secret"
  access_ttl: "2h"
  refresh_ttl: "168h"
build:
  max_concurrent: 2
  workspace_dir: "${data_dir}/workspaces"
  artifact_dir: "${data_dir}/artifacts"
  log_dir: "${data_dir}/logs"
  cache_dir: "${data_dir}/caches"
storage:
  root: "${data_dir}/storage"
  attachment_max_bytes: 20971520
  doc_import_max_bytes: 104857600
encryption:
  key: "${ENCRYPTION_KEY}"
admin:
  username: "${ADMIN_USER}"
  password: "${ADMIN_PASS}"
  display_name: "管理员"
EOF
}

ensure_embed_dist() {
  mkdir -p "$ROOT/cmd/server/dist"
  if [[ ! -f "$ROOT/cmd/server/dist/index.html" ]]; then
    if [[ -f "$ROOT/web-v2/dist/index.html" ]]; then
      rm -rf "$ROOT/cmd/server/dist"
      cp -R "$ROOT/web-v2/dist" "$ROOT/cmd/server/dist"
    else
      printf '%s\n' '<!doctype html><html><head></head><body>smoke placeholder</body></html>' \
        >"$ROOT/cmd/server/dist/index.html"
    fi
  fi
}

build_server_bin() {
  ensure_embed_dist
  local out="${1:-$SMOKE_TMP/bedrock}"
  (cd "$ROOT" && CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=smoke" -o "$out" ./cmd/server)
  printf '%s' "$out"
}
