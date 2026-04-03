#!/usr/bin/env bash

set -euo pipefail

cd /workspace/vue-vben-admin

if [[ -n "${NPM_REGISTRY:-}" ]]; then
  pnpm config set registry "${NPM_REGISTRY}"
fi

if [[ ! -d node_modules/.pnpm ]]; then
  echo "Restoring prebuilt frontend dependencies..."
  tar -xzf /opt/vue-vben-admin-prebuilt/deps.tgz -C /workspace/vue-vben-admin
fi

exec pnpm -F @vben/web-antd run dev --host 0.0.0.0 --port "${WEB_CONTAINER_PORT:-5666}"
