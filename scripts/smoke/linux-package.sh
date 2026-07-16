#!/usr/bin/env bash
# Cross-compile Linux amd64/arm64 Server + Deploy Agent; optional start smoke for amd64 on Linux hosts.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

ensure_dirs
OUT="$SMOKE_TMP/dist"
rm -rf "$OUT"
mkdir -p "$OUT"

ensure_embed_dist

echo "==> cross-compile Server + Agent"
(
  cd "$ROOT"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=smoke" \
    -o "$OUT/bedrock-linux-amd64" ./cmd/server
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=smoke" \
    -o "$OUT/bedrock-agent-linux-amd64" ./cmd/agent
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=smoke" \
    -o "$OUT/bedrock-linux-arm64" ./cmd/server
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=smoke" \
    -o "$OUT/bedrock-agent-linux-arm64" ./cmd/agent
)

(
  cd "$OUT"
  sha256sum bedrock-linux-amd64 bedrock-agent-linux-amd64 \
    bedrock-linux-arm64 bedrock-agent-linux-arm64 > SHA256SUMS
)

ls -la "$OUT"
echo "artifacts:"
cat "$OUT/SHA256SUMS"

HOST_ARCH="$(uname -m)"
HOST_OS="$(uname -s)"
if [[ "$HOST_OS" == "Linux" && "$HOST_ARCH" == "x86_64" ]]; then
  DATA="$SMOKE_TMP/pkg-data"
  CFG="$SMOKE_TMP/pkg-config.yaml"
  PORT="${SMOKE_PORT:-18082}"
  rm -rf "$DATA"
  mkdir -p "$DATA"
  SMOKE_PORT="$PORT" write_smoke_config "$CFG" "$DATA" sqlite "$PORT"
  echo "==> start amd64 package smoke"
  "$OUT/bedrock-linux-amd64" --config "$CFG" >"$SMOKE_TMP/pkg-server.log" 2>&1 &
  PID=$!
  cleanup() { kill "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; }
  trap cleanup EXIT
  wait_http "http://127.0.0.1:${PORT}/api/v1/health" 80
  TOKEN="$(api_login "http://127.0.0.1:${PORT}")"
  curl -fsS "http://127.0.0.1:${PORT}/api/v1/auth/me" -H "Authorization: Bearer $TOKEN" >/dev/null
  echo "PASS: linux amd64 package start+login"
else
  echo "SKIP: runtime package start (host=$HOST_OS/$HOST_ARCH); binaries produced for amd64+arm64"
fi

echo "PASS: linux-package smoke (binaries + checksums)"
