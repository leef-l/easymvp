#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/vue-vben-admin"

HEAP_MB="${EASYMVP_BUILD_HEAP_MB:-768}"
BUILD_LABEL="${EASYMVP_BUILD_LABEL:-web-antd ci build}"

if ! [[ "$HEAP_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_BUILD_HEAP_MB must be an integer, got: $HEAP_MB" >&2
  exit 2
fi

if (( HEAP_MB <= 0 )); then
  echo "EASYMVP_BUILD_HEAP_MB must be positive" >&2
  exit 2
fi

echo "Running ${BUILD_LABEL} in CI container"
echo "  root: $ROOT_DIR"
echo "  heap_mb: $HEAP_MB"
echo "  verify_build: ${EASYMVP_WEB_ANTD_VERIFY_BUILD:-0}"
echo "  workflow_bundle: ${EASYMVP_WEB_ANTD_WORKFLOW_BUNDLE:-0}"
echo "  bundle_entry: ${EASYMVP_WEB_ANTD_BUNDLE_ENTRY:-}"
echo "  bundle_out_dir: ${EASYMVP_WEB_ANTD_BUNDLE_OUT_DIR:-}"

cd "$ROOT_DIR"
CI=1 \
NODE_OPTIONS="--max-old-space-size=${HEAP_MB}" \
UV_THREADPOOL_SIZE="${UV_THREADPOOL_SIZE:-1}" \
npm_config_jobs="${npm_config_jobs:-1}" \
npm_config_maxsockets="${npm_config_maxsockets:-4}" \
pnpm_config_child_concurrency="${pnpm_config_child_concurrency:-1}" \
pnpm_config_workspace_concurrency="${pnpm_config_workspace_concurrency:-1}" \
pnpm_config_network_concurrency="${pnpm_config_network_concurrency:-1}" \
TURBO_CONCURRENCY="${TURBO_CONCURRENCY:-1}" \
pnpm -C "$APP_DIR" --filter @vben/web-antd exec vite build --mode production
