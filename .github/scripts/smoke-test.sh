#!/usr/bin/env bash
set -euo pipefail

# Verifies a deployed service is actually serving requests — not a re-run
# of the correctness suite (Stages 1-4), a minimal check against the real,
# running container, per docs/rules/smoke-testing.md. A failing check here
# means the deploy mechanism reported success but the service isn't
# actually up, which the deploy call's own status code cannot catch.
#
# Usage: smoke-test.sh <url> [json | frontend <expected-api-url>]
#   <url>      — full URL to GET, HTTP 200 is the passing signal.
#   json       — the response body must also be valid JSON (the backend's
#                /api/posts check).
#   frontend   — <url> is the frontend's root page. The main JS bundle it
#                references must also return 200, must no longer contain
#                the build-time placeholder __REACT_APP_API_URL__, and must
#                contain <expected-api-url>: proof that the container's
#                entrypoint substituted this environment's backend URL
#                (ADR-006). nginx serves index.html with 200 even when that
#                substitution failed, so the root page alone can't show it.

if [ "$#" -lt 1 ]; then
  echo "::error::usage: smoke-test.sh <url> [json | frontend <expected-api-url>]" >&2
  exit 1
fi

URL="$1"
MODE="${2:-}"

if [ "$MODE" = "frontend" ] && [ -z "${3:-}" ]; then
  echo "::error::frontend mode needs the expected API URL as third argument" >&2
  exit 1
fi

# GET <url> into <file>, failing unless the response is HTTP 200.
# --max-time bounds each request so a service that never answers fails the
# job instead of hanging it. 90 s covers a Render Free cold start, which
# Render documents at about one minute.
fetch_ok() {
  local url="$1" out="$2" code
  code=$(curl -sS --max-time 90 -o "$out" -w "%{http_code}" "$url")
  if [ "$code" != "200" ]; then
    echo "::error::smoke test failed: $url returned HTTP $code" >&2
    exit 1
  fi
}

echo "Smoke testing: $URL" >&2
fetch_ok "$URL" /tmp/smoke-test-body

if [ "$MODE" = "json" ]; then
  if ! jq empty /tmp/smoke-test-body 2>/dev/null; then
    echo "::error::smoke test failed: $URL returned HTTP 200 but the body is not valid JSON" >&2
    exit 1
  fi
fi

if [ "$MODE" = "frontend" ]; then
  EXPECTED_API_URL="$3"
  # Create React App injects exactly one main bundle into index.html.
  BUNDLE_PATH=$(grep -o '/static/js/main\.[0-9a-f]*\.js' /tmp/smoke-test-body | head -n 1 || true)
  if [ -z "$BUNDLE_PATH" ]; then
    echo "::error::smoke test failed: $URL served no reference to a main JS bundle" >&2
    exit 1
  fi
  BUNDLE_URL="${URL%/}${BUNDLE_PATH}"
  fetch_ok "$BUNDLE_URL" /tmp/smoke-test-bundle
  if grep -q '__REACT_APP_API_URL__' /tmp/smoke-test-bundle; then
    echo "::error::smoke test failed: $BUNDLE_URL still contains the __REACT_APP_API_URL__ placeholder" >&2
    exit 1
  fi
  if ! grep -qF "$EXPECTED_API_URL" /tmp/smoke-test-bundle; then
    echo "::error::smoke test failed: $BUNDLE_URL does not contain the expected API URL $EXPECTED_API_URL" >&2
    exit 1
  fi
  echo "OK: $BUNDLE_URL points at $EXPECTED_API_URL" >&2
fi

echo "OK: $URL" >&2
