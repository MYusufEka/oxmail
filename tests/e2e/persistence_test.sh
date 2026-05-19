#!/usr/bin/env bash
set -euo pipefail

# persistence_test.sh — Send email → restart → email still accessible
# Verifies data survives container restarts.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

suite_start "Persistence Test"

# ---------------------------------------------------------------------------
# Step 1: Wait for stack healthy
# ---------------------------------------------------------------------------

wait_for_healthy 120

# ---------------------------------------------------------------------------
# Step 2: Seed a domain and user if not already present
# ---------------------------------------------------------------------------

api_post "/api/domains" "{\"name\":\"${DOMAIN}\"}" >/dev/null 2>&1 || true
api_post "/api/users" "{\"email\":\"persist@${DOMAIN}\",\"password\":\"${PASSWORD}\",\"displayName\":\"Persist User\"}" >/dev/null 2>&1 || true

# ---------------------------------------------------------------------------
# Step 3: Send a uniquely identifiable email
# ---------------------------------------------------------------------------

UNIQUE_SUBJECT="Persistence-$(date +%s)"

RESPONSE=$(api_post "/api/mail/send?auth_user=persist@${DOMAIN}" \
  "{\"from\":\"persist@${DOMAIN}\",\"to\":[\"persist@${DOMAIN}\"],\"subject\":\"${UNIQUE_SUBJECT}\",\"bodyText\":\"This email must survive a restart.\"}")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "Send persistence test email" "200" "$STATUS"

sleep 2

# ---------------------------------------------------------------------------
# Step 4: Verify email is in inbox before restart
# ---------------------------------------------------------------------------

RESPONSE=$(api_get "/api/mail/inbox?user=persist@${DOMAIN}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "Inbox accessible before restart" "200" "$STATUS"
assert_contains "Email present before restart" "$BODY" "$UNIQUE_SUBJECT"

# Count messages before restart
INBOX_COUNT_BEFORE=$(echo "$BODY" | grep -o "\"subject\"" | wc -l | tr -d ' ')
print_info "Inbox count before restart: ${INBOX_COUNT_BEFORE}"

# ---------------------------------------------------------------------------
# Step 5: Restart the stack
# ---------------------------------------------------------------------------

print_info "Restarting docker compose..."
docker compose restart

# ---------------------------------------------------------------------------
# Step 6: Wait for stack to come back healthy
# ---------------------------------------------------------------------------

wait_for_healthy 120

# ---------------------------------------------------------------------------
# Step 7: Verify email still accessible after restart
# ---------------------------------------------------------------------------

RESPONSE=$(api_get "/api/mail/inbox?user=persist@${DOMAIN}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "Inbox accessible after restart" "200" "$STATUS"
assert_contains "Email present after restart" "$BODY" "$UNIQUE_SUBJECT"

# Count messages after restart
INBOX_COUNT_AFTER=$(echo "$BODY" | grep -o "\"subject\"" | wc -l | tr -d ' ')
print_info "Inbox count after restart: ${INBOX_COUNT_AFTER}"

# Verify count unchanged
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$INBOX_COUNT_AFTER" -ge "$INBOX_COUNT_BEFORE" ]; then
  TESTS_PASSED=$((TESTS_PASSED + 1))
  print_pass "Inbox count preserved (${INBOX_COUNT_BEFORE} → ${INBOX_COUNT_AFTER})"
else
  TESTS_FAILED=$((TESTS_FAILED + 1))
  print_fail "Inbox count changed (${INBOX_COUNT_BEFORE} → ${INBOX_COUNT_AFTER})"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

suite_end
