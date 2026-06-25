# ADR-003 — Image Tagging Strategy: Commit SHA

## Status

Accepted

## Date

2025

## Context

Docker images pushed to ghcr.io must be tagged in a way that enables traceability, reproducibility, and safe rollback. The tagging strategy determines whether it is possible to know exactly what code is running in each environment at any point in time.

## Decision

Tag every image with the Git commit SHA (`${{ github.sha }}`). Never use `latest` as the sole tag in any pipeline step or deployment configuration.

## Rationale

- **Immutability:** a commit SHA is a fixed, unique identifier. Once an image is tagged with a SHA, that tag always refers to the same image. It cannot be overwritten by a subsequent build.
- **Traceability:** given the tag of a running image, it is possible to identify the exact commit — and therefore the exact code — that produced it. This makes debugging and auditing straightforward.
- **Reproducibility:** deploying the same SHA tag at any future point produces the same running image. There is no ambiguity about what version is being deployed.
- **Safe rollback:** because previous images are never overwritten, rolling back to a known good state is as simple as redeploying the previous SHA tag. No rebuild is required.
- **Environment parity:** QA and PROD deploy the same SHA tag. This is the technical proof that both environments run the same artifact.

## Why `latest` is explicitly prohibited

`latest` is a mutable tag. Every new build overwrites the previous `latest`. This creates several problems:

- It is impossible to know which commit produced the currently running `latest` image without checking external metadata.
- Rollback requires knowing the previous SHA externally, as `latest` no longer points to it.
- If QA and PROD both pull `latest` at different times, they may run different images despite appearing to use the same tag.
- A direct correction was given during academic review of a prior project (forum-app-qa-pipeline) where `latest` was used in the deploy configuration. This decision explicitly addresses that correction.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| `latest` only | Mutable, no traceability, no safe rollback. Explicitly prohibited. |
| Semantic versioning (`v1.0.0`) | Requires manual version management or additional tooling. Adds complexity without benefit for a project where every commit to `main` is a deployable unit. |
| Branch name + SHA (`main-a3f9c2b`) | Adds redundant information. The SHA alone is sufficient for traceability. |
| Build number | Sequential but not directly linked to source code. A SHA is more informative and already available in GitHub Actions without additional configuration. |

## Consequences

- Every pipeline run produces an image tagged with the SHA of the triggering commit.
- Render services are configured to pull the specific SHA tag, not `latest`.
- The deploy hook call in the pipeline must pass the SHA tag to Render as part of the deploy configuration.
- Images accumulate in ghcr.io over time. Periodic cleanup of old images may be needed in the long term, but is out of scope for this project.