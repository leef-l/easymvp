#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"
EASYMVP_BUILD_LABEL="web-antd full production build" \
./scripts/web-antd-ci-build.sh
