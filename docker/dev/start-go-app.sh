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
LOG_BASE_PATH="${GF_LOG_PATH:-/workspace/admin-go/logs}"
LOG_DIR="${LOG_BASE_PATH}/${APP_NAME}"

mkdir -p "${CONFIG_DIR}" "${LOG_DIR}"

REDIS_CONFIG=""
if [[ -n "${REDIS_ADDR:-}" ]]; then
  REDIS_CONFIG=$(cat <<EOF

redis:
  default:
    address: "${REDIS_ADDR}"
    pass: "${REDIS_PASS:-}"
    db: 0
EOF
)
fi

cat > "${CONFIG_FILE}" <<EOF
server:
  address: ":${APP_PORT}"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

database:
  default:
    link: "mysql:${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT:-3306})/${DB_NAME}?charset=utf8mb4&loc=Local&parseTime=true"
    debug: false
${REDIS_CONFIG}

logger:
  path: "${LOG_DIR}"
  level: "${GF_LOG_LEVEL:-all}"
  stdout: ${GF_LOG_STDOUT:-false}
  rotateSize: "${GF_LOG_ROTATE_SIZE:-100M}"
  rotateExpire: "${GF_LOG_ROTATE_EXPIRE:-7d}"
  rotateBackupLimit: ${GF_LOG_ROTATE_BACKUP_LIMIT:-10}
  stStatus: 0

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

# mvp 服务启动前自动执行数据库迁移 + 种子数据（仅 mvp 服务负责，避免并发冲突）
if [[ "${APP_NAME}" == "mvp" ]]; then
  MIGRATE_DIR="/workspace/admin-go/manifest/sql/mysql"
  SEED_FILE="/workspace/admin-go/manifest/sql/seed/mysql_seed.sql"
  MIGRATE_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT:-3306})/${DB_NAME}?multiStatements=true"

  if command -v migrate >/dev/null 2>&1 && [[ -d "${MIGRATE_DIR}" ]]; then
    echo "[migrate] Running database migrations..."
    MIGRATE_OUTPUT=$(migrate -path "${MIGRATE_DIR}" -database "${MIGRATE_URL}" up 2>&1) || true
    MIGRATE_EXIT=${PIPESTATUS[0]:-$?}

    if echo "${MIGRATE_OUTPUT}" | grep -q "no change"; then
      echo "[migrate] No new migrations to apply."
    elif echo "${MIGRATE_OUTPUT}" | grep -qiE "^[0-9]+/"; then
      echo "[migrate] Migrations applied: ${MIGRATE_OUTPUT}"
      SCHEMA_CHANGED=true
    else
      echo "[migrate] ${MIGRATE_OUTPUT}"
    fi

    # 首次建表后自动导入种子数据
    if [[ "${SCHEMA_CHANGED:-}" == "true" ]] && [[ -f "${SEED_FILE}" ]]; then
      echo "[seed] Checking if seed data is needed..."
      cd /workspace/admin-go
      if GF_GCFG_FILE="${CONFIG_FILE}" go run ./app/mvp/tools/dbctl seed -file "${SEED_FILE}" 2>&1; then
        echo "[seed] Seed data applied successfully."
      else
        echo "[seed] Seed skipped or already applied (this is normal for existing databases)."
      fi
    fi
  else
    echo "[migrate] migrate binary or migrations dir not found, skipping."
  fi
fi

cd "/workspace/admin-go/app/${APP_NAME}"
exec go run main.go
