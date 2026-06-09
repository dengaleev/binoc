#!/bin/sh
# Provision the binoc service dashboard into SigNoz via its HTTP API.
# Runs as the dashboard-provisioner one-shot service in docker-compose.yml
# once signoz is healthy. Credentials come from the same SIGNOZ_USER_ROOT_*
# env vars the signoz service uses (single-sourced via a YAML anchor).
#
# Note: the all-in-one image's SPA catch-all shadows /api/v1/login, so we
# authenticate via the v2 session endpoint instead.

set -e

SIGNOZ_URL="http://signoz:8080"
TITLE="binoc service"

login_payload="{\"email\":\"$SIGNOZ_USER_ROOT_EMAIL\",\"password\":\"$SIGNOZ_USER_ROOT_PASSWORD\",\"orgId\":\"$SIGNOZ_USER_ROOT_ORG_ID\"}"

# Root user reconciliation can lag the health endpoint by a moment — retry.
token=""
for _ in $(seq 1 30); do
  token=$(wget -q -O- --header 'Content-Type: application/json' \
    --post-data "$login_payload" "$SIGNOZ_URL/api/v2/sessions/email_password" \
    2>/dev/null | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p') || true
  if [ -n "$token" ]; then break; fi
  sleep 2
done

if [ -z "$token" ]; then
  echo "error: could not log in to SigNoz" >&2
  exit 1
fi

auth="Authorization: Bearer $token"

if wget -q -O- --header "$auth" "$SIGNOZ_URL/api/v1/dashboards" \
  | grep -q "\"title\":\"$TITLE\""; then
  echo "dashboard already exists"
  exit 0
fi

wget -q -O /dev/null --header "$auth" --header 'Content-Type: application/json' \
  --post-data "$(cat /dashboard.json)" "$SIGNOZ_URL/api/v1/dashboards"

echo "dashboard provisioned"
