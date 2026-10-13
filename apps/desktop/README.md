# FortyOne Desktop

A Tauri 2 desktop window for the existing local FortyOne web app. The web app and its backend keep running independently; this workspace attaches to the web server you already use.

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

On macOS, native traffic lights share one flat header with Back, Forward, search, and the app's existing invite, help, notifications, profile, and contextual create controls. The header follows the web app's light, dark, or system theme and matches its background without Liquid Glass. Drag empty header space to move the window, or double-click it to maximize and restore. Back and Forward are disabled when the webview has no history in that direction. Use the existing sidebar edge handle to expand or collapse the sidebar.

File → New Story (`Cmd+N`) opens the existing creation dialog, and View → Search (`Cmd+K`) opens app search. The creation label follows the workspace's terminology, and actions remain disabled until the app is ready and the relevant permissions are available.

The native Navigate menu provides Back, Forward, Reload (`Cmd+R`), Open Start Page, Open in Browser, and Developer Tools (`Cmd+Option+I`). Window size and position are saved between launches. Links to other sites open in your system browser; the desktop window stays on the configured app origin.

The configured app origin in the main webview can call four narrowly scoped native commands to synchronize appearance and menu availability, read navigation availability, go back or forward, and drag or maximize the window. These commands have no file, network, or record-writing access. Header and native menu actions open the existing web app interfaces, which retain their existing validation and permissions.

System-browser OAuth/SSO sign-in and callback handoff are not implemented yet. Native notifications, deep links, automatic updates, installers, and offline support are also outside this local prototype. The existing web app, API configuration, and network connection are still required. Keep both the web server and Go API running while using the desktop app.

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

Direct Tauri CLI access is available for diagnostics:

```sh
pnpm --filter desktop tauri info
```

The workspace uses dedicated desktop script names so the repository's ordinary `pnpm dev` and `pnpm build` commands do not unexpectedly launch a native window or build native applications. Windows packaging and native behavior still need validation on Windows.

## Desktop icon

The desktop icon uses the user-provided FortyOne artwork, preserved in `src-tauri/icons/source.png`. The Tauri icon CLI generated the PNG, macOS ICNS, and Windows ICO assets from that source without redesigning the artwork. Restart the desktop development process after changing these assets to rebuild the native icon.

## Desktop squircles

Desktop webviews without native `corner-shape` support load Hyperellipse on the client. Its CSS carriers are scoped to the desktop provider; ordinary web browsers keep their existing rendering. The fallback uses the same radius tokens as the native squircle styles. Form controls, circular shapes, translucent backdrops, the main content canvas, and controls with decorative pseudo-elements retain their existing corners. Add `data-squircle="off"` to opt a surface out.

Hyperellipse is pinned to 1.0.5 with a pnpm patch. The patch preserves author inline styles during rendering and cleanup, releases observers and listeners for removed elements, refreshes focus-within ancestors, and prevents its temporary paint changes from being sampled as transparent CSS transitions. Transform and opacity transitions remain native. Paint-contained and overflow-clipped surfaces keep their border ring on the host background so descendant clipping cannot erase it. Keep the patch when upgrading until the upstream implementation covers these cases; verify task cards (including their thin borders), hover and focus states, theme changes, and portaled menus in WKWebView after an upgrade.
