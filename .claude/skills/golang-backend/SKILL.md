---
name: golang-backend
description: Conventions and checklist for building/reviewing the QuranLingo Go backend — routing, PostgreSQL access, auth, background jobs, and security baselines for a Duolingo-style gamified learning API. Use whenever writing, reviewing, or planning Go backend code in this project.
---

# Go backend conventions — QuranLingo

## Project layout (as implemented)

- `cmd/api` — HTTP server entrypoint. `cmd/seed` — populates starter course content (`make backend-seed`).
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
