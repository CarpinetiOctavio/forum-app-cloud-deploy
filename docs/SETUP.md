# Setup

## Prerequisites

```bash
go version     # 1.24 or higher (see backend/go.mod)
node --version # 18 or higher (matches .github/workflows/ci.yml)
npm --version
```

> The Go module was renamed from `tp06-testing` to `forum-app-ci-testing`
> when this repo was rebuilt as a portfolio piece. Any reference to the old
> name in git history is intentional and traceable to ADR-004.

**Installing Go:**
```bash
# macOS
brew install go
# Ubuntu/Debian
sudo apt install golang-go
# Windows: https://go.dev/dl/
```

**Installing Node.js:**
```bash
# macOS
brew install node
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
# Windows: https://nodejs.org/
```

## Clone and install

```bash
git clone https://github.com/CarpinetiOctavio/forum-app-ci-testing.git
cd forum-app-ci-testing

cd backend && go mod download && cd ..
cd frontend && npm install && cd ..
```

## Running locally

```bash
# Terminal 1 — backend
cd backend && go run cmd/api/main.go
# http://localhost:8080

# Terminal 2 — frontend
cd frontend && npm start
# http://localhost:3000
```

## Database schema

`backend/internal/database/database.go` creates `backend/database.db` on first
run (`InitDB("./database.db")` in `main.go`), using `CREATE TABLE IF NOT
EXISTS` so re-running the app never wipes existing data. Three tables:

```
users                          posts                           comments
├── id (PK)                    ├── id (PK)                     ├── id (PK)
├── email (UNIQUE)              ├── title                       ├── post_id (FK → posts)
├── password                    ├── content                     ├── user_id (FK → users)
├── username                    ├── user_id (FK → users)         ├── content
└── created_at                  └── created_at                  └── created_at
```

Both foreign keys that reference `users`, and the one referencing `posts`,
are declared `ON DELETE CASCADE`. Deleting a user deletes their posts *and*
every comment on those posts; deleting a post deletes its comments. This
matters for local testing: seeding a user, a few posts, and comments, then
deleting the user, is enough to verify the cascade — no manual cleanup of
orphaned rows is needed, and none is possible to forget. Indexes exist on
all three foreign-key columns (`idx_posts_user_id`,
`idx_comments_post_id`, `idx_comments_user_id`).

Tests never touch this file — every repository is mocked in the service-layer
tests, so `database.db` is only ever created by actually running the app
(`go run cmd/api/main.go`) or the E2E suite.

## Frontend service layer

Components never call `axios` directly. `frontend/src/services/authService.ts`
and `postService.ts` wrap every HTTP call and return typed data
(`response.data`, not the full Axios response). A component only ever calls
`authService.login(credentials)` or `postService.createPost(data, userId)` —
it has no idea `axios` is involved, what the exact URL is, or that HTTP is
happening at all.

This pays off in two concrete ways: changing the API base URL or switching
HTTP libraries touches two files instead of every component that makes a
request, and testing a component only requires mocking `axios` once (via
`jest.mock('axios')`) rather than stubbing HTTP calls scattered across
component code.

## What this repo does NOT need

Unlike the later repos in this series, this repo requires no external
infrastructure:
- No GitHub Secrets — the CI pipeline uses no external service (see ADR-004).
  Coverage is uploaded as an internal pipeline artifact, not to Codecov or
  any third-party reporting tool.
- No container registry, no cloud hosting account, no Render/AWS/GCP
  configuration — those belong to `forum-app-cloud-deploy`.
- No SonarCloud token — static analysis is out of this repo's scope (see
  `docs/rules/testing.md`); it belongs to `forum-app-qa-pipeline`.

Cloning the repo and running the two install commands above is the entire
setup.
