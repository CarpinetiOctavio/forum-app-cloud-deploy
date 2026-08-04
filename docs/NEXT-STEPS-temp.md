# Next steps — temporary, delete once complete

This file is scratch, not part of this repo's permanent documentation
structure. Delete it once the plan below is fully implemented — it should
never be a long-lived doc alongside `docs/rules/` and `docs/decisions/`.

**Correction (re-verified 2026-08-04):** this file's step 1 and
`ADR-000`'s Consequences previously stated the backend test suite as
"47/47 passing." Re-run directly with a clean test cache
(`go test ./tests/services/... -v -count=1`, `--- PASS` count checked by
hand): the real, current count is **50/50 passing, 88.2% coverage**. The
test files did not change between this repo's initial commit and its
module rename — only import paths did — so 47 was already wrong when
first written here, not a regression from anything done since. Corrected
per `docs/rules/verification.md`.

## Why `README.md`, `docs/COMMANDS.md`, `docs/SETUP.md` are empty right now

Emptied deliberately, not lost or broken. All three still had
`forum-app-qa-pipeline`'s own content (wrong course scope, wrong pipeline
description, wrong commands) inherited from the template copy — keeping
that in place risked a new session reading them and treating `qa-pipeline`'s
TP7 content as if it described this repo. Per `docs/rules/documentation.md`'s
own rule ("documentation follows implementation, not the other way
around"), each gets written for real once there's real implementation to
document, not before:

- **`README.md`** — write once the app is containerized and deployed, or at
  minimum once Docker images build successfully locally. Needs, at
  minimum: this repo's real place in the series (TP8, third repo, built
  from `qa-pipeline@v1.0.0`), the real tech stack, real Docker/run
  commands, and — once Render exists — the real Live Environments section.
  Do **not** carry over any number from `cloud-deploy-legacy`'s README
  (e.g. the "Issues Resolved: 47" boilerplate `cloud-deploy-legacy-audit-results.md`
  §6 confirmed was never real) — every claim gets verified against this
  repo's own real state before being written.
- **`docs/COMMANDS.md`** — write incrementally, one section at a time, as
  each real command becomes runnable (local Docker build, then local
  Render deploy commands once services exist, etc.). Never write a command
  that doesn't work yet as a placeholder.
- **`docs/SETUP.md`** — write once real Render services and GitHub secrets
  exist, documenting the actual setup performed, not a plan for it.
  `cloud-deploy-legacy`'s own `SETUP.md` is a reasonable structural
  reference for what sections this needs (per `ADR-007`'s real secret
  names now, not `-legacy`'s) — read for structure, not copied for content.

## Docker implementation plan — in order, each step verified before the next

**1. Backend source fix (`ADR-005`).** `backend/cmd/api/main.go`: replace
the hardcoded `database.InitDB("./database.db")` with an
`os.Getenv("DATABASE_PATH")` read, falling back to the same
`./database.db` value. Verify: `go build ./...`, `go test
./tests/services/... -cover -coverpkg=./internal/services/...` — must
still be 50/50 passing, unchanged coverage. This does not touch
`internal/services/` — see `ADR-005`'s own scope justification if that
constraint gets questioned.

**2. Frontend source fix (`ADR-006`).** `frontend/src/services/postService.ts`
and `authService.ts`: replace the hardcoded `const API_URL =
'http://localhost:8080/...'` with a `process.env.REACT_APP_API_URL ||
'http://localhost:8080'` base. Verify: `npm test -- --coverage
--watchAll=false` — must still be passing at whatever the real, currently
verified frontend count is (not assumed from this file's own prior,
uncorrected "47/47" claim on the backend side — see the note at the top of
this file).

**3. Backend Dockerfile.** Per `docs/rules/docker.md` in full — multi-stage
(`golang:X-alpine` builder, `alpine` final, not `scratch`, per Rule 2's
CGO/`go-sqlite3` finding — confirm the exact Go/Alpine versions to pin
against what's actually in `go.mod` first), non-root user with the
`WORKDIR`-then-`chown`-then-`COPY` order from
`cloud-deploy-legacy-transferable-knowledge-results.md` §3 (skip the
two-step bug already documented there, don't rediscover it), pinned base
image versions (Rule 6). Verify: `docker build -t forum-backend
./backend` succeeds; `docker run --rm -p 8080:8080 forum-backend` starts
and `curl localhost:8080/api/posts` returns `[]`.

**4. Frontend Dockerfile.** Per `docs/rules/docker.md` and `ADR-006`'s
exact mechanism — `node:X-alpine` builder with `ARG
REACT_APP_API_URL=__REACT_APP_API_URL__`, `nginx:X-alpine` final stage,
`docker-entrypoint.sh` performing the placeholder `sed` substitution
before `nginx -g 'daemon off;'`. Verify: `docker build -t forum-frontend
./frontend` succeeds; run with a test `REACT_APP_API_URL`, confirm the
built JS actually contains the injected value (grep the served bundle),
not just that the container starts.

**5. Local verification of both together.** Both containers running
locally, frontend pointed at the local backend via `REACT_APP_API_URL`,
confirm a real request round-trips. This is the last step before touching
CI — don't add Stage 5/6 to `ci.yml` until both Dockerfiles are proven
locally.

**6. `ci.yml` Stage 5/6 (Docker build + push to ghcr.io).** Per
`docs/rules/pipeline.md`. Bootstrap ordering from
`cloud-deploy-legacy-transferable-knowledge-results.md` §1/§2 applies
here: first push must be manual (`--platform linux/amd64` — required even
on non-Apple-Silicon dev machines if the pipeline itself doesn't push
first), packages linked to this repository afterward, workflow-level
*and* job-level `permissions: packages: write` both required.

**7. Render services + `docs/SETUP.md` (real, not the qa-pipeline
leftover).** Four services per `ADR-007`'s naming, four deploy hook
secrets plus `RENDER_API_KEY`, two GitHub environments (`qa` no
protection, `prod` required reviewer). Write `SETUP.md` as this actually
gets done, step by step — not in advance of doing it.

**8. `ci.yml` Stage 7-9 (deploy QA, approval gate, deploy PROD) +
`docs/rules/smoke-testing.md`'s requirement.** Per `ADR-007`: deploy hooks
with `imgURL` set to the real SHA, `RENDER_API_KEY` used only for the
read-only "Retrieve deploy" status poll feeding the smoke test. Confirm
the smoke test actually fails loud if pointed at a broken deploy before
trusting it as a real gate — same "verify the check itself" discipline as
everything else in this series.

## After step 8

`README.md` gets its final real content once there's a real, running
system to describe. Delete this file once step 8 is done and `README.md`
is written for real.
