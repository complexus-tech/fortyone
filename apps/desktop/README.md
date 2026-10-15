# FortyOne Desktop

The default Tauri 2 app runs the **full Projects application**. It preserves the existing stories, documents, boards, settings, navigation, and account flows. Desktop capabilities extend that interface; the separate bundled React client is experimental and must be selected explicitly with a `:native` script.

## Commands

Run these from the repository root after `pnpm install`, stable Rust, and the platform Tauri prerequisites. On macOS, install Apple's command-line tools with `xcode-select --install`.

| Command                                      | Result                                                                             |
| -------------------------------------------- | ---------------------------------------------------------------------------------- |
| `pnpm dev:desktop`                           | Full Projects app in Tauri, using the already running local web app.               |
| `pnpm dev:desktop:hosted`                    | Compatibility alias for the full Projects development shell.                       |
| `pnpm build:desktop`                         | Debug executable for the full Projects shell; no installer.                        |
| `pnpm package:desktop`                       | Production Projects shell packaged as a local macOS DMG or Windows NSIS installer. |
| `pnpm release:desktop`                       | Distribution packaging with required signing and updater configuration.            |
| `pnpm check:desktop`                         | Launcher/bridge tests, Rust checks/tests, and experimental renderer checks.        |
| `pnpm --filter desktop desktop:dev:native`   | Explicit experimental bundled client, Vite on port 5178.                           |
| `pnpm --filter desktop desktop:build:native` | Debug build of the experimental bundled client.                                    |

Start the existing local services separately:

```sh
make -C apps/server develop
pnpm dev:projects
pnpm dev:desktop
```

The launcher finds Rust in `~/.cargo/bin` without modifying shell configuration. Normal monorepo `pnpm dev` and `pnpm build` do not launch or package the desktop app. Only the approved local PostgreSQL/Redis configuration is supported; the launcher does not provision a backend.

The full development app defaults to `http://localhost:3000`; native API verification uses `http://localhost:8000`. A different web port can be selected with `FORTYONE_DESKTOP_URL=http://localhost:3001 pnpm dev:desktop`. Only loopback HTTP(S) URLs are accepted. Match the local API's cookie and exact CORS-origin configuration. Production uses the fixed `https://cloud.fortyone.app` entry point and approved workspace subdomains, ignoring the development override.

## Full Projects app: native capabilities

The hosted bridge adds native window chrome, window-state restoration, history controls, native New Story/Search menu actions, and workspace-switch feedback to the existing UI. Workspace switching immediately shows the target and a spinner, prevents duplicate clicks, and retains progress through document navigation. A timeout offers Retry and Return. Saving the last workspace is a bounded preference write, not an authentication check.

Desktop controls appear in the existing **Settings → Preferences** and **Settings → Notifications** pages. They manage launch at login, whether closing the window keeps the process running, and system notifications on this computer. Saves show pending and failure states. Older desktop builds without these optional methods retain their existing behavior; ordinary browser settings do not gain native controls.

The menu-bar/tray menu opens FortyOne, New Story, Search, or Quit. On macOS, the tray uses a thin outline of the existing logo with a transparent, padded canvas, about 10% smaller than the previous solid mark. It is a template icon so the system chooses its color; the Dock/app icon keeps its artwork. Regenerate the 36×36 template PNG with `swift apps/desktop/scripts/prepare-tray-icon.swift` after editing its SVG. Closing the window with background operation enabled hides it. Disabling that option makes closing quit. Dock reopening restores the macOS window.

Window close, tray Quit, and app Quit wait for the main app's local draft writes. Failure or a 12-second timeout keeps the window open and surfaces the save error. A later explicit close offers Retry saving, Stay open, or Close without saving. The sanitized workspace URL is persisted before hide, quit, or an approved updater restart, rather than relying on a second native exit callback. Network synchronization is not required to close after local writes succeed; saved outbox work remains on disk.

## Authentication and local storage

The default app uses its ordinary authenticated webview session. Native code privately reads only the configured API's HttpOnly session cookie, verifies the account and workspace, and stores a cookie fingerprint with the native binding. Cookie values are never returned through IPC. This flow does not require the experimental client's Keychain session or custom callback scheme.

Native query caches, drafts, pending operations, and private page HTML are scoped by the exact origin, account, and workspace. Commands verify the current window, cookie binding, and workspace lease again before access. Bindings expire after seven days without online verification. Confirmed denial and account/logout transitions clear the relevant binding and server-owned cached material; draft/outbox handling must preserve work until the user completes the recovery or discard flow. The renderer does not receive general filesystem, arbitrary URL, or arbitrary cookie access.

