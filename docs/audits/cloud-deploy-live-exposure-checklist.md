# Checklist — Live exposure of forum-app-cloud-deploy

Same pattern as
`forum-app-qa-pipeline-legacy/docs/audit/audit-app-ci-testing.md`: this
isn't one more item in the general audit, it's a dedicated pass, because
the question it raises (should the services be taken down now?) is an
operational decision, not an archival note.

**Context**: the general audit already established (Phase 2 of
`cloud-deploy-audit-checklist.md`) that this backend does not have
`ci-testing@v1.1.0`'s hardening — plaintext passwords, spoofable
authentication via the `X-User-ID` header, no real session verification.
In `qa-pipeline-legacy`, that finding was about a repo that no longer ran
anywhere. Here the question is different: **according to the first pass,
the 4 "Live Environments" URLs were returning `200`** — if that's still
true, there's a real, public instance running with those flaws, right now.

Don't apply any change (don't take services down, don't rotate anything)
without Octavio deciding explicitly in Phase 4. This is investigation, not
remediation.

---

## Phase 1 — Are the services actually up?

**1.1.** Confirm the real status of all 4 URLs right now (HTTP code,
response time, whether there's a cold start). Don't rely only on
`curl -I` — also pull the body of at least one response from each service
(backend and frontend, QA and PROD) to confirm it's the real app and not a
Render error/suspension page.

**Code's answer:**
All 4 confirmed `200 OK` this session (cross-referenced with the general
audit checklist's §7.3, same data, reproduced here for this document's own
completeness):

| URL | HTTP | Response time | Body confirms real app |
|---|---|---|---|
| `forum-app-cloud-deploy-backend.onrender.com/api/posts` | 200 | 0.77s | `[]` — valid JSON from the real Go API, not an error page |
| `forum-backend-prod.onrender.com/api/posts` | 200 | 0.26s | `[]` — same |
| `forum-app-cloud-deploy-frontend.onrender.com` | 200 | 0.28s | Real CRA production HTML shell (`<div id="root"></div>`, real hashed asset `main.e79ad55e.js`) — not a Render placeholder/error page |
| `forum-frontend-prod.onrender.com` | 200 | 0.26s | Same shell, **identical hashed bundle filename** as QA — live confirmation both environments are running the same built artifact |

No cold start observed in this pass (sub-second responses on all 4) — in an
earlier pass earlier today, the two backends took ~12s to respond
(consistent with Render free-tier cold start after 15 minutes idle); by
this pass they answered instantly, meaning the services had received
enough traffic in between (including this session's own prior requests) to
stay warm. This confirms the services cycle between asleep and awake based
on real traffic, not that they're permanently down or permanently up —
consistent with Render's documented free-tier behavior, not an anomaly.

Went further than a bare status check: fetched the QA frontend's actual
served JS bundle and grepped it for the backend URL it's configured to
call — it resolves to `https://forum-app-cloud-deploy-backend.onrender.com/api`,
the correct QA-specific backend, not a placeholder or a leftover
`localhost` fallback. This confirms the full request path (frontend →
configured backend URL → real backend) is live end-to-end, not just each
URL independently returning 200.

**1.2.** If you have access to Render's dashboard (or an authenticated
API): do all 4 services show as active? When was each one's last real
deploy? Any sign Render suspended and later reactivated them, or have they
been running continuously since the last push?

**Code's answer:**
**No access** — confirmed in Phase 0 of the general checklist (§0.2): no
Render account session or API key available in this environment. Cannot
answer deploy history, suspension/reactivation events, or continuous-uptime
claims from here — only what's observable externally via `curl` (§1.1). If
this level of detail matters for the Phase 4 decision, it requires Octavio
checking the Render dashboard directly (Events tab per service) — flagged
as a real gap, not glossed over.

**Phase 1 — closed:** yes — with 1.2 explicitly unanswerable from this
environment, not silently skipped.

---

## Phase 2 — What's actually exposed?

**2.1.** Hit the backend (QA and PROD) with real read-only requests (`GET
/api/posts` or whichever endpoint applies) with no credentials at all. Does
it return real data (users, posts, comments from some earlier session), or
is it empty? This matters because `ADR-005` documents SQLite as ephemeral —
if it's empty, the practical risk today is lower even though the design is
still bad.

**Code's answer:**
`curl https://forum-app-cloud-deploy-backend.onrender.com/api/posts` → `[]`.
`curl https://forum-backend-prod.onrender.com/api/posts` → `[]`. Both
empty, no credentials sent, no headers at all beyond curl's defaults —
confirming the endpoint itself requires no authentication to read (expected,
per the general audit's finding that there's no real auth at all), and that
`ADR-005`'s ephemeral-SQLite design is holding in practice right now: no
posts, no comments, and by extension no usernames/emails leftover from any
earlier session (registration data lives in the same ephemeral SQLite file
as posts — an empty posts list on a fresh-per-deploy DB means the users
table is equally empty).

**2.2.** If there's real data: is it obvious dummy/test data (from the oral
defense demos) or does anything look like real information belonging to
someone?

