# ADR-005: Adding edit functionality to posts and comments

**Date:** 2026-07-29
**Status:** Accepted

## Context

TP7's scope — code coverage, static analysis, and E2E integration testing —
requires, as one of its minimum integration test scenarios, a full **update**
flow for a record, alongside create and error-handling flows (see the
assignment text linked from this repo's audits, or `ingsoft3ucc/TPs_2025`'s
`07-code-coverage_integration-tests.md`). Implementing that scope for this
app means the app needs an edit operation for the Cypress suite to exercise.

## Decision

Add edit functionality to both posts and comments — not just one, since TP7's
requirement is about the update *pattern* existing somewhere real in the app,
and posts and comments are the two mutable resource types this app has. This
is new application code, built here in `qa-pipeline`, not backported to
`ci-testing` (whose own scope is closed — it never runs E2E tests and has no
need for this feature) or deferred to `cloud-deploy` (this isn't a deployment
or infrastructure concern).

This is justified under the exceeds-scope criterion in
`docs/rules/documentation.md` (category d): implementing TP7's own testing
requirements for this app means the operations those requirements describe
need to exist. Building the feature is how this repo's E2E testing work gets
to be about something real, the same way a coverage gate is only meaningful
once there's something to measure.

## Design constraint — authorization must live in the service layer

This repo already has a documented, empirical contrast between two ways the
same kind of check ("only the author can act on this resource") was
implemented in the inherited codebase:

- `DeletePost`'s authorship check lives in `post_service.go` — mockable,
  and the existing unit test (`TestDeletePost_NotTheAuthor`) genuinely
  exercises it, because the comparison happens in code the test's mock
  doesn't shortcut.
- `DeleteComment`'s authorship check lives in the repository layer's SQL
  (`WHERE ... AND user_id = ?`) — the corresponding unit test mocks the
  repository directly, so it only proves the service propagates whatever
  the mock is told to return. The real authorization logic is never
  exercised by any test in this codebase (see the app-security audit in
  `ci-testing`'s `docs/audits/`).

`EditPost` and `EditComment` must follow `DeletePost`'s pattern, not
`DeleteComment`'s: the authorship comparison (`post.UserID != userID` /
equivalent for comments) happens in the service, in Go, before any
repository call — not left to a SQL `WHERE` clause that a mocked test can't
reach. This is not a style preference; it's the difference between a test
that proves something and one that doesn't.

See [docs/diagrams/edit-delete-authorization-matrix.svg](../diagrams/edit-delete-authorization-matrix.svg) for this asymmetry laid out across all four methods.

## Scope of the feature itself

Kept intentionally minimal — this exists to make TP7's update-flow
requirement real, not to build a full editing experience:

- `EditPost`: update title and/or content of an existing post. Same
  validation rules as `CreatePost` (non-empty title/content). Author-only.
- `EditComment`: update the content of an existing comment. Same validation
  as `CreateComment`. Author-only.
- No edit history, no "edited" timestamp/label, no partial-field PATCH
  semantics — a full replace of the editable fields, matching the
  simplicity of the rest of this app's design. These can be added later if a
  concrete reason emerges; not by default.

## Alternatives considered and rejected

**Reinterpret "update" loosely** (e.g., treat changing a post's comment
count, or some other incidental state change, as satisfying the requirement).
Rejected: the consigna is explicit about a "full update flow" for a record —
building the actual operation is what the requirement is asking for.
Reinterpreting it to fit what already exists would document a testing
strategy for something other than what TP7 describes.

**Add edit only to posts, not comments.** Rejected: cheaper, but arbitrary —
nothing about TP7's requirement or this app's design favors one resource
over the other, and having both gives the E2E suite a second, independent
instance of the authorization pattern to validate, which is worth more than
the marginal implementation cost.

## Consequences

- New backend: `EditPost` and `EditComment` methods on the existing
  services, new handler routes, unit tests for both (success, validation
  failure, not-found, not-the-author) following `DeletePost`'s
  service-layer authorization pattern.
- New frontend: edit UI on `PostDetail` (post) and wherever comments are
  rendered, reusing `CreatePost`/`CommentForm`'s validation where
  reasonable rather than duplicating it.
- The Cypress "update flow" test exercises this real feature end-to-end —
  including, ideally, an attempted edit by a non-author, given this repo's
  existing standard of testing the authorization boundary, not just the
  happy path.
- Coverage strategy (`ADR-001`) is unaffected in its boundaries — these are
  Services/Components-layer additions, measured the same as existing code.