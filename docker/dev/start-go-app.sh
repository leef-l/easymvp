#!/usr/bin/env bash

set -euo pipefail

APP_NAME="${1:-}"
APP_PORT="${2:-8000}"

if [[ -z "${APP_NAME}" ]]; then
  echo "missing app name"
  exit 1
fi

if [[ -z "${DB_HOST:-}" || -z "${DB_USER:-}" || -z "${DB_PASSWORD:-}" || -z "${DB_NAME:-}" ]]; then
  echo "database env is incomplete"
  exit 1
fi

CONFIG_DIR="/workspace/admin-go/.runtime-config"
CONFIG_FILE="${CONFIG_DIR}/${APP_NAME}.yaml"

mkdir -p "${CONFIG_DIR}"

cat > "${CONFIG_FILE}" <<EOF
server:
  address: ":${APP_PORT}"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

database:
  default:
    link: "mysql:${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT:-3306})/${DB_NAME}?charset=utf8mb4&loc=Local&parseTime=true"
    debug: false

logger:
  level: "info"
  stdout: true

jwt:
  secret: "${JWT_SECRET:-easymvp-secret-key}"
  expire: 24
EOF

export GF_GCFG_FILE="${CONFIG_FILE}"

export PATH="/usr/local/go/bin:/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/root/.local/share/uv/tools/aider-chat/bin:/root/.local/share/uv/tools/openhands/bin:${PATH:-}"

if [[ -x "/root/.local/share/uv/tools/aider-chat/bin/aider" ]]; then
  ln -sf /root/.local/share/uv/tools/aider-chat/bin/aider /usr/local/bin/aider
fi

if [[ -x "/root/.local/share/uv/tools/openhands/bin/openhands" ]]; then
  ln -sf /root/.local/share/uv/tools/openhands/bin/openhands /usr/local/bin/openhands
fi

if [[ -x "/root/.local/share/uv/tools/openhands/bin/openhands-acp" ]]; then
  ln -sf /root/.local/share/uv/tools/openhands/bin/openhands-acp /usr/local/bin/openhands-acp
fi

cd "/workspace/admin-go/app/${APP_NAME}"
exec go run main.go
