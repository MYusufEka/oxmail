#!/usr/bin/env bash
set -euo pipefail

# run-all.sh — Orchestrator that runs all E2E tests and generates a report.
# Usage: ./tests/e2e/run-all.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
EVIDENCE_DIR="${PROJECT_ROOT}/.sisyphus/evidence/e2e"

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
# Setup
# =============================================================================

mkdir -p "$EVIDENCE_DIR"

TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0
SUITE_RESULTS=()
OVERALL_START=$(date +%s)
REPORT_FILE="${EVIDENCE_DIR}/report-$(date +%Y%m%d-%H%M%S).txt"

# =============================================================================
# Test suites to run
# =============================================================================

SUITES=(
  "lifecycle_test.sh"
  "persistence_test.sh"
  "security_test.sh"
  "api_test.sh"
)

# =============================================================================
# Banner
# =============================================================================

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║         Oxmail E2E Test Suite                ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${CYAN}Started:${NC} $(date '+%Y-%m-%d %H:%M:%S')"
echo -e "  ${CYAN}Suites:${NC}  ${#SUITES[@]}"
echo ""

# =============================================================================
# Run each suite
# =============================================================================

for suite in "${SUITES[@]}"; do
  SUITE_PATH="${SCRIPT_DIR}/${suite}"
  TOTAL_SUITES=$((TOTAL_SUITES + 1))

  if [ ! -f "$SUITE_PATH" ]; then
    echo -e "  ${RED}✗${NC} ${suite} — file not found"
    FAILED_SUITES=$((FAILED_SUITES + 1))
    SUITE_RESULTS+=("FAIL ${suite} (not found)")
    continue
  fi

  if [ ! -x "$SUITE_PATH" ]; then
    chmod +x "$SUITE_PATH"
  fi

  SUITE_START=$(date +%s)
  SUITE_OUTPUT="${EVIDENCE_DIR}/${suite%.sh}.log"

  echo -e "${BOLD}▶ Running: ${suite}${NC}"

  if bash "$SUITE_PATH" > "$SUITE_OUTPUT" 2>&1; then
    SUITE_END=$(date +%s)
    SUITE_DURATION=$((SUITE_END - SUITE_START))
    PASSED_SUITES=$((PASSED_SUITES + 1))
    SUITE_RESULTS+=("PASS ${suite} (${SUITE_DURATION}s)")
    echo -e "  ${GREEN}✓${NC} ${suite} — passed (${SUITE_DURATION}s)"
  else
    SUITE_END=$(date +%s)
    SUITE_DURATION=$((SUITE_END - SUITE_START))
    FAILED_SUITES=$((FAILED_SUITES + 1))
    SUITE_RESULTS+=("FAIL ${suite} (${SUITE_DURATION}s)")
    echo -e "  ${RED}✗${NC} ${suite} — failed (${SUITE_DURATION}s)"
    echo -e "    ${YELLOW}Log:${NC} ${SUITE_OUTPUT}"
  fi
done

# =============================================================================
# Summary
# =============================================================================

OVERALL_END=$(date +%s)
OVERALL_DURATION=$((OVERALL_END - OVERALL_START))

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║                  Summary                     ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""

for result in "${SUITE_RESULTS[@]}"; do
  if [[ "$result" == PASS* ]]; then
    echo -e "  ${GREEN}✓${NC} ${result#PASS }"
  else
    echo -e "  ${RED}✗${NC} ${result#FAIL }"
  fi
done

echo ""
echo -e "  ${BOLD}Total:${NC}    ${TOTAL_SUITES} suites"
echo -e "  ${GREEN}Passed:${NC}   ${PASSED_SUITES}"
echo -e "  ${RED}Failed:${NC}   ${FAILED_SUITES}"
echo -e "  ${CYAN}Duration:${NC} ${OVERALL_DURATION}s"
echo ""

# =============================================================================
# Write report file
# =============================================================================

{
  echo "Oxmail E2E Test Report"
  echo "======================"
  echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
  echo "Duration: ${OVERALL_DURATION}s"
  echo ""
  echo "Results:"
  for result in "${SUITE_RESULTS[@]}"; do
    echo "  $result"
  done
  echo ""
  echo "Summary: ${PASSED_SUITES}/${TOTAL_SUITES} suites passed"
  if [ "$FAILED_SUITES" -gt 0 ]; then
    echo "Status: FAILED"
  else
    echo "Status: PASSED"
  fi
} > "$REPORT_FILE"

echo -e "  ${CYAN}Report:${NC} ${REPORT_FILE}"
echo ""

# =============================================================================
# Exit code
# =============================================================================

if [ "$FAILED_SUITES" -gt 0 ]; then
  echo -e "${RED}${BOLD}E2E TESTS FAILED${NC}"
  exit 1
else
  echo -e "${GREEN}${BOLD}ALL E2E TESTS PASSED${NC}"
  exit 0
fi
