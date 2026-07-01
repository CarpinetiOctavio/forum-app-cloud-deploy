# Commands Reference

Quick reference for running, testing, and operating the forum-app-cloud-deploy
project. Each command includes context (when and why to use it) and the
expected output to confirm it worked.

---

## Before you start — Clean up your environment

Run these commands before cloning and starting the project on any machine.
They ensure that ports are free, no stale containers are running, and Docker
is operational.

---

**1. Kill any process using port 8080 (backend)**
```bash
lsof -ti :8080 | xargs kill -9 2>/dev/null; echo "done"
```
Use when starting the backend locally and port 8080 is already in use —
typically a previous run left the process alive.
Expected: `done` with no other output. If nothing was running on that port,
`kill` has nothing to do and `2>/dev/null` suppresses the error silently.

**2. Kill any process using port 3000 (frontend)**
```bash
lsof -ti :3000 | xargs kill -9 2>/dev/null; echo "done"
```
Use when starting the React dev server and port 3000 is already taken.
Expected: same as above — `done` with no errors.

**3. Verify both ports are free**
```bash
lsof -i :8080 -i :3000
```
Use immediately after the kill commands to confirm nothing is still listening.
Expected: no output at all. Any line in the output means a process is still
holding that port.

**4. Stop and remove any running project containers**
```bash
docker ps --filter "name=forum" --format "{{.Names}}"
```
Use to see which project containers are currently running.
Expected: empty output if none are running. If container names appear, stop
and remove them:
```bash
docker ps --filter "name=forum" -q | xargs docker rm -f 2>/dev/null; echo "done"
```
Expected: `done`. Each removed container prints its ID on a separate line
before `done`.

**5. Verify Docker Desktop is running**
```bash
docker info --format "Server Version: {{.ServerVersion}}"
```
Use before any `docker build` or `docker run` command.
Expected: `Server Version: 27.x.x` (or similar). If Docker Desktop is not
running you will get `Cannot connect to the Docker daemon` — open Docker
Desktop and wait for it to finish starting before retrying.

**6. Remove all local project images (optional — use before a clean demo)**
```bash
docker rmi forum-backend forum-frontend 2>/dev/null; echo "done"
```
Use when you want to demonstrate a full build from scratch. After running this, `docker build` downloads base images and compiles everything from zero — nothing is cached.
Expected: `done`. Each removed image prints its ID before `done`. If the images didn't exist, `2>/dev/null` suppresses the error silently.

---

## 1. Git

**Clone the repository**
```bash
git clone https://github.com/CarpinetiOctavio/forum-app-cloud-deploy.git
cd forum-app-cloud-deploy
```
Use when setting up the project on a new machine.
Expected: repository cloned, `cd` puts you in the project root.

**Stage specific files and commit**
```bash
git add path/to/file
git commit -m "your message"

---
Empty commit:
git commit --allow-empty -m "chore: trigger full pipeline run"
```
Use after making a change. Prefer staging specific files over `git add .`
to avoid accidentally committing secrets or binaries.
Expected: `[master abc1234] your message` — one new commit created.

**Push to remote**
```bash
git push origin main
```
Use to push local commits to the GitHub remote. This triggers the CI/CD
pipeline on every push to `main`.
Expected: `master -> main` confirmation line. The pipeline run appears
immediately in the GitHub Actions tab.

**Reviewer comment for PROD**
```bash
chore: verify full pipeline end-to-end after fresh clone
```
---

## 2. Run locally

**Start the backend**
```bash
cd backend
go run cmd/api/main.go
```
Use to run the Go API server locally for development or E2E testing.
Backend must be running before starting Cypress.
Expected: `database initialized successfully` followed by
`Server running on :8080`.

**Start the frontend**
```bash
cd frontend
npm start
```
Use to run the React dev server locally.
Expected: browser opens at `http://localhost:3000` and the forum UI loads.

---

## 3. Run tests locally

**Backend unit tests with coverage**
```bash
cd backend
go test ./tests/services/... -v -cover -coverpkg=./internal/services/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```
Use to verify all backend tests pass and check the coverage percentage
before pushing.
Expected: all tests print `--- PASS`, final line shows
`total: (statements) 97.3%` or similar — must be above 70%.

**Frontend unit tests with coverage**
```bash
cd frontend
npm test -- --coverage --watchAll=false
```
Use to run all Jest tests and generate the coverage report.
Expected: `Tests: 47 passed, 47 total` and a coverage table showing
≥70% across all metrics.

**Cypress E2E tests**

Backend and frontend must both be running first (see section 2).
```bash
cd frontend
npx cypress run
```
Use to run the full E2E suite headlessly.
Expected: `15 passing` — all specs in `cypress/e2e/blog/` green.

To open the Cypress interactive runner instead:
```bash
npx cypress open
```

---

## 4. Docker

**Build backend image locally**
```bash
docker build -t forum-backend ./backend
```
Use to verify the backend Dockerfile builds correctly.
Expected: `exporting to image` → `naming to ... forum-backend:latest done`
with no errors. Build takes ~60 seconds the first time (CGO compilation).

