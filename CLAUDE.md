# CLAUDE.md — forum-app-cloud-deploy

## Purpose of this file
Operating context for any AI assistant working in this repository. This repo is
TP8 of a graded course series, built as a copy of `forum-app-qa-pipeline@v1.0.0`
— not continued from `forum-app-cloud-deploy-legacy`, the original TP8 build
(see `ADR-000` for why). It is the third of three repos in a pipeline series
(ci-testing → qa-pipeline → cloud-deploy), each with a strictly bounded scope.

## Scope boundary — do not cross
This repo's scope is containerization, a container registry, CI/CD deployment
to QA and PROD environments, secrets management, and a manual approval gate
between environments — built on top of the testing and quality-gate work
`qa-pipeline` already established, not a re-litigation of it. Explicitly out
of scope: expanding or modifying the inherited test suite's scope, coverage
gate thresholds, SonarCloud rule configuration, or Cypress mocking strategy —
those are `qa-pipeline`'s decisions (and, transitively, `ci-testing`'s),
inherited as-is, not reopened here. Also out of scope: changes to backend or
frontend application logic in general — this repo containerizes and deploys
the app, it does not modify what the app does — **with one accepted
exception, real authentication (see below), not a general license to touch
application logic elsewhere.**

**Exception on record: real authentication (`ADR-001`).** `qa-pipeline`
recorded, in the commit that wrote `docs/rules/testing-and-quality-gates.md`
(`8b07d0d`, `forum-app-qa-pipeline`), that the app's spoofable
`X-User-ID`-header authentication model was "deferred to cloud-deploy" —
that intent never made it into `qa-pipeline`'s own committed prose, only its
commit message. `ADR-001` accepts that handoff explicitly. This is scope
acceptance only — the mechanism (JWT, sessions, or another approach) is not
yet designed and gets its own follow-up ADR once it is. Any other
application-logic change still requires the same exceeds-scope justification
below; this is the one specific instance on record, not a precedent for
touching business logic generally.

The general exception mechanism: this repo may exceed TP8's literal scope
when doing so is a condition for this repo's own declared guarantees to be
real, not for general improvement — see `docs/rules/documentation.md`'s
ADR-justification category (d) and `ADR-000` for the criterion and its
limits. `ADR-001` is the concrete instance currently on record; if another
arises, it gets its own ADR under the same criterion, not folded silently
into unrelated work.

## Methodology (see ADR-000)
Decisions in this repo are not modeled on how `forum-app-cloud-deploy-legacy`
solved the same problem — that repo's source is obsolete (confirmed in
`cloud-deploy-legacy-audit-results.md`: pre-hardening, missing features
`qa-pipeline` already has, a documented deploy mechanism that never matched
what was really implemented). `-legacy` is a source of decisions and their
reasoning to understand, not a file to copy and patch — the same way
`qa-pipeline` used `qa-pipeline-legacy`, and the same correction that had to
be made explicitly during this repo's own starter-verification process (see
`docs/audits/cloud-deploy-starter-verification-checklist.md`) after an early
draft of that checklist got this wrong.

Decisions here may, and often should, build on `forum-app-qa-pipeline`'s own
decisions where directly relevant — referenced by link (see
`docs/rules/documentation.md`'s cross-repo references section), never
duplicated. Each decision must be grounded independently in software
engineering fundamentals relevant to this repo's own scope — established
concepts and practice, not "the other repo does it this way," and not "the
legacy repo already decided this." If this series extends beyond this repo,
a decision fully grounded here becomes the baseline that propagates forward
— never the reverse.

## Rules-file convention — MUST/MUST NOT, always with a why
*(new)* Every rule stated in `docs/rules/` uses **MUST** / **MUST NOT** /
**MUST NOT, except** for anything that is actually mandatory, each one
immediately followed by its justification. A plain declarative sentence
without that markup is describing something (current state, context,
history), not imposing a requirement — don't treat it as one, and don't
write a new requirement without the markup and the why. This exists so a
rule can be enforced or checked mechanically, not interpreted differently
by different sessions.

## Initialization protocol
Before writing or modifying anything in a session:
1. Read every file in `docs/rules/` in full.
2. Read every ADR in `docs/decisions/` in full, in order.
3. Verify the current state of the repo against what the documentation claims
   (test counts, file structure, CI steps, branch/ruleset state, which
   secrets and environments actually exist) — do not assume the docs are
   accurate. See `docs/rules/verification.md` for the standard of evidence
   this requires; a prose summary from a prior session is not sufficient on
   its own.
4. Report findings and proposed next steps. Wait for explicit approval before
   writing anything.

## Decision-making authority
This assistant proposes and fundamenta options. It does not decide. Any change
affecting infrastructure, deployment mechanism, scope, or documentation
structure requires Octavio's explicit approval before being written.

## Requirements for any proposed change, in order
1. Scope check — strictly within TP8's boundary, or explicitly justified under
   the exceeds-scope criterion above?
2. Fundamentation check — grounded in a real software engineering concept
   relevant to containerization, registries, deployment, or CI/CD, not just
   "it works" or "the legacy repo did it this way"?
   A change that fails either check gets flagged, not implemented.

## Documentation standard
English for all prose (README, ADRs, SETUP, COMMANDS). Test names in English —
inherited transitively through `qa-pipeline`'s codebase from `ci-testing`'s own
`ADR-006` — not a decision made in this repo, so not re-argued here (see
`docs/rules/documentation.md`).

## AI usage disclosure
Claude acts as a conceptual auditor and writing assistant — never as
decision-maker for container/registry/hosting choices, deployment mechanism,
or secrets management. All design decisions were made by Octavio Carpineti;
Claude's role was surfacing inconsistencies, verifying claims against the
actual repo state, grounding proposals in software engineering fundamentals,
and drafting documentation for review.