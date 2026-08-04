# Operating rules — verification (forum-app-cloud-deploy)

## "Done" requires evidence, not a summary
During this repo's own setup, more than one status report described something as
complete when it wasn't: four audit documents that existed only locally, a branch
ruleset reported as empty that was actually active, a working branch with real
changes that was never pushed to origin, a README claimed to be reverted that was
actually freshly rewritten. None of these were caught by re-reading the report —
they were caught by checking the actual repo state directly.

Before reporting a task as done: provide the thing that makes it checkable — a
commit hash, a GitHub compare link, the literal command output, a diff — not a
prose description of what should be true. If it can't be checked without trusting
the report, say what's unverified, not that it's done.

## When reviewing someone else's "done"
The same standard applies in reverse: don't accept "done" at face value just
because a report is confident or detailed. Re-derive the specific, checkable claim
(a file exists at this path, a number matches this count, a branch is at this
commit) and confirm it directly when the tooling allows it.

## Numbers specifically
Any test count, coverage percentage, or line count stated in an ADR, README, or
audit must be checked against the actual file or command output before being
trusted — not assumed correct because it appears in a document that looks
authoritative. This repo's own predecessor (`forum-app-cloud-deploy-legacy`)
had multiple instances of stale or false claims surviving undetected across
several documents at once — the same wrong deploy-hook mechanism repeated,
unverified, in four separate files, and the same false "backend logic
inherited from qa-pipeline" claim repeated in two — the standard here is to
verify each one, not to inherit the habit of not checking.