#!/usr/bin/env bash
# Three-database smoke: always runs SQLite; Postgres/MySQL when DSN/env available.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

ensure_dirs

run_driver() {
  local driver="$1"
  local port="$2"
  local data="$SMOKE_TMP/db-$driver"
  local cfg="$SMOKE_TMP/db-$driver.yaml"
  local log="$SMOKE_TMP/db-$driver.log"
  local base="http://127.0.0.1:${port}"

  rm -rf "$data"
  mkdir -p "$data"
  write_smoke_config "$cfg" "$data" "$driver" "$port"

  local bin
  bin="$(build_server_bin "$SMOKE_TMP/bedrock-$driver")"
  echo "==> $driver on :$port"
  "$bin" --config "$cfg" >"$log" 2>&1 &
  local pid=$!
  if ! wait_http "$base/api/v1/health" 90; then
    echo "FAIL: $driver did not become healthy" >&2
    tail -n 80 "$log" >&2 || true
    kill "$pid" 2>/dev/null || true
    return 1
  fi
  local token
  token="$(SMOKE_PORT="$port" api_login "$base")"
  curl -fsS "$base/api/v1/auth/me" -H "Authorization: Bearer $token" >/dev/null
  kill "$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
  echo "PASS: $driver"
}

PASSED=()
SKIPPED=()

run_driver sqlite 18180
PASSED+=(sqlite)

if [[ -n "${BEDROCK_CONTRACT_POSTGRES_DSN:-}" ]] || [[ -n "${BEDROCK_SMOKE_POSTGRES:-}" ]]; then
  export BEDROCK_DB_HOST="${BEDROCK_DB_HOST:-127.0.0.1}"
  export BEDROCK_DB_PORT="${BEDROCK_DB_PORT:-5432}"
  export BEDROCK_DB_NAME="${BEDROCK_DB_NAME:-bedrock_smoke}"
  export BEDROCK_DB_USER="${BEDROCK_DB_USER:-bedrock}"
  export BEDROCK_DB_PASSWORD="${BEDROCK_DB_PASSWORD:-bedrock}"
  if run_driver postgres 18181; then
    PASSED+=(postgres)
  else
    echo "FAIL: postgres smoke" >&2
    exit 1
  fi
else
  SKIPPED+=(postgres)
  echo "SKIP: postgres (set BEDROCK_SMOKE_POSTGRES=1 or BEDROCK_CONTRACT_POSTGRES_DSN)"
fi

if [[ -n "${BEDROCK_CONTRACT_MYSQL_DSN:-}" ]] || [[ -n "${BEDROCK_SMOKE_MYSQL:-}" ]]; then
  export BEDROCK_DB_HOST="${BEDROCK_DB_HOST:-127.0.0.1}"
  export BEDROCK_DB_PORT="${BEDROCK_DB_PORT:-3306}"
  export BEDROCK_DB_NAME="${BEDROCK_DB_NAME:-bedrock_smoke}"
  export BEDROCK_DB_USER="${BEDROCK_DB_USER:-bedrock}"
  export BEDROCK_DB_PASSWORD="${BEDROCK_DB_PASSWORD:-bedrock}"
  if run_driver mysql 18182; then
    PASSED+=(mysql)
  else
    echo "FAIL: mysql smoke" >&2
    exit 1
  fi
else
  SKIPPED+=(mysql)
  echo "SKIP: mysql (set BEDROCK_SMOKE_MYSQL=1 or BEDROCK_CONTRACT_MYSQL_DSN)"
fi

echo "three-db summary: passed=${PASSED[*]} skipped=${SKIPPED[*]:-none}"
# Contract-tagged unit tests remain the authoritative multi-driver matrix when services exist.
if [[ ${#PASSED[@]} -lt 1 ]]; then
  exit 1
fi
echo "PASS: three-db smoke (at least SQLite)"
