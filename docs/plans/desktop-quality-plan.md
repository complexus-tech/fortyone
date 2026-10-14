# FortyOne desktop performance and quality plan

Status: initial performance changes and the React performance pass implemented locally; UX and visual changes proposed for review.

The first priority is moving between pages and opening tickets. The quality target is consistently fast, predictable daily use on the Mac, measured with repeatable workloads. A score out of ten should follow those results rather than replace them.

## Initial findings and changes

The patched Hyperellipse fallback has a substantial cost in a controlled native WKWebView test. In the 240-card fixture, eight page replacement and ticket insertion cycles took a median 7.90 seconds with the fallback and 2.20 seconds without it. The median trial p95 frame gap was 643ms with the fallback and 36ms without it. The 24-card fixture also showed increased CPU use and frame delays. The [benchmark report and reproducible harness](../../apps/desktop/scripts/performance/README.md) include all trials, focus exclusions, source hashes, and methodology.

These are rendering workload measurements, not production page-load times. The fixture excludes React, routing, authentication, and network requests, and its observed native frame cadence was about 30Hz. It does not establish the performance of the separate landing page.

The local implementation now:

- Removes the JavaScript squircle fallback from Projects workspace and settings shells. Browsers with native corner support keep squircles; unsupported engines use existing CSS rounded corners. The landing integration remains separate.
- Mounts one responsive ticket details tree. Previously the hidden mobile and desktop trees each mounted title and description editors; the initial editor count falls from four to two.
- Uses the same freshness rules for ticket hover prefetch and actual readers, avoiding repeated requests for fresh data. Rows that open a dialog skip the explicit full-page route prefetch.
- Hydrates the workspace list from the membership check already completed in the same authorized server request. It also removes the unrelated startup invoice prefetch, whose key did not match the billing reader.
- Restores a validated production workspace URL after normal close or quit. Only an allowed HTTPS origin and supported page path are saved; query strings and fragments are excluded. Sign-out clears the saved location. Authentication and workspace authorization still run normally on reopening.

The performance work does not change the base font or color and border tokens. Separate theme token edits are outside this pass. Native route restoration requires a rebuilt desktop binary; web changes require the updated Projects application. No deployment or replacement of the installed app is included.

## React performance pass

The follow-up implementation addresses repeated work during navigation, ticket editing, and large Strategy Maps:

- Keeps Maya's runtime mounted in the workspace and shares it between the full page and chat rail. Navigation preserves the active conversation, stream, draft, attachments, approvals, and voice session. Explicit conversation selection, new chat, and workspace changes still establish the appropriate session. Stable action and visibility contexts prevent streaming updates from notifying unrelated controls.
- Starts loading ticket editor code alongside intent-based data prefetch and direct ticket loading. The module stays dynamically imported; preloading does not mount editors. A failed speculative import can be retried when opening the ticket.
- Captures immutable editor documents on each edit and serializes only when the debounced save runs or flushes. Title and description saves remain independent; navigation, blur, pending uploads, and editor destruction preserve the correct content and ticket ID.
- Separates board display settings from selected ticket IDs. A 250-consumer regression fixture verifies that selection changes do not notify card display consumers, while actual display-setting changes do. Stable ticket keys also preserve row state when titles change.
- Streams objective metadata and onboarding progress outside the workspace shell's blocking query group. Access checks remain mandatory. Short freshness windows avoid immediate duplicate status and running-sprint reads; sprint mutations explicitly invalidate the running-sprint cache.
- Stabilizes Strategy Map query snapshots and caches connector geometry independently of the camera. A 500-objective fixture verifies that 30 camera-only updates within the same query window perform no layout rebuilds, connection-list rebuilds, path calculations, storage writes, or API requests. Data changes, moved nodes, workspace changes, and new query windows still refresh the relevant results. Existing viewport culling remains in place.

These are automated work-count and behavior checks, not measured production latency or frame-rate improvements. Authenticated before/after timings remain a release verification step.

## Measure the complete application

Use a production-mode profiling build with Web Inspector enabled only for that build, plus an authenticated test workspace. Keep the ordinary release build's debugging restrictions. Record browser network requests, script execution, layout, paint, allocations, and WebContent CPU together; [WebKit Timelines](https://webkit.org/web-inspector/timelines-tab/) provides these views. Capture separate runs without profiling overhead for the final interaction timings.

Use the same Mac, visible window, viewport, theme, dataset, and network for before/after runs. Record actual frame cadence and focus, exclude paired runs affected by backgrounding, and separate cold cache from warm cache. Run at least 20 samples per flow and report p50, p95, sample count, and range; increase samples when results vary. Compare native WKWebView with Safari and Chromium. Read-only checks against the installed production app establish navigation behavior but accessibility-tool timings are not browser performance measurements.

