#!/bin/sh
set -e

# backup.sh — Create WAL-safe SQLite backup with integrity verification and retention.
# Usage: ./scripts/backup.sh
# Env: DB_PATH or OXMAIL_DB_PATH (default: oxmail.db)
#      BACKUP_DIR or OXMAIL_BACKUP_DIR (default: ./backups)
#      BACKUP_RETAIN_DAYS (default: 7)

log_failure() {
  echo "[backup] FAILED: $1" >&2
}

log_success() {
  echo "[backup] success: $1"
}

cleanup_failed_backup() {
  if [ -n "${BACKUP_PATH:-}" ] && [ -f "$BACKUP_PATH" ]; then
    rm -f "$BACKUP_PATH"
  fi
}

fail() {
  log_failure "$1"
  cleanup_failed_backup
  exit 1
}

DB_PATH="${DB_PATH:-${OXMAIL_DB_PATH:-oxmail.db}}"
BACKUP_DIR="${BACKUP_DIR:-${OXMAIL_BACKUP_DIR:-./backups}}"
BACKUP_RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"
TIMESTAMP=$(date +%Y-%m-%d-%H%M%S)
BACKUP_PATH="$BACKUP_DIR/oxmail-${TIMESTAMP}.db"

command -v sqlite3 >/dev/null 2>&1 || fail "sqlite3 CLI not found"
[ -f "$DB_PATH" ] || fail "database not found: $DB_PATH"

case "$BACKUP_RETAIN_DAYS" in
  ''|*[!0-9]*) fail "BACKUP_RETAIN_DAYS must be non-negative integer" ;;
esac

mkdir -p "$BACKUP_DIR" || fail "could not create backup directory: $BACKUP_DIR"

SOURCE_CHECK=$(sqlite3 "$DB_PATH" "PRAGMA integrity_check;" 2>&1) || fail "source integrity check failed: $SOURCE_CHECK"
[ "$SOURCE_CHECK" = "ok" ] || fail "source integrity check returned: $SOURCE_CHECK"

sqlite3 "$DB_PATH" ".backup '$BACKUP_PATH'" || fail "sqlite .backup failed: $BACKUP_PATH"

BACKUP_CHECK=$(sqlite3 "$BACKUP_PATH" "PRAGMA integrity_check;" 2>&1) || fail "backup integrity check failed: $BACKUP_CHECK"
[ "$BACKUP_CHECK" = "ok" ] || fail "backup integrity check returned: $BACKUP_CHECK"

if [ "$BACKUP_RETAIN_DAYS" -gt 0 ]; then
  find "$BACKUP_DIR" -type f -name 'oxmail-*.db' -mtime +"$BACKUP_RETAIN_DAYS" -exec rm -f {} \; || fail "retention cleanup failed"
fi

log_success "$BACKUP_PATH"
