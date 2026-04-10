#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/vue-vben-admin"
LOCK_FILE="${TMPDIR:-/tmp}/easymvp-web-antd-typecheck.lock"

HEAP_MB="${EASYMVP_TYPECHECK_HEAP_MB:-1536}"
MIN_AVAILABLE_MB="${EASYMVP_TYPECHECK_MIN_AVAILABLE_MB:-2048}"

if ! [[ "$HEAP_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_TYPECHECK_HEAP_MB must be an integer, got: $HEAP_MB" >&2
  exit 2
fi

if ! [[ "$MIN_AVAILABLE_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_TYPECHECK_MIN_AVAILABLE_MB must be an integer, got: $MIN_AVAILABLE_MB" >&2
  exit 2
fi

available_mb="$(awk '/MemAvailable:/ { printf "%d", $2 / 1024 }' /proc/meminfo)"
if [[ -z "$available_mb" ]]; then
  echo "Unable to read MemAvailable from /proc/meminfo" >&2
  exit 2
fi

if (( available_mb < MIN_AVAILABLE_MB )); then
  echo "Skip web-antd full typecheck: available memory ${available_mb}MB < required ${MIN_AVAILABLE_MB}MB" >&2
  exit 3
fi

mkdir -p "$(dirname "$LOCK_FILE")"
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  echo "Another web-antd typecheck is already running: $LOCK_FILE" >&2
  exit 4
fi

echo "Running web-antd full typecheck on current server"
echo "  root: $ROOT_DIR"
echo "  heap_mb: $HEAP_MB"
echo "  mem_available_mb: $available_mb"
echo "  lock_file: $LOCK_FILE"

nice_cmd=(nice -n 15)
ionice_cmd=()
if command -v ionice >/dev/null 2>&1; then
  ionice_cmd=(ionice -c3)
fi

cd "$ROOT_DIR"
env \
  CI=1 \
  NODE_OPTIONS="--max-old-space-size=${HEAP_MB}" \
  npm_config_maxsockets=4 \
  pnpm_config_child_concurrency=1 \
  pnpm_config_workspace_concurrency=1 \
  TURBO_CONCURRENCY=1 \
  "${nice_cmd[@]}" \
  "${ionice_cmd[@]}" \
  pnpm -C "$APP_DIR" exec vue-tsc --noEmit -p apps/web-antd/tsconfig.json
