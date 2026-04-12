#!/usr/bin/env bash
set -euo pipefail

if (( $# != 1 )); then
  echo "Usage: $0 <entry-file-relative-to-apps/web-antd>" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_ROOT="$ROOT_DIR/vue-vben-admin/apps/web-antd"
ENTRY_RELATIVE="${1#./}"
ENTRY_PATH="$APP_ROOT/$ENTRY_RELATIVE"

if [[ "$ENTRY_RELATIVE" != src/* ]]; then
  echo "Entry must be under apps/web-antd/src: $ENTRY_RELATIVE" >&2
  exit 2
fi

if [[ ! -f "$ENTRY_PATH" ]]; then
  echo "Entry file not found: $ENTRY_PATH" >&2
  exit 2
fi

TEMP_ENTRY="$(mktemp "$APP_ROOT/easymvp-entry-bundle.XXXXXX.ts")"
cleanup() {
  rm -f "$TEMP_ENTRY"
}
trap cleanup EXIT

cat >"$TEMP_ENTRY" <<EOF
import './$ENTRY_RELATIVE';

export default {};
EOF

TEMP_ENTRY_BASENAME="$(basename "$TEMP_ENTRY")"
OUT_DIR_SUFFIX="$(printf '%s' "$ENTRY_RELATIVE" | tr '/.' '--')"

cd "$ROOT_DIR"
EASYMVP_BUILD_LABEL="web-antd entry bundle build: ${ENTRY_RELATIVE}" \
EASYMVP_WEB_ANTD_VERIFY_BUILD=1 \
EASYMVP_WEB_ANTD_BUNDLE_ENTRY="$TEMP_ENTRY_BASENAME" \
EASYMVP_WEB_ANTD_BUNDLE_OUT_DIR="dist-entry-verify-${OUT_DIR_SUFFIX}" \
./scripts/web-antd-ci-build.sh
