**Comparison Target**

- Source visual truth:
  - `/Users/joseph/Downloads/Screenshot 2026-08-22 at 6.49.23 PM.png` (`746 x 1426` pixels)
  - `/Users/joseph/Downloads/Screenshot 2026-08-22 at 6.49.31 PM.png` (`626 x 1108` pixels)
  - `/Users/joseph/Downloads/Screenshot 2026-08-22 at 6.49.41 PM.png` (`420 x 1420` pixels)
  - `/Users/joseph/Downloads/Screenshot 2026-08-22 at 6.49.49 PM.png` (`570 x 1358` pixels)
- Intended implementation: FortyOne Projects desktop application shell with the edge control positioned above center, a prominent resting bar that morphs into a directional arrow, short Expand/Collapse tooltips, and expanded/collapsed states.
- Implementation screenshot: unavailable. The local Projects preview at `http://127.0.0.1:3000/` opened the sign-in screen rather than the authenticated application shell.
- Intended viewport: the active desktop Projects viewport, matched across expanded, collapsed, resting, and hovered states.
- Density normalization: not performed because an authenticated implementation capture was unavailable.

**State**

- Source: Stripe dashboard sidebar edge toggle in expanded/resting, expanded/hovered, collapsed/resting, and collapsed/hovered states.
- Browser result: unauthenticated FortyOne sign-in screen.
- Primary interactions tested: none in the rendered application shell because it was not reachable without signing in.
- Console errors checked: not applicable to the unreachable authenticated shell.

**Full-view Comparison Evidence**

- Blocked. All four source references were opened successfully, but no same-state authenticated implementation screenshot could be captured.

**Focused Region Comparison Evidence**

- Blocked. The sidebar edge, resting bar, directional hover icon, tooltip, and width transition require the authenticated application shell for a valid comparison.

**Findings**

- [P1] Authenticated implementation capture unavailable
  Location: local Projects preview.
  Evidence: the source shows four sidebar interaction states; the browser opens the FortyOne sign-in screen.
  Impact: visual fidelity and the live expanded/collapsed interaction cannot be verified against the references.
  Fix: open the local Projects application in an authenticated in-app browser session and capture the same four states.

**Open Questions**

- None about implementation intent. The remaining blocker is preview authentication.

**Implementation Checklist**

- Capture the expanded resting state.
- Hover or keyboard-focus the control and capture the collapse direction and tooltip.
- Toggle the sidebar and capture the collapsed resting state.
- Hover or keyboard-focus the control and capture the expand direction and tooltip.
- Verify typography, spacing, colors, icon fidelity, copy, focus treatment, and width transition against the source references.

**Comparison History**

- Pass 1: blocked before visual comparison because the local browser session was unauthenticated. No visual fixes were made from browser evidence.
- Pass 2: source implementation was refined from user feedback, but post-fix capture remains blocked by the unauthenticated local browser session.
- Pass 3: the two-stroke rotation and fade was replaced with a single thinner path that bends into a shallower angle; authenticated post-fix capture remains unavailable.
- Pass 4: the morph control now uses the semantic primary color and a slightly taller icon.
- Pass 5: the original command-bar toggle was removed, leaving the edge morph control as the only sidebar toggle; authenticated post-fix capture remains unavailable.

**Follow-up Polish**

- None classified until authenticated captures are available.

final result: blocked

---

# Pricing Page — Calendly Full-page Reference

**Comparison Target**

- Source visual truth:
  - `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_Vzv9Am/Screenshot 2026-08-23 at 11.41.33 PM.png` (`2950 x 1664` pixels).
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/calendly-desktop-00.jpg` through `calendly-desktop-25.jpg`, captured from the current public page in small scroll increments.
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/calendly-pricing-mobile-reference.jpg` (`390 x 844` CSS-pixel viewport).
- Intended implementation: the FortyOne `/pricing` route, retaining FortyOne content, colors, controls, and icon library while matching the reference page hierarchy and interaction patterns.
- Implementation screenshots:
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/fortyone-pricing-final-desktop-top.jpg`
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/fortyone-pricing-final-desktop-cards.jpg`
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/fortyone-pricing-final-card-bottoms.jpg`
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/fortyone-pricing-final-comparison.jpg`
  - `/Users/joseph/.codex/visualizations/2026/08/23/01a03081-6730-7d13-bdd9-355045e88c59/fortyone-pricing-mobile-final.jpg`
