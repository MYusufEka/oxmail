#!/bin/sh
set -e

# restore.sh — Safely restore one SQLite backup after stopping oxmail-api.
# Usage: ./scripts/restore.sh backups/oxmail-YYYY-MM-DD-HHMMSS.db
# Env: DB_PATH or OXMAIL_DB_PATH (default: oxmail.db)
#      COMPOSE_FILES (default: -f docker-compose.yml)
#      COMPOSE_PROFILES (optional, example: --profile prod)
#      SKIP_DOCKER=1 to skip container stop/start for local temp DB tests

log_failure() {
  echo "[restore] FAILED: $1" >&2
}

log_success() {
  echo "[restore] success"
}

fail() {
  log_failure "$1"
  exit 1
}

usage() {
  echo "Usage: $0 BACKUP_FILE" >&2
}

[ "$#" -eq 1 ] || { usage; fail "expected one backup file argument"; }

BACKUP_PATH="$1"
DB_PATH="${DB_PATH:-${OXMAIL_DB_PATH:-oxmail.db}}"
COMPOSE_FILES="${COMPOSE_FILES:--f docker-compose.yml}"
COMPOSE_PROFILES="${COMPOSE_PROFILES:-}"
SKIP_DOCKER="${SKIP_DOCKER:-0}"
API_WAS_RUNNING=0

command -v sqlite3 >/dev/null 2>&1 || fail "sqlite3 CLI not found"
[ -f "$BACKUP_PATH" ] || fail "backup file not found: $BACKUP_PATH"

BACKUP_CHECK=$(sqlite3 "$BACKUP_PATH" "PRAGMA integrity_check;" 2>&1) || fail "backup integrity check failed: $BACKUP_CHECK"
[ "$BACKUP_CHECK" = "ok" ] || fail "backup integrity check returned: $BACKUP_CHECK"

restore_api() {
  if [ "$SKIP_DOCKER" != "1" ] && [ "$API_WAS_RUNNING" -eq 1 ]; then
    # shellcheck disable=SC2086
    docker compose $COMPOSE_PROFILES $COMPOSE_FILES up -d oxmail-api >/dev/null 2>&1 || true
  fi
}
trap restore_api EXIT INT TERM

if [ "$SKIP_DOCKER" != "1" ]; then
  command -v docker >/dev/null 2>&1 || fail "docker CLI not found"
  if docker compose $COMPOSE_PROFILES $COMPOSE_FILES ps oxmail-api >/dev/null 2>&1; then
    RUNNING_STATE=$(docker compose $COMPOSE_PROFILES $COMPOSE_FILES ps --status running --services oxmail-api 2>/dev/null || true)
    if [ "$RUNNING_STATE" = "oxmail-api" ]; then
      API_WAS_RUNNING=1
      docker compose $COMPOSE_PROFILES $COMPOSE_FILES stop oxmail-api >/dev/null || fail "could not stop oxmail-api"
    fi
  fi
fi

DB_DIR=$(dirname "$DB_PATH")
mkdir -p "$DB_DIR" || fail "could not create database directory: $DB_DIR"
TEMP_RESTORE=$(mktemp "${DB_DIR}/.restore.XXXXXX") || fail "could not create temporary restore file"
rm -f "$TEMP_RESTORE"

sqlite3 "$BACKUP_PATH" ".backup '$TEMP_RESTORE'" || {
  rm -f "$TEMP_RESTORE"
  fail "sqlite restore backup copy failed"
}

RESTORED_CHECK=$(sqlite3 "$TEMP_RESTORE" "PRAGMA integrity_check;" 2>&1) || {
  rm -f "$TEMP_RESTORE"
  fail "restored database integrity check failed: $RESTORED_CHECK"
}
[ "$RESTORED_CHECK" = "ok" ] || {
  rm -f "$TEMP_RESTORE"
  fail "restored database integrity check returned: $RESTORED_CHECK"
}

mv "$TEMP_RESTORE" "$DB_PATH" || {
  rm -f "$TEMP_RESTORE"
  fail "could not replace database: $DB_PATH"
}
rm -f "$DB_PATH-wal" "$DB_PATH-shm"

if [ "$SKIP_DOCKER" != "1" ] && [ "$API_WAS_RUNNING" -eq 1 ]; then
  docker compose $COMPOSE_PROFILES $COMPOSE_FILES up -d oxmail-api >/dev/null || fail "could not restart oxmail-api"
fi
trap - EXIT INT TERM
log_success
