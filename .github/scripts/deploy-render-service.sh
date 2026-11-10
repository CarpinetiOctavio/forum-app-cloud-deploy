#!/usr/bin/env bash
set -euo pipefail

# Triggers a Render deploy hook with a specific, already-built image, polls
# until the deploy reaches a terminal state, then resolves and prints the
# service's real public URL. Used by Stage 7 (deploy-qa) and Stage 9
# (deploy-prod) — see docs/rules/pipeline.md and ADR-007/ADR-009.
#
# Per ADR-007: the API key is used exclusively for read-only GET calls
# (status polling, service lookup) — never to trigger or modify anything.
# The hook call is the only write operation in this script.
#
# Every HTTP call checks its status code. Transient failures (network error,
# 429, 5xx) are retried; any other non-2xx response fails immediately, so a
# mistyped hook or an invalid API key surfaces as its real HTTP code instead
# of being retried as if it were a deploy already in progress.
#
# Only the final resolved URL goes to stdout; everything else goes to
# stderr, so callers can capture the URL cleanly via command substitution.
#
# Usage: deploy-render-service.sh <hook_url> <image_url> <api_key>

if [ "$#" -ne 3 ]; then
  echo "::error::usage: deploy-render-service.sh <hook_url> <image_url> <api_key>" >&2
  exit 1
fi

HOOK_URL="$1"
IMAGE_URL="$2"
API_KEY="$3"

CURL_TIMEOUT=30
BODY=$(mktemp)
trap 'rm -f "$BODY"' EXIT

# Performs one HTTP request, writing the body to $BODY and printing the
# status code. A network-level failure (DNS, timeout, refused) prints "000"
# instead of aborting the script under `set -e`, so callers can treat it as
# transient.
http_request() {
  local code
  code=$(curl -sS -o "$BODY" -w '%{http_code}' --max-time "$CURL_TIMEOUT" "$@") || code="000"
  echo "$code"
}

is_transient() {
  case "$1" in
    000|429|5??) return 0 ;;
    *) return 1 ;;
  esac
}

SERVICE_ID=$(echo "$HOOK_URL" | grep -oE 'srv-[A-Za-z0-9]+' | head -n1)
if [ -z "$SERVICE_ID" ]; then
  echo "::error::could not extract a service ID (srv-...) from the hook URL" >&2
  exit 1
fi
echo "Service ID: $SERVICE_ID" >&2

# --- Trigger the deploy. Retries with backoff on a 2xx response without a
#     deploy id (a deploy is already in progress) and on transient failures;
#     fails immediately on any other status code ---
DEPLOY_ID=""
BACKOFF=15
TRIGGER_ATTEMPTS=5
for attempt in $(seq 1 "$TRIGGER_ATTEMPTS"); do
  CODE=$(http_request -X POST "${HOOK_URL}&imgURL=${IMAGE_URL}")
  if [[ "$CODE" == 2?? ]]; then
    DEPLOY_ID=$(jq -r '.deploy.id // empty' "$BODY" 2>/dev/null || true)
    if [ -n "$DEPLOY_ID" ]; then
      break
    fi
    REASON="HTTP $CODE without a deploy id (a deploy is likely already in progress)"
  elif is_transient "$CODE"; then
    REASON="transient failure (HTTP $CODE)"
  else
    echo "::error::deploy hook for service $SERVICE_ID returned HTTP $CODE — check the deploy hook secret" >&2
    echo "Response: $(cat "$BODY")" >&2
    exit 1
  fi
  if [ "$attempt" -eq "$TRIGGER_ATTEMPTS" ]; then
    echo "::error::could not trigger a deploy for service $SERVICE_ID after ${TRIGGER_ATTEMPTS} attempts (last: ${REASON})" >&2
    exit 1
  fi
  echo "Attempt ${attempt}/${TRIGGER_ATTEMPTS}: ${REASON} — retrying in ${BACKOFF}s" >&2
  sleep "$BACKOFF"
  BACKOFF=$((BACKOFF * 2))
done
echo "Deploy ID: $DEPLOY_ID" >&2

# --- Poll until the deploy reaches a terminal state ---
# Terminal success: live. Terminal failure: build_failed, update_failed,
# pre_deploy_failed, canceled, deactivated (a deactivated deploy will never
# become live). Anything else keeps waiting (created, queued,
# build_in_progress, update_in_progress, pre_deploy_in_progress).
MAX_ATTEMPTS=30
INTERVAL=20
STATUS=""
for attempt in $(seq 1 "$MAX_ATTEMPTS"); do
  CODE=$(http_request -H "Authorization: Bearer ${API_KEY}" \
    "https://api.render.com/v1/services/${SERVICE_ID}/deploys/${DEPLOY_ID}")
  if [[ "$CODE" == 2?? ]]; then
    STATUS=$(jq -r '.status // empty' "$BODY" 2>/dev/null || true)
  elif is_transient "$CODE"; then
    echo "Attempt ${attempt}/${MAX_ATTEMPTS}: transient failure (HTTP $CODE), waiting ${INTERVAL}s" >&2
    sleep "$INTERVAL"
    continue
  else
    echo "::error::polling deploy $DEPLOY_ID returned HTTP $CODE — check RENDER_API_KEY" >&2
    exit 1
  fi
  case "$STATUS" in
    live)
      echo "Deploy $DEPLOY_ID is live" >&2
      break
      ;;
    build_failed|update_failed|pre_deploy_failed|canceled|deactivated)
      echo "::error::deploy $DEPLOY_ID reached terminal failure state: $STATUS" >&2
      exit 1
      ;;
    *)
      echo "Attempt ${attempt}/${MAX_ATTEMPTS}: status=$STATUS, waiting ${INTERVAL}s" >&2
      sleep "$INTERVAL"
      ;;
  esac
done

if [ "$STATUS" != "live" ]; then
  echo "::error::timed out waiting for deploy $DEPLOY_ID to go live (last status: $STATUS)" >&2
  exit 1
fi

# --- Resolve the service's real public URL (read-only GET, same key) ---
CODE=$(http_request -H "Authorization: Bearer ${API_KEY}" \
  "https://api.render.com/v1/services/${SERVICE_ID}")
if [[ "$CODE" != 2?? ]]; then
  echo "::error::looking up service $SERVICE_ID returned HTTP $CODE" >&2
  exit 1
fi
SERVICE_URL=$(jq -r '.serviceDetails.url // empty' "$BODY" 2>/dev/null || true)

if [ -z "$SERVICE_URL" ]; then
  echo "::error::could not resolve public URL for service $SERVICE_ID" >&2
  exit 1
fi

echo "Resolved URL: $SERVICE_URL" >&2
echo "$SERVICE_URL"