**Build frontend image locally**
```bash
docker build -t forum-frontend ./frontend
```
Use to verify the frontend Dockerfile and entrypoint build correctly.
Expected: `Compiled successfully` in the builder stage, then
`naming to ... forum-frontend:latest done`.

**Run backend container locally**
```bash
docker run --rm -p 8080:8080 forum-backend
```
Use to verify the binary starts correctly inside the container.
Expected: `database initialized successfully` and `Server running on :8080`.
Test with `curl http://localhost:8080/api/posts` — should return `[]`.

**Run frontend container locally**
```bash
docker run --rm -p 3000:80 forum-frontend
```
Use to verify the nginx entrypoint replaces the URL placeholder and serves
the app.
Expected: container starts with no errors. `http://localhost:3000` shows
the forum UI. Check the replacement ran:
```bash
docker run --rm -e REACT_APP_API_URL=https://example.com/api -d --name test-frontend forum-frontend
docker exec test-frontend sh -c "grep -c 'example.com' /usr/share/nginx/html/static/js/main.*.js"
docker stop test-frontend
```
Expected output: `1` — the placeholder was replaced with the injected URL.

**Build for linux/amd64 (required when pushing from Apple Silicon)**
```bash
docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:latest \
  --push ./backend
```
Use when pushing images manually to ghcr.io from a Mac with M-series chip.
Render runs on amd64 — omitting `--platform` produces an incompatible image.
Expected: `pushing layers done` and `naming to ghcr.io/...` with no errors.

---

## 5. Trigger pipeline manually from GitHub Actions

Pushing any commit (including an empty one) triggers the pipeline
automatically. Use an empty commit to trigger without a code change:
```bash
git commit --allow-empty -m "chore: trigger pipeline run"
git push origin master:main
```
Expected: a new run appears immediately at
`github.com/CarpinetiOctavio/forum-app-cloud-deploy/actions`.

To re-run only the failed jobs of an existing run without pushing:
1. Open the failed run in GitHub Actions
2. Top-right → **Re-run jobs** → **Re-run failed jobs**

---

## 6. View pipeline logs in real time

**Via GitHub CLI**
```bash
gh run list --repo CarpinetiOctavio/forum-app-cloud-deploy --limit 5
```
Shows the five most recent runs with their IDs and status.
Expected: table with columns `STATUS`, `NAME`, `WORKFLOW`, `ID`.

```bash
gh run watch <RUN_ID> --repo CarpinetiOctavio/forum-app-cloud-deploy
```
Streams live log output for the run until it completes.
Expected: job names appear as they start; pass/fail shown inline.

```bash
gh run view <RUN_ID> --log --repo CarpinetiOctavio/forum-app-cloud-deploy
```
Dumps the full log of a completed run to stdout.

**Via browser**
Open `github.com/CarpinetiOctavio/forum-app-cloud-deploy/actions`, click
the run, then click any job to see its live log.

---

## 7. Approve PROD gate in GitHub

The `deploy-prod` job pauses until a required reviewer approves it.

1. Open the pipeline run in GitHub Actions
2. Click the `Deploy to PROD` job — it shows **"Waiting for review"**
3. Click **Review deployments**
4. Check the `prod` environment box
5. Click **Approve and deploy**

Expected: the `deploy-prod` job resumes immediately and deploys both
`forum-backend-prod` and `forum-frontend-prod` on Render.

---

## 8. Verify Render deployed correctly

**Check deploy status in Render dashboard**

Open `dashboard.render.com`, select each service, and check the
**Events** tab. A successful deploy shows:
`Deploy live for <service-name>`

**Verify the backend is responding**
```bash
curl http://localhost:8080/api/posts
```
Expected: `[]` (empty array) — database is fresh on each deploy (SQLite
on ephemeral filesystem, see ADR-005).

**Verify the frontend is serving the app**

Open `https://<your-frontend-qa-url>` in a browser.
Expected: forum UI loads and the Login / Register page is displayed.

**Verify the correct SHA was deployed**

In Render → service → **Events**, the deploy event shows the image URL
including the commit SHA tag:
`ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:<sha>`

Cross-check the SHA against the GitHub Actions run that triggered the
deploy.

**Verify all 4 services are responding**

Useful when the Render dashboard does not show all services in the main view.

Backend QA:
```bash
curl https://forum-app-cloud-deploy-backend.onrender.com/api/posts
```
Expected: `[]`

Backend PROD:
```bash
curl https://forum-backend-prod.onrender.com/api/posts
```
Expected: `[]`

Frontend QA:
```bash
curl -I https://forum-app-cloud-deploy-frontend.onrender.com
```
Expected: `HTTP/2 200`

Frontend PROD:
```bash
curl -I https://forum-frontend-prod.onrender.com
```
Expected: `HTTP/2 200`

---

## 9. Intentionally fail the pipeline

