#!/bin/sh
# Provision the binoc service dashboard into SigNoz.
# Usage: ./provision-dashboard.sh [compose-project-name]
# Runs after 'make up STACK=signoz' — copies SQLite DB from the container,
# inserts the dashboard, and copies it back.

set -e

PROJECT="${1:-binoc-signoz}"
CONTAINER="${PROJECT}-signoz-1"
DB_REMOTE="/var/lib/signoz/signoz.db"
DB_LOCAL="/tmp/signoz-provision.db"
DASHBOARD_JSON="$(dirname "$0")/dashboard.json"

# Copy DB from container
docker cp "$CONTAINER:$DB_REMOTE" "$DB_LOCAL"

# Insert dashboard if not exists
python3 - "$DB_LOCAL" "$DASHBOARD_JSON" <<'PYEOF'
import sqlite3, sys, json

db_path, json_path = sys.argv[1], sys.argv[2]
conn = sqlite3.connect(db_path)
c = conn.cursor()

# Check if dashboard already exists
c.execute("SELECT count(*) FROM dashboard WHERE json_extract(data, '$.title') = 'binoc service'")
if c.fetchone()[0] > 0:
    print("Dashboard already exists, skipping")
    sys.exit(0)

# Get org_id from users
c.execute("SELECT org_id FROM users LIMIT 1")
row = c.fetchone()
org_id = row[0] if row else "default"

# Read dashboard JSON
with open(json_path) as f:
    data = f.read()

import uuid
dashboard_id = str(uuid.uuid4())
c.execute(
    "INSERT INTO dashboard (id, created_at, updated_at, created_by, updated_by, data, locked, org_id) VALUES (?, datetime('now'), datetime('now'), '', '', ?, 0, ?)",
    (dashboard_id, data, org_id)
)
conn.commit()
print("Dashboard provisioned successfully")
PYEOF

# Copy DB back
docker cp "$DB_LOCAL" "$CONTAINER:$DB_REMOTE"
rm -f "$DB_LOCAL"
