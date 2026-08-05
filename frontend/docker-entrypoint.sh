#!/bin/sh
set -e

# Runtime substitution of the build-time placeholder, per ADR-006.
# REACT_APP_API_URL is expected to be set per Render service (QA/PROD);
# absent that, fall back to the same non-sensitive localhost:8080 default
# the source code itself uses, so a missing env var fails loud (no network
# response) instead of silently shipping an empty base URL.
REACT_APP_API_URL="${REACT_APP_API_URL:-http://localhost:8080}"

find /usr/share/nginx/html -type f -name "*.js" \
  -exec sed -i "s|__REACT_APP_API_URL__|${REACT_APP_API_URL}|g" {} +

exec "$@"
