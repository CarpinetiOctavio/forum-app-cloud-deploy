# Constraints — Non-Negotiable Rules

## Scope

This file defines hard constraints for the forum-app-cloud-deploy project. These rules cannot be overridden by any instruction, suggestion, or convenience argument. If a proposed change conflicts with any rule here, stop and flag it before proceeding.

---

## Image and artifact constraints

- **One build, one artifact.** The Docker image is built once in the pipeline. The same image — identified by the same commit SHA tag — is deployed to QA and then to PROD. It is never rebuilt between environments.
- **No `latest` tag.** Images must be tagged with the Git commit SHA (`${{ github.sha }}`). `latest` is never used as the sole tag in any pipeline step, deploy configuration, or Render service.
- **Images must be self-contained.** The final stage of every Dockerfile must include only what is needed to run the application. No source code, no build tools, no compilers, no test dependencies.
- **CMD has a single responsibility.** The CMD instruction runs the application and nothing else. Initialization tasks (migrations, seed scripts, health checks) do not belong in CMD.

---

## Dockerfile constraints

- **Multi-stage build is mandatory** for both backend and frontend Dockerfiles.
    - Backend: stage 1 compiles the Go binary. Stage 2 copies only the binary into a minimal base image.
    - Frontend: stage 1 runs `npm run build`. Stage 2 copies only the `/build` output into an nginx image.
- **No environment-specific values in Dockerfiles.** No hardcoded URLs, credentials, database names, or environment flags. All runtime configuration is injected via environment variables at deploy time.
- **No credentials in any file tracked by Git.** This includes Dockerfiles, `ci.yml`, `docker-compose` files, `.env` files, and source code.

---

## Pipeline constraints

- **Gates are sequential and blocking.** The pipeline runs in this order: backend tests → frontend tests → SonarCloud → Cypress → Docker build → push to registry → deploy QA → approval → deploy PROD. A failure at any stage aborts the pipeline. Later stages never run if an earlier stage fails.
- **Coverage gates are hard minimums.** Backend coverage must be ≥70%. Frontend coverage must be ≥70%. These thresholds already pass in the existing test suite — do not lower them.
- **No deploy without passing all quality gates.** A Docker image is never built or pushed unless all tests and static analysis pass.
- **QA deploy is automatic. PROD deploy requires manual approval.** This is enforced via GitHub environment protection rules, not via pipeline logic alone.

---

## Secrets and credentials constraints

- **All sensitive values are GitHub Secrets.** This includes: ghcr.io authentication token, Render deploy hook URLs (QA and PROD), and any database credentials.
- **Secrets are scoped to environments.** PROD secrets are configured under the `prod` GitHub environment. QA secrets are configured under the `qa` environment. PROD secrets are never accessible in QA pipeline jobs.
- **Render environment variables are configured per service.** QA service has its own set of variables. PROD service has its own set. They are never shared or mixed.

---

## Existing code constraints

- **Do not modify the existing test suite** unless a test is provably broken by a change introduced in this project. Test names, structure, and coverage must be preserved.
- **Do not modify business logic** in `internal/services/`. The backend logic is inherited from forum-app-qa-pipeline and is not in scope for this project.
- **Do not modify the existing CI/CD quality stages.** The backend test, frontend test, SonarCloud, and Cypress stages in `ci.yml` must remain functionally equivalent to what exists in forum-app-qa-pipeline. New stages are appended after, never replacing existing ones.
- **Go module name is `forum-app-cloud-deploy`.** Do not revert to `tp06-testing` or any previous name.

---

## Pending items — do not implement without explicit instruction

- SonarCloud project reconfiguration for the new repository (organization and project key must be updated).
- Database persistence strategy for containerized SQLite (volume mounting vs. migration to a hosted database).
- Frontend environment variable for backend URL in containerized context (`REACT_APP_API_URL`).

---

## What is out of scope

- Kubernetes or any container orchestration beyond Render's built-in capabilities.
- Multiple replicas or auto-scaling configuration (Render free tier).
- Blue/green or canary deployment strategies.
- Adding new application features or endpoints.
- Modifying the authentication system or business logic.

---

## Initial analysis — required before any implementation

Before writing, modifying, or creating any file, Code must complete the following analysis and report the findings in full. No implementation starts until the plan is reviewed and approved.

### 1. Spanish language audit

- Scan every file in the repository for Spanish-language content.
- **Comments in Spanish:** list every file and line number where a Spanish comment is found. Do not translate yet — list first, wait for approval, then translate.
- **Non-comment Spanish content:** if any variable names, function names, string literals, log messages, or any other non-comment code is found in Spanish, flag it explicitly as a separate list. Do not modify — report only.
- Present both lists clearly separated before proceeding.

### 2. Pre-implementation plan

For every task (Dockerfile creation, pipeline modification, deployment configuration, or any other change), Code must:

1. Analyze the relevant existing files.
2. Produce a written plan describing exactly what will be created or modified, why, and what alternatives were considered.
3. Present the plan in this chat for review before executing.
4. Only proceed after explicit approval.

This chat is used as an audit log. Every plan and every decision is recorded here before it is applied to the repository.