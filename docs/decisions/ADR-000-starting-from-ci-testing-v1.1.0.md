# ADR-000: Starting from a copy of forum-app-ci-testing@v1.1.0, not continuing forum-app-qa-pipeline-legacy

**Date:** 2026-07-27
**Status:** Accepted

## Context

`forum-app-qa-pipeline` existed as a mirror of this course's TP7 assignment, which
itself descended from the **original** TP6 submission — not from
`forum-app-ci-testing`'s rebuilt-from-zero `v1.0.0`. A working session on 2026-06-24
did real, dated work on it (a SonarCloud project-key fix, partial translation of
comments and error strings, expanded test coverage, five ADRs), but that session
built on top of the pre-existing base rather than rebuilding it from ci-testing's
already-corrected baseline.

Two independent audits — one run directly against the repo in this project's
qa-pipeline chat, one run by a separate Claude Code session tasked with sweeping the
full repo — found that this base still carries concrete technical debt that predates
`ci-testing`'s correction pass, not just documentation lag:

- **CI workflow still targets Node 20** and uses `codecov/codecov-action@v3` as an
  external coverage dependency in both the backend and frontend test jobs.
  `ci-testing` already resolved both: Node 24, and internal artifact uploads instead
  of Codecov.
- **The CI trigger still includes a `develop` branch** (`on: push: branches: [main,
  master, develop]`). `ci-testing` evaluated and explicitly rejected a `develop`
  branch model for a single committer; it reappears here unexamined.
- **The `quality-summary` job prints hardcoded test counts and coverage
  percentages** regardless of the jobs that ran before it (`if: always()`, not "runs
  only if all preceding jobs pass" as its own ADR-005 claims) — and those hardcoded
  numbers (35 backend / 39 frontend / 89 total) were independently verified against
  the actual test suite (47 Go + 47 Jest = 94, or 109 including the 15 Cypress
  specs) and found to be stale on every field, not just the total.
- **Two of the five existing ADRs have content problems, not just formatting
  ones**: `ADR-002-coverage-strategy.md` is a ~95%-literal duplicate of
  `ADR-001-tech-stack.md` — same Context, Decision, Rationale, and Alternatives
  sections about Go/React/SQLite, with a single orphaned sentence about
  `collectCoverageFrom` in its Consequences section. It does not document a coverage
  strategy at all: no rationale for the 70% threshold, for measuring only
  `internal/services/`, or for excluding `handlers`/`repository`/`router`.
  `ADR-005-cicd-pipeline.md`'s own description of the `quality-summary` job does not
  match the YAML it describes, and its "Problems encountered and resolved" section
  omits both the Codecov dependency and the `develop` branch trigger — neither was
  ever a documented decision, both were unexamined scaffold defaults.

The application source code itself (not the CI/testing layer) was also audited, in
parallel, once it became clear that `qa-pipeline-legacy`'s backend and frontend are
byte-for-byte the same code as `ci-testing`'s. That audit surfaced a shared
authentication design gap; it is documented and resolved separately, in
`ci-testing`'s own `ADR-008`, per this series' forward-propagation rule — it does
not bear on the decision recorded in this ADR, which is specifically about which
codebase and CI baseline this repository starts from.

## Decision

`forum-app-qa-pipeline` is built as a **copy** — not a fork, and not a continuation
in place — of `forum-app-ci-testing` at tag `v1.1.0`. TP7's scope (coverage
measurement with a 70% gate, static analysis via SonarCloud, E2E tests via Cypress,
and quality gates that block the pipeline) is added on top of that base.

