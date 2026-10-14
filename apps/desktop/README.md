# FortyOne Desktop

A Tauri 2 desktop app for FortyOne. Package the production app for this Mac, or use the development window with your existing local web server.

## Package the production app on macOS

With stable Rust, Apple's command line tools, and workspace dependencies installed as described below, run this from the repository root:

```sh
pnpm package:desktop
```

This creates a release-mode `FortyOne.app` and a DMG using the `production` Cargo feature and `src-tauri/tauri.production.conf.json`. The app is named **FortyOne**, uses the separate bundle identifier `app.fortyone.desktop`, and always opens [cloud.fortyone.app](https://cloud.fortyone.app). Local `FORTYONE_DESKTOP_URL` and inherited Tauri configuration overrides are removed from the packaging environment. No local web server or Go API is needed to run this version.

With the default Cargo target directory, find the DMG in `apps/desktop/src-tauri/target/release/bundle/dmg/`; the app is inside it. The build targets the Mac running the command: Apple silicon on this machine. Open the DMG, drag **FortyOne** into **Applications**, eject the image, and launch FortyOne from Applications. Sign in separately from FortyOne Development or your browser.

The local package uses [ad-hoc signing](https://v2.tauri.app/distribute/sign/macos/#ad-hoc-signing) (`-`), with no Apple credentials or notarization. macOS may require allowing the app in **System Settings → Privacy & Security**. This is a local installation build, not a notarized public release.

The DMG contains the native shell and loads the production web frontend over HTTPS; it does not bundle your local Next.js files. Internet access and the production services are required. Frontend changes appear only after they are deployed to the production site. To pick up native changes, run `pnpm package:desktop` again, quit the installed app, and replace it in Applications using the new DMG.

## Run on macOS

Install [stable Rust](https://rustup.rs/) and Apple's command line tools if they are not installed:

```sh
xcode-select --install
```

From the repository root, install workspace dependencies once:

```sh
pnpm install
```

The Go API must run separately with your existing local configuration and approved PostgreSQL and Redis endpoints. If it is not already running, start it from the repository root in its own terminal:

```sh
make -C apps/server develop
```

Start the web app in your usual terminal. If it is already running, keep using that process:

```sh
pnpm dev:projects
```

In another terminal, open the desktop app:

```sh
pnpm dev:desktop
```

The first launch compiles the Rust dependencies and takes longer. Later launches reuse the build cache. Web changes hot reload through the existing Next.js development server. Quit the desktop app or press Control-C in its terminal to stop the desktop development process.

The launcher checks that the web app is reachable before compiling. It also finds Rust in `~/.cargo/bin` when that directory is missing from the current shell's PATH; it does not change your shell configuration.

## A different local URL

The default is `http://localhost:3000`. To use the port where your web app is already running:

```sh
FORTYONE_DESKTOP_URL=http://localhost:3001 pnpm dev:desktop
```

Only HTTP or HTTPS URLs on `localhost`, `127.0.0.1`, or `[::1]` are accepted. Credentials and URL fragments are rejected. Use the same hostname your API uses for session cookies (normally `localhost` for both). A different app port must also be allowed by the API's existing CORS configuration. HTTPS requires a certificate trusted by both Node.js and macOS.

## Sign in and local scope

The desktop webview has its own cookie store. Signing in to the web app in Safari or Chrome does not sign in to the desktop app. Enter your email, then type the emailed six-digit code inside the desktop window. Clicking the email's link signs in to your browser instead. The desktop session uses the existing persistent HttpOnly cookie.

On macOS, native traffic lights share one header with Back, Forward, search, and the app's existing invite, help, notifications, profile, and contextual create controls. The header and sidebar form a continuous translucent surface over the native macOS Sidebar material, with a subtle theme tint and an opaque main content canvas. The material follows the web app's light, dark, or system theme, window focus, and macOS accessibility preferences. Drag empty header space to move the window, or double-click it to maximize and restore. Back and Forward are disabled when the webview has no history in that direction. Use the existing sidebar edge handle to expand or collapse the sidebar.

Translucency requires both the updated native shell and the matching web frontend. The bridge advertises support so older desktop builds and ordinary browser pages retain their opaque surfaces. Rebuild the desktop app after native changes; production also requires deploying the frontend changes. Tauri's transparent macOS webview uses its `macos-private-api` feature because WebKit's public background-color API does not clear the page canvas. This is compatible with the current locally distributed DMG, but a future Mac App Store release would require revisiting that implementation.

Login, sign-up, and onboarding extend to the top of the desktop window with only the native traffic lights visible. Their transparent top drag region retains the same move and double-click behavior; it adds no toolbar spacing or Back/Forward buttons. This scoped presentation is included in the native bridge, so rebuilding the DMG updates it without deploying the web app. Browser pages, workspace headers, and settings keep their existing layout.

File → New Story (`Cmd+N`) opens the existing creation dialog, and View → Search (`Cmd+K`) opens app search. The creation label follows the workspace's terminology, and actions remain disabled until the app is ready and the relevant permissions are available.

The native Navigate menu provides Back, Forward, Reload (`Cmd+R`), Open Start Page, Open in Browser, and Developer Tools (`Cmd+Option+I`, development builds only). Window size and position are saved between launches. Local development stays on its configured app origin. Production keeps the sign-in host and workspace subdomains inside the same window, using the existing hosted routing and session cookies. Links to other sites open in your system browser.

The production app also restores the last approved workspace page after a normal close or quit. It saves the current webview location at shutdown, including pages reached through client-side navigation, without writing to disk during ordinary navigation. Only known app routes on validated workspace origins are eligible; authentication callbacks, public/shared links, query strings, and URL fragments are excluded. Sign-out or returning to cloud sign-in clears the saved location. Missing, invalid, or unreadable state falls back to cloud sign-in, and local development retains its configured start URL. This restores the page, not open dialogs, unsaved edits, or crash state.

The configured app origin and validated production workspace origins in the main webview can call four narrowly scoped native commands to synchronize appearance and menu availability, read navigation availability, go back or forward, and drag or maximize the window. Production permissions are registered for each exact HTTPS origin as it is visited; infrastructure hosts and unrelated sites receive none. These commands have no file, network, or record-writing access. Header and native menu actions open the existing web app interfaces, which retain their existing validation and permissions.

System-browser OAuth/SSO sign-in and callback handoff are not implemented yet. Native notifications, deep links, automatic updates, and offline support are also outside this prototype. Local development still requires the existing web app, API configuration, and network connection. Keep both the web server and Go API running while using FortyOne Development; the packaged production app uses the hosted services instead.

## Sign-in troubleshooting

If sign-in reports "Failed to send magic link" with "Load failed", first check that the Go API is running. The desktop launcher checks the web server; loading the sign-in page does not prove the API is available. With the default local API port, check readiness:

```sh
curl --fail http://localhost:8000/readiness
```

The response should report `"status":"ok"` and `"phase":"ready"`. If the connection fails, start the API with `make -C apps/server develop`; if readiness fails, check its terminal for configuration or dependency errors. Use your configured API port if it differs. Once the API is ready, retry sign-in. If the request still fails, check the API logs and ensure its CORS configuration allows the exact web app origin, including the port.

If your local API is configured to send email through `localhost:1025`, run a local SMTP inbox such as [Mailpit](https://mailpit.axllent.org/docs/install/) in a separate terminal:

```sh
brew install mailpit
mailpit --listen 127.0.0.1:8025 --smtp 127.0.0.1:1025
```

Open [the local inbox](http://localhost:8025), request a new sign-in email, and enter its six-digit code in the desktop window. These messages stay in the local inbox rather than reaching your email provider. Use the code instead of the email link, which signs in to your browser. This step is only needed for an API configuration using local SMTP; keep your existing email setup otherwise.

## Check and build

```sh
pnpm check:desktop
pnpm build:desktop
```

The check command runs launcher validation tests, Rust formatting, compilation checks, and Rust tests. The build command creates a debug executable for local use without signing or generating an installer. It captures `FORTYONE_DESKTOP_URL` at build time, so set that variable for a custom port when building too. The output remains dependent on the running local web app.

For production-only native changes, also run `cargo test --locked --features production` and `cargo check --locked --features production` from `apps/desktop/src-tauri`. Route restoration tests use isolated temporary directories and cover safe URL normalization, auth/public-route exclusions, missing or corrupt state, logout clearing policy, and preserving the previous file when replacement fails. They do not launch the installed app; close/quit/relaunch behavior still needs a packaged-app smoke test before distribution.

Direct Tauri CLI access is available for diagnostics:

```sh
pnpm --filter desktop tauri info
```

The workspace uses dedicated desktop script names so the repository's ordinary `pnpm dev` and `pnpm build` commands do not unexpectedly launch a native window or build native applications. Windows packaging and native behavior still need validation on Windows.

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
