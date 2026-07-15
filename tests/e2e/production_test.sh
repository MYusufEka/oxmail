#!/usr/bin/env bash
set -euo pipefail

# production_test.sh — Production-specific E2E tests
# Tests TLS, SMTP STARTTLS, IMAPS, auth enforcement, and HTTP→HTTPS redirect.
#
# Usage: ./tests/e2e/production_test.sh [domain]
# Requires: openssl, curl

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

# =============================================================================
# Configuration
# =============================================================================

PROD_DOMAIN="${1:-${OXMAIL_DOMAIN:-mail.example.com}}"
PROD_IP="${OXMAIL_PUBLIC_IP:-127.0.0.1}"
TIMEOUT=10

# Override API_URL for production
API_URL="https://${PROD_DOMAIN}"

# =============================================================================
# Prerequisites check
# =============================================================================

check_prerequisites() {
  local missing=()

  if ! command -v openssl &>/dev/null; then
    missing+=("openssl")
  fi
  if ! command -v curl &>/dev/null; then
    missing+=("curl")
  fi

  if [ ${#missing[@]} -gt 0 ]; then
    echo -e "${RED}Missing required tools: ${missing[*]}${NC}"
    echo "Install them and retry."
    exit 1
  fi
}

# =============================================================================
# TLS Tests (Port 443)
# =============================================================================

test_https_tls() {
  suite_start "HTTPS TLS (port 443)"

  # Test TLS connection
  local tls_output
  tls_output=$(echo | openssl s_client -connect "${PROD_DOMAIN}:443" -servername "${PROD_DOMAIN}" 2>&1 || true)

  # Verify certificate is present
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$tls_output" | grep -q "BEGIN CERTIFICATE"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "TLS certificate present on port 443"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "No TLS certificate on port 443"
  fi

  # Verify certificate matches domain
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  local cert_cn
  cert_cn=$(echo "$tls_output" | openssl x509 -noout -subject 2>/dev/null | grep -o "CN = [^/]*" | cut -d= -f2 | tr -d ' ' || echo "")
  local cert_san
  cert_san=$(echo "$tls_output" | openssl x509 -noout -ext subjectAltName 2>/dev/null || echo "")

  if echo "$cert_cn $cert_san" | grep -q "${PROD_DOMAIN}"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "Certificate matches domain ${PROD_DOMAIN}"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "Certificate does not match domain (CN=${cert_cn})"
  fi

  # Verify TLS version (should be 1.2 or 1.3)
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$tls_output" | grep -qE "TLSv1\.[23]"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    local tls_version
    tls_version=$(echo "$tls_output" | grep -oE "TLSv1\.[23]" | head -1)
    print_pass "TLS version is ${tls_version}"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "TLS version is not 1.2 or 1.3"
  fi

  suite_end
}

# =============================================================================
# SMTP STARTTLS Tests (Port 587)
# =============================================================================

test_smtp_starttls() {
  suite_start "SMTP STARTTLS (port 587)"

  # Test SMTP banner
  local smtp_output
  smtp_output=$(echo "EHLO test.local" | openssl s_client -connect "${PROD_DOMAIN}:587" -starttls smtp -quiet 2>&1 || true)

  # Verify STARTTLS negotiation succeeded
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$smtp_output" | grep -qE "(250|220)"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "SMTP STARTTLS connection established on port 587"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "SMTP STARTTLS failed on port 587"
  fi

  # Verify TLS certificate on SMTP
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  local smtp_tls
  smtp_tls=$(echo "" | openssl s_client -connect "${PROD_DOMAIN}:587" -starttls smtp 2>&1 || true)
  if echo "$smtp_tls" | grep -q "BEGIN CERTIFICATE"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "SMTP TLS certificate present"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "No TLS certificate on SMTP port 587"
  fi

  # Test SMTPS on port 465 (implicit TLS)
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  local smtps_output
  smtps_output=$(echo "" | openssl s_client -connect "${PROD_DOMAIN}:465" 2>&1 || true)
  if echo "$smtps_output" | grep -q "BEGIN CERTIFICATE"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "SMTPS (implicit TLS) working on port 465"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "SMTPS not available on port 465"
  fi

  suite_end
}

# =============================================================================
# IMAPS Tests (Port 993)
# =============================================================================

test_imaps() {
  suite_start "IMAPS (port 993)"

  # Test IMAPS connection (implicit TLS)
  local imaps_output
  imaps_output=$(echo "" | openssl s_client -connect "${PROD_DOMAIN}:993" 2>&1 || true)

  # Verify TLS certificate
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$imaps_output" | grep -q "BEGIN CERTIFICATE"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "IMAPS TLS certificate present on port 993"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "No TLS certificate on IMAPS port 993"
  fi

  # Verify IMAP banner after TLS
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  local imap_banner
  imap_banner=$(echo "a001 LOGOUT" | openssl s_client -connect "${PROD_DOMAIN}:993" -quiet 2>/dev/null || true)
  if echo "$imap_banner" | grep -qi "OK\|IMAP\|Dovecot"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "IMAP server responding over TLS"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "IMAP server not responding on port 993"
  fi

  # Verify TLS version on IMAPS
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$imaps_output" | grep -qE "TLSv1\.[23]"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    local tls_ver
    tls_ver=$(echo "$imaps_output" | grep -oE "TLSv1\.[23]" | head -1)
    print_pass "IMAPS using ${tls_ver}"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "IMAPS not using TLS 1.2+"
  fi

  suite_end
}

# =============================================================================
# Auth Enforcement Tests
# =============================================================================

test_auth_enforcement() {
  suite_start "Auth Enforcement"

  # Test unauthenticated access to protected endpoint → 401
  local response
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    -k "https://${PROD_DOMAIN}/api/domains" 2>/dev/null || echo "000")

  assert_status_code "GET /api/domains without auth returns 401" "401" "$response"

  # Test unauthenticated access to users endpoint → 401
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    -k "https://${PROD_DOMAIN}/api/users" 2>/dev/null || echo "000")

  assert_status_code "GET /api/users without auth returns 401" "401" "$response"

  # Test health endpoint is public (no auth required)
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    -k "https://${PROD_DOMAIN}/api/health" 2>/dev/null || echo "000")

  assert_status_code "GET /api/health is public (200)" "200" "$response"

  # Test invalid token → 401
  response=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    -k -H "Authorization: Bearer invalid-token-here" \
    "https://${PROD_DOMAIN}/api/domains" 2>/dev/null || echo "000")

  assert_status_code "GET /api/domains with invalid token returns 401" "401" "$response"

  suite_end
}

