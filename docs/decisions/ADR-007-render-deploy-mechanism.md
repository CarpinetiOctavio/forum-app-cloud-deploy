# ADR-007: Render Deploy Mechanism — Deploy Hooks for Triggering, API for Status Only

**Date:** 2026-08-04
**Status:** Accepted

## Context

`docs/rules/pipeline.md` left this decision explicitly open since it was
first written: `cloud-deploy-legacy` documented a Render **deploy hook**
mechanism in four separate files (its own `ADR-002-hosting.md`,
`deployment.md`, `pipeline.md`, `constraints.md` — unrelated to, and
predating, this repository's own differently-numbered `ADR-002`, about the
container registry) and actually implemented a different one — the **Render
REST API**, authenticated with an account-scoped `RENDER_API_KEY` — without
ever reconciling the two (`cloud-deploy-legacy-audit-results.md` §4). This
repository does not inherit either choice by default; per
`docs/rules/documentation.md`'s "documentation follows implementation"
rule, the decision is made once, deliberately, before either mechanism is
implemented — not asserted from what `-legacy` happened to run.

**Why the mechanism matters beyond "which HTTP call to make":** `ADR-003`'s
commit-SHA tagging strategy is not just an auditability preference. Its real
value — surfaced directly in this project's own TP8 oral defense, when
asked to justify SHA tagging — is the ability to build and validate several
images through the complete pipeline in advance, then choose which specific,
already-validated one to deploy at the moment it's needed: one image built
and gated for a scheduled release window, a different one held ready as a
rollback target, both already proven by every quality gate before the
deploy decision is made. That capability is only real if the deploy
mechanism can target an exact tag or digest, not merely "redeploy whatever
this service currently points to." Any mechanism that can't do that would
quietly defeat the reason `ADR-003` exists.

## Investigation — verified directly against Render's current documentation

Both candidate mechanisms were checked against Render's real, current API
reference (`render.com/docs/deploy-hooks`,
`api-docs.render.com/reference/create-deploy`,
`api-docs.render.com/reference/authentication`), not assumed from
`-legacy`'s precedent or general familiarity with either. Four findings,
each independently verifiable:

