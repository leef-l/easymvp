#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/core"
TMP_ROOT="${TMPDIR:-/tmp}/easymvp-core-smoke-$$"
BIN_PATH="$TMP_ROOT/easymvp-core"
DATA_ROOT="$TMP_ROOT/var"
DB_PATH="$DATA_ROOT/data/easymvp.db"
MIGRATION_PATH="$CORE_DIR/manifest/migrations"
HOST_PORT="${HOST_PORT:-18000}"
HEALTH_URL="http://127.0.0.1:${HOST_PORT}/api/v3/system/healthz"
PID=""

cleanup() {
  if [ -n "$PID" ]; then
    kill "$PID" >/dev/null 2>&1 || true
    wait "$PID" >/dev/null 2>&1 || true
  fi
  rm -rf "$TMP_ROOT"
}
trap cleanup EXIT

mkdir -p "$TMP_ROOT"

echo "== Build apps/core smoke binary =="
(
  cd "$CORE_DIR"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$BIN_PATH" ./main.go
)

echo
echo "== Run apps/core smoke process =="
"$BIN_PATH" \
  --safe-mode \
  --port "$HOST_PORT" \
  --data-root "$DATA_ROOT" \
  --db-path "$DB_PATH" \
  --migration-path "$MIGRATION_PATH" \
  >/tmp/easymvp-core-smoke.log 2>&1 &
PID="$!"

echo
echo "== Probe healthz =="
for _ in $(seq 1 30); do
  if curl --fail --silent "$HEALTH_URL" >/dev/null; then
    echo "healthz ok: $HEALTH_URL"
    exit 0
  fi
  sleep 1
done

echo "healthz probe failed: $HEALTH_URL" >&2
cat /tmp/easymvp-core-smoke.log >&2 || true
exit 1