This demonstrates that the quality gates block the pipeline before any
Docker image is built or deployed. Three failure scenarios are available,
each aborting the pipeline at a different stage.

---

### Scenario A — Break a backend unit test (aborts at stage 1)

**Step 1 — Uncomment the intentional failure line**

In `backend/tests/services/auth_service_test.go`, find `TestRegister_Success`
and uncomment the line marked `// line in order to show how to break the pipeline`.

**Step 2 — Verify locally before pushing**
```bash
cd backend
go test ./tests/services/... -v 2>&1 | head -20
```
Expected: `FAIL` on `TestRegister_Success`, `FAIL` on the package.

**Step 3 — Push**
```bash
git add backend/tests/services/auth_service_test.go
git commit -m "test: intentional backend failure to demonstrate pipeline gate"
git push origin main
```

**Step 4 — Observe**

`Backend Tests (Go)` fails. All downstream jobs (`SonarCloud`, `Cypress E2E`,
`Docker Build & Push`, `Deploy to QA`, `Deploy to PROD`) are skipped — they
never run. No image is built or pushed to ghcr.io.

**Step 5 — Restore**

Comment the line back and push:
```bash
git add backend/tests/services/auth_service_test.go
git commit -m "chore: clean up intentional test failure after pipeline gate demonstration"
git push origin main
```

---

### Scenario B — Break a frontend unit test (aborts at stage 2)

**Step 1 — Uncomment the intentional failure line**

In `frontend/src/components/CreatePost/CreatePost.test.tsx`, find the line
marked `// line in order to show how to break the pipeline` and uncomment it.

**Step 2 — Verify locally before pushing**
```bash
cd frontend
npm test -- --watchAll=false 2>&1 | tail -20
```
Expected: `FAIL` on `CreatePost.test.tsx`, test suite fails.

**Step 3 — Push**
```bash
git add frontend/src/components/CreatePost/CreatePost.test.tsx
git commit -m "test: intentional frontend failure to demonstrate pipeline gate"
git push origin main
```

**Step 4 — Observe**

Backend tests pass but `Frontend Tests (React)` fails. `SonarCloud`,
`Cypress E2E`, `Docker Build & Push`, and deploy jobs are skipped.

**Step 5 — Restore**

Comment the line back and push with message:
`chore: clean up intentional test failure after pipeline gate demonstration`

---

### Scenario C — Break an E2E test (aborts at stage 4)

**Step 1 — Uncomment the intentional failure line**

In `frontend/cypress/e2e/blog/full-flow.cy.js`, find the line marked
`// line in order to show how to break the pipeline` and uncomment it.

**Step 2 — Verify locally before pushing**
```bash
cd frontend
npx cypress run --spec "cypress/e2e/blog/full-flow.cy.js"
```
Expected: `1 failing` on the full flow test.

**Step 3 — Push**
```bash
git add frontend/cypress/e2e/blog/full-flow.cy.js
git commit -m "test: intentional E2E failure to demonstrate pipeline gate"
git push origin main
```

**Step 4 — Observe**

Backend and frontend unit tests pass, SonarCloud passes, but `Cypress E2E`
fails. `Docker Build & Push`, `Deploy to QA`, and `Deploy to PROD` are
skipped — no image is built despite all unit tests passing.

**Step 5 — Restore**

Comment the line back and push with message:
`chore: clean up intentional test failure after pipeline gate demonstration`

---

## 10. Rollback in Render

If a PROD deploy introduces a regression, roll back to the last known
good image. The SHA tag on every image makes this safe — old images are
never overwritten in ghcr.io.

**Step 1 — Find the last known good SHA**
```bash
git log --oneline -10
-
git log -1 --format="%H" <SHORT_SHA>
```

Identify the commit SHA of the last good release. Alternatively, find it
in GitHub Actions → the last green `Deploy to PROD` run → step
`Deploy backend to PROD` log shows the full SHA in the image URL.

**Step 2 — Trigger a rollback deploy via Render API**

Replace `SHA_DEL_COMMIT` with the actual commit SHA (full 40-character hash).

Backend PROD:
```bash
curl -X POST "https://api.render.com/v1/services/srv-d8vf577lk1mc738suglg/deploys" \
  -H "Authorization: Bearer $RENDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"imageUrl\":\"ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:SHA_DEL_COMMIT\"}"
```

Frontend PROD:
```bash
curl -X POST "https://api.render.com/v1/services/srv-d8vf69uq1p3s73b3rkvg/deploys" \
  -H "Authorization: Bearer $RENDER_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"imageUrl\":\"ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend:SHA_DEL_COMMIT\"}"
```

Expected: HTTP `201 Created` response from the Render API. The service
Events tab shows a new deploy starting with the previous SHA image.

**Step 3 — Verify rollback**

Once the deploy completes, repeat the verification steps from section 8.
The Events tab should show the previous SHA in the image URL.

**Why rollback is always possible:** images are tagged with commit SHA
and never deleted or overwritten in ghcr.io (see ADR-003). Any
previously deployed SHA can be redeployed at any time without rebuilding.
