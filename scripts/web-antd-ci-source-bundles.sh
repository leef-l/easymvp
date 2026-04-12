#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"
EASYMVP_WEB_ANTD_ENTRY_BUNDLE_RUNNER="$ROOT_DIR/scripts/web-antd-ci-entry-bundle.sh" \
./scripts/web-antd-source-bundles.sh
