#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$ROOT_DIR/scripts/github-actions-only.sh" "$(basename "${BASH_SOURCE[0]}")" ".github/workflows/web-antd-guard.yml"

cd "$ROOT_DIR"
EASYMVP_BUILD_LABEL="web-antd verification build" \
EASYMVP_WEB_ANTD_VERIFY_BUILD=1 \
./scripts/web-antd-build-safe.sh
