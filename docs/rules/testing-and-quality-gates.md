# Operating rules — testing and quality gates (forum-app-qa-pipeline)

## Coverage scope
The 70% gate applies to `internal/services/` (backend) and the components/services
layer (frontend), per `ADR-001-coverage-strategy`. Handlers, Repository, Router,
Database, Models are deliberately excluded — that is a documented decision, not a
gap. Don't propose adding coverage requirements to an excluded layer without a new
ADR justifying why the original exclusion no longer holds.

## Branch protection ruleset must track new pipeline jobs
The ruleset on `main`/`staging` does not automatically pick up new `ci.yml` jobs —
each one must be added by hand as a required status check (Settings → Rules → edit
the ruleset → add the job's exact name under "Require status checks to pass").
Verified via the GitHub API (`rulesets/19909818`), not assumed from the workflow
file alone: the ruleset currently requires all 7 of `Test Summary`, `Backend Tests
(Go)`, `Frontend Tests (React)`, `Backend Build`, `Frontend Build`, `SonarCloud
Code Analysis`, and `Cypress E2E` — every job `ci.yml` defines.

`Test Summary` is required *in addition to* the 6 individual jobs, not in their
place, even though it aggregates all of them. Reason: `Test Summary`'s own step
(`docs/decisions/ADR-004`) explicitly fails the job if any dependency's result is
`failure` *or* `cancelled` — that `cancelled` case is handled deliberately, in
code we can read. Whether the ruleset itself blocks a merge on an individual
required check landing in a `cancelled` (not `failure`) state isn't confirmed with
the same certainty — GitHub's own docs are not unambiguous on this point for
`cancelled` specifically. Keeping `Test Summary` as a required check alongside the
6 individual ones means that gap, if it exists, doesn't matter: `Test Summary`
covers it either way.

Adding a new job to `ci.yml` must still include adding it here by hand — that
part of the gap (the ruleset not auto-tracking new jobs) is unchanged by this
fix. Treat "add a new required job" as part of implementing that job, not a
follow-up task to remember separately.

## Before changing `sonar.coverage.exclusions` or any SonarCloud config
An exclusion added to make the Quality Gate pass is not a valid reason on its own —
it needs the same justification bar as an ADR (why this file/layer shouldn't be
measured, not just that measuring it is inconvenient). If a gate fails because
excluded/untested code shows up as uncovered, the fix is understanding why it's
failing, not silently widening the exclusion list.

## What a test actually proves — state it explicitly
Distinguish, in both code review and documentation, between what a test verifies
and what it appears to verify. Concretely: a mocked repository test proves the
service's logic is correct given an input — it does not prove the input itself is
trustworthy. A Cypress test using `cy.intercept()` proves UI behavior given a
response — it does not prove the backend produces that response, or that an
unauthenticated request would be rejected. When writing or reviewing a test,
name the actual guarantee, not the apparent one.

## Recreating scenarios from `forum-app-qa-pipeline-legacy`
Test scenarios catalogued in `docs/audits/qa-pipeline-legacy-transferable-knowledge-results.md`
as adding real coverage are a reference for *what* to test, not *how*. Write the
test code from scratch against this repo's actual implementation — don't adapt
`-legacy`'s test code, which was never validated against a documented testing
standard of its own (see that repo's own audit for why).

One specific consequence of "what, not how": `-legacy`'s catalog includes a
group of `ShouldPropagateError_When...Fails` tests, one per service method,
each nearly identical — verifying only that an error from the mocked
repository propagates as-is (see
`docs/audits/qa-pipeline-legacy-audit-results.md`). The scenario itself (error
propagation) is worth keeping for each method it applies to; the one-test-per-
method structure is not. When these are recreated here, consolidate them into
a single table-driven Go test per service (one `[]struct{...}` of cases, one
loop with `t.Run` per case) instead of porting one near-identical function per
method — same scenarios covered, less duplicated setup, and a new method
needing the same check becomes one more row instead of one more function.

## Authentication vs. authorization
This repo inherits `forum-app-ci-testing`'s authentication model as-is (see that
repo's `ADR-008` for the known limitation and its rationale for deferral). Tests
written here that touch authorization logic (e.g., "only the author can delete")
must document that they validate the authorization comparison, not identity
verification — the latter doesn't exist yet in this codebase. Don't write or
review such a test as if it closes that gap.

**User-existence checks in mutation methods don't add real protection, given
that same limitation.** `DeleteComment`'s `userRepo.FindByID(userID)` call
(kept as-is, not backported to `EditComment` or reconciled with `DeletePost`,
which has no equivalent check) only confirms the given `userID` exists as a
row in `users` — it can't verify the request actually comes from that user,
because there's no session or token behind `X-User-ID` for it to check
against. Any `userID` that has ever authored a post or comment already
exists in the table by definition, so a forged `X-User-ID` reusing that
value passes this check the same way a legitimate request would, whether or
not the check is present. This is why the asymmetry between `DeletePost`
(no user-existence check), `DeleteComment` (has one), and `EditPost`/
`EditComment` (neither has one, mirroring `DeletePost`) is being left as-is
rather than unified: with no real identity verification behind `X-User-ID`,
none of these checks change what's actually protected, so reconciling them
— including editing `DeleteComment`, an already-merged method with its own
tests — isn't worth the risk for a gain that's cosmetic, not functional.