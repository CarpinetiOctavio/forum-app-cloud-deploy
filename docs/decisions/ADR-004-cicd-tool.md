# ADR-004 — CI/CD Tool: GitHub Actions

## Status

Accepted

## Date

2025

## Context

The project needs a CI/CD tool to automate the full pipeline: running tests, static analysis, building Docker images, pushing to a registry, and deploying to QA and PROD. The tool must integrate with the GitHub repository, support environment-based approval gates, and be free of cost for public repositories.

## Decision

Use GitHub Actions as the CI/CD tool.

## Rationale

- **Inherited from forum-app-qa-pipeline:** the existing quality assurance pipeline (stages 1–4) is already implemented in GitHub Actions. Replacing the tool would require rewriting the entire existing pipeline from scratch, introducing risk and providing no benefit.
- **Native GitHub integration:** workflows are triggered by repository events (push, pull request) without external webhook configuration. The `GITHUB_TOKEN` provides built-in authentication to ghcr.io. Environment protection rules for the approval gate are a first-class GitHub feature, not a workaround.
- **Environment support:** GitHub Actions supports named environments with protection rules (required reviewers), which is the mechanism used to implement the manual approval gate before PROD deployment.
- **Free for public repositories:** GitHub Actions provides unlimited minutes for public repositories, which covers this project's pipeline without cost.
- **No additional tooling:** the pipeline runs entirely within GitHub. No external CI/CD service needs to be configured, authenticated, or maintained.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| GitLab CI/CD | Would require migrating the repository and all history to GitLab. The existing pipeline in GitHub Actions would be lost. No benefit justifies the migration cost. |
| CircleCI | Requires a separate account, external webhook configuration, and a different YAML syntax. Free tier is limited (6,000 minutes/month). Adds an external dependency for no gain over GitHub Actions. |
| Jenkins | Self-hosted, requires infrastructure setup and maintenance. Disproportionate complexity for this project's scope. |
| Azure DevOps Pipelines | No existing Azure infrastructure. Adds an external platform dependency. Free tier for public projects is sufficient but setup complexity exceeds GitHub Actions for this use case. |

## Consequences

- All pipeline configuration lives in `.github/workflows/ci.yml`.
- The existing quality stages (backend tests, frontend tests, SonarCloud, Cypress) are preserved and extended with new stages for Docker build, registry push, and deployment.
- Two GitHub environments (`qa` and `prod`) must be configured in repository settings before the pipeline can deploy.
- The `prod` environment must have a required reviewer protection rule configured before the first PROD deployment.