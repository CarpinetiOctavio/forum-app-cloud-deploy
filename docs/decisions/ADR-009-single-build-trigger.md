# ADR-009: Single Build Trigger — QA and PROD Deploy Both Originate from `main`

**Date:** 2026-08-05
**Status:** Accepted

## Context

`docs/rules/pipeline.md`, as originally written for this repo, mapped
Stage 7 (deploy to QA) to `if: github.ref == 'refs/heads/staging'` and
Stage 9 (deploy to PROD) to `if: github.ref == 'refs/heads/main'`. This
mapping was stated as "this repo's own, not inherited from anywhere," with
no comparison against alternatives and no ADR backing it — a gap in itself,
per `docs/rules/documentation.md`'s requirement that every relevant
decision trace to a real justification.

Designing Stage 9 surfaced why that gap mattered. `staging` and `main` are
different branches, receiving pushes at different times. Because
`docs/rules/pipeline.md`'s Trigger section configures
`push: branches: ['feature/**', staging, main]` with no per-job branch
condition on Stages 5–6 (Docker build + push to `ghcr.io`), every push to
either branch independently builds and pushes its own image, tagged with
that push's own commit SHA. A push to `staging` and the later merge commit
that lands on `main` are, in git, different commits — even when the tree
content is identical, a Docker build is not reproducible byte-for-byte
between two separate invocations (layer timestamps, Go build IDs). Stage 9,
triggered by the `main` push, would either have to trust that its own
freshly-rebuilt image happened to be equivalent to whatever Stage 7 already
validated on `staging` — an unverifiable assumption — or reach back into
git history (`HEAD^2` on the merge commit) to recover the exact SHA Stage 7
had validated and redeploy that instead of its own run's build.

Both of those are attempts to reconcile two independently-built images after
the fact. Checking why reconciliation was needed at all led to
`docs/rules/constraints.md` — the document `pipeline.md` itself declares as
its authority ("every rule here derives from a constraint stated there").
It already states, as a non-negotiable MUST, independent of and prior to
anything written in `pipeline.md`:

> **MUST**: build the image exactly once per pipeline run. The same image,
> identified by the same commit-SHA tag, is what gets deployed to QA and,
> after approval, to PROD — never rebuilt between the two.

`constraints.md` also states its own conflict-resolution rule: *"If a rules
file appears to permit something this file forbids, this file wins — stop
and flag the conflict, do not resolve it by improvising."* The
`staging`-triggers-QA / `main`-triggers-PROD mapping is exactly this case:
by construction, it requires two separate pipeline runs to produce what
`constraints.md` requires to be one image from one run. This was not
noticed when `pipeline.md`'s Stage 7/9 mapping was originally written.

A second, mechanical inconsistency compounds the first. `pipeline.md`'s own
Trigger section states, as its own MUST: `push: branches: ['feature/**']`
only — yet the same section's supporting text cites `qa-pipeline@v1.0.0`'s
real trigger as `push: ['feature/**', staging, main]`, and the real,
current `ci.yml` implements that broader pattern, not the narrower one the
document states as its own rule. This is what mechanically causes Stage 5–6
to fire on `staging` pushes at all — the literal trigger for the whole
problem this ADR resolves.

Octavio's own prior implementation of this same application, submitted for
this course's final assignment, never had this problem: a single pipeline
run, triggered once code reached `main`, built one image, deployed it to
QA, and — after manual approval — promoted that same image to PROD. This
ADR's decision restores that model, informed by why the current repo's
design had drifted from it.

## Decision

Stage 7's trigger changes from `if: github.ref == 'refs/heads/staging'` to
`if: github.ref == 'refs/heads/main'`. Stage 7 and Stage 9 now both run
within the single pipeline run triggered by a push to `main`, consuming the
one image that run's Stage 5–6 built — never a separately-built one.
`HEAD^2` resolution, the two-parent guard, and disabling squash/rebase
merges are no longer needed; the problem they were built to work around no
longer exists once there is only ever one build per promotable release.

