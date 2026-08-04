# Operating rules — Deployment (forum-app-cloud-deploy)

## Purpose and authority
This file states how Render itself is configured and how this repo reasons
about promoting an artifact through it. Read `docs/rules/constraints.md`
first. **This file does not re-decide the deploy mechanism** — decided in
`ADR-007` (deploy hooks trigger, `RENDER_API_KEY` reads status only),
stated as a requirement in `docs/rules/pipeline.md`'s Stage 7/9 section.
Referenced here, not duplicated — the same mistake `cloud-deploy-legacy`
made by describing the same wrong mechanism independently in four files
(its own `ADR-002-hosting.md`, `deployment.md`, `pipeline.md`, and
`constraints.md`) is avoided by having exactly one place state it.

## Services — four, not two, named explicitly to avoid the error already found once
**MUST**: configure four Render services — backend and frontend, separately,
for each of QA and PROD.

**Why this is stated explicitly, not left implicit**:
`cloud-deploy-legacy-audit-results.md` §3/§6 found its own `ADR-002`
stating "two Render services... `forum-app-qa` and `forum-app-prod`," while
its real, working setup needed four — the ADR's own stated names didn't
even match its own `SETUP.md`. Stating the real count and a naming
convention here, before any service is created, is how this repo avoids
writing the same contradiction into its own docs.

**Naming convention** (adapt if a real constraint on Render forces a
different name — report it if so, don't silently rename without saying):
`forum-backend-qa`, `forum-frontend-qa`, `forum-backend-prod`,
`forum-frontend-prod`.

## Environment variables — determined by analysis, not guessed in advance
**MUST NOT** document a specific list of required environment variables
before that list has been produced by actually reading this repo's real
backend and frontend source.

**Why this is a MUST NOT, not just a suggestion to be careful**:
`cloud-deploy-legacy`'s own `deployment.md` did exactly the opposite — it
listed `DATABASE_PATH` as "TBD — pending SQLite persistence decision" and
shipped that placeholder as if it were documentation. A variable list
written before the analysis is not a smaller version of the real thing,
it's a guess wearing the format of a fact — exactly what
`docs/rules/documentation.md`'s "documentation follows implementation"
rule exists to prevent.

**Required process**: analyze the real backend source for every
environment variable it reads (port, database path, any credential),
analyze the real frontend source for how it resolves the backend URL at
runtime (see `docs/rules/docker.md` Rule 1 — this repo's actual mechanism
for CRA's build-time inlining problem), and report the real list — per
service, QA and PROD values distinguished — before configuring anything on
Render or writing anything further in this file.

## Persistence — this repo's own decision, not `-legacy`'s
This repo's application code is `qa-pipeline@v1.0.0`'s, not
`cloud-deploy-legacy`'s — its real database/persistence behavior **MUST**
be verified directly against that inherited code, not assumed to match
`-legacy`'s `ADR-005` (ephemeral SQLite) just because the topic is the
same. `-legacy`'s `ADR-005` is a well-reasoned document worth reading for
*how* to reason about this trade-off (real alternatives, real costs, an
honest scope statement) — it is not evidence of what this repo's own
inherited code actually does. This repo's own persistence ADR gets written
once that's verified, informed by `-legacy`'s reasoning, not copied from
its conclusion.

## Secrets
Per `ADR-007`: five secrets total, not the two-secret (`RENDER_DEPLOY_HOOK_QA`/`_PROD`)
shape `-legacy`'s documentation described, and not the four-`RENDER_SERVICE_ID_*`
shape `-legacy` actually implemented instead — this repo's own real
mechanism needs a different five.

| Secret | Scope | Value | Used for |
|---|---|---|---|
| `RENDER_DEPLOY_HOOK_BACKEND_QA` | `qa` environment | Hook URL, `forum-backend-qa` | Stage 7, backend deploy trigger |
| `RENDER_DEPLOY_HOOK_FRONTEND_QA` | `qa` environment | Hook URL, `forum-frontend-qa` | Stage 7, frontend deploy trigger |
| `RENDER_DEPLOY_HOOK_BACKEND_PROD` | `prod` environment | Hook URL, `forum-backend-prod` | Stage 9, backend deploy trigger |
| `RENDER_DEPLOY_HOOK_FRONTEND_PROD` | `prod` environment | Hook URL, `forum-frontend-prod` | Stage 9, frontend deploy trigger |
| `RENDER_API_KEY` | Repository level | Render account API key | Read-only "Retrieve deploy" polling only (both stages) — **MUST NOT** be used for any write call, per `ADR-007`'s Consequences |

**MUST**: scope the four hook secrets correctly — QA-only under the `qa`
GitHub environment, PROD-only under `prod`, never both; `RENDER_API_KEY`
is repository-level since both stages' status polling need it and Render
offers no narrower-scoped key to split it by environment (`ADR-007`,
finding 2). **MUST NOT**: log a secret's value in any pipeline output, or
write a real value into any docs file "as a working example" —
`cloud-deploy-legacy-transferable-knowledge-results.md` §5 is the
specific, documented cost of that exact mistake happening once already.

## Image promotion — the guarantee this whole repo exists to demonstrate
**MUST**: the image deployed to PROD is the exact image — same digest, same
commit-SHA tag — that was deployed to QA. Render (or whatever the chosen
mechanism is) redeploys an existing tag for PROD; it does not receive a
freshly built one.

**Why this is the one rule this file treats as most load-bearing**: this is
`docs/rules/docker.md` Rule 1's guarantee (build once, promote everywhere)
made concrete at the deployment layer — if PROD ever runs an image QA
didn't run first, the QA gate stopped meaning what it claims to mean,
regardless of how correct every earlier stage was.

**Verification**: the image digest referenced by the QA deploy and the
image digest referenced by the PROD deploy, for the same pipeline run,
MUST be identical — checkable directly against the deploy mechanism's own
logs, not assumed from the fact that both stages "succeeded."

## Rollback — principle now, exact steps once the mechanism is decided
**MUST**: be able to redeploy a previous known-good commit-SHA tag without
rebuilding anything. This is only possible because every image is tagged
with its commit SHA and never overwritten (`docker.md` Rule 5) — rollback
is a consequence of that rule, not a separate feature bolted on.

The exact rollback steps (which dashboard action, which API call) depend on
the mechanism decided in `pipeline.md` — written here once that decision
exists, not guessed at now the way `-legacy`'s rollback section ended up
being the one file that accidentally matched its real mechanism while
three others didn't.

## QA vs. PROD — what's allowed to differ, what's never allowed to differ
**Allowed to differ**: runtime environment variables, resource allocation
(CPU/memory, within whatever the free tier permits), the deploy trigger's
approval requirement (automatic for QA, gated for PROD per
`docs/rules/pipeline.md`).

**MUST NOT differ, ever**: the Docker image (digest, not just tag string),
the application code, the Dockerfile that built it.

## Before configuring or documenting anything in this file
1. Complete the environment-variable analysis above and report it.
2. Verify this repo's real persistence behavior against its own inherited
   code (not `-legacy`'s).
3. Confirm the mechanism decision in `pipeline.md` exists before writing
   the Secrets or Rollback sections with real specifics.
4. Report findings and wait for approval before creating any Render
   service or writing further into this file.