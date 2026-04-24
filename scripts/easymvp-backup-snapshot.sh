#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CORE_DIR="$ROOT_DIR/apps/core"
DATA_ROOT="${EASYMVP_DATA_ROOT:-$CORE_DIR/var}"
DB_PATH="${EASYMVP_DB_PATH:-$DATA_ROOT/data/easymvp.db}"
MIGRATION_PATH="${EASYMVP_MIGRATION_PATH:-$CORE_DIR/manifest/migrations}"
BACKUP_ROOT="${EASYMVP_BACKUP_ROOT:-$DATA_ROOT/backups}"

usage() {
  cat <<'EOF'
Usage:
  scripts/easymvp-backup-snapshot.sh snapshot [label]
  scripts/easymvp-backup-snapshot.sh verify <snapshot_dir>
  scripts/easymvp-backup-snapshot.sh restore <snapshot_dir> <target_data_root>

Environment:
  EASYMVP_DATA_ROOT              defaults to apps/core/var
  EASYMVP_DB_PATH                defaults to $EASYMVP_DATA_ROOT/data/easymvp.db
  EASYMVP_MIGRATION_PATH         defaults to apps/core/manifest/migrations
  EASYMVP_BACKUP_ROOT            defaults to $EASYMVP_DATA_ROOT/backups
  EASYMVP_BACKUP_INCLUDE_FILES=1 optionally copies projects/settings/diagnostics
  EASYMVP_RESTORE_OVERWRITE=1    allows overwriting target_data_root/data/easymvp.db
EOF
}

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

sanitize_label() {
  local label="${1:-manual}"
  printf "%s" "$label" | tr -c 'A-Za-z0-9._-' '-' | sed 's/^-*//;s/-*$//' | cut -c 1-80
}

copy_dir_if_exists() {
  local src="$1"
  local dst="$2"
  [ -d "$src" ] || return 0

  mkdir -p "$(dirname "$dst")"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a "$src/" "$dst/"
  else
    mkdir -p "$dst"
    cp -a "$src/." "$dst/"
  fi
}

copy_sqlite_db() {
  local source="$1"
  local target="$2"

  mkdir -p "$(dirname "$target")"
  if command -v sqlite3 >/dev/null 2>&1; then
    local escaped_target
    escaped_target="$(printf "%s" "$target" | sed "s/'/''/g")"
    if sqlite3 "$source" ".backup '$escaped_target'"; then
      return 0
    fi
    echo "WARN: sqlite3 .backup failed; falling back to file copy with WAL sidecars" >&2
  fi

  cp -p "$source" "$target"
  for suffix in -wal -shm; do
    if [ -f "$source$suffix" ]; then
      cp -p "$source$suffix" "$target$suffix"
    fi
  done
}

verify_snapshot() {
  local snapshot_dir="${1:-}"
  [ -n "$snapshot_dir" ] || fail "snapshot_dir is required"
  [ -d "$snapshot_dir" ] || fail "snapshot_dir not found: $snapshot_dir"
  [ -f "$snapshot_dir/db/easymvp.db" ] || fail "snapshot db missing: $snapshot_dir/db/easymvp.db"
  [ -f "$snapshot_dir/manifest/snapshot.env" ] || fail "snapshot manifest missing"
  [ -d "$snapshot_dir/manifest/migrations" ] || fail "migration manifest missing"

  if command -v sqlite3 >/dev/null 2>&1; then
    local integrity
    integrity="$(sqlite3 "$snapshot_dir/db/easymvp.db" 'PRAGMA integrity_check;' | tr -d '\r')"
    [ "$integrity" = "ok" ] || fail "sqlite integrity_check failed: $integrity"
  else
    echo "WARN: sqlite3 not found; skipped sqlite integrity_check" >&2
  fi

  echo "EASYMVP_BACKUP_VERIFY path=$snapshot_dir status=ok"
}

