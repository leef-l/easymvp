#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
bash "${repo_root}/scripts/github-actions-only.sh" "$(basename "${BASH_SOURCE[0]}")" ".github/workflows/backend-guard.yml"

cd "${repo_root}/admin-go"
export GF_GCFG_PATH="${GF_GCFG_PATH:-app/mvp/manifest/config}"
export GF_GCFG_FILE="${GF_GCFG_FILE:-config.yaml}"

go run ./app/mvp/regressioncheck "${repo_root}/test-workspaces/regression-manifest.json"