The main app persists an allowlist of existing query families, retaining at most 250 query entries and 8 MiB per saved workspace snapshot. Startup warms dependencies such as workspace/settings metadata before useful work collections. When already offline, cache restoration uses the native cookie-bound account/workspace lease without waiting for another HTTP verification; reconnect still verifies the session online. Cached views keep their normal components. Routine downloads and synchronization stay in the background. A compact Offline badge replaces the trial/subscription footer while offline; saved-change review and cache recovery are available in Desktop App settings. Data that has never been downloaded remains online-only. Story/document draft recovery is integrated with the existing editors.

Offline cold startup uses native-owned private HTML and a desktop-only service worker for public Next assets. Private HTML is not stored in browser CacheStorage. Each page is limited to 4 MiB, with at most 24 pages per workspace and 64 MiB per origin/account, including saved security policies. Pages expire after seven days. Only approved workspace routes are eligible; authentication routes and secret-bearing query parameters are rejected. Settings HTML is limited to the settings root, account profile, preferences, notifications, and workspace members. Security, developer/API-key, billing, integrations, and unknown settings pages are excluded. The saved Content-Security-Policy is preserved. An offline restart first revalidates the native session before a short-lived, one-use worker response returns the HTML.

Service-worker updates wait until all windows controlled by the previous worker are closed before activating. This preserves in-flight Next responses and one-use offline handoffs. Hiding a window to the tray does not close it; fully quit and reopen to activate a waiting update. The first worker installation activates normally, and updates retain the existing public asset cache for downloaded pages.

Local QA has exercised downloaded My Work reload, full quit/reopen, offline story creation, and subsequent operation replay in the actual full Projects app. The queued creation survived restart and produced exactly one server story and matching receipt after reconnect, with an empty local queue. Neither that evidence nor the unit tests proves that all dynamically loaded Next chunks, fonts, or route transitions are available in a packaged offline WKWebView. Do not describe downloaded query data alone as a complete offline application, or assume all web mutations work offline. The hosted outbox targets story creation/updates, comments, and staged story attachments; unsupported mutations still require connectivity. Pending work and conflicts can be reviewed in Desktop App settings. Active-workspace replay runs on reconnect, foreground activation, and every 30 seconds while pending work exists. Conflicts show server and local values and require explicit review before retrying. A write with an unknown server outcome keeps its original operation ID and cannot be discarded. Updates without a saved server comparison cannot be discarded until that comparison is available.

Offline links and the story context menu, related-document menu, and settings back action verify a downloaded destination and flush active editors before a full navigation. A failed write or missing page keeps the current page open with a message. Other programmatic router flows still require route-specific runtime validation; cached query data does not by itself provide a saved page for a new route.

Story attachments can be staged locally before upload, up to 10 MiB per file and 128 MiB of retained bytes per origin/account. Pending file bytes survive restart and upload through the existing story attachment flow with stable operation IDs. Locally retained files can be previewed or exported through native, account-scoped commands. This does not download every historical server attachment or provide general offline document/media parity.

Collaborative documents retain Yjs state and permissions in account/workspace-scoped drafts; reopening a downloaded document resumes its collaboration session. A restored document epoch or revoked editing access preserves unsynced work and blocks unsafe merging. Recovery export uses the native save dialog and requires a confirmed file write, with no newer local edits, before replacing the active recovery draft. Closed documents are not synchronized by a separate background collaboration worker.

## Notifications

The full Projects app reuses its **existing active-workspace SSE connection**. It does not create a polling loop or an account-wide native subscription. Live notification payloads are normalized, checked against the current account/workspace, and sent through the native bridge. Historical inbox loads and SSE reconnects do not produce old system alerts. Rust deduplicates each account/workspace/notification/version durably, while a newly coalesced notification version can alert again. Foreground windows are not unconditionally suppressed.

The native badge mirrors the active workspace's unread query. Notification clicks validate the current account/workspace and route through the existing notification detail page or inbox; arbitrary URLs are not accepted. A notification from another no-longer-active workspace is not silently authorized from its old payload.

Notifications require the preference to be enabled and the process to remain running. A hidden window can keep the existing stream connected; fully quitting stops delivery. macOS uses `UNUserNotificationCenter` to request alert and sound authorization when the notification switch is enabled, check system authorization before delivery, and request banners, Notification Center entries, and the default sound while the app is in the foreground. Focus/Do Not Disturb and OS notification settings still apply. Successful submission means macOS accepted the request; it does not prove that a banner was shown or a sound was heard. Windows retains the Tauri plugin, whose desktop permission API does not inspect the actual OS setting; an installed app and platform QA are required for notification identity and click behavior. macOS uses a numeric Dock badge; Windows uses an overlay indication. The unread badge is independent of the system notification preference.

