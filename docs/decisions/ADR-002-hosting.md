# ADR-002 — Cloud Hosting: Render.com

## Status

Accepted

## Date

2025

## Context

The application needs to be deployed to two independent cloud environments — QA and PROD — that can run Docker containers pulled from a registry, support environment variable configuration per service, provide public URLs, and be triggered programmatically from a CI/CD pipeline. The solution must be free of cost.

## Decision

Use Render.com as the cloud hosting provider for both QA and PROD environments.

## Rationale

- **Docker-native:** Render supports deploying directly from a container registry. It pulls the image, runs it, and manages the container lifecycle — no build step on the hosting side.
- **Deploy hooks:** Render provides a unique HTTP endpoint per service (deploy hook) that triggers a new deployment when called. This integrates cleanly with GitHub Actions via an HTTP POST step, without requiring a Render-specific GitHub Action or SDK.
- **Independent service configuration:** QA and PROD are two separate Render services. Each has its own environment variables, resources, and deploy hook. They are fully isolated from each other.
- **Free tier:** Render's free tier supports web services with enough availability for academic and demonstration purposes.
- **Zero-downtime deployment:** Render starts the new container, waits for it to become healthy, and only then terminates the previous one.

## Alternatives considered

| Alternative | Reason not chosen |
|-------------|------------------|
| Google Cloud Run | Free tier is generous, but setup requires a GCP project, IAM configuration, and the `gcloud` CLI or a GitHub Action with service account credentials. Higher setup complexity for the same outcome. |
| Fly.io | Viable free tier, but deployment requires the `flyctl` CLI in the pipeline, adding a tool dependency. Less intuitive environment separation between QA and PROD. |
| Railway.app | Free tier is credit-based ($5/month) and can be exhausted unpredictably. Deploy hook support is less straightforward than Render. |
| AWS App Runner | Requires AWS account, IAM roles, and ECR or Docker Hub as registry. Cost and setup complexity exceed the project's constraints. |
| Heroku | No longer offers a free tier for container-based deployments. |

## Consequences

- Two Render services must be created manually: `forum-app-qa` and `forum-app-prod`.
- Each service is configured to pull from ghcr.io using the commit SHA tag.
- Deploy hook URLs are stored as GitHub Secrets scoped to their respective environments (`qa` and `prod`).
- Render free tier imposes resource limits (shared CPU, 512MB RAM, spin-down after inactivity). This is acceptable for QA and demonstration purposes but would not be suitable for production workloads with real users.
- SQLite persistence behavior on Render's free tier must be resolved before deployment. See `docs/rules/constraints.md` — pending items.