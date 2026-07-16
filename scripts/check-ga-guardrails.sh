#!/usr/bin/env bash
# GA guardrails: fail if a supported 1.x→2.0 data migration path is introduced.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

fail=0

# Ban executable migration entrypoints that claim to upgrade 1.x data into 2.0.
# Schema migrations under internal/platform/migration/ are expected (2.0-only).
for path in \
  scripts/migrate-1x \
  scripts/migrate-from-1x.sh \
  scripts/upgrade-from-1x.sh \
  cmd/migrate1x \
  internal/legacy/migrate
do
  if [[ -e "$path" ]]; then
    echo "ERROR: found 1.x→2.0 data migration entrypoint: $path" >&2
    fail=1
  fi
done

# Affirmative "we support 1.x migration" claims (negations are fine).
bad_docs="$(rg -n --glob '*.md' \
  '(提供|支持|可用).{0,12}1\.x.{0,12}(数据)?迁移|(数据)?迁移.{0,12}1\.x.{0,12}(支持|可用)' \
  docs AGENTS.md .agents 2>/dev/null || true)"
if [[ -n "$bad_docs" ]]; then
  filtered="$(printf '%s\n' "$bad_docs" | rg -v '不提供|未提供|不支持|不可用|不迁移|禁止|不做|明确不做|~~|作为支持路径' || true)"
  if [[ -n "$filtered" ]]; then
    echo "$filtered" >&2
    echo "ERROR: docs appear to claim 1.x data migration is supported" >&2
    fail=1
  fi
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "GA guardrails OK (no 1.x data-migration support path)"
