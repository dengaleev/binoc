#!/bin/sh
# Provision the binoc service dashboard into SigNoz via the API.
# Usage: ./provision-dashboard.sh [compose-project-name]
#
# Workaround: SigNoz's all-in-one image serves its SPA on the same port as
# the API, and the SPA catch-all shadows /api/v1/login. We use the v2 session
# endpoint instead and call the API from inside the Docker network (via the
# zookeeper container which has curl).

set -e

PROJECT="${1:-binoc-signoz}"
SIGNOZ_URL="http://signoz:8080"
CURL_CONTAINER="${PROJECT}-zookeeper-1"

EMAIL="admin@example.com"
PASSWORD='Admin@Signoz1!'
ORG_ID="019d0000-0000-7000-8000-000000000001"

DASHBOARD_JSON="$(dirname "$0")/dashboard.json"

# Copy dashboard JSON into the curl container
docker cp "$DASHBOARD_JSON" "$CURL_CONTAINER:/tmp/dashboard.json"

# Login
LOGIN_PAYLOAD=$(printf '{"email":"%s","password":"%s","orgId":"%s"}' "$EMAIL" "$PASSWORD" "$ORG_ID")
printf '%s' "$LOGIN_PAYLOAD" > /tmp/signoz-login.json
docker cp /tmp/signoz-login.json "$CURL_CONTAINER:/tmp/login.json"

TOKEN=$(docker exec "$CURL_CONTAINER" curl -sf -X POST "$SIGNOZ_URL/api/v2/sessions/email_password" \
  -H 'Content-Type: application/json' -d @/tmp/login.json \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")

if [ -z "$TOKEN" ]; then
  echo "Error: failed to get access token" >&2
  exit 1
fi

# Check if dashboard exists
EXISTS=$(docker exec "$CURL_CONTAINER" curl -sf "$SIGNOZ_URL/api/v1/dashboards" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -c "
import sys, json
dashboards = json.load(sys.stdin).get('data', [])
print('yes' if any(d.get('data', {}).get('title') == 'binoc service' for d in dashboards) else 'no')
")

if [ "$EXISTS" = "yes" ]; then
  echo "Dashboard already exists, skipping"
  exit 0
fi

# Create dashboard
docker exec "$CURL_CONTAINER" curl -sf -X POST "$SIGNOZ_URL/api/v1/dashboards" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @/tmp/dashboard.json > /dev/null

echo "Dashboard provisioned successfully"
