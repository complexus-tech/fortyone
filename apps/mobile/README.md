# FortyOne Mobile

FortyOne's Expo application for iOS and Android. The current baseline is Expo SDK 57, React Native 0.86, and React 19.2. It uses development builds with native modules; Expo Go is not a supported way to validate this app. Dependency versions are recorded in `package.json` and the root `pnpm-lock.yaml`.

This codebase is being prepared for its first store release. Passing TypeScript, tests, and Metro export checks does not establish that authentication, native rendering, signing, or store delivery work on physical devices.

## Local development

Use Node.js 22.13 or newer, pnpm 9.3.0, and the platform toolchain required by the installed Expo SDK. iOS compilation requires macOS and Xcode; Android compilation requires the Android SDK and a compatible JDK.

From the repository root:

```sh
pnpm --filter mobile... install --frozen-lockfile
cp apps/mobile/.env.example apps/mobile/.env.local
```

Set the public API and web application URLs to the team's development environment. The web app and API must both include the mobile sign-in endpoints in this repository. Public Expo variables are compiled into the app and must never contain credentials. HTTPS is required outside development; development HTTP is restricted to localhost addresses by `lib/http/config.ts`. A physical device needs a reachable, trusted HTTPS development environment.

```sh
pnpm --filter mobile start
SENTRY_DISABLE_AUTO_UPLOAD=true pnpm --filter mobile ios
SENTRY_DISABLE_AUTO_UPLOAD=true pnpm --filter mobile android
```

The native commands generate/build local native projects and require working platform tooling. `web` starts a browser development target, but native authentication and platform behavior must be tested in the development client. Generated `ios/`, `android/`, `.expo/`, and `dist/` output are not source-controlled.

## Implemented behavior

- Home overview, joined teams, My Work, grouped story lists, search, objectives, sprints, and story details.
- Story creation and updates, properties, labels, archive/delete/restore actions, links, subtasks, and comments through the existing API.
- Paginated in-app inbox with read/unread/delete actions and native story/objective destinations. Unsupported native destinations open the canonical FortyOne web page. This is an inbox, not operating-system push delivery.
- Shared HTML rendering and editing through TipTap in Expo DOM components/WebView. Existing descriptions prefer `descriptionHTML` with escaped plain-text fallback. Editors send HTML plus plain text, retain supported web document structures, sanitize content, and preserve local drafts by account, workspace, and document. Drafts are cleared on sign-out/account replacement. Rendering existing media does not implement camera capture or attachment upload.
- Explicit loading, retry, empty, and mutation failure feedback. Grouped stories use a single virtualized SectionList with per-group pagination.
- Protected native routes, scoped request cancellation and persistence, theme preferences, and platform-specific controls.

