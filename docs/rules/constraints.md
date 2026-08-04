# Operating rules — constraints (forum-app-cloud-deploy)

## Purpose and authority
This file states this repo's non-negotiable rules. Every other rules file
(`docker.md`, and any `deployment.md`/`pipeline.md` written once the hosting
and deploy-mechanism decisions exist) must be consistent with this one. If a
rules file appears to permit something this file forbids, this file wins —
stop and flag the conflict, do not resolve it by improvising.

## The TP8 assignment floor — verified against the real text
Every rule below satisfies, at minimum, the graded assignment
(`08-contenedores-automatizacion.md`, `ingsoft3ucc/TPs_2025`): a justified
container registry; QA and PROD container deploys with real environment
differentiation; a CI/CD pipeline with commit-SHA-tagged images (never
`latest` as the sole tag) and a manual gate between environments; scoped
secrets management; and a documented alternative for every decision,
defensible orally. This is the floor this repo must clear, not the ceiling —
see `docs/decisions/ADR-000` for what "more than the floor" means here and
why.

## Image and artifact constraints
- **MUST**: build the image exactly once per pipeline run. The same image,
  identified by the same commit-SHA tag, is what gets deployed to QA and,
  after approval, to PROD — never rebuilt between the two.
    - Why: an image rebuilt per environment means QA validated a different
      artifact than the one PROD runs — the entire point of a QA gate is
      testing the thing that will actually ship (Humble & Farley,
      "build once, promote everywhere").
    - Verification: the image digest referenced by the `deploy-qa` step and
      the image digest referenced by the `deploy-prod` step for the same
      pipeline run MUST be identical.
- **MUST NOT**: tag any image `latest` as its only tag, in any workflow
  step, Dockerfile, or Render service configuration. MUST tag with
  `${{ github.sha }}`.
    - Why: `latest` is not a version — it can't tell you which commit a
      running container corresponds to, which breaks rollback and audit
      trails. This repo's own `ADR-003`-equivalent documents the full
      rationale; this line exists here because it's non-negotiable, not
      optional pending further justification.
- **MUST**: contain, in the final stage of every Dockerfile, only what
  running the application requires — no compiler, no source code, no test
  dependencies, no package-manager cache.
    - Why: attack-surface reduction. A compiler or shell utility present in a
      running container is machinery available to anyone who gains any
      code-execution foothold in it, and it does nothing for the container's
      actual job.
- **MUST**: give `CMD` exactly one responsibility — starting the
  application process. MUST NOT run migrations, seed data, health checks, or
  any other initialization inside `CMD`.
    - Why: conflating "start the app" with "prepare the environment" makes
      failures ambiguous (did the app fail, or did setup fail?) and makes the
      image's behavior depend on execution order that isn't visible from the
      Dockerfile alone.

## Dockerfile constraints
- **MUST**: use a multi-stage build for both the backend and frontend
  Dockerfiles — see `docker.md` for the required stage structure and why.
- **MUST NOT**: bake any environment-specific value into a Dockerfile —
  no URL, credential, database path, or flag that differs between QA and
  PROD. **MUST**: inject every such value at container runtime.
    - Why: this is the mechanism that makes the "one image, both
      environments" rule above actually true, not just declared. See
      `docker.md`'s portable-artifact section for the full grounding
      (Twelve-Factor App, factor III).
- **MUST NOT**: commit any credential, token, or secret value to any file
  tracked by git — Dockerfiles, `ci.yml`, `.env` files, or source code.
  This includes writing a "verified working example" with a real value
  in a docs file. `cloud-deploy-legacy-transferable-knowledge-results.md`
  §5 documents exactly this mistake happening once already, in this repo's
  own predecessor, and what it cost to fix — the constraint exists because
  that specific failure is on record, not hypothetically.

## Pipeline constraints
- **MUST**: run quality gates sequentially and block on failure — an
  earlier stage failing MUST prevent every later stage from running,
  including image build and any deploy.
- **MUST NOT**: build or push a Docker image unless every test and
  static-analysis gate already passed.
