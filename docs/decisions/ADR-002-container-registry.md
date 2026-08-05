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

## Amendment (2026-08-05): bootstrap packages recreated from scratch, not reused

**Context.** Performing the manual bootstrap push (`cloud-deploy-legacy-transferable-knowledge-results.md`
§2) found that `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-{backend,frontend}`
already existed — public, with a full version history of commit-SHA tags
dating back to 2026-07-01, and linked to `forum-app-cloud-deploy-legacy`
(the archived predecessor, per `ADR-000`). No version in that history had
ever been tagged `latest`; the bootstrap push's own one-time `latest` tag
(sanctioned by §1/§2 for this specific purpose) was the first.

**Alternatives considered:**

| Alternative | Reason not chosen |
|---|---|
| Reuse the existing package: re-link it to this repository ("Connect repository," pointed at `forum-app-cloud-deploy` instead of `-legacy`), keep the prior history, note the mixed provenance | Rejected: mixing tags pushed by two different repositories' pipelines under one package name weakens the "one tag, one commit, one repo" traceability `ADR-003`'s SHA-tagging scheme exists to guarantee. Not a technical defect — both prior and new tags would still resolve correctly — but a provenance-cleanliness problem for a portfolio artifact meant to demonstrate this repository's own pipeline, not a merged trail of two. |
| Delete both packages and redo the bootstrap from scratch | **Chosen.** Produces a package whose entire version history belongs to this repository's own pipeline, with no reconciliation note needed to explain an inherited tag range. |

**Decision:** option 2. Octavio deletes both packages manually via GitHub's
UI — the same standard already applied to SonarCloud project configuration
and branch rulesets in this repository: a destructive, irreversible
account-level action is never performed by this assistant, only proposed
and waited on.

**Consequences.**
- `-legacy`'s ghcr.io version history (2026-07-01 through 2026-08-04) is
  lost from ghcr.io itself once deleted. This is not a real loss of
  information: `-legacy`'s own git history, which is what that history
  traces back to, remains fully intact and accessible in the archived
  `forum-app-cloud-deploy-legacy` repository — ghcr.io was never the
  system of record for it.
- The bootstrap push (`docker buildx build --platform linux/amd64 ...
  --push`, tagged `latest` per §1/§2's sanctioned one-time exception) is
  redone against the empty package names, once Octavio confirms both
  deletions are complete.
- `ADR-002`'s original decision (ghcr.io as the registry) and its original
  Consequences are unaffected — this amendment only corrects how the two
  specific packages this repository uses got created, not the choice of
  registry itself.
