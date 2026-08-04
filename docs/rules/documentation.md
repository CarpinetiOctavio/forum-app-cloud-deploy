# Operating rules — documentation (forum-app-cloud-deploy)

## ADR requirements
Every ADR must include: the problem/question, the alternatives actually considered,
the decision, and the justification. The justification must trace to one of: (a) a
concept from software engineering literature/practice relevant to this repo's scope
(containerization, container registry, CI/CD deployment, cloud hosting), (b) an
explicit scope boundary of TP8, (c) verified evidence from this repo's own history,
or (d) a condition for this repo's own declared guarantees to be real, even where it
exceeds TP8's literal scope — see `ADR-000` for the criterion and its limits.
"Because the other repo does it this way" is not a valid justification on its own.

## Cross-repo references
Where a decision here builds on something already resolved in an earlier repo of
the series (`forum-app-ci-testing`, `forum-app-qa-pipeline`), link to that repo's
own ADR — do not duplicate its text. `docs/decisions/` in this repo holds only
decisions made in this repo.

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
this repo — not the intended TP8 scope in advance. The README's description of
this repo's *scope and place in the series* is the exception: stating intent
("this repo will add a container registry, Docker builds, and deploy stages")
is not the same claim as stating a result ("the pipeline deploys to PROD via
the Render API"), and only the latter needs to wait. A diagram, screenshot, or
rules file describing a mechanism, config, or pipeline state from a different
repo, an earlier plan, or a different point in time doesn't satisfy this
either, even if it was once accurate or was the original intent (see
`forum-app-cloud-deploy-legacy`'s own case: its `ADR-002-hosting.md` and
three separate `docs/rules/` files kept describing a Render deploy-hook mechanism as the plan
long after the real, implemented mechanism had become the Render REST API —
only `SETUP.md` and `COMMANDS.md` were updated to match, and the other four
never were, independently wrong in the same direction). It has to be true of
*this* repo *now*. When a TP8 feature is implemented, its documentation is
added in the same commit or PR — never staged ahead of the code as a
placeholder, and never left undocumented or unreconciled after the real
implementation changes what it describes.

## Language
English for all prose. Test names in English, inherited from
[`forum-app-ci-testing`'s `ADR-006`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-006-test-name-translation.md) —
not a decision made in this repo, so not re-justified here.