#!/bin/sh
# Dovecot quota warning script
# Called by Dovecot when a user exceeds 95% of their storage quota
# Arguments: %u (username/email)

USER="$1"
PERCENT="$2"  # Set by Dovecot when using quota_warning = storage=95%%

if [ -z "$USER" ]; then
  exit 0
fi

# Log to syslog so it appears in container logs
logger -t "quota-warning" "User $USER has exceeded 95% of mailbox quota ($PERCENT% used)"

# If the API health endpoint is reachable, send a notification
# This is a best-effort warning — silent failure is acceptable
ENDPOINT="${QUOTA_WARNING_URL:-http://localhost:8080/api/quota/warning}"
CONTENT="{\"email\":\"$USER\",\"percent\":\"${PERCENT:-95}\"}"

if command -v curl > /dev/null 2>&1; then
  curl -s -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -d "$CONTENT" > /dev/null 2>&1 || true
fi
