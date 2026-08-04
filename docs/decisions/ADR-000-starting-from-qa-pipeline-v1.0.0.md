# ADR-000: Starting from a copy of forum-app-qa-pipeline@v1.0.0, not continuing forum-app-cloud-deploy-legacy

**Date:** 2026-08-04
**Status:** Accepted

## Context

`forum-app-cloud-deploy-legacy` was built in July 2026 for this course's TP8
final defense — unlike `qa-pipeline-legacy` (a raw, unedited October 2025
submission with no portfolio intent), `-legacy` was already built with some
portfolio awareness: it has 8 ADRs, a full `docs/rules/` set, `SETUP.md`,
`COMMANDS.md`, diagrams, and screenshots. That made it a real candidate to
continue building on directly, not an obvious rebuild-from-scratch the way
`qa-pipeline-legacy` was.

Two rounds of audit work — an initial pass, then a second, checklist-driven
pass that re-verified every finding against running code, live deployed
services, and external APIs (`cloud-deploy-audit-checklist.md`,
`cloud-deploy-live-exposure-checklist.md`, both closed and reviewed) —
compiled into `cloud-deploy-legacy-audit-results.md` and
`cloud-deploy-legacy-transferable-knowledge-results.md`. Together they found
that `-legacy`'s apparent maturity does not extend to its code lineage or to
several of its documents:

- **`-legacy`'s backend does not descend from `qa-pipeline@v1.0.0`**, despite
  its own `constraints.md` and `testing.md` both stating it does, in matching
  wording. Diffed directly against real tagged snapshots of all three prior
  repos: `-legacy`'s business logic is the closest sibling of `ci-testing`'s
  **pre-hardening** `v1.0.0` state (43 non-comment lines of difference — a
  module path and a constants refactor, nothing behavioral) — not
  `qa-pipeline@v1.0.0` (181 lines of difference, almost entirely
  `EditPost`/`EditComment`, absent from `-legacy` entirely) and not
  `ci-testing@v1.1.0` (71 lines of difference on `auth_service.go` alone,
  from the hardening pass `-legacy` never received: no `bcrypt`, no
  `http.MaxBytesReader`, 9 confirmed raw `err.Error()` responses sent to the
  client). `-legacy`'s own git history opens on 2026-06-24 with a single
  commit importing an entire pre-existing `tp06-testing`-named project
  wholesale — the real prior development happened outside that repo's git
  history entirely.
