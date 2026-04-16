#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$ROOT_DIR/scripts/github-actions-only.sh" "$(basename "${BASH_SOURCE[0]}")" ".github/workflows/web-antd-guard.yml"

cd "$ROOT_DIR"
EASYMVP_TYPECHECK_PROJECT="apps/web-antd/tsconfig.mvp-workflow.json" \
EASYMVP_TYPECHECK_LABEL="web-antd mvp workflow typecheck" \
./scripts/web-antd-typecheck-safe.sh
