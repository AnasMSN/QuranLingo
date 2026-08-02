---
name: golang-backend
description: Conventions and checklist for building/reviewing the QuranLingo Go backend — routing, PostgreSQL access, auth, background jobs, and security baselines for a Duolingo-style gamified learning API. Use whenever writing, reviewing, or planning Go backend code in this project.
---

# Go backend conventions — QuranLingo

## Documentation rule (read this first)

Any change that affects app behavior, API contracts, gamification rules, admin capabilities, or setup steps **must update `README.md` and add a `CHANGELOG.md` entry in the same change** — not as a follow-up, not left for the user to remember. See the root `README.md`'s "Versioning & documentation process" section for the format (Keep a Changelog + SemVer) and the `/change-history` snapshot convention. This applies whether the change was requested by name or discovered incidentally while doing something else.

## Project layout (as implemented)

- `cmd/api` — HTTP server entrypoint. `cmd/seed` — populates course content (`make backend-seed`). `cmd/hashpassword` — prints a bcrypt hash for `ADMIN_PASSWORD_HASH` (`make backend-admin-hash password=...`).
- `internal/config` — env loading (`godotenv` reads `backend/.env`, falls back to real env vars).
- `internal/db` — `pgxpool` connection setup; `internal/db/migrations` — `golang-migrate` SQL files; `internal/db/seed` — course/skill/lesson/exercise seed data.
- `internal/models` — plain structs, no ORM tags.
- `internal/repository` — every SQL statement in the app, as free functions taking `(ctx, q Querier, ...)`. `Querier` is a small interface (`Exec`/`Query`/`QueryRow`) implemented by both `*pgxpool.Pool` and `pgx.Tx`, so the same function works standalone or inside a transaction.
- `internal/service` — business logic (`auth_service.go`, `content_service.go`, `lesson_service.go`, `leaderboard_service.go`). Handlers never touch the repository directly except for simple lookups (e.g. `GET /me`).
- `internal/handler` — thin HTTP layer + `router.go` (chi). `internal/middleware` — JWT auth middleware. `internal/httpx` — shared JSON response/error helpers.
- Handlers stay thin: parse/validate request → call a service function → shape response.

## Database (PostgreSQL)

- **Deviation from the original plan**: hand-written parameterized SQL via `pgx/v5` in `internal/repository`, not `sqlc`/`gorm`. Reasoning: same safety guarantees (fully parameterized, no string concat), zero codegen toolchain to install/run/keep in sync. Revisit `sqlc` only if the repository layer grows large enough that hand-maintaining scan/insert boilerplate becomes the bottleneck.
- Every schema change goes through a `golang-migrate` migration file (`internal/db/migrations/00000N_name.{up,down}.sql`), checked into version control, never applied by hand against prod.
- Wrap multi-step state changes in a single DB transaction — see `LessonService.Submit` for the reference pattern (lesson completion + XP transaction + user XP/hearts/streak update, all committed together, with idempotency-key unique-constraint handling for concurrent retries).
- Implemented schema: `users`, `refresh_tokens`, `courses`, `skills`, `lessons`, `exercises`, `exercise_options`, `user_lesson_completions`, `xp_transactions`. There is deliberately no separate `user_skill_progress`/streak table — lesson-unlock status and skill-completion are derived at read time from `user_lesson_completions` (see `ContentService.GetCourseTree`), and streak/hearts/XP live directly on `users` since they're always updated alongside a completion anyway. Revisit only if derived-status queries become a performance problem at scale.
- **Known gotcha**: Supabase's direct `db.<ref>.supabase.co` host is IPv6-only on many projects — unreachable from IPv4-only networks/sandboxes. Use the connection pooler host (Session or Transaction mode, from Project Settings → Database → Connection string) for `DATABASE_URL` if you hit `no route to host` on an IPv6 address.

## API design

- REST + JSON, documented with an OpenAPI spec. Only reach for GraphQL if a concrete screen needs many differently-shaped partial views.
- Validate all input at the handler boundary (types, ranges, required fields) before it reaches a service.
- Idempotency: any endpoint that can be double-submitted by a flaky mobile connection (lesson-complete, purchase confirmation) must be safe to retry — use an idempotency key or check-then-no-op logic.

## Auth

- Implemented in `internal/service/auth_service.go`: JWT access token (HS256, short-lived, `golang-jwt/jwt/v5`) + opaque refresh token (32 random bytes, hex-encoded, SHA-256 hashed before storage, rotated and revoked on every use).
- Passwords hashed with `bcrypt` (`golang.org/x/crypto/bcrypt`). Never implement custom crypto/hashing.
- `internal/middleware/auth.go` validates the `Authorization: Bearer` header and injects the user ID into request context (`middleware.UserID(ctx)`).

## Admin panel