## Packaging and distribution

Default packaging uses the `production` feature and the full hosted app. It does **not** enable `native-client`. A local macOS package is ad-hoc signed and not notarized; a local Windows package is not a signed public release. Packaging removes inherited signing/updater credentials and arbitrary Tauri overrides. With the default target directory, installers are under `apps/desktop/src-tauri/target/release/bundle/`.

Release packaging requires:

- `FORTYONE_UPDATER_PUBLIC_KEY`, HTTPS `FORTYONE_UPDATER_URL`, and `TAURI_SIGNING_PRIVATE_KEY` (plus its password when required).
- macOS: Developer ID `APPLE_SIGNING_IDENTITY` and either `APPLE_API_ISSUER`/`APPLE_API_KEY`/`APPLE_API_KEY_PATH` or `APPLE_ID`/`APPLE_PASSWORD`/`APPLE_TEAM_ID` for notarization.
- Windows: `FORTYONE_WINDOWS_CERTIFICATE_THUMBPRINT` and `FORTYONE_WINDOWS_TIMESTAMP_URL`.

Updater signatures and OS installer signing are separate requirements. Publish the matching signed update metadata/artifacts and validate installed upgrade behavior. The main app's desktop preferences include an explicit update check and reviewed install/restart action. Development or unconfigured packages show that updates are unavailable. The empty `plugins.updater.pubkey` in the base/local package configuration only allows the Tauri plugin to initialize; it does not enable update checks or installation. Release packaging rejects a blank key, and native update controls require the production feature, a nonempty compiled public key, and a compiled HTTPS endpoint. Native code rechecks the requested version, verifies the downloaded signature, and flushes local drafts before installation. A failed flush prevents installation; close/quit is blocked while an update is in progress. Local tests do not execute an installer or prove installed upgrade behavior. Rebuild the installed app for Rust/bridge changes; the main UI is served by the deployed Projects app. macOS transparency uses `macos-private-api`; Mac App Store distribution requires revisiting that implementation.

## Experimental bundled client

The opt-in `native-client` path has a separate React renderer and native SQLite/synchronization model. Its bundled startup, broad native downloads, rich editor collaboration, native media/export, account-wide SSE, recovery export, and updater UI belong to that experimental path. They are not evidence that equivalent flows have shipped in the default full Projects app.

It authorizes through the system browser at `/auth/desktop`, exchanges a single-use code with S256 PKCE, and retains its token in the OS credential store. Production callbacks use `fortyone-desktop://login`; development uses `fortyone-desktop-dev://login`, accepted only by development Projects/API configuration. On macOS, register a dedicated development `.app` declaring that scheme before callback QA; a raw executable does not register it.

Rebuilding an ad-hoc macOS app changes its signing identity and may cause a fresh Keychain authorization prompt. Use a consistent signing identity and bundle identifier for repeatable signed QA. The native credential cache serializes reads and latches denied/unavailable results so background work does not repeatedly reopen a canceled prompt; an explicit retry or browser reconnect is required. This credential-store behavior belongs to the experimental client. See Apple's [code-signing requirements](https://developer.apple.com/documentation/technotes/tn3127-inside-code-signing-requirements).

Experimental collaboration uses `FORTYONE_COLLABORATION_URL` and exact service `COLLABORATION_ALLOWED_ORIGINS`: `tauri://localhost`, the installed Windows `http://tauri.localhost` or `https://tauri.localhost` mapping, and development `http://localhost:5178`. Keep these alongside approved web origins. An unavailable collaboration service does not make live collaboration available offline.

## Validation and deployment

Deploy the required Projects/native/API contracts together. Operation replay requires migration `000218_desktop_operation_receipts.up.sql` through the normal deployment process; editing source or running tests does not apply it. Existing API sessions, database, Redis/SSE, storage, and collaboration services retain their normal operational requirements.

For the local QA environment, migration 218 was applied on 2026-10-15 and its clean version, receipt table, constraints, and API readiness were verified. This records the local database change only; production migration and deployment are separate steps.

On 2026-10-15, the rebuilt production hosted shell replaced the local `/Applications/FortyOne.app`, preserving the `app.fortyone.desktop` identifier and existing user storage. The updated app reopened the authenticated live workspace successfully. The previous bundle was retained in a local installer backup. This installation does not deploy the Projects frontend, API, worker, or database migration to production; those contracts must ship separately before the installed app can use their new behavior.

