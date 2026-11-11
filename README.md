# Forum App — Cloud Deploy

Work in progress. This repository extends the pipeline of [forum-app-qa-pipeline](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline) with container images, a container registry, and deployment to QA and production on Render. Today `main` runs tests and quality gates and builds and publishes the images to ghcr.io; deployment is not implemented on `main` yet.

This README will be rewritten when deployment works. Until then, the decisions made so far are in [`docs/decisions/`](docs/decisions/).

## AI usage disclosure

This repository is built with substantial help from Claude (Anthropic), used through Claude Code and claude.ai: as an auditor of the reasoning and the documentation, to verify claims against the real state of the repository, and as a writer of documentation and code drafts for review. Claude does not make decisions: scope, design, tooling and every change that lands are decided and approved by Octavio Carpineti. Not every commit Claude contributed to carries a `Co-Authored-By` trailer.
