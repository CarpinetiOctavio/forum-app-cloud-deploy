#!/usr/bin/env bash
set -euo pipefail

# Verifies a deployed service is actually serving requests — not a re-run
# of the correctness suite (Stages 1-4), a minimal check against the real,
# running container, per docs/rules/smoke-testing.md. A failing check here
# means the deploy mechanism reported success but the service isn't
# actually up, which the deploy call's own status code cannot catch.
#
# Usage: smoke-test.sh <url> [json]
#   <url>  — full URL to GET, HTTP 200 is the passing signal.
#   json   — if passed literally as the second argument, the response body
#            must also be valid JSON (used for the backend's /api/posts
#            check; omitted for the frontend's static page).

if [ "$#" -lt 1 ]; then
  echo "::error::usage: smoke-test.sh <url> [json]" >&2
  exit 1
fi

URL="$1"
EXPECT_JSON="${2:-}"

echo "Smoke testing: $URL" >&2

HTTP_CODE=$(curl -sS -o /tmp/smoke-test-body -w "%{http_code}" "$URL")

if [ "$HTTP_CODE" != "200" ]; then
  echo "::error::smoke test failed: $URL returned HTTP $HTTP_CODE" >&2
  exit 1
fi

if [ "$EXPECT_JSON" = "json" ]; then
  if ! jq empty /tmp/smoke-test-body 2>/dev/null; then
    echo "::error::smoke test failed: $URL returned HTTP 200 but the body is not valid JSON" >&2
    exit 1
  fi
fi

echo "OK: $URL ($HTTP_CODE)" >&2