Stages 1–6 are **unchanged** — they continue running on every push
regardless of branch (`feature/**`, `staging`, `main`), per the existing
Trigger section and matching `qa-pipeline@v1.0.0`'s own real precedent.
Confirmed explicitly with Octavio: this is deliberate, not an oversight —
fast feedback on every push, and by the time code reaches `main` it has
already been validated twice (the `feature` push itself, and the
`staging` push/PR). The images `feature/**` and `staging` pushes build are
never referenced by any deploy stage; they accumulate unused in `ghcr.io`,
an already-accepted cost (`ADR-003`'s Consequences section), not a new one
introduced here.

## Rationale

**This is not a new design choice — it is compliance with a rule this repo
already committed to.** `constraints.md` predates this decision and
already required build-once, promote-everywhere (Humble & Farley); this
ADR corrects an implementation that had drifted from a constraint already
in force, rather than introducing a new one.

**The manual approval gate only means what it claims to mean if what's
approved is what ships.** With two independently-built images, approving
"the QA build" and deploying "the `main` build" are approvals of two
different artifacts that happen to share source code — the gate becomes a
formality, not a real control. `docker.md` Rule 1 and `deployment.md`'s
"Image promotion" section both name this as the specific property being
protected.

**A single trigger makes the guarantee true by construction, not by
verification after the fact.** `HEAD^2` would have made the correct
outcome *reachable* through git topology, but only for the specific case
of a real two-parent merge commit — fragile to how the PR was actually
merged, and dependent on a guard to fail safely when that assumption
breaks. Removing the second trigger removes the ambiguity itself; there is
never a second candidate image to reconcile against.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| `HEAD^2` resolution + hard guard on Stage 9 | Technically sound for the case of a genuine two-parent merge commit (verified against this repo's real git history), but solves the symptom, not the cause — still requires reasoning about merge topology to answer "which image is the validated one," and requires disabling squash/rebase repo-wide (a side effect touching every future PR, not just `staging`→`main`) to keep the guard's precondition reliably true. |
| Persist the QA-validated SHA across runs (artifact, GitHub Deployments API lookup, or similar), and have Stage 9 read it | Solves the same problem as `HEAD^2` without depending on merge-commit shape, but adds a second real piece of infrastructure (cross-run state) to maintain and audit, for a problem a trigger change removes outright. No justification for the added complexity once the simpler fix is available. |
| Keep the current mapping, accept two independently-built images from identical source as "close enough" | Rejected outright, not just deprioritized: `deployment.md`'s verification clause is digest identity, not source-code identity — "close enough" doesn't satisfy a MUST stated as a checkable equality. Docker builds are not reproducible byte-for-byte by default in this pipeline (no confirmed deterministic build flags in either Dockerfile), so the two images are not guaranteed identical even in principle. |

## Consequences

- `docs/rules/pipeline.md`'s Stage table, Trigger section, Stages 7/9
  branch-mapping subsection, and GitHub environments table all require a
  corresponding rewrite — tracked as the immediate follow-up to this ADR,
  not deferred.
- The `qa` GitHub environment's `deployment_branch_policy` changes from
  "restricted to `staging`" to "restricted to `main`" — required for Stage
  7 to be able to run at all once its trigger changes; leaving the old
  policy in place would make GitHub block the very deploy this ADR intends
  to enable.
- QA validation of a deployed artifact now happens only after code has
  already reached `main`, never before. This is a real trade-off, not a
  free improvement: it was possible, under the old mapping, to catch a
  deploy-time issue on `staging` before promoting to `main`. Accepted
  because Stages 1–4 (code-level tests, static analysis, E2E) already gate
  both `feature → staging` and `staging → main`, and because
  `smoke-testing.md` already establishes that deploy-time issues (env var
  wiring, port binding, container startup) are a distinct failure domain
  from code-level tests — one that Stage 7 continues to catch, just one
  merge later than before.
- `HEAD^2` resolution, the two-parent guard, and any repo Settings change
  to merge strategies are no longer required and are not implemented.
- Rollback (`deployment.md`'s Rollback section) is unaffected — it already
  depended only on SHA tags never being overwritten (`docker.md` Rule 5),
  not on which stage triggered which deploy.