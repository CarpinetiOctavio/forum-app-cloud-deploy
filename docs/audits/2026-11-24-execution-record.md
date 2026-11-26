# Execution record — 2026-11-24

A dated record of the pipeline and rollback runs that serve as this repository's evidence. It is not updated: later runs are not added here.

Run IDs, jobs and deployments come from the GitHub API and the run logs; times are UTC, as GitHub shows them. GitHub keeps Actions logs for 90 days, so each section links the screenshot that remains after that ([`docs/screenshots/`](../screenshots/README.md)) and says what to look at in it.

GitHub's run list numbers the rollback runs "Rollback #1" to "#4", and #1 is the rejected input of section 3.1. This record names runs by ID; "rollback 1" to "rollback 3" are runs #2 to #4.

## 1. One build per run, a different digest per build

Commits `7ee46bf` and `3b158c0` have identical `backend/` and `frontend/` trees, and their runs pushed different digests for both images: backend `b1714332…` and `3c17899b…`, frontend `1a76a1f8…` and `5b2d5aed…` (the `Image:` lines of the PROD deploy in run `43306631470`, [screenshot 02](../screenshots/02-digests-run-43306631470.png), and in run `43341781168`, [screenshot 03](../screenshots/03-digests-run-43341781168.png)).

## 2. The pipeline on `main`: run `45336308644`

Push to `main` at `a14fb88`, the merge of pull request #25, at 05:52:11. All eleven jobs succeeded ([screenshot 04](../screenshots/04-pipeline-run-on-main.png): the job graph, from the tests to Deploy to PROD).

- **Quality gate.** The SonarCloud job waited for the gate: `QUALITY GATE STATUS: PASSED`.
- **One image, promoted.** QA and PROD received the same index digests, and Render reported the same platform manifest for each in both environments:

  | Image | Digest deployed (QA and PROD) | Manifest reported by Render |
  |---|---|---|
  | backend | `sha256:3b941a5638a4ea10a712088ce1af8e9fd19f4722de9506d697184febf6b0c9b9` | `sha256:d5ab114931d4fd683de6b745f8ae55ac9da5e42bdc357ac4179d821b71bdd1c6` |
  | frontend | `sha256:f26a49dc84c9ca3e86159a6b2bf77f04f9cb488acebb2c6b4232de4e228b4b09` | `sha256:cc8aed4239a5e80f1ce0adbe40edf21234353b6e71e2452e7351011802fd0940` |

- **Approval.** `Deploy to PROD (Render)` was created at 05:57:45, waited for the required reviewer, and ran from 06:01:08 to 06:01:59 ([screenshot 05](../screenshots/05-prod-approval-run-45336308644.png): the approval under "Deployment protection rules"; the Artifacts table below it lists build records and coverage reports, not images, and its digests are those of the artifact archives).
- The smoke tests passed in QA and in PROD. The run created the `prod` deployment `8019972867`.

## 3. Rollback

The three rollbacks were run from `main` (`a14fb88`) with these inputs:

| Rollback | Target | Backend digest | Frontend digest |
|---|---|---|---|
| 1 | `3b158c0`, the previous PROD | `sha256:3c17899b08e6d09f982d8340a147c0baca4895c7e8b5a6696169f4854c13521f` | `sha256:5b2d5aed55cad445e77a1031e3eb891ab5065cf2fbd3955a6380eb2e7c48a1c6` |
| 2 and 3 | `a14fb88`, back to the current PROD | `sha256:3b941a56…` (section 2) | `sha256:f26a49dc…` (section 2) |

### 3.1 A malformed input is rejected: run `45438714033`

The commit SHA `3b158c0` was passed instead of a digest. `Validate rollback digests` failed on the format check, before any deploy hook was called; `Roll back QA` and `Roll back PROD` were skipped and no deployment was created ([screenshot 06](../screenshots/06-rollback-rejects-malformed-input.png): the annotation "'3b158c0' is not a digest of the form sha256:<64 hex characters>.").

### 3.2 Rollback 1, to `3b158c0`: run `45439653236`

- **Validation.** Both digests were accepted against deployment `7754168946`, created by run `43341781168` of `ci.yml`.
- **Deploy.** QA ran from 16:39:43 to 16:40:35; PROD, after approval, from 16:43:02 to 16:43:53 ([screenshot 07](../screenshots/07-rollback-1-approved.png)). Both received the input digests, Render reported the manifests `8657b0fc…` and `36cf7a42…` in both, and the smoke tests passed.
- The run created the `prod` deployment `8034330468`, which records `a14fb88`: the commit of `main` when the run started, not the commit of the images it deployed.

### 3.3 Rollback 2, to `a14fb88`: run `45442208645`

The validation counts only deployments created by `ci.yml` (`ADR-007`, amendment of 2026-11-23). When it ran (16:52:47 to 16:53:02), the API listed the `prod` deployments in this order, newest first:

