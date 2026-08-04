# ADR-004: CI/CD Tool — GitHub Actions

**Date:** 2026-08-04
**Status:** Accepted

## Context

This repository's pipeline does not start from zero: it starts by
inheriting `forum-app-qa-pipeline@v1.0.0`'s already-working, already-tested
CI/CD pipeline (backend/frontend tests, coverage gates, SonarCloud, Cypress
— `docs/rules/pipeline.md` Stages 1–4) and extending it with Docker build,
registry push, and a two-environment deploy with a manual gate (Stages 5–9).
The CI/CD tool decision is therefore not "which tool should a new project
use" in the abstract — it is "does anything justify discarding a working,
already-validated pipeline implementation to re-platform it," which is a
meaningfully different and higher bar.

## Decision

Continue using **GitHub Actions**.

## Rationale

**Continuity has a real cost attached to reversing it, not just an
inertia argument.** Stages 1–4 are not a stub — they are a working
implementation with real, encoded lessons: `sonar.tests`/`sonar.sources`
scoping bugs found and fixed across multiple iterations, an `npm ci`
lockfile-drift failure mode discovered and resolved, a `wait-on`
scheme-vs-method (`HEAD` vs `GET`) bug found and fixed, an `S6505`
unpinned-`npx`-script supply-chain finding remediated. Every one of those
is tool-specific knowledge (GitHub Actions' exact caching semantics, its
artifact-passing model between jobs, its exact interaction with SonarCloud's
official scan action) that a different CI/CD platform would not simply
carry over — it would need to be rediscovered against the new platform's
own equivalent mechanisms, the same category of wasted rediscovery this
project's own audits (`cloud-deploy-legacy-transferable-knowledge-results.md`)
exist specifically to avoid.

**Environment protection rules are a first-class platform feature this
project's design depends on, not a workaround.** The manual approval gate
before PROD (`docs/rules/pipeline.md` Stage 8) is implemented via GitHub's
own environment protection rules — enforced outside the workflow YAML
itself, by the platform. `docs/rules/constraints.md` and
`docs/rules/pipeline.md` both ground this explicitly: a gate implemented
only as an `if:` condition inside `ci.yml` is editable by the same push
that would bypass it; a platform-level rule is not. Reproducing this
property on a CI/CD tool without native environment-scoped, protected
deploy gates would mean building a weaker approximation of a guarantee
GitHub Actions already provides for free.

**Native registry authentication compounds with `ADR-002`.** `GITHUB_TOKEN`
authenticating to ghcr.io with no additional secret is a GitHub
Actions–specific integration; choosing a different CI/CD tool would either
forfeit that property or require re-deriving an equivalent for a different
registry, undermining the reasoning `ADR-002` already established.

**No cost pressure pushes toward an alternative.** GitHub Actions is free
for public repositories with no minutes cap that matters at this project's
scale — there is no budget argument for switching, only potential setup
cost with no corresponding benefit.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| GitLab CI/CD | Requires migrating the repository itself off GitHub — abandoning native `GITHUB_TOKEN` registry auth (`ADR-002`) and GitHub's environment protection rules (this repo's approval-gate mechanism) simultaneously, for no problem this project actually has. |
| CircleCI | Separate account and external webhook configuration; free tier caps at 6,000 minutes/month with no corresponding benefit over GitHub Actions' native integration; introduces a second platform's credential and trust surface for no functional gain. |
| Jenkins | Self-hosted — requires provisioning, securing, and maintaining infrastructure that GitHub Actions provides managed and free. Disproportionate operational burden for this project's scope and lifecycle. |
| Azure DevOps Pipelines | No existing Azure footprint elsewhere in this project to amortize the setup cost against; free tier for public projects is adequate but the migration cost from a working GitHub Actions pipeline isn't justified by any specific capability Azure DevOps offers that this project needs. |

## Consequences

- All pipeline configuration continues to live in `.github/workflows/ci.yml`,
  extended in place, not replaced.
- Stages 1–4's existing logic, commands, and coverage thresholds are
  inherited unmodified — this is `qa-pipeline`'s scope, restated as a hard
  constraint in `docs/rules/constraints.md`, not reopened by this decision.
- The two GitHub environments (`qa`, `prod`) referenced throughout
  `docs/rules/pipeline.md` and `docs/rules/deployment.md` are a direct
  consequence of this tool choice — they do not exist as a concept
  independent of GitHub Actions' own environment feature.
- Every tool-specific gotcha already catalogued in
  `cloud-deploy-legacy-transferable-knowledge-results.md` (the two-part
  `GITHUB_TOKEN`/ghcr.io permission requirement, the arm64/amd64 mismatch on
  manual pushes, the non-root Docker user sequence) is real operational
  knowledge this decision preserves the relevance of — switching tools would
  have made each of those items moot in one direction and a fresh unknown in
  another.
