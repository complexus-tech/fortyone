# Design QA

## Scope

- Public feedback roadmap cards
- Contributor profile navigation and feedback history
- Bottom-right board-creation CTA
- Organization-specific public portal favicon

## Visual references

- Contributor profile reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_D18xP2/Screenshot 2026-07-20 at 7.38.23 AM.png`
- Roadmap implementation capture: `/tmp/fortyone-roadmap-qa.png`
- Contributor profile implementation capture: `/tmp/fortyone-author-profile-qa.png`

The reference and implementation captures were reviewed together. The reference is a profile-page direction rather than a pixel-identical FortyOne screen, so the implementation keeps FortyOne's existing navigation, spacing, colors, typography, and component scale.

## Checklist

- [x] Card title is the first element and uses the Projects Kanban title scale.
- [x] Author avatar uses the Kanban-style compact, squared avatar with a subtle outline.
- [x] Board is rendered as a color-tinted outlined pill.
- [x] Roadmap vote metadata is icon plus count with no container.
- [x] Author name opens a public contributor profile.
- [x] Contributor profile lists only that author's feedback and supports pagination.
- [x] Empty roadmap columns remain visually distinct without becoming prominent.
- [x] Bottom-right CTA is fixed, text-only, and reads “Create your own board”.
- [x] CTA preserves the feedback-settings destination through signup and onboarding.
- [x] Public portal metadata uses the organization logo for icon, shortcut icon, and Apple touch icon.
- [x] Portals without an organization logo use a 64 × 64 peach-dot PNG fallback.

## Functional verification

- Roadmap author link opened `/portal/art-circles/people/8a798112-90fe-495e-9f1c-f36655e3d8ab` successfully.
- The contributor page returned only Joseph Mukorivo's submitted feedback.
- The CTA resolved to signup with the nested onboarding and feedback-settings callback intact.
- The fallback favicon endpoint returned HTTP 200 with `image/png`, RGBA, and 64 × 64 dimensions.

## Result

Passed. No visible overflow, unintended card nesting, missing metadata, or broken primary navigation was found in the final browser pass.

---

# Strategy and Objectives Design QA

## References

- Objectives reference: `/Users/joseph/.codex/generated_images/019fb9c4-44cf-7572-bee1-f9ec161c2bd0/exec-ae19cf53-a803-4c5b-9600-123604f4e144.png`
- Strategy Map reference: `/Users/joseph/.codex/generated_images/019fb9c4-44cf-7572-bee1-f9ec161c2bd0/exec-ee460f02-969b-412a-b5e4-9703e72782eb.png`

## Review

- Strategy hierarchy is explicit: Ultimate Goal, Strategic Pillar, Objective, and expandable Key Result nodes.
- The canvas uses the product's existing typography, surfaces, borders, and monochrome palette.
- Objective-to-pillar alignment is visible and editable from each objective card.
- Key results are loaded only when an objective is expanded and render as nested measurable outcomes.
- Objectives timeline has stable Owner, Status, Objective, and Duration columns with the Gantt chart alongside.
- Timeline owner and status controls remain editable, and date dragging/navigation behavior is preserved.
- Temporary unauthenticated preview routes were removed after visual verification.

## Result

final result: passed

---

# Objectives Timeline Interaction Design QA

## Reference and comparison

- Linear details-panel reference: `/Users/joseph/Downloads/Screenshot 2026-08-01 at 10.04.13 AM.png` (2926 × 1580)
- FortyOne implementation capture: `/tmp/fortyone-objectives-timeline-implementation.png` (1280 × 720)
- Side-by-side normalized comparison: `/tmp/fortyone-objectives-timeline-comparison.png` (2560 × 720)

The reference and implementation were reviewed together at a normalized 1280 × 720 density. The implementation intentionally retains FortyOne's application shell, typography, spacing, and semantic colors while matching the reference interaction hierarchy.

## Review

- Objective rows share the same `3.5rem` CSS height on both sides of the timeline, and bars retain their original `2.5rem` height.
- Key-result markers are intentionally omitted from the timeline; key results remain available in the objective details panel.
- Off-screen objective locators use circular chevron buttons with a light border, translucent surface, and backdrop blur.
- Chevron placement uses the rendered sticky-pane width, keeping right-side controls 12px from the actual visible chart edge across root font sizes.
- Off-screen controls show a muted `MMM yyyy - MMM yyyy` range after left chevrons and before right chevrons.
- The selected objective panel is widened to 34rem and uses an opaque sidebar surface so timeline lines cannot bleed through.
- Stored objective HTML renders through the shared Box `html` API with the existing rich-text prose styles.
- Timeline bars reuse the Maya chat pop-up surface colors and backdrop blur in both themes.
- Browser verification found no runtime warnings, errors, overflow, or broken primary interaction.

final result: passed

---

# Objectives Timeline Alignment and Surface Design QA

## Source and implementation evidence

- Row-height source crop: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_mcpv1X/Screenshot 2026-08-01 at 10.55.31 AM.png` (732 × 304)
- Panel-opacity source: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_yiBmU3/Screenshot 2026-08-01 at 10.41.04 AM.png` (2456 × 1334)
- Final timeline capture: `/tmp/fortyone-objectives-timeline-final.png` (1280 × 720)
- Final panel capture: `/tmp/fortyone-objective-panel-final.png` (1280 × 720)
- Focused row comparison: `/tmp/fortyone-objectives-row-comparison.png` (1464 × 304)
- Full panel comparison: `/tmp/fortyone-objectives-panel-comparison.png` (2560 × 720)
- Browser viewport: 1280 × 720 CSS pixels, dark mode.

## Findings and comparison history

- Earlier P1: the left objective row resolved from `h-14` to 45.5px while the chart row was hard-coded to 56px. Both now use the same `3.5rem` CSS length, and the final capture shows aligned row boundaries.
- Earlier P1: the translucent details panel allowed timeline grid lines to show through its content. The panel and sticky header now use an opaque `bg-sidebar` surface with no background alpha.
- Earlier P2: key-result dots made the timeline bar feel visually noisy. Timeline KR markers were removed; key results remain in the details panel.
- Earlier P2: off-screen names consumed horizontal space. They were replaced with circular chevron buttons using shared icons, light borders, and `backdrop-blur-2xl`.
- The right chevron now measures the rendered sticky pane instead of assuming a 576px width, so its anchor follows the real browser-side edge of the chart.
- Earlier P2: header control heights, radii, and order differed. Zoom now leads, followed by a low-opacity separator, Today, the 2.1rem Timeline/List switcher, and New Objective.
- The objective bar and chevron controls reuse the Maya chat pop-up surface tokens: `bg-white/80`, `dark:bg-surface/80`, matching borders, and `backdrop-blur-2xl`.

## Required fidelity surfaces

- Typography: existing FortyOne type scale and weights are unchanged; rich objective HTML retains the established prose styles.
- Spacing and layout: fixed and timeline rows share one CSS height; controls share one height and rounded-xl treatment.
- Colors and tokens: floating timeline elements use existing Maya surfaces; the objective panel uses the opaque sidebar token.
- Image and icon quality: shared `ChevronLeftIcon` and `ChevronRightIcon` are used; no custom-drawn assets were introduced.
- Copy and content: objective names are removed from off-screen locators without changing navigation behavior.

## Functional verification

- Left chevron navigation scrolled to the off-screen objective.
- Today recentered the timeline.
- Objective selection opened the opaque details panel, and the close control dismissed it.
- Browser console contained no warnings or errors.
- The temporary unauthenticated QA route was removed after verification.

final result: passed

---

# Objectives Timeline Controls Design QA

## Reference

- User screenshot: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_w1GxLF/Screenshot 2026-08-01 at 9.45.26 AM.png`

