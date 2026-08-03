# Brief — Audit of forum-app-qa-pipeline-legacy before creating the new repo

## Portfolio context

This work is part of a GitHub portfolio for Octavio Carpineti, who is going on
exchange to Santa Clara University in August 2026. The portfolio is not a
collection of coursework assignments — it is a demonstration of a reasoning
pattern: why each decision was made, what was discarded and why, before how it
was implemented.

The relevant series here is: **forum-app-ci-testing → forum-app-qa-pipeline →
forum-app-cloud-deploy**, moving in a single direction. Each repo is grounded
independently in its own scope; fixes made in one repo carry forward through the
series, never backward.

- `forum-app-ci-testing`: closed, tag `v1.0.0`. Rebuilt from scratch keeping the
  scope of TP6 (unit testing). Has the full documentation standard: `/docs` with
  8 ADRs, `CLAUDE.md`, `docs/rules/`, `SETUP.md`, `COMMANDS.md`, diagrams,
  screenshots, `LICENSE`. Verified directly against the repo — the state is
  real, not aspirational.
- `forum-app-qa-pipeline-legacy` (the current repo, renamed with the `-legacy` suffix
  so there is no ambiguity about which one is being audited): is a mirror of the
  TP7 assignment from the Software Engineering 3 course, which in turn stems
  from the **original** TP6 (not the rebuilt ci-testing). It had one real work
  session on June 24, 2026 (9 commits: SonarCloud fix, partial translation,
  coverage expansion, drafting of 5 ADRs) but was never resolved from the
  corrected baseline that ci-testing already has.

## Decision already made — do not reopen this discussion

A **new** repo named `forum-app-qa-pipeline` is going to be created (the
previous repo was renamed to `forum-app-qa-pipeline-legacy`, so the clean name is
now free for the repo about to be built), starting from a **copy** (not a fork
— a fork visually subordinates the repo on GitHub and cannot be reverted
without support) of the content of `forum-app-ci-testing` at its `v1.0.0` tag.
On top of that base, the TP7 scope is added: coverage with a 70% gate, static
analysis with SonarCloud, E2E tests with Cypress, and quality gates that block
the pipeline.

The current repo (`forum-app-qa-pipeline-legacy`) **is not deleted or touched
yet**. It is kept intact as a source of information to analyze. The reason for
starting from scratch instead of continuing on top of it: it was verified that
the current code and pipeline carry inconsistencies and technical debt that
predate the fix already made in ci-testing, and it makes no sense to re-derive
those same fixes here when they already exist, resolved and documented, in the
base being started from.

**What is being rescued**: information. The 109 tests (47 Go + 47 Jest + 15
Cypress) and the 5 existing ADRs contain reasoning and test cases that go well
beyond the minimum required by the course (3 E2E flows). Before rewriting
everything from scratch, it's necessary to determine what from that is worth
porting or using as reference.

## Your task now: full audit of forum-app-qa-pipeline-legacy

**Do not build the new repo yet.** This phase is only analysis and cataloging
of the existing repo:
`https://github.com/CarpinetiOctavio/forum-app-qa-pipeline-legacy`.

### Findings already verified (do not re-derive them, they are already confirmed with evidence)

**The problems hypothesized from the ci-testing chat do not reproduce
(verified directly against the code, no need to repeat the check):**
- The coverage command already correctly specifies
  `-coverpkg=./internal/services/...` in `ci.yml`.
- The Go version in the workflow (`1.24`) is compatible with the one in
  `go.mod` (`1.24.1`) — there is no real desync.
- The GitHub Actions are already on `@v4`/`@v5`, not `@v3`.
- The Go cache already correctly points to `backend/go.sum`, not the root.
- There are no unfilled `tu-usuario`-style placeholders in the README.

**Confirmed with concrete evidence:**
- `go.mod` declares module `tp06-testing` — a course-assignment identifier, not
  the real repo name. It also leaks into a sample block in the README.
- `backend/tests/desc.md` documents "PostService (8 tests)" when the real file
  has 36 tests in that category — outdated by a factor of ~4.5x, left
  unupdated when the suite grew from TP6 to TP7.
- Documentation mixing Spanish and English: `desc.md` and
  `desc-funcionalidades.md` are entirely in Spanish; there are test fixtures in
  Spanish (`PostList.test.tsx`: "Post de otro usuario"); test names in
  Spanglish (`TestDeleteComment_NoEsAutor`, `TestDeleteComment_UsuarioNoExiste`,
  `TestDeleteComment_PostNoExiste`).
- The README does **not** have the test counts wrong — it matches exactly the
  real count from the code (109 = 47+47+15). The number that is wrong is the
  hardcoded `quality-summary` in `ci.yml`, which prints a fixed "89 tests",
  outdated and not reading real results from the jobs.
