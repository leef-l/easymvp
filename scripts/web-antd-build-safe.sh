#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$ROOT_DIR/scripts/github-actions-only.sh" "$(basename "${BASH_SOURCE[0]}")" ".github/workflows/web-antd-guard.yml"
APP_DIR="$ROOT_DIR/vue-vben-admin"
LOCK_FILE="${TMPDIR:-/tmp}/easymvp-web-antd-build.lock"

CPU_CORE="${EASYMVP_PNPM_CPU_CORE:-0}"
MEMORY_LIMIT_MB="${EASYMVP_PNPM_MEMORY_MB:-1024}"
HEAP_MB="${EASYMVP_BUILD_HEAP_MB:-768}"
MIN_AVAILABLE_MB="${EASYMVP_BUILD_MIN_AVAILABLE_MB:-1024}"
BUILD_LABEL="${EASYMVP_BUILD_LABEL:-web-antd production build}"

if ! command -v systemd-run >/dev/null 2>&1; then
  echo "systemd-run is required to enforce the 1 CPU / 1G memory guard" >&2
  exit 5
fi

if ! [[ "$CPU_CORE" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_PNPM_CPU_CORE must be an integer, got: $CPU_CORE" >&2
  exit 2
fi

if ! [[ "$MEMORY_LIMIT_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_PNPM_MEMORY_MB must be an integer, got: $MEMORY_LIMIT_MB" >&2
  exit 2
fi

if ! [[ "$HEAP_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_BUILD_HEAP_MB must be an integer, got: $HEAP_MB" >&2
  exit 2
fi

if ! [[ "$MIN_AVAILABLE_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_BUILD_MIN_AVAILABLE_MB must be an integer, got: $MIN_AVAILABLE_MB" >&2
  exit 2
fi

if (( HEAP_MB <= 0 || MEMORY_LIMIT_MB <= 0 || MIN_AVAILABLE_MB <= 0 )); then
  echo "CPU/memory limits must be positive integers" >&2
  exit 2
fi

if (( HEAP_MB >= MEMORY_LIMIT_MB )); then
  echo "EASYMVP_BUILD_HEAP_MB must be smaller than EASYMVP_PNPM_MEMORY_MB (${HEAP_MB} >= ${MEMORY_LIMIT_MB})" >&2
  exit 2
fi

available_mb="$(awk '/MemAvailable:/ { printf "%d", $2 / 1024 }' /proc/meminfo)"
if [[ -z "$available_mb" ]]; then
  echo "Unable to read MemAvailable from /proc/meminfo" >&2
  exit 2
fi

if (( available_mb < MIN_AVAILABLE_MB )); then
  echo "Skip ${BUILD_LABEL}: available memory ${available_mb}MB < required ${MIN_AVAILABLE_MB}MB" >&2
  exit 3
fi

mkdir -p "$(dirname "$LOCK_FILE")"
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  echo "Another ${BUILD_LABEL} is already running: $LOCK_FILE" >&2
  exit 4
fi

echo "Running ${BUILD_LABEL} on current server"
echo "  root: $ROOT_DIR"
echo "  cpu_core: $CPU_CORE"
echo "  memory_limit_mb: $MEMORY_LIMIT_MB"
echo "  heap_mb: $HEAP_MB"
echo "  mem_available_mb: $available_mb"
echo "  lock_file: $LOCK_FILE"

nice_cmd=(nice -n 15)
ionice_cmd=()
if command -v ionice >/dev/null 2>&1; then
  ionice_cmd=(ionice -c3)
fi

cd "$ROOT_DIR"
set +e
systemd-run --scope --quiet \
  -p "AllowedCPUs=${CPU_CORE}" \
  -p "CPUQuota=100%" \
  -p "MemoryMax=${MEMORY_LIMIT_MB}M" \
  -p "MemorySwapMax=0" \
  -p "TasksMax=256" \
  env \
  CI=1 \
  GOMAXPROCS=1 \
  GOMEMLIMIT=768MiB \
  NODE_OPTIONS="--max-old-space-size=${HEAP_MB}" \
  UV_THREADPOOL_SIZE=1 \
  npm_config_jobs=1 \
  npm_config_maxsockets=4 \
  pnpm_config_child_concurrency=1 \
  pnpm_config_workspace_concurrency=1 \
  pnpm_config_network_concurrency=1 \
  TURBO_CONCURRENCY=1 \
  "${nice_cmd[@]}" \
  "${ionice_cmd[@]}" \
  pnpm -C "$APP_DIR" --filter @vben/web-antd exec vite build --mode production
status=$?
set -e

if (( status == 137 || status == 143 )); then
  echo "${BUILD_LABEL} was terminated under the 1 CPU / 1G guard (exit=${status})" >&2
fi

exit "$status"
