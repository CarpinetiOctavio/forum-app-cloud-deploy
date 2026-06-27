# Project Setup Guide

This guide documents the full configuration process required to run the
forum-app-cloud-deploy CI/CD pipeline from scratch. Follow each step in
order — later steps depend on earlier ones.

---

## Prerequisites

- GitHub account with admin access to the repository
- Render.com account (free tier is sufficient)
- Docker Desktop installed and running locally

---

## Step 1 — Create a GitHub Personal Access Token (PAT)

The PAT is used to push Docker images to ghcr.io from your local machine
during initial setup. The pipeline itself uses `GITHUB_TOKEN` after that.

1. GitHub → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. **Generate new token (classic)**
3. Select scopes: `write:packages`, `read:packages`, `repo`
4. Copy the token immediately — it is not shown again

---

## Step 2 — Build and push initial images to ghcr.io

The pipeline cannot push images to a package that does not yet exist.
Push an initial `latest`-tagged image from your local machine to create
the packages, then link them to the repository in Step 3.

**Important:** build with `--platform linux/amd64`. Render runs on amd64.
Building on Apple Silicon (arm64) without this flag produces incompatible
images.

```bash
echo YOUR_PAT | docker login ghcr.io -u carpinetioctavio --password-stdin

docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:latest \
  --push ./backend

docker buildx build --platform linux/amd64 \
  -t ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend:latest \
  --push ./frontend
```

---

## Step 3 — Link packages to the repository

`GITHUB_TOKEN` in Actions can only write to packages that are linked to
the repository from which the workflow runs. Do this for both packages.

For each package (`forum-app-cloud-deploy-backend` and `forum-app-cloud-deploy-frontend`):

1. `github.com/carpinetioctavio` → **Packages** tab
2. Click the package → **Package settings**
3. Under **Connect repository** → select `CarpinetiOctavio/forum-app-cloud-deploy`

---

## Step 4 — Configure Workflow permissions

1. Repository → **Settings** → **Actions** → **General**
2. Under **Workflow permissions** → select **Read and write permissions**
3. Save

---

## Step 5 — Create Render services

Create four Web Services in Render. For each service:

- **Type:** Web Service
- **Source:** Existing image registry
- **Registry:** `ghcr.io`
- **Registry credentials:** username = `carpinetioctavio`, password = PAT with `read:packages` scope

| Service name | Image | Port |
|---|---|---|
| `forum-backend-qa` | `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:latest` | 8080 |
| `forum-frontend-qa` | `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend:latest` | 80 |
| `forum-backend-prod` | `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-backend:latest` | 8080 |
| `forum-frontend-prod` | `ghcr.io/carpinetioctavio/forum-app-cloud-deploy-frontend:latest` | 80 |

### Environment variables per service

**forum-backend-qa** and **forum-backend-prod:**

| Variable | Value |
|---|---|
| `PORT` | `8080` |
| `DATABASE_PATH` | `./database.db` |

**forum-frontend-qa:**

| Variable | Value |
|---|---|
| `REACT_APP_API_URL` | Public URL of `forum-backend-qa` + `/api` (e.g. `https://forum-backend-qa.onrender.com/api`) |

**forum-frontend-prod:**

| Variable | Value |
|---|---|
| `REACT_APP_API_URL` | Public URL of `forum-backend-prod` + `/api` (e.g. `https://forum-backend-prod.onrender.com/api`) |

After creating each service, copy its **Service ID** from the dashboard URL
(`srv-xxxxxxxxxxxxxxxx`) — you will need it in Step 7.

---

## Step 6 — Create GitHub environments

Repository → **Settings** → **Environments** → **New environment**

| Environment | Protection rule |
|---|---|
| `qa` | None (automatic deploy) |
| `prod` | **Required reviewers** → add your GitHub username → Save |

---

## Step 7 — Configure secrets

### Repository-level secret

**Settings** → **Secrets and variables** → **Actions** → **New repository secret**

| Secret | Value |
|---|---|
| `RENDER_API_KEY` | API key from Render → **Account Settings** → **API Keys** → Create API Key |

### Environment `qa` secrets

**Settings** → **Environments** → `qa` → **Add secret**

| Secret | Value |
|---|---|
| `RENDER_SERVICE_ID_BACKEND_QA` | `srv-xxxxxxxxxxxxxxxx` (ID of `forum-backend-qa`) |
| `RENDER_SERVICE_ID_FRONTEND_QA` | `srv-xxxxxxxxxxxxxxxx` (ID of `forum-frontend-qa`) |

### Environment `prod` secrets

**Settings** → **Environments** → `prod` → **Add secret**

| Secret | Value |
|---|---|
| `RENDER_SERVICE_ID_BACKEND_PROD` | `srv-xxxxxxxxxxxxxxxx` (ID of `forum-backend-prod`) |
| `RENDER_SERVICE_ID_FRONTEND_PROD` | `srv-xxxxxxxxxxxxxxxx` (ID of `forum-frontend-prod`) |
