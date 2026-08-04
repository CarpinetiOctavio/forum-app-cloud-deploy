# ADR-002: Container Registry — GitHub Container Registry (ghcr.io)

**Date:** 2026-08-04
**Status:** Accepted

## Context

TP8's assignment requires a container registry with justified authentication
— a place to store and distribute the Docker images this pipeline builds,
reachable from two independent systems that must never share a long-lived
credential unnecessarily: GitHub Actions (which builds and pushes) and
Render (which pulls and runs). The registry sits on the trust boundary
between "code that compiled" and "artifact that runs in production" — per
the SLSA supply-chain framework's build-integrity model, that boundary is
exactly where a credential compromise has the most leverage: a registry
push credential with excessive scope doesn't just leak data, it lets an
attacker substitute what QA validated for something PROD later runs. The
registry choice is therefore evaluated primarily on credential-scope and
authentication surface, not on storage cost or UI convenience.

**Table stakes, stated plainly rather than left implicit:** this decision
converges with the TP8 assignment's own "100%-free" reference architecture
("Ejemplo 1" of `08-contenedores-automatizacion.md`), which names ghcr.io
directly. That coincidence does not substitute for this ADR's own
reasoning — the alternatives below are evaluated on their real technical
merits for this project, and would lead to the same conclusion even absent
the assignment's suggestion, for the reason in the next paragraph.

## Decision

Use **GitHub Container Registry (ghcr.io)**.

## Rationale

**Native, ephemeral authentication — the deciding factor.** GitHub Actions
authenticates to ghcr.io with `GITHUB_TOKEN`, a credential that is
auto-generated per workflow run, scoped to that run, and expires when the
job ends. No additional registry credential needs to exist as a long-lived
GitHub Secret for the push step. This isn't a convenience — every
long-lived credential a pipeline holds is an asset an attacker can target
independently of any code vulnerability; removing one from the design
removes an entire class of exposure, not just a setup step. This is the
same reasoning `docs/rules/pipeline.md`'s Stage 6 requirement
(`permissions: packages: write` at the workflow level) and this repo's own
`ADR-001` (accepting real authentication as in-scope) are grounded in: a
credential's blast radius, not just whether a task technically works.

**Same trust domain as the pipeline that uses it.** The repository, the
CI/CD pipeline, and the registry all live under one GitHub identity and one
permissions model. A cross-platform registry (Docker Hub, GitLab, a cloud
provider's own) introduces a second authentication domain and a second
place secrets can leak from, for no functional gain this project needs.

**Free, with no functional restriction that matters here.** ghcr.io permits
unlimited public image storage at no cost — satisfying the project's
zero-budget constraint without a storage-tier decision to revisit later.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Docker Hub | Requires a separate, long-lived credential as a GitHub Secret purely for registry auth — reintroducing the exact credential-surface problem `GITHUB_TOKEN` avoids. Free-tier pull-rate limits are also a real operational risk for a QA/PROD deploy pipeline that pulls on every promotion. |
| GitLab Container Registry | Would require migrating the repository itself to GitLab — abandoning the native `GITHUB_TOKEN` integration and the environment-protection-rule mechanism this project's approval gate depends on (see `docs/rules/pipeline.md`). No technical problem this project has is solved by that migration. |
| Amazon ECR | Requires a full AWS account, IAM role/policy configuration, and a distinct credential-management surface (IAM access keys or OIDC federation) — real setup cost and a second cloud provider's trust boundary, for a project with no other AWS dependency. |
| Azure Container Registry | Same class of cost and complexity as ECR, with no existing Azure footprint in this project to amortize it against. |

## Consequences

- Images are pushed as `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-{backend,frontend}`.
- `GITHUB_TOKEN` is sufficient for the pipeline's own push step; no registry
  secret needs to be created or rotated for CI.
- Render, a separate system, still needs read access to a public ghcr.io
  image — for a public package this requires no credential at all; if any
  package is ever made private, a scoped `read:packages` PAT is the
  minimum-privilege credential for that one purpose, not the push-capable
  token used by CI.
- Per `docs/decisions/ADR-001`'s own precedent and
  `cloud-deploy-legacy-transferable-knowledge-results.md` §2, the first
  push to a not-yet-existing package must be manual (bootstrap), and both
  packages must be linked to this repository afterward for `GITHUB_TOKEN`
  to have write access on subsequent runs — this is an implementation
  detail of *using* this decision, not a condition of making it.
