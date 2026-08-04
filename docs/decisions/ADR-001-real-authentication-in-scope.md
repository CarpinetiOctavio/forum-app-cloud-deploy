# ADR-001: Real authentication is in scope for this repo — accepted, not yet designed

**Date:** 2026-08-04
**Status:** Accepted (scope only — the mechanism is a separate, future decision)

## Context

The application this series builds on has never had real authentication.
Every mutation endpoint trusts a client-supplied `X-User-ID` header with no
session, token, or password check behind it once login succeeds —
`forum-app-ci-testing`'s own `ADR-008` first documented this as a known,
accepted limitation, deferred rather than fixed at the time, because
`ci-testing`'s scope was unit testing, not application security.

`qa-pipeline` inherited the same design and made its own decision about it,
recorded in `docs/rules/testing-and-quality-gates.md`: it deliberately left
a related asymmetry unfixed (`DeleteComment`'s `userRepo.FindByID` check,
absent from `DeletePost`/`EditPost`/`EditComment`) because, with no real
identity verification behind `X-User-ID`, reconciling that asymmetry
"isn't worth the risk for a gain that's cosmetic, not functional." That
file's own committed text explains *why the asymmetry stays* but does not
say *who addresses the underlying model* — that half of the reasoning is in
the commit that introduced the file, not in the file itself
(`8b07d0d`, `forum-app-qa-pipeline`): *"neither check adds real protection
under this app's current X-User-ID-header authentication model — already
documented as a known limitation in ci-testing's ADR-008, **deferred to
cloud-deploy**."* `qa-pipeline`'s `CLAUDE.md` independently confirms the
shape of that handoff without naming authentication specifically: it scopes
"Docker, containerization, deployment" out to this repository by name,
while never scoping authentication to itself the way it scopes those away.

This repository's own audit of `cloud-deploy-legacy` closed the loop on
severity: the design isn't theoretical. A request presenting no credential
beyond a client-supplied `X-User-ID` value was confirmed, live, against
`-legacy`'s real deployed QA service, to delete another account's content —
reproduced independently twice, once by Claude and once by Octavio himself
by hand (`cloud-deploy-legacy-audit-results.md` §7).

This repository's own `docs/rules/constraints.md` and `CLAUDE.md`, as
first drafted, stated the opposite of this ADR's decision — that modifying
the authentication model is `qa-pipeline`'s scope, not this repo's. That
was written from `cloud-deploy-legacy`'s own audit alone, before
`qa-pipeline`'s real files (not `-legacy`'s) were checked for an explicit
handoff. They're corrected as a consequence of this ADR, not independently.

## Decision

Real authentication is **in scope** for `forum-app-cloud-deploy`. This
repository accepts the handoff `qa-pipeline` recorded for itself in
`8b07d0d` and does not defer it further — there is no fourth repo in the
series for it to land on next.