- **MUST**: deploy to QA automatically on a passing pipeline run. **MUST**:
  require a human approval, enforced via GitHub environment protection (not
  by pipeline logic alone), before any deploy to PROD.
    - Why "via environment protection, not pipeline logic alone": pipeline
      logic runs inside the same YAML anyone with write access can edit — a
      gate implemented only as an `if:` condition is a gate the same push
      that bypasses it can also remove. GitHub's environment-level
      `required_reviewers` rule is enforced outside the workflow file itself.
- **Coverage gate thresholds are `qa-pipeline`'s own decision, inherited as
  data, not re-decided here.** This repo's test suite is the one it started
  with via the initial template copy from `forum-app-qa-pipeline@v1.0.0` —
  verified real (not the false "inherited from qa-pipeline" claim
  `cloud-deploy-legacy`'s own constraints file made about *itself*; see
  `cloud-deploy-legacy-audit-results.md` §2 for why that claim was false
  there). The threshold itself, and any change to it, is `qa-pipeline`'s
  scope, not this repo's — see `CLAUDE.md`'s scope boundary.

## Secrets and environments
- **MUST**: store every sensitive value as a GitHub Secret — registry
  authentication, any Render credential, any database credential. **MUST
  NOT**: store any of these as a plaintext value anywhere else, including
  in a "for reference" comment in a docs file.
- **MUST**: scope PROD-only secrets to the `prod` GitHub environment and
  QA-only secrets to the `qa` environment. A PROD secret MUST NOT be
  readable from a QA pipeline job.
- **MUST**: configure Render environment variables per service — the QA
  service's variables and the PROD service's variables are never the same
  set, even when some individual values happen to match.

## Existing code — scope boundary, restated with its real justification
- **MUST NOT** modify `internal/services/` business logic, the inherited
  test suite's names/structure/coverage, or the inherited CI quality
  stages (backend test, frontend test, SonarCloud, Cypress) — these are
  `qa-pipeline`'s decisions, inherited via the template copy and verified
  real (see `docs/decisions/ADR-000` and `docs/rules/pipeline.md`'s
  Stages 1–4 section, which link to `qa-pipeline`'s own relevant ADRs
  where this repo's rules build on a decision made there). This is
  `CLAUDE.md`'s scope boundary,
  restated here as a hard constraint, not a new rule. **Exception**: real
  authentication (`ADR-001`) will require changes here once its mechanism
  is designed — the one accepted case where this constraint doesn't hold,
  not a general opening to touch business logic.
- Some inherited test descriptions remain in Spanish, deliberately —
  `qa-pipeline`'s own `ADR-008` decided this, for reasons specific to that
  repo. That is not a leftover to "clean up" here; translating it would be
  modifying inherited content this repo has no standing to change.
- **MUST**: keep the Go module name `forum-app-cloud-deploy` once renamed
  per the starter-verification checklist. MUST NOT revert to
  `forum-app-qa-pipeline` or any earlier name.

## Out of scope for this repo
Kubernetes or any orchestration beyond what Render provides natively;
multiple replicas or autoscaling (Render's free tier doesn't offer them,
and there's no budget constraint requiring them — see `ADR-000`); blue/green
or canary deployment strategies; any new application feature, endpoint, or
business-logic change **other than real authentication**.

**Real authentication is the one accepted exception — see `ADR-001`, not a
contradiction of this section.** `qa-pipeline` recorded, in the commit that
wrote `docs/rules/testing-and-quality-gates.md` (`8b07d0d`), that the
spoofable `X-User-ID`-header model was "deferred to cloud-deploy" — intent
that never reached `qa-pipeline`'s own committed prose. Combined with
`cloud-deploy-legacy-audit-results.md` §7 confirming the design
live-exploitable, this repo accepts the scope in `ADR-001`. Accepting scope
is not the same claim as having designed or shipped the fix: the mechanism
is a separate, future decision, documented in its own ADR once real design
work happens — nothing about this exception licenses touching any other
application feature, endpoint, or business logic.