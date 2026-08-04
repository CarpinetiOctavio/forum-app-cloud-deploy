# Audit — forum-app-cloud-deploy (pre-`-legacy` rename)

**Date:** 2026-08-03

Full audit of `forum-app-cloud-deploy` ahead of renaming it
`forum-app-cloud-deploy-legacy` and building a new `cloud-deploy` from
`forum-app-qa-pipeline@v1.0.0` — same treatment already applied to
`forum-app-qa-pipeline-legacy` before it. This document compiles two rounds
of work: an initial pass, and a second, checklist-driven pass (working
documents `cloud-deploy-audit-checklist.md` and
`cloud-deploy-live-exposure-checklist.md`, both closed and reviewed by
Octavio) that re-verified every finding against running code, live
services, and external APIs rather than static reading alone. Every claim
below is backed by a command actually run, a file/line reference, or a live
request actually sent — not inferred from documentation. No file in this
repo was modified as part of this audit except the creation of this
document and its companion, `cloud-deploy-legacy-transferable-knowledge-results.md`.

---

## 0. Executive summary

- **This repo's backend does not descend from `qa-pipeline@v1.0.0`**,
  despite `docs/rules/constraints.md` stating it does. Diffed directly
  against real tagged snapshots of all three prior repos: this backend's
  business logic is a near-exact match for `ci-testing@v1.0.0`
  (43 non-comment lines of difference — module path and a constants
  refactor, nothing behavioral), not `qa-pipeline@v1.0.0` (181 lines of
  difference, almost entirely `EditPost`/`EditComment`, which don't exist
  here) and not `ci-testing@v1.1.0` (71 lines of difference on
  `auth_service.go` alone, from the hardening pass this repo never
  received). This repo's own git history opens on 2026-06-24 with a single
  commit that imports an entire pre-existing `tp06-testing`-named project
  wholesale — the real coursework development happened before this
  repository's git history begins.
- **Both test suites pass with real, reproducible coverage matching the
  README** — re-run twice across two sessions: backend 47/47 at 97.3%,
  frontend 47/47 at ~97.7%. All 15 Cypress tests were read individually
  (not just grepped) and then actually executed live against a real local
  backend+frontend: 15/15 pass, and confirmed line-by-line that none of
  them ever reaches the backend — every network call, in every spec, is
  `cy.intercept()`-mocked.
- **The authorization gap generalizes exactly like `qa-pipeline-legacy`,
  and is now confirmed exploitable live, twice, by two independent
  parties**: `DeletePost`'s authorship check lives in the service layer
  (real code, genuinely enforced); `DeleteComment`'s lives in the
  repository layer (raw SQL), is excluded from coverage measurement by
  `sonar-project.properties`, and — like `DeletePost` — only checks whether
  a client-supplied number matches a stored owner, never whether the client
  proved control of that identity. See §7.
- **`ADR-008`'s stale coverage numbers (86.5%/92.44%) were real when
  written**, not fabricated: introduced 2026-06-24, cited accurately in
  `ADR-008` on 2026-07-02 at 18:09, and made stale 51 minutes later by a
  different commit that updated `testing.md`/the README without touching
  the ADR.
- **The documented deploy mechanism (Render deploy hooks) does not match
  the real one (Render REST API + an account-scoped API key)**, across
  four separate files (`ADR-002`, `docs/rules/deployment.md`,
  `docs/rules/pipeline.md`, `docs/rules/constraints.md`) — not one doc
  that drifted, but four independently wrong in the same way, in the same
  direction: all read like a pre-implementation plan that was never
  reconciled once the real Render-API mechanism was built (`SETUP.md` and
  `COMMANDS.md`, the two docs that were evidently touched again afterward,
  are both accurate). `ADR-002` additionally understates the real service
  count (states "two," the real setup needs four).
- **The false claim that this backend's logic is "inherited from
  forum-app-qa-pipeline" appears in two files with matching wording**, not
  one: `docs/rules/constraints.md` and `docs/rules/testing.md` both state
  it, almost verbatim. Confirmed false by direct diff against real tagged
  snapshots (§2) — this backend is the closest sibling of `ci-testing`'s
  pre-hardening `v1.0.0`, not `qa-pipeline@v1.0.0`.
