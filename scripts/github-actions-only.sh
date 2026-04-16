#!/usr/bin/env bash
set -euo pipefail

if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
  exit 0
fi

script_name="${1:-this script}"
workflow_path="${2:-.github/workflows/}"

cat >&2 <<EOF
${script_name} is GitHub Actions-only.
Local test/build/guard execution is disabled by EasyMVP engineering rules.
Trigger the corresponding workflow instead: ${workflow_path}
EOF
exit 64
