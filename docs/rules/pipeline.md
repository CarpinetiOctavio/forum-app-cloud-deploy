# Operating rules — Pipeline (forum-app-cloud-deploy)

## Purpose and authority
This file states the required structure of this repo's CI/CD pipeline.
Read `docs/rules/constraints.md` first — every rule here derives from a
constraint stated there. Docker-specific requirements referenced below are
`docs/rules/docker.md`'s scope, not restated here.

## Trigger — verified against the series' real pattern, not invented
**MUST**: `push: branches: ['feature/**']`.
**MUST**: `pull_request: branches: [staging, main]`
**Why this exact shape, not a guess**: confirmed directly against
`ci-testing@v1.1.0` (`push: ['feature/**']`, `pull_request: [staging, main]`)
and `qa-pipeline@v1.0.0` (`push: ['feature/**', staging, main]`,
`pull_request: [staging, main]`) — this is the series' established
convention: `pull_request` gates the merge with required status checks;
`push` gives fast feedback on feature branches and re-confirms on the
branches that actually receive merges. Neither prior repo conditions any
job by branch internally — every job runs the same way regardless of which
of the three triggered it. This repo follows that, **except** for Stages 7
and 9 below, which are genuinely new — neither `ci-testing` nor
`qa-pipeline` ever deployed anything, so neither had a reason to map a
branch to an environment. This repo does.

**MUST NOT** repeat `cloud-deploy-legacy`'s actual mistake here either:
its trigger was `[main, master, develop]` — three branches, none of them
`staging`, with `deploy-qa` unconditioned by branch and the `qa`
environment's `deployment_branch_policy` left `null`
(`cloud-deploy-legacy-audit-results.md` §4.5). That was wrong in the
opposite direction from a bare `[main]` — too broad, not too narrow — and
both are wrong for the same underlying reason: neither was checked against
what the trigger is actually supposed to gate before being written.

## Stage order — sequential and blocking
Every stage MUST complete successfully before the next starts. A failure
at any stage MUST abort the entire pipeline — no later stage runs. No
stage is skipped, parallelized, or reordered without an ADR documenting
why.

| Order | Stage | Status | Runs on |
|-------|-------|--------|---------|
| 1–4 | Backend tests, frontend tests, SonarCloud analysis, Cypress E2E | Inherited — see below | `feature/**`, `staging`, `main` |
| 5 | Docker image build | New to this repo | `feature/**`, `staging`, `main` |
| 6 | Push to ghcr.io | New to this repo | `feature/**`, `staging`, `main` |
| 7 | Deploy to QA (Render) + smoke test | New — mechanism decided, `ADR-007` | `staging` only |
| 8 | Manual approval gate | New | `main` only |
| 9 | Deploy to PROD (Render) + smoke test | New — same mechanism as Stage 7 | `main` only |

## Stages 1–4 — inherited, verified, not re-decided here
This repo's test suite and quality-gate jobs are what this repo started
with via the initial template copy from `forum-app-qa-pipeline@v1.0.0` —
currently (as of the starter-verification checklist)
`Backend Tests (Go)`, `Frontend Tests (React)`, `Backend Build`,
`Frontend Build`, `SonarCloud Code Analysis`, `Cypress E2E`, and
`Test Summary`. **MUST NOT** modify their logic, coverage thresholds, or
commands — that is `qa-pipeline`'s scope, not this repo's, per `CLAUDE.md`.
For the reasoning behind this exact shape (why these two extra jobs, why
gated this way, why parallelized where they are), see `qa-pipeline`'s own
[`ADR-004-pipeline-extension`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/blob/main/docs/decisions/ADR-004-pipeline-extension.md)
— not duplicated here, and not kept as a local file in this repo's own
`docs/decisions/` per the cross-repo referencing convention
(`docs/rules/documentation.md`).