- Viewports: desktop `1280 x 720`; mobile `390 x 844`.

**State**

- Route: `/pricing`.
- Themes: light and dark.
- Billing: annual for visual comparison; desktop monthly radio and mobile billing switch were exercised interactively.
- Primary interactions tested: billing changes update Professional from `$5.60` to `$7` and Business from `$8` to `$10`; the comparison plan dock stays visible during long-table scrolling; FAQ rows expand and collapse.

**Full-view Comparison Evidence**

- The implementation preserves the existing FortyOne navigation and adds the pricing hero inside the established rounded landing shell.
- Four equal-width cards use the reference's contrasting outer body, inset summary panel, serif price, aligned long feature body, and elevated popular-plan frame.
- Desktop keeps the normal-weight description under the heading and native billing radios. Mobile follows the source's compact centered heading, single billing switch, and one-column card flow.
- Card CTAs use FortyOne's shared `size="lg"` dimensions at both breakpoints.

**Focused Region Comparison Evidence**

- Calendly and FortyOne card-body captures were opened together at the same desktop viewport. Both use icon-led feature groups, full-contrast item text, fine dividers, aligned card endings, and a tall popular-plan rail.
- FortyOne category icons come from the existing icon package and use semantic `primary`, `secondary`, `info`, and `success` surfaces instead of copied Calendly assets.
- The detailed comparison follows the source's serif section heading, repeated plan controls, alternating feature rows, category icon headings, compact included markers, and long grouped data. Its measured fixed dock avoids the repository's known overflow/sticky failure mode.
- The pricing FAQ reuses FortyOne's existing content in the reference's centered rounded shell and left-side expand icon treatment.

**Findings**

- [P3] Product content intentionally differs from the reference
  Location: plan names, prices, descriptions, and feature lists.
  Evidence: the source describes Calendly scheduling products; the implementation retains FortyOne's Hobby, Professional, Business, and Enterprise plans.
  Impact: text lengths and line wrapping differ while the visual hierarchy remains aligned.
  Resolution: accepted product constraint.

- [P3] Popular frame uses the FortyOne palette
  Location: Business plan perimeter.
  Evidence: the reference uses Calendly's blue-to-multicolor rail; the implementation uses FortyOne secondary, info, and primary theme colors.
  Impact: brand color differs by design without changing card anatomy.
  Resolution: explicitly requested and accepted.

- [P3] FortyOne has fewer add-on states than Calendly
  Location: lower pricing-card feature groups.
  Evidence: Calendly shows unavailable and optional Notetaker/Callie groups; FortyOne's current plan data has no equivalent add-on states.
  Impact: FortyOne card copy is shorter, so shared minimum height is used to preserve the long aligned silhouette without inventing unavailable product data.
  Resolution: accepted product-data constraint.

**Comparison History**

- Pass 1: introduced the rounded pricing shell, inset summary cards, serif pricing, arrow CTAs, and a solid popular-plan frame.
- Pass 2: reduced the CTAs to the shared FortyOne button size, replaced hard-coded card colors with theme surfaces, and rebuilt the popular frame as a theme-token gradient.
- Pass 3: reviewed the normalized reference and implementation together; no P1 or P2 visual mismatch remained.
- Pass 4: set card CTAs to the shared `size="lg"` variant and strengthened semantic surface contrast in both themes.
- Pass 5: reduced the summary-panel height, moved the page description below the heading, and replaced the segmented billing control with Calendly-style native radios and a savings badge.
- Pass 6: reduced the feature-body horizontal padding from `1.5rem` to `1.25rem` on desktop and increased the savings badge height from `1.25rem` to `1.5rem`.
- Pass 7: replaced flat feature lists with categorized groups, real FortyOne icons, full-contrast item text, fine dividers, and taller aligned card bodies.
- Pass 8: replaced the short limits/features table with the full category-based comparison and fixed plan dock.
- Pass 9: matched the source mobile hierarchy with a centered compact heading, single billing switch, one-column cards, and horizontally scrollable comparison data.
- Pass 10: compared source and implementation captures together at desktop and mobile sizes, verified light/dark contrast, exercised billing and dock interactions, and found no P1 or P2 mismatch.
- Pass 11: lightened the light-mode card body from `accent` to `surface-muted`, changed the inset summary to `surface`, and replaced the heavier default elevation with the low-opacity semantic `shadow` token.
- Pass 12: replaced generic shared icons with a pricing-only nine-icon set covering work, access, teams, planning, organization, administration, support, deployment, and scale in one consistent rounded-stroke language.
- Pass 13: simplified the weakest category icons into cleaner task, hierarchy, key, chat, deployment-stack, and growth marks with lighter strokes at card scale.
- Pass 14: narrowed the comparison table, enlarged and de-emphasized its plan CTAs, replaced alternating row fills with subtle continuous plan columns, and changed boolean values to compact rounded-square checks.
- Pass 15: removed the pricing FAQ so the detailed comparison flows directly into the final CTA.