# =============================================================================
# HTTP → HTTPS Redirect Tests
# =============================================================================

test_http_redirect() {
  suite_start "HTTP → HTTPS Redirect"

  # Test HTTP redirect to HTTPS
  local redirect_status
  redirect_status=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    "http://${PROD_DOMAIN}/" 2>/dev/null || echo "000")

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ "$redirect_status" = "301" ] || [ "$redirect_status" = "308" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "HTTP → HTTPS redirect (${redirect_status})"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "Expected 301/308 redirect, got ${redirect_status}"
  fi

  # Verify redirect location header points to HTTPS
  local redirect_location
  redirect_location=$(curl -s -I --max-time "$TIMEOUT" \
    "http://${PROD_DOMAIN}/" 2>/dev/null | grep -i "^location:" | tr -d '\r' || echo "")

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if echo "$redirect_location" | grep -qi "https://"; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "Redirect location uses HTTPS"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "Redirect location does not use HTTPS (${redirect_location})"
  fi

  # Test API path also redirects
  local api_redirect
  api_redirect=$(curl -s -o /dev/null -w "%{http_code}" \
    --max-time "$TIMEOUT" \
    "http://${PROD_DOMAIN}/api/health" 2>/dev/null || echo "000")

  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  if [ "$api_redirect" = "301" ] || [ "$api_redirect" = "308" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    print_pass "API path HTTP → HTTPS redirect (${api_redirect})"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    print_fail "API path not redirecting (got ${api_redirect})"
  fi

  suite_end
}

# =============================================================================
# Main
# =============================================================================

main() {
  check_prerequisites

  echo ""
  echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}║   Oxmail Production Integration Tests       ║${NC}"
  echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
  echo ""
  echo -e "  Domain:  ${CYAN}${PROD_DOMAIN}${NC}"
  echo -e "  IP:      ${CYAN}${PROD_IP}${NC}"
  echo ""

  local total_passed=0
  local total_failed=0
  local suites_failed=0

  # Run all test suites
  if test_https_tls; then
    total_passed=$((total_passed + TESTS_PASSED))
  else
    total_passed=$((total_passed + TESTS_PASSED))
    total_failed=$((total_failed + TESTS_FAILED))
    suites_failed=$((suites_failed + 1))
  fi

  if test_smtp_starttls; then
    total_passed=$((total_passed + TESTS_PASSED))
  else
    total_passed=$((total_passed + TESTS_PASSED))
    total_failed=$((total_failed + TESTS_FAILED))
    suites_failed=$((suites_failed + 1))
  fi

  if test_imaps; then
    total_passed=$((total_passed + TESTS_PASSED))
  else
    total_passed=$((total_passed + TESTS_PASSED))
    total_failed=$((total_failed + TESTS_FAILED))
    suites_failed=$((suites_failed + 1))
  fi

  if test_auth_enforcement; then
    total_passed=$((total_passed + TESTS_PASSED))
  else
    total_passed=$((total_passed + TESTS_PASSED))
    total_failed=$((total_failed + TESTS_FAILED))
    suites_failed=$((suites_failed + 1))
  fi

  if test_http_redirect; then
    total_passed=$((total_passed + TESTS_PASSED))
  else
    total_passed=$((total_passed + TESTS_PASSED))
    total_failed=$((total_failed + TESTS_FAILED))
    suites_failed=$((suites_failed + 1))
  fi

  # Final summary
  echo ""
  echo -e "${BOLD}══════════════════════════════════════════════${NC}"
  echo -e "${BOLD}  PRODUCTION TEST SUMMARY${NC}"
  echo -e "${BOLD}══════════════════════════════════════════════${NC}"
  echo ""
  echo -e "  Total:   $((total_passed + total_failed)) tests"
  echo -e "  Passed:  ${GREEN}${total_passed}${NC}"
  echo -e "  Failed:  ${RED}${total_failed}${NC}"
  echo ""

  if [ "$suites_failed" -gt 0 ]; then
    echo -e "  ${RED}${BOLD}PRODUCTION TESTS FAILED${NC} (${suites_failed} suite(s) with failures)"
    echo ""
    exit 1
  else
    echo -e "  ${GREEN}${BOLD}ALL PRODUCTION TESTS PASSED${NC}"
    echo ""
    exit 0
  fi
}

main "$@"