| Flow                           | Record                                                                               | Proposed acceptance target                                                                    |
| ------------------------------ | ------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------- |
| Cold app launch                | Launch to first useful screen, authentication and network separately                 | Useful content within 2 seconds on the defined test network; immediate loading feedback       |
| Warm page navigation           | Click to usable destination, preserving navigation and focus                         | p95 at or below 200ms when required code and data are cached                                  |
| Ticket opening                 | Click to editable title and description, cold and warm                               | Cached p95 at or below 200ms; uncached p95 below 1 second on the defined test network         |
| Search                         | Key press to visible input and result update                                         | Input feedback below 100ms; separate server result latency                                    |
| Board and list                 | Scroll, selection, filtering, dragging, and ticket opening at normal and large sizes | No recurring main-thread tasks above 50ms; frame gaps stay close to the measured idle cadence |
| Maya and documents             | Typing and navigation while streaming or collaborating                               | Typing stays responsive; no repeating full-page layout work                                   |
| Idle and long sessions         | CPU after settling; memory after 30 navigation and ticket cycles                     | No sustained idle CPU loop or growing retained editor/listener count                          |
| Reconnect and workspace switch | Recovery, authorization, subscriptions, and cache isolation                          | No duplicate event streams, stale workspace data, or lost pending input                       |

Measure cache hits, first-load JavaScript, editor construction, server response time, and request waterfalls before choosing the next optimization. Ticket editor code now preloads alongside data; verify the remaining cold-open cost with actual traces. Virtualize large lists only if row rendering is the measured bottleneck; verify keyboard navigation, drag behavior, and variable heights before adopting it.

## Desktop integration sequence

1. **Verify restoration in a packaged build.** Cover normal quit, window close, reopening, sign-out, revoked membership, expired session, malformed saved data, and production workspace subdomains. The current implementation restores a page, not unsaved drafts or an open ticket dialog.
2. **Add reliable startup and network recovery.** A small local startup surface should remain usable when hosted content cannot load, with retry and open-in-browser actions. Approve its visual treatment with the UX work below before implementation.
3. **Complete browser authentication and deep links.** Design a one-time state/PKCE exchange that establishes the existing HttpOnly webview session without putting reusable tokens in URLs. Validate workspace targets, cold/warm launch, cancellation, expired callbacks, and SSO. [Tauri deep linking](https://v2.tauri.app/plugin/deep-linking/) supplies the OS entry point; the secure session exchange also needs backend work.
4. **Add native notifications.** Reuse existing preferences and event IDs; deduplicate against the in-app stream, avoid notifying for the focused conversation, and open a validated destination on click. Request permission at a meaningful moment. [Tauri notifications](https://v2.tauri.app/plugin/notification/) provide native delivery.
5. **Make releases dependable.** Add Developer ID signing, notarization, a separate updater key, and a hosted release manifest. Tauri requires signatures for updater artifacts; that key is distinct from Apple's signing identity. Follow the [macOS signing](https://v2.tauri.app/distribute/sign/macos/) and [updater](https://v2.tauri.app/plugin/updater/) requirements. Verify upgrade, failed-download recovery, and rollback before enabling automatic checks.

The app remains online-first. Offline data editing and conflict resolution would be a separate product commitment.

## Everyday UX proposal

These changes await review before implementation:

- **Navigation and ticket continuity:** preserve list scroll position, selection, and return focus; make keyboard shortcuts discoverable; verify Escape, Back, Forward, and Cmd-click consistently. Keep draft safety during route changes and resizing.
- **Loading and recovery:** acknowledge clicks immediately, keep useful cached content visible while refreshing, and show actionable retry states. Avoid replacing the entire workspace for a local request.
- **Information hierarchy:** review ticket row metadata, date labels, property density, and truncation in real narrow and wide windows. Prioritize the title and next action without hiding important status.
- **Accessibility:** test complete keyboard flows, clear focus indicators, VoiceOver labels, reduced motion, and comfortable target sizes for dense toolbar controls.

Start with My Work to ticket details and back, then global search, then Maya. Review screenshots and a small interactive prototype of each proposed change before altering the shared shell.

## Visual polish proposal

Keep the current base font, brand colors, flat window chrome, native traffic lights, and draggable header. Do not introduce Liquid Glass.

Review light and dark modes side by side for icon weight and optical alignment, search-field contrast, control heights, borders, hover/selected states, and spacing. Use existing component tokens so each correction is consistent. Test compact and wide windows, with and without the sidebar, and confirm that ordinary Safari corners remain polished after removing the interactive fallback.

## Release gates

Ship the performance changes after focused tests, Projects type checking and production build, native tests and lint, and visual regression checks of ticket page, dialog, and notification layouts at mobile and desktop widths. Re-run the authenticated flow measurements after deployment before claiming a production speed improvement.

The UX and visual proposals need approval first. Excellent desktop integration also requires the authentication, recovery, notification, and release checks above; route restoration alone does not meet that target.