1. **Both mechanisms can target an exact tag or digest — functionally
   equivalent on the one property that matters most.** A deploy hook
   accepts an `imgURL` query parameter
   (`?imgURL=ghcr.io/.../backend:<sha>`); the API's `POST
   /v1/services/{id}/deploys` accepts an `imageUrl` body parameter. Render's
   own documentation states the identical constraint for both, nearly
   word-for-word: *"All components of `imgURL` besides the tag or digest
   must match your service's default image URL"* (deploy hooks) and *"The
   host, repository, and image name all must match the currently configured
   image for the service"* (API). Neither mechanism can point a service at
   an arbitrary external image; both can select among tags of the image the
   service already tracks. The premise that a deploy hook can only replay
   `latest` is false — this was an assumption made and then corrected within
   this same decision process, not settled on first read.
2. **Render does not offer a service-scoped or read-only API key.** Two
   independent, current sources confirm this: `render.com/docs/mcp-server`
   states *"Render API keys are broadly scoped. They grant access to all
   workspaces and services your account can access"*; `api-docs.render.com/reference/authentication`
   states *"An API key provides access to all workspaces you belong to."*
   There is no narrower credential to request instead — this rules out
   "just use a scoped key" as an option before it can be considered.
3. **A deploy hook's response carries a deploy ID that unlocks read-only
   status polling.** A successful hook call returns `{"deploy":
   {"id": "dep-..."}}`. That ID can be passed to the API's own read-only
   `GET` "Retrieve deploy" endpoint to poll real deploy status. This
   resolves a real, previously-open question in this repository's own
   `docs/rules/smoke-testing.md` — how the pipeline knows a deploy has
   actually finished, rather than assuming completion from the trigger
   call's own status code, before running the smoke test against it.
4. **Edge case, not a defect**: if the Render workspace's overlapping-deploy
   policy is set to `Wait` and a deploy is already in progress, a hook call
   returns `202` with no deploy ID. Relevant specifically to the
   back-to-back-deploys scenario the tagging strategy is meant to support
   (§ above) — handled in Consequences, not a reason to discard the
   mechanism.

## Decision

**Deploy hooks trigger every deploy** (QA backend, QA frontend, PROD
backend, PROD frontend — four hook URLs, one per service), each called with
the `imgURL` parameter set to the exact SHA-tagged image this pipeline run
built. **The Render API, authenticated with `RENDER_API_KEY`, is used
exclusively for read-only status polling** (`GET` "Retrieve deploy") to
determine when a triggered deploy has finished, feeding
`docs/rules/smoke-testing.md`'s requirement — never to trigger a deploy or
modify any service's configuration.

## Rationale

**Why not the API for triggering too, now that the functional gap turned
out not to exist.** With both mechanisms equally capable of targeting an
exact tag, the deciding factor is exposure if a secret leaks — and here the
two are not close. A leaked deploy hook URL, per Render's own documented
`imgURL` constraint, can only ever trigger a redeploy of the same image path
that one service already tracks; it cannot read account data, enumerate
other services, or point a service at an image outside its own registry
path. A leaked `RENDER_API_KEY`, per Render's own confirmed broad-scoping
(finding 2), grants that access across every service in the account. Four
narrowly-scoped secrets, each contestable independently, is a materially
smaller blast radius than one broadly-scoped secret used for the same
number of write operations — the same reasoning `ADR-001` and `ADR-002`
already apply to credential surface generally, applied here to the one
place this repository actually holds write access to production
infrastructure.

**Why the API key is still needed, and why calling it "read-only" requires
a precise caveat, not a comfortable one.** Finding 2 means there is no way
to issue a key that is *incapable* of writing — the key used for status
polling is exactly as powerful as one used for triggering deploys; the
constraint that it's only ever called for `GET` requests is enforced by
this pipeline's own code, not by anything Render grants at the credential
level. This decision reduces exposure by minimizing what the key is *used*
for, not by reducing what it *could* do if it leaked. That distinction is
recorded here explicitly so it doesn't get silently overstated later as "we
use a read-only key."

**Why this doesn't reopen the "simpler is one secret" argument for the
API.** One secret is operationally simpler to provision, but this project's
own established pattern (`ADR-001` accepting real authentication as
in-scope specifically because a spoofable design was confirmed
exploitable) already establishes that this series treats a larger,
confirmed attack surface as worth the operational cost of avoiding it, not
as a convenience trade worth taking for its own sake.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Render API for both triggering and status (what `-legacy` actually ran) | Rejected once the functional gap motivating it (belief that hooks couldn't target a specific tag) was checked and found false — with equivalent functionality, there's no longer a reason to accept the larger blast radius. |
| Deploy hooks only, no status polling, smoke test fires on a fixed delay after the trigger call | Rejected: a fixed delay is either too short (smoke test runs against the still-terminating old container, per Render's zero-downtime swap) or wastefully long — real status polling, from finding 3, is available for the cost of one read-only credential already required for exactly this purpose. |
| A service-scoped or read-only Render API key | Does not exist — confirmed against two current, independent sources (finding 2). Not a real option to weigh, not a gap in this analysis. |
| Store the deploy hook URLs unauthenticated as build-time constants rather than GitHub Secrets | Rejected without extended consideration: `docs/rules/constraints.md` already treats every Render credential, hook URLs included, as a value that must be a GitHub Secret — a hook URL is a bearer credential (anyone with it can trigger a deploy) regardless of its narrower blast radius relative to an API key. |

## Consequences

- `docs/rules/pipeline.md`'s Stage 7/9 section is rewritten to state this
  mechanism as a requirement, replacing the two-option list it held while
  this decision was open.
- `docs/rules/deployment.md`'s Secrets section is rewritten to name the
  real secrets this decision requires: `RENDER_DEPLOY_HOOK_{BACKEND,FRONTEND}_{QA,PROD}`
  (four, one per service, each a bearer URL) and `RENDER_API_KEY` (one,
  repository-level, used only for the read-only status call) — not the
  `RENDER_DEPLOY_HOOK_QA`/`RENDER_DEPLOY_HOOK_PROD` two-secret shape
  `-legacy`'s documentation (never its implementation) described, and not
  the `RENDER_SERVICE_ID_*` four-secret shape `-legacy` actually used
  instead.
- Every deploy call MUST include `imgURL`/`imageUrl` set to this pipeline
  run's exact commit SHA — omitting it would silently fall back to
  whatever tag the service last had configured, defeating `ADR-003`'s
  guarantee.
- The pipeline MUST branch on a hook response containing `202` with no
  deploy ID (finding 4) — treated as "a deploy is already in progress,"
  not as a failure, with a retry/backoff rather than an immediate error.
- `RENDER_API_KEY`'s code paths are restricted to `GET` calls by this
  pipeline's own implementation, not by anything Render enforces — any
  future change that adds a write call using this same secret must be
  treated as reopening this ADR's blast-radius analysis, not as a minor
  extension.