The old repository is renamed `forum-app-qa-pipeline-legacy`, archived on GitHub
(read-only, marked with the platform's own "Archived" badge), and kept as a
reference — specifically because it contains close to 110 tests exceeding the
course's stated minimum (3 E2E flows), and reasoning captured in its five ADRs that
may still be valid even where the code implementing it is being replaced. Its
README carries a short notice marking it superseded and pointing back to this ADR.
What gets ported, adapted, or discarded from it — test by test, ADR by ADR — is
decided and recorded separately, informed by both audits, as the new repository is
built; this ADR only settles that the new repository's code and CI baseline come
from `ci-testing@v1.1.0`, not from `qa-pipeline-legacy`.

## Alternatives considered and rejected

**Continue building on `qa-pipeline-legacy` in place**, applying the same fixes
`ci-testing` already made (Node version, Codecov, `develop` trigger, quality-summary
logic) directly to it. Rejected: every one of those fixes has already been designed,
implemented, and reasoned through once, in `ci-testing`. Redoing them here would not
be polishing a base — it would be re-deriving corrections that already exist,
against a codebase that also still carries the un-rebuilt TP6 lineage's other
issues. The series' own forward-propagation methodology (`ci-testing`'s `ADR-000`)
exists specifically to prevent this: corrections belong at their origin point and
travel forward from there.

**Fork `ci-testing` on GitHub** rather than copy it. Rejected: a fork is
permanently marked as "forked from" in GitHub's UI, is excluded from the default
repository view on a profile, and cannot be reversed without contacting GitHub
support. Confirmed against GitHub's own documentation: a fork also carries the
entire commit history of the parent repository, and commits to a fork do not
count toward the owner's contribution graph — a repository generated from a
template, by contrast, starts with a single commit and its commits do count.
This repo is meant to stand as its own portfolio entry in the series, not as a
visually subordinate branch of `ci-testing`.

**Replace `qa-pipeline-legacy`'s contents in place** (keep the existing repository,
history, and URL, but overwrite its contents with `ci-testing@v1.1.0` as a new
first commit). Rejected: the existing repository has no external references worth
preserving (no stars, watchers, or a URL anyone has bookmarked — it is a course
mirror), so preserving its identity adds historical noise without a corresponding
benefit. A clean copy with an honest ADR explaining why the mirror was abandoned
achieves the same transparency without that noise.

## Consequences

- The new repository's CI/CD baseline and translated documentation are inherited
  correctly from day one, rather than needing to be re-derived inside this repo.
  The Go module path is a separate matter: it is copied verbatim as
  `forum-app-ci-testing` at first, which is simply wrong for this repository — it
  identifies a different repo in the series, the same category of problem as the
  stale `tp06-testing` identifier already flagged in `qa-pipeline-legacy`. This is
  a mechanical rename (updating the module path and its import statements), not a
  re-derivation of any decision `ci-testing` made, and is corrected as part of
  setting this repository up, not deferred. `ci-testing`'s eleven existing ADRs at
  the time of this decision (`ADR-000` through `ADR-010`) are **not** copied as
  files into this repository — per the cross-repo referencing convention adopted
  alongside this ADR, `docs/decisions/` here starts empty and holds only
  decisions made in this repo; where one of them builds on something
  `ci-testing` already resolved, it links to `ci-testing`'s own ADR rather than
  duplicating its text.
- The three `ci-testing@v1.1.0` security-hardening fixes (bcrypt password hashing,
  request body size limits, the `GetAllPosts` error-leak fix — `ci-testing`'s
  `ADR-008`) are inherited as well. The one item `ADR-008` explicitly left as an
  accepted, deferred limitation — the absence of real authentication behind
  `X-User-ID` — is inherited too, and is addressed in its own ADR in this
  repository once the rest of TP7's scope is underway, not folded into this one.
- `qa-pipeline-legacy` remains on GitHub, renamed, archived, and untouched, as the
  reference the audits were run against — not deleted, so that the reasoning trail
  (mirror → audited → abandoned in favor of a clean copy, with cause) stays
  verifiable rather than asserted.
- This is the same category of decision `ci-testing`'s own `ADR-000` recorded for
  itself (resolving forward, not mirroring backward) — applied here one stage later
  in the same series, for the same reason.