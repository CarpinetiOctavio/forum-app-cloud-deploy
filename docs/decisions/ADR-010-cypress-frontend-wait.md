# ADR-010: Waiting for the Frontend in the Cypress Job with `wait-on@9.1.0`

**Date:** 2026-11-11
**Status:** Accepted

## Context

The Cypress job (Stage 4) is inherited from `forum-app-qa-pipeline`
([`ADR-003`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline/blob/main/docs/decisions/ADR-003-cypress.md)),
and two rules forbid modifying it: `docs/rules/constraints.md` (the inherited
quality stages) and `docs/rules/pipeline.md` (their logic and commands; this
change is to commands). It broke without any change here: green on
2026-08-18 (run `32092409720`), then 3 failures out of 3 on a
documentation-only pull request (run `42946771152`). Each time, the action's
own wait failed within milliseconds, well short of its 60-second timeout:
`Socket closed before the connection was established`.

The action is pinned by SHA (v7.4.1), but it runs on the Node.js bundled with
the GitHub-hosted runner, and that changed: runner `2.336.0` → `2.338.2`,
Node `24.18.0` → `24.22.1`. The action waits with `got@11.8.6`, which retries
only on listed error codes such as `ECONNREFUSED`. Reproduced in containers
with the same call: on Node 24.18.0 a refused connection is retried; on
24.22.1, Node also emits `ERR_SOCKET_CLOSED_BEFORE_CONNECTION`, which `got`
does not retry, so the wait ends at once. Pinning the action pins its code,
not the runtime it runs on, and that runtime cannot be pinned on hosted
runners.

## Decision

The frontend is started in its own step and waited for with the tool this
job already uses for the backend:

```yaml
- name: Start the frontend dev server
  working-directory: ./frontend
  run: |
    npm start &
    npx --ignore-scripts --yes wait-on@9.1.0 http-get://localhost:3000 --timeout 60000
```

The action's `start`, `wait-on` and `wait-on-timeout` inputs are removed; its
version is not changed. **Nothing else changes:** not the specs, the mocks,
the coverage thresholds, the install step, the start command, the 60-second
timeout, or the GET request. Only the code that decides "the frontend is
ready" moves.

## Rationale

- **Scope:** criterion (d) of `docs/rules/documentation.md`. Stage 4 is a
  required check; while it fails nothing merges, including Stages 7–9, the
  deploy pipeline this repository exists to show. This restores the
  inherited stage; it does not change what it tests.
- **Why `wait-on`:** its HTTP check treats any request error as "not ready"
  and polls every 250 ms until the timeout, without looking at error codes.
  Verified on Node 24.22.2 against a server that came up after 4 s.
- **Limits, stated so they are not overstated later:** the wait now runs on
  the Node from `setup-node`, which is pinned to major `24` only (24.19.0 in
  August, 24.22.2 now), so it is less exposed, not immune. The action still
  runs on the runner's Node for everything else. This is not independence
  from the runner.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Wait on `127.0.0.1` instead of `localhost` | The failure reproduces without IPv6. |
| Update the action to v7.4.9 (`got@16`, which retries correctly) | Eight releases at once, and the wait stays on the runner's runtime: the same fragility. |
| Self-hosted runner or digest-pinned job container | Persistent runners are discouraged for public repositories; a job container does not change the Node that JavaScript actions run on. |

## Consequences

- This pull request's own CI run is the check that Stage 4 passes again.
- `constraints.md` is not edited: this ADR is the one place recording the
  exception.
- `forum-app-qa-pipeline` has the same latent failure; it is not fixed there,
  since corrections in this series travel forward only.
- No scheduled run is added, so future runner drift will again surface on
  the next push or pull request.
