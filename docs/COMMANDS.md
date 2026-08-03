# COMMANDS.md — forum-app-qa-pipeline

Commands that work today, against this repo's current state (inherited from
`forum-app-ci-testing@v1.1.0`). Sections for SonarCloud, Cypress, and the
coverage gate are added here once those pieces actually exist in this repo —
see `docs/rules/documentation.md`.

## Backend

```bash
cd backend

# Run the server (http://localhost:8080)
go run cmd/api/main.go

# Run all unit tests
go test ./tests/services/... -v

# Run with coverage (measures internal/services/ only — see ADR-001 once written)
go test ./tests/services/... -v -cover -coverpkg=./internal/services/...

# Generate an HTML coverage report
go test ./tests/services/... -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -html=coverage.out

# Coverage summary in terminal
go tool cover -func=coverage.out
```

## Frontend

```bash
cd frontend

# Run the dev server (http://localhost:3000)
npm start

# Run tests once
npm test -- --watchAll=false

# Run tests with coverage
npm test -- --coverage --watchAll=false
```

## Git / branching

```bash
# Feature branches merge into staging, staging merges into main
git checkout -b feature/<name> staging

# Trigger the CI pipeline manually (empty commit)
git commit --allow-empty -m "trigger pipeline"
git push
```