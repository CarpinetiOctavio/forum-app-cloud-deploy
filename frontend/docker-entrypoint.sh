#!/bin/sh
set -e

# Runtime substitution of the build-time placeholder, per ADR-006.
# REACT_APP_API_URL is expected to be set per Render service (QA/PROD);
# absent that, fall back to the same non-sensitive localhost:8080 default
# the source code itself uses, so a missing env var fails loud (no network
# response) instead of silently shipping an empty base URL.
REACT_APP_API_URL="${REACT_APP_API_URL:-http://localhost:8080}"

# /usr/share/nginx/html is root-owned and read-only to the nginx user
# (docker:S6504 — the runtime user must not be able to modify what it
# serves). The substitution therefore happens on a copy in /tmp, which is
# writable by any user (sticky-bit 1777) with no chown needed; nginx's
# `root` directive points at /tmp/html instead (set at build time).
mkdir -p /tmp/html
cp -r /usr/share/nginx/html/. /tmp/html/

find /tmp/html -type f -name "*.js" \
  -exec sed -i "s|__REACT_APP_API_URL__|${REACT_APP_API_URL}|g" {} +

exec "$@"