**This ADR decides scope only, not mechanism.** Whether the real fix is
JWT, server-side sessions, or another approach; how it interacts with the
container/deployment work that is this repo's original TP8 scope; and
what changes to `internal/services/`, `internal/handlers/`, and the
frontend it requires — none of that is decided here. It gets designed,
with real alternatives weighed, and recorded in its own ADR once that
design work happens — the same discipline already applied to this repo's
deploy-mechanism decision (`docs/rules/pipeline.md`'s Stage 7/9 section):
accept the scope and state it plainly now, don't write the mechanism ahead
of actually deciding it.

## Rationale

**Why accept now, in an ADR of its own, rather than let it stay an
unstated assumption**: the same failure this repo's own audit found
repeatedly in `-legacy` — a real decision that exists only in someone's
intent, never reconciled into a document a future reader can find — is
what happened here too, one repo earlier in the series. `qa-pipeline`'s
real intent to defer this forward existed from the moment `8b07d0d` was
written; it just never made it into committed prose. Writing this ADR now,
explicitly, with the commit that expressed the original intent cited by
hash, closes that gap here rather than repeating it a third time.

**Why this repo and not a new one**: the series has three stages
(`ci-testing` → `qa-pipeline` → `cloud-deploy`), each with its own closed
scope; nothing about the portfolio's structure anticipates a fourth. A
limitation that started as "not this repo's scope, defer it" at every
earlier stage has to resolve somewhere the chain ends, or the portfolio's
own forward-propagation methodology (`ADR-000`) never actually delivers a
version of the app with real authentication — only a growing paper trail
of repos agreeing it's someone else's problem.

**Why scope-only, not mechanism, in this same ADR**: `docs/rules/documentation.md`'s
own rule — documentation follows implementation, not the other way around
— exists because of exactly this failure mode, confirmed four separate
times in `-legacy`'s own history (§ of the audit). Prescribing JWT or
sessions here, before any real design or code exists, would be writing the
same kind of pre-implementation plan that never got reconciled with reality
in four of `-legacy`'s own documents. Accepting the scope is a fact that's
true today and doesn't need implementation to back it; the mechanism isn't.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Leave it exactly where `qa-pipeline` left it — inherited, accepted, unaddressed, permanently | Rejected: contradicts `qa-pipeline`'s own recorded intent (`8b07d0d`) to hand this forward, and leaves a confirmed, live-exploitable design in whatever this repo eventually deploys, with no repo left in the series to defer it to again. |
| Design and implement the fix now, in the same ADR that accepts the scope | Rejected: no real analysis of the mechanism has happened yet — writing a solution before that analysis is the precise mistake this repo's own audit spent two full passes cataloguing in `-legacy`. Scope and mechanism are separable decisions; only one of them is actually settled right now. |
| Treat this as `qa-pipeline`'s unfinished business, ask that repo's own future session to write the fix there instead | Rejected: `qa-pipeline`'s own `CLAUDE.md` scopes deployment/containerization away from itself by name, and its own commit already states the deferral was intentional, not an oversight to send back. Re-opening `qa-pipeline` to do this would itself violate the series' forward-propagation rule — corrections travel forward, not backward. |

## Consequences

- `docs/rules/constraints.md`'s "out of scope" section and `CLAUDE.md`'s
  scope boundary, both of which currently state the opposite of this
  decision, are corrected as part of this ADR landing — not left
  contradicting a decision that supersedes them.
- A follow-up ADR, written once real design work happens, decides and
  documents the actual mechanism (JWT vs. session vs. another approach),
  its alternatives, and its interaction with this repo's container/deploy
  work — this ADR does not pre-empt it. **That follow-up ADR MUST read
  `qa-pipeline`'s own
  [`ADR-005-edit-functionality`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/blob/main/docs/decisions/ADR-005-edit-functionality.md)
  before being written** — not kept as a local file here, per the
  cross-repo referencing convention, but the reasoning it contains is a
  direct prerequisite: replacing the raw `X-User-ID` header with a verified
  identity necessarily touches every existing authorization check that
  currently reads it (`DeletePost`'s service-layer comparison,
  `DeleteComment`'s SQL `WHERE` clause), not just the login/session flow.
  `ADR-005` already established, with real evidence from this codebase,
  that the service-layer placement is the one that's actually testable and
  the SQL placement is a blind spot. The follow-up ADR inherits that
  finding by reading it there, rather than re-deriving it or duplicating
  it into a file of this repo's own.
- Until that follow-up ADR lands and is implemented, this repository's own
  containerized, deployed version of the app carries the same
  spoofable-authentication design `-legacy` had, confirmed live-exploitable
  — accepting the scope to fix it is not the same claim as having fixed it
  yet. Any live services this repo deploys before that follow-up work
  lands inherit the same exposure `-legacy`'s did.
- `qa-pipeline`'s own files are not edited to state this handoff more
  explicitly — per the series' forward-propagation rule, corrections travel
  forward only. This ADR, citing `8b07d0d` by hash, is where that
  reconciliation lives from now on.