## Review

- The Owner, Status, Objective, and Duration pane is widened by 2rem, from 34rem to 36rem.
- Today and Zoom are in the page header immediately before the Timeline/List switcher.
- The timeline column labels have a dedicated header row and are no longer crowded by controls.
- Timeline/List remains available, as requested.
- Today recenters the timeline and all three zoom options remain interactive.
- The authenticated production route was not bypassed; the temporary preview route was removed after verification.
- Browser verification found no runtime errors or visible overflow at the tested desktop viewport.

## Result

final result: passed

---

# Objectives Today Marker Design QA

## Source and implementation evidence

- Source visual truth: `/Users/joseph/Downloads/Screenshot 2026-08-01 at 10.03.02 AM.png` (2048 × 1303 pixels).
- Implementation screenshot: unavailable because the available in-app browser session redirects the objectives route to login.
- Intended implementation viewport: desktop, dark mode, authenticated objectives timeline.
- Density normalization: not applicable because an implementation capture could not be produced.
- State: objectives timeline centered on the current date, with the details panel closed.

## Comparison evidence

- Full-view comparison: blocked; the source screenshot was opened at original resolution, but the authenticated implementation state was unavailable.
- Focused marker comparison: blocked for the same authentication reason.
- Primary interactions and browser console: not tested because the implementation route could not be reached past login.

