# ADR-011: Foreign keys enforced on every SQLite connection — correcting an inherited defect

**Date:** 2026-11-23
**Status:** Accepted

## Context

The schema in `backend/internal/database/database.go` declares three foreign
keys with `ON DELETE CASCADE`: `posts.user_id` and `comments.user_id`
referencing `users`, and `comments.post_id` referencing `posts`. SQLite
enforces foreign keys only on connections that enable them, and they are off
by default
([SQLite, "Enabling Foreign Key Support"](https://www.sqlite.org/foreignkeys.html#fk_enable)).
`InitDB` did not enable them, so no foreign key was checked and deleting a
post left its comments in the database.

The defect comes from `forum-app-ci-testing` and reached this repository
through `forum-app-qa-pipeline`, whose
[`ADR-006`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/blob/main/docs/decisions/ADR-006-sqlite-and-unenforced-foreign-keys.md)
records it, with the choice of SQLite, and does not fix it. `database.go` and
the `InitDB` call are the same here; `cmd/api/main.go` only adds the
`DATABASE_PATH` variable (`ADR-005`). That ADR's last Consequence leaves the
decision to a later repository that inherits the code.

Reproduced against this repository's backend image, built from `main`
(`3b158c0`): through the API, a user, a post and a comment are created and
the post is deleted; the database file, copied out of the container, still
holds the comment (`SELECT COUNT(*) FROM comments` returns `1`, and
`PRAGMA foreign_key_check` reports it).

## Decision

Foreign keys are enabled on every connection, through the connection string:

```go
db, err := sql.Open("sqlite3", filepath+"?_foreign_keys=1")
```

The driver, `github.com/mattn/go-sqlite3` v1.14.32, reads `_foreign_keys`
(or `_fk`) from the connection string and runs `PRAGMA foreign_keys` on each
connection it opens (`sqlite3.go`, the option parsing and the per-connection
setup in `Open`). Without a `file:` prefix, it takes the path up to the `?`.
`DATABASE_PATH` is a file path without a query (`ADR-005`,
`docs/rules/deployment.md`).

## The criterion

A decided exclusion and a discovered error are treated differently. An
exclusion decided for this repository, such as real authentication
(`ADR-001` and its amendment), is declared. A discovered error — a false
statement, or a mechanism that does not do what it states — is corrected
forward, in the repository where it is found, when the fix does not change
what the application does for its users and needs no design. Here the
schema states a cascade that the connection never applies, and the fix is
one line that applies it.

The fix does not change what users see. `CreatePost` checks that the user
exists, and `CreateComment` checks the post and the user, before inserting
(`internal/services/post_service.go`); `GetComments` returns "post not
found" for a deleted post, so the orphaned comments were never shown. The
one change is that deleting a post now deletes its comments.

## The reasons `qa-pipeline` gave for not fixing it

`qa-pipeline`'s `ADR-006` gives three reasons ("Why the defect is not fixed
here"), summarized and answered in order:

- **The series corrects a defect at its origin and propagates the
  correction forward, and the origin, `forum-app-ci-testing`, is closed.** The
  correction goes forward, to the repository that is still open, which is
  what `forum-app-ci-testing`'s
  [`ADR-000`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-000-resolving-forward-not-mirroring-backward.md)
  asks: decisions propagate forward, never backward. Neither closed
  repository is edited.
- **Data integrity is outside TP7's scope.** It is outside TP8's scope as
  well. The correction does not rest on scope: it rests on the criterion
  above, which applies to errors, not to features.
- **No guarantee the repository declares depends on it, so criterion (d)
  does not apply.** The same holds here, and this record does not invoke
  criterion (d) of `docs/rules/documentation.md`.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| `PRAGMA foreign_keys = ON` executed once in `InitDB` | It enables the connection that runs it. `database/sql` keeps a pool and opens more connections as needed, and those start with foreign keys off. |
| Delete a post's comments explicitly in `DeletePost` | It repeats in Go a rule the schema already declares, and changes repository code instead of applying the declaration. |
| Move to PostgreSQL | It needs design and a second running service; `ADR-005` records the cost of a hosted database. |
| Declare the defect as a limitation, as `qa-pipeline` does | It is an error with a one-line fix that needs no design, which the criterion above corrects. |

## Verification

Without touching the inherited test suite, which is outside this
repository's scope (`CLAUDE.md`):

- **Before and after, in the backend image.** The same API sequence run
  against the image built from `main` and against the image with this
  change, then `sqlite3` on the database file copied out of each container:

  | | `main` (`3b158c0`) | This change |
  |---|---|---|
  | `SELECT COUNT(*) FROM comments` after deleting the post | `1` | `0` |
  | `PRAGMA foreign_key_check` | reports the comment | empty |
  | `GET /api/posts/1/comments` after deleting the post | `404` "post not found" | `404` "post not found" |

  The image is where the check runs because it is what gets deployed, and
  because it compiles the cgo driver.
- **The inherited suite is unchanged and passes:** `go test
  ./tests/services/... -coverpkg=./internal/services/...`, 88.2% of
  statements, as before. It mocks the repositories and never opens the
  database.
- **Regression in CI:** the `Cypress E2E` job runs the real backend, and
  `delete-comment-real.cy.ts` creates a real post and comment with foreign
  keys enforced; the smoke tests run after each deploy.

The response through the API is the same before and after, so the change
cannot be observed through the deployed services.

## Errors recognized in this record

- `docs/audits/cloud-deploy-audit-checklist.md` states that "the DB
  schema's cascades are already correct". It confirmed the declaration, not
  that SQLite applies it. The audit is a closed record and is not edited.

## Consequences

- Deleting a post deletes its comments. Deleting a user by hand would
  delete their posts and comments; the application has no operation that
  deletes users.
- If a post is deleted between `CreateComment`'s check that it exists and
  the `INSERT` of the comment, the `INSERT` now fails with a foreign key
  error, instead of leaving an orphaned comment.
- QA and PROD databases are ephemeral (`ADR-005`): each deploy starts with
  an empty database, so there are no existing rows to migrate. Enabling
  foreign keys does not recheck rows already stored.
- `CLAUDE.md`'s scope boundary, which keeps application logic out of
  scope, names this correction as an exception and links here.
