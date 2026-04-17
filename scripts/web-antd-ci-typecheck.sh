#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/vue-vben-admin"

HEAP_MB="${EASYMVP_TYPECHECK_HEAP_MB:-768}"
TYPECHECK_LABEL="${EASYMVP_TYPECHECK_LABEL:-web-antd full typecheck}"
TYPECHECK_PROJECT="${EASYMVP_TYPECHECK_PROJECT:-apps/web-antd/tsconfig.json}"

if ! [[ "$HEAP_MB" =~ ^[0-9]+$ ]]; then
  echo "EASYMVP_TYPECHECK_HEAP_MB must be an integer, got: $HEAP_MB" >&2
  exit 2
fi

if (( HEAP_MB <= 0 )); then
  echo "EASYMVP_TYPECHECK_HEAP_MB must be positive" >&2
  exit 2
fi

echo "Running ${TYPECHECK_LABEL} in CI container"
echo "  root: $ROOT_DIR"
echo "  heap_mb: $HEAP_MB"
echo "  tsconfig: $TYPECHECK_PROJECT"
echo "  skip_lib_check: 1"

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
pnpm -C "$APP_DIR" exec vue-tsc --noEmit --skipLibCheck -p "$TYPECHECK_PROJECT"
