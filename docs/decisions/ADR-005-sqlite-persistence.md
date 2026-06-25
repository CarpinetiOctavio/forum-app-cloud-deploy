# ADR-005 — SQLite Persistence Strategy: Ephemeral Filesystem

## Status

Accepted

## Date

2025

## Context

The backend uses SQLite via `go-sqlite3` (CGO). The database file is written to the local container filesystem (`./database.db`). Render web services — on both free and paid tiers — run Docker containers with ephemeral filesystems: any data written inside the container is lost when a new deploy starts a new container. On the free tier, services also spin down after 15 minutes of inactivity, which may cause additional data loss between requests.

A durable solution requires either a persistent volume mounted into the container, or migrating the backend to a hosted relational database.

## Decision

Accept ephemeral SQLite for QA and PROD on Render free tier. The database file does not persist between deploys. This is a known and documented limitation for this project.

## Rationale

- **Scope constraint:** migrating from SQLite to a hosted PostgreSQL database (Supabase, Neon, Railway) requires replacing `go-sqlite3` with a PostgreSQL driver, updating all SQL queries, and modifying `database.go`. This is a change to business infrastructure explicitly outside the scope of this project, which extends forum-app-qa-pipeline only with containerization and deployment — not with backend changes.
- **Schema is self-healing:** `database.go` creates all tables using `CREATE TABLE IF NOT EXISTS`. On every container startup, the schema is recreated automatically. The application is functional immediately after each deploy without manual intervention.
- **Demo and academic context:** this project is defended in an oral exam and demonstrated in a controlled setting. Persistent data across deploys is not a requirement for that context.
- **Render Disk is not available on the free tier:** Render's persistent disk feature starts at $0.25/GB/month. It is a paid feature and cannot be used within the zero-budget constraint of this project.
- **Tradeoff is known and documented:** accepting ephemeral storage here is a deliberate architectural decision, not an oversight. It demonstrates understanding of the limitation and the path to resolving it in a real production context.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| Render Disk (persistent volume) | Paid feature — not available on free tier. Would solve the problem at $0.25/GB/month on a paid plan. |
| Migrate to hosted PostgreSQL (Supabase, Neon, Railway) | Requires backend code changes outside the declared scope of this project. All three offer a free tier and would be the correct choice for a real production workload. |
| Render PostgreSQL addon | Same problem as above — requires migrating away from SQLite, which touches `database.go`, `go.mod`, and all repository implementations. Out of scope. |
| Initialize database from a seed file on startup | Does not solve persistence — data written by users is still lost on redeploy. Only reduces the empty-state problem, not the data-loss problem. |

## Consequences

- Data written by users (registrations, posts, comments) is lost on every new deploy.
- The application starts in a clean state after each deploy — all tables exist but are empty.
- This behavior is acceptable for demonstration and academic evaluation purposes.
- For a real production deployment, the correct path is: upgrade to a paid Render plan and mount a Render Disk at the `DATABASE_PATH`, or migrate the backend to a hosted PostgreSQL database and remove the `go-sqlite3` dependency.
- The `DATABASE_PATH` environment variable (pending item in `docs/rules/constraints.md`) should still be wired up in `main.go` so that the path is configurable at runtime — this allows switching to a mounted volume path without rebuilding the image, if a paid plan is adopted in the future.
