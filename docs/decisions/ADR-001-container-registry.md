# ADR-001 — Container Registry: GitHub Container Registry (ghcr.io)

## Status

Accepted

## Date

2025

## Context

The CI/CD pipeline needs a registry to store and distribute Docker images between the build stage and the deployment environments. The registry must be accessible from GitHub Actions (build side) and from Render (deploy side), support private image storage at no cost, and integrate with the existing GitHub-based workflow.

## Decision

Use GitHub Container Registry (ghcr.io) as the container registry for this project.

## Rationale

- **Native GitHub integration:** GitHub Actions authenticates to ghcr.io using the built-in `GITHUB_TOKEN` — no additional credentials or secrets need to be configured for the push step. This reduces setup complexity and eliminates one category of potential secret exposure.
- **Same ecosystem:** the repository, the CI/CD pipeline, and the registry all live within GitHub. This minimizes friction between tools and keeps the workflow coherent.
- **Free for public repositories:** ghcr.io provides unlimited private and public image storage for GitHub accounts at no cost, which satisfies the project's zero-budget constraint.
- **Image visibility control:** images can be scoped to the repository, making access management straightforward.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| Docker Hub | Requires a separate account and credentials secret for authentication in the pipeline. Free tier limits pulls, which could cause pipeline failures. |
| GitLab Container Registry | Would require migrating the repository to GitLab, which contradicts the decision to use GitHub Actions and the existing GitHub workflow. |
| Amazon ECR | Requires AWS account setup, IAM credentials management, and incurs costs beyond the free tier for this use case. |
| Azure Container Registry | Same cost and setup complexity concerns as ECR. Unnecessary for a project that does not use Azure. |

## Consequences

- Authentication in the pipeline uses `GITHUB_TOKEN` — no registry-specific secret is needed for the push step.
- Images are referenced as `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend` and `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend`.
- Render must be configured to pull from ghcr.io. If the repository is private, a personal access token with `read:packages` scope must be configured in Render as a registry credential.