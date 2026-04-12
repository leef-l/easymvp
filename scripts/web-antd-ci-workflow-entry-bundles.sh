#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

DEFAULT_ENTRIES=(
  "src/views/mvp/workflow/objective.vue"
  "src/views/mvp/workflow/situation.vue"
  "src/views/mvp/workflow/dashboard.vue"
  "src/views/mvp/workflow/execution.vue"
  "src/views/mvp/workflow/review.vue"
  "src/views/mvp/workflow/accept.vue"
  "src/views/mvp/workflow/verification.vue"
  "src/views/mvp/workflow/autonomy.vue"
)

if (( $# > 0 )); then
  ENTRIES=("$@")
else
  ENTRIES=("${DEFAULT_ENTRIES[@]}")
fi

total="${#ENTRIES[@]}"
if (( total == 0 )); then
  echo "No workflow entries provided" >&2
  exit 2
fi

cd "$ROOT_DIR"
for idx in "${!ENTRIES[@]}"; do
  entry="${ENTRIES[$idx]}"
  echo "[$((idx + 1))/$total] ${entry}"
  ./scripts/web-antd-ci-entry-bundle.sh "$entry"
done

echo "web-antd workflow entry bundle verification passed: ${total} file(s)"
