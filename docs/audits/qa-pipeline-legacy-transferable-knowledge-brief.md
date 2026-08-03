# Brief — Transferable knowledge sweep in forum-app-qa-pipeline-legacy, before the historical revert

## Context

`forum-app-qa-pipeline-legacy` is going to be reverted to its real historical
state (prior to the polish session that translated and enriched its ADRs) —
the goal is for it to end up as a faithful snapshot of how the repo looked at
the correct point in its history, not a hybrid with today's reasoning mixed
in. Once the revert is done, anything that wasn't extracted beforehand is
lost — the same problem that already happened with `qa-pipeline`'s README,
which had no prior commit to fall back to.

Before that revert, two types of content need to be separated out of what
got enriched during this session:
1. **Specific to `-legacy`'s code**: empirical findings about THAT particular
   code (e.g. "we found 47 duplicated strings in this repo's handlers").
   This isn't transferable as-is — `qa-pipeline`'s code isn't identical, so
   the specific number doesn't apply.
2. **Behavior of the tools/platforms themselves** (SonarCloud, Cypress, npm),
   discovered while building the real implementation in `-legacy`. This IS
   transferable — it's going to happen to `qa-pipeline` the same way once it
   reaches that same stage of TP7, because it doesn't depend on the code, it
   depends on how the tool behaves.

## Already found (don't re-derive, but do verify)

In `docs/decisions/ADR-003-sonarcloud.md` and `ADR-005-cicd-pipeline.md` of
`-legacy`:

1. `sonar.language` doesn't support multiple values in SonarCloud — it has to
   be omitted, auto-detection already handles Go and TypeScript without it
   being declared.
2. Automatic Analysis conflicts with CI-triggered analysis — the automatic
   one has to be disabled from the project's configuration in SonarCloud.
3. Quality Gate failing on "Coverage on New Code: 0.0%" after commits that
   only extract constants (with no corresponding new test) — resolved by
   changing the project's "New Code Definition" from "Previous Version" to
   "Number of days: 30" in SonarCloud's configuration (a project-settings
   change, not something that lives in `sonar-project.properties`).
4. Race condition in Cypress when starting the frontend with `npm start &` —
   resolved by using `cypress-io/github-action@v6`, which manages the
   process lifecycle via `wait-on` instead of a manual `&`.
5. `package-lock.json` desync with `npm ci` — resolved by always using
   `npm install` (not installing individual packages) and committing the
   updated lock file.

## What's left to review and that I didn't get to cover

1. **`docs/decisions/ADR-004-cypress.md` from `-legacy`** — haven't read it
   yet. Given that `ADR-005-cicd-pipeline.md` already had a Cypress gotcha
   (the race-condition one), it's reasonable that this ADR, specifically
   dedicated to Cypress, has more — look for any other behavior of the tool
   itself (configuration, timeouts, selectors, whatever) that doesn't depend
   on this repo's specific code.
2. **`docs/decisions/ADR-001-tech-stack.md` from `-legacy`** — already
   decided that it doesn't get ported as an ADR (`qa-pipeline`'s stack
   references `ci-testing`, not this file), but check whether it has any
   loose operational fact unrelated to the stack decision itself that's
   worth rescuing the same way as the 5 points above.
3. **Real comparison between `-legacy`'s `ADR-002-coverage-strategy.md`**
   (the one rewritten this session) **and `qa-pipeline`'s
   `ADR-001-coverage-strategy.md`** (the one that already exists there,
   prescriptive, with no empirical results yet) — point out concretely what
   one has that the other doesn't, not just whether "they look similar."
4. **The `docs/diagrams/layered-architecture.svg` diagram** — beyond moving
   it as-is to `qa-pipeline/docs/diagrams/` (already decided), check whether
   it has any note or fact in the SVG itself (comments, text) that isn't
   already covered in the ADRs and that would be lost if the image is just
   copied without reading it first.
5. **Anything else that comes up** following the same criterion (is it about
   this repo's code, or about how a tool `qa-pipeline` is going to use the
   same way behaves?) that isn't on this list — this brief is a floor, not a
   ceiling.

## Expected output format

A list of findings, each classified as transferable (with the evidence and
why it doesn't depend on `-legacy`'s specific code) or not transferable (and
why). Don't apply anything yet — this is input for the handoff document to
`qa-pipeline`'s new chat, not a change to any repo.
