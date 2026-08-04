# Transferable knowledge — forum-app-cloud-deploy (pre-`-legacy`)

**Date:** 2026-08-03

Platform and tooling gotchas from this repo that are worth carrying into the
new `cloud-deploy` build so they aren't rediscovered from scratch. Kept
separate from `cloud-deploy-legacy-audit-results.md` on purpose — that
document is an architectural/documentation audit; this one is pure
"how the platform actually behaves," extracted with light editing from this
repo's own `ADR-007-deployment-issues.md`, `SETUP.md`, and `COMMANDS.md`,
plus two items found independently in this session that ADR-007 doesn't
cover.

None of this is a recommendation for the new repo's architecture — it's
operational memory: what broke, why, and the fix that worked, so the same
hour isn't spent rediscovering it.

---

## 1. Apple Silicon (arm64) builds are incompatible with Render

**Symptom**: a container built and pushed manually from a MacBook with an
M-series chip fails to start on Render with an architecture mismatch.

**Cause**: `docker build` / `docker buildx build` without `--platform`
default to the host architecture. On Apple Silicon that's `arm64`. Render's
infrastructure runs `amd64` only.

**Fix**: always pass `--platform linux/amd64` explicitly for any image
intended for Render, when building locally:
```bash
docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/<image-name>:latest \
  --push ./backend
```

**Does not affect CI**: GitHub Actions' `ubuntu-latest` runners are `amd64`
natively, so `docker/build-push-action` in the pipeline produces the correct
architecture with no extra configuration. This only bites when pushing
manually from a Mac — e.g. the one-time bootstrap push needed before the
pipeline can push to a package that doesn't exist yet (see §2).

---

## 2. `GITHUB_TOKEN` can't push to ghcr.io — two separate permission problems, both required

**Symptom**: `docker-build-push` job fails with `permission_denied:
write_package`, even after adding `permissions: packages: write` at the
**job** level in `ci.yml`.

**Cause — two compounding issues, fixing only one is not enough:**

1. `permissions: packages: write` must also be declared at the **workflow
   level** (top of `ci.yml`, not just inside the job) for `GITHUB_TOKEN` to
   get write access to the package registry at all.
2. A package created by pushing manually with a personal PAT is **not**
   automatically linked to any repository. `GITHUB_TOKEN`, scoped to the
   repository the workflow runs in, can only write to packages explicitly
   linked to that repository — regardless of the `permissions` block.

**Fix — both steps required:**
- Add `permissions: contents: read` / `packages: write` at the workflow-level
  block in `ci.yml`, in addition to (not instead of) the job-level block.
- For each package: GitHub → your account → **Packages** tab → select the
  package → **Package settings** → **Connect repository** → select the repo.

**Bootstrap ordering note** (from `SETUP.md`): the pipeline can't push to a
package that doesn't exist yet, so the very first images have to be pushed
manually (with a PAT, `--platform linux/amd64` per §1, tagged `latest` just
for that one-time bootstrap) to create the packages — *then* linked to the
repo, *then* the pipeline can push with `GITHUB_TOKEN` from then on.

---

## 3. Non-root Docker user — two separate fixes, in sequence

**Part A — SonarCloud flags root containers (`docker:S6471`)**

**Symptom**: SonarCloud blocks the pipeline: final Dockerfile stage has no
`USER` instruction, so the app runs as `root` inside the container —
flagged as a security vulnerability.

**Fix**: add a non-root system user/group in the final stage, transfer
binary ownership at copy time:
```dockerfile
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder --chown=appuser:appgroup /app/server .
USER appuser
```
The **builder** stage stays root — `gcc`/the Go compiler need it to compile.
Only the final runtime stage needs the non-root user.

**Part B — the non-root user then can't write to its own working directory**

**Symptom**: once Part A is applied, the container starts, but SQLite fails
at runtime: `failed to initialize database: unable to open database file:
no such file or directory`.

**Cause**: `WORKDIR /app` creates the directory as `root:root` *before*
`USER appuser` takes effect. `COPY --chown=appuser:appgroup` transfers
ownership of the **files** copied in, but not of the **directory** itself —
`appuser` still can't create new files (like `database.db`) inside it.

