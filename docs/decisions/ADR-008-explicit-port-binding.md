# ADR-008: Explicit `PORT` Binding — Reading the Environment Instead of Trusting Render's Auto-Detection

**Date:** 2026-08-05
**Status:** Accepted

## Context

`docs/rules/deployment.md`'s environment-variable analysis (a precondition
for configuring anything on Render, per that file's own "Before
configuring or documenting anything" section) required reading this
repo's real backend and frontend source for every value they read from
the environment — not assumed from `-legacy` or from general familiarity
with Render.

That analysis found `DATABASE_PATH` (backend, `ADR-005`) and
`REACT_APP_API_URL` (frontend, `ADR-006`) already handled. It also found
that **neither container's listen port was configurable at all**:

```go
// backend/cmd/api/main.go, before this ADR
if err := http.ListenAndServe(":8080", r); err != nil {
```

```dockerfile
# frontend/Dockerfile, before this ADR
RUN sed -i 's/listen  *80;/listen 8080;/' /etc/nginx/conf.d/default.conf
```

Both ports were fixed at `8080`, one hardcoded in Go source, the other
baked into the image at build time.

**Verified directly against Render's current documentation**
(`render.com/docs/web-services`), not assumed from general platform
familiarity:

> "We recommend binding your HTTP server to the port defined by the
> `PORT` environment variable."
>
> "The default value of `PORT` is `10000` for all Render web services."
>
> "If you bind your HTTP server to a different port, Render is *usually*
> able to detect and use it. If Render fails to detect a bound port, your
> web service's deploy fails and displays an error in your logs."

Render does not strictly require reading `PORT` — auto-detection is a
documented fallback — but Render's own language ("*usually*", "if
detection fails, deploy fails") describes a real, acknowledged failure
mode, not a hypothetical one. This is the same category of risk `ADR-005`
and `ADR-006` already treat as worth closing rather than tolerating:
`docs/rules/docker.md`'s Rule 1 guarantee ("one image, unmodified, runs in
both QA and PROD") is not actually true of a container whose ability to
receive traffic depends on a platform's best-effort port-sniffing.

**Frontend/nginx checked separately, not assumed to behave like the
backend by analogy.** Render's documentation does not carve out a
Docker-specific exception to the `PORT`/auto-detection behavior above —
searched specifically for any distinction between native-runtime and
Docker-sourced web services, found none — so the same risk was treated as
applying to the `nginx`-based frontend container too, absent evidence it
doesn't.

**Implementing the frontend fix surfaced a second, independent finding**:
the straightforward approach — have `docker-entrypoint.sh` `sed` the port
directly into `/etc/nginx/conf.d/default.conf` at container start — fails.
Verified directly, not assumed:

```
$ docker exec <container> sh -c "sed -i 's/listen  *[0-9]*;/listen 9090;/' /etc/nginx/conf.d/default.conf"
sed: can't create temp file '/etc/nginx/conf.d/default.confXXXXXX': Permission denied
```

`/etc/nginx/conf.d/` is root-owned; the container runs as the non-root
`nginx` user since `ADR-006`'s own amendment (`docker:S6471`/`S6504`).
This is the identical shape of problem that amendment already solved for
`/usr/share/nginx/html` — a runtime-user process needing to modify
something the image ships root-owned and read-only — encountered again
here, on a different file, for a different reason.

## Decision