**This is not the same claim `cloud-deploy-legacy`'s own `pipeline.md`
made about itself** ("Stages 1–4 correspond to the pipeline inherited from
forum-app-qa-pipeline") — that claim was checked directly against both
repos' real `ci.yml` and found false: different action-pinning style,
different step names, different coverage-check implementation, different
job names entirely in places. This repo's version of the same claim is
true because it was verified against the real, current state of this
repo's own `ci.yml` — not assumed by analogy to what `-legacy` said about a
different repo.

**Before touching anything here**: MUST re-verify the current real job
names and structure directly (`grep`/`cat` on this repo's actual `ci.yml`),
since names may have changed since this document was written. Never
assume this table stays accurate without checking.

## Stage 5 — Docker image build
**MUST**: build both images only after stages 1–4 have all passed. **MUST**
follow every rule in `docs/rules/docker.md`. **MUST** tag both images with
`${{ github.sha }}`.

## Stage 6 — Push to ghcr.io
**MUST**: push to `ghcr.io/carpinetioctavio`, tagged with the commit SHA
only. **MUST**: declare `permissions: packages: write` at both the
workflow level and the job level in `ci.yml` — a job-level-only grant is
not sufficient (`cloud-deploy-legacy-transferable-knowledge-results.md`
§2 documents this exact two-part requirement and the bootstrap ordering
it implies: the first image push to a not-yet-existing package must happen
manually, before the package can be linked to this repository, before the
pipeline can push to it with `GITHUB_TOKEN`).

## Stages 7 and 9 — branch mapping and mechanism, both decided

**Branch mapping — this repo's own, not inherited from anywhere**:
- **Stage 7 MUST** run `if: github.ref == 'refs/heads/staging'` — a merge
  into `staging` is what promotes code to the QA environment.
- **Stage 9 MUST** run `if: github.ref == 'refs/heads/main'` — a merge into
  `main` is what's eligible for PROD, still gated by the manual approval
  below regardless of branch.
- Neither `ci-testing` nor `qa-pipeline` had two deploy environments to map
  branches onto — this mapping exists because this repo does, not because
  it was copied from anywhere.
- **Both stages end with a smoke test, not just a deploy call** — see
  `docs/rules/smoke-testing.md`. A deploy mechanism reporting success is not
  the same claim as the deployed service actually serving requests; Stage 7
  and Stage 9 are not complete until that's verified against the real,
  running container, not assumed from the deploy call's own status code.

**Mechanism — decided, see `ADR-007` for the full investigation and
rationale.** `cloud-deploy-legacy` documented a Render deploy hook
mechanism in four separate files and actually implemented a different one
(the Render REST API, account-scoped `RENDER_API_KEY`) without ever
reconciling the two (`cloud-deploy-legacy-audit-results.md` §4). This repo
decided explicitly, verified directly against Render's current API
documentation rather than assuming either of `-legacy`'s two conflicting
answers:

- **Stage 7 and Stage 9 MUST trigger their deploys via Render deploy
  hooks** — `POST` to the service's hook URL with the `imgURL` query
  parameter set to this pipeline run's exact `${{ github.sha }}`-tagged
  image. **MUST NOT** omit `imgURL` — doing so silently redeploys whatever
  tag the service last had configured, defeating `ADR-003`'s guarantee.
  Four hook URLs, one per service (`forum-backend-qa`, `forum-frontend-qa`,
  `forum-backend-prod`, `forum-frontend-prod`).
- **MUST** poll the Render API's read-only "Retrieve deploy" endpoint
  (`GET /v1/services/{id}/deploys/{deployId}`), using the deploy ID the
  hook's own response returns, to determine when the deploy has actually
  finished before running the smoke test against it — not a fixed delay.
  **MUST NOT** use `RENDER_API_KEY` for anything other than this `GET`
  call — no write operation through this pipeline is authorized to use it,
  per `ADR-007`'s Consequences.
- **MUST** treat a `202` hook response with no deploy ID (the
  overlapping-deploys `Wait`-policy case) as "a deploy is already in
  progress," with a retry/backoff — not as a failure.

`ADR-007` documents why: both mechanisms turned out to be functionally
equivalent for targeting an exact tag (an assumption checked and corrected
mid-decision, not assumed from either of `-legacy`'s two answers), which
made the deciding factor blast radius if a secret leaks — four
narrowly-scoped hook URLs versus one Render API key confirmed, against
Render's own current documentation, to have no narrower scope available at
all.

## GitHub environments
| Environment | Protection rule | `deployment_branch_policy` |
|---|---|---|
| `qa` | None (automatic deploy) | MUST be explicitly configured to allow only `staging` — `cloud-deploy-legacy-audit-results.md` §4/§5 confirmed `null` (not "restricted," genuinely unset) was part of why an auto-deploy from a non-`main` branch was possible there |
| `prod` | Required reviewer: repository owner, enforced via GitHub environment protection | MUST be explicitly configured to allow only `main` |

**Why "via environment protection, not pipeline logic alone"**: restated
from `constraints.md` — a gate implemented only as an `if:` condition in
`ci.yml` is editable by the same push that would bypass it. GitHub's
environment-level rule is enforced outside the workflow file. This is a
second, independent layer on top of the `if: github.ref == ...` conditions
above — it does not replace them; either one alone would leave a gap the
other closes.

## Before writing or modifying `ci.yml`
1. Read the current, real `ci.yml` in full — do not assume this document's
   stage table is still accurate.
2. Report the current real job names and structure.
3. Confirm the Stage 7/9 mechanism decision has been made and documented in
   its own ADR before writing either deploy stage.
4. Confirm both Dockerfiles exist and build successfully before writing
   Stage 5.
5. Confirm `sonar-project.properties` points at this repo's own SonarCloud
   project (not `qa-pipeline`'s, still inherited unmodified as of the last
   checked state) before relying on Stage 3 passing meaning anything.
6. Report the proposed `ci.yml` change and wait for approval before writing
   it.