- **Real, confirmed coverage gap**: the business rule "only the author can
  delete their comment" lives in `SQLitePostRepository.DeleteComment`
  (repository layer, real SQL access). The test `TestDeleteComment_NoEsAutor`
  mocks that whole layer
  (`mockRepo.On("DeleteComment", ...).Return(errors.New(...))`) and only
  verifies that the service propagates the mock's error — it does not exercise
  the real logic. Cypress does not exercise it either: it uses `cy.intercept()`
  to mock the backend (documented in ADR-004). **There is no repository-layer
  test file at all**
  (`find backend -iname "*repository*test*"` returns nothing). The rule marked
  as "CRITICAL" in the documentation has no real coverage anywhere across the
  109 tests.

### What you need to analyze and what couldn't be verified from this chat

1. **Generalize the mocked-authorization check**: `DeletePost` is also
   documented with a "user is not the author (BUSINESS RULE)" case in
   `desc.md`. Check whether it has the same pattern as `DeleteComment` —
   authorization delegated to the repository layer, mocked in the test, with
   no real coverage. Look for the same pattern in any other mutation operation
   (create, edit, delete) across the whole backend.
2. **Run the actual suite**: confirm whether the documented coverage (97.3%
   backend, 97.64% frontend) is real by running the tests, not just what the
   README says. Confirm whether the coverage command actually works without
   errors (Octavio mentioned suspecting a "broken coverage command" — this
   could not be verified without running the toolchain).
3. **Compare `desc-funcionalidades.md` against the real `ci.yml`**: the
   ci-testing chat hypothesized that the functionality documentation might
   describe a different pipeline than the one that actually runs. Only the
   first 30 lines (academic definitions) were reviewed — the rest of the
   document (SonarCloud, Cypress, CI/CD sections) still needs to be compared
   against the real `ci.yml`, step by step.
4. **Sweep the 7 loose `desc.md` files** (`backend/internal/database`,
   `models`, `repository`, `services`, `tests`, `frontend/src`,
   `frontend/src/services`) and the old `desc/` folder
   (`desc-arquitectura.md`, `desc-funcionalidades.md`) — catalog which ones
   have rescuable conceptual information (even if it's in Spanish and needs
   translating/reformulating) and which are pure redundancy with what already
   exists in the ADRs.
5. **Catalog the 109 tests** (47 Go + 47 Jest + 15 Cypress) one by one: what
   scenario each one covers, whether the business logic it tests exists the
   same way in the `ci-testing` v1.0.0 code or changed when it was rebuilt
   from scratch, and whether the case is genuinely worth porting (real edge
   case) or redundant.
6. **Review the 5 existing ADRs** (`docs/decisions/`) with this assessment
   already done as a starting point, not to repeat it but to deepen it:
    - ADR-001 (tech-stack), ADR-003 (sonarcloud), ADR-004 (cypress): solid,
      reusable with polish of form. Confirm there is nothing else to review in
      detail.
    - ADR-002 (coverage-strategy): **broken** — it is a literal duplicate of
      ADR-001's content with the wrong title. It documents no real coverage
      strategy. It needs to be written from scratch: why 70%, why measure only
      `internal/services/`, why exclude handlers/repository/router.
    - ADR-005 (cicd-pipeline): the "Problems encountered and resolved" section
      doesn't mention external Codecov or the `develop` branch trigger — neither
      of the two was ever a documented decision, they were left over from
      scaffolding without review. The description of the `quality-summary` job
      does not match the real YAML (it says it "runs only if all previous jobs
      pass" but the YAML has `if: always()`, and it says it "provides a
      consolidated status" but it only prints fixed text). This ADR needs a real
      rewrite of those sections, not just wording adjustments.

### Expected output format

An inventory for Octavio to review and decide on — do not apply changes
directly or start building the new repo yet. The inventory needs to separate:

1. **Tests**: which to port as-is, which to adapt, which to discard as
   redundant — with the reason for each case.
2. **Conceptual documentation** (`desc.md` and related): what information is
   worth preserving (even if the format changes) and what gets discarded.
3. **ADRs**: for each of the 5, whether it's reused with polish of form or
   needs a content rewrite — with the reason.
4. **New findings** that emerge from your own analysis and are not already in
   this brief.
5. **Open questions** that require a decision from Octavio before being able
   to move forward (for example: what to do about the lack of real
   authorization coverage — is an integration test against real SQLite added
   to the new repo, or is it documented as an accepted scope limitation?).

## Non-negotiable principle

You are a conceptual auditor and drafting assistant — never a decision maker.
Everything you propose in the inventory is for Octavio to decide, with
explicit justification from you on each recommendation. No design decision is
made automatically or applied without his review.
