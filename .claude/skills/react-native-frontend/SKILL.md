---
name: react-native-frontend
description: Conventions and checklist for building/reviewing the QuranLingo React Native (Expo) frontend — offline-first data layer, state management, auth token storage, navigation, and security baselines for a Duolingo-style gamified learning app. Use whenever writing, reviewing, or planning React Native/mobile code in this project.
---

# React Native frontend conventions — QuranLingo

Lives in `mobile/` at the repo root, alongside the Go backend in `backend/`. Run it via the root `Makefile` (`make frontend-start`, `make frontend-ios`, `make frontend-android`) rather than invoking `npx expo` directly, so command entrypoints stay consistent with the backend.

## Project layout

- Expo managed workflow, TypeScript throughout — only eject to a custom dev client if a native module truly requires it.
- Structure: `app/` or `src/` with `screens/`, `components/`, `navigation/`, `api/` (typed client), `db/` (local storage layer), `store/` (client state).
- Screens stay thin: fetch/derive data via hooks, render via components. Keep business logic (SRS scheduling, XP display rules) out of screen components.

## State management

- **Server state**: TanStack Query (React Query) for all API data — caching, retries, background refetch. Don't hand-roll fetch/cache logic.
- **Local/UI state**: React Context or Zustand for lightweight global state (current lesson session, UI toggles). Avoid Redux boilerplate unless the app's state complexity actually demands it.

## Offline-first data layer

- This is a core design constraint, not an add-on: learners use the app on commutes/low signal.
- Cache lesson/course/exercise content locally with `expo-sqlite` (or WatermelonDB if sync complexity grows) so lessons are usable offline.
- Queue progress events (lesson completed, exercise answered, streak tick) locally when offline; sync to the backend when connectivity returns.
- Conflict resolution: keep it simple — server is the source of truth for XP/streak/progress; client-queued events are replayed and the server recomputes canonical state (matches the backend's "never trust client for correctness" rule). Don't try to merge conflicting client/server state on-device.
- Every queued sync event needs a client-generated idempotency key so a retried sync can't double-award XP.

## Networking

- Typed API client generated from (or kept in sync with) the backend's OpenAPI spec — don't hand-write duplicate request/response types on both sides.
- Central place for auth header injection, 401 → refresh-token flow, and retry/backoff — not scattered per-screen fetch calls.

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
