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
