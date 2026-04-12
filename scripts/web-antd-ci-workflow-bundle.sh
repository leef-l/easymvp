#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"
EASYMVP_BUILD_LABEL="web-antd workflow bundle build" \
EASYMVP_WEB_ANTD_VERIFY_BUILD=1 \
EASYMVP_WEB_ANTD_WORKFLOW_BUNDLE=1 \
./scripts/web-antd-ci-build.sh