- **The spoofable-authentication design this predates is not hypothetical —
  it was confirmed exploitable against `-legacy`'s real, live, deployed QA
  service**, independently, twice: once by Claude (the `DeleteComment` case,
  the specific finding this entire audit chain traces back to, originating
  in `qa-pipeline-legacy`'s own audit) and once by Octavio himself,
  reproducing the `DeletePost` case by hand against the same live backend.
  Both confirmed that a request presenting no credential beyond a
  client-supplied `X-User-ID` header can delete another account's content.
- **Several of `-legacy`'s documents describe a plan that was never
  reconciled with what was actually implemented.** Four separate files
  (`-legacy`'s own `ADR-002-hosting.md`, `docs/rules/deployment.md`,
  `docs/rules/pipeline.md`, `docs/rules/constraints.md`) describe deploying
  via a Render deploy hook; the real, implemented mechanism is the Render
  REST API authenticated with an account-scoped `RENDER_API_KEY` — only
  `SETUP.md` and `COMMANDS.md` were ever updated to match. `-legacy`'s
  `ADR-002-hosting.md` separately understates the real service count
  (states two, the real setup needs four; this repository's own,
  differently-numbered `ADR-002` is about the container registry, an
  unrelated topic — see `ADR-002-container-registry.md`). The README's
  "Issues Resolved: 47" claim is boilerplate copied from
  `qa-pipeline-legacy`'s own README template, never a number computed
  against `-legacy`'s actual SonarCloud project — confirmed by exhausting
  every angle of SonarCloud's API: this project has never recorded more
  than 31 issues total, 6 ever resolved. The pipeline auto-deploys to QA on
  a push to `develop` or `master`, not only `main`, undocumented. A GitHub
  environment-review approval left unactioned expires silently after
  exactly 30 days, also undocumented, and was the actual explanation for a
  run that looked like an unexplained deploy failure.
- **Not everything found was a defect.** `-legacy`'s `ADR-005`
  (SQLite persistence) and `ADR-006` (frontend runtime config) are solid,
  independently-reasoned decisions, not just restatements of the course's
  own "100%-free" reference architecture example (which `-legacy`'s registry
  and hosting choices otherwise track closely). Several real,
  non-code-specific platform gotchas were found and are worth keeping
  regardless of this decision: the arm64/amd64 mismatch when pushing from
  Apple Silicon, the two-part `GITHUB_TOKEN`/ghcr.io permission requirement,
  the two-part non-root-Docker-user fix (a SonarCloud rule, then a
  `WORKDIR`-ownership runtime failure it doesn't catch), `go-sqlite3`'s CGO
  requirement ruling out a `scratch` final stage, and the general lesson
  that a plan document needs a deliberate reconciliation pass once
  implementation diverges from it, not just an initial draft. These are
  catalogued in `cloud-deploy-legacy-transferable-knowledge-results.md`, not
  repeated here.

## Decision

`forum-app-cloud-deploy` is built as a **copy** — not a fork, and not a
continuation in place — of `forum-app-qa-pipeline` at tag `v1.0.0`. TP8's
scope (a justified container registry, Docker containerization, CI/CD
deployment to QA and PROD environments with a manual approval gate between
them, and scoped secrets management) is added on top of that base.

`forum-app-cloud-deploy-legacy` is renamed with the `-legacy` suffix,
archived on GitHub (read-only, the platform's own "Archived" badge), and
kept as a reference — its two audit documents are copied into this repo's
own `docs/audits/` verbatim, the same way `qa-pipeline`'s own `docs/audits/`
holds byte-for-byte copies of `qa-pipeline-legacy`'s. Its README carries a
notice marking it superseded and pointing back to this ADR. Separately from
the archival itself: `-legacy`'s four live Render services (QA and PROD,
backend and frontend) are suspended, not deleted — a reversible action,
decided once the live-exposure audit confirmed the spoofable-auth design was
exploitable against them in practice, not just in source.

What gets ported, adapted, or discarded from `-legacy` — ADR by ADR, rule
file by rule file — is decided and recorded separately as this repository is
built, informed by both audit documents; this ADR only settles that this
repository's code and CI baseline come from `qa-pipeline@v1.0.0`, not from
`cloud-deploy-legacy`.

## Alternatives considered and rejected

**Continue building on `cloud-deploy-legacy` in place**, applying the
hardening and features it's missing (bcrypt, `MaxBytesReader`, the
`err.Error()` leak fix, `EditPost`/`EditComment`) directly to it. Rejected:
every one of those fixes already exists, implemented and tested, in
`qa-pipeline@v1.0.0` — redoing them here would not be polishing a base, it
would be re-deriving corrections that already exist against a codebase that
also still carries `ci-testing`'s own pre-hardening lineage, not just a
missing feature or two. The series' forward-propagation methodology exists
specifically to prevent this: corrections belong at their origin point and
travel forward from there, the same reasoning `qa-pipeline`'s own `ADR-000`
already recorded for itself, one stage earlier.

**Fork `qa-pipeline` on GitHub** rather than copy it. Rejected for the same
reasons `qa-pipeline`'s own `ADR-000` rejected forking `ci-testing`: a fork
is permanently marked "forked from" in GitHub's UI, is excluded from a
profile's default repository view, doesn't count its commits toward the
owner's contribution graph, and can't be reversed without contacting GitHub
support. A repository generated from a template starts with a single commit
that does count, and stands as its own portfolio entry rather than a
visually subordinate branch.

**Replace `cloud-deploy-legacy`'s contents in place** (keep the existing
repository, history, and URL, overwrite its contents with
`qa-pipeline@v1.0.0` as a new first commit). Rejected, though for a
different reason than `qa-pipeline-legacy`'s equivalent case: `-legacy` is
not a content-free mirror with nothing worth preserving — it has real,
dated ADRs, a real oral defense behind it, and now two real audit documents
explaining exactly what in it was sound and what wasn't. That history is
worth keeping intact and inspectable under its own identity (renamed,
archived), not overwritten and lost the way an in-place replacement would.

