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

SERVICE_ID=$(echo "$HOOK_URL" | grep -oE 'srv-[A-Za-z0-9]+' | head -n1)
if [ -z "$SERVICE_ID" ]; then
  echo "::error::could not extract a service ID (srv-...) from the hook URL" >&2
  exit 1
fi
echo "Service ID: $SERVICE_ID" >&2

# --- Trigger the deploy, retrying with backoff on the "already in
#     progress" case (202 response, no deploy id in the body) ---
DEPLOY_ID=""
BACKOFF=15
for attempt in $(seq 1 5); do
  RESPONSE=$(curl -sS -X POST "${HOOK_URL}&imgURL=${IMAGE_URL}")
  DEPLOY_ID=$(echo "$RESPONSE" | jq -r '.deploy.id // empty')
  if [ -n "$DEPLOY_ID" ]; then
    break
  fi
  echo "Attempt ${attempt}/5: no deploy id in response (a deploy is likely already in progress) — retrying in ${BACKOFF}s" >&2
  echo "Response: $RESPONSE" >&2
  sleep "$BACKOFF"
  BACKOFF=$((BACKOFF * 2))
done

if [ -z "$DEPLOY_ID" ]; then
  echo "::error::no deploy id returned for service $SERVICE_ID after 5 attempts" >&2
  exit 1
fi
echo "Deploy ID: $DEPLOY_ID" >&2

# --- Poll until the deploy reaches a terminal state ---
# Terminal success: live. Terminal failure: build_failed, update_failed,
# pre_deploy_failed, canceled. Anything else keeps waiting (queued,
# build_in_progress, update_in_progress, pre_deploy_in_progress, created,
# deactivated) — verified against Render's real deployStatus enum.
MAX_ATTEMPTS=30
INTERVAL=20
STATUS=""
for attempt in $(seq 1 "$MAX_ATTEMPTS"); do
  STATUS=$(curl -sS -H "Authorization: Bearer ${API_KEY}" \
    "https://api.render.com/v1/services/${SERVICE_ID}/deploys/${DEPLOY_ID}" | jq -r '.status')
  case "$STATUS" in
    live)
      echo "Deploy $DEPLOY_ID is live" >&2
      break
      ;;
    build_failed|update_failed|pre_deploy_failed|canceled)
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
SERVICE_URL=$(curl -sS -H "Authorization: Bearer ${API_KEY}" \
  "https://api.render.com/v1/services/${SERVICE_ID}" | jq -r '.serviceDetails.url // empty')

if [ -z "$SERVICE_URL" ]; then
  echo "::error::could not resolve public URL for service $SERVICE_ID" >&2
  exit 1
fi

echo "Resolved URL: $SERVICE_URL" >&2
echo "$SERVICE_URL"
