#!/usr/bin/env bash
set -euo pipefail

# seed.sh — Create test domain, users, and emails via API
# Usage: ./scripts/seed.sh
# Env: OXMAIL_API_URL (default: http://localhost:8080)

API_URL="${OXMAIL_API_URL:-http://localhost:8080}"
DOMAIN="local.test"
PASSWORD="TestPass123!"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_ok() { echo -e "${GREEN}✓${NC} $1"; }
print_err() { echo -e "${RED}✗${NC} $1"; }

# Helper: make API call and check response
api_post() {
  local endpoint="$1"
  local data="$2"
  local description="$3"

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$data" \
    "${API_URL}${endpoint}")

  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | sed '$d')

  if [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 300 ]; then
    print_ok "$description"
  else
    print_err "$description (HTTP $HTTP_CODE)"
    echo "  Response: $BODY"
    return 1
  fi
}

echo "=== Oxmail Seed Script ==="
echo "API: ${API_URL}"
echo ""

# Step 1: Wait for API to be healthy
echo "--- Checking API health ---"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"${SCRIPT_DIR}/wait-healthy.sh"
echo ""

# Step 2: Create domain
echo "--- Creating domain ---"
api_post "/api/domains" \
  "{\"domain\":\"${DOMAIN}\"}" \
  "Created domain: ${DOMAIN}"
echo ""

# Step 3: Create users
echo "--- Creating users ---"
api_post "/api/users" \
  "{\"email\":\"alice@${DOMAIN}\",\"password\":\"${PASSWORD}\",\"name\":\"Alice Test\"}" \
  "Created user: alice@${DOMAIN}"

api_post "/api/users" \
  "{\"email\":\"bob@${DOMAIN}\",\"password\":\"${PASSWORD}\",\"name\":\"Bob Test\"}" \
  "Created user: bob@${DOMAIN}"
echo ""

# Step 4: Send test emails
echo "--- Sending test emails ---"
api_post "/api/mail/send" \
  "{\"from\":\"alice@${DOMAIN}\",\"to\":\"bob@${DOMAIN}\",\"subject\":\"Hello Bob!\",\"body\":\"This is test email 1 from Alice.\"}" \
  "Email 1: alice → bob"

api_post "/api/mail/send" \
  "{\"from\":\"bob@${DOMAIN}\",\"to\":\"alice@${DOMAIN}\",\"subject\":\"Re: Hello Bob!\",\"body\":\"Hey Alice! Got your email. This is test email 2.\"}" \
  "Email 2: bob → alice"

api_post "/api/mail/send" \
  "{\"from\":\"alice@${DOMAIN}\",\"to\":\"bob@${DOMAIN}\",\"subject\":\"Meeting tomorrow\",\"body\":\"Can we meet at 10am? Test email 3.\"}" \
  "Email 3: alice → bob"

api_post "/api/mail/send" \
  "{\"from\":\"bob@${DOMAIN}\",\"to\":\"alice@${DOMAIN}\",\"subject\":\"Re: Meeting tomorrow\",\"body\":\"Sure, 10am works! Test email 4.\"}" \
  "Email 4: bob → alice"

api_post "/api/mail/send" \
  "{\"from\":\"alice@${DOMAIN}\",\"to\":\"bob@${DOMAIN}\",\"subject\":\"Project update\",\"body\":\"Everything is on track. Test email 5.\"}" \
  "Email 5: alice → bob"
echo ""

# Step 5: Generate DKIM key
echo "--- Generating DKIM key ---"
api_post "/api/domains/${DOMAIN}/dkim" \
  "{}" \
  "Generated DKIM key for ${DOMAIN}"
echo ""

# Summary
echo "=== Seed Complete ==="
echo "  Domain:  ${DOMAIN}"
echo "  Users:   alice@${DOMAIN}, bob@${DOMAIN}"
echo "  Password: ${PASSWORD}"
echo "  Emails:  5 test emails sent"
echo "  DKIM:    Generated for ${DOMAIN}"
