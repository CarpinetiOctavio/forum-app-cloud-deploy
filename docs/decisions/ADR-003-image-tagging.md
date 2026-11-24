# ADR-003: Image Tagging Strategy — Commit SHA, Never `latest`

**Date:** 2026-08-04
**Status:** Accepted

## Context

TP8's assignment explicitly forbids `latest` as an image's sole tag. That
requirement is not arbitrary; it encodes a real property continuous-delivery
practice depends on. Humble and Farley's *Continuous Delivery* names it
directly: a deployment pipeline's guarantees only hold if the artifact that
passes every gate is the *exact* artifact that ships — "build binaries once"
is the mechanism, and that mechanism only works if the binary in question
can be named unambiguously and permanently. A mutable tag breaks that naming
guarantee at the one point where it matters most: the moment a human (or an
approval gate) decides whether to promote an artifact to production.

This project's own non-negotiable constraint — the same Docker image,
unmodified, deployed to QA and then, after approval, to PROD
(`docs/rules/constraints.md`) — has no way to be *verified* without a
tagging scheme that can prove two deploys pulled the same bytes. Tagging is
therefore not a labeling convenience here; it is the mechanism that makes an
otherwise-unverifiable promise checkable.

## Decision

Tag every image pushed to ghcr.io with the triggering commit's full SHA
(`${{ github.sha }}`). **Never** use `latest` as an image's only tag, in any
workflow step, Dockerfile, or Render service configuration.

## Rationale

**Immutability is the property being purchased, not just traceability.** A
commit SHA is a content-independent, collision-resistant identifier assigned
once, by git, and never reused. Once `ghcr.io/.../backend:<sha>` exists, that
tag refers to one specific set of bytes for as long as the image isn't
explicitly deleted — there is no operation in normal pipeline use that
silently repoints it, the way every `docker push ...:latest` silently
repoints `latest`.

**Traceability follows from immutability, not the other way around.** Given
a running container's tag, the exact commit — and therefore the exact
reviewed, tested source — that produced it is recoverable with no external
bookkeeping. This is what makes an incident review or a rollback decision
start from a fact instead of an assumption.

**Rollback becomes a lookup, not a rebuild.** Because every previously
deployed SHA remains resolvable and untouched in the registry, "redeploy the
last known-good version" reduces to "redeploy this specific, already-built
tag" — no rebuild step, and therefore no risk that a rebuild produces
different bytes than what was actually running before (a real risk with
non-reproducible builds, sidestepped entirely by never rebuilding for
rollback in the first place).

**Environment parity becomes checkable, not asserted.** `docs/rules/deployment.md`
states that the QA and PROD image digests, for a given pipeline run, MUST be
identical — SHA tagging is what makes that a verifiable equality check
(compare two tag strings) instead of an unverifiable claim about two
separate build invocations having somehow produced the same output.

## Why `latest` specifically is excluded, not just discouraged

`latest` is not a version: it is a pointer that every push reassigns, so it
fails each property above. It does not say which commit is running;
redeploying "the version from before" requires knowing that version's
identity from some other source; and if QA and PROD pull `latest` at
different times, they can run different images while both report
"running latest".

The same conclusion was given as a correction during this series' academic
review of `forum-app-qa-pipeline`, where `latest` appeared in a deploy
configuration. The reasoning above is this repository's own grounding and
holds regardless of that correction.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| `latest` only | Mutable by construction (see above), and excluded by the assignment. |
| Semantic versioning (`v1.2.3`) | Requires a human or a tool to decide, for every commit, whether it's a major/minor/patch change — overhead with no corresponding benefit when every commit to `main` that passes the pipeline is, by this project's own design, a deployable unit. Semver communicates compatibility intent to external consumers of a published artifact; this image has exactly one internal consumer (Render, via this same pipeline), so that signal has no audience. |
| Branch name + SHA (`main-a3f9c2b`) | The SHA alone is already globally unique and sufficient for lookup; prefixing it with the branch adds a second piece of state that can drift from the truth (a commit can exist on more than one branch) without adding any real disambiguation. |
| Sequential build number | Monotonic and simple, but not self-describing — recovering "which commit is build #482" requires consulting the CI system's own external run history, reintroducing exactly the external-memory dependency `latest` has and a SHA does not. |

## Consequences

- Every pipeline run that reaches the Docker-build stage produces an image
  tagged with that run's commit SHA — no separate versioning scheme to
  maintain in parallel.
- Render service configurations reference a specific SHA tag, never
  `latest`, for both QA and PROD.
- The deploy step (mechanism still open, see `docs/rules/pipeline.md`
  Stage 7/9) passes the SHA tag explicitly, regardless of which mechanism
  is eventually chosen — this decision is independent of that one.
- Images accumulate in ghcr.io indefinitely under this scheme; periodic
  cleanup of old, superseded SHAs is a real operational concern for a
  long-lived project but is out of scope for this repository's current
  size and lifecycle.

## Amendment (2026-11-11): deploys reference the digest

The deploy step and Render configuration now reference images by digest
(see `ADR-007`'s amendment); the SHA tag is still published for
traceability.

## Amendment (2026-11-23): what the SHA tag guarantees and what it does not

Four statements above do not hold:

- **The tag can move.** The Rationale says no operation in normal pipeline
  use repoints a SHA tag. Re-running a run builds the image again for the
  same commit and pushes it under the same tag; builds are not reproducible
  (`docs/rules/docker.md` Rule 6), so the tag moves to a different digest
  (`ADR-007`'s amendment of 2026-11-11). A digest cannot move.
- **Rollback and environment parity go by digest, not by tag.** "Rollback
  becomes a lookup" and "Environment parity becomes checkable" rely on the
  tag naming one set of bytes. Because it can move, deploys pass the digest
  (`ADR-007`'s amendment) and rollback takes digests
  (`docs/rules/deployment.md`, Rollback).
- **The deploy mechanism is not open.** The Consequences call it "still
  open"; `ADR-007` decided it.
- **Pull request runs tag a merge commit that no branch contains.** On a
  `pull_request` event, `github.sha` is the commit GitHub creates to test
  the merge of the pull request into its base. Run `43341192421` (a pull
  request from `staging`) pushed its images under `df55240`, "Merge 15e525b
  into 7ee46bf", which is on no branch. Those images trace to a test merge,
  not to a commit in the history, and are never deployed: only a push to
  `main` deploys (`ADR-009`).

The decision is unchanged: every image is tagged with `${{ github.sha }}`,
never `latest`. The tag traces an image to the commit that produced it; the
digest identifies what is deployed.
