#!/usr/bin/env bash
set -euo pipefail

# Checks that an image digest may be used as a rollback target: it must be
# an image of <image_repo> whose commit-SHA tag currently points to it, and
# that commit must have a successful deployment from `main` in the `prod`
# environment. Read-only: it inspects the public registry and reads the
# GitHub Deployments API; it never calls a deploy hook. See
# docs/rules/deployment.md (Rollback) and ADR-007's amendment of 2026-11-23.
#
# Limit: GitHub records the commit of a deployment, not the image digest, and
# this compares against the commit's current tag. A digest that was in PROD
# is rejected if a rerun later moved its tag; a digest from a rerun that never
# reached PROD is accepted if its commit did. ADR-007's amendment of
# 2026-11-23.
#
# Requires GH_TOKEN (deployments: read) and GITHUB_REPOSITORY.
#
# Usage: validate-rollback-digest.sh <image_repo> <digest>

if [ "$#" -ne 2 ]; then
  echo "::error::usage: validate-rollback-digest.sh <image_repo> <digest>" >&2
  exit 1
fi

IMAGE_REPO="$1"
DIGEST="$2"

if [[ ! "$DIGEST" =~ ^sha256:[0-9a-f]{64}$ ]]; then
  echo "::error::'${DIGEST}' is not a digest of the form sha256:<64 hex characters>." >&2
  exit 1
fi

if ! docker buildx imagetools inspect "${IMAGE_REPO}@${DIGEST}" >/dev/null 2>&1; then
  echo "::error::${IMAGE_REPO}@${DIGEST} does not exist in the registry." >&2
  exit 1
fi

DEPLOYMENTS=$(gh api --paginate \
  "repos/${GITHUB_REPOSITORY}/deployments?environment=prod&ref=main&per_page=100" \
  --jq '.[] | "\(.id) \(.sha)"')

while read -r ID SHA; do
  [ -n "${ID:-}" ] || continue
  SUCCEEDED=$(gh api "repos/${GITHUB_REPOSITORY}/deployments/${ID}/statuses?per_page=100" \
    --jq 'any(.[]; .state == "success")')
  [ "$SUCCEEDED" = "true" ] || continue
  TAGGED=$(docker buildx imagetools inspect "${IMAGE_REPO}:${SHA}" \
    --format '{{.Manifest.Digest}}' 2>/dev/null || true)
  if [ "$TAGGED" = "$DIGEST" ]; then
    echo "${IMAGE_REPO}@${DIGEST} is tagged ${SHA}, deployed to prod from main (deployment ${ID})." >&2
    exit 0
  fi
done <<< "$DEPLOYMENTS"

echo "::error::${IMAGE_REPO}@${DIGEST} is not the current image of any commit deployed to prod from main." >&2
exit 1
