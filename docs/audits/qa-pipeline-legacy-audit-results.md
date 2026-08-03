# Audit results — forum-app-qa-pipeline-legacy

Inventory produced following `auditoria.md`. This is an input for Octavio to decide on — no recommendation here is applied, all of them have an explicit reason. No file in the repo was touched except for the creation of this document.

---

## 0. Executive summary

- **The coverage command is not broken.** Backend and frontend run clean, real coverage (97.3% / 97.64%) matches exactly what's documented in the README.
- **The real authorization gap is more specific than it looked**: `DeletePost` does validate authorship in the service (well tested). Only `DeleteComment` delegates the rule to the repository and mocks it in the test — the original brief finding does not generalize to all mutations, it's specific to one method.
- **New finding more serious than the previous one**: the entire authentication model is spoofable — there's no JWT/session, any request authenticates with an `X-User-ID` header with no verification, and passwords are stored in plaintext. This reframes the question of "do we add a real integration test?" — fixing the test doesn't fix the design.
- **Of the 109 tests, 57 have no equivalent in `ci-testing@v1.0.0`** (24 backend + 18 frontend + 15 Cypress) — porting won't mostly be "adapting," it will be "recreating against the already-rebuilt base."
- **Of the 9 `desc.md`/`desc/` documents, most have outdated numbers**; the one that contributes the most unique conceptual reasoning is `desc-arquitectura.md`, but it's also the one with the most stale data.
- **Of the 5 ADRs**: 001, 003, 004 solid. 002 is broken (needs to be rewritten almost entirely). 005 needs a targeted rewrite of 2 sections.

---

## 1. Real environment (verified, phase 1)

| Command | Result |
|---|---|
| `go test ./tests/services/... -v -cover -coverpkg=./internal/services/...` | 47/47 tests OK, **97.3%** coverage — matches README |
| `npm test -- --coverage --watchAll=false` | 47/47 tests OK (8 suites), **97.64%** coverage — matches README |

