#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNNER="${EASYMVP_WEB_ANTD_ENTRY_BUNDLE_RUNNER:-$ROOT_DIR/scripts/web-antd-ci-entry-bundle.sh}"
SHARD_INDEX="${EASYMVP_WEB_ANTD_SHARD_INDEX:-1}"
SHARD_TOTAL="${EASYMVP_WEB_ANTD_SHARD_TOTAL:-1}"

if ! [[ "$SHARD_INDEX" =~ ^[0-9]+$ ]] || ! [[ "$SHARD_TOTAL" =~ ^[0-9]+$ ]]; then
  echo "Shard index/total must be positive integers" >&2
  exit 2
fi

if (( SHARD_INDEX <= 0 || SHARD_TOTAL <= 0 || SHARD_INDEX > SHARD_TOTAL )); then
  echo "Invalid shard selection: index=${SHARD_INDEX} total=${SHARD_TOTAL}" >&2
  exit 2
fi

if [[ ! -x "$RUNNER" ]]; then
  echo "Entry bundle runner is not executable: $RUNNER" >&2
  exit 2
fi

mapfile -t SOURCE_FILES < <("$ROOT_DIR/scripts/web-antd-source-files.sh")
TOTAL_SOURCE_COUNT="${#SOURCE_FILES[@]}"
if (( TOTAL_SOURCE_COUNT == 0 )); then
  echo "No web-antd source files found" >&2
  exit 2
fi

SELECTED_FILES=()
for idx in "${!SOURCE_FILES[@]}"; do
  ordinal=$((idx + 1))
  if (((ordinal - 1) % SHARD_TOTAL == SHARD_INDEX - 1)); then
    SELECTED_FILES+=("${SOURCE_FILES[$idx]}")
  fi
done

SELECTED_COUNT="${#SELECTED_FILES[@]}"
if (( SELECTED_COUNT == 0 )); then
  echo "No source files selected for shard ${SHARD_INDEX}/${SHARD_TOTAL}" >&2
  exit 2
fi

echo "Running web-antd full-source bundle shard ${SHARD_INDEX}/${SHARD_TOTAL}"
echo "  root: $ROOT_DIR"
echo "  runner: $RUNNER"
echo "  total_source_files: $TOTAL_SOURCE_COUNT"
echo "  shard_source_files: $SELECTED_COUNT"

cd "$ROOT_DIR"
for idx in "${!SELECTED_FILES[@]}"; do
  entry="${SELECTED_FILES[$idx]}"
  echo "[$((idx + 1))/$SELECTED_COUNT] ${entry}"
  "$RUNNER" "$entry"
done

echo "web-antd full-source bundle shard passed: ${SHARD_INDEX}/${SHARD_TOTAL} (${SELECTED_COUNT}/${TOTAL_SOURCE_COUNT})"
