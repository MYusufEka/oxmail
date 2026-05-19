#!/usr/bin/env bash
set -euo pipefail

# api_test.sh — All CRUD endpoints, health, logs
# Comprehensive API coverage for domains, users, aliases, DKIM, health, and logs.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

suite_start "API Test"

# ---------------------------------------------------------------------------
# Step 1: Wait for stack healthy
# ---------------------------------------------------------------------------

wait_for_healthy 120

# ---------------------------------------------------------------------------
# Step 2: Health endpoint
# ---------------------------------------------------------------------------

print_header "Health Endpoint"

RESPONSE=$(api_get "/api/health")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")

assert_status_code "GET /api/health" "200" "$STATUS"
assert_contains "Health reports API status" "$BODY" "api"
assert_contains "Health reports postfix status" "$BODY" "postfix"
assert_contains "Health reports dovecot status" "$BODY" "dovecot"
assert_contains "Health reports redis status" "$BODY" "redis"

# ---------------------------------------------------------------------------
# Step 3: Domain CRUD
# ---------------------------------------------------------------------------

print_header "Domain CRUD"

# Create
RESPONSE=$(api_post "/api/domains" "{\"name\":\"apitest.example\"}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "POST /api/domains (create)" "201" "$STATUS"
assert_json_field "Domain name returned" "$BODY" "name" "apitest.example"

# List
RESPONSE=$(api_get "/api/domains")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/domains (list)" "200" "$STATUS"
assert_contains "Domain in list" "$BODY" "apitest.example"

# Get single
RESPONSE=$(api_get "/api/domains/apitest.example")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/domains/apitest.example" "200" "$STATUS"
assert_json_field "Domain name matches" "$BODY" "name" "apitest.example"

# Delete
RESPONSE=$(api_delete "/api/domains/apitest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "DELETE /api/domains/apitest.example" "200" "$STATUS"

# Verify deleted
RESPONSE=$(api_get "/api/domains/apitest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "GET deleted domain → 404" "404" "$STATUS"

# ---------------------------------------------------------------------------
# Step 4: User CRUD
# ---------------------------------------------------------------------------

print_header "User CRUD"

# Create domain first (users need a domain)
api_post "/api/domains" "{\"name\":\"usertest.example\"}" >/dev/null 2>&1

# Create user
RESPONSE=$(api_post "/api/users" "{\"email\":\"testuser@usertest.example\",\"password\":\"${PASSWORD}\",\"displayName\":\"Test User\"}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "POST /api/users (create)" "201" "$STATUS"
assert_contains "User email in response" "$BODY" "testuser@usertest.example"

# List users
RESPONSE=$(api_get "/api/users")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/users (list)" "200" "$STATUS"
assert_contains "User in list" "$BODY" "testuser@usertest.example"

# Delete user
RESPONSE=$(api_delete "/api/users/testuser@usertest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "DELETE /api/users/testuser@usertest.example" "200" "$STATUS"

# Verify deleted
RESPONSE=$(api_get "/api/users/testuser@usertest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "GET deleted user → 404" "404" "$STATUS"

# Cleanup domain
api_delete "/api/domains/usertest.example" >/dev/null 2>&1

# ---------------------------------------------------------------------------
# Step 5: Alias CRUD
# ---------------------------------------------------------------------------

print_header "Alias CRUD"

# Setup: create domain and user
api_post "/api/domains" "{\"name\":\"aliastest.example\"}" >/dev/null 2>&1
api_post "/api/users" "{\"email\":\"real@aliastest.example\",\"password\":\"${PASSWORD}\",\"displayName\":\"Real User\"}" >/dev/null 2>&1

# Create alias
RESPONSE=$(api_post "/api/aliases" "{\"source\":\"alias@aliastest.example\",\"destination\":\"real@aliastest.example\"}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "POST /api/aliases (create)" "201" "$STATUS"
assert_contains "Alias source in response" "$BODY" "alias@aliastest.example"

# List aliases
RESPONSE=$(api_get "/api/aliases")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/aliases (list)" "200" "$STATUS"
assert_contains "Alias in list" "$BODY" "alias@aliastest.example"

# Delete alias
RESPONSE=$(api_delete "/api/aliases/alias@aliastest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "DELETE /api/aliases/alias@aliastest.example" "200" "$STATUS"

# Verify deleted
RESPONSE=$(api_get "/api/aliases/alias@aliastest.example")
STATUS=$(extract_status "$RESPONSE")
assert_status_code "GET deleted alias → 404" "404" "$STATUS"

# Cleanup
api_delete "/api/users/real@aliastest.example" >/dev/null 2>&1
api_delete "/api/domains/aliastest.example" >/dev/null 2>&1

# ---------------------------------------------------------------------------
# Step 6: DKIM
# ---------------------------------------------------------------------------

print_header "DKIM"

# Setup domain
api_post "/api/domains" "{\"name\":\"dkimtest.example\"}" >/dev/null 2>&1

# Generate DKIM key
RESPONSE=$(api_post "/api/domains/dkimtest.example/dkim" "{}")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "POST /api/domains/dkimtest.example/dkim (generate)" "200" "$STATUS"
assert_contains "DKIM record in response" "$BODY" "dkimtest.example"

# Get DKIM key
RESPONSE=$(api_get "/api/domains/dkimtest.example/dkim")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/domains/dkimtest.example/dkim" "200" "$STATUS"
assert_not_empty "DKIM public key present" "$BODY"

# Cleanup
api_delete "/api/domains/dkimtest.example" >/dev/null 2>&1

# ---------------------------------------------------------------------------
# Step 7: Logs endpoint
# ---------------------------------------------------------------------------

print_header "Logs Endpoint"

RESPONSE=$(api_get "/api/logs")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/logs" "200" "$STATUS"

RESPONSE=$(api_get "/api/logs?limit=5")
STATUS=$(extract_status "$RESPONSE")
BODY=$(extract_body "$RESPONSE")
assert_status_code "GET /api/logs?limit=5" "200" "$STATUS"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

suite_end
