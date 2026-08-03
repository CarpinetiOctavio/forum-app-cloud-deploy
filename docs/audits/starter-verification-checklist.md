# Brief — Final verification of the forum-app-qa-pipeline starter

## Context

Before moving TP7 development (coverage gate, SonarCloud, Cypress) to a new
session, I need to confirm the real state of every piece of the "starter"
(getting what was inherited from `ci-testing@v1.1.0` clean before building
anything new). There was more than one case along the way where a "done"
report didn't match what was actually in the repo (files that stayed local-
only without being pushed, a ruleset reported as empty that was actually
active). That's why this request isn't "tell me the state" in general — it's
a closed checklist, item by item, with verifiable proof for each one (a
GitHub link, or the command you ran and its output), not a prose summary.

## Checklist to confirm

For each item: **Done** (with a link or output proving it), **Pending** (with
exactly what's missing), or **N/A** (with the reason). Don't use "done"
without one of the three labels and its evidence.

### Repo infrastructure
1. `staging` branch — does it exist, and does it share an initial commit with `main`?
2. Branch protection ruleset — what required checks are configured right now?
   (exact name of each one, not just "yes, it's active")
3. `go.mod` — is the module path still `forum-app-ci-testing`, or was it
   already renamed? If renamed, to what, and were the imports updated in the
   files that reference it?
4. `node-version` in `.github/workflows/ci.yml` — is it still `'18'` in the 4
   `setup-node` steps, or is it already on `'24'`? If it changed, in all 4
   steps equally?

### Documentation inherited from ci-testing
5. The 11 original ci-testing ADRs, the 2 diagrams, and the 12 screenshots
   inherited from the template copy — are they actually deleted from the
   working tree? On which branch, committed or only staged?
6. The 4 new ADRs (coverage-strategy, sonarcloud, cypress,
   pipeline-extension) — do they exist as files? On which branch? Do they
   reference ci-testing's ADRs via a GitHub link, or duplicate content?
7. `ADR-000-starting-from-ci-testing-v1.1.0.md` — is it in `docs/decisions/`
   exactly as drafted in the previous chat?
8. `CLAUDE.md` — was it rewritten for this repo's scope (TP7), or does it
   still describe ci-testing's scope (TP6)?
9. `README.md` — it was lost in an earlier incident (`git reset --hard`) and
   had to be redone from scratch. Does a new version exist, or is it still
   empty / untouched since the incident?
10. `docs/rules/` — which files remain? Confirm specifically whether
    `testing.md` (ci-testing's scope) was removed and whether
    `documentation.md` was kept.
11. `docs/SETUP.md` — confirmed it already has the database schema and
    frontend service-layer sections. Is it still on the
    `docs/setup-database-frontend-sections` branch unmerged, or was it
    already merged to `main`?

### Distribution of `-legacy`'s `desc.md` files (per the already-closed triage)
12. `database/desc.md` → `SETUP.md` — confirmed already done, don't re-verify.
13. `frontend/src/services/desc.md` → `SETUP.md` — confirmed already done,
    don't re-verify.
14. `services/desc.md` (business rules) — **this one specifically was NOT
    supposed to be integrated yet**, it stays aside for an
    authorization/scope ADR to be written in the new session. Confirm it
    wasn't integrated into any existing ADR by mistake.
15. `tests/desc.md` — it wasn't supposed to be transcribed as new content,
    only referenced via a link to `ADR-003-mocking-strategy` and
    `ADR-009-comment-language-and-aaa-convention` from ci-testing. Confirm no
    new documentation was created from this file.
16. `models/desc.md`, `repository/desc.md`, `frontend/src/desc.md` — these
    were supposed to be discarded with no action. Confirm they don't show up
    integrated anywhere.
17. `-legacy`'s `desc-arquitectura.md` — it was integrated as prose into one
    of the 4 new ADRs. **Pending confirmation of whether it was also created
    as a diagram in `docs/diagrams/`** — if not, do it now, verifying every
    fact (coverage, database name, constructor names) against this repo's
    real code before putting it in the diagram, not copying it from the
    original.
18. `-legacy`'s `desc-funcionalidades.md` — only two specific fragments had
    been flagged as worth rescuing from this one (real SonarCloud
    troubleshooting on "Quality Gate: Coverage on New Code," and the
    historical note that this originated as a university assignment).
    Confirm whether those two fragments were integrated anywhere, or were
    left unaddressed.

### Final confirmation
19. For every branch with unmerged changes (`docs/setup-database-frontend-sections`,
    `docs/replace-inherited-adrs-and-setup`, and any other that exists): are
    they actually pushed to the remote? Paste the compare link for each
    branch against `main` (`github.com/.../compare/main...branch-name`), not
    just a text confirmation.

## Expected output format

The 19 items, in order, each with its label (Done/Pending/N/A) and its
evidence. If something ended up in an intermediate or ambiguous state, say so
— don't round it up to "Done."

---

## Answers (verified 2026-07-28, against the real repo — commands below each item)

### Repo infrastructure

**1. `staging` branch** — Done.
It exists (`origin/staging`). `git merge-base main origin/staging` → `10ac064`,
which is `main`'s current commit (`10ac064486...`) — meaning `staging` is
currently at the same point as `main`, it hasn't diverged yet. Both share the
same "Initial commit" (`74b9209`) as their furthest common ancestor.

**2. Ruleset — exact required checks** — Done.
`gh api repos/CarpinetiOctavio/forum-app-qa-pipeline/rules/branches/main`
returns 4 active rules in ruleset `19909818`:
- `pull_request` (`required_approving_review_count: 0` — no reviewers required)
- `required_status_checks` → **`"Test Summary"`** (the single required check, exact name)
- `non_fast_forward`
- `deletion` (protection against deleting the branch)

**3. `go.mod`** — Pending.
`head -3 backend/go.mod` → `module forum-app-ci-testing`. Not renamed. The
decision to rename it did get documented, in the paragraph you added to
`ADR-000` ("This is a mechanical rename... corrected as part of setting this
repository up, not deferred") — but the execution (the rename + updating
imports) hasn't happened yet.

**4. `node-version` in `ci.yml`** — Pending, and I'm correcting the item's
premise: there aren't 4 `setup-node` steps, there are **2**
(`grep -n "uses: actions/setup-node"` → lines 57 and 114, inside
`frontend-tests` and `frontend-build`). The other 2 version-pinning steps
that do exist are `setup-go` (lines 19 and 95), not `setup-node`. The 2 real
`node-version` entries are still `'18'` (lines 59 and 116). No Go or Node
step changed version.

### Documentation inherited from ci-testing

**5. The 11 inherited ADRs + 2 diagrams + 12 screenshots** — Pending.
On `origin/main` (what's actually published right now) all 11 are still
there in full — confirmed with `git ls-tree -r origin/main --name-only`.
Locally, on the `docs/replace-inherited-adrs-and-setup` branch, they're
deleted from the working tree and **staged** (`git rm` already run) but
**not committed** — that branch's last commit is still `bf317c2` (the
`SETUP.md` one). That branch, on top of that, was never pushed to origin
(see item 19).

**6. The 4 new ADRs** — Exist as files on disk
(`docs/decisions/ADR-001-coverage-strategy.md`,
`ADR-002-sonarcloud.md`, `ADR-003-cypress.md`,
`ADR-004-pipeline-extension.md`), same state as item 5: staged on
`docs/replace-inherited-adrs-and-setup`, not committed, not pushed. On
references: `ADR-003` and `ADR-004` each link once to a real `ci-testing`
ADR on GitHub (`grep -c "github.com/.../forum-app-ci-testing"` → 1 and 1);
`ADR-001` and `ADR-002` have no links at all because they cover topics
`ci-testing` never had (coverage gate, SonarCloud) — there's nothing there to
duplicate or reference.

**7. `ADR-000-starting-from-ci-testing-v1.1.0.md`** — Yes, it's in
`docs/decisions/` with the content drafted in the previous chat, plus the
paragraph you added afterward about renaming `go.mod`. Same state as items 5
and 6: staged, not committed, not pushed.

**8. `CLAUDE.md`** — Pending. Still untouched: the first line reads
`# CLAUDE.md — forum-app-ci-testing`, describes "This repo is TP6," and its
"Scope boundary" explicitly excludes coverage gates, SonarCloud, and E2E —
exactly what this repo now has to build. It was not rewritten for TP7 at any
point in this session.

**9. `README.md`** — Pending, and in worse shape than "untouched": the file
is **empty** (0 bytes). The diff against `origin/main`
(`git diff origin/main -- README.md`) shows the original 372 lines
(inherited from `ci-testing`, with its title and badges) were deleted but no
new content was written yet. There's no commit to fall back to — as already
flagged in the previous chat, this had to be rebuilt from scratch and still
hasn't started.

**10. `docs/rules/`** — `documentation.md` and `testing.md` remain.
**`testing.md` was NOT removed** — it's still complete and outdated: the
title says `(forum-app-ci-testing)`, it describes TP6's scope ("Unit tests
only, Services layer only... Repository and Handlers layers are deliberately
untested... that's a decision (ADR-002), not an omission"). The `ADR-002` it
cites is the old one (`testing-scope-services-layer`) — once the pending
commit from item 5 lands, `ADR-002` in this repo will be `sonarcloud.md`,
not the one `testing.md` is citing. This is a new finding, not anticipated
in the original checklist.

**11. `docs/SETUP.md`** — Still on `docs/setup-database-frontend-sections`,
**not merged to `main`**. Confirmed: the branch is pushed to origin
(`git branch -r` lists it), but there's no open PR for it
(`gh pr list --head docs/setup-database-frontend-sections` → empty) and it
isn't merged. Compare link in item 19.

### Distribution of `-legacy`'s `desc.md` files

**12. `database/desc.md` → `SETUP.md`** — Done, not re-verified in detail
(as requested), just confirming the file lives on the branch from item 11.

**13. `frontend/src/services/desc.md` → `SETUP.md`** — Done, same note as
item 12.

**14. `services/desc.md` (business rules)** — Done (confirmed it was NOT
integrated). Grepping this file's specific content (`Register()`,
`Login()`, `CreatePost()`, `DeletePost()`, "only the author can...") against
the 4 new ADRs and `SETUP.md`: the only matches are 3 generic uses of the
phrase "business rule(s)" in `ADR-001-coverage-strategy.md`, as an abstract
architectural concept ("coverage over code that contains business logic") —
none of the specific methods or validations the original file describes show
up. The file itself remains intact and untouched in `-legacy`, reserved for
the future authorization/scope ADR.

**15. `tests/desc.md`** — Done (confirmed no new content was created).
Grepping "Arrange.*Act.*Assert" / "AAA pattern" against the 4 new ADRs and
`SETUP.md`: no matches. The only thing that happened to this file was the
discard, already committed and pushed in `-legacy`.

**16. `models/desc.md`, `repository/desc.md`, `frontend/src/desc.md`** —
Done (confirmed they don't show up integrated). Grepping distinctive phrases
from those 3 files against `docs/decisions/*.md` and `docs/SETUP.md`: no
matches. All 3 files remain intact, untouched, in `-legacy`.

**17. `desc-arquitectura.md` — diagram** — Correcting the item's premise: the
prose integration **was not into any of qa-pipeline's 4 new ADRs** — it was
into `-legacy`'s own `ADR-002-coverage-strategy.md`, in the previous session.
The diagram was created (`docs/diagrams/layered-architecture.svg`, verified
fact by fact against the real code before drawing it: 97.3%/97.64% instead
of the old 86.5%/92.44%, `database.db` instead of `blog.db`, real
constructors `NewSQLiteUserRepository`/`NewAuthService`/etc.), but it **also
lives in `-legacy`**, not in this repo. `qa-pipeline` has no architecture
diagram of its own — its `docs/diagrams/` is empty right now (the 2
inherited from `ci-testing` are pending deletion per item 5, and nothing new
was put in their place).

**18. `desc-funcionalidades.md` — the 2 fragments** — Partial.
The "Quality Gate: Coverage on New Code 0.0%" fix → Done, integrated into
`-legacy`'s `ADR-003-sonarcloud.md` (committed and pushed). The historical
note that "this was originally a university assignment" → Pending, not
integrated anywhere; it's still mentioned only inside the audit documents
themselves (`audit-results.md` / `audit.md` in `-legacy`, and their copies in
`qa-pipeline/docs/audits/`), where it had originally been raised as open
question #5 and never resolved — it was never decided on or applied.

### Final confirmation

**19. Branches with unmerged changes** — Partial, one real link and one pending:
- `docs/setup-database-frontend-sections` — pushed. Compare:
  https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/compare/main...docs/setup-database-frontend-sections
- `docs/replace-inherited-adrs-and-setup` — **NOT pushed**, exists only in
  this local working tree (`git branch -r` doesn't list it). No compare link
  is possible until it's pushed — the earlier push attempt was interrupted
  at your request so you could commit it yourself.
