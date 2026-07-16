#!/usr/bin/env bash
# Documented restart-recovery verification via Go unit tests (BuildRun / AgentRun / install jobs).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

echo "==> BuildRun scheduler recovery (queued / interrupted)"
go test ./internal/engine/ -run 'TestSchedulerRecovery' -count=1

echo "==> AgentRun / CLI install recovery (package tests covering RecoverOnStartup)"
go test ./internal/ai/service/ -count=1

echo "==> Ops dev-environment install recovery"
go test ./internal/ops/service/ -count=1

echo "PASS: restart-recovery (unit/integration tests)"
