#!/usr/bin/env bash
set -euo pipefail

if (( $# != 1 )); then
  echo "Usage: $0 <entry-file-relative-to-apps/web-antd>" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bash "$ROOT_DIR/scripts/github-actions-only.sh" "$(basename "${BASH_SOURCE[0]}")" ".github/workflows/web-antd-guard.yml"
APP_ROOT="$ROOT_DIR/vue-vben-admin/apps/web-antd"
ENTRY_RELATIVE="$1"
ENTRY_PATH="$APP_ROOT/$ENTRY_RELATIVE"

if [[ ! -f "$ENTRY_PATH" ]]; then
  echo "Entry file not found: $ENTRY_PATH" >&2
  exit 2
fi

mkdir -p "$APP_ROOT/node_modules/.tmp"
TEMP_TSCONFIG="$(mktemp "$APP_ROOT/node_modules/.tmp/easymvp-web-antd-entry-typecheck.XXXXXX.json")"
cleanup() {
  rm -f "$TEMP_TSCONFIG"
}
trap cleanup EXIT

cat >"$TEMP_TSCONFIG" <<EOF
{
  "\$schema": "https://json.schemastore.org/tsconfig",
  "extends": "../../tsconfig.json",
  "include": [
    "../../$ENTRY_RELATIVE"
  ]
}
EOF

cd "$ROOT_DIR"
EASYMVP_TYPECHECK_PROJECT="$TEMP_TSCONFIG" \
EASYMVP_TYPECHECK_LABEL="web-antd entry typecheck: ${ENTRY_RELATIVE}" \
./scripts/web-antd-typecheck-safe.sh
