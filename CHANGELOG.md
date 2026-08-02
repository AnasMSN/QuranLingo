# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versioning follows [Semantic Versioning](https://semver.org/).

A full snapshot of `README.md` as of each version below is kept in [`/change-history`](./change-history).

## [1.4.0] — 2026-08-02

### Added
- **Per-question audio**: `exercises.audio_url` (migration `000005`) stores an optional plain public URL to a pronunciation clip — the backend never uploads, validates, or proxies it, just stores/returns it as-is. `AUDIO_BASE_URL` documented in `backend/.env.example` as the simplest hosting option (a public Supabase Storage bucket).
- **Admin**: an "Audio URL" field on the manual add/edit question forms, an optional `audio_url` CSV import column, and a **🔊 Play sound** button in the question preview (a plain `<audio>` element whose `play()` errors are swallowed silently).
- **Mobile**: `expo-audio` (`mobile/src/utils/sound.ts`'s `playAudioUrl`) auto-plays a question's audio every time it's shown, plus a **🔊 Hear again** button to replay it. A missing/unreachable clip fails completely silently on both admin and mobile — this is the normal state for content with no recording yet, not an error.

## [1.3.0] — 2026-08-01

### Added
- **Edit existing (live) questions**: `/admin/questions` now lists every already-**approved** question for the selected lesson, in the same Android-lesson-screen-styled preview as drafts, each with its own inline **Edit** form. Edits to a live question take effect immediately (`AdminService.UpdateQuestion`, no re-approval step) — no more editing content only by re-running the seeder.

### Changed
- **Multiple correct answers per question**: since Arabic words often carry both a general and one or more specific valid meanings, a multiple-choice question is no longer limited to exactly one correct option. Admins check as many options as are acceptable (checkboxes, not a radio button) in both the manual form and the edit form; CSV's `correct_option` column now accepts `;`-separated option numbers (e.g. `1;3`). The learner is graded correct if they pick *any* option flagged correct (`gradeAnswer` didn't need to change — it already just checked the selected option's `is_correct`). `exercises.correct_answer` now stores every correct option's text joined with `" / "` for the post-lesson results screen.
- `AdminService.UpdateDraftQuestion` renamed to `UpdateQuestion` and its "draft only" restriction was dropped, since it's now also the path used to edit live questions.

## [1.2.0] — 2026-08-01

### Added
- **Admin question management** (`/admin/questions`): operators can add multiple-choice questions to any lesson without touching the seed data — one at a time via a form, or in bulk via CSV import (UTF-8, so Arabic with harakat pastes in as-is). Every new question starts as a **draft**, invisible to the app, and is rendered in an Android-lesson-screen-styled preview (matching `mobile/src/screens/LessonScreen.tsx`) with the correct answer highlighted for the operator only. Drafts can be edited or deleted before an explicit **Approve** click publishes them.
- `exercises.status` column (`draft`/`approved`, migration `000004`) — `ContentService`/`LessonService` only ever see `approved` rows; draft questions can't leak to the mobile app or be graded before review.

### Changed
- **Multiple-choice only**: removed the `translate` (type-the-answer) exercise type end-to-end — DB constraint (`exercises_type_check`), Go models/grading (`gradeAnswer` simplified, `AnswerInput.text_answer` removed), the course seeder (now always emits multiple-choice), and the mobile lesson screen (dropped the text-input branch, always renders options). Legacy `translate` rows in the DB were converted to `multiple_choice` by the migration; `make backend-seed` backfills their options.

## [1.1.0] — 2026-07-31

### Changed
- **Admin panel auth**: replaced the browser-native HTTP Basic Auth prompt with a dedicated `/admin/login` sign-in page and a signed, `HttpOnly`/`SameSite=Strict` session cookie (`ADMIN_SESSION_SECRET`, HMAC-SHA256, 12h expiry). `/admin/logout` clears the session. The admin panel now requires all three of `ADMIN_USERNAME`, `ADMIN_PASSWORD_HASH`, and `ADMIN_SESSION_SECRET` to be set (previously just the first two) — still fails closed if any is missing. There remains no admin registration flow of any kind; the only admin account is whatever these env vars resolve to.

### Added
- `make backend-admin-secret` — generates a random 64-char hex secret for `ADMIN_SESSION_SECRET`.

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
