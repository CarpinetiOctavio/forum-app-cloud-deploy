# Lot traceability

The analogy was written by Octavio Carpineti; Claude corrected it against the repository, translated it and added the links.

## The concept

At a pharmaceutical company, a medicine, for example an ibuprofen 600 mg tablet in its "Plus" variant, is made from a formula with a version. Each production run gives a **lot** with its **lot number**, and all the boxes of a lot are the same; two lots made from the same formula version carry the same printed label ("formula version X") and different lot numbers. Quality control releases **a lot**, not the formula, and a lot made later does not inherit that release, because it is another lot. After a failure, a recall goes back to a lot already released, identified by its number, not to "the previous formula": manufacturing again from that formula gives a new lot that no one checked.

## The transfer

![Two lanes with the same stages. Pharmaceutical: formula version X, manufacturing run, lot L1, release by quality control, distribution; a second run of the same formula gives lot L2, same label, not released. Pipeline: commit C, build, image with digest D1, QA with smoke test and approval, PROD; a rerun of the same commit gives digest D2, same tag, which now points to it, never deployed. If a later lot or release fails, the recall goes back to lot L1 and the rollback to digest D1, through release again; going back by tag reaches the last build, which no one released. Side panel: the app is a two-piece kit, backend and frontend, each with its own lot, validated separately.](diagrams/lot-traceability.svg)

The figure maps each step; the evidence is that two builds of identical code produced different digests (runs `43306631470` and `43341781168`, [execution record](audits/2026-11-24-execution-record.md), section 1; screenshots [02](screenshots/02-digests-run-43306631470.png) and [03](screenshots/03-digests-run-43341781168.png)), that a rerun of the same commit moves its tag to the new image and that PROD receives the digest QA tested, without a rebuild ([ADR-009](decisions/ADR-009-single-build-trigger.md); [ADR-007](decisions/ADR-007-render-deploy-mechanism.md), amendment of 2026-11-11), and that a rollback goes back to a digest that was in PROD, through QA and approval again ([ADR-007](decisions/ADR-007-render-deploy-mechanism.md), amendment of 2026-11-23).

## Where it breaks

- **A two-piece kit.** The app has two pieces, backend and frontend, each with its own lot. At a pharmaceutical company, a kit is released as a unit. Here the rollback validates each lot on its own ([`rollback.yml`](../.github/workflows/rollback.yml)), so it can assemble a kit from two released lots that were never tested together. And in a normal deploy the pieces are delivered one after the other ([`ci.yml`](../.github/workflows/ci.yml)): if the second fails, the kit is left mixed.
- **A physical lot runs out.** An image can be copied without limit, so going back to an old lot is possible as long as it exists in the registry.