There is no implemented biometric sign-in, push token registration, background sync service, or offline mutation queue. The presence of `expo-updates` does not establish an operational OTA release pipeline. NativeWind remains on the v5 preview line because the app uses Tailwind 4; its native styling needs device regression testing before release. The root pnpm override pins only `@expo/metro-config > lightningcss` to 1.30.1 for a reproduced native CSS AST visitor regression ([upstream issue](https://github.com/parcel-bundler/lightningcss/issues/1081)). It does not change the web Tailwind pipeline. Keep the narrow pin until an upstream fix passes native bundle and device checks.

## Authentication and data lifetime

Sign-in opens the FortyOne web app in the system authentication browser. The mobile app creates a random state and a PKCE verifier/challenge and keeps the pending transaction in SecureStore. After web authentication, `/auth/mobile` obtains a short-lived, single-use code from the API and returns to `fortyone://login`. The native app validates the callback and exchanges the code with its verifier at `auth/mobile/exchange`.

The API creates a separate native session using its existing `fortyone_session` cookie. The mobile client stores only that validated first-party cookie and its expiry/account metadata in Expo SecureStore, and sends it explicitly through `expo/fetch` to the configured API origin. It does not depend on sharing the browser cookie jar with React Native. Write requests include the configured web origin required by the backend. Session state is checked on startup; confirmed unauthorized/expired sessions are removed, while a connection failure preserves the existing session and shows retry feedback.

React Query keys and MMKV snapshots include the account and workspace. Workspace/account changes replace the query client and cancel old requests. Logout destroys the account's persisted snapshots and local drafts; late callbacks cannot repersist them. Cached reads can be shown offline when an existing unexpired session is available. Writes require connectivity and fail visibly; they are not queued for background replay.

SecureStore contains the session credential. Query snapshots use MMKV and editor drafts use AsyncStorage, without application-level encryption. Release privacy/backup settings must reflect the fact that project content can be stored on the device. Server revocation can fail while offline; local sign-out still removes the device credential and reports that remote revocation was not confirmed.

## Code organization

```text
app/                   Expo Router routes, protected layout, error recovery
components/ui/         Shared native primitives and platform variants
components/rich-text/  DOM editor/viewer, HTML policy, draft repository
modules/               Feature screens, hooks, queries, actions, and DTOs
lib/http/              Trusted-origin transport, cancellation, HTTP errors
lib/auth*              Native sign-in protocol and cookie validation
lib/session-*          Credential storage and query-cache lifecycle
lib/query-*            Account/workspace keys and persistence provider
lib/observability*     Conditional crash reporting and privacy filter
store/                 Zustand session and local UI state
constants/             Scoped query factories, colors, and terminology
```

Keep network DTO validation at the query/action boundary. Reuse scoped key factories and shape-aware story cache helpers rather than assuming every cached list has the same structure. Cross-account transitions must reset persistence before old asynchronous work can affect the next session. Rich-text format changes need round-trip tests against the shared editor schema.

## Checks and CI

```sh
pnpm --filter mobile type-check
pnpm --filter mobile lint
pnpm --filter mobile test
EXPO_NO_DOTENV=1 EXPO_PUBLIC_API_URL=https://api.example.invalid EXPO_PUBLIC_APP_URL=https://app.example.invalid SENTRY_DISABLE_AUTO_UPLOAD=true pnpm --filter mobile export:verify
pnpm --filter mobile doctor
```

`test` discovers local `*.test.ts` files and uses Node's test runner with `tsx`. Tests cover auth/session contracts, storage ordering, cache isolation/cancellation, optimistic changes, notification pagination/destinations, and editor HTML/draft behavior. They do not contact a backend or monitoring service.

`.github/workflows/mobile-quality.yml` installs the frozen mobile dependency graph, then runs types, lint, tests, and iOS/Android Metro exports. It uses public placeholder origins, has no publishing credentials, and disables Sentry uploads. Exporting bundles does not compile/sign native apps or run device UI tests.

## Crash reporting

Sentry is configured manually through `app.config.ts`, `metro.config.js`, and `lib/observability.ts`. Runtime reporting initializes only in a non-development build with `EXPO_PUBLIC_SENTRY_DSN` set. The root route error boundary provides a recovery screen independently of auth, theme, and query providers.

The JavaScript event filter drops messages, request bodies/headers, user fields, arbitrary extra context, breadcrumbs, and source snippets; it preserves stack locations and selected build/device metadata. Screenshots, view hierarchy, failed-request capture, logs, automatic sessions, performance tracing, profiling, and replay are disabled or unconfigured. Native crash processing has its own SDK path and requires release validation of its actual payload; a JavaScript unit test is not proof of native privacy behavior.

The managed release owner configures `SENTRY_ORG`, `SENTRY_PROJECT`, and a secret `SENTRY_AUTH_TOKEN` in the build environment for source-map and native-symbol upload. Never prefix the token with `EXPO_PUBLIC_`, commit it, or include it in app config. Local/CI builds set `SENTRY_DISABLE_AUTO_UPLOAD=true`. No Sentry account setup or diagnostic event submission is part of the automated tests. See the [official Expo Sentry guide](https://docs.expo.dev/guides/using-sentry/) for the managed build integration.

## Managed mobile delivery

Production builds, signing credentials, store submissions, and over-the-air updates are owned by the internal mobile release process. Do not run an ad hoc production EAS build or store submission from a personal Expo account.

Use the checked-in development profile only after receiving access through the team's managed Expo organization. Changes to `eas.json`, native identifiers, entitlements, signing, or release channels require review from the mobile release owner.

Before the first release, the release owner still needs evidence for:

- Coordinated API/web rollout and real iOS/Android sign-in, cancellation, cold-start callbacks, expiry/revocation, account changes, and offline sign-out.
- Native builds on the supported Xcode/Android toolchains; physical-device navigation, keyboard/editor behavior, low-memory recovery, accessibility, and large-list performance. Include Android edge-to-edge/predictive back and iOS sheet/native-control behavior.
- Supported HTML fixtures round-tripped between web and mobile, including mentions, media, tables, draft recovery, server rejection, and concurrent edits. Confirm attachment/unsupported content expectations before expanding editor features.
- Correct managed Expo ownership, identifiers, signing, store metadata/privacy declarations, permissions, backups/data retention, and release/update channels. Any OTA process must validate runtime compatibility and rollback separately.
- Symbolicated JavaScript/native crash reports from an authorized release test, inspected event payloads, source-map/symbol access, retention, and alert routing.

## Related applications

The mobile app shares API contracts with [Projects](../projects/) and the [Go API](../server/). Authentication changes must stay compatible across all three applications.
