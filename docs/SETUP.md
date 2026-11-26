# Setup

What a copy of this repository needs to run locally and to run its pipeline. The commands are in [`COMMANDS.md`](COMMANDS.md).

## Local

- Go 1.24 (`backend/go.mod`), Node.js 24 (the version `ci.yml` uses) and Docker.
- `git clone https://github.com/CarpinetiOctavio/forum-app-cloud-deploy.git`, then `go mod download` in `backend/` and `npm ci` in `frontend/`.

The Go module is `forum-app-cloud-deploy`. The repository started as a copy of `forum-app-qa-pipeline@v1.0.0`, so earlier names appear in its history ([ADR-000](decisions/ADR-000-starting-from-qa-pipeline-v1.0.0.md)).

## GitHub

- **Workflow permissions** (Settings → Actions → General): read. Every job in `ci.yml` and `rollback.yml` declares its own `permissions`.
- **Ruleset** on `staging` and `main`, with no bypass: a pull request is required, the branch cannot be deleted or force-pushed, and nine checks must pass: `Backend Tests (Go)`, `Frontend Tests (React)`, `Backend Build`, `Frontend Build`, `SonarCloud Code Analysis`, `Cypress E2E`, `Test Summary`, `Docker Build & Push (Backend)` and `Docker Build & Push (Frontend)`.
- **Environments:** `qa` and `prod`, each accepting deployments only from `main`, with no bypass for administrators; `prod` also requires a reviewer ([ADR-007](decisions/ADR-007-render-deploy-mechanism.md)).
- **Secrets**, each loaded with the interactive prompt of `gh secret set <name>` (add `--env qa` or `--env prod` for an environment secret), so the value never appears in a command:

  | Secret | Scope | Where it comes from |
  |---|---|---|
  | `SONAR_TOKEN` | repository | a SonarCloud token |
  | `RENDER_API_KEY` | repository | Render → Account Settings → API Keys; used only for read-only calls ([ADR-007](decisions/ADR-007-render-deploy-mechanism.md)) |
  | `RENDER_DEPLOY_HOOK_BACKEND_QA`, `RENDER_DEPLOY_HOOK_FRONTEND_QA` | `qa` | Render → each QA service → Settings → Deploy Hook |
  | `RENDER_DEPLOY_HOOK_BACKEND_PROD`, `RENDER_DEPLOY_HOOK_FRONTEND_PROD` | `prod` | Render → each PROD service → Settings → Deploy Hook |

## Container registry

The images are `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend` and `-frontend` ([ADR-002](decisions/ADR-002-container-registry.md)). The pipeline pushes with `GITHUB_TOKEN`, which cannot push to a package that does not exist yet or is not linked to this repository, so each package is set up once by hand ([ADR-002](decisions/ADR-002-container-registry.md), Consequences and amendment of 2026-08-05):

1. Push a first `linux/amd64` image manually (`COMMANDS.md`, Registry).
2. In the package settings: **Connect repository** to `forum-app-cloud-deploy`, and under **Manage Actions access** give the repository the **Write** role.
3. Make the package public, so Render pulls it without a credential.

## SonarCloud

Project `CarpinetiOctavio_forum-app-cloud-deploy` in the organization `carpinetioctavio` (`sonar-project.properties`), on the Free plan, which has a quality gate only for `main` and for pull requests into `main` ([ADR-009](decisions/ADR-009-single-build-trigger.md), amendment of 2026-11-23).

## Render

One project with two environments, QA and Production, and four web services on the Free plan, each created from an existing image in ghcr.io:

| Service | Environment | Variables |
|---|---|---|
| `forum-backend-qa` | QA | none |
| `forum-frontend-qa` | QA | `REACT_APP_API_URL` = the URL of `forum-backend-qa` |
| `forum-backend-prod` | Production | none |
| `forum-frontend-prod` | Production | `REACT_APP_API_URL` = the URL of `forum-backend-prod` |

- Each service is created with the commit-SHA tag of an image already in the registry. From then on, every deploy comes from the pipeline through the service's deploy hook, which passes the image digest ([ADR-007](decisions/ADR-007-render-deploy-mechanism.md)).
- **Auto-Deploy is off** on all four services, so nothing deploys outside the pipeline.
- The frontend substitutes `REACT_APP_API_URL` into its bundle when the container starts ([ADR-006](decisions/ADR-006-frontend-runtime-config.md)). Both containers listen on `PORT` ([ADR-008](decisions/ADR-008-explicit-port-binding.md)), and the backend uses `./database.db` when `DATABASE_PATH` is not set ([ADR-005](decisions/ADR-005-sqlite-persistence.md)).