**Fix**: `chown` the directory explicitly, as root, between `WORKDIR` and
`COPY`:
```dockerfile
WORKDIR /app
RUN chown appuser:appgroup /app
COPY --from=builder --chown=appuser:appgroup /app/server .
USER appuser
```

**Takeaway for the new repo**: if a non-root user is added to satisfy a
SonarCloud/security rule, budget for this second fix too — it only shows up
at container *runtime*, not at build time or in SonarCloud's static
analysis, so it can look like the first fix was sufficient until the app
actually tries to write to disk.

---

## 4. Frontend runtime config for a single cross-environment image (CRA)

**Problem**: Create React App inlines `process.env.REACT_APP_*` at **build
time** — there's no way to read an env var from the container host at
runtime once the bundle exists. But the constraint (in this repo, and
plausibly in the new one too) is one image, deployed identically to QA and
PROD, with only the backend URL differing between them.

**Pattern that worked** — entrypoint placeholder replacement:

1. Dockerfile builder stage: `ARG REACT_APP_API_URL=__REACT_APP_API_URL__`
   — this makes CRA bake the **literal placeholder string** into the bundle
   instead of `undefined` (important: if the arg is unset, CRA's `||`
   fallback in source code would otherwise kick in at build time and there'd
   be nothing left for the entrypoint to find and replace).
2. A shell entrypoint script runs at container start, before nginx:
   ```sh
   #!/bin/sh
   set -e
   API_URL="${REACT_APP_API_URL:-http://localhost:8080/api}"
   find /usr/share/nginx/html -name '*.js' -exec sed -i "s|__REACT_APP_API_URL__|${API_URL}|g" {} \;
   exec nginx -g 'daemon off;'
   ```
3. `ENTRYPOINT ["/docker-entrypoint.sh"]` replaces the Dockerfile's `CMD`.

