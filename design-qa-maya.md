# Maya empty-state design QA

- Source visual truth: `/Users/joseph/Downloads/Screenshot 2026-09-04 at 10.01.56 AM.png`
- Implementation: `http://localhost:3000/unauthorized/maya-preview` (temporary read-only preview route removed after capture)
- Implementation screenshot: captured in the Codex in-app browser; the browser tool did not expose a filesystem path
- Source pixels: 2340 x 1528
- Desktop implementation pixels/CSS viewport: 1280 x 720 at device scale 1
- Responsive implementation pixels/CSS viewport: 390 x 844 at device scale 1
- State: dark theme, empty Maya conversation
- Normalization: compared the central assistant identity, composer, actions, and prompt-card region. The source includes unrelated product chrome and uses a larger viewport, so the comparison used proportional placement rather than raw pixel coordinates.

## Full-view comparison evidence

The implementation carries over the source hierarchy: quiet header actions, a personalized centered greeting, one dominant composer, and four compact prompt cards beneath it. FortyOne's existing typography, controls, and pricing gradient replace the source product's branding. At the user's request, the Maya icon and ambient background treatment were removed; the greeting and gradient edge now carry the hierarchy.

## Focused-region comparison evidence

- Typography: the personalized `Hi, {firstName}! Ask Maya anything.` greeting is the sole display heading, with a short muted capability line. UI copy remains in the app's existing font and optical scale.
- Layout rhythm: the composer is the widest and highest-contrast object; its width returns to the original `max-w-4xl`. Suggestions form a single four-column row on desktop, two columns at tablet widths, and one column on mobile.
- Colors and tokens: the composer uses a six-pixel frame with the pricing sequence `secondary -> info -> primary`, including the darker token mixes used by the pricing surface. The solid input surface has no blur or drop shadow. A separate, low-opacity ambient wash now sits on the page background. The frame and inner surface use larger radii, with a `2.75rem` squircle treatment in supporting browsers.
- Image and icon quality: there are no new raster assets or approximate custom icons. Existing FortyOne icons are used for assistant actions and suggested prompts.
- Copy: prompts remain tied to real Maya capabilities and retain workspace terminology rather than copying Brain2 labels. Supporting descriptions are intentionally constrained to one line.

## Findings

No actionable P0, P1, or P2 visual differences remain for the approved interpretation. The implementation is intentionally inspired by the source rather than a brand-identical clone.

## Comparison history

1. First desktop capture: the structure, gradient frame, and four-card row matched the intended hierarchy. The added Maya icon was visually unnecessary.
2. First mobile capture: the composer controls wrapped at 390 px, creating avoidable density.
3. Second desktop capture: removed the ambient page gradient and composer shadows, narrowed the composer, applied a tertiary Live Voice treatment, and shortened the prompt cards with bare side icons and single-line descriptions.
4. Third desktop capture: restored the original suggestion typography, moved bare icons above the copy, increased the greeting and description gap, restored the wider composer, and added squircle corners. The first capture exposed one-line descriptions escaping their cards; adding explicit width and overflow containment fixed the P2 issue.
5. Final desktop and mobile captures: each description now truncates within its own card. The personalized greeting wraps cleanly and all persistent composer controls remain visible at 390 px.
6. Composer-control refinement: promoted the Attach, Record, and Live Voice hover surface to their empty-state resting appearance, made Send fully round, and removed the helper line beneath the empty composer.
7. Prompt-copy refinement: retained the four bottom cards while shortening their labels and descriptions to the reference's direct, task-oriented style.
8. Width and frame refinement: moved the suggestion row into the composer's exact `max-w-4xl` parent, doubled the gradient frame to six pixels, and increased the corner radii to balance the heavier edge.
9. Page-flow refinement: converted the transparent Maya toolbar to an absolute overlay, added initial conversation clearance while allowing messages to scroll behind it, subtly restored the background-only gradient, increased the prompt-description type and gap, and raised the squircle radius again.
10. Prompt-card density refinement: increased prompt icons from 18 to 19 pixels while reducing vertical padding and minimum height for a more compact row.
11. Composer density refinement: reduced the empty-state input and surface height by eight pixels without changing its width, controls, frame thickness, or corner geometry.
12. Supporting-copy refinement: changed the empty-state line to `Plan what's next, find what matters, or move work forward.` for a clearer default-homepage proposition.

## Verification limits

The production Maya route requires an authenticated workspace session that was not available in the in-app browser. The visual comparison used a temporary route with the production CSS modules and prompt component; source-level tests and TypeScript cover the production component wiring. A signed-in conversation-state browser check remains separate from this empty-state QA.

## Final result

final result: passed