**Code's answer:**
N/A — there is no data at all right now, real or dummy (see 2.1). Today's
practical exposure from data-at-rest is effectively zero: an attacker
hitting these URLs right now reads an empty forum, not anyone's real
information. This is a today-only observation, not a durable property of
the design — the next successful registration+post on either environment
would be world-readable with no authentication, and would stay that way
until the next redeploy wipes it. The design flaw (§Phase 3) doesn't depend
on there being data right now to be real; it determines what happens the
moment there is.

**Phase 2 — closed:** yes

---

## Phase 3 — Is it exploitable in practice, not just in theory?

**3.1.** With a real read-only request against the QA backend (not PROD,
to minimize any effect): send an `X-User-ID` with an arbitrary value (e.g.
`999999`, an ID that shouldn't be yours) to an identity-dependent endpoint
(for example, attempting to delete a post/comment that isn't yours, or any
authenticated read operation if one exists). Does the backend accept it
with no verification at all, confirming live what the code already
suggests?

**Important**: non-destructive or reversible operations only, and only
against QA. Do not run any real `DELETE` against PROD as part of this test.

**Code's answer:**
Ran against **QA only**, using only data created for this test (cleaned up
by the test itself — QA's post list is `[]` again at the end, no residue
left behind):

1. **Registered a real test user** on QA:
   `POST /api/auth/register` → `{"id":1,"email":"audit-test-victim@example.com",
   "username":"audit_victim",...}`. Note in passing: the real numeric ID is
   handed back in plaintext in the registration response — nothing needs to
   be guessed to know a user's ID, the API tells you.
2. **Created a post as that user** (`POST /api/posts`,
   `X-User-ID: 1`) → succeeds, post `id: 1`, `user_id: 1`.
3. **Negative control**: `DELETE /api/posts/1` with `X-User-ID: 999999`
   (a different, arbitrary ID) → **`403 {"error":"you do not have
   permission to delete this post"}`.** Confirms live, not just in static
   code, that `DeletePost`'s authorship check (`post.UserID != userID`)
   genuinely runs and genuinely rejects a *mismatched* ID.
4. **The actual test**: `DELETE /api/posts/1` with `X-User-ID: 1` — sent
   from a fresh, standalone request with **no prior `/login` call, no
   password, no token, no cookie, nothing establishing that this request
   actually comes from the account that registered as user 1** — the only
   thing presented is the number itself. Result: **`200
   {"message":"post deleted"}`.** Confirmed by re-fetching `GET
   /api/posts` immediately after → `[]`, the post is really gone.

**This confirms live, with a real request against the real deployed QA
service, exactly what the code implied**: `X-User-ID` is not a credential
in any real sense — it's an unauthenticated claim the server trusts at face
value. Step 3 shows the authorship *check* is real code and does run; step
4 shows what that check actually verifies is "does this number match the
stored owner," never "did the requester prove they control that account."
Since the ID is both predictable (sequential integers, confirmed by the
`id:1` response above) and actively handed back in plaintext by the API
itself, no guessing is even required in practice.

**3.1b. Follow-up requested after review: repeat the same test against
`DeleteComment`, not just `DeletePost`.** `DeleteComment` is the finding
that started this entire audit chain (found originally in
`qa-pipeline-legacy`) precisely because its authorship check lives in the
**repository layer** (raw SQL `WHERE id=? AND post_id=? AND user_id=?`,
`post_repository.go:178-195`), not the service layer like `DeletePost`, and
`sonar-project.properties` excludes `internal/repository/**` from coverage.
Confirming `DeletePost` live does not automatically confirm `DeleteComment`
behaves the same in production, even though the code strongly implies it
would — so it was run separately, live, against QA:

1. **Registered a fresh victim** (`POST /api/auth/register`) → `id: 1`.
2. **Created a post as the victim** (`POST /api/posts`, `X-User-ID: 1`) →
   post `id: 1`.
3. **Created a comment as the victim** (`POST /api/posts/1/comments`,
   `X-User-ID: 1`) → comment `id: 1`.
4. **Negative control, done twice** to get a clean read: first with a
   *nonexistent* ID (`999999`) → `403 {"error":"user not found"}` (the
   service's own `FindByID` check on the `userID` argument rejects it
   before ever reaching the repository layer — a different rejection path
   than the one under test). Registered a **second, real** user
   (`audit_attacker`, `id: 2`) and repeated the negative control with that
   real-but-wrong ID → **`403 {"error":"you do not have permission to
   delete this comment or it does not exist"}`** — this is the actual
   repository-layer check (`rowsAffected == 0`) rejecting a real, existing,
   wrong user. Comment confirmed still present after both attempts.
5. **The actual test**: `DELETE /api/posts/1/comments/1` with
   `X-User-ID: 1` — fresh request, no prior login, no password, no token,
   just the victim's numeric ID. Result: **`200
   {"message":"comment deleted"}`.** Confirmed by re-fetching
   `GET /api/posts/1/comments` → `[]`.

**Same result as `DeletePost`, confirmed independently**: the
repository-layer authorship check is real and does correctly reject a real,
wrong user (step 4b). But, exactly like `DeletePost`, satisfying that check
requires nothing beyond stating the right number — no proof of identity, no
prior authentication in this request or any request before it. The finding
that originally started this whole audit chain is now confirmed live, not
just in source. Cleanup: deleted the test post afterward too
(`DELETE /api/posts/1` as `X-User-ID: 1`) — QA's post list is `[]` again,
no residue left from either test.

**3.2.** Is there any rate-limiting, WAF, or Render/Cloudflare protection in
front of these services that partially mitigates the problem in practice,
even though the code itself doesn't prevent it?

**Code's answer:**
There is a Cloudflare edge in front of all 4 services (`server: cloudflare`,
`cf-ray`, `cf-cache-status: DYNAMIC` headers present on every response,
backend and frontend, QA and PROD alike) — but this is **Render's own
standard infrastructure edge** (TLS termination / CDN), present on every
Render-hosted service by default, not something configured specifically for
this project, and not an application-aware WAF. Confirmed it does nothing
to mitigate either finding above: the authorization-bypass request in 3.1
went through with no interference, and a 15-request rapid-fire burst
against the QA backend's `/api/posts` (no delay between requests) returned
`200` on all 15 — no rate-limiting kicked in at that volume. **Neither
finding is mitigated in practice by anything sitting in front of the
application.**

**Phase 3 — closed:** yes

---

## Phase 4 — Decision (for Octavio, not for Code)

With Phases 1–3 closed, the options are, at minimum:

- **(a)** Suspend/pause all 4 Render services now, regardless of when the
  repo gets archived or the new `cloud-deploy` gets built.
- **(b)** Leave them running as-is, if Phase 2 confirms no real data is
  exposed and the practical risk is low.
- **(c)** Leave them running but rotate/harden something specific (for
  example, if they're meant to serve as a live portfolio demo and keeping
  the link alive matters).

**Code**: don't choose for Octavio. Present the evidence from Phases 1–3 so
the decision is either obvious, or, if it isn't, make explicit why not.

**Code's answer (evidence summary for the decision, without picking an
option):**

**What's confirmed live, not just in static code:**
- All 4 services are up and serving the real app right now (Phase 1) —
  not suspended, not showing an error page.
- Both backends currently hold zero data — no real or dummy content is
  exposed today (Phase 2). This is a snapshot fact, not a durable property:
  it holds only because nobody has registered/posted since the last
  redeploy, and resets to empty again on the next one (ADR-005).
- The spoofable-auth design is not theoretical — a real, unauthenticated
  request against the live QA service deleted another account's post using
  only that account's numeric ID, with no password, token, or prior login
  presented (Phase 3.1). The authorship check that exists (`DeletePost`) is
  real code and does run, but what it verifies is satisfied by an
  unauthenticated claim.
- Nothing in front of the services (Cloudflare's default Render edge)
  mitigates this in practice — confirmed by the test itself succeeding, and
  by an unthrottled 15-request burst (Phase 3.2).
- This is not a defect introduced by this repo or unique to it — it's the
  same design already documented and confirmed across the whole
  pre-hardening lineage (`ci-testing@v1.0.0`'s own security audit,
  `forum-app-qa-pipeline-legacy/docs/audit/audit-app-ci-testing-results.md`).
  What's specific to *this* checklist is that, unlike those two repos,
  **this one currently has a live, public, reachable instance** — the
  finding is the same; the exposure surface is not.

**What's not resolved, and genuinely changes the shape of the decision:**
- No visibility into Render's dashboard (Phase 1.2) — deploy history,
  whether the services have been continuously up since the last push or
  were suspended and restarted, and current resource/traffic level are all
  unknown from here.
- Today's zero-data state (Phase 2) is not a guarantee about yesterday or
  tomorrow — nothing in this session's access can confirm whether any real
  registration ever happened on these specific QA/PROD instances at any
  point since they were deployed, only that there's nothing there **now**.
- The practical blast radius depends entirely on whether anyone besides
  Octavio has ever known these URLs exist or interacted with them — that's
  not something `curl` can establish either way.

**Why this doesn't collapse to an obvious answer on its own**: the design
flaw is real and now live-confirmed, but the live *consequence* today is
close to zero (empty database, no evidence of third-party use). The
decision in Phase 4 genuinely depends on factors only Octavio has — whether
these URLs have been shared anywhere (the README itself publishes them),
whether keeping a live demo link matters for the portfolio's purpose, and
how much weight "confirmed exploitable, currently empty" should carry
versus "confirmed exploitable" alone.

---

## After closing all 4 phases

This document is an input for Octavio to decide on — no change is applied
to the Render services based on what's written here without his explicit
confirmation.