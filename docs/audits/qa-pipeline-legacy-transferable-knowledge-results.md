# Transferable knowledge findings — forum-app-qa-pipeline-legacy sweep

Findings produced following `qa-pipeline-legacy-transferable-knowledge-brief.md`. Sweep run 2026-07-28 — nothing applied, read-only.

---

## 1. `ADR-004-cypress.md` from `-legacy` — reread in full

**Transferable:**
- **`cy.reload()` + inconsistent logout → flaky test, removed.** The ADR
  says: "the test was flaky due to `cy.reload()` causing inconsistent
  logout behavior". This is Cypress behavior (`cy.reload()` resets the
  page's JS context) combined with an app pattern (in-memory React session,
  no `localStorage`) that **is the same pattern in `qa-pipeline`** —
  confirmed in `ci-testing`'s security audit: "no finding... all auth state
  is transient in React memory." Since `qa-pipeline` inherits that same
  session design, any Cypress test there that uses `cy.reload()` expecting
  the session to persist is going to run into the same problem. Worth
  documenting *before* writing the tests, not rediscovering the hard way.
- The rest of the "CI/CD integration" section (the official action manages
  startup/health-check/cleanup, `npm start &` causes race conditions and
  zombie processes) is the same information as finding #4 already listed in
  "Already found" — not new, just confirms it's well captured there.

**Not transferable / already covered:**
- "Node.js 20+ required — Cypress 13+ dropped support for Node 18" — already
  in `qa-pipeline`'s own `ADR-003-cypress.md` (written in the previous
  session as a Consequence), no need to bring it over again.
- The detail of the 15 specific tests (5+5+4+1) and the "delete own post"
  test removed for being flaky — is inventory of `-legacy`'s code, not tool
  behavior. Not transferable.

## 2. `ADR-001-tech-stack.md` from `-legacy` — reread in full

**Nothing transferable beyond the stack decision itself.** The whole
document (Context/Decision/Rationale/Alternatives/Consequences) is
specifically about why Go+React+SQLite — there's no loose operational tool
fact (it doesn't mention SonarCloud, Cypress, or npm anywhere). Confirms
what was already decided in the brief: there's nothing to rescue from here
beyond the stack decision itself, which doesn't get ported as its own ADR
anyway.

## 3. Concrete comparison: `ADR-002-coverage-strategy.md` (`-legacy`, rewritten
this session) vs. `ADR-001-coverage-strategy.md` (`qa-pipeline`, already existing)

They have the same backbone (why 70%, why only services/components, the
same 3 rejected alternatives, the same trade-off note about excluding
`repository/`). Concrete differences found:

- **`-legacy` has an "Achieved coverage" section with real numbers**
  (97.3%/97.64%, 109 tests, real `go test`/`npm test` command and output).
  **`qa-pipeline` has no such section at all** — correct, that's exactly
  the kind of empirical result that shouldn't be copied yet. Nothing to fix
  here, it's the separation working as intended.
- **`-legacy`'s backend exclusion table has 7 rows, `qa-pipeline`'s has 5.**
  `-legacy` includes `backend/cmd/**` and `backend/tests/mocks/**` as table
  rows (verified against `-legacy`'s real `sonar-project.properties`, which
  does exist). `qa-pipeline` only mentions `cmd/` in a loose sentence below
  the table, and doesn't mention `tests/mocks/**` anywhere — because
  `qa-pipeline` doesn't have a `sonar-project.properties` yet (confirmed it
  doesn't exist), so that table was written against the directory structure
  alone, not against a real config. **This isn't necessarily an error to
  fix right now** — it's the logical consequence of `qa-pipeline` not
  having SonarCloud implemented yet — but if, once `qa-pipeline`'s real
  `sonar-project.properties` gets implemented, it ends up needing to
  exclude `tests/mocks/**` too (very likely, it's going to have the same
  generated-mocks folder), it's worth adding it to `ADR-001`'s table at that
  point, not assuming the current table is already complete.

## 4. `docs/diagrams/layered-architecture.svg` — reread in full, nothing hidden

There are no comments or extra text hidden in the SVG beyond what renders
visibly — the `<!-- Frontend -->`, `<!-- API Layer -->`, etc. are just the
SVG's own organizational labels, with no extra information.

**There IS a concrete problem for when it gets moved:** the diagram's
footer literally says `"Coverage scope and exclusion rationale:
ADR-002-coverage-strategy.md"` — that's the filename in `-legacy`. In
`qa-pipeline` the equivalent ADR is `ADR-001-coverage-strategy.md` (a
different number). If the SVG is copied as-is, that text ends up pointing
at a file that doesn't exist in `qa-pipeline`. That line in the SVG needs
to be edited when moving it, it's not a plain copy.

**Separate note about the diagram's own content** (not a new finding, it's
a classification clarification): the real constructor names
(`NewAuthService`, `NewPostService`, `NewSQLiteUserRepository`,
`NewSQLitePostRepository`) and the `database.db` filename are, strictly
speaking, empirical findings about `-legacy`'s code — brief category 1, not
transferable as a general rule. But since it was already verified in this
same session that `qa-pipeline`'s `database.go` and constructors are
**identical** to `-legacy`'s (both are copies of `ci-testing`), in this
specific case they do apply as-is — not because they're tool behavior, but
because the source code is the same. Worth making that explicit in the
handoff, so the exception doesn't get generalized.

## 5. Other things that came up, same criterion

- **Cypress: dead `.ts` files in `cypress/support/`** (a finding from
  `audit-results.md` in `-legacy`, not from an ADR — flagging it anyway
  since the brief says "floor, not ceiling"): `cypress.config.ts` points
  `supportFile` at `cypress/support/e2e.js`, not the `.ts` one;
  `commands.ts`/`e2e.ts` remain as Cypress's default scaffold, unused, while
  the real code lives in the `.js` versions. This **is transferable as a
  setup-hygiene warning**: if `qa-pipeline` generates Cypress with
  `cypress open`/`cypress init`'s default TS scaffolding and then writes the
  real commands in `.js` without deleting the `.ts` files or pointing
  `supportFile` correctly, it's going to end up with the same problem — it
  doesn't depend on this repo's particular code, it depends on how Cypress
  generates its default scaffold when the project is TypeScript.
- Found nothing additional in `ADR-003-sonarcloud.md` or
  `ADR-005-cicd-pipeline.md` beyond the 5 points already listed in "Already
  found" — reread them in full for this response and they confirm what was
  already flagged, no new findings there.