## Findings

- No code-level blocker remains. The main navigation contains no duplicate Zoom or Today controls.
- The timeline now uses a thin peach current-day line, an `MMM d` peach label, an 80% line opacity, and an 8px label radius.
- The marker is rendered below the objective details panel, whose opaque surface prevents the line from bleeding through.
- The objective panel top edge moved down one spacing step while retaining its bottom alignment and scroll clearance above Ask Maya AI.
- The new-objective short summary now matches the description font size and uses a shorter two-row input.

## Comparison history

- No visual iteration was possible because the implementation capture was blocked by authentication.

## Implementation checklist

- [x] Remove duplicate controls from the main page header.
- [x] Add an accurately positioned current-day marker for week, month, and quarter zoom levels.
- [x] Apply the requested peach color, 80% line opacity, taller label, and 8px radius.
- [x] Keep the marker behind the objective details panel.
- [x] Move property values slightly right.
- [x] Match short-summary and description text sizing while reducing summary height.
- [ ] Re-run the visual comparison in an authenticated browser session.

final result: blocked

---

# Sidebar Team Menu Design QA

## Source and implementation evidence

- Source visual truth: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_KhOx7i/Screenshot 2026-08-08 at 9.37.47 AM.png` (948 × 576 pixels, dark mode).
- Implementation screenshot: unavailable because the connected in-app browser opens `http://localhost:3000/` at `Login - FortyOne`.
- Intended viewport and state: authenticated desktop Projects view in dark mode with the Your Teams menu open and seven joined teams available.
- CSS viewport, implementation pixel size, and density normalization: unavailable without the authenticated implementation state.

## Comparison evidence

- Full-view comparison: blocked because an authenticated implementation capture could not be produced.
- Focused menu comparison: blocked for the same reason; a combined source-and-implementation comparison input could not be created.
- A focused source review identified two visible targets: horizontal group insets around every menu item and joined/public team rows immediately below the management actions.

## Functional verification

- Manage Teams, Create new team, joined-team rows, and public-team rows now render inside the shared padded `Command.Group` component.
- Joined teams are no longer filtered out because they are already visible in the sidebar; all joined teams are listed under Your teams.
- Public teams remain available under Join a team, and a stable empty message renders when neither collection contains a team.
- The local `first` workspace now contains six deterministic seeded teams and six matching memberships for Joseph Mukorivo; the seed was rerun without creating duplicates.
- Focused ESLint and Projects type checking pass. React Doctor reports only findings in pre-existing or concurrently changed components outside this focused menu change.
- Focused Jest remains blocked before test discovery by the repository's Jest 29/TypeScript 7 `fileExists` configuration failure.

## Required fidelity surfaces

