#!/usr/bin/env bash
set -euo pipefail

# scripts/test-e2e.sh — Entry point for E2E tests (called by `make test-e2e`)
# Usage: ./scripts/test-e2e.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

exec bash "${PROJECT_ROOT}/tests/e2e/run-all.sh" "$@"
