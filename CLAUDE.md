# CLAUDE.md — forum-app-qa-pipeline

## Purpose of this file
Operating context for any AI assistant working in this repository. This repo is TP7
of a graded course series, built as a copy of `forum-app-ci-testing@v1.1.0` — not
continued from `forum-app-qa-pipeline-legacy`, the original TP7 mirror (see
`ADR-000` for why). It is the second of three repos in a pipeline series
(ci-testing → qa-pipeline → cloud-deploy), each with a strictly bounded scope.

## Scope boundary — do not cross
This repo's scope is coverage measurement with a 70% gate, static analysis with
SonarCloud, E2E testing with Cypress, and quality gates that block the pipeline —
built on top of the Services-layer unit testing `ci-testing` already established,
not a re-litigation of it. Explicitly out of scope: Docker, containerization,
deployment — those belong to `cloud-deploy`. Also out of scope, ordinarily:
expanding unit-test *scope* to layers `ci-testing`'s `ADR-002` already excluded
(Handlers, Repository, Router) — that boundary is inherited, not reopened here.

The one exception: this repo may exceed TP7's literal scope when doing so is a
condition for this repo's own declared guarantees to be real, not for general
improvement — see `docs/rules/documentation.md`'s ADR-justification category (d)
and `ADR-000` for the criterion and its limits. The authorization-vs-authentication
test work (see `docs/rules/testing-and-quality-gates.md`) is the concrete instance
of this exception currently on record; it is not a license to expand scope
elsewhere without the same justification.

## Methodology (see ADR-000)
Decisions in this repo are not modeled on how `forum-app-cloud-deploy` solved the
same problem — that would import a later stage's reasoning into an earlier one.
They may, and often should, build on `forum-app-ci-testing`'s own decisions where
directly relevant — referenced by link (see `docs/rules/documentation.md`'s
cross-repo references section), never duplicated. Each decision must be grounded
independently in software engineering fundamentals relevant to this repo's own
scope — established concepts and practice, not "the other repo does it this way."
Once a decision here is fully fundamented, it becomes the baseline that propagates
forward to `cloud-deploy` — never the reverse.

## Initialization protocol
Before writing or modifying anything in a session:
1. Read every file in `docs/rules/` in full.
2. Read every ADR in `docs/decisions/` in full, in order.
3. Verify the current state of the repo against what the documentation claims (test
   counts, file structure, CI steps, branch/ruleset state) — do not assume the docs
   are accurate. See `docs/rules/verification.md` for the standard of evidence this
   requires; a prose summary from a prior session is not sufficient on its own.
4. Report findings and proposed next steps. Wait for explicit approval before
   writing anything.

## Decision-making authority
This assistant proposes and fundamenta options. It does not decide. Any change
affecting test behavior, scope, or documentation structure requires Octavio's
explicit approval before being written.

## Requirements for any proposed change, in order
1. Scope check — strictly within TP7's boundary, or explicitly justified under the
   exceeds-scope criterion above?
2. Fundamentation check — grounded in a real software engineering concept relevant
   to coverage, static analysis, or E2E testing, not just "it works" or "the other
   repo does it this way"?
   A change that fails either check gets flagged, not implemented.

## Documentation standard
English for all prose (README, ADRs, SETUP, COMMANDS). Test names in English,
inherited from `ci-testing`'s codebase and its own `ADR-006` — not a decision made
in this repo, so not re-argued here (see `docs/rules/documentation.md`).

## AI usage disclosure
Claude acts as a conceptual auditor and writing assistant — never as decision-maker
for test design, mocking strategy, coverage scope, or quality-gate configuration.
All design decisions were made by Octavio Carpineti; Claude's role was surfacing
inconsistencies, verifying claims against the actual repo state, grounding
proposals in software engineering fundamentals, and drafting documentation for
review.