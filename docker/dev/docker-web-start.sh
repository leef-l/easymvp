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

# 确保 monorepo internal 包已构建（stub 生成 dist/）
if [[ ! -f internal/vite-config/dist/index.mjs ]]; then
  echo "Building internal packages (stub)..."
  pnpm -r run stub --if-present
fi

exec pnpm -F @vben/web-antd run dev --host 0.0.0.0 --port "${WEB_CONTAINER_PORT:-5666}"
