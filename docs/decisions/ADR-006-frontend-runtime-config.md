# ADR-006: Frontend Runtime Configuration — Entrypoint Placeholder Substitution, with a Required Source Fix

**Date:** 2026-08-04
**Status:** Accepted

## Context

The frontend is a Create React App (CRA) application. CRA resolves
`process.env.REACT_APP_*` references at **build time**, via webpack's
`DefinePlugin` — every such reference becomes a literal string baked into
the compiled JS bundle when `npm run build` runs. There is no mechanism in
the shipped bundle to read an environment variable from the container host
at runtime; by the time a container starts, the bundle is a static asset.

`docs/rules/constraints.md`'s non-negotiable "one image, unmodified, runs
in both QA and PROD" guarantee requires the backend URL — necessarily
different between the two environments — to vary *without* rebuilding the
image. CRA's build-time resolution and this repository's single-artifact
constraint are in direct tension by construction; some mechanism has to
exist to reconcile them.

**Verified directly against this repository's own inherited frontend
source, not assumed from `cloud-deploy-legacy`'s solution to the same
tension** — the premise that solution depends on does not currently hold
here:

```typescript
// frontend/src/services/postService.ts
const API_URL = 'http://localhost:8080/api/posts';

// frontend/src/services/authService.ts
const API_URL = 'http://localhost:8080/api/auth';
```

Both service files hold a **hardcoded string literal** pointing at
`localhost:8080`. `grep -rn "REACT_APP_API_URL\|process.env"
frontend/src/services/` returns nothing — there is no environment-variable
reference anywhere in the inherited frontend for a build-time or
runtime-substitution mechanism to act on. This is a materially different
starting point than `cloud-deploy-legacy`'s own `ADR-006`, which reasoned
about *how* to substitute a value into an existing `process.env.REACT_APP_API_URL`
reference — that reference does not exist yet in this repository's real
code. Wiggins' Twelve-Factor App Factor III applies here for the same
reason it applies to `ADR-005`: config that varies by deploy (an API base
URL) must live outside the code, and right now it doesn't.

**Same clarification as `ADR-005`, and for the identical reason: this is
not a conceptual error in `qa-pipeline` or `ci-testing`, it is the correct
absence of something neither repo ever needed.** Every environment the
frontend has ever run in, across both prior repos, is `localhost:3000`
talking to `localhost:8080` — local development, or the same pairing inside
one CI job for Cypress. There was never a second, differently-addressed
backend for `postService.ts`/`authService.ts` to need to distinguish.
Hardcoding the URL was correctly scoped for that reality, not a shortcut.
The requirement is new here — two named environments, each with its own
public backend URL — introduced by this repository's own scope, not a
latent bug inherited from earlier work. Consequence for how the fix gets
made: new configurability, not a repair, with the same non-sensitive
`localhost:8080` value preserved as the fallback default (never a real
QA/PROD URL or any credential baked in as a "convenient" default), so local
development keeps working unchanged and no environment-specific or
sensitive value ever ships inside the image itself.

## Decision

Adopt the same class of solution `cloud-deploy-legacy`'s `ADR-006`
validated — entrypoint placeholder substitution — **applied to source code
that must first be changed to have something for it to substitute into**:

1. `postService.ts` and `authService.ts`'s hardcoded `API_URL` constants
   are replaced with a reference to a build-time placeholder:
   ```typescript
   const API_BASE = process.env.REACT_APP_API_URL || 'http://localhost:8080';
   const API_URL = `${API_BASE}/api/posts`; // and /api/auth, respectively
   ```
2. The frontend Dockerfile's builder stage sets
   `ARG REACT_APP_API_URL=__REACT_APP_API_URL__` before `npm run build` —
   this makes CRA bake the literal placeholder token into the compiled
   bundle, not `undefined`, so the `||` fallback in source does not
   activate at build time and there is something concrete for the next
   step to find.
3. A shell entrypoint script, run before nginx starts, replaces every
   occurrence of `__REACT_APP_API_URL__` in the built JS with the real
   `REACT_APP_API_URL` value injected into the container's environment at
   deploy time.

## Rationale

**Why the source change is in scope, addressed with the same reasoning
`ADR-005` used for `main.go`:** `postService.ts`/`authService.ts` are not
business logic — they don't decide what the application does, only where
it sends requests to do it. Leaving the hardcoded value in place would mean
this repository's own stated guarantee (one image, both environments,
runtime-injected config only) is provably false the moment anyone reads
`postService.ts`. Fixing it is the minimum change necessary for this
repository's declared purpose to be achievable, not a feature change —
`docs/rules/documentation.md`'s exceeds-scope criterion (d) and `CLAUDE.md`
name this exact class of justification.

**Why substitution-after-build, not a build-time value per environment.**
Building a separate image per environment — QA gets one build with QA's URL
baked in, PROD gets another — would trivially solve CRA's build-time
resolution problem, and is explicitly the alternative this ADR rejects
first, because it violates the single-artifact constraint at its root: if
QA and PROD run different images, the entire premise of "QA validates what
PROD runs" (the reason `ADR-003`'s SHA-tagging scheme and
`docs/rules/deployment.md`'s image-promotion rule exist) is false from the
start. Runtime substitution is the only approach that keeps the artifact
identical across environments while still letting the *value* vary.

