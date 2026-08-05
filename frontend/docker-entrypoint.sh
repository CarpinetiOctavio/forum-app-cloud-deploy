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

# Same reasoning, applied to nginx's own listen port (ADR-008): Render
# injects PORT per service, and the port nginx binds to must be able to
# vary without rebuilding the image, same as REACT_APP_API_URL. Render
# assigns a default of 10000 if PORT is unset; 8080 here matches this
# repo's own existing convention (the backend's fallback, EXPOSE 8080),
# not Render's own default -- deliberate, not an oversight.
# /etc/nginx/conf.d/ is root-owned and read-only to the nginx user, same
# S6504 reasoning as above -- the whole directory (not just default.conf)
# is copied to /tmp/conf.d, not just the one file known to need editing
# today, so a future base-image update that adds another .conf file there
# doesn't silently stop being served. nginx.conf's own `include` directive
# was repointed at /tmp/conf.d/*.conf at build time (a fixed path, not an
# environment-specific value, so build-time is fine per docker.md Rule 1).
PORT="${PORT:-8080}"
mkdir -p /tmp/conf.d
cp -r /etc/nginx/conf.d/. /tmp/conf.d/
sed -i "s/listen  *[0-9]*;/listen ${PORT};/" /tmp/conf.d/default.conf

exec "$@"
