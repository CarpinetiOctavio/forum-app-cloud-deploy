# Forum App — Cloud Deploy

**Author:** Octavio Carpineti
**Course:** Software Engineering III — Universidad Católica de Córdoba (UCC)
**Year:** 2025

Full-stack forum application containerized and deployed to the cloud with a complete CI/CD pipeline. This project builds on [forum-app-qa-pipeline](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline), extending it with Docker containerization, a container registry, and automated deployment to QA and production environments on Render via GitHub Actions.
---

## Table of Contents

1. [Project Description](#project-description)
2. [Tech Stack](#tech-stack)
3. [Prerequisites](#prerequisites)
4. [Installation](#installation)
5. [Running the Project](#running-the-project)
6. [Running Tests](#running-tests)
7. [Quality Tools](#quality-tools)
8. [CI/CD Pipeline](#cicd-pipeline)
9. [Docker & Cloud Deploy](#docker--cloud-deploy)
10. [Project Structure](#project-structure)
11. [Troubleshooting](#troubleshooting)
12. [Metrics](#metrics)

---

## Project Description

A mini social network built with React (frontend) and Go (backend) that implements:

- User registration and authentication
- Post creation, listing, and deletion
- Comment system on posts
- Permission validation (only the author can delete their own content)

This repository represents the full deployment lifecycle of the application: from local development through automated testing, static analysis, containerization, and cloud deployment across QA and production environments.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24 + SQLite |
| Frontend | React 18 + TypeScript |
| Unit Testing | Go testing + Jest |
| E2E Testing | Cypress |
| Static Analysis | SonarCloud |
| CI/CD | GitHub Actions |
| Containerization | Docker (multi-stage builds) |
| Container Registry | GitHub Container Registry (ghcr.io) |
| Cloud Hosting | Render.com (QA + PROD) |

---

## Prerequisites

```bash
# Verify installed versions:
go version     # 1.24 or higher
node --version # 20 or higher
npm --version  # 10 or higher
docker --version # any recent version
```

**Installing Go:**
```bash
# macOS
brew install go

# Ubuntu/Debian
sudo apt install golang-go

# Windows: https://go.dev/dl/
```

**Installing Node.js:**
```bash
# macOS
brew install node

# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Windows: https://nodejs.org/
```

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/CarpinetiOctavio/forum-app-cloud-deploy.git
cd forum-app-cloud-deploy
```

### 2. Install Backend Dependencies

```bash
cd backend
go mod download
cd ..
```

### 3. Install Frontend Dependencies

```bash
cd frontend
npm install
cd ..
```

---

## Running the Project

### Option 1: Manual (recommended for local development)

**Terminal 1 — Backend:**
```bash
cd backend
go run cmd/api/main.go
```

Backend runs at `http://localhost:8080`. Expected output:
```
Server running at http://localhost:8080
Database initialized
```

**Terminal 2 — Frontend:**
```bash
cd frontend
npm start
```

Frontend runs at `http://localhost:3000` and opens automatically in the browser.

### Option 2: Build and Run

**Backend:**
```bash
cd backend
go build -o app cmd/api/main.go
./app
```

**Frontend:**
```bash
cd frontend
npm run build
serve -s build -l 3000  # requires: npm install -g serve
```

---

## Running Tests

### Backend Unit Tests

```bash
cd backend

# Run all tests
go test ./tests/services/... -v

# Run with coverage
go test ./tests/services/... -v -cover -coverpkg=./internal/services/...

# Generate HTML coverage report
go test ./tests/services/... -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -html=coverage.out

# Coverage summary in terminal
go tool cover -func=coverage.out
```

Expected output:
```
=== RUN   TestRegister_Success
--- PASS: TestRegister_Success (0.00s)
...
PASS
coverage: 86.5% of statements in ./internal/services
ok      forum-app-cloud-deploy/tests/services     0.537s
```

### Frontend Unit Tests

```bash
cd frontend

# Run once
npm test -- --watchAll=false

# Run with coverage
npm test -- --coverage --watchAll=false

# Open HTML coverage report
open coverage/lcov-report/index.html
```

Expected output:
```
Test Suites: 8 passed, 8 total
Tests:       39 passed, 39 total
Coverage:    92.44% statements
```

### E2E Tests — Cypress

> Backend and frontend must be running before executing Cypress tests.

```bash
# Terminal 1: backend
cd backend && go run cmd/api/main.go

# Terminal 2: frontend
cd frontend && npm start

# Terminal 3: Cypress
cd frontend

# Interactive mode
npx cypress open

# Headless mode (used in CI/CD)
npx cypress run
```

Expected output:
```
Running:  auth.cy.js        (1 of 4)  ✓ 5 tests passing
Running:  posts.cy.js       (2 of 4)  ✓ 5 tests passing
Running:  comments.cy.js    (3 of 4)  ✓ 4 tests passing
Running:  full-flow.cy.js   (4 of 4)  ✓ 1 test passing

Total: 15 tests passing
```

---

## Quality Tools

### SonarCloud (Static Analysis)

```
URL: https://sonarcloud.io/project/overview?id=CarpinetiOctavio_forum-app-cloud-deploy
Organization: carpinetioctavio
```

Local analysis (optional):
```bash
docker run --rm \
  -e SONAR_HOST_URL="https://sonarcloud.io" \
  -e SONAR_TOKEN="your-token" \
  -v "$(pwd):/usr/src" \
  sonarsource/sonar-scanner-cli
```

### Code Coverage

**Backend:**
```bash
cd backend
go test ./tests/services/... -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -func=coverage.out | grep total
go tool cover -html=coverage.out
```

**Frontend:**
```bash
cd frontend
npm test -- --coverage --watchAll=false
open coverage/lcov-report/index.html
```

---

## CI/CD Pipeline

The pipeline runs automatically on every push or pull request to `main`.

**Stages:**
1. Backend unit tests + coverage
2. Frontend unit tests + coverage
3. SonarCloud static analysis
4. Cypress E2E tests
5. Docker image build (multi-stage)
6. Push to GitHub Container Registry (`ghcr.io`) — tagged with commit SHA
7. Deploy to QA (Render) — automatic
8. Manual approval gate
9. Deploy to PROD (Render) — requires explicit approval

**Quality gates — pipeline aborts if:**
- Backend coverage < 70%
- Frontend coverage < 70%
- SonarCloud Quality Gate fails
- Any unit test fails
- Any E2E test fails

**Trigger pipeline manually:**
```bash
git commit --allow-empty -m "chore: trigger pipeline"
git push
```

---

## Docker & Cloud Deploy

### Building Images Locally

```bash
# Backend
docker build -t forum-backend ./backend

# Frontend
docker build -t forum-frontend ./frontend
```

### Environments

| Aspect | QA | PROD |
|--------|----|------|
| Service | Render.com | Render.com |
| Docker image | Same image (same SHA tag) | Same image (same SHA tag) |
| Deploy | Automatic on passing tests | Requires manual approval |
| Resources | Minimal | Standard |
| Credentials | QA environment variables | PROD environment variables |

### Image Versioning

Images are tagged with the Git commit SHA (`${{ github.sha }}`), not `latest`. This ensures full traceability: every deployed image maps to an exact commit, and rollback is as simple as redeploying a previous tag.

### Secrets

All sensitive values (registry credentials, Render deploy hooks, database credentials) are stored as GitHub Secrets and injected at runtime. Nothing sensitive is hardcoded in Dockerfiles or workflow YAML files.

---

## Project Structure

```
forum-app-cloud-deploy/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                  # Server entry point
│   ├── internal/
│   │   ├── handlers/                    # HTTP handlers
│   │   │   ├── auth_handler.go
│   │   │   ├── post_handler.go
│   │   │   └── utils.go
│   │   ├── services/                    # Business logic (86.5% coverage)
│   │   │   ├── auth_service.go
│   │   │   └── post_service.go
│   │   ├── repository/                  # Data access interfaces
│   │   │   ├── user_repository.go
│   │   │   └── post_repository.go
│   │   ├── models/                      # Data structures
│   │   │   ├── users.go
│   │   │   └── post.go
│   │   ├── database/                    # Database configuration
│   │   │   └── database.go
│   │   └── router/                      # Route configuration
│   │       └── router.go
│   ├── tests/
│   │   ├── services/                    # 35 unit tests
│   │   │   ├── auth_service_test.go
│   │   │   └── post_service_test.go
│   │   └── mocks/
│   │       ├── mock_user_repository.go
│   │       └── mock_post_repository.go
│   ├── Dockerfile                       # Multi-stage build
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── components/                  # React components (92.44% coverage)
│   │   │   ├── Login/
│   │   │   ├── PostList/
│   │   │   ├── CreatePost/
│   │   │   ├── PostDetail/
│   │   │   ├── CommentList/
│   │   │   └── CommentForm/
│   │   ├── services/                    # HTTP services
│   │   │   ├── authService.ts
│   │   │   ├── authService.test.ts
│   │   │   ├── postService.ts
│   │   │   └── postService.test.ts
│   │   ├── types/
│   │   │   └── index.ts
│   │   ├── App.tsx
│   │   └── index.tsx
│   ├── cypress/
│   │   ├── e2e/
│   │   │   └── blog/                    # 15 E2E tests
│   │   │       ├── auth.cy.js
│   │   │       ├── posts.cy.js
│   │   │       ├── comments.cy.js
│   │   │       └── full-flow.cy.js
│   │   └── support/
│   ├── Dockerfile                       # Multi-stage build
│   ├── cypress.config.js
│   ├── package.json
│   └── package-lock.json
│
├── .github/
│   └── workflows/
│       └── ci.yml                       # Full CI/CD pipeline
│
├── docs/
│   ├── decisions/                       # Architecture Decision Records (ADR)
│   └── rules/                           # AI assistant operating instructions
├── sonar-project.properties
└── README.md
```

---

## Troubleshooting

### Backend won't start

```bash
# Check port 8080
lsof -i :8080
kill -9 <PID>

# Verify Go installation
go version

# Clean and reinstall dependencies
cd backend
rm go.sum
go mod tidy
go mod download
```

### Frontend won't start

```bash
# Check port 3000
lsof -i :3000

# Clean and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Cypress tests fail

```bash
# Verify backend and frontend are running
curl http://localhost:8080/api/health
curl http://localhost:3000

# Clear Cypress cache
npx cypress cache clear
npx cypress install

# Run with verbose logs
DEBUG=cypress:* npx cypress run
```

### Pipeline fails in GitHub Actions

```bash
# Check logs at: GitHub > Actions > failed run

# Common causes:
# 1. package-lock.json out of sync
cd frontend
rm package-lock.json
npm install
git add package-lock.json
git commit -m "fix: regenerate package-lock.json"
git push

# 2. Tests failing locally — always run tests locally before pushing
```

---

## Metrics

| Metric | Target | Result | Status |
|--------|--------|--------|--------|
| Backend Coverage | ≥70% | 86.5% | ✅ |
| Frontend Coverage | ≥70% | 92.44% | ✅ |
| Total Tests | — | 89 tests | ✅ |
| SonarCloud Quality Gate | Pass | PASSED | ✅ |
| Issues Resolved | ≥3 | 47 issues | ✅ |
| Code Duplications | <3% | 0.0% | ✅ |

---

**Author:** Octavio Carpineti
**GitHub:** https://github.com/CarpinetiOctavio
**Repository:** https://github.com/CarpinetiOctavio/forum-app-cloud-deploy