**Backend**: read `PORT` from the environment, falling back to the
existing `8080` default:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
log.Printf("🚀 Server running at http://localhost:%s", port)
if err := http.ListenAndServe(":"+port, r); err != nil {
```

**Frontend**: apply the same `/tmp`-copy pattern `ADR-006`'s amendment
already established, to a second, independent target. `nginx.conf`'s
`include` directive is repointed at build time (a fixed path, not an
environment-specific value, so build-time is fine per `docker.md` Rule 1)
from `/etc/nginx/conf.d/*.conf` to `/tmp/conf.d/*.conf`. At container
start, `docker-entrypoint.sh` copies the **entire** `/etc/nginx/conf.d/`
directory (not only `default.conf`) into `/tmp/conf.d/`, then substitutes
the real `PORT` into the copy:

```sh
PORT="${PORT:-8080}"
mkdir -p /tmp/conf.d
cp -r /etc/nginx/conf.d/. /tmp/conf.d/
sed -i "s/listen  *[0-9]*;/listen ${PORT};/" /tmp/conf.d/default.conf
```

The original `/etc/nginx/conf.d/` stays root-owned and untouched; nginx
reads its config from `/tmp/conf.d/` instead, per the build-time
`include` change.

## Rationale

**Why explicit reads, not trusting auto-detection, for both containers.**
Render's own documented risk ("usually", "if detection fails, deploy
fails") is exactly the kind of implicit, best-effort behavior
`docs/rules/docker.md` Rule 1 and `ADR-005`/`ADR-006` already reject in
other forms — the difference between "the platform's guess about my
container happens to be right" and "my container declares what it needs
and gets it" is the same distinction that motivated reading
`DATABASE_PATH` and `REACT_APP_API_URL` from the environment instead of
hardcoding them. `8080` is kept as the fallback for both — not Render's
own `10000` default — to match this repo's existing convention
(`EXPOSE 8080` in both Dockerfiles, the backend's prior hardcoded value),
so local development and any script assuming `8080` is unaffected.

**Why the frontend fix copies the whole `conf.d/` directory, not just
`default.conf`.** `docs/rules/deployment.md`'s own environment-variable
process already models the discipline this follows: verify against real,
current state, don't assume based on what's true today staying true
indefinitely. `nginx:1.31.3-alpine3.24`'s image ships exactly one file in
`conf.d/` today — but a future base-image bump could add another, and a
fix that only copies the one file known to matter right now would
silently stop serving that hypothetical second file, a regression with no
error message. Copying the directory costs nothing extra and removes that
failure mode entirely.

**Why this doesn't reopen `ADR-006`'s amendment.** The mechanism
(root-owned original, runtime-writable copy in `/tmp`, `nginx` pointed at
the copy) is identical in shape but applied to an independent target —
nginx's own configuration, not the served application bundle. Verified
separately that the fix doesn't reintroduce `docker:S6504`: the original
`/etc/nginx/conf.d/` remains provably unwritable by the `nginx` user
(`touch` against it still returns `Permission denied`) after the fix, the
same check `ADR-006`'s amendment used for `/usr/share/nginx/html`.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|---|---|
| Trust Render's auto-detection, change nothing | Rejected: Render's own documentation describes this as a best-effort fallback with a real, named failure mode ("if Render fails to detect... deploy fails"), not a guarantee — leaving it unaddressed means this repo's declared "one image, both environments" guarantee depends on a platform behavior outside this repo's control, the same category of risk `ADR-005`/`ADR-006` already closed elsewhere. |
| `sed` the port directly into `/etc/nginx/conf.d/default.conf` at container start (the original, untested proposal) | Rejected once verified failing: the file is root-owned, the container runs as non-root `nginx`, and `sed -i` cannot create its temp file there — confirmed with a real `Permission denied`, not assumed. |
| Give the `nginx` user write access to `/etc/nginx/conf.d/` via `--chown` or a runtime `chown` | Rejected: this is exactly the pattern `docker:S6504` and `ADR-006`'s amendment already rejected for `/usr/share/nginx/html` — the runtime user would be able to modify nginx's own routing configuration, a larger blast radius than the served static files it was originally scoped to. |
| Copy only `default.conf` to `/tmp/conf.d/`, not the whole directory | Rejected: works for the image's current contents, but silently stops covering any additional `.conf` file a future base-image version might ship in the same directory, with no error to surface the gap. |

## Consequences

- `backend/cmd/api/main.go` reads `PORT` from the environment, falling
  back to `8080` — every existing local-dev and CI workflow that never
  set this variable continues to work unchanged. Verified: `go build
  ./...` exits 0; `go test ./tests/services/... -cover
  -coverpkg=./internal/services/...` still 50/50 passing, 88.2% coverage,
  unchanged.
- `frontend/Dockerfile`'s build-time port `sed` is removed; `nginx.conf`'s
  `include` path is repointed at `/tmp/conf.d/*.conf` at build time
  instead. `frontend/docker-entrypoint.sh` gains the runtime copy-and-
  substitute step for the whole `conf.d/` directory.
- `PORT` must be set per Render service (or left unset, falling back to
  `8080`) — `docs/rules/deployment.md`'s environment-variable list is
  updated to include it for both services, alongside `DATABASE_PATH` and
  `REACT_APP_API_URL`.
- Verified end-to-end against the real, rebuilt frontend image, not just
  reasoned about: running with `PORT=9095` set, the original
  `/etc/nginx/conf.d/default.conf` and `nginx.conf` stay root-owned and
  unmodified, the runtime copy in `/tmp/conf.d/default.conf` carries the
  substituted port, `touch` against the original directory still fails
  with `Permission denied`, and the container actually serves HTTP `200`
  on the configured port — combined with `REACT_APP_API_URL`'s own
  substitution still working correctly in the same run.