- Operator-only, server-rendered HTML at `/admin` (`internal/handler/admin_handler.go`, templates embedded from `internal/handler/templates/admin.html` + `admin_login.html` via `//go:embed`) — not a JSON API, so it intentionally bypasses `internal/httpx`.
- **Auth is a dedicated login page, not HTTP Basic Auth.** `GET/POST /admin/login` renders a form and checks the submission with `middleware.VerifyAdminCredentials` (bcrypt-compared against `ADMIN_USERNAME`/`ADMIN_PASSWORD_HASH`), then issues a signed session cookie via `middleware.NewAdminSessionCookie` (`ADMIN_SESSION_SECRET`, HMAC-SHA256 over `username|expiry`, 12h TTL, `HttpOnly`+`SameSite=Strict`, `Secure` only when `cfg.Env == "production"`). `middleware.AdminSessionAuth` gates every other `/admin/*` route and redirects to `/admin/login` (not a 401) on missing/invalid/expired cookies. `POST /admin/logout` clears the cookie.
- **There is no admin registration flow anywhere in the codebase** — the only admin account that can ever exist is whatever `ADMIN_USERNAME`/`ADMIN_PASSWORD_HASH` resolve to in the environment. Never add a signup path for admins; if multi-operator support is needed later, that's a real `is_admin` DB-role redesign, not an extension of this env-var credential.
- Routes are only mounted (`router.go`) if `ADMIN_USERNAME`, `ADMIN_PASSWORD_HASH`, **and** `ADMIN_SESSION_SECRET` are all set — fails closed rather than mounting with an empty/guessable credential or session key. Keep this pattern for any future admin-only capability.
- `ADMIN_PASSWORD_HASH` contains `$` — always single-quote it in `.env`, or a shell `source` (as `backend-migrate-up`/`down` do) will expand it as positional parameters. Generate it with `make backend-admin-hash password=...` and the session secret with `make backend-admin-secret`.
- Reuse constants from `service` (e.g. `service.MaxHearts`) instead of redefining them in admin code.
- **Question management** (`/admin/questions`, same handler/templates package): lets an operator add multiple-choice questions to any lesson via a manual form or bulk CSV import, without touching the seeder. Every question created here starts as `exercises.status = 'draft'` — `AdminService.ApproveQuestion` is the only path that flips it to `'approved'`, and only `'approved'` rows are ever visible to `ContentService.GetLessonDetail` / graded by `LessonService.Submit` (both filter on status in `repository.ListExercisesByLesson`). The preview page (`templates/admin_questions.html`) renders each question (draft or already-live) styled like `mobile/src/screens/LessonScreen.tsx` (same colors/layout), via a shared `{{define "questionCard"}}` partial, so the operator proofs it exactly as a learner would see it, with every correct option flagged for the operator's eyes only. Drafts can additionally be approved (`ApproveQuestion`) or deleted (`DeleteDraftQuestion`, draft-only — never approved/live questions). `AdminService.UpdateQuestion` edits either a draft or an already-live question (live edits apply immediately, no re-approval).
- **Per-question audio** (`exercises.audio_url`, migration `000005`): a plain public URL to a pronunciation clip, stored and returned as-is — the backend never validates, fetches, or proxies it (audio is served directly from wherever it's hosted, e.g. a public Supabase Storage bucket; see `AUDIO_BASE_URL` guidance in `.env.example`). It's optional everywhere: `CreateDraftQuestion`/`UpdateQuestion`/CSV import (`audio_url` column) all accept a blank value, and `ContentService.ExerciseDTO` uses `json:"audio_url,omitempty"` so questions with no clip just omit the field rather than sending `""`. A blank/broken URL is an intentionally unremarkable state — nothing server-side treats it as an error.
- **Questions can have more than one correct option** — Arabic words often have a general sense plus one or more specific ones. `validateQuestionOptions` requires 2-4 options with *at least* one marked correct (not exactly one); `gradeAnswer` in `lesson_service.go` already handles this correctly unchanged, since it only checks whether the learner's selected option has `is_correct = true`. The admin UI uses checkboxes (`correct_options` form field, multi-value) rather than a radio button; CSV's `correct_option` column accepts `;`-separated option numbers (`parseCorrectOptionNumbers` in `admin_service.go`). `correctAnswerText` joins every correct option's text with `" / "` into `exercises.correct_answer` for the post-lesson results screen.
- **Multiple-choice is the only exercise type** — `models.ExerciseType` has a single value (`ExerciseMultipleChoice`); the DB's `exercises_type_check` constraint enforces it too. Never reintroduce a free-text/"translate" exercise type without also updating the mobile lesson screen, which no longer has a text-input rendering path.

## Background jobs / async

- **Not yet implemented.** XP/hearts/streak updates currently happen synchronously inside the lesson-submission request, inside one DB transaction — correct and simple at MVP scale.
- When it's time to add async work (streak-reset reminders, push notifications, leaderboard precomputation), use `asynq` (Redis-backed). Jobs must be idempotent — a redelivered job should not double-award XP or double-send a notification.

## Security checklist (apply to every PR touching the backend)

- [ ] No trust in client-submitted scores/XP/streak values — server recomputes and validates.
- [ ] All queries parameterized; no string-built SQL.
- [ ] Sensitive endpoints (progress/XP writes, content-listing) rate-limited. Implemented with `go-chi/httprate`, keyed by `RemoteAddr` (`httprate.LimitByIP`, not `LimitByRealIP`) — the real-IP variant trusts `X-Forwarded-For`/`X-Real-IP`, which is spoofable by any client unless a trusted reverse proxy is confirmed to be setting those headers. Switch to a real-IP-aware key only once a specific proxy/load balancer is in front of this API and its trust boundary is known.
- [ ] No secrets logged or returned in API responses/error messages.
- [ ] New DB migrations reviewed for backward compatibility (won't break the currently-deployed API version during rollout).
- [ ] Payment flows never touch raw card data — delegate to Stripe/RevenueCat.

## Testing

- Service-layer unit tests with a real Postgres (via `testcontainers-go` or a dockerized test DB) rather than mocking the DB — integration behavior (constraints, transactions) matters more than isolation here.
