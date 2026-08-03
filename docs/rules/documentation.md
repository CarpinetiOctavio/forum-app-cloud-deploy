# Operating rules — documentation (forum-app-qa-pipeline)

## ADR requirements
Every ADR must include: the problem/question, the alternatives actually considered,
the decision, and the justification. The justification must trace to one of: (a) a
concept from software engineering literature/practice relevant to this repo's scope
(coverage measurement, static analysis, E2E testing, CI/CD quality gates), (b) an
explicit scope boundary of TP7, (c) verified evidence from this repo's own history,
or (d) a condition for this repo's own declared guarantees to be real, even where it
exceeds TP7's literal scope — see `ADR-000` for the criterion and its limits.
"Because the other repo does it this way" is not a valid justification on its own.

## Cross-repo references
Where a decision here builds on something already resolved in an earlier repo of
the series (`forum-app-ci-testing`), link to that repo's own ADR — do not duplicate
its text. `docs/decisions/` in this repo holds only decisions made in this repo.

## Before writing an ADR
Confirm with Octavio any fact that can't be verified from code or git history directly
(why an incident happened, whether a choice was deliberate or a shortcut). Frame
inferred causes as "most probable, given available evidence," never as certainty.

## README
Explains the why of the repo's existence and its place in the series, referencing
ADRs for detail instead of repeating it.

## Documentation follows implementation, not the other way around
`SETUP.md`, `COMMANDS.md`, `docs/screenshots/`, `docs/diagrams/`, and the README's
usage sections describe only tools, commands, and results that already exist in
this repo — not the intended TP7 scope in advance. The README's description of
this repo's *scope and place in the series* is the exception: stating intent
("this repo will add coverage gates, SonarCloud, and Cypress") is not the same
claim as stating a result ("coverage is at 97%"), and only the latter needs to
wait. A diagram or screenshot showing metrics, test counts, or a pipeline state
from a different repo or a different point in time doesn't satisfy this either,
even if the underlying facts were once real somewhere — it has to be true of
*this* repo *now* (see the `layered-architecture.svg` case: accurate for
`-legacy`'s June 2026 state, not for this repo's actual test counts, so it
wasn't ported over as-is). When a TP7 feature is implemented, its documentation
is added in the same commit or PR — never staged ahead of the code as a
placeholder, and never left undocumented after landing.

## Language
English for all prose. Test names in English, inherited from
[`forum-app-ci-testing`'s `ADR-006`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-006-test-name-translation.md) —
not a decision made in this repo, so not re-justified here.