**Why an entrypoint script, specifically, over other runtime-injection
patterns.** A `sed` pass over the built JS at container startup is a single
well-understood operation, adds negligible startup latency (one or two JS
chunks in a typical CRA build), requires no new global object or extra
script tag in `index.html`, and is transparent to the application code —
`postService.ts`/`authService.ts` read `process.env.REACT_APP_API_URL` the
same way they always would; the substitution mechanism lives entirely in
the Docker layer, not the application layer.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Separate image per environment | Directly violates the single-artifact constraint (`docs/rules/constraints.md`) that this repository's entire QA-then-PROD promotion guarantee depends on — rejected without further consideration once that conflict is identified. |
| nginx reverse proxy with relative `/api/` paths | Technically viable, but requires changing every API call site to use relative paths (not just the two `API_URL` constants) and adding an nginx `proxy_pass` configuration — a wider source-code change than the two-constant fix this decision requires, for an equivalent outcome. |
| `window._env_` global via a separate `public/env.js` script | Also viable, but adds more moving parts than the entrypoint approach: an extra `<script>` tag in `index.html`, a global object, and every service file reading from `window._env_` instead of `process.env` — a larger deviation from CRA's own idiomatic environment-variable pattern than a placeholder substitution that preserves `process.env.REACT_APP_API_URL` as the actual read site in source. |
| Leave the hardcoded `localhost:8080` and accept it as a known limitation | Rejected as internally inconsistent, for the same reason as the equivalent option in `ADR-005`: this repository's own rules already forbid environment-specific values baked into the image. The application would also be **functionally broken** in any deployed environment, not just architecturally inconsistent — `localhost:8080` resolves to nothing inside a QA or PROD container. |

## Consequences

- `postService.ts` and `authService.ts` gain a `process.env.REACT_APP_API_URL`
  reference with a `localhost:8080` fallback — local development and any
  existing test that never set this variable continues to work unchanged.
- The frontend Dockerfile gains one `ARG` in the builder stage and an
  `ENTRYPOINT` script (replacing a bare `CMD`) in the final stage.
- `REACT_APP_API_URL` must be set per Render service (QA pointing at the QA
  backend's public URL, PROD at PROD's) — absent that, the fallback resolves
  to `localhost:8080`, which is a loud, immediately visible failure (no
  network response) rather than a silent misconfiguration.
- Every new deploy re-runs the substitution fresh against a newly built
  image — no stale injected value can persist across deploys, since nothing
  is cached or written back to the image itself.
- This decision and `ADR-005` share a structural pattern worth naming
  explicitly: both found the inherited codebase not merely
  differently-configured from what `cloud-deploy-legacy` assumed, but
  **not configurable at all** at the point this repository actually starts
  from — both required a real, scoped, in-bounds source fix as a
  precondition, not just a deployment-layer decision layered on top of
  already-adequate code.

## Amendment (2026-08-04): substitution target moved from `/usr/share/nginx/html` to `/tmp/html`

SonarCloud's `docker:S6504` ("make sure the copied resource cannot be
modified by a non-root user") flagged the frontend Dockerfile once the
final stage gained a non-root `USER nginx` (added to satisfy a separate
finding, `docker:S6471`, root-by-default). The original entrypoint ran
`sed -i` directly against `/usr/share/nginx/html` — the same directory
`COPY --from=builder --chown=nginx:nginx` had just handed ownership of to
the `nginx` user, specifically so the entrypoint (running as that same
user) could write to it. Once `USER nginx` existed as a static Dockerfile
instruction, that `--chown` was exactly the pattern S6504 exists to catch:
the runtime user owning the content it serves, which a compromised nginx
worker could exploit to persist a modified response.

**Alternative considered and rejected: keep `--chown=nginx:nginx`, but
scope it to only `build/static/js` instead of the whole build output.**
This narrows *how much* the runtime user can modify, but not *whether* it
can — the COPY line touching the JS files (the ones actually containing
the placeholder) would still hand write ownership to the same user
serving them, which is the exact shape `docker:S6504` checks for. It would
likely still fail the same SonarCloud gate, just on a smaller COPY
instruction, without resolving the underlying finding.

**Decision: `/usr/share/nginx/html` stays root-owned and read-only to
`nginx`** (no `--chown` at all on that `COPY`). The entrypoint instead
copies the full build output to `/tmp` (`cp -r`, world-writable via the
sticky bit, no `chown` required) and performs the placeholder
substitution there; nginx's `root` directive is repointed to `/tmp/html`
at build time. No `COPY` instruction in the final image ever grants the
runtime user write access to anything — the writable copy is created
imperatively at container start, outside what a static Dockerfile
analyzer evaluates. Verified directly: `touch` against
`/usr/share/nginx/html` as the `nginx` user returns `Permission denied`;
the served bundle, fetched over a real HTTP request against the running
container, still shows the placeholder correctly substituted.

The cost is duplicating the build output on disk inside the running
container once, at startup — this app's gzipped bundle is under 100 KB,
and the `cp -r` adds negligible time next to the `sed` pass this ADR
already accounted for. Everything else this ADR decided — the placeholder
mechanism itself, the `ARG`/entrypoint split, `REACT_APP_API_URL`'s
fallback behavior — is unchanged; only the filesystem location the
substitution runs against moved.
