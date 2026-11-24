# Decisions

Each Architecture Decision Record (ADR) states one decision: its context, the alternatives considered and the reasons for the choice. This index only says what each record decides; the reasoning lives in the record itself. Later corrections are added as dated amendments at the end of the record; the original text is not rewritten. Decisions this repository builds on from [`forum-app-ci-testing`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/tree/main/docs/decisions) and [`forum-app-qa-pipeline`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/tree/main/docs/decisions) are linked from the record that uses them, not copied here.

| ADR | Area | What it decides |
|---|---|---|
| [000](ADR-000-starting-from-qa-pipeline-v1.0.0.md) | Method | The repository starts as a copy of `forum-app-qa-pipeline@v1.0.0`, not as a continuation of the original TP8 build (`forum-app-cloud-deploy-legacy`). |
| [001](ADR-001-real-authentication-in-scope.md) | Scope | Accepted real authentication as in scope; reversed by its amendment of 2026-11-11: authentication is out of scope, and the spoofable design is declared as a limitation. |
| [002](ADR-002-container-registry.md) | Registry | GitHub Container Registry, pushed to with the run's own `GITHUB_TOKEN`. Amended 2026-08-05 (the packages were recreated, not reused from the predecessor). |
| [003](ADR-003-image-tagging.md) | Image tags | Every image is tagged with its commit SHA, never `latest` alone. Amended 2026-11-11 (deploys reference the digest) and 2026-11-23 (what the tag guarantees and what it does not). |
| [004](ADR-004-cicd-tool.md) | CI/CD tool | GitHub Actions, extending the inherited pipeline instead of replacing it. |
| [005](ADR-005-sqlite-persistence.md) | Persistence | Ephemeral SQLite on Render's free tier, with the database path read from `DATABASE_PATH`. |
| [006](ADR-006-frontend-runtime-config.md) | Frontend configuration | One frontend image for QA and PROD: the backend URL is substituted into the built bundle when the container starts. Amended 2026-08-04 (the substitution runs on a copy in `/tmp`) and 2026-11-23 (it runs at each container start, not per image). |
| [007](ADR-007-render-deploy-mechanism.md) | Deploy | Render deploy hooks trigger every deploy; the Render API key is used only for read-only calls. Amended 2026-11-11 (deploys pass the image digest), 2026-11-23 (rollback through the same mechanism, and its limits) and 2026-11-24 (a deploy waiting for PROD approval holds the concurrency group). |
| [008](ADR-008-explicit-port-binding.md) | Ports | Both containers listen on the port in `PORT`, instead of relying on Render's detection. |
| [009](ADR-009-single-build-trigger.md) | Pipeline trigger | QA and PROD deploy from the same run on `main`, so both receive the image that run built. Amended 2026-11-23 (the quality gate on the path to `main`, and the rollback claim). |
| [010](ADR-010-cypress-frontend-wait.md) | E2E testing | The Cypress job starts the frontend in its own step and waits for it with `wait-on`. Amended 2026-11-12 (validation run) and 2026-11-23 (background processes, and the fix in `qa-pipeline`). |
| [011](ADR-011-enforce-sqlite-foreign-keys.md) | Database | SQLite foreign keys are enforced on every connection, correcting a defect inherited from the earlier repositories. |

No record includes a diagram.

On 2026-11-23, ADR-000, 001, 003 and 005 to 009 were condensed: passages that repeated other sections, the rules files or other records were removed, and no statement was changed. The previous text is in the git history, one commit per record.
