# Forum App — Cloud Deploy

## What is this project

Full-stack forum application (Go backend + React/TypeScript frontend) containerized and deployed to the cloud with a complete CI/CD pipeline. This repository extends [forum-app-qa-pipeline](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline) by adding Docker containerization, a container registry, and automated deployment to QA and production environments.

**Course:** Software Engineering III (Ingeniería de Software III) — Universidad Católica de Córdoba (UCC), 2025.
**Author:** Octavio Carpineti.
**Repository:** https://github.com/CarpinetiOctavio/forum-app-cloud-deploy

This project will be defended in an oral exam — every technical decision must be justifiable from a software engineering perspective.

---

## Tech stack — do not change without consulting

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24 + SQLite |
| Frontend | React 18 + TypeScript |
| Unit testing | Go testing + Jest |
| E2E testing | Cypress |
| Static analysis | SonarCloud |
| CI/CD | GitHub Actions |
| Containerization | Docker (multi-stage builds) |
| Container registry | GitHub Container Registry (ghcr.io) |
| Cloud hosting | Render.com (QA + PROD) |

---

## Project structure — respect strictly

```
forum-app-cloud-deploy/
├── backend/
│   ├── cmd/api/main.go              # Server entry point
│   ├── internal/
│   │   ├── handlers/                # HTTP handlers — no business logic
│   │   ├── services/                # Business logic
│   │   ├── repository/              # Data access interfaces
│   │   ├── models/                  # Data structures
│   │   ├── database/                # Database configuration
│   │   └── router/                  # Route configuration
│   ├── tests/
│   │   ├── services/                # Unit tests
│   │   └── mocks/                   # Test mocks
│   └── Dockerfile                   # Multi-stage build
├── frontend/
│   ├── src/
│   ├── cypress/                     # E2E tests
│   └── Dockerfile                   # Multi-stage build
├── .github/
│   └── workflows/
│       └── ci.yml                   # Full CI/CD pipeline
├── docs/
│   ├── rules/                       # AI assistant operating instructions
│   └── decisions/                   # Architecture Decision Records (ADR)
├── sonar-project.properties
└── README.md
```

---

## Essential commands

```bash
# Run backend locally
cd backend && go run cmd/api/main.go

# Run frontend locally
cd frontend && npm start

# Backend unit tests
cd backend && go test ./tests/services/... -v

# Backend tests with coverage
cd backend && go test ./tests/services/... -v -cover -coverpkg=./internal/services/...

# Frontend unit tests
cd frontend && npm test -- --watchAll=false

# Frontend tests with coverage
cd frontend && npm test -- --coverage --watchAll=false

# E2E tests (backend and frontend must be running)
cd frontend && npx cypress run

# Build Docker image — backend
docker build -t forum-backend ./backend

# Build Docker image — frontend
docker build -t forum-frontend ./frontend
```

---

## Pipeline stages — in order

1. Backend unit tests + coverage gate (≥70%)
2. Frontend unit tests + coverage gate (≥70%)
3. SonarCloud static analysis
4. Cypress E2E tests
5. Docker image build (multi-stage, single artifact)
6. Push to ghcr.io — tagged with commit SHA (`${{ github.sha }}`), never `latest`
7. Deploy to QA (Render) — automatic on passing all gates
8. Manual approval gate (GitHub environment protection rule)
9. Deploy to PROD (Render) — same image, same SHA tag

**Critical rule:** if any gate in steps 1–4 fails, the pipeline aborts. No image is built. Nothing is deployed.

---

## Non-negotiable constraints

- Images are tagged with commit SHA — never with `latest`
- The same image artifact is deployed to both QA and PROD — never rebuilt
- No credentials, tokens, or secrets in Dockerfiles, YAML files, or source code
- Backend Dockerfile must use multi-stage build — builder stage and final minimal stage
- Frontend Dockerfile must use multi-stage build — builder stage and final nginx stage
- CMD in Dockerfiles must have a single responsibility: run the application
- Environment-specific configuration is injected at runtime via environment variables — never baked into the image

---

## References — read all at the start. Consult by @mention each time they are referenced

- `docs/rules/pipeline.md` — CI/CD pipeline stages, gates, and GitHub Actions workflow structure
- `docs/rules/docker.md` — Dockerfile rules, multi-stage build requirements, image constraints
- `docs/rules/deployment.md` — Render configuration, environments, secrets, deploy hooks, rollback
- `docs/rules/testing.md` — existing test suite, how to run tests, coverage gates, naming conventions
- `docs/rules/constraints.md` — non-negotiable restrictions, what not to touch, pending items
- `docs/decisions/ADR-001-container-registry.md` — why ghcr.io
- `docs/decisions/ADR-002-hosting.md` — why Render
- `docs/decisions/ADR-003-image-tagging.md` — why commit SHA and not latest
- `docs/decisions/ADR-004-cicd-tool.md` — why GitHub Actions

---

## Initialization — run this on first load

When opening this project for the first time, execute the following steps in order before doing anything else. Do not skip steps. Do not start any implementation until all steps are complete and reported.

### Step 1 — Read all rule files

Read every file in `docs/rules/` in this order:
1. `docs/rules/constraints.md`
2. `docs/rules/docker.md`
3. `docs/rules/pipeline.md`
4. `docs/rules/deployment.md`
5. `docs/rules/testing.md`

### Step 2 — Read all ADRs

Read every file in `docs/decisions/`:
1. `docs/decisions/ADR-001-container-registry.md`
2. `docs/decisions/ADR-002-hosting.md`
3. `docs/decisions/ADR-003-image-tagging.md`
4. `docs/decisions/ADR-004-cicd-tool.md`

### Step 3 — Analyze the full repository

Scan every file in the repository. Report the following:

**3a. Spanish language audit**
- List every file and line number where a Spanish-language comment is found. Do not translate yet.
- List separately any non-comment content found in Spanish (variable names, function names, string literals, log messages). Do not modify — report only.

**3b. Project structure summary**
- Confirm the current directory structure matches what is documented in this file.
- Flag any discrepancy.

**3c. Existing pipeline summary**
- Read `.github/workflows/ci.yml` in full.
- Report the current stage structure.
- Confirm what stages are inherited from forum-app-qa-pipeline and what stages are missing (Docker build, registry push, QA deploy, approval gate, PROD deploy).

**3d. Pending items status**
- Read the pending items in `docs/rules/constraints.md`.
- For each pending item, report what you found in the source code that is relevant to resolving it.

### Step 4 — Report and wait

Present the full report from Step 3 before doing anything else. No file is created or modified until the report is reviewed and a plan is approved.