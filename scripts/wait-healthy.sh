#!/usr/bin/env bash
set -euo pipefail

# wait-healthy.sh — Poll API health endpoint until ready
# Usage: ./scripts/wait-healthy.sh
# Env: OXMAIL_API_URL (default: http://localhost:8080)
#       TIMEOUT (default: 120 seconds)

API_URL="${OXMAIL_API_URL:-http://localhost:8080}"
TIMEOUT="${TIMEOUT:-120}"
INTERVAL=2
ELAPSED=0

echo "Waiting for API at ${API_URL}/health ..."

while [ "$ELAPSED" -lt "$TIMEOUT" ]; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_URL}/health" 2>/dev/null || echo "000")

  if [ "$STATUS" = "200" ]; then
    echo ""
    echo "API is healthy! (took ${ELAPSED}s)"
    exit 0
  fi

  printf "."
  sleep "$INTERVAL"
  ELAPSED=$((ELAPSED + INTERVAL))
done

echo ""
echo "ERROR: API did not become healthy within ${TIMEOUT}s"
exit 1
