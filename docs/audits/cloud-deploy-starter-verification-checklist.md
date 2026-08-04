# Brief — Final verification of the forum-app-cloud-deploy starter

## Context

Adapted from `forum-app-qa-pipeline`'s own starter-verification checklist
(`docs/audits/starter-verification-checklist.md`), same purpose: confirm the
real state of everything inherited via the template copy from
`forum-app-qa-pipeline@v1.0.0`, before building anything new — item by item,
with verifiable evidence, not a prose summary.

**Structural difference from the qa-pipeline precedent, worth having in
mind while answering this**: `qa-pipeline` had to write its new ADRs from a
blank page, because `ci-testing` had nothing on coverage/SonarCloud/Cypress
to draw from. This repo is different — `forum-app-cloud-deploy-legacy`
already has 8 ADRs, 5 `docs/rules/` files, `SETUP.md`, and `COMMANDS.md`
directly on-topic (containers, registry, hosting, deploy pipeline). Most of
this checklist is about **porting that content forward with the specific
fixes already catalogued** in `cloud-deploy-legacy-audit-results.md`, not
writing from scratch.

**Why the process differs, not just what it produces**: `qa-pipeline-legacy`
was the actual, raw October 2025 academic submission for TP7 — built once,
under exam conditions, with no portfolio intent at the time. That's why its
starter process was mostly rebuild-from-scratch (empty README, borrowed
ADRs to delete wholesale, nothing reusable on its actual topic).
`cloud-deploy-legacy` is different in kind, not just in topic: it was built
in July 2026, for the TP8 final defense, already with some portfolio intent
— which is exactly why it already has 8 ADRs and a full `docs/rules/` set
worth auditing in the first place, rather than nothing to audit at all. The
work this checklist verifies is not "demands less than qa's did" — it's a
different kind of rigor: rescuing real prior decision-making, findings, and
transferable knowledge accurately, rather than generating it new. Getting
this rescue right (not silently dropping something worth keeping, not
silently carrying forward something the audit already found broken) is what
"starting optimally" means for this repo specifically.

For each item: **Done** (with a link or output proving it), **Pending**
(with exactly what's missing), or **N/A** (with the reason). Don't use
"Done" without evidence.

## Checklist to confirm

### Repo infrastructure
1. `staging` branch — does it exist, and does it share an initial commit
   with `main`?
2. Branch ruleset on `main` — created yet, or still open (this is a fresh
   repo, unlike qa-pipeline's checklist this isn't "verify an existing one,"
   it's "does one exist at all")?
3. `go.mod` — still `module forum-app-qa-pipeline` (inherited from the
   template), or renamed to `forum-app-cloud-deploy`? If renamed, were the
   imports in every file that references the module path updated too?
4. Secrets — `RENDER_API_KEY`, `SONAR_TOKEN`, and the 4
   `RENDER_SERVICE_ID_*` do **not** copy via a template generation (GitHub
   never carries secrets across). Have new Render services, a new SonarCloud
   project, and new GitHub secrets been created for this repo specifically,
   or is the pipeline still pointing at nothing?

### Documentation inherited from qa-pipeline (candidates for deletion, not adaptation)
5. `qa-pipeline`'s own `ADR-000` through `ADR-005` (6 files, all about
   `qa-pipeline`'s own decisions — coverage strategy, SonarCloud, Cypress,
   pipeline extension, edit functionality) — deleted from
   `docs/decisions/`, no exceptions? Anything that belongs to a `qa-pipeline`
   or `ci-testing` ADR gets referenced by link from wherever this repo's own
   docs need it — never duplicated, never kept as a retained file. Concretely
   for `ADR-005-edit-functionality.md`: this container serves the exact
   `EditPost`/`EditComment` feature that ADR documents, so the future
   authorization ADR (item 17) should link to
   `github.com/CarpinetiOctavio/forum-app-qa-pipeline/blob/main/docs/decisions/ADR-005-edit-functionality.md`
   directly — the file itself does not get copied or kept here.
6. `qa-pipeline`'s `docs/audits/` (5 files: `qa-pipeline-legacy-audit-brief.md`,
   `-results.md`, `qa-pipeline-legacy-transferable-knowledge-brief.md`,
   `-results.md`, `starter-verification-checklist.md`) — deleted? All 5 are
   about `qa-pipeline`'s own history, not this repo's.