- **The pipeline auto-deploys to QA on push to `develop`/`master`, not just
  `main`** — confirmed in `ci.yml`'s trigger and in GitHub's environment
  config (`deployment_branch_policy: null` on both environments) — nothing
  technical narrows this beyond "in practice nobody pushes to those
  branches."
- **The README's "47 issues resolved" claim is confirmed false, with a
  known origin**: exhausted every reasonable SonarCloud API angle (open
  issues, resolved issues, unfiltered total, resolution-type breakdown,
  full analysis history, security hotspots) — this project's SonarCloud
  data has never recorded more than 31 issues total, 6 ever resolved. The
  number "47" is boilerplate: the identical line exists in
  `forum-app-qa-pipeline-legacy`'s README, confirming it was copied across
  the series' README template rather than computed for this project.
- **The `deploy-prod` run that looked like an unexplained failure
  (`28621757517`) is fully explained**: the PROD approval request sat
  `waiting` for exactly 30 days, to the second, then GitHub's
  environment-review expiry auto-failed the job with zero steps run — not
  a Render error, not a bad deploy.
- **Live services are confirmed up, confirmed exploitable, and a decision
  has already been made**: see §7. This is not an open question in this
  document.
- **New, minor finding**: every "create" endpoint (`Register`, `CreatePost`,
  `CreateComment`) returns `"created_at":"0001-01-01T00:00:00Z"` — Go's
  zero-value for an unset `time.Time` — in its own creation response, even
  though the database correctly stores the real timestamp. Confirmed in
  code: see §8.

---

## 1. Real environment — tests, coverage, Cypress

Re-run across two separate sessions, not reused from a single pass:

| Check | Result |
|---|---|
| `go test ./tests/services/... -v -cover -coverpkg=./internal/services/...` | 47/47 `--- PASS`, **97.3%** coverage — matches README, `testing.md`, `COMMANDS.md` |
| `npm test -- --coverage --watchAll=false` | 47/47 pass (8 suites), **97.67%** statements / 98.78% lines — matches README's 97.64% within normal metric-choice variance |
| All 4 Cypress spec files, read individually, all 15 tests | Every network-dependent test wraps its calls in `cy.intercept()`, including every `beforeEach`. The one test with no intercept (`posts.cy.js`, missing-title case) needs none — HTML5 validation blocks the submit client-side before any request would fire. **Flatly: 0 of 15 ever reach the real backend.** |
| `npx cypress run` against a real local backend + frontend | **15/15 passing**, ~15s total, nothing behaves differently live vs. the static read — no flake, no incidental real-backend interaction. Backend's post list confirmed still `[]` after the run: no writes landed from any of the 15 tests. |
| SonarCloud quality gate (`api/qualitygates/project_status`) | `OK` — 6 conditions, all on new code (100% new coverage, 0% new duplication, all ratings pass), the default "new code" template, not custom-tightened |

The suspicion that any test/coverage number in this repo is fabricated does
not reproduce anywhere. Every current number in the README, `testing.md`,
and `COMMANDS.md` is real and independently reproducible.

**`ADR-008`'s numbers, precisely dated, not just flagged as stale**: exact
commit timeline confirms 86.5%/92.44% were the real, accurate coverage from
2026-06-24 (introduced in commit `9dc3c50`) through 2026-07-02. `ADR-008`
was created at 18:09 on 2026-07-02 (`04c524f`), citing those numbers
correctly — accurate the moment it was written. A separate commit 51
minutes later (`1cd8dd1`, 19:00, "fix stale docs") updated `testing.md` and
the README to the current 97.3%/97.64% without updating `ADR-008`'s
citation. The staleness is a same-day, 51-minute miss in a docs-consistency
pass, not a long-stale or fabricated figure.

---

## 2. This repo's real lineage

`docs/rules/constraints.md` states: *"The backend logic is inherited from
forum-app-qa-pipeline."* The same claim is repeated, not just implied, in
`docs/rules/testing.md` — its own opening line: *"This file documents the
existing test suite inherited from forum-app-qa-pipeline,"* and again in
its first section: *"The test suite was established in forum-app-qa-pipeline
and must be preserved exactly as inherited."* Two files, matching wording,
same false premise. Confirmed false by direct diff against real tagged
snapshots of all three prior repos (`ci-testing@v1.0.0`, `ci-testing@v1.1.0`,
`qa-pipeline@v1.0.0`, checked out via `git worktree` alongside this repo):

