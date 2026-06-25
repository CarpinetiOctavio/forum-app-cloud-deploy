# Docker — Rules and Requirements

## Purpose

This file defines how Docker must be used in this project. Read `docs/rules/constraints.md` before this file. Any conflict between this file and `constraints.md` resolves in favor of `constraints.md`.

---

## Multi-stage build — mandatory for both Dockerfiles

Every Dockerfile in this project uses a multi-stage build. The rationale is separation between build environment and runtime environment: the final image contains only what is needed to run the application, not what was needed to build it.

**Backend Dockerfile — required structure:**

- Stage 1 (`builder`): starts from a Go base image, copies source code, compiles the binary. The output is a single statically linked executable.
- Stage 2 (final): starts from a minimal base image (`alpine` or `scratch`), copies only the compiled binary from stage 1. Nothing else.

**Frontend Dockerfile — required structure:**

- Stage 1 (`builder`): starts from a Node base image, copies `package.json` and `package-lock.json`, runs `npm ci`, copies source code, runs `npm run build`. The output is the `/build` directory.
- Stage 2 (final): starts from an `nginx:alpine` base image, copies only the `/build` output from stage 1 into nginx's serving directory.

---

## Layer ordering — cache efficiency

Instructions in each Dockerfile must be ordered from least-frequently-changing to most-frequently-changing:

1. Base image (`FROM`)
2. Working directory (`WORKDIR`)
3. Dependency files (`COPY go.mod go.sum ./` or `COPY package*.json ./`)
4. Dependency installation (`RUN go mod download` or `RUN npm ci`)
5. Source code (`COPY . .`)
6. Build step (`RUN go build` or `RUN npm run build`)

This order ensures that the dependency installation layer is cached and not re-executed on every source code change.

---

## CMD — single responsibility

The `CMD` instruction must only start the application process. It must not:

- Run database migrations
- Run seed scripts
- Execute health checks
- Perform any initialization that belongs to the environment, not the application

If database migrations are required, they are handled as a separate concern outside the Dockerfile. See `docs/rules/constraints.md` — pending items.

---

## Environment variables — runtime injection only

No environment-specific value is baked into any Dockerfile. This includes:

- Database connection strings or file paths
- API URLs
- Port numbers (unless they are truly fixed and environment-agnostic)
- Any value that differs between QA and PROD

All such values are injected at runtime via environment variables configured in Render per service. The image is environment-agnostic — the same image runs in QA and PROD without modification.

---

## Image tagging — commit SHA only

Images pushed to ghcr.io are tagged with the Git commit SHA (`${{ github.sha }}`). This is enforced in `ci.yml` and must not be changed to `latest` or any mutable tag. See `docs/decisions/ADR-003-image-tagging.md` for the full rationale.

---

## Base image selection

- Go final stage: prefer `alpine` over `scratch` unless the binary is fully statically linked with no CGO dependencies. **Read the pending item on SQLite and CGO in `docs/rules/constraints.md` before writing the backend Dockerfile** — `go-sqlite3` uses CGO, which affects base image choice and build flags.
- Frontend final stage: `nginx:alpine`.
- Always pin base image versions — never use unversioned tags like `alpine` or `node`. Use `alpine:3.19`, `node:20-alpine`, `nginx:1.25-alpine`, etc.

---

## .dockerignore

Both `backend/` and `frontend/` must have a `.dockerignore` file. At minimum:

**Backend `.dockerignore`:**
```
*.db
*.out
/tmp
```

**Frontend `.dockerignore`:**
```
node_modules
build
*.log
```

This prevents unnecessary files from being sent to the Docker build context, keeping builds fast and images clean.

---

## What Code must do before writing any Dockerfile

1. Read `docs/rules/constraints.md` — pending items section.
2. Analyze the existing backend source to determine whether CGO is required (due to `go-sqlite3`) and report the finding before choosing the base image for the Go final stage.
3. Analyze the existing frontend source to determine the correct build output directory and the API URL configuration (`REACT_APP_API_URL` or equivalent) before writing the frontend Dockerfile.
4. Report findings and proposed Dockerfile structure before writing any file.