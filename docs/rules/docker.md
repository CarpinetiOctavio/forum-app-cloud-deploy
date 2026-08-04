# Operating rules — Docker (forum-app-cloud-deploy)

## Purpose and authority
This file governs how a Dockerfile in this repo must be written. Read
`docs/rules/constraints.md` first — every rule here derives from a
constraint stated there; if this file and `constraints.md` ever appear to
conflict, `constraints.md` wins.

## Rule 1 — one image, unmodified, runs in both QA and PROD
**MUST**: the image built by the pipeline is deployed to QA and, after
approval, to PROD without being rebuilt or altered in between.

**Why**: this is the continuous-delivery "build once, promote everywhere"
principle (Humble & Farley) combined with the Twelve-Factor App's strict
separation of config from code (Wiggins, factor III). An image that differs
between QA and PROD — even slightly — means QA tested something other than
what ships; the gate stops meaning what it claims to mean.

**How this MUST be implemented, not just intended**:
- **MUST NOT** write any environment-specific value into a Dockerfile: no
  database path, API URL, port (unless genuinely fixed and
  environment-agnostic), or any value that differs between QA and PROD.
- **MUST** inject every such value at container **runtime**, via
  environment variables set per Render service — never at build time.
- If the frontend's build tool inlines a value at build time in a way that
  would otherwise defeat this rule (Create React App's
  `process.env.REACT_APP_*` is resolved when the bundle is built, not when
  the container starts): **MUST** use a mechanism that substitutes the real
  value into the already-built bundle at container startup instead (an
  entrypoint script performing the substitution, run before the web server
  starts, is one such mechanism — not the only possible one, but it must
  exist in some form). **MUST NOT** conclude that this rule doesn't apply
  to the frontend because CRA makes runtime injection harder — harder to
  implement is not the same as not required.

**How to verify compliance, not just assume it**: the image digest deployed
to QA and the image digest deployed to PROD, for the same pipeline run,
**MUST** be identical. If they differ, Rule 1 was violated somewhere in the
pipeline, regardless of what any Dockerfile says.

## Rule 2 — the runtime image ships one binary and nothing it doesn't need
**MUST**: use a multi-stage build for both the backend and frontend
Dockerfiles.

**Why**: separating the build environment from the runtime environment is
attack-surface reduction, not just image-size discipline — a compiler,
package-manager cache, or source tree present in a *running* container is
usable by anyone who gains any code-execution foothold in it, and none of
it is needed for the container to do its one job.

**Required structure**:
- **Backend**: stage 1 (`builder`) compiles a single, statically linked Go
  binary from source. Stage 2 (final) starts from a minimal base image and
  `COPY`s **only** that compiled binary from stage 1. No source file, no Go
  toolchain, is present in stage 2.
- **Frontend**: stage 1 (`builder`) runs `npm ci` then `npm run build`.
  Stage 2 (final, an `nginx:alpine`-family image) `COPY`s **only** the build
  output directory from stage 1 into nginx's serving directory. No
  `node_modules`, no source, no Node runtime, is present in stage 2.

**Before writing the backend Dockerfile, MUST determine and report**:
whether this repo's real build has a CGO dependency (`go-sqlite3` uses CGO)
— this decides whether `scratch` is viable for the final stage or `alpine`
is required, and MUST be verified against this repo's own `go.mod` and a
real build attempt, not assumed from any other repo's answer to the same
question.

## Rule 3 — layer order follows change frequency, least to most
**MUST** order Dockerfile instructions: base image → working directory →
dependency manifest (`go.mod`/`go.sum` or `package*.json`) → dependency
installation → source code → build step.

**Why**: Docker caches each layer and invalidates it, and every layer after
it, on the first change it detects. Ordering dependency installation before
source-code copy means a source-only commit reuses the cached dependency
layer instead of reinstalling everything. Getting this order backward is
not incorrect in the sense of producing a broken image — the image still
works — but every build reinstalls every dependency from scratch, which is
a real, avoidable cost this rule exists to avoid.

## Rule 4 — `CMD` starts the application, and only the application
**MUST**: `CMD` (or `ENTRYPOINT`, if used instead) starts the application
process as its final action. **MUST NOT**: run database migrations, seed
data, execute health checks, or perform any setup that belongs to the
deployment environment rather than to the application itself, inside `CMD`.

**Why**: see `constraints.md`. If this repo needs a migration or seed step,
that is a decision made and documented on its own terms — with its own
ADR if it's non-obvious — not implemented by adding a line to `CMD` because
it was the path of least resistance at the time.

## Rule 5 — every image tag is a commit SHA
**MUST**: tag every image pushed to the registry with `${{ github.sha }}`.
**MUST NOT**: use `latest` as an image's only tag, anywhere.

**Why**: restated from `constraints.md` — `latest` cannot be traced back to
the commit that produced it, which breaks rollback (you can't redeploy "the
image from three deploys ago" if every deploy overwrote the same tag) and
breaks audit trail (a running container's tag doesn't tell you what code
it's running).

## Rule 6 — every base image version is pinned
**MUST**: pin every base image to a specific version tag — `alpine:3.19`,
not `alpine`; `node:22-alpine`, not `node`; `nginx:1.27-alpine`, not
`nginx:alpine` alone if a more specific pin is available.

**Why**: an unpinned tag is not a fixed reference — `alpine` today and
`alpine` in three weeks can resolve to different underlying images. A
Dockerfile that hasn't changed producing a different image later is a
reproducibility failure, and it fails silently — nothing errors, the build
just quietly ships something different than last time.

## Rule 7 — `.dockerignore` exists for both `backend/` and `frontend/`
**MUST**: both directories have a `.dockerignore` excluding, at minimum,
local databases, build artifacts, logs, and `node_modules`.

**Why**: the build context is everything Docker can potentially copy into
an image layer. A file that's excluded from the context can't leak into an
image by accident, even via an overly broad `COPY . .` — this is a
defense specifically against a mistake elsewhere in the Dockerfile, not a
substitute for writing `COPY` carefully.

## Before writing any Dockerfile — mandatory, in order
1. Determine and report the CGO/SQLite finding from Rule 2.
2. Determine and report this repo's real frontend build-output directory
   and its real mechanism for resolving Rule 1's runtime-injection
   requirement — verified against this repo's actual source, not assumed
   from `cloud-deploy-legacy` or from `qa-pipeline`.
3. Present both findings and the proposed Dockerfile structure for review.
   **MUST NOT** write either Dockerfile before this review is explicitly
   approved.