- Typography: existing FortyOne `Text` styles and menu heading weight are retained.
- Spacing and layout: the shared `Command.Group` horizontal padding now applies to every actionable row; no menu group opts out with `px-0`.
- Colors and visual tokens: existing warm menu surfaces, hover colors, borders, and team colors are unchanged.
- Image and icon quality: existing Hugeicons-based team, plus, and chevron icons are reused at stroke width 2.
- Copy and content: management actions remain first, followed by Your teams and Join a team sections.

## Comparison history

- Earlier P1: the screenshot showed no team rows below the management actions. The menu now renders complete joined and public team collections.
- Earlier P2: menu action rows reached the container edges. All rows now inherit the shared command-group inset.
- Post-fix visual evidence remains unavailable until the local app is authenticated in the connected browser.

## Implementation checklist

- [x] Restore padded groups around every menu section.
- [x] Display every joined team below the management actions.
- [x] Continue displaying joinable public teams.
- [x] Seed six joined teams for overflow testing.
- [ ] Capture and compare the authenticated open-menu state at 948 × 576.

final result: blocked

---

# Documents Related Work Sidebar and Header Design QA

## Source and implementation evidence

- Open Related Work reference: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 10.47.25 PM.png` (1022 × 1472 pixels).
- Collapsed Related Work reference: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 10.47.37 PM.png` (958 × 1408 pixels).
- Document header reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_ChBSdF/Screenshot 2026-08-06 at 10.50.41 PM.png` (2026 × 378 pixels).
- Implementation screenshot: unavailable because the connected in-app browser opens `http://localhost:3000/` at `Login - FortyOne`; a connected Chrome session was unavailable.
- Intended viewport and state: authenticated desktop Documents editor in dark mode, with Related Work open and collapsed.

## Comparison evidence

- Full-view comparison: blocked because an authenticated implementation capture could not be produced.
- Focused header, sidebar, resize-divider, toggle, and empty-state comparison: blocked for the same reason.
- The source references were opened at original resolution and used to define the implementation structure, copy, and proportions.

## Functional verification

- Related Work now uses the same resizable divided-panel primitive as the sprint page, with a 30% default sidebar and 25–40% resize bounds.
- The open sidebar has a close control, full-width relationship action, centered empty state, and functional relationship list/remove actions.
- The collapsed state exposes an icon-only edge control with a `Show related work` tooltip.
- The document header groups the title and working overflow menu on the left and keeps document access on the right; unsupported Pin and Favorite controls were not added as static UI.
- Focused formatting and ESLint pass with zero warnings, Projects type checking passes, `git diff --check` passes, and React Doctor reports 92/100 with advisory maintainability findings only.

## Required fidelity surfaces

- Typography: existing FortyOne type tokens are retained; the header title uses the larger semibold hierarchy from the reference and Related Work uses normal product font sizes.
- Spacing and layout: the right rail uses sprint-page resize proportions, six-unit panel gutters, a full-width action, and a vertically centered empty state.
- Colors and visual tokens: the sidebar uses the app surface and existing half-pixel resizable divider instead of the previous translucent floating overlay.
- Image and icon quality: existing Hugeicons-based document, relationship, and close icons are reused; no generated or approximate assets were introduced.
- Copy and content: the empty state follows the supplied `Nothing to see here!` structure while using FortyOne's `document` terminology.

## Comparison history

- Earlier P1: Related Work was a floating overlay that covered the editor and did not match the sprint-page layout. It now participates in the resizable document layout.
- Earlier P2: the relationship toggle occupied the top navigation. It now appears as the collapsed edge control shown in the reference.
- Earlier P2: the top navigation separated the title from its overflow menu. They are now grouped together, while access remains right-aligned.
- Post-fix visual evidence remains unavailable until the local app is authenticated in the connected browser.

## Implementation checklist

- [x] Replace the floating Related Work overlay with a resizable sidebar.
- [x] Match the supplied open and collapsed sidebar hierarchy.
- [x] Move the collapsed toggle to the editor edge.
- [x] Reorganize the top navigation using only working controls.
- [ ] Capture and compare the authenticated open and collapsed states.

final result: blocked

---

# Documents Editor Media and List Refinement Design QA

