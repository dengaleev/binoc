#!/bin/sh
# Playground mode: skip authentication
export IS_LOCAL_APP_MODE="DANGEROUSLY_is_local_app_mode💀"

# Override the baked-in __ENV.js so the frontend knows it's local mode
echo 'window.__ENV = {"NEXT_PUBLIC_IS_LOCAL_MODE":"true","NEXT_PUBLIC_APP_VERSION":"2.22.1"};' \
  > /app/packages/app/packages/app/public/__ENV.js

source /etc/local/entry.base.sh
