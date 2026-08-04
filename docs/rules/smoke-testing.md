# Operating rules — Smoke tests (forum-app-cloud-deploy)

## Purpose
This file states the requirement for verifying a deployment succeeded, as
a concern separate from verifying the code is correct. Read
`docs/rules/pipeline.md` first — these checks are the final step inside
Stage 7 (deploy to QA) and Stage 9 (deploy to PROD), not new numbered
stages of their own.

## Why this is a different failure domain than Stages 1–4
Stages 1–4 (backend tests, frontend tests, SonarCloud, Cypress) verify that
the *code* behaves correctly, against a local or CI-provisioned
environment. None of them can catch a failure that only exists in the
*deployed* container: a misconfigured environment variable, a port
mismatch, a container that crashes on startup on Render's actual
infrastructure but never got exercised that way in CI, a database path
that doesn't resolve the way it did locally. A pipeline that stops at "the
deploy step returned success" is trusting the deploy mechanism's own
report of itself — exactly the class of failure
`docs/rules/verification.md` already warns against in general
("don't accept 'done' at face value"), applied here to infrastructure
instead of to a status report.

This is the same distinction container orchestration platforms encode as
liveness/readiness probes (Kubernetes' terminology, even though this repo
doesn't run Kubernetes — the concept transfers: a process existing is not
the same claim as a process being ready to serve traffic) and the same
distinction continuous-delivery practice calls a post-deploy smoke test —
a minimal, fast check against the real running thing, not a re-run of the
correctness suite.

## Requirement
**MUST**: after Stage 7 (deploy to QA) reports success, run an automated
HTTP check directly against the real, newly deployed QA service before
that job is considered complete. **MUST**: run the equivalent check
against PROD after Stage 9.

**MUST NOT**: treat "the deploy mechanism returned success" (whatever
Stage 7/9's chosen mechanism, per `pipeline.md`, reports) as sufficient
evidence the deployed service is actually serving requests correctly —
the deploy call succeeding means Render/the API accepted the request, not
that the container is healthy once it starts.

## What the check verifies — to be confirmed against real code, not assumed
**MUST**: verify, before writing the actual check, whether this repo's
inherited backend (from `qa-pipeline@v1.0.0`) exposes a dedicated health
endpoint. If it does, the smoke test hits that endpoint. If it does not,
the smoke test hits a real, already-existing read endpoint
(`GET /api/posts` is the known candidate, per this repo's own audit
history) and treats "200, valid JSON" as the passing signal — **MUST NOT**
add a new health endpoint to the inherited backend to make this easier;
that would be modifying `qa-pipeline`'s inherited code, out of scope per
`CLAUDE.md`.

**MUST**: check both the backend's own URL directly and the frontend's
served page — a passing backend check does not verify the frontend
container started correctly, and the two are separate containers, separate
potential failure points.

## Failure behavior
**MUST**: a failing QA smoke test blocks Stage 8 from being meaningfully
actionable — at minimum, the pipeline's own output must make a failing
smoke test impossible to miss before a human approves promotion to PROD.
GitHub's approval gate is a manual action Octavio controls; this
requirement is about the information available to him at that moment, not
about the pipeline technically preventing a click.

**MUST**: a failing PROD smoke test is the trigger condition for
`docs/rules/deployment.md`'s rollback procedure — not a separate incident
process invented ad hoc when it happens.

## Before implementing this
1. Confirm whether a health endpoint exists in the inherited backend, per
   the requirement above, and report the finding.
2. Report the proposed check (endpoint, expected response, timeout) for
   both backend and frontend before writing it into `ci.yml`.