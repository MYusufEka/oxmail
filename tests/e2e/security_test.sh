#!/usr/bin/env bash
set -euo pipefail

# security_test.sh — Open relay rejected, auth required (prod mode), rate limiting
# Verifies security boundaries of the mail server.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

suite_start "Security Test"

# ---------------------------------------------------------------------------
# Step 1: Wait for stack healthy
# ---------------------------------------------------------------------------

wait_for_healthy 120

# ---------------------------------------------------------------------------
# Step 2: Open relay test — attempt to relay to external domain
# ---------------------------------------------------------------------------

print_header "Open Relay Protection"

# Attempt to send email to an external domain without being an authorized user
RESPONSE=$(api_post "/api/mail/send" \
  "{\"from\":\"spammer@evil.com\",\"to\":[\"victim@external.org\"],\"subject\":\"Spam\",\"bodyText\":\"Open relay test.\"}")
STATUS=$(extract_status "$RESPONSE")

# Should be rejected (403 or 400 — not 200)
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$STATUS" -ge 400 ]; then
  TESTS_PASSED=$((TESTS_PASSED + 1))
  print_pass "Open relay rejected (HTTP $STATUS)"
else
  TESTS_FAILED=$((TESTS_FAILED + 1))
  print_fail "Open relay NOT rejected (HTTP $STATUS — expected 4xx)"
fi

# Attempt relay from local user to external domain
RESPONSE=$(api_post "/api/mail/send?auth_user=alice@${DOMAIN}" \
  "{\"from\":\"alice@${DOMAIN}\",\"to\":[\"someone@external.org\"],\"subject\":\"Relay attempt\",\"bodyText\":\"Should be blocked in dev mode.\"}")
STATUS=$(extract_status "$RESPONSE")

# External delivery must be blocked/rejected; HTTP 200 means possible relay acceptance.
TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$STATUS" -ge 400 ]; then
  TESTS_PASSED=$((TESTS_PASSED + 1))
  print_pass "External relay blocked/rejected (HTTP $STATUS)"
else
  TESTS_FAILED=$((TESTS_FAILED + 1))
  print_fail "External relay accepted/possible relay (HTTP $STATUS — expected 4xx/5xx)"
fi

# ---------------------------------------------------------------------------
# Step 3: Authentication required in production mode
# ---------------------------------------------------------------------------

print_header "Authentication Enforcement"

# Test unauthenticated access to protected endpoints
PROTECTED_ENDPOINTS=("/api/domains" "/api/users" "/api/mail/inbox?user=alice@${DOMAIN}")

for endpoint in "${PROTECTED_ENDPOINTS[@]}"; do
  # Make request without auth token (clear any existing token)
  SAVED_TOKEN="$AUTH_TOKEN"
  AUTH_TOKEN=""

  RESPONSE=$(api_get "$endpoint")
  STATUS=$(extract_status "$RESPONSE")

  AUTH_TOKEN="$SAVED_TOKEN"

  # In production mode, should return 401
  # In dev mode, may return 200 (auth relaxed) — both are acceptable
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ "$STATUS" = "401" ] || [ "$STATUS" = "200" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "Auth check on ${endpoint} (HTTP $STATUS)"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "Unexpected auth response on ${endpoint} (HTTP $STATUS)"
  fi
done

# If in production mode, verify 401 specifically
if [ "${OXMAIL_MODE:-dev}" = "production" ]; then
  AUTH_TOKEN=""

  RESPONSE=$(api_get "/api/domains")
  STATUS=$(extract_status "$RESPONSE")
  assert_status_code "Unauthenticated GET /api/domains → 401" "401" "$STATUS"

  RESPONSE=$(api_post "/api/domains" "{\"name\":\"unauthorized.test\"}")
  STATUS=$(extract_status "$RESPONSE")
  assert_status_code "Unauthenticated POST /api/domains → 401" "401" "$STATUS"

  RESPONSE=$(api_get "/api/users")
  STATUS=$(extract_status "$RESPONSE")
  assert_status_code "Unauthenticated GET /api/users → 401" "401" "$STATUS"
fi

# ---------------------------------------------------------------------------
# Step 4: Login rate limiting — 6 rapid attempts should trigger 429
# ---------------------------------------------------------------------------

print_header "Rate Limiting"

# Make 6 rapid failed login attempts
RATE_LIMITED=false
for i in $(seq 1 6); do
  RESPONSE=$(api_post "/api/auth/login" "{\"email\":\"admin@${DOMAIN}\",\"password\":\"wrong-password-${i}\"}")
  STATUS=$(extract_status "$RESPONSE")

  if [ "$STATUS" = "429" ]; then
    RATE_LIMITED=true
    print_info "Rate limited after attempt ${i}"
    break
  fi
done

TESTS_TOTAL=$((TESTS_TOTAL + 1))
if [ "$RATE_LIMITED" = true ]; then
  TESTS_PASSED=$((TESTS_PASSED + 1))
  print_pass "Rate limiting triggered after rapid login attempts"
else
  TESTS_FAILED=$((TESTS_FAILED + 1))
  print_fail "Rate limiting NOT triggered after 6 rapid attempts (last status: $STATUS)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

suite_end
