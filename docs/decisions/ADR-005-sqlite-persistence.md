# ADR-005: SQLite Persistence Strategy — Ephemeral, with a Required Configurability Fix

**Date:** 2026-08-04
**Status:** Accepted

## Context

The backend persists to SQLite via `go-sqlite3` (a CGO binding — see
`docs/rules/docker.md` Rule 2's base-image consequence). Render web
services, on every tier including free, run containers with ephemeral
filesystems: any file written inside a running container is lost the
moment a new deploy replaces it, and free-tier services additionally spin
down after 15 minutes of inactivity. A SQLite file living on that
filesystem inherits both failure modes.

**Verified directly against this repository's own inherited code, not
assumed from `cloud-deploy-legacy`'s conclusion on the same topic** — per
`docs/rules/deployment.md`'s own requirement that persistence behavior be
checked against *this repo's* code, not inherited by analogy:

```go
// backend/cmd/api/main.go
db, err := database.InitDB("./database.db")
```

The database path is a **hardcoded string literal**. There is no
`os.Getenv("DATABASE_PATH")` anywhere in `main.go`, or anywhere else in
`backend/`. This is a materially different starting point than
`cloud-deploy-legacy`'s own `ADR-005` assumed: that document reasoned about
*where* to point a configurable `DATABASE_PATH` once the ephemeral-vs-durable
decision was made; this repository's actual inherited code isn't
configurable at all yet. Wiggins' Twelve-Factor App names this precisely as
Factor III (*"Store config in the environment... strictly separated from
code"*) — the current code fails that factor outright, independent of
where the file happens to live on disk.

**This is not a conceptual error to attribute to `qa-pipeline` or
`ci-testing` — it is the correct, natural consequence of neither repo ever
needing this.** `ci-testing` never deploys the application anywhere; its
tests run against mocked repositories or a locally started process.
`qa-pipeline` does run the real backend, but only ever in one context: a
single CI job, on `localhost`, for the duration of the Cypress suite
(`docs/rules/pipeline.md`'s inherited Stage 4) — there was never a second,
independently reachable environment in either repo's lifecycle for a
configurable path to distinguish between. Hardcoding `./database.db` was
the right amount of complexity for that scope, not a shortcut around a
requirement that already existed. The requirement is new here, introduced
by this repository's own scope — two named, independently reachable
environments — not a pre-existing gap this repository discovered. The
practical consequence for how the fix gets made: it has to be done as new,
first configurability for this codebase, not as a repair of something that
regressed — no assumption from `-legacy`'s equivalent code carries over,
and the fallback default must remain the same non-sensitive value
(`./database.db`) local development already relies on, with no credential
or environment-specific detail hardcoded in its place.

## Decision

**Accept ephemeral SQLite storage for both QA and PROD on Render's free
tier** — the same conclusion `cloud-deploy-legacy`'s own `ADR-005` reached,
reasoned independently here against this repository's real constraints, not
copied from that conclusion. **Additionally, and as a precondition for this
decision to be implementable at all**: `main.go`'s hardcoded database path
is replaced with an environment-variable read, with the current hardcoded
value retained as the fallback default:

```go
dbPath := os.Getenv("DATABASE_PATH")
if dbPath == "" {
    dbPath = "./database.db"
}
db, err := database.InitDB(dbPath)
```

## Rationale

**Why the configurability fix is in scope despite `docs/rules/constraints.md`'s
restriction on touching `internal/services/`:** the change above is to
`cmd/api/main.go`, the composition root — not to `internal/services/`, not
to any business rule, not to what the application does. It is exactly the
"environment-specific configuration injected at runtime" requirement
`docs/rules/docker.md` Rule 1 and `docs/rules/constraints.md` already state
as non-negotiable for this repository's own scope — containerizing and
deploying an app that hardcodes its own storage path is not actually
possible to do correctly, since "one image, environment-agnostic, runtime
config only" (this repo's central guarantee) cannot be true of a binary
that ignores its environment by construction. This is not a business-logic
change; it is the minimum wiring necessary for this repository's stated
purpose to be achievable at all, which is precisely the "condition for this
repository's own declared guarantees to be real" exceeds-scope criterion
`docs/rules/documentation.md` and `CLAUDE.md` already define — the same
class of justification `ADR-001` used to accept real authentication as
in-scope.

**Why ephemeral is accepted rather than fixed with durable storage.**
Three real, independently sufficient reasons, not one preference:

1. **Cost.** Render Disk (persistent volume) starts at $0.25/GB/month — a
   real, verifiable figure, not a rounding-error justification — and this
   project operates under a zero-budget constraint.
2. **Scope boundary.** Migrating to a hosted relational database (Supabase,
   Neon, Railway Postgres) would replace `go-sqlite3` with a different
   driver, rewrite every query in `internal/repository/`, and touch
   `database.go`'s schema-creation logic — squarely the business-logic and
   `internal/services/`-adjacent surface `docs/rules/constraints.md`
   excludes from this repository's scope, unlike the `main.go` fix above,
   which touches none of it.
3. **The schema is self-healing, which changes what "ephemeral" actually
   costs functionally.** `database.go`'s `createTables` uses
   `CREATE TABLE IF NOT EXISTS` — the schema recreates itself automatically
   on every container start. The application is fully functional
   immediately after any redeploy; the only real cost is loss of
   previously written rows, not application downtime or a broken
   deployment.

**Why this doesn't contradict `ADR-001`'s acceptance of real authentication
as in-scope.** `ADR-001` accepts a security-relevant fix because leaving it
unaddressed was confirmed exploitable and had no other repo left in the
series to land on. Migrating off SQLite is a durability improvement, not a
security fix, has a clear and correct academic-portfolio-scoped answer
(accept the trade-off, document it, provide the escape hatch), and remains
`qa-pipeline`'s architectural surface to change if ever warranted — the two
decisions use the same exceeds-scope test and reach different, independently
justified conclusions.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Render Disk (persistent volume) | Solves the problem correctly, but is a paid feature ($0.25/GB/month minimum) — incompatible with this project's zero-budget constraint. The escape hatch above (`DATABASE_PATH` now genuinely configurable) means adopting this later, on a paid plan, requires zero code change — only a Render configuration change. |
| Migrate to hosted PostgreSQL | The architecturally correct answer for a real production workload — but requires rewriting `internal/repository/`'s SQL and `go-sqlite3`'s CGO dependency, which is `qa-pipeline`'s owned code surface per `docs/rules/constraints.md`, not this repository's to modify as a side effect of containerization. |
| Seed the database from a fixture file on every startup | Addresses the *appearance* of an empty-state problem, not the actual data-loss problem — a user's real registration or post is still lost on redeploy either way. Doesn't solve what ephemeral storage actually costs. |
| Leave `main.go`'s path hardcoded, accept ephemeral storage as-is | Rejected as internally inconsistent: this repository's own `docs/rules/docker.md` Rule 1 and `constraints.md` already forbid baking environment-specific values into the running artifact. Accepting ephemeral storage without fixing the hardcoded path would mean this repo's own stated rules are violated by the code it ships, silently. |

## Consequences

- `cmd/api/main.go` reads `DATABASE_PATH` from the environment, falling
  back to `./database.db` when unset — every existing local-dev and CI
  workflow that never set this variable continues to work unchanged.
- Data written by users (registrations, posts, comments) does not survive
  a redeploy on either QA or PROD, on the free tier, by design — documented
  here as a deliberate, cost-driven trade-off, not an oversight.
- If a paid Render plan is ever adopted, mounting a Render Disk at the
  configured `DATABASE_PATH` durable-izes storage with no further code
  change — the configurability fix above is the entire prerequisite,
  already satisfied.
- This decision is independently reviewable from `ADR-001`: accepting one
  exceeds-scope change (authentication) does not imply accepting others by
  default — each is justified on its own terms, against the same explicit
  test, not by precedent alone.
