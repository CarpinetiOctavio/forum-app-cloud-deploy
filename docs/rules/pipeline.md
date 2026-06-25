# Pipeline — CI/CD Structure and Rules

## Purpose

This file defines the structure, stages, and rules of the GitHub Actions CI/CD pipeline. Read `docs/rules/constraints.md` before this file. Any conflict resolves in favor of `constraints.md`.

---

## Trigger

The pipeline runs on every push or pull request to `main`.

---

## Stage order — sequential and blocking

Each stage must complete successfully before the next one starts. A failure at any stage aborts the entire pipeline. No stage is skipped, parallelized, or reordered without explicit justification documented in an ADR.

| Order | Stage | Type | Abort on failure |
|-------|-------|------|-----------------|
| 1 | Backend unit tests + coverage | Quality gate | Yes |
| 2 | Frontend unit tests + coverage | Quality gate | Yes |
| 3 | SonarCloud static analysis | Quality gate | Yes |
| 4 | Cypress E2E tests | Quality gate | Yes |
| 5 | Docker image build | Build | Yes |
| 6 | Push to ghcr.io | Publish | Yes |
| 7 | Deploy to QA (Render) | Deploy | Yes |
| 8 | Manual approval gate | Gate | Blocks |
| 9 | Deploy to PROD (Render) | Deploy | Yes |

---

## Stage specifications

### Stage 1 — Backend unit tests + coverage

- Working directory: `backend/`
- Command: `go test ./tests/services/... -v -cover -coverpkg=./internal/services/...`
- Coverage threshold: ≥70% — pipeline aborts if not met
- Coverage report must be generated and uploaded as a pipeline artifact

### Stage 2 — Frontend unit tests + coverage

- Working directory: `frontend/`
- Command: `npm test -- --coverage --watchAll=false`
- Coverage threshold: ≥70% — pipeline aborts if not met
- Coverage report must be generated and uploaded as a pipeline artifact

### Stage 3 — SonarCloud static analysis

- Uses `SonarSource/sonarcloud-github-action`
- Requires `SONAR_TOKEN` secret (repository-level)
- Configuration in `sonar-project.properties` — must be updated for the new repository before this stage runs
- Pipeline aborts if SonarCloud Quality Gate fails
- **Pending:** `sonar-project.properties` must be reconfigured for `CarpinetiOctavio/forum-app-cloud-deploy`. Code must analyze the existing file and report what needs to change before modifying it.

### Stage 4 — Cypress E2E tests

- Requires backend and frontend to be running as services within the pipeline job
- Command: `npx cypress run`
- All 15 existing E2E tests must pass
- Pipeline aborts if any test fails

### Stage 5 — Docker image build

- Builds backend and frontend images using their respective Dockerfiles
- Multi-stage build — see `docs/rules/docker.md`
- No image is built if stages 1–4 have not all passed
- Images are tagged with `${{ github.sha }}`

### Stage 6 — Push to ghcr.io

- Registry: `ghcr.io/carpinetioctavio`
- Authentication: uses `GITHUB_TOKEN` (automatically available in GitHub Actions — no additional secret required)
- Images pushed: `forum-app-cloud-deploy-backend` and `forum-app-cloud-deploy-frontend`
- Tag: commit SHA only — never `latest`

### Stage 7 — Deploy to QA

- Triggered by HTTP POST to the Render QA deploy hook
- Deploy hook URL stored as GitHub Secret under the `qa` environment: `RENDER_DEPLOY_HOOK_QA`
- Render pulls the image tagged with the current commit SHA from ghcr.io
- Stage succeeds when Render confirms the deploy

### Stage 8 — Manual approval gate

- Implemented via GitHub environment protection rules on the `prod` environment
- Required reviewer: repository owner
- Pipeline pauses indefinitely until approved or rejected
- Approval is logged in GitHub with reviewer identity and timestamp

### Stage 9 — Deploy to PROD

- Triggered by HTTP POST to the Render PROD deploy hook
- Deploy hook URL stored as GitHub Secret under the `prod` environment: `RENDER_DEPLOY_HOOK_PROD`
- Render pulls the same image — same commit SHA tag — that was deployed to QA
- Stage succeeds when Render confirms the deploy

---

## GitHub environments

Two environments must be configured in the repository settings:

| Environment | Protection rule | Secrets |
|-------------|----------------|---------|
| `qa` | None (automatic) | `RENDER_DEPLOY_HOOK_QA` |
| `prod` | Required reviewer: repository owner | `RENDER_DEPLOY_HOOK_PROD` |

PROD secrets are scoped to the `prod` environment and are not accessible in any job that does not specify `environment: prod`.

---

## Existing stages — do not modify

Stages 1–4 correspond to the pipeline inherited from `forum-app-qa-pipeline`. Their logic, commands, coverage thresholds, and quality gates must remain functionally equivalent to the original. New stages (5–9) are appended after the existing ones.

**Code must read the existing `ci.yml` before writing any pipeline modification.** Report the current structure and confirm what is being preserved versus what is being added before making any change.

---

## What Code must do before writing or modifying ci.yml

1. Read the existing `ci.yml` in full.
2. Report the current stage structure.
3. Confirm which stages are inherited and which are new.
4. Verify that `sonar-project.properties` has been updated before including the SonarCloud stage.
5. Confirm that both Dockerfiles exist and are valid before writing the build stage.
6. Report proposed `ci.yml` structure before writing any file.