## Consequences

- This repository's Go module was already renamed from `forum-app-qa-pipeline`
  (the template-inherited name) to `forum-app-cloud-deploy` as part of this
  repo's starter-verification process, with all 12 dependent import
  statements updated in the same pass — confirmed by a real `go build`,
  `go vet`, and full test run (50/50 passing) after the rename, not assumed
  from the `go.mod` edit alone. (The "47/47" figure recorded at the time —
  in this ADR, in `docs/NEXT-STEPS-temp.md`, and in
  `docs/audits/cloud-deploy-starter-verification-checklist.md` — was already
  wrong when written: the test files did not change between the initial
  commit and the rename, only their import paths did. Corrected here once
  re-verified directly, per `docs/rules/verification.md`.)
- `qa-pipeline@v1.0.0`'s own `ADR-000` through `ADR-005` are **not** copied
  as files into this repository's `docs/decisions/` — per the cross-repo
  referencing convention (`docs/rules/documentation.md`), a decision here
  that builds on something `qa-pipeline` already resolved links to
  `qa-pipeline`'s own ADR rather than duplicating it. Three of the six
  (`ADR-000`, `ADR-004`, `ADR-005`) were kept temporarily during this
  repo's starter-verification process specifically because each one is a
  direct structural template for something this repo still had to write
  (this ADR itself, this repo's own future pipeline documentation, and its
  future authorization ADR, respectively) — `ADR-000`'s file is deleted as
  part of this ADR landing, its pattern now absorbed here.
- The security-hardening fixes from `ci-testing@v1.1.0` (`bcrypt`,
  `http.MaxBytesReader`, the `err.Error()` leak fix) are inherited
  transitively through `qa-pipeline@v1.0.0`, not re-derived in this
  repository.
- **The deeper, spoofable-authentication design — the client-supplied
  `X-User-ID` header with no real session behind it — is now this
  repository's own scope to resolve, not something inherited and left
  alone.** `qa-pipeline`'s committed prose
  (`docs/rules/testing-and-quality-gates.md`) explains why it left a
  related asymmetry unreconciled but doesn't itself say where the
  underlying model gets addressed; the commit that wrote that file
  (`8b07d0d`) does: *"already documented as a known limitation in
  ci-testing's ADR-008, deferred to cloud-deploy."* That intent never made
  it into the file's own committed text — the same category of gap this
  repo's own audit found four times over in `-legacy` (a decision that
  exists only as someone's intent, never reconciled into a document a
  later reader can find). This ADR does not repeat that gap: the handoff
  is now recorded explicitly, in `ADR-001`, citing that commit by hash.
  Accepting the scope is not the same claim as having designed or
  implemented the fix — `ADR-001` accepts scope only; the mechanism is a
  separate, future decision.
- `cloud-deploy-legacy` remains on GitHub, renamed, archived, and untouched
  in content, as the reference both audits were run against — not deleted,
  so the reasoning trail (built for TP8 → audited twice → superseded in
  favor of a clean copy, with cause) stays verifiable rather than asserted.
  Its four live Render services are suspended separately from the archival
  itself, once decided.
- This is the same category of decision `qa-pipeline`'s own `ADR-000`
  recorded for itself, applied here one stage later in the same series, for
  the same reason.
