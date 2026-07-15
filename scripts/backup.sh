#!/bin/sh
set -e

# backup.sh — Copy SQLite DB from oxmail-api container to local backups/
# Usage: ./scripts/backup.sh
# Env: OXMAIL_DB_PATH (default: /app/data/oxmail.db)
#      OXMAIL_BACKUP_DIR (default: ./backups)

DB_PATH="${OXMAIL_DB_PATH:-/app/oxmail.db}"
BACKUP_DIR="${OXMAIL_BACKUP_DIR:-./backups}"
DATE=$(date +%Y-%m-%d)

mkdir -p "$BACKUP_DIR"
docker cp "oxmail-api:${DB_PATH}" "$BACKUP_DIR/oxmail-${DATE}.db"
docker cp "oxmail-api:${DB_PATH}-wal" "$BACKUP_DIR/oxmail-${DATE}.db-wal" 2>/dev/null || true
docker cp "oxmail-api:${DB_PATH}-shm" "$BACKUP_DIR/oxmail-${DATE}.db-shm" 2>/dev/null || true
echo "Backup saved: $BACKUP_DIR/oxmail-${DATE}.db"