7. `qa-pipeline`'s `CLAUDE.md` — still describes `qa-pipeline`/TP7 scope, or
   rewritten for this repo's TP8 scope?
8. `qa-pipeline`'s `docs/rules/verification.md` and `documentation.md` —
   **not qa-pipeline-specific in substance** (evidence-over-assertion
   discipline, ADR justification requirements). Recommend keeping both,
   with the repo name in each file's own title updated. Confirm: kept, and
   updated?
9. `qa-pipeline`'s `docs/rules/testing-and-quality-gates.md` — this one
   **is** qa-pipeline-specific (coverage gates, SonarCloud thresholds for
   TP7's own test suite). Deleted, in favor of a rule file for this repo's
   own testing reality (100% mocked Cypress, the `internal/repository/`
   coverage exclusion — already documented accurately in the audit, §4.1/§6)?
10. `qa-pipeline`'s `docs/diagrams/`, `docs/screenshots/` — cleared of
    qa-pipeline-specific content, ready for this repo's own?

### New authorship, informed by `cloud-deploy-legacy` (not copied from it)

**Correction, 2026-08-04**: earlier drafts of this section described this as
"porting forward with fixes" — treating `cloud-deploy-legacy`'s files as
artifacts to copy and patch. That was wrong, and inconsistent with how
`qa-pipeline` actually used `qa-pipeline-legacy`: its source code is
obsolete (confirmed throughout `cloud-deploy-legacy-audit-results.md` —
pre-hardening, missing `EditPost`/`EditComment`, wrong deploy mechanism in
three files). Everything below is written fresh for this repo, informed by
what the audit already established about which decisions were genuinely
well-reasoned and which weren't — not by opening the old file and editing
it. The one exception is the audit trail itself (item 22): those documents
are historical record, not implementation, and get copied verbatim — same
as `qa-pipeline`'s `docs/audits/` holding byte-for-byte copies of
`qa-pipeline-legacy`'s own audit docs.
11. `ADR-001` through `ADR-008` equivalents — **written fresh** for this
    repo, each one informed by §3 of the audit (which decisions were
    mandatory-by-assignment vs. voluntary, which alternatives were real vs.
    filler, which verdicts already say "fine as substance" vs. "needs a real
    rewrite") — not by opening `cloud-deploy-legacy`'s files and editing
    them. Concretely:
    - Registry, tagging, CI/CD-tool, persistence, and frontend-runtime-config
      decisions (`-legacy`'s `ADR-001`, `003`, `004`, `005`, `006`) were
      already verdicted "fine as substance" — the new ADR for each can reach
      the same decision and even reuse the same real alternatives-considered
      (they're genuine, per the audit), but write the document itself new,
      verifying anything specific to this repo's own state (service names,
      real numbers) rather than assuming `-legacy`'s specifics still apply.
    - Hosting/deploy-mechanism (`-legacy`'s `ADR-002`) needs the most actual
      new thinking, not just new prose — the audit found its documented
      mechanism (deploy hooks) never matched what was really implemented
      (Render API). This repo's own ADR should settle that mechanism
      decision on its own merits before writing it down, not inherit
      whichever one `-legacy` happened to end up running.
    - The postmortem (`-legacy`'s `ADR-007`) is a different kind of
      document — it recorded incidents specific to that repo's own build
      (the arm64 mismatch, the ghcr.io permission bootstrap, the non-root
      Docker user sequence, the API-key leak). Those are exactly the kind of
      transferable knowledge this repo's own build can use to avoid
      repeating the same hour of debugging — but they belong in this
      checklist's own transferable-knowledge document (already covered:
      `cloud-deploy-legacy-transferable-knowledge-results.md`, item 22),
      not recreated as a new postmortem ADR here. If a new incident happens
      during this repo's own build, it gets its own new postmortem.
12. `docs/rules/{constraints,deployment,docker,pipeline,testing}.md`
    equivalents — written fresh for this repo's own real pipeline and
    real constraints, once those exist (this is downstream of items 11 and
    of the infrastructure work still in progress — Render/SonarCloud/ghcr.io).
    `-legacy`'s versions are useful for knowing what categories of rule this
    repo needs (deploy mechanism, Docker conventions, test-suite scope) and
    which specific claims turned out false against real code (the
    deploy-hook description, the "E2E covers repository layer" claim) — not
    as a base file to edit.
13. `docs/SETUP.md` equivalent — written fresh once the real infrastructure
    (Render services, SonarCloud project, secrets — still in progress) is
    real. `-legacy`'s `SETUP.md` was audited as accurate for what it
    described, so it's a reliable guide to what sections this repo's own
    `SETUP.md` needs, not a file to carry over.
14. `docs/COMMANDS.md` equivalent — same: written fresh, informed by
    `-legacy`'s structure and by the audit's specific finding (real service
    IDs were left exposed in its rollback section, §4.4/§6) — meaning this
    repo's own version should use placeholders from the start, not because
    an old file got redacted, but because it's known from the start not to
    write real infrastructure IDs into a docs file.
15. `README.md` — written fresh. Needs to cover, at minimum, what the audit
    flagged as wrong in `-legacy`'s version (the "Issues Resolved: 47" claim
    — real number is 6, boilerplate origin identified in §7.1/§6) and what's
    different about this repo's own app (this codebase has
    `EditPost`/`EditComment`, inherited from `qa-pipeline@v1.0.0` —
    `-legacy`'s app description predates that feature entirely, since its
    code never had it).

### `cloud-deploy-legacy`'s `desc.md` files — same correction applies
Per the audit (§6): the triage verdicts from `qa-pipeline-legacy`'s own
audit transfer directly onto these 7 files (which of them are worth
understanding vs. discarding) — but "worth understanding" still means
informing new authorship, not transcription. `qa-pipeline`'s own precedent
confirms this: its `database/desc.md`-equivalent content didn't get pasted
into the new `SETUP.md`, it informed what that fresh `SETUP.md` needed to
cover, verified against real code before being written.
16. `database/desc.md`, `frontend/src/services/desc.md` — do they inform
    this repo's own fresh `SETUP.md` (item 13), verified against this
    repo's real code before being written, not transcribed?
17. `services/desc.md` (business rules) — **left aside deliberately**,
    reserved for a future authorization/scope ADR (same pattern as
    qa-pipeline's item 14) — confirm it wasn't integrated anywhere by
    mistake, and that this repo's real authorization split (`DeletePost` in
    the service layer, `DeleteComment` in the repository layer — the
    finding this entire audit chain traces back to) is what that future ADR
    should document, not just the original file's business rules.
18. `tests/desc.md` — referenced via a link to the relevant ADR/rule file,
    not transcribed as new content?
19. `models/desc.md`, `repository/desc.md`, `frontend/src/desc.md` —
    discarded with no action, confirmed not integrated anywhere?
20. `desc-arquitectura.md`, `desc-funcionalidades.md` — **N/A**.
    `cloud-deploy-legacy` has no `desc/` folder at its root and never did —
    confirmed by direct comparison against `qa-pipeline-legacy`, which has
    both files (restored during its own historical revert). Items 17–18 of
    the qa-pipeline checklist have no equivalent here.

### New content specific to this repo
21. `ADR-000` (copy, not fork, not mirror — same mechanism as
    `qa-pipeline`'s own `ADR-000`) — drafted, citing
    `cloud-deploy-legacy-audit-results.md` as its evidence, the way
    `qa-pipeline`'s `ADR-000` cites `qa-pipeline-legacy`'s two audits?
22. This repo's own `docs/audits/` — the two final documents
    (`cloud-deploy-legacy-audit-results.md`,
    `cloud-deploy-legacy-transferable-knowledge-results.md`) and the two
    checklists (`cloud-deploy-audit-checklist.md`,
    `cloud-deploy-live-exposure-checklist.md`) copied in?

### Final confirmation
23. For every branch with unmerged changes: are they actually pushed to the
    remote? Paste the compare link for each (`.../compare/main...branch-name`),
    not just a text confirmation.

## Expected output format

The 23 items, in order, each with its label (Done/Pending/N/A) and its
evidence. If something ended up in an intermediate or ambiguous state, say
so — don't round it up to "Done."

---

## Answers (verified 2026-08-04, against the real repo — commands below each item)

**Baseline**: `git log --all --oneline` → one commit, `7ca7078` ("Initial
commit", 2026-08-03). `staging` and `main` share it, not diverged.

### Repo infrastructure

**1. `staging` branch** — Done.
`git fetch origin` → `origin/staging` exists. `git merge-base main
origin/staging` and `git log -1 --format="%H" main` both return
`7ca7078b3becafc7defc5e5b12e705340348fb93` — same commit, not diverged.

**2. Branch ruleset on `main`** — Done, imported from `qa-pipeline`.
`gh api .../rulesets` → `staging-and-main-protection`, `target: branch`,
`enforcement: active`, created `2026-08-04T00:52:09-03:00`. Required
status checks still list `qa-pipeline`'s 7 own job names (`Test Summary`,
`SonarCloud Code Analysis`, `Cypress E2E`, `Backend Build`,
`Backend Tests (Go)`, `Frontend Build`, `Frontend Tests (React)`) — none of
`cloud-deploy`'s future Docker/deploy jobs yet. Known, deferred update, not
a gap — this repo's own `ci.yml` doesn't exist yet to know the real job
names from.

**3. `go.mod`** — Done.
`backend/go.mod` line 1: `module forum-app-cloud-deploy`. All 12 `.go`
files that imported the old `forum-app-qa-pipeline` path updated to match
(`cmd/api/main.go`, both `tests/mocks/*.go`, both
`tests/services/*_test.go`, `internal/repository/{user,post}_repository.go`,
`internal/handlers/{auth,post}_handler.go`,
`internal/services/{auth,post}_service.go`, `internal/router/router.go`).
`grep -rn "forum-app-qa-pipeline" backend --include="*.go"` → empty.
`go build ./...`, `go vet ./...` both exit 0. `go test
./tests/services/... -cover -coverpkg=./internal/services/...` → 50/50
pass (including `EditPost`/`EditComment` tests, confirming this template
carries `qa-pipeline@v1.0.0`'s full feature set), 88.2% coverage.
**Correction (re-verified 2026-08-04):** this line originally recorded
"47/47" — re-run with a clean test cache and a manual `--- PASS` count,
the real figure at this checkpoint (test files unchanged since) is 50/50.
47 was already wrong when this checklist was written, not a later
regression. Corrected per `docs/rules/verification.md`.

**4. Secrets** — Partial: SonarCloud sub-piece Done, Render sub-piece
deliberately deferred (Octavio's sequencing decision, see below).

SonarCloud:
- New project created, confirmed via public API
  (`sonarcloud.io/api/components/search?organization=carpinetioctavio&q=cloud-deploy`).
  Both this project and the pre-existing legacy one had their keys updated
  in the SonarCloud dashboard (requires an authenticated session this
  environment doesn't have — done by Octavio): legacy
  `CarpinetiOctavio_forum-app-cloud-deploy` →
  `CarpinetiOctavio_forum-app-cloud-deploy-legacy`; this repo's new project
  → the now-freed clean `CarpinetiOctavio_forum-app-cloud-deploy`.
  Re-confirmed via the same API call, both correct.
- `SONAR_TOKEN` secret created and loaded — confirmed by name, not value:
  `gh api .../actions/secrets` → `{"total_count":1,"secrets":[{"name":
  "SONAR_TOKEN","created_at":"2026-08-04T04:30:44Z",...}]}`.
- `sonar-project.properties` updated in **both** repos to match:
  `CarpinetiOctavio_forum-app-cloud-deploy` here, `..._forum-app-cloud-deploy-legacy`
  in `-legacy` — plus 2 references in `-legacy`'s own `README.md` (the
  Quality Gate badge and the plain-text SonarCloud link), which would
  otherwise have kept pointing `-legacy`'s badge at *this* repo's live
  quality gate status.

Render — deliberately deferred: `gh api .../actions/secrets` shows only
`SONAR_TOKEN`; `gh api .../environments` → `{"total_count":0}`, no
`qa`/`prod` environments either. **Sequencing decision (Octavio,
2026-08-04)**: Render/Docker setup depends on Dockerfiles and
Docker-build/deploy stages in `ci.yml` that don't exist yet — real build
work, not starter housekeeping — and belongs to the phase *after* this
starter checklist closes and `-legacy` is archived, not folded into
finishing item 4 now. Item 4 stays **Partial** on that basis.

### Documentation inherited from qa-pipeline

**Correction to this section's own model, 2026-08-04 (Octavio)**: the
first pass at items 5–10 treated "not `cloud-deploy`-specific in content"
as the deletion criterion. Wrong test — the real one is "will this file be
consulted again as a template/reference while writing this repo's own
equivalent document." Each of `qa-pipeline`'s 6 ADRs and 5 diagrams was
read in full under that corrected test before any deletion:

- **`ADR-000` kept temporarily, then deleted (2026-08-04)** once its
  copy-not-fork-not-replace-in-place pattern was absorbed into this repo's
  own `ADR-000-starting-from-qa-pipeline-v1.0.0.md` — see item 21.
- **`ADR-004`, `ADR-005` kept temporarily, also deleted (2026-08-04),
  correction to the correction**: once this session confirmed *exactly*
  when and why each would be consulted again (`ADR-004`'s pipeline shape
  for `pipeline.md`'s Stage 1–4 section; `ADR-005`'s service-layer-not-SQL
  authorization finding for the real-authentication follow-up ADR
  `ADR-001` defers), keeping the files themselves stopped being
  justified — per the cross-repo referencing convention
  (`docs/rules/documentation.md`), once the *when/why* is pinned down
  precisely enough to write as a citation, the file becomes exactly the
  kind of duplication that convention exists to prevent. Both are now
  linked from where they'll actually be needed (`pipeline.md`'s Stage 1–4
  section; `ADR-001`'s Consequences) instead of sitting in this repo's own
  `docs/decisions/`.
- **`ADR-001`, `ADR-002`, `ADR-003` deleted** — `docs/decisions/`:
  `git rm ADR-001-coverage-strategy.md ADR-002-sonarcloud.md
  ADR-003-cypress.md`. None of these have a future "this repo's own
  version" to inform — coverage strategy, SonarCloud config, and Cypress
  mocking strategy are inherited and referenced by link if ever needed,
  never re-decided here.

**5. `qa-pipeline`'s own `ADR-000`–`ADR-005`** — Done, all 6 resolved.
None are copied as files into this repo's own `docs/decisions/` as of
2026-08-04 — the 3 kept temporarily as active reference (`ADR-000`,
`ADR-004`, `ADR-005`) were each deleted once either absorbed into this
repo's own new ADR or pinned down precisely enough to cite by link instead.
Confirmed: `ls docs/decisions/` →
`ADR-000-starting-from-qa-pipeline-v1.0.0.md`,
`ADR-001-real-authentication-in-scope.md` only — both this repo's own.

**6. `qa-pipeline`'s `docs/audits/`** — Partial, same correction. The 4
audit files (`qa-pipeline-legacy-audit-brief.md`, `-results.md`,
`qa-pipeline-legacy-transferable-knowledge-brief.md`, `-results.md`)
deleted — read in full first (this session), confirmed their value was in
demonstrating *how and why* that sweep was conducted (methodology:
separate code-specific findings from tool/platform-behavior findings), not
in their specific qa-pipeline content — that methodology was already
applied fresh to `cloud-deploy-legacy` rather than needing the old files
kept as a live reference. **`starter-verification-checklist.md` (the
`qa-pipeline` precedent to this very document) kept** — still being
consulted directly, mid-use, to complete this checklist's own remaining
items. `ls docs/audits/` → `cloud-deploy-starter-verification-checklist.md`,
`starter-verification-checklist.md` only.

**7. `qa-pipeline`'s `CLAUDE.md`** — Still Pending, not resolved this pass.
Real authorship work (TP8 scope, this repo's own series position) — same
category as items 11–15, deliberately deferred per the sequencing decision
in item 4, not attempted alongside the housekeeping-only items 5/6/9/10.

**8. `qa-pipeline`'s `docs/rules/verification.md`, `documentation.md`** —
Done. Both kept, per the item's own recommendation, and updated —
mechanical in `verification.md` (title only, plus swapping the
illustrative predecessor-repo example from `qa-pipeline-legacy` to
`cloud-deploy-legacy`'s own, freshly confirmed, stale-claims findings);
slightly more than title-only in `documentation.md` (the ADR-justification
topic list and every `TP7`/`TP8` reference updated to this repo's real
scope, the cross-repo-reference line extended to include
`forum-app-qa-pipeline` alongside `forum-app-ci-testing`, and the one
illustrative example replaced — `layered-architecture.svg`, which doesn't
exist in this repo's own history, swapped for the deploy-hook-vs-Render-API
drift actually found in `cloud-deploy-legacy` this session). Both
substitutions used facts already established, not new authorship.
`grep -n "TP7\|qa-pipeline\b"` on both files afterward → only two
remaining mentions, both intentional (one inside the quoted false-claim
example, one the correct cross-repo link).

**9. `qa-pipeline`'s `docs/rules/testing-and-quality-gates.md`** — Done.
`git rm docs/rules/testing-and-quality-gates.md`. Confirmed
qa-pipeline-specific in substance (coverage gates and SonarCloud
thresholds for TP7's own suite) with no future "this repo's own version"
planned — this repo inherits that testing reality unchanged, item 12's
future rule file only needs to describe the *addition* of Docker/deploy,
not re-litigate coverage/SonarCloud scope.

**10. `qa-pipeline`'s `docs/diagrams/`, `docs/screenshots/`** — Done, with
the same nuance as item 5: 4 of 5 diagrams and all 9 screenshots deleted,
1 diagram kept. Each of the 5 diagrams was opened and read as raw SVG
before deciding, not assumed from the filename:
- **`ci-pipeline.svg` kept** — visually documents the same 7-job shape
  `ADR-004` describes in prose; this repo's own future pipeline diagram
  will extend this exact shape (new boxes appended after `summary`) and
  should likely reuse its color/layout conventions for visual consistency
  across the series. `git rm` not run on this file.
- **`cypress-mocked-vs-real.svg`, `layered-architecture.svg`,
  `edit-delete-authorization-matrix.svg`, `silent-due-to-not-exercising.svg`
  deleted** — each illustrates a `qa-pipeline`-specific fact (its own test
  composition, its own app's internal layers, a matrix already fully
  covered in `ADR-005`'s text, three specific incidents from `qa-pipeline`'s
  own CI history) with no future "this repo needs to draw something like
  this" use case identified.
- **All 9 screenshots deleted** — literal images of `qa-pipeline`'s own
  pipeline runs, SonarCloud dashboard, and Edit-feature UI. Not
  reconstructable as a template the way a hand-authored SVG or a prose ADR
  is; this repo will need its own fresh screenshots once its own pipeline
  and app exist.

`git build`/`go vet`/`go test` re-run after all deletions to confirm
nothing in `backend/` was affected (deletions were entirely under `docs/`):
50/50 pass, 88.2% coverage, unchanged from item 3. (Originally recorded as
"47/47" here too — same correction as item 3 above, same re-verified real
figure: 50/50.)

### New content specific to this repo

**21. `ADR-000`** — Done (2026-08-04), ahead of the rest of items 11–23,
at Octavio's explicit request. `docs/decisions/ADR-000-starting-from-qa-pipeline-v1.0.0.md`
written fresh, same mechanism as `qa-pipeline`'s own `ADR-000`
(Context/Decision/Alternatives/Consequences, the same three alternatives
considered — continue in place, fork, replace in place — each with
`cloud-deploy`-specific reasoning, not copied reasoning), citing
`cloud-deploy-legacy-audit-results.md` and
`cloud-deploy-legacy-transferable-knowledge-results.md` as its evidence the
way `qa-pipeline`'s `ADR-000` cites `qa-pipeline-legacy`'s two audits.
Specific content grounded in this session's real findings: the lineage diff
evidence (§2 of the audit), both live-exploit reproductions (Claude's and
Octavio's own, independently), the four-file deploy-hook mismatch, the
`X-User-ID` authentication gap explicitly left unresolved and out of this
repo's scope (per `constraints.md`), and the Render-services-suspended
decision. `qa-pipeline`'s own `ADR-000-starting-from-ci-testing-v1.1.0.md`
deleted immediately after — its pattern is now fully absorbed, nothing left
to reference it for. Two stale cross-references caught and fixed in the
same pass: this checklist's own item 5 note, and confirmed no other file
in this repo references the deleted filename (`README.md`'s mention is
`qa-pipeline`'s own inherited content, out of scope until item 15).

### Addendum, 2026-08-04 — `ADR-001`, outside the original 23 items

Not part of the original checklist, and not starter housekeeping — real
content, written at Octavio's explicit request before archiving `-legacy`.
`docs/decisions/ADR-001-real-authentication-in-scope.md` accepts, as this
repo's own scope, the real-authentication work `qa-pipeline` recorded (only
in a commit message, `8b07d0d` — never in its own committed prose) as
"deferred to cloud-deploy." Found by re-reading `qa-pipeline`'s real git
log and files directly, not the inherited copies in this repo, at
Octavio's specific prompt to look past `docs/rules/`/`CLAUDE.md` (both
early-written) for a later, final decision. Scope accepted; mechanism
(JWT/session/other) explicitly deferred to its own future ADR — not
designed here. `CLAUDE.md` and `docs/rules/constraints.md`, both of which
stated the opposite before this, were corrected in the same pass.
Consequence for `ADR-004`/`ADR-005` (item 5, revised above): once this
ADR pinned down precisely when and why each gets consulted again, keeping
either file locally stopped being justified — both deleted, both linked
from their real point of future use instead.

### Addendum, 2026-08-04 — item 11 (`ADR-001` through `ADR-008` equivalents), substantially done

Never answered in the original pass; real progress since makes leaving it
blank misleading. **7 of this repo's own 8 ADRs exist**, at Octavio's
explicit request, ahead of the original "defer all real authorship until
after archiving `-legacy`" sequencing (§ item 4's own sequencing note) —
he judged ADRs specifically (decisions, not implementation) safe to write
early, distinct from Docker/Render construction work, which stays
deferred:

- `ADR-000` (starting point) — item 21, already closed above.
- `ADR-001` (real authentication, scope only) — this addendum's own
  subject, above.
- `ADR-002`–`ADR-006` (registry, tagging, CI/CD tool, persistence,
  frontend runtime config) — written fresh, each verified against §3 of
  `cloud-deploy-legacy-audit-results.md`'s per-ADR verdict table
  (`docs/rules/documentation.md`'s ADR-justification bar, not `-legacy`'s
  conclusions taken on faith). `ADR-005` and `ADR-006` specifically
  surfaced a real, load-bearing finding neither `-legacy`'s equivalent
  ADRs nor this checklist's original item 11 text anticipated: this
  repo's actual inherited code (`main.go`, `postService.ts`,
  `authService.ts`) has **zero** environment-variable configurability at
  all — not differently configured from `-legacy`'s assumption, not
  configured at all. Both ADRs required a real, scoped source fix as a
  documented precondition, not just a deployment-layer decision layered on
  top of already-adequate code.
- `ADR-007` (Render deploy mechanism — this repo's equivalent of
  `-legacy`'s own hosting `ADR-002`, different number here since this
  repo's own `ADR-002` is a different topic) — the one item 11 flagged as
  needing "the most actual new thinking, not just new prose." Resolved:
  verified directly against Render's current API documentation (not
  `-legacy`'s precedent either way) that deploy hooks and the Render API
  are functionally equivalent for targeting an exact tag — a belief to the
  contrary was formed and then corrected mid-investigation, not assumed
  correct on first read. Decision: deploy hooks trigger, `RENDER_API_KEY`
  restricted to read-only status polling.
- `ADR-008` equivalent (test-language) — **N/A, not just unwritten**: this
  repo's inherited test suite comes from `qa-pipeline`, whose own `ADR-008`
  already settled Spanish test descriptions for its own repo;
  `docs/rules/constraints.md` already states this inherited content isn't
  this repo's to re-decide. No future document needs writing here.

`docs/rules/pipeline.md` and `docs/rules/deployment.md` were both updated
in the same pass to state `ADR-007`'s real mechanism as a requirement,
replacing the two-option "open decision" language both carried before.
`go build`/`go vet`/`go test` re-confirmed clean after this addendum's
work — none of it touched `backend/` beyond the two source fixes `ADR-005`/
`ADR-006` themselves specify are still pending implementation (the ADRs
are written; the actual `main.go`/`postService.ts`/`authService.ts` edits
are not — that's real build work, still correctly deferred).

**What's left of item 11**: nothing to *decide* — implementing what these
7 ADRs already specify (the `main.go`/frontend source fixes, the
Dockerfiles, the actual `ci.yml` stages 5–9) is real build work, the same
category already deferred in item 4's own sequencing note.