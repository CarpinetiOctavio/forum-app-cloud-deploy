# Commands

Commands for this repository. The test and build commands are the ones `ci.yml` runs; what each command needs is in [`SETUP.md`](SETUP.md).

## Backend

```bash
cd backend
go run cmd/api/main.go     # http://localhost:8080

# Tests and coverage; ci.yml fails below 70% (qa-pipeline's ADR-001)
go test ./tests/services/... -v -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -func=coverage.out
```

## Frontend

```bash
cd frontend
npm start                                   # http://localhost:3000
npm test -- --coverage --watchAll=false     # fails below the 70% threshold in package.json
npm run build
```

## Cypress

With the backend and the frontend running in two other terminals:

```bash
cd frontend
npx cypress run
```

## Docker

```bash
docker build -t forum-backend ./backend
docker run --rm -p 8080:8080 forum-backend

docker build -t forum-frontend ./frontend
docker run --rm -p 3000:8080 -e REACT_APP_API_URL=http://localhost:8080 forum-frontend
```

The frontend container listens on 8080 and is mapped here to 3000; `REACT_APP_API_URL` is the backend URL its bundle calls.

## Registry

The first image of each package, before the pipeline can push to it (`SETUP.md`, Container registry). `docker login` prompts for a token with `write:packages`:

```bash
docker login ghcr.io -u <github-user>
docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:<commit-sha> --push ./backend
docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend:<commit-sha> --push ./frontend
```

## Pipeline

Pushes to `feature/**` run stages 1–6; a pull request into `staging` or `main` runs them on the merge; the merge into `main` builds, deploys to QA and waits for approval before PROD ([ADR-009](decisions/ADR-009-single-build-trigger.md)). The digests each run deployed are in the `Image:` lines of its deploy jobs.

```bash
# Run the pipeline on a feature branch without changes
git commit --allow-empty -m "trigger pipeline"
git push
```

## Smoke tests

The same checks the deploy jobs run, against any deployed service:

```bash
bash .github/scripts/smoke-test.sh <backend-url>/api/posts json
bash .github/scripts/smoke-test.sh <frontend-url> frontend <backend-url>
```

## Rollback

Redeploys two digests that already reached PROD from `main`: QA first, then PROD after approval ([ADR-007](decisions/ADR-007-render-deploy-mechanism.md), amendment of 2026-11-23).

```bash
gh workflow run rollback.yml --ref main \
  -f backend_digest=sha256:<64 hex> \
  -f frontend_digest=sha256:<64 hex>
```

The run validates both digests before calling any deploy hook, and PROD waits for approval on the run's page.
