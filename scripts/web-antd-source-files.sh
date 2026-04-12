#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$ROOT_DIR"
find vue-vben-admin/apps/web-antd/src -type f \( -name '*.vue' -o -name '*.ts' -o -name '*.tsx' \) \
  | sed 's#^vue-vben-admin/apps/web-antd/##' \
  | LC_ALL=C sort