**Cost**: negligible — one `sed` pass over one or two JS chunks at
container startup, re-run fresh on every new container (no stale values
carried across deploys, since it's not persisted anywhere).

**Alternatives considered and rejected in this repo** (all still valid to
re-evaluate for the new one): building a separate image per environment
(breaks the single-artifact constraint outright); an nginx reverse proxy
with relative `/api/` paths (works, but means changing every API call site
in the frontend to use relative URLs); a `window._env_` global loaded from
a separate `public/env.js` script (works, more moving parts — an extra
script tag in `index.html`, a global, and every service file reading from
it instead of `process.env`).

---

## 5. Render API key hardcoded, then leaked to a public repo — incident and fix

**What happened**: the rollback instructions in `docs/COMMANDS.md` §10 were
written with a real, working `RENDER_API_KEY` value pasted in directly
(`export RENDER_API_KEY="rnd_..."`), as a "verified, copy-pasteable"
example, instead of a placeholder. Since the repo is public, the key was
exposed to anyone with read access from the commit that introduced it
onward — commit history keeps it readable even after a later commit removes
it from the current file.

**Caught by**: manual review ahead of using the repo as a portfolio piece —
**not** by any automated pipeline gate. There is no secret-scanning step in
this pipeline or as a pre-commit hook.

**Fix applied**:
1. Replaced the real value with a placeholder in `docs/COMMANDS.md`
   (`export RENDER_API_KEY="your-render-api-key"`).
2. **Rotated the key in the Render dashboard** — this is the part that
   actually neutralizes the leak. Editing the file after the fact does
   nothing about a key that's already sitting in git history; only
   invalidating the leaked value at the source does.
3. Updated the corresponding `RENDER_API_KEY` GitHub Actions secret used by
   `deploy-qa`/`deploy-prod` with the new value.

**What the fix missed** (found independently this session, not in
`ADR-007`): the same rollback section also had two real Render **service
IDs** (`srv-...`) hardcoded, filled in at the same time as the API key.
Those were never redacted or rotated — service IDs aren't secrets in the
same sense (they can't authenticate anything on their own), but they're
still real infrastructure identifiers left exposed in a section of the repo
that a parallel file (`SETUP.md`) treats as sensitive enough to
placeholder. Worth building a secret-scanning step (even a lightweight
regex pre-commit hook for `rnd_`, `sk_`, `srv-` patterns) into the new
repo's pipeline from the start, rather than relying on a manual pass before
each portfolio review — this incident was caught once by luck of timing,
not by any repeatable mechanism.

**Takeaway for the new repo**: "verified working example" in docs is a
trap — a real value is more convincing to write in the moment, and it's
exactly what turns a docs file into a credential leak. Placeholder values
that are obviously placeholders (`your-render-api-key`, `srv-xxxxxxxxxxxxxxxx`)
cost nothing in clarity and remove the risk entirely.

---

## 6. GitHub environment-review requests expire silently after 30 days

Not documented anywhere in this repo (found independently this session, via
the GitHub API, investigating an apparently-unexplained `deploy-prod`
failure — full writeup in the audit doc, §4.8). A `deploy-prod` job that
sits waiting for a required-reviewer approval, if never approved or
rejected, is auto-failed by GitHub after **exactly 30 days** — with an empty
step list, so the job log gives no indication of what "failed." If the new
repo keeps a manual-approval PROD gate, this is worth knowing going in:
an abandoned approval doesn't stay pending forever, and when it expires it
looks exactly like a real pipeline failure at a glance.

---

## 7. Render free-tier services: cold-start delay depends on recent traffic, not a fixed number

Observed empirically this session, not documented anywhere in this repo.
The same 4 URLs were hit twice, hours apart: the first pass measured
~12 seconds to first byte on both backends (consistent with Render's
documented free-tier spin-down after 15 minutes idle); a later pass, after
several rounds of manual `curl`/exploit-verification traffic in between,
measured sub-second responses on all 4 services. The wake-up cost isn't a
constant — it depends entirely on how recently *any* request (including a
health check, a monitoring ping, or manual testing traffic like this
session's own) touched the service. Worth knowing for the new repo if a
live demo link is part of the portfolio pitch: the "first click after a
while" experience is real and will vary, not something a fixed loading
spinner or timeout budget can fully paper over on the free tier.

---

## 8. `go-sqlite3` needs CGO — incompatible with a `scratch` final stage

`go-sqlite3` (this app's SQLite driver) is a CGO binding, not pure Go —
compiling it requires `CGO_ENABLED=1` and a C toolchain (`gcc`/`musl-dev`)
in the builder stage. The consequence carries into the final stage's base
image choice: a `CGO_ENABLED=1` binary is dynamically linked against libc
and will not run on `FROM scratch` (which has no libc at all) — it needs
at minimum a minimal libc-providing image like `alpine`. This is a Go/CGO
fact, not something specific to this app's code: any project using
`go-sqlite3` (or another CGO-based dependency) hits the same constraint,
regardless of what the rest of the codebase looks like. `-legacy`'s own
`docs/rules/docker.md` states this correctly (*"prefer alpine over scratch
unless the binary is fully statically linked with no CGO dependencies"*) —
worth carrying the reasoning forward, not just the conclusion, since a
future dependency change (e.g., swapping to a pure-Go SQLite driver) would
flip which base image is actually optimal.

---

## 9. Plan documents drift silently once implementation changes — update them as part of finishing the work, not separately

Not a tool-behavior gotcha like the other 8 items — a process pattern,
found by comparing which of `-legacy`'s docs match its real, running
pipeline and which don't. `docs/SETUP.md` and `docs/COMMANDS.md` correctly
describe the Render REST API deploy mechanism this repo actually uses.
`docs/decisions/ADR-002-hosting.md` and three separate `docs/rules/` files
(`deployment.md`, `pipeline.md`, and `constraints.md`) all instead describe
Render **deploy hooks** — an earlier, simpler mechanism that was
apparently the original plan and never got implemented, or was replaced
without anyone going back to update the four documents that had already
committed it to writing. The split is clean: the two documents that
happened to get touched again after the real mechanism was built are
accurate; the four that were written once, early, and never revisited are
wrong — in the same specific way, independently, in each of the four.

**Takeaway for the new repo**: a rules file or an ADR that describes a
plan *before* the corresponding code exists is normal and fine — but it
needs a deliberate re-read once the real implementation lands, specifically
checking "does this still describe what's actually running," not just
"did I write this once already." Treating that reconciliation pass as
part of finishing a feature (the same way tests or docs review already
are) would have caught this before it reached four files instead of one.
