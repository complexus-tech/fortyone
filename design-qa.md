# Walkthrough Maya Composer Design QA

- Source visual truth: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_yk2x0p/Screenshot 2026-08-31 at 4.14.47 PM.png`
- Implementation screenshot: `/Users/joseph/.codex/visualizations/2026/08/29/01a04f41-71b0-7e72-95ca-0cdc78eedbd2/maya-walkthrough-localhost-final.png`
- Preview URL: `http://localhost:3000/first/roadmap?view=gantt`
- Viewport: 1202 x 950 CSS pixels in the authenticated desktop in-app browser
- Source pixels: 1428 x 1400
- Implementation pixels: 1202 x 950 at the browser's default device scale
- State: quota-limited Maya walkthrough step with the composer open, spotlighted, focused, and interactive

## Full-view comparison evidence

The source and localhost implementation were inspected together. Both place the walkthrough panel directly above the Maya composer, keep the composer visible inside the dimmed application, use the orange focus outline on both regions, and retain the long action label on one line. The localhost capture uses a shorter viewport, so the panel is more compact while preserving the source composition and hierarchy.

## Focused region comparison evidence

The composer target resolves to `478 x 129.375` CSS pixels at `(705.125, 800.125)`. Its spotlight is `494 x 145.375` at `(697.125, 792.125)`, confirming the intended eight-pixel padding on every side. The panel shows `Try Maya with a real request [5 / 7]`, an inline `Shift + M` shortcut, and the enabled `Write my first Maya message` action.

## Findings

- A fresh transition initially exposed a stale fallback even though the composer was mounted and visible. React Compiler was reusing a DOM-derived measurement because the reducer revision used only to force rendering was discarded.
- The overlay now stores the resolved target and geometry as explicit React state, then refreshes it through mount, resize, intersection, scroll, and short stability checks.
- A clean Back -> Minimize Maya -> Next transition reopened Maya, focused the message input, displayed the real action, and produced no new browser warnings or errors.
- Historical hydration/sidebar and demo-avatar warnings in the long-running development session predate this focused transition and are outside this walkthrough change.

## Comparison history

- Initial pass: blocked at login before the authenticated localhost tab was available.
- First authenticated pass: correct after hot reload, but a fresh closed-chat transition remained on `Getting things ready…`.
- Final pass: fresh localhost transition resolves the composer immediately and matches the source interaction and layout.

## Primary interactions tested

1. Open Help and restart Product tour.
2. Start the guided setup and advance through the completed task, calendar, and objective steps.
3. Enter the Maya step from a closed chat rail.
4. Move Back, minimize Maya, and return with Next to verify repeatable target mounting and focus.
5. Confirm the quota-limited state still exposes the composer and actionable CTA. Prompt submission behavior is covered by the focused Maya hook tests without sending a live message during visual QA.

final result: passed