## Source and implementation evidence

- Table-density and misplaced-controls source: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 8.09.48 PM.png`.
- Shortcut Documents references: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.22.33 PM.png`, `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.22.43 PM.png`, and `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.23.07 PM.png`.
- Linear slash-command reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_JdtS3R/Screenshot 2026-08-05 at 11.17.44 PM.png`.
- Implementation screenshot: unavailable because the connected in-app browser opened `http://localhost:3000` at `Login - FortyOne`.
- Intended state: authenticated desktop Documents home and editor, with a table selected, Related Work open, and uploaded media selected.

## Comparison evidence

- Full-view and focused side-by-side comparisons are blocked until an authenticated implementation state is available.
- The available visual references were inspected before implementation; no temporary authentication bypass or preview route was added.

## Functional and source verification

- Slash commands use `CodeXmlIcon` and `LayoutTable01Icon`; Diagram, GIF, and Attach file are removed.
- Table cells use compact vertical padding, prose margins are reset inside cells, and table actions are anchored to the selected table in the document scroll container without a nested portal.
- Related Work is closed by default and opens as a Maya-style floating translucent panel over a subtle backdrop without reducing editor width.
- Document access uses the requested keyhole lock for Only me and a two-person outline icon for workspace/member access.
- Images and MP4 videos support picker, paste, and drop uploads. Images persist width and alignment, support resize controls, and swap optimistic blob previews for access-checked stable document-media URLs.
- Template-card hover changes only the border. Empty states have no card or icon well, and document results render directly beneath a divider with zero horizontal row padding and neutral title hover.
- Focused Projects type checking and ESLint, Prettier, UI package type checking/builds, focused Go document/attachment tests, `gofmt -d`, and `git diff --check` pass.
- React Doctor scores the changed React surface 92/100 with no errors; remaining findings are advisory component-organization, exact icon-vector precision, and animation-token warnings.

## Result

final result: blocked

---

# Documents Landing, Templates, and Editor Design QA

## Source and implementation evidence

- Shortcut recent-documents reference: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.22.33 PM.png` (2048 × 1159 pixels).
- Shortcut my-documents reference: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.22.43 PM.png` (2048 × 1159 pixels).
- Shortcut document-and-access reference: `/Users/joseph/Downloads/Screenshot 2026-08-06 at 7.23.07 PM.png` (2048 × 1159 pixels).
- Linear slash-command reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_JdtS3R/Screenshot 2026-08-05 at 11.17.44 PM.png`.
- Implementation screenshot: unavailable because `http://localhost:3000/fin-kenya/docs` redirected the connected in-app browser to `Login - FortyOne`.
- Intended viewport and state: authenticated desktop Documents home and editor in dark mode, with recent documents and a populated template document.

## Comparison evidence

- Full Documents-home comparison: blocked because the authenticated implementation state could not be reached.
- Focused template-card, document-table, slash-menu, and table-editor comparisons: blocked for the same authentication reason.
- The three source screenshots were inspected at original resolution before implementation. A side-by-side comparison input could not be produced without a post-change implementation capture.

## Functional verification

- The Documents header now reuses the Notifications expandable-search interaction; the shared input uses normal body text sizing across Notifications, Feedback, Intake, and Documents.
- Recent documents, My documents, Shared with me, and Templates are URL-addressable views, and the sidebar retains quick links to recently updated documents.
- Blank, Meeting notes, Project brief, One-to-one, and Weekly update cards create real workspace-visible documents with trusted TipTap-compatible HTML and searchable text in one API operation.
- My and shared views render a document table with last-updated, access, visible related-work count, and owner-only archive actions.
- Visible relationship counts reuse team-membership visibility rules, and Shared with me now requires an explicit document-member relationship.
- Workspace guests remain read-only across API responses and UI controls; guests selected for restricted access are stored as viewers.
- The editor supports headings, inline formatting, links, quotes, lists, checklists, URL media, attachments, code, dividers, resizable tables, and contextual row/column/table actions.
- Slash-menu teardown is nullable and idempotent, and Escape exits the dedicated TipTap suggestion plugin rather than only hiding its popover.
- Projects type checking, focused ESLint, the Next.js 16.3 Turbopack production build, and focused Go document tests pass.
- React Doctor reports no Documents warning; its remaining warning is an unrelated existing memoization finding in story details.