**Follow-up Polish**

- None required for the implemented source-derived page scope.

final result: passed

---

# Landing Footer — Calendly Layout Reference

**Comparison Target**

- Source visual truth:
  - `/Users/joseph/Downloads/Screenshot 2026-08-24 at 12.06.49 AM.png` (`3004 x 1684` pixels).
  - `/Users/joseph/Downloads/Screenshot 2026-08-24 at 12.07.00 AM.png` (`2972 x 1688` pixels).
- Implementation screenshots:
  - `/tmp/fortyone-footer-final-desktop.png` (`1280 x 720` pixels), top composition.
  - `/tmp/fortyone-footer-in-context.png` (`1280 x 720` pixels), utility and legal rows in the homepage.
- Comparison artifact: `/tmp/fortyone-footer-final-comparison.png`.
- Viewport: `1280 x 720` CSS pixels in the in-app browser.
- Density normalization: the source and implementation top regions were each normalized to `640` pixels wide for the combined comparison.

**State**

- Route: `/`, light theme.
- Primary interactions tested: light and dark theme controls; the selected option updates `aria-pressed` and the root theme class.
- Console checked: no active error dialog remained after the current landing and pricing routes compiled.
- Responsive check: `390 x 844` CSS pixels with no horizontal overflow.

**Full-view Comparison Evidence**

- The footer uses the reference's inset rounded frame, oversized left statement, three balanced link columns, low wordmark, divider, social controls, and quiet final utility row.
- FortyOne's existing theme-aware hero gradient, semantic text colors, wordmark, icons, routes, and appearance behavior replace the reference brand treatment by design.
- The homepage hero and pricing hero now share the same larger `3rem` mobile and `4rem` desktop panel radii as the footer.

**Focused Region Comparison Evidence**

- The reference and implementation top regions were reviewed together in `/tmp/fortyone-footer-final-comparison.png`.
- The final homepage footer utility capture confirms the compact Appearance pill, single-line copyright, and Legal-column-only privacy and terms links.
- The footer wordmark was reduced to `1.75rem` on mobile and `2rem` from the small breakpoint.

**Findings**

- [P3] Palette intentionally differs from the reference
  Location: footer frame and text.
  Evidence: the source uses Calendly navy; the implementation reuses FortyOne's subtle hero gradient and semantic light/dark tokens.
  Impact: brand color differs while the spatial composition remains aligned.
  Resolution: explicitly requested and accepted.

- [P3] Product content intentionally differs from the reference
  Location: link groups and utility controls.
  Evidence: the implementation preserves FortyOne routes and its theme control instead of Calendly downloads and language controls.
  Impact: exact column lengths differ without weakening hierarchy.
  Resolution: accepted product constraint.

**Comparison History**

- Pass 1: translated the reference into a dark brand-ink card with the existing FortyOne destinations and controls.
- Pass 2: replaced the long mobile stack with a two-column directory and verified no horizontal overflow.
- Pass 3: switched from the reference's dark palette to the existing theme-aware hero gradient, removed the frame border, and increased the footer, homepage hero, and pricing hero radii.
- Pass 4: restored the compact Appearance switch, moved copyright into the final legal row, removed duplicate privacy and terms links, reduced the wordmark, and changed the statement to `Keep what matters moving.`
- Pass 5: verified the final footer on the real homepage, the pricing route, the shared `62px` rendered desktop radius, theme switching, link uniqueness, and responsive overflow.

**Follow-up Polish**

- None required for the approved direction.

final result: passed
