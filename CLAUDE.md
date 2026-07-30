# QuranLingo

A Duolingo-style language-learning app. Backend and frontend are two separate apps living in **one monorepo**.

## Repository layout

- `backend/` — Go API (module root, `cmd/api` entrypoint).
- `mobile/` — React Native (Expo) app.
- Root `Makefile` — the single entrypoint for running/building/testing either app. Always use `make <target>` (`make backend-run`, `make frontend-start`, `make dev`, etc. — run `make help` for the full list) instead of `cd`-ing into a subfolder and invoking `go`/`npm`/`expo` directly, so command usage stays consistent for anyone (or any agent) working in this repo.

## Stack

- **Backend**: Go, **chi** router (`internal/handler/router.go`). PostgreSQL (Supabase-hosted) is the source of truth. Redis/`asynq` background jobs are documented as a future addition — not yet wired up; XP/streak/hearts updates currently happen synchronously inside the request that triggers them, which is fine at MVP scale.
- **DB access**: hand-written parameterized SQL via `pgx/v5` in `internal/repository/` (a deliberate deviation from the originally-planned `sqlc` — same safety guarantees, one less codegen toolchain to run). `golang-migrate` for schema migrations (`backend/internal/db/migrations`).
- **Object storage**: S3-compatible bucket for audio/image lesson assets — not yet implemented (no audio content in the seeded course yet). Never serve binary assets through the API process or store them in Postgres.
- **Frontend**: React Native (Expo unless there's a specific native-module reason not to), TypeScript. Not yet scaffolded.
- **Offline**: local SQLite / WatermelonDB on the client caching lesson content and queuing progress events for sync — this is a core design constraint, not an afterthought, since language learners use the app offline (commute, low signal).
- **API shape**: REST + JSON. Implemented endpoints: `POST /auth/{register,login,refresh}`, `GET /me`, `GET /courses/{code}`, `GET /lessons/{id}`, `POST /lessons/{id}/submit`, `GET /leaderboard/weekly`, `GET /healthz`.
- **Auth**: JWT access token (short-lived, HS256) + opaque refresh token (rotated and revoked on use, hashed at rest) — implemented in `internal/service/auth_service.go`. Passwords hashed with bcrypt.
- **Gamification core** (`internal/service/lesson_service.go`): lesson submission grades answers server-side only, awards XP, decrements hearts on wrong answers (with a 4h lazy refill once hearts hit 0), and updates the daily streak — all in one DB transaction. Submissions require a client-generated `idempotency_key` so retried requests can't double-award XP.
- **Seed content**: `backend/internal/db/seed` + `make backend-seed` populates the "Arabic Basics" course — 5 skills (Greetings, Numbers, Family, Common Phrases, Food & Drink), one lesson each, 5 exercises per lesson (alternating multiple-choice/translate). Upsert-based, safe to re-run.

## Security baselines (non-negotiable)

- **Server-side correctness only.** Lesson score, "answer correct," XP, streak increments must be recomputed and validated server-side. Never trust client-submitted scores — this is the #1 cheat vector in gamified apps.
- **Token storage on RN**: `expo-secure-store` (Keychain/Keystore-backed) — never AsyncStorage. Only use `react-native-keychain` directly if the app moves to a bare/custom dev client workflow.
- **Parameterized queries only** — no string-built SQL, ever.
- **Rate-limit** progress/XP endpoints and any endpoint that enumerates course/lesson content.
- **No real secrets in the RN bundle.** Anything shipped in the app binary is extractable; client-side keys must be public-safe (publishable keys only).
- **COPPA/GDPR**: if users may be under 13 or in the EU/UK, bake in data-handling/consent rules from day one — far cheaper than retrofitting.
- **Payments**: route through Stripe/RevenueCat if/when subscriptions are added. Never touch raw card data (stay out of PCI scope).

## Conventions

- Prefer a modular monolith for the backend until there's a concrete reason to split services — don't pre-architect microservices.
- Keep the Go backend and React Native frontend as clearly separated apps/directories with their own dependency management, even if they end up in one repo.