Focused Projects tests cover native notification filtering, SSE lifecycle, settings interactions, cache scope, and draft behavior. Rust tests cover native scope checks, cached pages, preferences, notification receipts, and operation handling. The service-worker tests are server-independent contract tests. Keep experimental renderer tests separate from default full-app evidence.

Before release, test an installed signed build on macOS and Windows: ordinary main-app login/SSO, workspace switching, denied/expired sessions, offline cold launch and route changes, queued work across quit/relaunch, conflicts and uncertain outcomes, close-to-tray/reopen, real notification delivery/click activation, and signed updates. Unit tests or a successful compilation do not establish deployed or OS-level behavior. Record PostgreSQL integration-test gaps separately when `TEST_DATABASE_URL` is unavailable.

Validation checklist:

- Implemented in the main app: native preferences, active-workspace SSE notifications/badge, scoped query/page/draft persistence, reviewed outbox replay, staged story files, collaboration recovery export, durable close/quit, and explicit signed updater controls.
- Observed in local macOS full-app QA on 2026-10-15 using a production Next build: native launch, My Work reload and quit/reopen while web/API services were stopped, a durable offline creation replayed exactly once, cached Profile/Preferences/Notifications navigation, cached task custom-field definitions, the Working offline indicator, and visible Switching to FortyOne feedback during a deliberately delayed API response. Native notification preferences were enabled through the actual main app. This is narrower than a signed production or Windows validation.
- Observed in the installed production hosted macOS app on 2026-10-15: the updated bundle reopened the signed-in workspace, the desktop notification preference was enabled, and macOS showed desktop alerts, Notification Center, and sound enabled for FortyOne. After a fresh live workspace update, the user confirmed both a system banner and sound. The notification adapter and main-app forwarding/settings have 27 passing targeted tests; this does not establish click activation, Focus behavior, or other platforms.
- Still requires installed runtime evidence: all dynamic chunks/fonts and remaining offline routes, attachment recovery/upload after restart, macOS notification click activation and lifecycle, Windows notification delivery/activation, permission denial, and a real signed upgrade. Automated tests do not send live notifications, execute an installer, or apply migrations.

For local sign-in failures, check the API before changing WebView configuration:

```sh
curl --fail http://localhost:8000/readiness
pnpm --filter desktop tauri info
```

A login page loading does not prove API readiness or local email delivery. Keep the web app's existing SMTP/Mailpit setup.

## Desktop icon

The desktop icon uses the user-provided FortyOne artwork, preserved unchanged in `src-tauri/icons/source.png`. The derived `source-cropped.png` enlarges the white 4 by 23%: it crops 24 pixels from each side of the original 256×256 image, scales that 208×208 crop back to the canvas, and restores the original alpha mask so the black tile keeps its exact corner silhouette. The mark is resampled, never redrawn.

To reproduce the crop and desktop assets on macOS, run from the repository root:

```sh
swift apps/desktop/scripts/prepare-icon.swift
icon_output_dir="$(mktemp -d)"
pnpm --filter desktop tauri icon "$PWD/apps/desktop/src-tauri/icons/source-cropped.png" --output "$icon_output_dir"
for icon in icon.png 32x32.png 128x128.png 128x128@2x.png icon.icns icon.ico; do
  cp "$icon_output_dir/$icon" apps/desktop/src-tauri/icons/
done
rm -r "$icon_output_dir"
```

Only the configured desktop PNG, macOS ICNS, and Windows ICO outputs are copied; the CLI's unused mobile assets stay out of the repository. Repackage the desktop app after changing these assets to include the new icon.

## Squircle compatibility

Projects uses native CSS squircles where supported and ordinary CSS rounded corners elsewhere, including macOS webviews without `corner-shape` support. The Hyperellipse fallback is disabled in the interactive workspace and settings shells: controlled WKWebView measurements showed substantial scripting and frame-delay costs during board updates and ticket opening. Existing radius, color, border, and font tokens remain in use.

The separate marketing landing page still uses the shared fallback; its workload has not been benchmarked. The opt-in benchmark and its measured results live in [`scripts/performance`](scripts/performance).

Hyperellipse is pinned to 1.0.5 with a pnpm patch. The patch preserves author inline styles during rendering and cleanup, releases observers and listeners for removed elements, refreshes focus-within ancestors, and prevents its temporary paint changes from being sampled as transparent CSS transitions. Transform and opacity transitions remain native. Paint-contained and overflow-clipped surfaces keep their border ring on the host background so descendant clipping cannot erase it. Keep the patch when upgrading until the upstream implementation covers these cases; verify task cards (including their thin borders), hover and focus states, theme changes, and portaled menus in Safari and WKWebView after an upgrade.
