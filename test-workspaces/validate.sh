#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"

cd "${repo_root}/admin-go"
go run ./app/mvp/regressioncheck "${repo_root}/test-workspaces/regression-manifest.json"