The suspicion of a "broken coverage command" **does not reproduce**. `npm install` required installing `node_modules` from scratch (it wasn't present), with no errors — 58 vulnerabilities reported by `npm audit`, all in transitive dependencies of `react-scripts` (Create React App, deprecated), not actionable without eject/migration — informational mention, not a finding within this scope.

---

## 2. Mocked authorization — generalized pattern (phase 2)

The backend only has 4 mutations (there's no `Edit`/`Update` at all, confirmed by grep): `CreatePost`, `DeletePost`, `CreateComment`, `DeleteComment`. Of these, only `DeletePost` and `DeleteComment` have an authorship rule.

| Operation | Where does the rule live? | Does the test actually exercise it? |
|---|---|---|
| `DeletePost` | **Service** (`post_service.go:113`, `if post.UserID != userID`) | **Yes** — `TestDeletePost_NoEsAutor` only mocks `FindByID`, the authorship comparison runs on real code, it's not stubbed. |
| `DeleteComment` | **Repository** (SQL `WHERE id=? AND post_id=? AND user_id=?`, `post_repository.go:178-195`) | **No** — `TestDeleteComment_NoEsAutor` mocks `mockRepo.On("DeleteComment",...).Return(errors.New(...))` directly; the real SQL never runs. |

In other words: the brief's finding **does not generalize** to all mutations — it's specific to `DeleteComment`. `DeletePost` is genuinely well covered.

**On the Cypress test** (`posts.cy.js:110`, "should not show delete button on other users' posts"): confirmed that it is **purely UI hiding**. It uses `cy.intercept('GET', '**/api/posts', {...})` to mock the whole backend response, and only checks `cy.contains('Eliminar').should('not.exist')`. It never sends a real DELETE request. The fact that the button doesn't appear in the DOM doesn't mean the server rejects the request if someone sends it straight to the API (with curl, Postman, etc.) — these are two different guarantees and today only the first one exists. **There is no test, in any of the 3 suites, that verifies the server's real rejection of a `DeleteComment` attempt by a non-author.**

There is no repository-layer test file at all (confirmed, `find backend -iname "*repository*test*"` returns nothing) — the SonarCloud coverage gate and `go test -coverpkg` **explicitly exclude** `backend/internal/repository/**` (`sonar-project.properties`), so no gate in the pipeline can detect this gap even if it looked for it.

---

## 3. Test inventory (phase 4) — 109 tests

Legend for the "ci-testing v1.0.0 baseline" column: **(a)** an equivalent exists (renamed to English) → candidate to **discard as redundant**. **(b)** doesn't exist → candidate to **port/recreate** because it adds coverage the new base doesn't have.

### 3.1 Backend Go — `auth_service_test.go` (11 tests)

The 11 tests (Register: Success/EmptyEmail/InvalidEmail/ShortPassword/EmptyUsername/DuplicateEmail; Login: Success/EmptyEmail/EmptyPassword/UserDoesNotExist/IncorrectPassword) are **(a)** — all exist in `ci-testing@v1.0.0`, renamed to English (`TestRegister_EmptyEmail`, etc.), same scenario, same mock-only.

**Recommendation:** discard as "port," already resolved in the new base. Nothing additional to rescue beyond what the base already has.

### 3.2 Backend Go — `post_service_test.go` (36 tests)

**Block 1 — first 12 tests (Success/NotFound/basic error for Create/DeletePost/DeleteComment)**: all **(a)**, exist in v1.0.0 renamed. Includes `TestDeletePost_NoEsAutor` → `TestDeletePost_NotTheAuthor` and `TestDeleteComment_NoEsAutor` → `TestDeleteComment_NotTheAuthor` (both mock-only, see section 2 on which of the two exercises real logic).

**Recommendation:** discard as "port as-is" — but the **scenario** of `DeletePost_NotTheAuthor` and `DeleteComment_NotTheAuthor` is worth revisiting in the new repo specifically because of the finding in section 2, not because of a portability gap.

**Block 2 — 24 tests with no equivalent in v1.0.0 (b)**, two subgroups:

- **13 scenario tests (happy path + specific negative)**: `GetAllPosts_Success/Empty`, `GetPostByID_Success/InvalidID/NotFound`, `CreateComment_Success/EmptyContent/PostNotFound/UserNotFound`, `GetCommentsByPostID_Success/PostNotFound/Empty`, `CreatePost_ShouldReturnError_WhenTitleTooShort`. They cover paths that today **do not exist at all** in the new base (`ci-testing` has no tests for these methods).
  **Recommendation:** strong candidates to port/recreate — they add real coverage, not redundant with anything.

- **11 `ShouldPropagateError_When...Fails` tests**: nearly identical, one for each service method, all verifying that an error from the mocked repo propagates as-is.
  **Recommendation:** diminishing marginal value — they're formulaic (same pattern repeated 11 times). Porting all of them is more volume/coverage %; consolidating them into a Go table-driven style covers the same ground with less code. See open question #4.

### 3.3 Frontend Jest — Components (35 tests)

| File | Tests | Status vs. v1.0.0 |
|---|---|---|
| `CommentForm.test.tsx` | 5 | **(b)** whole file doesn't exist in v1.0.0, even though the component is implemented there |
| `CreatePost.test.tsx` | 5 | **(b)** same |
| `PostDetail.test.tsx` | 4 | **(b)** same |
| `CommentList.test.tsx` | 7 | 5 **(a)**, 2 **(b)** (`should show alert when comment deletion fails`, `should show success message after comment is deleted`) |
| `Login.test.tsx` | 7 | 6 **(a)**, 1 **(b)** (`should show fallback error when registration fails without response`) |
| `PostList.test.tsx` | 7 | 6 **(a)**, 1 **(b)** (`should show alert when delete request fails`) |

**Recommendation:**
- The 3 complete files (14 tests) are the highest-value finding to port: `ci-testing` has the components but **zero** coverage of them. Porting/recreating is almost pure upside.
- The 4 specific error/alert-handling tests (`CommentList` ×2, `Login` ×1, `PostList` ×1) are genuine edge cases (what happens if a delete request fails) that the new base doesn't cover — good candidates.
- The rest (24 happy-path tests already present in v1.0.0, renamed to English) → discard as redundant.

### 3.4 Frontend Jest — Services (12 tests)

`authService.test.ts` (5) and `postService.test.ts` (7): all **(a)** — exist in v1.0.0, and in the case of `postService.test.ts` the new base is a **superset** (14 tests there vs. 7 here, adds error-propagation tests that -legacy doesn't have).

**Recommendation:** discard entirely — not just redundant, the new base is already better in this specific file.

### 3.5 Cypress E2E (15 tests)

The 4 files (`auth.cy.js` 5, `comments.cy.js` 4, `full-flow.cy.js` 1, `posts.cy.js` 5) are entirely **(b)** — `ci-testing@v1.0.0` has no Cypress folder, no E2E infrastructure there at all.

**Recommendation:** porting is the only option if E2E is wanted in the new repo (there's no "adapt," it's "build"). Important caveat to carry over: **all 15 are 100% mock-only** (`cy.intercept` in each) — none exercises the real Go backend or real SQLite, despite being categorized as "E2E." See section 2 and open question #3.

### 3.6 Quantitative summary

| Category | Count | Default recommendation |
|---|---|---|
| Redundant with v1.0.0 (discard) | 52 (11 auth + 12 post + 29 frontend happy-path/services) | Don't port |
| Adds real coverage (port/recreate) | 24 backend + 18 frontend + 15 Cypress = 57 | Port, with a caveat on the 11 `ShouldPropagateError` |
| Total | 109 of 109 cataloged with a clear verdict | — |

---

## 4. Conceptual documentation — 9 `desc.md`/`desc/` files

| File | Verdict | Reason |
|---|---|---|
| `backend/internal/database/desc.md` | **Partially rescue** | Explains the 3-table schema and the `ON DELETE CASCADE` cascade (confirmed in code) — not obvious just from a surface reading of the code. Overlaps with ADR-001 on the "why SQLite" part; that part is discarded, the rest is translated/reworded. |
| `backend/internal/models/desc.md` | **Discard** | It's source code copied inside a `.md`, with no original text. Also out of sync: it has comments in Spanish that in the real code have already been translated to English. No documentation value. |
| `backend/internal/repository/desc.md` | **Discard / low value** | The "why interfaces" argument is already covered as well or better in ADR-001. Also outdated: it doesn't mention `DeleteComment()` in the list of operations. |
| `backend/internal/services/desc.md` | **Rescue** | The only place that explicitly documents the business rules (e.g., "only the author can delete their post," marked as a business rule) — no ADR covers this. Needs updating: it also doesn't mention `DeleteComment()`. |
| `backend/tests/desc.md` | **Partially rescue** | The reasoning behind "why mock" (isolation, AAA pattern) is transferable pedagogical content. Discard the counts (already known to be outdated) and the test-naming section, which doesn't reflect the suite's real Spanglish/BDD-English mix. |
| `frontend/src/desc.md` | **Low value, not a priority** | Generic content about "why mock axios," not project-specific. Outdated counts in 3 of the 8 suites it mentions; doesn't even cover the other 5. |
| `frontend/src/services/desc.md` | **Partially rescue** | The service-layer abstraction argument (axios replaceable without touching components) is generic but well-explained reasoning, with no outdated data (no count table). Nice-to-have, not critical. |
| `desc/desc-arquitectura.md` | **Rescue with heavy correction** | The document with the most genuine architectural reasoning (layer diagram, request flow, justification of what's tested at each layer) — but also the one with the most stale data: coverage 86.5%/92.44% (real 97.3%/97.64%), mostly wrong per-component test counts, code snippets that don't literally match the real code (`blog.db` vs. the real `database.db`, different constructor names). If rescued, every numeric fact needs to be verified against the code before porting any sentence. |
| `desc/desc-funcionalidades.md` | **Discard almost all of it** | It's an academic exam guide (`cd ~/IngSW3/...` commands, 28 screenshots to show professors) — the whole format doesn't apply to a portfolio. Anchored to an old project identity (a different GitHub org/repo, a different `sonar.projectKey`), confirming it predates the rename. Two specific nuggets are rescuable: the decision not to fix the test-naming convention flagged by SonarCloud (already in ADR-003, this doc only adds the "professor's recommendation" citation — low value for a portfolio) and the resolution of the "Quality Gate: Coverage on New Code 0.0%" issue by changing "New Code Definition" to 30 days (real SonarCloud troubleshooting, no overlap with any ADR — this could be worth documenting if the same issue is hit again in the new repo). |

**Cross-cutting note**: `repository/desc.md` and `services/desc.md` **both** omit `DeleteComment()` from their respective listings — a consistent signal that this feature was added late in the TP7 cycle and no conceptual doc was updated, in either layer.

---

## 5. ADRs (phase 5)

| ADR | Verdict | Reason |
|---|---|---|
| **ADR-001** (tech-stack) | Reuse with polish of form | Solid, clear rationale, well-justified alternatives. Nothing else to review in detail. |
| **ADR-002** (coverage-strategy) | **Rewrite from scratch** | Confirmed: it's a ~95% literal duplicate of ADR-001 (same context, decision, Go/React/SQLite rationale, same Node/Angular/Postgres alternatives). It has **exactly one real sentence** of coverage strategy, orphaned in the "Consequences" section (`collectCoverageFrom` aligned with `sonar.coverage.exclusions`). As of today **there is no real ADR documenting**: why 70%, why measure only `internal/services/`, why exclude `handlers`/`repository`/`router`/`database`/`models`. This last question (why exclude `repository`) is also the root cause of why the coverage gate can never detect the gap from section 2 — it's worth having the rewritten ADR mention this explicitly as a conscious trade-off, not as an omission. |
| **ADR-003** (sonarcloud) | Reuse with polish of form | Solid. Confirmed the "47 duplicated string literals" fact and the documented decision not to fix test naming — both remain accurate. Nothing else to review in detail. |
| **ADR-004** (cypress) | Reuse with polish of form | Solid, and honest about the mocks-vs-real trade-off ("mocked E2E tests do not validate real backend integration" — stated explicitly, not a hidden new finding). Only nuance: the ADR calls this "E2E" despite being 100% mock — not an error in the ADR, but if the new repo is rewritten it's worth deciding whether the "E2E" name is the right one given what they actually test (see open question #3). |
| **ADR-005** (cicd-pipeline) | **Targeted rewrite, not a full one** | The general structure (7 jobs, parallelization, caching) is correct and well explained. "Problems encountered and resolved" needs to be rewritten (adding Codecov and the `develop`/`master` trigger, neither documented today) along with the description of the `quality-summary` job (it says it "runs only if all previous jobs pass" but the YAML has `if: always()`; it says it "provides a consolidated status" but it only prints fixed text — and that text is **more outdated than the brief said**: not just the "89 tests" total, also "35 tests/86.5%" for backend and "39 tests/92.44%" for frontend, all incorrect). Additional finding: the ASCII pipeline diagram in ADR-005 itself doesn't show `backend-build`/`frontend-build` as boxes, even though the prose and the `needs:` do mention them — a minor internal inconsistency to fix in the rewrite. |

---

## 6. New findings (not in the original brief)

1. **Authentication/authorization model spoofable by design.** There's no JWT or session: `Login` (`auth_service.go:70-99`) only validates credentials and returns the user, without issuing any token. Every subsequent request authenticates via an `X-User-ID` header that the client sets freely (`post_handler.go`), with no cryptographic verification that it corresponds to a real session. Passwords are compared and stored in plaintext (`auth_service.go:57,93`, explicitly commented "in production: hash with bcrypt" — never implemented). This is more serious than the specific `DeleteComment` gap: even adding a real integration test against SQLite, the model itself remains bypassable by anyone who crafts the header by hand. See open question #2.

2. **Dead `.ts` files in the Cypress support folder.** `cypress.config.ts` points `supportFile` to `'cypress/support/e2e.js'` (not the `.ts` one). `commands.ts`/`e2e.ts` are Cypress's default scaffold, never touched; the real code (`login` custom command, `mockBackend`, comments in Spanish) lives only in the `.js` versions, which never go through the TypeScript compiler even though the rest of the frontend project is TS.

3. **`quality-summary` in `ci.yml` has all its numbers hardcoded wrong, not just the total.** The brief already flagged the fixed "89 tests"; confirmed it also prints "35 tests/86.5%" for backend (real: 47/97.3%) and "39 tests/92.44%" for frontend (real: 47/97.64%) — it's a static `echo` that never read real results from any job, in any field.

4. **Codecov isn't documented anywhere.** `codecov/codecov-action@v3` runs in `backend-tests` and `frontend-tests` on every push (`ci.yml` lines 53-58, 91-96) — not mentioned in any of the 5 ADRs or any of the 9 `desc.md` files.

---

## 7. Open questions for Octavio

1. **(Already anticipated in the brief)** The real coverage gap for `DeleteComment` — is a real integration test against SQLite added in the new repo, or is it documented as an accepted scope limitation?

2. **New, and more fundamental**: given that the whole authentication model is spoofable (finding #1 in section 6), does it make sense to fix it in the new repo even though it exceeds the nominal scope of "QA pipeline" (TP7), or is it explicitly documented as a known academic-scope limitation — given that the repo's stated focus is testing/QA, not application security?

3. **Honesty of the testing pyramid**: the 15 Cypress tests are 100% mocked against the backend, never actually exercising it (same as today). Is that approach kept as-is in the new repo (documenting it with the same level of honesty ADR-004 already has), or is at least one real E2E test added against the real backend+SQLite as an integration smoke test, given that none of the 3 suites ever reaches the real database?

4. **Granularity of the 13 `ShouldPropagateError_When...Fails` tests** (backend, section 3.2): are they ported as-is (more volume, higher coverage number) or consolidated into a more idiomatic Go table-driven style (same scenario covered, less code, more maintainable)?

5. **`desc-funcionalidades.md`**: my recommendation is to discard almost the entire document except for the 2 specific SonarCloud nuggets mentioned in section 4 (is there value in rescuing anything else, for example as historical reference that a "university coursework" stage existed before the portfolio?).

6. **Rewritten ADR-002**: should the new version explicitly mention that excluding `repository/` from the coverage gate was what allowed the `DeleteComment` authorization gap to go unnoticed (retrospective honesty), or would you rather the ADR limit itself to justifying the decision going forward without that specific self-criticism?
