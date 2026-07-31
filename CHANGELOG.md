# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versioning follows [Semantic Versioning](https://semver.org/).

A full snapshot of `README.md` as of each version below is kept in [`/change-history`](./change-history).

## [1.0.0] — 2026-07-31

Initial tagged release. Backend and mobile app both functional end-to-end against a hosted Supabase Postgres instance.

### Added
- **Backend (Go)**: chi-routed REST API — auth (register/login/refresh with rotated, hashed refresh tokens), course/lesson content endpoints with server-computed lesson/skill locking, lesson submission with server-side grading, XP, hearts, streaks, and idempotency-key-protected double-submission prevention, weekly leaderboard.
- **Database**: PostgreSQL schema via `golang-migrate` (`users`, `refresh_tokens`, `courses`/`skills`/`lessons`/`exercises`/`exercise_options`, `user_lesson_completions`, `xp_transactions`).
- **Course content**: "125 Words of the Qur'an" — 25 skills × 5 words, fully vocalized with harakat, sourced from Dr. Abdulazeez Abdulraheem's *125 Words of the Qur'an* (Understand Al-Qur'an Academy). Upsert-based seeder (`make backend-seed`).
- **Mobile app (Expo/React Native, TypeScript)**: Duolingo-style skill-tree UI, full auth flow, lesson-taking flow, leaderboard, profile. Offline-first: SQLite content cache + a submission queue that syncs on reconnect, reusing the same idempotency key so replays can't double-award XP.
- **Admin panel**: server-rendered HTML page at `/admin` (HTTP Basic Auth, bcrypt-hashed operator credential) to list users and refill hearts — bulk "Refill All Hearts" and per-user refill — without waiting out the normal 4-hour lazy refill.
- **Dev tooling**: `make backend-admin-hash` to generate the admin password hash; root `Makefile` as the single entrypoint for both apps.
- **Documentation**: comprehensive root `README.md`, this changelog, and `/change-history` version snapshots.

### Changed
- Course content replaced the original placeholder "Arabic Basics" course (5 skills, everyday vocabulary) with "125 Words of the Qur'an" — course code changed from `arabic-basics` to `quran-125` (updated in both the seeder and the mobile app's course query).
- Expo SDK pinned to 54 for compatibility with the current Expo Go client app.

### Fixed
- `LearnScreen.tsx`: a `useMemo` hook was called after an early `return` (loading state), causing React's "Rendered more hooks than during the previous render" error — moved above all early returns so hooks stay unconditional.
- Mobile `.env` pointed at `localhost`, which resolves to the phone itself on a physical device rather than the dev machine — switched to the machine's LAN IP.
- `cmd/seed`'s hardcoded 30s context timeout was sized for the old 25-word course; raised to 5 minutes for the 125-word course's ~400+ sequential upserts over the network.