| Comparison | Diff size (comment lines stripped) |
|---|---|
| `post_service.go`: this repo vs. `ci-testing@v1.0.0` | **43 lines** — import path + an `Err*` constants refactor + doc-comment wording. No behavioral difference. |
| `post_service.go`: this repo vs. `ci-testing@v1.1.0` | 43 lines — identical size, because `post_service.go` itself is untouched between ci-testing's two tags |
| `post_service.go`: this repo vs. `qa-pipeline@v1.0.0` | **181 lines** — almost entirely `EditPost`/`EditComment`, confirmed absent here (`grep -n "func.*Edit"` returns nothing; no trace in any test, doc, or commit message across this repo's 43-commit history) |
| `auth_service.go`: this repo vs. `ci-testing@v1.1.0` | 71 lines — the bcrypt hardening pass |

**This backend's business logic is the closest sibling of `ci-testing`'s
original, pre-hardening `v1.0.0` state** — not simply "older than
everything else," specifically that snapshot.

Hardening checked point by point, all four confirmed absent:
- **bcrypt**: absent. `auth_service.go:57,93` still carry the unimplemented
  "in production: hash with bcrypt" comments verbatim.
- **`http.MaxBytesReader`**: absent from every handler (present in 4 places
  in `ci-testing@v1.1.0`).
- **The `err.Error()` leak fix**: absent — **9** raw `err.Error()` responses
  confirmed sent to the client (`auth_handler.go:35,55`;
  `post_handler.go:57,68,86,116,152,170,204`).
- **Rulesets migration**: absent. `gh api .../rulesets` → `[]`, re-confirmed
  twice this audit (rate limit checked clean both times). `ci-testing` and
  `qa-pipeline` both carry a `staging-and-main-protection` ruleset,
  `target: branch`, `enforcement: active` — genuinely enforced, not draft.

**The repo's own git history is younger than it looks.** The actual first
commit (`9dc3c50`, 2026-06-24 23:12:46 -0300) creates the entire file tree
at once — `ci.yml`, `CLAUDE.md`, full `backend/`/`frontend/` trees — with
`backend/go.mod` already reading `module tp06-testing`. Eleven minutes
later, the third commit (`be62589`, 23:23:31) renames it to
`forum-app-cloud-deploy`. This repository's git history never contains the
`tp06-testing` phase as a live, evolving codebase — the actual coursework
development happened before this git repo existed and was imported as a
single working-tree snapshot.

None of this is a defect to fix in this repo — it's the confirmed reason
the new `cloud-deploy` build starts from `qa-pipeline@v1.0.0` rather than
continuing here (ADR-000, forward-propagation, already decided).

---

## 3. Architecture decisions — mandatory-by-assignment vs. voluntary

The TP8 assignment (`08-contenedores-automatizacion.md`) requires, at
minimum: a justified container registry; QA and PROD container deploys with
environment-appropriate config; a full CI/CD pipeline with versioned,
never-`latest` image tags and a manual gate between environments; secret
management; and documented alternatives for each decision. Its own
"100%-free" reference example (`Ejemplo 1`) specifies almost exactly this
repo's stack: *"Container Registry: GitHub Container Registry | Hosting:
Render.com o Fly.io | CI/CD: GitHub Actions."*

| ADR | Requirement level | Alternatives | Verdict |
|---|---|---|---|
| **ADR-001** (container-registry) | Registry itself mandatory; ghcr.io specifically is the assignment's own suggested example, not chosen among true equals | Real — Docker Hub, GitLab, ECR, ACR each get a concrete, checkable downside specific to this project's constraints | Fine as substance; add one sentence acknowledging the overlap with the assignment's example so it doesn't read as unnoticed |
| **ADR-002** (hosting) | Two-env deploy mandatory; Render is again the assignment's example, though the ADR does independently argue against the assignment's other listed option (Fly.io) | Real — Cloud Run, Fly.io, Railway, App Runner, Heroku each concretely differentiated | **Needs a real rewrite.** Describes deploy hooks; the real `ci.yml` uses the Render REST API (`RENDER_API_KEY` + `POST /v1/services/{id}/deploys`) — confirmed mismatched across this ADR and two rule files (§4). States "two services"; the real setup needs four (backend+frontend × QA+PROD), and doesn't even match `SETUP.md`'s own service names |
| **ADR-003** (image-tagging) | Mandatory (assignment explicitly forbids `latest` as sole tag) | Real, honestly framed — `latest` included to explain why it's forbidden, not as a contender | Fine as-is. Matches `ci.yml` exactly. Cites a real prior correction from `qa-pipeline`'s own academic review |
| **ADR-004** (cicd-tool) | Mandatory-adjacent; GitHub Actions is the assignment's example, but the ADR's strongest argument (stages 1–4 already built in Actions during `qa-pipeline`) is independent of that | Real — GitLab CI, CircleCI, Jenkins, Azure DevOps each concretely scoped | Fine as-is, already honest about mixing both reasons |
| **ADR-005** (sqlite-persistence) | **Fully discretionary** — the assignment requires "appropriate resources," not a specific persistence strategy | Real — Render Disk (real $/GB figure), hosted Postgres (correctly scoped as out-of-project-scope, not "worse"), seed-on-startup (correctly dismissed as solving the wrong problem) | Fine as-is — **the clearest, most self-aware trade-off statement in the repo**, and not traceable to the assignment's example at all |
| **ADR-006** (frontend-runtime-config) | **Fully discretionary**, and the most substantial ADR in the repo — the assignment implies runtime config but not this mechanism | Real — separate images (correctly identified as a constraint violation), nginx reverse proxy, `window._env_`, each concretely scoped | Fine as-is — verified the described mechanism against the real Dockerfile/entrypoint script, matches line for line |
| **ADR-007** (deployment-issues, a postmortem) | N/A | N/A | All 5 documented incidents (a–e) verified present/fixed in current code, one by one — arm64/amd64 flag, workflow-level `permissions`, non-root user, `/app` chown, and the API-key placeholder fix. One caveat: the account-level rotation claimed in (e) isn't independently verifiable by reading files, but the `RENDER_API_KEY` GitHub secret's own `updated_at` timestamp (2026-07-02, one day after the docs fix) is real, independent evidence a rotation did happen |
| **ADR-008** (spanish-test-descriptions) | N/A | Real — translate (correctly rejected: same lines carry UI-coupled assertions), leave unannotated (rejected: unreadable to non-Spanish speakers), rewrite to `Should_When` (rejected: that convention is stated to apply only to new tests) | Decision and rationale need no changes; the two cited numbers need correcting to 97.3%/97.64% (or removing entirely — an ADR citing a metric a different file owns is exactly what caused this staleness, see §1) |

**Bottom line**: ADR-005 and ADR-006 are the two decisions in this repo not
handed to it by the assignment's own example — independent judgment is most
visible there, and they're also the two most solid documents in the set.
ADR-001/002/004 sit on top of the assignment's suggested stack; the
alternatives tables are real, but if the new repo wants to read as
"reasoned from scratch" rather than "followed the reference example," those
are the three to either deepen or diverge from.

---

## 4. Pipeline and deploy mechanism

Real trigger, `ci.yml` lines 3–7:
```yaml
on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master, develop ]
```
`docs/rules/pipeline.md` states: *"The pipeline runs on every push or pull
request to `main`."* **Does not match** — three branches trigger it, not
one. Every job from `docker-build-push` onward gates only on
`github.event_name == 'push'`, with no branch check. Confirmed via
`gh api .../environments`: `qa`'s `deployment_branch_policy` is `null`
(not "restricted to a branch" — genuinely unset), and `qa` has no
protection rules at all — so a push to `develop` or `master` would build
images, push to ghcr.io, and auto-deploy to QA with nothing in GitHub's
config narrowing it. `prod` still stops at its required-reviewer gate
regardless of source branch, so PROD itself is safe; QA is not
branch-scoped.

**Deploy mechanism, real vs. documented**: `deploy-qa`/`deploy-prod` call
```
POST https://api.render.com/v1/services/${SERVICE_ID}/deploys
Authorization: Bearer ${RENDER_API_KEY}
```
— the Render REST API, using one repository-level, account-scoped
`RENDER_API_KEY` plus four environment-scoped `RENDER_SERVICE_ID_*`
secrets (all 6 confirmed to exist by name via `gh api .../actions/secrets`
and `.../environments/{qa,prod}/secrets`). **Four** files, not three,
describe a Render **deploy hook** instead — a bare POST to a per-service
URL, with secrets named `RENDER_DEPLOY_HOOK_QA`/`_PROD` that don't exist
in this repo's actual secret set: `ADR-002`, `docs/rules/deployment.md`,
`docs/rules/pipeline.md`, **and `docs/rules/constraints.md`** (its
"Secrets and credentials constraints" section lists "Render deploy hook
URLs (QA and PROD)" among the values that must be GitHub Secrets — the
same wrong mechanism, in the one file explicitly described as carrying
non-negotiable, load-bearing rules). `docs/COMMANDS.md` §10 (rollback) is
the only file whose commands match the real mechanism. **This isn't
equivalent risk if either leaks**: a deploy hook can only redeploy one
service's already-configured tag; the real `RENDER_API_KEY` is scoped to
the whole Render account and can target any service with any image.

**Pattern behind the mismatch, not just its extent**: `SETUP.md` and
`COMMANDS.md` are the only two documents in this repo that match the real,
implemented mechanism. Both were evidently written or updated *after* the
real Render-API approach was implemented and working. `ADR-002` and the
three rule files above all read like a pre-implementation plan — written
when deploy hooks were still the intended approach — that was never
reconciled once the real mechanism changed. The lesson for the new repo
isn't just "use the Render API" (already known) — it's that a plan
document (an ADR, a rules file) needs a deliberate update pass once
implementation diverges from what it describes, or it silently turns into
a historical artifact of intent rather than a description of what's
running.

**The `28621757517` "deploy-prod failure" is fully explained, not a
mystery**: full deployment-status timeline via `gh api`:

| Timestamp | State |
|---|---|
| 2026-07-02T21:14:42Z | `waiting` (required-reviewer gate) |
| 2026-08-01T21:14:42Z | `failure`, job `steps: []` |

Exactly 30 days, to the second. This is GitHub's environment-review request
expiry — Octavio never approved this specific run, and it aged out with a
job log that shows nothing (`gh api .../jobs/.../logs` → `404`, because a
job with zero executed steps has no log to produce). Not a Render error,
not a bad SHA, not a secrets problem.

---

## 5. Repository configuration (GitHub)

| Aspect | State |
|---|---|
| `qa` environment | No protection rules, `deployment_branch_policy: null` |
| `prod` environment | `required_reviewers` (Octavio only), `deployment_branch_policy: null`, `can_admins_bypass: true` (irrelevant in practice — Octavio is the only admin) |
| `main` branch protection (classic) | `404 Branch not protected` |
| Rulesets | `[]` — none, `target: branch`/`enforcement: active` confirmed present on `ci-testing` and `qa-pipeline` for direct contrast |
| Secrets | 6 confirmed by name: repo-level `RENDER_API_KEY` (created 2026-06-26, **updated 2026-07-02** — independent evidence of the ADR-007e rotation) and `SONAR_TOKEN`; `qa`-scoped `RENDER_SERVICE_ID_{BACKEND,FRONTEND}_QA`; `prod`-scoped `RENDER_SERVICE_ID_{BACKEND,FRONTEND}_PROD` |

**`main` has zero protection of any kind** — no ruleset, no classic branch
protection, no required review, no required status checks. Combined with
§4: a direct push to `main` immediately triggers the full pipeline
including an automatic QA deploy.

---

## 6. Documentation findings

**The 7 loose `desc.md` files are the same documents already catalogued in
`qa-pipeline-legacy`'s audit, one generation further back.** Structural
comparison (heading counts, code-fence counts, line counts) is an exact
match on all 7 files — these are the untouched Spanish originals
`qa-pipeline-legacy` translated to English in its June 24 2026 session.
Spot-checked the specific staleness that audit already flagged
(`tests/desc.md`, "PostService (8 tests)" vs. a real count of 36/47): it's
identical, word for word, in both repos — **the translation carried the
stale number forward unchanged**, so there's no "corrected version" of
these files to prefer; `qa-pipeline-legacy/audit-results.md` §4's verdicts
(rescue/partially rescue/discard) transfer directly.

**`testing.md` contradicts itself internally.** Its table states 109 tests
(47+47+15) — matches reality exactly. Its own "before adding" section says
*"Confirm that all 89 existing tests pass"* — 89 doesn't match anything
real in this repo (not the total, not any subtotal). Provenance unconfirmed
with certainty; this repo's `ci.yml` has no `quality-summary`-style job to
trace it to (unlike `qa-pipeline-legacy`'s, which does), making it more
likely this line was copy-pasted from a `qa-pipeline`-era doc during
scaffolding and never updated.

**`testing.md` also asserts something directly contradicted by this
audit's own Cypress finding**: *"Handlers and repository implementations
are not unit tested — they are covered by E2E tests via Cypress."* Section
1 already established that all 15 Cypress tests are 100% mocked and never
reach the backend — so `internal/repository/` (where `DeleteComment`'s
authorization SQL lives) has **no** real coverage from any layer, not
partial E2E coverage as this line claims.

**The README's "Issues Resolved: 47" is confirmed false, with a known
origin.** SonarCloud's issues API was queried every reasonable way:
`resolved=true` → 6; `resolved=false` → 25; no filter → 31 (consistent:
25+6); resolution-type breakdown → all 6 are genuine `FIXED`, no other
category hiding a larger count; full analysis history (11 analyses,
2026-06-25 to 2026-07-02) → open-issue count peaks at 30 on the very first
analysis and has never been higher; security hotspots (a separate
SonarCloud category, in case 47 blended both) → 0. **47 was never reached
at any point in this project's SonarCloud history.** Its actual origin:
`forum-app-qa-pipeline-legacy/README.md` has the identical line
(`| Issues Resolved | ≥3 | 47 issues | ✅ |`) — this is boilerplate carried
across the series' README template, never a number computed against this
project's real SonarCloud data.

**`ADR-002`'s service count is wrong**, independent of its deploy-mechanism
error (§4): states "two Render services... `forum-app-qa` and
`forum-app-prod`"; the real setup (`SETUP.md` §5, and the 4
`RENDER_SERVICE_ID_*` secrets `ci.yml` actually references) is four
services with different names entirely.

**Real Render PROD service IDs were hardcoded in `docs/COMMANDS.md` —
fixed as part of this audit.** `ADR-007(e)` documents rotating a leaked
`RENDER_API_KEY` from the same rollback section, but the two real
`srv-...` PROD service IDs alongside it were never redacted at the time —
`SETUP.md` used a placeholder (`srv-xxxxxxxxxxxxxxxx`) for the same kind of
value, `COMMANDS.md` had the real ones. Lower severity than the API-key
leak (a service ID alone can't authenticate anything), but real
infrastructure detail left exposed through the exact review that caught
and fixed the more serious leak alongside it. **Applied**: both instances
in `docs/COMMANDS.md` §10 replaced with the same `srv-xxxxxxxxxxxxxxxx`
placeholder `SETUP.md` already uses — the only change made to this repo's
source outside the two audit documents themselves.

---

## 7. Live exposure — confirmed exploitable, decision made

**This is not an open question in this document.** A dedicated pass
(`cloud-deploy-live-exposure-checklist.md`, closed and reviewed) confirmed
the spoofable-authentication design is exploitable against the real,
public, live QA deployment — not just in source — and the decision on what
to do about it has already been made.

**All 4 live URLs are up and serving the real app.** Confirmed with actual
response bodies, not just status codes: both backends return real JSON
(`[]`, currently empty — consistent with `ADR-005`'s ephemeral SQLite, no
real or dummy data exposed today), both frontends serve a real CRA
production bundle. The QA frontend's actual served JS was confirmed to
contain the correct QA backend URL (`ADR-006`'s runtime placeholder
mechanism genuinely working live, not just in source), and PROD serves the
byte-identical bundle filename as QA — live confirmation of `ADR-003`'s
"same artifact, same SHA" claim. Response times varied between passes
(~12s cold start on one pass, sub-second on a later one, after intervening
traffic) — consistent with Render's free-tier idle/wake behavior, not an
anomaly (see the companion transferable-knowledge document, §7).

**The exploit was confirmed live, independently, by two separate parties,
against QA only, non-destructively:**

- **`DeletePost` — reproduced independently by Octavio Carpineti,
  2026-08-03.** Run by hand, step by step, against the live QA backend:
  registered his own test user (`id: 3`), created a post (`id: 2`),
  confirmed the negative control (`X-User-ID: 999999` → `403`), then sent
  the real request — `X-User-ID: 3`, with no `/login` call and no
  password/token presented anywhere in the request — and received
  `{"message":"post deleted"}`, confirmed by a subsequent `GET` returning
  `[]`. This reproduction is Octavio's own, run independently of Code's
  tooling or session.
- **`DeleteComment` — reproduced by Code, 2026-08-03**, extending the same
  test to the specific endpoint that started this entire audit chain
  (originally found in `qa-pipeline-legacy`, because its authorship check
  lives in the repository layer, excluded from coverage by
  `sonar-project.properties`). Registered a victim (`id: 1`), a post, and a
  comment; ran the negative control twice — once against a nonexistent ID
  (`999999` → `403 "user not found"`, rejected earlier in the service
  layer) and once against a second real, registered "attacker" user
  (`id: 2` → `403 "you do not have permission..."`, the actual
  repository-layer check under test, confirmed genuinely rejecting a real
  wrong user); then sent the real request — `X-User-ID: 1`, no prior login
  anywhere — and received `{"message":"comment deleted"}`, confirmed by a
  subsequent `GET` returning `[]`.

Both tests cleaned up after themselves; QA's post list was `[]` at the end
of each. Checked for mitigating factors: all 4 services sit behind
Cloudflare, but this is Render's own default infrastructure edge, not an
application-aware WAF — it did nothing to block either exploit request, and
a 15-request unthrottled burst against the QA backend confirmed no
rate-limiting either. Render's dashboard (deploy history, suspension
events, third-party traffic) was not accessible in either session — this
remains a real, acknowledged gap, not something either verification could
close.

**Decision (Octavio, already made — not pending)**: the 4 Render services
(QA and PROD, backend and frontend) will be **suspended, not deleted** — a
reversible action — rather than left running or torn down permanently. This
closes Phase 4 of the live-exposure checklist. Execution (logging into the
Render dashboard and suspending each service) is Octavio's own action, not
something applied as part of this audit.

---

## 8. Minor new finding — zero-value `created_at` on every creation response

Every "create" endpoint's own response carries a broken timestamp: `POST
/api/auth/register`, `POST /api/posts`, and `POST /api/posts/{id}/comments`
all return `"created_at":"0001-01-01T00:00:00Z"` — Go's zero-value for an
unset `time.Time` — even though the row is correctly written with a real
timestamp in the database.

**Confirmed in code, not just observed live**: `user_repository.go:28-45`
and `post_repository.go:32-47` (`Create` for both users and posts) both
insert with `VALUES (?, ?, ?, datetime('now'))` in the SQL — the database
row is correct — but only read `LastInsertId()` back into the passed
struct's `.ID` field; `CreatedAt` is never re-populated from the insert.
The same in-memory struct (built earlier in the service layer with
`CreatedAt` never set) is what gets serialized straight back to the client.
Confirmed the same pattern generalizes to comments empirically: the `POST
/comments` response in the live-exposure test showed the zero-value, while
an immediately following `GET` of the same comment (a separate read from
the database) showed the correct, real timestamp — consistent with the
bug being confined to the synchronous creation response, not a data
integrity issue.

**Severity**: cosmetic/minor — no data is lost or wrong in the database,
only the immediate response to a create call. Worth fixing in the new repo
(either re-fetch after insert, or use `RETURNING`/set the Go-side timestamp
before the insert and pass the same value into the query) since it's the
kind of thing that shows up immediately in any UI rendering "posted just
now."

---

## 9. Open questions for Octavio

Trimmed to what genuinely remains undecided — resolved items (was `ADR-008`
ever accurate, is the API key really rotated, is "47 issues" real, is the
live-exposure decision made) are answered above, not repeated here.

1. **Deploy hook vs. Render API** (§4): which mechanism does the new repo
   adopt? Not equivalent blast radius on a leaked secret — a deploy hook is
   scoped to one service and can only redeploy its already-configured tag;
   the current `RENDER_API_KEY` approach is scoped to the whole account and
   can target any service or image.
2. **Cypress with a live backend it never calls** (§1): keep a CI job that
   boots a real backend process for specs that are 100% mocked, or cut it
   as dead weight? Same still-open question as `qa-pipeline-legacy`'s open
   question #3.
3. **`deploy-qa` branch scope** (§4): restrict explicitly to `main` in the
   new repo, or is "in practice nobody pushes to `develop`/`master`" an
   acceptable stance given there's no technical control preventing it
   today?
4. **30-day silent approval expiry** (§4): acceptable default for the new
   repo, or does an abandoned PROD approval need an explicit,
   shorter-than-default timeout so it doesn't sit looking pending
   indefinitely?
