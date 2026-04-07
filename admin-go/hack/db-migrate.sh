#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$ROOT_DIR/manifest/sql/mysql}"
MIGRATE_BIN="${MIGRATE_BIN:-migrate}"

export GF_GCFG_PATH="${GF_GCFG_PATH:-app/mvp/manifest/config}"
export GF_GCFG_FILE="${GF_GCFG_FILE:-config.yaml}"

usage() {
  cat <<'EOF'
Usage:
  ./hack/db-migrate.sh create <name>
  ./hack/db-migrate.sh up [N]
  ./hack/db-migrate.sh down [N]
  ./hack/db-migrate.sh version
  ./hack/db-migrate.sh goto <version>
  ./hack/db-migrate.sh force <version>

Environment:
  MIGRATE_BIN            Override migrate binary path. Default: migrate
  MIGRATE_DATABASE_URL   Override database URL. If empty, derive from GoFrame config.
  MIGRATIONS_DIR         Override migration directory. Default: manifest/sql/mysql

Notes:
  1. This project uses versioned SQL migrations for schema only.
  2. Bootstrap seed data is applied separately via: make db-seed
EOF
}

require_migrate() {
  if command -v "$MIGRATE_BIN" >/dev/null 2>&1; then
    return 0
  fi

  cat >&2 <<'EOF'
error: migrate binary not found.

Install one of:
  go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  brew install golang-migrate

Then re-run the command.
EOF
  exit 1
}

database_url() {
  if [[ -n "${MIGRATE_DATABASE_URL:-}" ]]; then
    printf '%s\n' "$MIGRATE_DATABASE_URL"
    return 0
  fi

  (
    cd "$ROOT_DIR"
    GF_GCFG_PATH="$GF_GCFG_PATH" GF_GCFG_FILE="$GF_GCFG_FILE" \
      go run ./app/mvp/tools/dbctl migrate-url
  )
}

run_migrate() {
  local db_url
  db_url="$(database_url)"
  exec "$MIGRATE_BIN" -path "$MIGRATIONS_DIR" -database "$db_url" "$@"
}

main() {
  local cmd="${1:-}"
  if [[ -z "$cmd" ]]; then
    usage
    exit 1
  fi

  case "$cmd" in
    create)
      require_migrate
      local name="${2:-}"
      if [[ -z "$name" ]]; then
        echo "error: migration name is required" >&2
        exit 1
      fi
      exec "$MIGRATE_BIN" create -ext sql -dir "$MIGRATIONS_DIR" -seq "$name"
      ;;
    up)
      require_migrate
      shift
      run_migrate up "$@"
      ;;
    down)
      require_migrate
      shift
      if [[ "$#" -eq 0 ]]; then
        run_migrate down 1
      else
        run_migrate down "$@"
      fi
      ;;
    version)
      require_migrate
      run_migrate version
      ;;
    goto|force)
      require_migrate
      local version="${2:-}"
      if [[ -z "$version" ]]; then
        echo "error: version is required" >&2
        exit 1
      fi
      run_migrate "$cmd" "$version"
      ;;
    -h|--help|help)
      usage
      ;;
    *)
      echo "error: unknown command: $cmd" >&2
      usage
      exit 1
      ;;
  esac
}

main "$@"
