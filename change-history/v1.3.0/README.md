# QuranLingo

A Duolingo-style app for learning the words of the Qur'an. Go backend + React Native (Expo) mobile app, in one monorepo.

**Current version: v1.3.0** — see [CHANGELOG.md](./CHANGELOG.md) for what changed and [/change-history](./change-history) for a full README snapshot at every release.

---

## Table of contents

- [What the app does](#what-the-app-does)
- [Repository layout](#repository-layout)
- [Stack](#stack)
- [Gamification rules](#gamification-rules)
- [Course content](#course-content)
- [Backend API](#backend-api)
- [Admin panel](#admin-panel)
- [Data model](#data-model)
- [Mobile app](#mobile-app)
- [Security baselines](#security-baselines)
- [Getting started](#getting-started)
- [Versioning & documentation process](#versioning--documentation-process)
- [Engineering conventions & best practices](#engineering-conventions--best-practices)

---

## What the app does

A learner works through a Duolingo-style skill tree (**125 Words of the Qur'an** — the 125 most frequent words in the Qur'an, together accounting for roughly half the text). Each skill unlocks a 5-word lesson; each lesson is a set of multiple-choice exercises, graded by the server. Correct lessons earn XP, wrong answers cost a heart, and daily activity builds a streak — the standard gamification loop, implemented with one hard rule: **the client never grades itself**. See [Gamification rules](#gamification-rules).

## Repository layout

```
backend/    Go API (module root, cmd/api entrypoint; cmd/seed and cmd/hashpassword are dev tools)
mobile/     React Native (Expo) app, TypeScript
change-history/   Full README.md snapshot as of each tagged version
CHANGELOG.md      Dated, itemized log of what changed release to release
Makefile    Single entrypoint for running/building/testing either app — see `make help`
```

Always use `make <target>` (`make backend-run`, `make frontend-start`, `make dev`, …) instead of `cd`-ing into a subfolder and invoking `go`/`npm`/`expo` directly, so usage stays consistent for anyone (or any agent) working in this repo.

## Stack

- **Backend**: Go, [chi](https://github.com/go-chi/chi) router (`backend/internal/handler/router.go`). PostgreSQL (Supabase-hosted) is the source of truth.
- **DB access**: hand-written parameterized SQL via `pgx/v5` in `backend/internal/repository/` (no ORM, no codegen). `golang-migrate` for schema migrations (`backend/internal/db/migrations`).
- **Background jobs**: not yet wired up (Redis/`asynq` documented as a future addition). XP/streak/hearts updates happen synchronously inside the request that triggers them — fine at current scale.
- **Object storage**: not yet implemented — no audio/image lesson assets in the seeded course yet. When added, binary assets go through an S3-compatible bucket, never through the API process or Postgres.
- **Frontend**: React Native, Expo SDK 54, TypeScript.
- **Server state**: TanStack Query (`@tanstack/react-query`).
- **Client state**: Zustand (`mobile/src/store/authStore.ts`).
- **Offline**: `expo-sqlite` caches course/lesson content and queues lesson submissions made while offline (`mobile/src/db/`) — see [Mobile app](#mobile-app).
- **Auth**: JWT access token (short-lived, HS256) + opaque refresh token (rotated and revoked on use, hashed at rest) — `backend/internal/service/auth_service.go`. Passwords hashed with bcrypt.

## Gamification rules

All of the following is enforced **server-side only**, inside `backend/internal/service/lesson_service.go`. The mobile app never computes a score — it submits raw answers and displays whatever the server returns.

- **Hearts**: start at 5 (max). Each wrong answer in a lesson costs 1 heart, down to a floor of 0.
- **Heart refill**: once hearts hit 0, the server stamps `hearts_refill_at = now + 4h`. This is a *lazy* refill — there's no background job — the check only happens the next time that user submits a lesson: if 4 hours have passed, hearts reset to 5 (`MaxHearts`) at that moment; otherwise the submission is rejected with "no hearts remaining."
- **XP**: a completed lesson awards its fixed `xp_reward` (10 by default) regardless of score, added to `total_xp`.
- **Score**: `correct_answers / total_exercises * 100`, computed by re-grading every answer server-side against the stored correct answer/option.
- **Streak**: `current_streak` increments if the learner's last activity was yesterday, resets to 1 if there's a gap, stays flat if they already practiced today. `longest_streak` tracks the historical max.
- **Idempotency**: every submission carries a client-generated `idempotency_key`. Retried/replayed submissions (e.g. from the mobile offline queue) with the same key return the original result instead of double-awarding XP — enforced by a DB unique constraint on `(user_id, idempotency_key)`, not just application logic.
- **Lesson/skill locking**: a lesson unlocks once the previous lesson in its skill is completed; a skill unlocks once every lesson in the previous skill is completed. The first skill/lesson is always unlocked. Computed in `content_service.go`, enforced both when listing the course tree (`GET /courses/{code}`) and when fetching a specific lesson (`GET /lessons/{id}` returns 403 if locked).

## Course content

Seeded by `make backend-seed` (`backend/internal/db/seed/`), upsert-based so it's safe to re-run. Course: **"125 Words of the Qur'an"** (`code: quran-125`) — 25 themed skills × 5 words each, sourced from *"125 Words of the Qur'an"* by Dr. Abdulazeez Abdulraheem (Understand Al-Qur'an Academy). Every word is written with full harakat (tashkeel/diacritics) for readability. Each word becomes one multiple-choice exercise; Arabic text is *display-only* — grading always compares against the English answer, never the Arabic string.

Beyond the seeded course, operators can add more multiple-choice questions to any lesson through the admin panel's [question management](#admin-panel) page — one at a time or in bulk via CSV — without touching the seed data or redeploying.

The `Word` struct also carries `Transliteration`, `Root` (Arabic trilateral root), `Type` (noun/verb/particle), and `Occurrences` (frequency in the Qur'an) as reference metadata. Not written to the DB by the current seeder — kept in the dataset so future exercise types (e.g. root-family drills) don't require re-deriving the data.

## Backend API

REST + JSON. Base URL is whatever `APP_PORT` is bound to (default `:8080`).

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET` | `/healthz` | none | liveness check |
| `POST` | `/auth/register` | none | `{email, password, display_name}` → `{user, tokens}` (201). Rate-limited 10/min/IP. |
| `POST` | `/auth/login` | none | `{email, password}` → `{user, tokens}`. Rate-limited 10/min/IP. |
| `POST` | `/auth/refresh` | none | `{refresh_token}` → `{tokens}` (rotates the refresh token). Rate-limited 10/min/IP. |
| `GET` | `/me` | Bearer JWT | current user's profile/stats |
| `GET` | `/courses/{code}` | Bearer JWT | full skill/lesson tree with per-item lock status for this user |
| `GET` | `/lessons/{id}` | Bearer JWT | exercises for one lesson, **correct answers withheld** |
| `POST` | `/lessons/{id}/submit` | Bearer JWT | `{idempotency_key, answers[]}` → grades, awards XP, updates hearts/streak. Rate-limited 30/min/IP. |
| `GET` | `/leaderboard/weekly` | Bearer JWT | top 50 users by this week's XP |
| `GET` | `/admin/login` | none | admin sign-in page |
| `POST` | `/admin/login` | none | `username`+`password` form → signed session cookie. Rate-limited 10/min/IP. |
| `GET` | `/admin` | admin session | operator panel — see below |
| `POST` | `/admin/logout` | admin session | clears the session cookie |
| `POST` | `/admin/users/refill-all` | admin session | refill every user's hearts to full |
| `POST` | `/admin/users/{id}/refill` | admin session | refill one user's hearts to full |
| `GET` | `/admin/questions` | admin session | question management — lesson picker, add/import forms, draft preview list |
| `POST` | `/admin/questions/new` | admin session | create one draft question from a form |
| `POST` | `/admin/questions/import` | admin session | bulk-create draft questions from an uploaded CSV |
| `POST` | `/admin/questions/{id}/update` | admin session | overwrite a draft's prompt/Arabic text/options |
| `POST` | `/admin/questions/{id}/approve` | admin session | publish a draft — makes it visible to the app |
| `POST` | `/admin/questions/{id}/delete` | admin session | reject/remove a draft |

`Authorization: Bearer <access_token>` is required on every JWT-protected route. Access tokens are short-lived (`JWT_ACCESS_TTL`, default 15m); the mobile client transparently refreshes on 401 using the refresh token (`JWT_REFRESH_TTL`, default 720h/30d).

## Admin panel

A small, server-rendered HTML page at `/admin` (Go `html/template`, no JS framework, no separate deployable) for the one operator task that currently needs a human: refilling hearts without waiting out the 4-hour lazy refill. It lists every user (email, hearts, refill timer, XP, streak, join date) with a **"Refill All Hearts"** button and a per-row **Refill** button.

- **Auth**: a dedicated sign-in page at `/admin/login` (not a browser-native HTTP Basic Auth prompt) checks the submitted username/password against a single operator credential (`ADMIN_USERNAME` / `ADMIN_PASSWORD_HASH` in `backend/.env`) with `bcrypt.CompareHashAndPassword`, then issues a signed, `HttpOnly`, `SameSite=Strict` session cookie (`ADMIN_SESSION_SECRET`, HMAC-SHA256, 12h expiry). This is completely separate from the JWT user-auth system — **there is no admin registration flow of any kind**; the only admin account that can ever exist is whatever `ADMIN_USERNAME`/`ADMIN_PASSWORD_HASH` resolve to in the environment. `/admin/logout` clears the cookie. If any of the three env vars is unset, none of the `/admin` routes are mounted at all (fails closed).
- Generate a password hash with `make backend-admin-hash password=yourpassword` and a session secret with `make backend-admin-secret`. **The hash contains `$` characters — always single-quote it in `.env`**, otherwise a shell `source` (as the Makefile's migrate targets do) will try to expand it as positional parameters and corrupt it.
- Not shipped in the mobile app binary — it's a backend-only surface, so nothing admin-capable ever ships inside the public app.
- Known MVP limitation: one shared operator credential, no audit log of who refilled what. Fine for a single-developer project; would need real per-admin accounts and an audit trail before handing admin access to more than one person.

### Question management (`/admin/questions`)

Lets an operator add multiple-choice questions to any lesson without touching the seed data, with a draft → preview → approve workflow so nothing reaches the app unreviewed:

- **Add one question** via a form (prompt, Arabic text with harakat, 2-4 options, check all that are correct), or **bulk-import a CSV** (`prompt,arabic_text,option_1..option_4,correct_option` — UTF-8, so Arabic pastes in as-is; `correct_option` is the 1-based `option_N` number(s) that are correct, `;`-separated for more than one, e.g. `1;3`). Bad rows are skipped and reported individually rather than failing the whole import.
- **More than one option can be marked correct.** Arabic words often carry a general sense and one or more specific ones (e.g. الْحَمْد ≈ "all praise and thanks" / "praise" / "thanks"), so a question isn't limited to one right answer — the learner is graded correct if they pick *any* option flagged correct. `exercise_options.is_correct` has never had a uniqueness constraint; only the admin-side validation previously required exactly one, and that's what changed.
- Every question created this way starts as a **draft** (`exercises.status = 'draft'`) — invisible to `GET /lessons/{id}` and to lesson grading, which both only ever see `status = 'approved'` rows.
- The page renders each draft as a **preview styled like the actual Android lesson screen** (same layout/colors as `mobile/src/screens/LessonScreen.tsx`), with every correct option highlighted for the operator's eyes only, so they can proof the question exactly as a learner would see it.
- An operator can **edit** any draft inline (fix a typo, add/swap correct answers) before publishing, or **delete** it outright. Clicking **Approve** flips it to `approved`, making it live immediately — no redeploy, no reseed.
- **Existing questions**: below the drafts, an "Existing questions" list shows every already-**approved**/live question for the selected lesson in the same Android-style preview, each with its own **Edit** form — the same `UpdateQuestion` path used for drafts, just without the approve/delete step, so changes to live questions take effect immediately.
- Only multiple-choice questions are supported anywhere in the app (see [Data model](#data-model)) — the form and CSV importer don't offer any other exercise type.



## Data model

Postgres, via `golang-migrate` migrations in `backend/internal/db/migrations/`:

- `users` — profile + gamification state (`total_xp`, `hearts`, `hearts_refill_at`, `current_streak`, `longest_streak`, `last_activity_date`).
- `refresh_tokens` — hashed, rotated, revocable.
- `courses` → `skills` → `lessons` → `exercises` → `exercise_options` — the content tree. `exercises.type` is always `multiple_choice` (the only supported exercise type); `correct_answer` and `exercise_options.is_correct` are never sent to the client. An exercise can have **more than one** option with `is_correct = true` (a word can have multiple valid meanings) — `exercises.correct_answer` stores all of them joined with `" / "` for display in post-lesson results. `exercises.status` is `draft` or `approved` — only `approved` rows are ever served to or graded for the mobile app; `draft` is how admin-authored questions stay invisible until reviewed (see [Admin panel](#admin-panel)).
- `user_lesson_completions` — one row per successful submission, unique on `(user_id, idempotency_key)`.
- `xp_transactions` — an XP ledger entry per completion (currently only reason: `lesson_complete`).

## Mobile app

Expo SDK 54, TypeScript, `mobile/src/`:

```
api/        axios client with single-flight token-refresh-on-401, config, secure token storage
components/ DuoButton, LessonNode, StatBadge — reusable Duolingo-style UI primitives
db/         expo-sqlite: course/lesson cache + offline submission queue
hooks/      useCourse/useLessonDetail (TanStack Query + cache fallback), useSyncPendingSubmissions
navigation/ RootNavigator (auth-gated stack) + MainTabs (Learn/Leaderboard/Profile)
screens/    Login, Register, Learn, Lesson, LessonResults, Leaderboard, Profile
store/      authStore (Zustand) — session state, bootstraps from SecureStore on launch
theme/      shared colors/radii
types/      TypeScript mirror of the backend JSON contracts
```

- **Token storage**: `expo-secure-store` (Keychain/Keystore-backed), never AsyncStorage.
- **Offline-first**: since `GET /lessons/{id}` withholds correct answers by design, the client *cannot* grade a lesson locally, online or offline. So the offline path is "queue the full answer set, defer to a 'saved offline, will sync' results screen" — never a locally-computed score. Queued submissions reuse the same `idempotency_key` generated at lesson start, so a submission that's retried after coming back online can't double-award XP even if the original request actually reached the server before the connection dropped. `useSyncPendingSubmissions` flushes the queue on every connectivity-regained event (`@react-native-community/netinfo`) and on app start.
- **Course/lesson data**: cached to SQLite on every successful fetch; if a fetch fails, the hook falls back to the last cached copy before rethrowing.
- Configure `EXPO_PUBLIC_API_URL` in `mobile/.env` (see `mobile/.env.example`) — a physical device needs your machine's LAN IP (`ipconfig getifaddr en0` on macOS), which changes whenever you switch Wi-Fi networks.

## Security baselines (non-negotiable)

- **Server-side correctness only** — lesson score, "answer correct," XP, streak are always recomputed server-side; the client is never trusted for any of it.
- **Token storage on RN**: `expo-secure-store` only.
- **Parameterized queries only** — no string-built SQL, anywhere.
- **Rate limiting** on auth endpoints, lesson submission, and the admin panel.
- **No real secrets in the RN bundle** — anything shipped in the app binary is extractable.
- **Admin credential** is bcrypt-hashed at rest; sign-in happens through its own `/admin/login` page (not a native HTTP Basic Auth prompt), issuing an `HttpOnly`/`SameSite=Strict` signed session cookie over a separate auth path from user JWTs. Fails closed if unconfigured. No admin registration flow exists anywhere in the app.
- **COPPA/GDPR**: not yet implemented — needs addressing before this app is exposed to under-13 or EU/UK users.
- **Payments**: not yet implemented — when added, route through Stripe/RevenueCat, never touch raw card data.

## Getting started

```bash
make help                          # see every available target

# Backend
cp backend/.env.example backend/.env   # fill in DATABASE_URL, JWT secrets, admin creds
make backend-migrate-up            # apply DB schema
make backend-seed                  # populate the "125 Words of the Qur'an" course
make backend-run                   # start the API on :8080
make backend-admin-hash password=yourpassword   # generate ADMIN_PASSWORD_HASH
make backend-admin-secret          # generate ADMIN_SESSION_SECRET

# Mobile
cp mobile/.env.example mobile/.env     # set EXPO_PUBLIC_API_URL for your setup
make frontend-install
make frontend-start                # scan the QR code in Expo Go

# Both at once
make dev                           # backend in background + Expo dev server in foreground
```

Full env var reference: `backend/.env.example`, `mobile/.env.example`.

## Versioning & documentation process

- **Semantic Versioning** (`MAJOR.MINOR.PATCH`). Current: **v1.3.0**.
- **[CHANGELOG.md](./CHANGELOG.md)** — [Keep a Changelog](https://keepachangelog.com/) format: one dated entry per release, grouped into Added/Changed/Fixed/Removed. This is the fast, diffable answer to "what changed and when."
- **[/change-history](./change-history)** — a full copy of `README.md` saved at every tagged version (`change-history/v1.0.0/README.md`, …). CHANGELOG.md tells you *what* changed; a change-history snapshot lets you see the *entire* app description as it stood at that point, without digging through git history or reconstructing state from diffs.
- **Git tags** (`v1.0.0`, `v1.1.0`, `v1.2.0`, `v1.3.0`, …) mark the commit each snapshot corresponds to.

**Rule for future changes**: any change that affects app behavior, API contracts, gamification rules, or setup steps must update `README.md` and add a `CHANGELOG.md` entry in the same change — not as a follow-up. This is written into `.claude/skills/golang-backend/SKILL.md` and `.claude/skills/react-native-frontend/SKILL.md` so it isn't dependent on remembering to do it by hand.

## Engineering conventions & best practices

- **Modular monolith** — one Go binary, one Postgres database. No microservices until there's a concrete reason to split (per `CLAUDE.md`).
- **Backend and frontend stay separate apps** with independent dependency management, sharing only a root `Makefile`, even though they live in one repo.
- **No ORM, no codegen** for the DB layer — hand-written parameterized SQL is the whole toolchain; one less thing to generate, debug, or version.
- **Server is the only source of truth for game state** — see [Gamification rules](#gamification-rules). This is the single most important rule in the codebase; almost every other backend decision follows from it.
- **Idempotency by construction, not by convention** — the uniqueness that prevents double-awarding XP is a DB constraint (`UNIQUE (user_id, idempotency_key)`), not just application-level checking, so it holds even under concurrent requests.
- **Fail closed on missing config** — the admin panel simply isn't mounted if its credentials aren't set, rather than mounting with an empty/default one.
- **Upsert-based seeding** — `make backend-seed` is safe to run repeatedly; content changes are edits to the seed data, not one-off scripts.
- **No premature abstraction** — e.g. the admin panel reuses `MaxHearts` from the lesson service instead of introducing a shared "settings" package for one constant; three similar lines beat a speculative shared layer.
- **Every behavior-affecting change ships with its documentation update** — see [Versioning & documentation process](#versioning--documentation-process).