| Deployment | SHA | Created by |
|---|---|---|
| `8034330468` | `a14fb88` | `rollback.yml`, run `45439653236` (rollback 1) |
| `8019972867` | `a14fb88` | `ci.yml`, run `45336308644` |
| `7754168946` | `3b158c0` | `ci.yml`, run `43341781168` |
| `7749736048` | `7ee46bf` | `ci.yml`, run `43306631470` |

Both digests were accepted against `8019972867`, the second in the list ([screenshot 08](../screenshots/08-rollback-2-validation-accepts-ci-deployment.png): each step ends "deployed to prod from main (deployment 8019972867)"). The first, `8034330468`, has the same SHA, and the tag `a14fb88` points to the same digests, but it was created by `rollback.yml`, so it was not counted. The script does not log the deployments it skips; the log is read against this list.

QA ran from 16:53:07 to 16:53:58 and passed its smoke tests. `Roll back PROD` was created at 16:53:59 and waited for approval ([screenshot 09](../screenshots/09-rollback-2-prod-waiting.png)). It was approved only after rollback 3 had been started.

### 3.4 Rollback 3, started while rollback 2 waited for approval: run `45442757655`

The deploy jobs of `ci.yml` and `rollback.yml` share the concurrency group `render-deploy`. Rollback 3 was started to see what a job of that group does while another waits for PROD approval.

- Validation passed from 16:55:40 to 16:55:53, against `8019972867`.
- `Roll back QA` was created at 16:55:54 and stayed pending; rollback 2's PROD job was not cancelled. The run page showed "A job in this workflow is waiting for Roll back PROD (Render) (Rollback #3) to complete before running" ([screenshot 10](../screenshots/10-rollback-3-qa-held-by-waiting-prod.png)). The `qa` environment has no protection rules.
- Rollback 2's PROD job was approved ([screenshot 11](../screenshots/11-rollback-2-approved.png)) and ran from 17:00:45 to 17:02:22. Rollback 3's QA started at 17:02:26, 4 s after it finished, and ran until 17:03:16.
- Rollback 3's QA was pending for 6 min 32 s: 4 min 51 s while rollback 2's PROD waited for approval, and 1 min 41 s while it ran.
- Rollback 3's PROD waited for approval from 17:03:17, was approved ([screenshot 12](../screenshots/12-rollback-3-approved.png)) and ran from 17:06:09 to 17:07:01.

What this shows about the group is recorded in `ADR-007`'s amendment of 2026-11-24.

### 3.5 The same sequence, as Render records it

The events of `forum-backend-prod` show the three PROD deploys at 16:43, 17:01 and 17:06 ([screenshot 13](../screenshots/13-render-prod-deploy-events.png), in local time, UTC−8): each "Triggered via Deploy Hook", with the index digest the hook passed on the right and the manifest Render runs in the title (`8657b0f` for `3b158c0`, `d5ab114` for `a14fb88`).

### 3.6 `prod` deployments after the sequence

In the order the API returns them, all with status `success`:

| Deployment | SHA | Created by |
|---|---|---|
| `8034923760` | `a14fb88` | `rollback.yml`, run `45442757655` |
| `8034677791` | `a14fb88` | `rollback.yml`, run `45442208645` |
| `8034330468` | `a14fb88` | `rollback.yml`, run `45439653236` |
| `8019972867` | `a14fb88` | `ci.yml`, run `45336308644` |
| `7754168946` | `3b158c0` | `ci.yml`, run `43341781168` |
| `7749736048` | `7ee46bf` | `ci.yml`, run `43306631470` |

PROD ended on the images of `a14fb88` (section 2).

## 4. The quality gate blocks a pull request into `main`: pull request #26

The ruleset on `staging` and `main` requires nine checks, including both Docker builds ([screenshot 01](../screenshots/01-docker-checks-required.png): pull request #24, "Required" next to each).

Pull request #26 (`evidence/gate-red` → `main`, "DO NOT MERGE — quality gate evidence") added one commit, `504ac38`, with `backend/internal/services/gate_probe.go`: 26 statements without a test. The branch is outside `feature/**`, so no push run started and no image of `504ac38` was pushed to ghcr.io. Run `45448056858` (`pull_request`):

- `Backend Tests (Go)` passed: coverage of `internal/services` dropped from 88.2% to 75.3%, above the 70% threshold.
- `SonarCloud Code Analysis` failed: `QUALITY GATE STATUS: FAILED`, on the condition "0.0% Coverage on New Code (required ≥ 80%)".
- `Test Summary` failed; the Docker and deploy jobs were skipped.
- Both failing jobs are required checks, and merging was blocked ([screenshot 14](../screenshots/14-quality-gate-blocks-pr-into-main.png)).

The size of the probe was chosen so that this condition, and not the coverage threshold, would fail: SonarCloud does not apply the coverage condition to fewer than 20 new lines, and around 40 untested statements bring the total coverage under 70%, failing `Backend Tests (Go)` first and skipping the SonarCloud job. The pull request was closed without merging at 17:29:02, and its branch was deleted.