## Required fidelity surfaces

- Layout: the compact header moves scope navigation upward, while template cards and the bordered table use the same dense split-pane hierarchy as the Shortcut references.
- Typography: document search, scope items, slash commands, cards, table headers, and empty states use the normal application text scale rather than auxiliary small labels.
- Buttons and surfaces: search and create actions reuse Notifications button sizing and standard radius; template cards and editor popovers use opaque product surfaces and established border tokens.
- Icons: Documents and sharing use the requested Hugeicons outline glyphs, and the remaining controls reuse the shared icon package without custom-drawn assets.
- Interaction: search, scopes, templates, table rows, archive menus, slash commands, and table actions are implemented as working controls.

## Result

## final result: blocked

# Objectives Board and Timeline Hover Design QA

## Source and implementation evidence

- Source visual truth: `/Users/joseph/Downloads/Screenshot 2026-08-01 at 12.53.47 PM.png` (962 × 1230 pixels).
- Implementation screenshot: unavailable because browser control is not exposed in the current session and the objectives route requires authentication.
- Intended implementation viewport: authenticated desktop objectives page in dark mode.
- State: Timeline view with the pointer over a date column; Board and List views grouped by status.

## Comparison evidence

- Full-view comparison: blocked because an authenticated implementation capture could not be produced.
- Focused hover-marker comparison: blocked for the same reason.
- Primary interactions and browser console: not tested in a rendered browser.

## Findings

- The shared Stories layout switcher now presents `Board` instead of `Kanban` without changing its persisted layout value.
- Objectives now supports Timeline, Board, and List layouts.
- Objective Board and List views reuse the Stories board dimensions, grouped-header rhythm, card surfaces, controls, and empty-group behavior.
- Board groups can be changed between Status, Lead, and Priority; objectives can be reordered and dragged between compatible groups.
- Timeline pointer movement now renders a neutral full-height guide and an `MMM d` date label independently of the peach Today marker.
- The guide remains below the objective details panel in the stacking order.

## Comparison history

- No visual iteration was possible because the implementation capture was blocked.

## Implementation checklist

- [x] Rename user-facing Kanban labels to Board.
- [x] Add grouped Objective Board and List layouts.
- [x] Add status, lead, and priority grouping controls for Board and List.
- [x] Add drag-to-update grouping behavior in Board view.
- [x] Add a pointer-following timeline date guide.
- [ ] Re-run the visual comparison in an authenticated browser session.

final result: blocked

---

# Duplicate Feedback and Story Similarity Design QA

## Source and implementation evidence

- Source visual truth: `/Users/joseph/Downloads/Screenshot 2026-08-05 at 11.53.16 AM.png` (1910 × 1304 pixels, light mode) and `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_VwyO1j/Screenshot 2026-08-05 at 11.54.43 AM.png` (1872 × 896 pixels, dark mode), together with the requested row and surface corrections.
- Implementation screenshot: unavailable because `http://localhost:3000` redirected the connected in-app browser to `Login - FortyOne`; Chrome browser control was unavailable.
- Intended viewport: authenticated desktop create-story and create-feedback dialogs in light and dark mode.
- Source CSS size and density: unavailable from the OS captures; source pixel dimensions are recorded above. Implementation size and density normalization could not be measured without the authenticated state.
- State: a title of at least 10 characters has produced one or more high-relevance similar records.

## Comparison evidence

- Full-view comparison: blocked before implementation capture because authentication was unavailable.
- Focused panel comparison: blocked for the same reason.
- The source screenshots were opened at original resolution. A side-by-side comparison input could not be produced because the post-change implementation capture is unavailable.
- Code-level inspection confirms that the result panel now shares the dialog surface token and backdrop blur, uses a taller lighter header, stronger dialog-divider borders, and compact read-only product rows.

