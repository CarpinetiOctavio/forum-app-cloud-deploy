# ADR-001: Coverage strategy — 70% threshold, services/components-only scope

**Date:** 2026-07-28
**Status:** Accepted

## Context

TP7's scope adds a coverage gate that blocks the pipeline below a minimum threshold. A coverage percentage is only a meaningful signal if it's measured over code that actually contains business logic — measuring the entire codebase, including HTTP wiring, database setup, and struct definitions, either forces writing low-value tests for trivial code, or produces a precise-looking number that says nothing about whether the business rules are actually verified.

**Why this repo needs its own coverage-strategy ADR, instead of linking to `ci-testing`'s [ADR-002](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-002-testing-scope-services-layer.md):** `ci-testing`'s `ADR-002` settles a *testing-scope* question — which layers get a unit test at all (Services, plus a handful of frontend components with a business-rule branch). It says nothing about a numeric threshold enforced in CI, because `ci-testing` never had a coverage gate — TP6 required demonstrating unit testing and mocking, not a pass/fail percentage. A coverage gate is a distinct concern from testing scope: it's possible to have the right layers under test (as `ci-testing` does) with no threshold blocking the pipeline on a regression, and it's possible to enforce a threshold over the wrong layers. Both questions happen to land on the same boundary here (Services/components), but that's this ADR's own conclusion, arrived at independently below — not inherited by reference, because `ci-testing`'s ADR doesn't address the question this one is answering.

## Decision

Minimum coverage threshold: **70%**, enforced independently in both backend and frontend:
- Backend: a CI step parses `go tool cover` output and fails with a non-zero exit code if the percentage is below 70.
- Frontend: Jest's native `coverageThreshold` in `package.json` (`branches`, `functions`, `lines`, `statements`, all 70).

Coverage is measured only over `backend/internal/services/` and `frontend/src/components/**` + `frontend/src/services/**` — not the rest of the codebase.

## Rationale

**Why 70% and not 100%?**

70% is a pragmatic floor: high enough to catch a real lack of testing discipline, low enough that it doesn't force chasing coverage on defensive branches that can't realistically be triggered or on boilerplate. A 100% requirement tends to produce tests written to satisfy the metric — assertion-free tests that just execute a line — rather than tests that verify behavior.

**Why only `services/` (backend) and `components/` + `services/` (frontend)?**

`backend/internal/services/` is where validations and business rules live — input validation, authorization checks, coordination between repositories. It's the layer with actual branching logic, which is what a coverage gate is meant to protect. The same argument applies to the frontend: components hold the UI logic, and the frontend `services/` layer holds the API-calling logic; everything excluded below is either wiring, generated boilerplate, or type-only code with no behavior to cover.

## Coverage exclusions

**Backend** (`internal/`):

| Package | Reason excluded | How it's tested instead |
|---|---|---|
| `handlers/` | Only maps HTTP request/response to service calls; no business logic of its own | Cypress E2E tests exercise the handlers through real HTTP requests |
| `repository/` | Executes SQL directly against a real connection; not practical to unit-test without a live database | Mocked in every service-layer test, so services are tested against the `Repository` interface, not a concrete implementation |
| `router/` | Route registration only — no conditional logic | Exercised indirectly by every E2E test that hits an endpoint |
| `database/` | Schema definition and connection setup — configuration, not logic | Out of scope for unit tests; would require integration tests against a real SQLite file |
| `models/` | Plain structs with JSON tags — no methods, no logic | N/A — nothing to execute |
| `tests/mocks/` | Test doubles used *by* the tests, not code under test | N/A — same reasoning as the frontend's `__mocks__/**` row, below |

Backend `cmd/` (entry point) is validated by the build succeeding, not by coverage.

**Frontend** (`src/`):

| Path | Reason excluded |
|---|---|
| `App.tsx` | Root component — only orchestrates other components and top-level view state; no logic of its own |
| `index.tsx` | React entry point — framework bootstrap |
| `reportWebVitals.ts` | Create React App boilerplate |
| `types/**` | Type/interface definitions only |
| `__mocks__/**` | Test doubles used *by* the tests, not code under test |

**SonarCloud-only** (`cypress/`, outside `frontend/src/` and never part of Jest's `collectCoverageFrom` scope above — this isn't a coverage-gate exclusion, it only exists in `sonar-project.properties`):

| Path | Reason excluded |
|---|---|
| `cypress/support/**` | Not a test double — this is real Cypress support code (e.g. the `login` custom command's own `cy.visit`/`cy.get`/`cy.wait` sequence). Excluded because no tool in this pipeline currently instruments it: Jest never executes it, and Cypress doesn't produce a coverage report for its own support files without `@cypress/code-coverage`, which wasn't set up (out of scope for the Cypress work that introduced this file). See Consequences below. |

## Alternatives considered and rejected

**100% coverage requirement.** Rejected: forces testing of trivial or purely defensive code paths, and incentivizes tests written to satisfy the number rather than to verify behavior.

**Measure the entire codebase, no exclusions.** Rejected: would require integration tests against a real database and HTTP server to meaningfully cover `repository/`, `handlers/`, and `database/` — those layers have no branching logic of their own to protect, so the cost isn't justified by the signal gained.

**A lower threshold (e.g. 50%).** Rejected: too permissive — would pass with large parts of the services layer left untested.

## Consequences

- Backend and frontend coverage scopes are intentionally aligned: `sonar.coverage.exclusions` (see [ADR-002](ADR-002-sonarcloud.md)) and Jest's `collectCoverageFrom` exclude the same categories of non-logic code, so both tools report coverage over comparable ground.
- Because `repository/` is excluded from measurement, any business rule implemented at that layer instead of in `services/` will not be reflected in the coverage number even without a dedicated test for it. This is a known limitation of scoping coverage to `services/` only — the coverage gate alone cannot guarantee every business rule is verified regardless of which layer it happens to live in.
- The Go build (`cmd/`) and the router are validated by the pipeline succeeding, not by a coverage percentage — a compilation failure or a broken route registration surfaces as a failed build/E2E job rather than a coverage regression.
- The backend table's `tests/mocks/` row was added when SonarCloud's `sonar.coverage.exclusions` was configured to mirror this table ([ADR-002](ADR-002-sonarcloud.md)) and, being a full-repo scan, made the gap visible: the frontend side already excluded `__mocks__/**` for being test doubles, but the backend table never listed its own equivalent. This isn't a new criterion — the reasoning was already written on the frontend side of this same table — it's applying it to close an asymmetry between the two halves that the `go test -coverpkg=./internal/services/...` gate alone never surfaced, because it never looked outside `services/` in the first place.
- The `cypress/support/**` exclusion is a known gap, not a deliberate design choice like the rest of this table: SonarCloud's own default Quality Gate condition on new-code coverage (≥80%, distinct from and stricter than this repo's own 70% gate) failed on a PR whose only real application-code change (`CommentList.tsx`) was itself at 85.71% — the failure came entirely from `cypress/support/commands.ts`'s 8 new lines sitting at 0%, confirmed via SonarCloud's own measures API before excluding anything. The exclusion exists because no tool currently measures that code, not because measuring it would be low-value the way `__mocks__/**` or `tests/mocks/` are — `commands.ts` carries real logic (a login sequence other specs depend on). It's revisable the moment Cypress coverage instrumentation (`@cypress/code-coverage` or equivalent) is added to this pipeline; until then, this and any future `cypress/support/**` file carry zero verification signal from any gate in this repo.
