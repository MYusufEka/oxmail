#!/usr/bin/env bash
set -euo pipefail

# lifecycle_test.sh — Full lifecycle: start → seed → send → read → verify logs
# Tests the complete email flow from domain creation to log verification.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

suite_start "Lifecycle Test"

# ---------------------------------------------------------------------------
# Step 1: Wait for stack to be healthy
# ---------------------------------------------------------------------------

wait_for_healthy 120

# ---------------------------------------------------------------------------
# Step 2: Create domain
# ---------------------------------------------------------------------------

RESPONSE=$(api_post "/api/domains" "{\"name\":\"${DOMAIN}\"}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "Create domain ${DOMAIN}" "201" "$STATUS"
assert_json_field "Domain name in response" "$BODY" "name" "$DOMAIN"

# ---------------------------------------------------------------------------
# Step 3: Create users
# ---------------------------------------------------------------------------

RESPONSE=$(api_post "/api/users" "{\"email\":\"alice@${DOMAIN}\",\"password\":\"${PASSWORD}\",\"displayName\":\"Alice Test\"}")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "Create user alice@${DOMAIN}" "201" "$STATUS"

RESPONSE=$(api_post "/api/users" "{\"email\":\"bob@${DOMAIN}\",\"password\":\"${PASSWORD}\",\"displayName\":\"Bob Test\"}")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "Create user bob@${DOMAIN}" "201" "$STATUS"

# ---------------------------------------------------------------------------
# Step 4: Send email from alice to bob
# ---------------------------------------------------------------------------

RESPONSE=$(api_post "/api/mail/send?auth_user=alice@${DOMAIN}" \
  "{\"from\":\"alice@${DOMAIN}\",\"to\":[\"bob@${DOMAIN}\"],\"subject\":\"Lifecycle Test Email\",\"bodyText\":\"Hello from lifecycle test.\"}")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "Send email alice → bob" "200" "$STATUS"

# Brief pause for delivery
sleep 2

# ---------------------------------------------------------------------------
# Step 5: Verify email appears in bob's inbox
# ---------------------------------------------------------------------------

RESPONSE=$(api_get "/api/mail/inbox?user=bob@${DOMAIN}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "Get bob's inbox" "200" "$STATUS"
assert_contains "Email in bob's inbox" "$BODY" "Lifecycle Test Email"

# ---------------------------------------------------------------------------
# Step 6: Verify log entry for the sent email
# ---------------------------------------------------------------------------

RESPONSE=$(api_get "/api/logs?limit=50")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "Get logs" "200" "$STATUS"
assert_contains "Log contains send event" "$BODY" "alice@${DOMAIN}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

suite_end