## Functional verification

- Similarity requests remain debounced but do not run before 10 trimmed title characters.
- Generic action words are removed before ranking and the normalized-title display threshold is raised from `0.2` to `0.45`.
- A read-only database score check returned `1.000` for “Create a monthly feedback digest”, `0.333` for “AI Feedback”, and `0.067` for “Create backlog”; only the first clears the display threshold.
- Story results include reference, title, assignee, status, and priority and remain advisory links.
- Feedback results include title, creator, status, and a working optimistic upvote control. High-confidence duplicates still redirect to the existing record.
- No checking, no-match, or error panel is displayed; the panel exists only when result rows exist.
- The full Go suite, Projects type checking, focused ESLint, formatting, and React Doctor completed successfully. Focused Jest execution remains blocked by the repository's Jest 29/TypeScript 7 config loader incompatibility before tests load.

## Required fidelity surfaces

- Typography: the heading increases from `text-xs` to `text-sm`, retains the established uppercase tracking, and result text uses the existing `Text` component weights.
- Spacing and layout: the heading has a 4rem minimum height; result rows use a compact 4rem minimum height with responsive metadata labels.
- Colors and visual tokens: the panel uses `bg-surface-elevated`, `border-border-strong`, a lighter `bg-surface-muted/70` heading, and `backdrop-blur-xl`, including the dialog's translucent dark surface.
- Image and icon quality: existing avatar, status, priority, vote, and chevron components are reused; no replacement raster or custom icon asset was introduced.
- Copy and content: story and feedback rows now show the requested product metadata without “Review story”, checking, or empty-result copy.

## Comparison history

- Earlier P1: generic verbs caused visibly unrelated candidates such as “Create backlog”. The normalized threshold and stop-word ranking now exclude that result in the database score check.
- Earlier P2: the panel header was too short and visually merged into the body. It now uses a taller, lighter, bordered header.
- Earlier P2: result rows looked like bespoke suggestion cards and omitted operational metadata. They now reuse the story-row visual vocabulary and public-feedback status/vote components.
- Post-fix browser evidence remains blocked by authentication, so these fixes cannot be visually passed yet.

## Result

final result: blocked

---

# Docs Workspace and Editor Design QA

## Source and implementation evidence

- Shortcut Docs split-pane reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_B3ZWeM/Screenshot 2026-08-05 at 11.17.10 PM.png`
- Linear slash-command reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_JdtS3R/Screenshot 2026-08-05 at 11.17.44 PM.png`
- Linear compact selection-toolbar reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_zwGzOl/Screenshot 2026-08-05 at 11.18.08 PM.png`
- Implementation screenshot: unavailable because the connected in-app browser opens the local app at the FortyOne sign-in screen.
- Intended viewport: authenticated desktop Docs route in dark mode.

## Comparison evidence

- Full-view comparison: blocked before implementation capture because authentication was unavailable.
- Focused slash-menu and selection-toolbar comparisons: blocked for the same reason.
- The local Docker daemon required to start a complete seeded API stack was unavailable, so a temporary authenticated environment could not be created.

## Functional verification

- The production build includes both `/[workspaceSlug]/docs` and `/[workspaceSlug]/docs/[documentId]`.
- Focused TypeScript and ESLint checks pass for the Docs UI and affected story/objective surfaces.
- The full Go test suite passes, including document service coverage.
- React Doctor reports no new Docs warning; its remaining finding is pre-existing in story detail.
- The editor toolbar uses a solid elevated surface and the stronger dark-mode border token.

## Implementation checklist

- [x] Add Docs after Strategy Map in workspace navigation.
- [x] Add a searchable split-pane document workspace.
- [x] Persist workspace, restricted, and private document access.
- [x] Add bidirectional Story and Objective relationships.
- [x] Add Linear-inspired slash commands.
- [x] Compact and strengthen the text-selection toolbar.
- [ ] Re-run the visual comparison in an authenticated browser session.

final result: blocked