create_snapshot() {
  local raw_label="${1:-manual}"
  local label
  local timestamp
  local snapshot_dir

  label="$(sanitize_label "$raw_label")"
  [ -n "$label" ] || label="manual"
  timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
  snapshot_dir="$BACKUP_ROOT/${timestamp}-${label}"

  [ -f "$DB_PATH" ] || fail "db file not found: $DB_PATH"
  [ -d "$MIGRATION_PATH" ] || fail "migration path not found: $MIGRATION_PATH"
  [ ! -e "$snapshot_dir" ] || fail "snapshot already exists: $snapshot_dir"

  mkdir -p "$snapshot_dir/db" "$snapshot_dir/manifest" "$snapshot_dir/metadata"
  copy_sqlite_db "$DB_PATH" "$snapshot_dir/db/easymvp.db"
  mkdir -p "$snapshot_dir/manifest/migrations"
  cp -p "$MIGRATION_PATH"/*.sql "$snapshot_dir/manifest/migrations/"

  if [ "${EASYMVP_BACKUP_INCLUDE_FILES:-0}" = "1" ]; then
    copy_dir_if_exists "$DATA_ROOT/projects" "$snapshot_dir/data_root/projects"
    copy_dir_if_exists "$DATA_ROOT/settings" "$snapshot_dir/data_root/settings"
    copy_dir_if_exists "$DATA_ROOT/diagnostics" "$snapshot_dir/data_root/diagnostics"
  fi

  cat >"$snapshot_dir/manifest/snapshot.env" <<EOF
created_at_utc=$timestamp
label=$label
root_dir=$ROOT_DIR
data_root=$DATA_ROOT
db_path=$DB_PATH
migration_path=$MIGRATION_PATH
include_files=${EASYMVP_BACKUP_INCLUDE_FILES:-0}
EOF

  {
    echo "snapshot_dir=$snapshot_dir"
    echo "db_bytes=$(wc -c <"$snapshot_dir/db/easymvp.db" | tr -d ' ')"
    echo "migration_count=$(find "$snapshot_dir/manifest/migrations" -maxdepth 1 -type f -name '*.sql' | wc -l | tr -d ' ')"
  } >"$snapshot_dir/metadata/summary.txt"

  verify_snapshot "$snapshot_dir" >/dev/null
  echo "EASYMVP_BACKUP path=$snapshot_dir db=$snapshot_dir/db/easymvp.db include_files=${EASYMVP_BACKUP_INCLUDE_FILES:-0}"
}

restore_snapshot() {
  local snapshot_dir="${1:-}"
  local target_data_root="${2:-}"
  local target_db

  [ -n "$snapshot_dir" ] || fail "snapshot_dir is required"
  [ -n "$target_data_root" ] || fail "target_data_root is required"
  verify_snapshot "$snapshot_dir" >/dev/null

  target_db="$target_data_root/data/easymvp.db"
  if [ -e "$target_db" ] && [ "${EASYMVP_RESTORE_OVERWRITE:-0}" != "1" ]; then
    fail "target db already exists; set EASYMVP_RESTORE_OVERWRITE=1 to overwrite: $target_db"
  fi

  mkdir -p "$target_data_root/data" "$target_data_root/backups" "$target_data_root/temp"
  cp -p "$snapshot_dir/db/easymvp.db" "$target_db"
  for suffix in -wal -shm; do
    if [ -f "$snapshot_dir/db/easymvp.db$suffix" ]; then
      cp -p "$snapshot_dir/db/easymvp.db$suffix" "$target_db$suffix"
    fi
  done

  copy_dir_if_exists "$snapshot_dir/data_root/projects" "$target_data_root/projects"
  copy_dir_if_exists "$snapshot_dir/data_root/settings" "$target_data_root/settings"
  copy_dir_if_exists "$snapshot_dir/data_root/diagnostics" "$target_data_root/diagnostics"

  echo "EASYMVP_RESTORE data_root=$target_data_root db=$target_db status=ok"
}

case "${1:-snapshot}" in
  snapshot|create)
    shift || true
    create_snapshot "${1:-manual}"
    ;;
  verify)
    shift || true
    verify_snapshot "${1:-}"
    ;;
  restore)
    shift || true
    restore_snapshot "${1:-}" "${2:-}"
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    usage >&2
    fail "unknown command: $1"
    ;;
esac
