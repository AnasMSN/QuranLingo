---
name: react-native-frontend
description: Conventions and checklist for building/reviewing the QuranLingo React Native (Expo) frontend — offline-first data layer, state management, auth token storage, navigation, and security baselines for a Duolingo-style gamified learning app. Use whenever writing, reviewing, or planning React Native/mobile code in this project.
---

# React Native frontend conventions — QuranLingo

Lives in `mobile/` at the repo root, alongside the Go backend in `backend/`. Run it via the root `Makefile` (`make frontend-start`, `make frontend-ios`, `make frontend-android`) rather than invoking `npx expo` directly, so command entrypoints stay consistent with the backend.

## Documentation rule (read this first)

Any change that affects app behavior, screens/navigation, offline sync, or setup steps **must update `README.md` and add a `CHANGELOG.md` entry in the same change** — not as a follow-up, not left for the user to remember. See the root `README.md`'s "Versioning & documentation process" section for the format (Keep a Changelog + SemVer) and the `/change-history` snapshot convention. This applies whether the change was requested by name or discovered incidentally while doing something else.

## Project layout (as implemented)

- Expo managed workflow (currently pinned to **SDK 54** for Expo Go compatibility — check before bumping), TypeScript throughout. Only eject to a custom dev client if a native module truly requires it.
- `mobile/src/`: `api/` (axios client + config + secure token storage), `components/` (DuoButton, LessonNode, StatBadge), `db/` (expo-sqlite cache + offline submission queue), `hooks/` (TanStack Query hooks), `navigation/` (RootNavigator + MainTabs), `screens/`, `store/` (Zustand `authStore`), `theme/`, `types/` (hand-written mirror of the backend JSON contracts — see Networking below), `utils/`.
- Screens stay thin: fetch/derive data via hooks, render via components. Keep business logic (lesson-flow state, XP display rules) out of screen components.
- **Hooks-order discipline**: every hook call (`useState`, `useMemo`, etc.) must run unconditionally on every render — never after an early `return`. We've hit "Rendered more hooks than during the previous render" from this exact mistake (a `useMemo` placed after a loading-state early return in `LearnScreen.tsx`); `expo lint`'s `react-hooks/immutability` rule catches the sibling mistake (mutating state during render) but not hook-ordering violations after an early return, so review this by eye.

## State management

- **Server state**: TanStack Query (React Query) for all API data — caching, retries, background refetch. Don't hand-roll fetch/cache logic.
- **Local/UI state**: React Context or Zustand for lightweight global state (current lesson session, UI toggles). Avoid Redux boilerplate unless the app's state complexity actually demands it.

## Offline-first data layer

- This is a core design constraint, not an add-on: learners use the app on commutes/low signal.
- Implemented in `mobile/src/db/`: `database.ts` (expo-sqlite setup), `courseCache.ts` (course/lesson content cache, read on fetch failure), `submissionQueue.ts` (queued lesson submissions). `useSyncPendingSubmissions` flushes the queue on every `NetInfo` connectivity-regained event and on app start.
- Because `GET /lessons/{id}` withholds correct answers by backend design, the client **cannot** grade a lesson locally under any circumstance — the offline path is "queue the full answer set, show a 'saved offline, will sync' result," never a locally-computed score. Don't build a local grading fallback; it would contradict the one rule the whole gamification system depends on.
- Conflict resolution: keep it simple — server is the source of truth for XP/streak/progress; client-queued events are replayed and the server recomputes canonical state. Don't try to merge conflicting client/server state on-device.
- Every queued sync event reuses the same client-generated `idempotency_key` created when the lesson attempt started, so a retried/replayed sync can't double-award XP even if the original request actually reached the server before the connection dropped.

## Networking

- `mobile/src/api/client.ts`: two axios instances — `http` (request interceptor injects the bearer token; response interceptor does single-flight refresh-on-401) and `rawHttp` (interceptor-free, used only for the `/auth/refresh` call itself, to avoid infinite recursion). `setOnAuthFailure` decouples the client from `store/authStore.ts` to avoid a circular import.
- `mobile/src/types/api.ts` is a hand-written TypeScript mirror of the backend JSON contracts (there's no OpenAPI spec to generate from) — when a backend response shape changes, update this file in the same change.

## Auth & secure storage

- Store access/refresh tokens with `expo-secure-store` (Expo managed workflow — backed by Keychain on iOS / Keystore on Android). Only reach for `react-native-keychain` directly if the project moves to a bare/custom dev client workflow.
- **Never** store tokens in `AsyncStorage`, plain state, or anywhere unencrypted.
- Handle refresh-token rotation transparently in the API client; on unrecoverable auth failure, force a clean logout (clear local queue/cache of anything sensitive).

## Navigation

- React Navigation (stack + tab navigators). Keep deep-link routes validated server-side reachable only for authorized content (don't assume a deep link implies entitlement).

## Performance

- Virtualize long lists (course/skill trees, leaderboards) with `FlashList` instead of a raw `ScrollView`/`FlatList` for large datasets.
- Lazy-load lesson audio/images from the CDN/object storage URLs the backend returns — don't bundle content assets into the app binary.

## Security checklist (apply to every PR touching the frontend)

- [ ] Tokens stored via `expo-secure-store`/Keychain/Keystore only — never `AsyncStorage` or plain JS state persisted to disk.
- [ ] No real secrets/API keys embedded in the bundle — only public-safe (publishable) keys, since anything shipped in the app is extractable.
- [ ] All server writes (lesson-complete, purchase) carry an idempotency key so retries/offline-sync can't double-apply.
- [ ] No client-side computation is trusted as final for XP/streak/correctness — it's optimistic UI only, reconciled against the server response.
- [ ] Deep links / universal links validate authorization server-side before rendering gated content.

## Testing

- Jest + React Native Testing Library for component/unit tests.
- Consider Maestro (simpler YAML flows) over Detox for E2E unless the team already knows Detox.
