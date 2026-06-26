# ADR-006 — Frontend Runtime Configuration: Entrypoint Placeholder Replacement

## Status

Accepted

## Date

2026

## Context

The frontend is a React application built with Create React App (CRA). It needs to call a backend API whose URL differs between environments: QA uses one Render service URL and PROD uses another.

CRA expands `process.env.REACT_APP_*` variables at build time using webpack's DefinePlugin. The values are inlined as string literals into the compiled JS bundle. There is no mechanism in the CRA runtime to read environment variables from the host — the bundle is static once built.

The pipeline non-negotiable constraint (see `docs/rules/constraints.md`) requires that the same Docker image artifact is deployed to both QA and PROD. Building a separate image per environment — each with a different `REACT_APP_API_URL` baked in — would violate this constraint.

## Decision

Build the frontend image once with a fixed placeholder string (`__REACT_APP_API_URL__`) embedded in the bundle. At container startup, an entrypoint shell script replaces that placeholder in all `.js` files under the nginx document root with the value of the `REACT_APP_API_URL` environment variable injected by Render at runtime. nginx then starts normally.

The `ARG REACT_APP_API_URL=__REACT_APP_API_URL__` instruction in the Dockerfile causes CRA to embed the literal placeholder string instead of `undefined`, ensuring the `||` fallback in the source code does not activate and the placeholder is always present in the bundle for the entrypoint to replace.

## Rationale

- **Preserves the single-artifact constraint:** the Docker image is built once and deployed identically to QA and PROD. The only difference between environments is the value of a runtime environment variable configured in Render.
- **No changes to React source code:** `authService.ts` and `postService.ts` continue to use `process.env.REACT_APP_API_URL` — the entrypoint approach is transparent to the application code.
- **Standard pattern:** runtime placeholder replacement via entrypoint script is the established industry approach for deploying SPAs in containers when environment-specific configuration is required.
- **Reversible:** if the frontend is migrated to a framework with native runtime env support (Next.js, Vite with SSR), this entrypoint can be removed without touching the application logic.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| Build separate images per environment (Option C) | Violates the non-negotiable constraint that the same artifact is deployed to QA and PROD. Rejected without further consideration. |
| nginx reverse proxy with relative API paths (Option B) | Requires changing all API call URLs in the frontend source to relative paths and adding an nginx `proxy_pass` directive for `/api/`. Wider code change for the same outcome. |
| `window._env_` pattern via `public/env.js` | Requires modifying `public/index.html` to load the config script, adding a `window._env_` global, and updating all service files to read from `window._env_`. More moving parts than the placeholder approach for equivalent result. |
| Accept hardcoded localhost fallback | The app deploys but cannot reach the backend in any cloud environment. Functionally broken in QA and PROD. Not acceptable. |

## Consequences

- The frontend Dockerfile gains one `ARG` instruction in the builder stage and an `ENTRYPOINT` script in the final stage, replacing the previous `CMD`.
- `REACT_APP_API_URL` must be set in each Render service (QA and PROD) pointing to the corresponding backend service public URL. If the variable is not set, the entrypoint falls back to `http://localhost:8080/api`.
- The `sed` replacement runs on every container startup. On the nginx:1.25-alpine image with a standard CRA bundle (one or two JS chunks), this adds negligible startup time.
- If the bundle is regenerated (new deploy with a new image), the entrypoint re-runs the replacement — no stale values persist across deploys.
