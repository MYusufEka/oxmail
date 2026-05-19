#!/usr/bin/env bash
set -euo pipefail

# helpers.sh — Shared test utilities for E2E tests
# Source this file: source "$(dirname "${BASH_SOURCE[0]}")/helpers.sh"

# =============================================================================
# Configuration
# =============================================================================

API_URL="${OXMAIL_API_URL:-http://localhost:8080}"
DOMAIN="local.test"
PASSWORD="TestPass123!"
AUTH_TOKEN=""

# =============================================================================
# Colors
# =============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# =============================================================================
# Counters
# =============================================================================

TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0
TEST_START_TIME=""
CURRENT_SUITE=""

# =============================================================================
# Output helpers
# =============================================================================

print_pass() { echo -e "  ${GREEN}✓${NC} $1"; }
print_fail() { echo -e "  ${RED}✗${NC} $1"; }
print_info() { echo -e "  ${CYAN}ℹ${NC} $1"; }
print_header() {
  echo ""
  echo -e "${BOLD}━━━ $1 ━━━${NC}"
  echo ""
}

# =============================================================================
# Test lifecycle
# =============================================================================

suite_start() {
  CURRENT_SUITE="$1"
  TEST_START_TIME=$(date +%s)
  TESTS_PASSED=0
  TESTS_FAILED=0
  TESTS_TOTAL=0
  print_header "$CURRENT_SUITE"
}

suite_end() {
  local end_time
  end_time=$(date +%s)
  local duration=$((end_time - TEST_START_TIME))
  echo ""
  echo -e "${BOLD}Results: ${GREEN}${TESTS_PASSED} passed${NC}, ${RED}${TESTS_FAILED} failed${NC} (${TESTS_TOTAL} total) in ${duration}s"
  if [ "$TESTS_FAILED" -gt 0 ]; then
    return 1
  fi
  return 0
}

# =============================================================================
# Assertion helpers
# =============================================================================

assert_status_code() {
  local description="$1"
  local expected="$2"
  local actual="$3"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ "$actual" = "$expected" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "$description (HTTP $actual)"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "$description (expected HTTP $expected, got HTTP $actual)"
  fi
}

assert_json_field() {
  local description="$1"
  local json="$2"
  local field="$3"
  local expected="$4"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  local actual
  actual=$(echo "$json" | grep -o "\"${field}\":[^,}]*" | head -1 | cut -d: -f2- | tr -d ' "')

  if [ "$actual" = "$expected" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "$description (.${field} = \"$actual\")"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "$description (expected .${field} = \"$expected\", got \"$actual\")"
  fi
}

assert_contains() {
  local description="$1"
  local haystack="$2"
  local needle="$3"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$haystack" | grep -q "$needle"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "$description (contains \"$needle\")"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "$description (does not contain \"$needle\")"
  fi
}

assert_not_empty() {
  local description="$1"
  local value="$2"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ -n "$value" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "$description (not empty)"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "$description (was empty)"
  fi
}

assert_greater_than() {
  local description="$1"
  local actual="$2"
  local threshold="$3"

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ "$actual" -gt "$threshold" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "$description ($actual > $threshold)"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "$description (expected > $threshold, got $actual)"
  fi
}

# =============================================================================
# HTTP helpers
# =============================================================================

api_get() {
  local endpoint="$1"
  local auth_header=""
  if [ -n "$AUTH_TOKEN" ]; then
    auth_header="-H \"Authorization: Bearer ${AUTH_TOKEN}\""
  fi
  eval curl -s -w "\n%{http_code}" ${auth_header} "${API_URL}${endpoint}"
}

api_post() {
  local endpoint="$1"
  local payload="${2:-{}}"
  local auth_header=""
  if [ -n "$AUTH_TOKEN" ]; then
    auth_header="-H \"Authorization: Bearer ${AUTH_TOKEN}\""
  fi
  eval curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    ${auth_header} \
    -d "'${payload}'" \
    "${API_URL}${endpoint}"
}

api_delete() {
  local endpoint="$1"
  local auth_header=""
  if [ -n "$AUTH_TOKEN" ]; then
    auth_header="-H \"Authorization: Bearer ${AUTH_TOKEN}\""
  fi
  eval curl -s -w "\n%{http_code}" -X DELETE ${auth_header} "${API_URL}${endpoint}"
}

# Extract HTTP status code from curl response (last line)
extract_status() {
  echo "$1" | tail -n1
}

# Extract body from curl response (everything except last line)
extract_body() {
  echo "$1" | sed '$d'
}

# =============================================================================
# Auth helpers
# =============================================================================

login() {
  local email="${1:-admin@${DOMAIN}}"
  local password="${2:-${PASSWORD}}"
  local response
  response=$(api_post "/api/auth/login" "{\"email\":\"${email}\",\"password\":\"${password}\"}")
  local body
  body=$(extract_body "$response")
  AUTH_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
}

# =============================================================================
# Wait helpers
# =============================================================================

wait_for_healthy() {
  local timeout="${1:-120}"
  local interval=2
  local elapsed=0

  print_info "Waiting for API at ${API_URL}/health ..."

  while [ "$elapsed" -lt "$timeout" ]; do
    local status
    status=$(curl -s -o /dev/null -w "%{http_code}" "${API_URL}/health" 2>/dev/null || echo "000")
    if [ "$status" = "200" ]; then
      print_info "API healthy (${elapsed}s)"
      return 0
    fi
    sleep "$interval"
    elapsed=$((elapsed + interval))
  done

  print_fail "API not healthy after ${timeout}s"
  return 1
}
