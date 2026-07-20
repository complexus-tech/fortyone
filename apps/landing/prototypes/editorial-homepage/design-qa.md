# Design QA

## Comparison target

- Source visual truth:
  - `/Users/joseph/development/complexus/fortyone/output/playwright/landing-reference-audit-2026-07-20/10-fortyone-home-desktop-current-hero.png`
  - `/Users/joseph/development/complexus/fortyone/output/playwright/landing-reference-audit-2026-07-20/02-userjot-home-desktop-loop.png`
  - `/Users/joseph/development/complexus/fortyone/output/playwright/landing-reference-audit-2026-07-20/09-complexus-home-desktop-testimonial.png`
- Browser-rendered implementation:
  - `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/40-final-desktop-hero.png`
  - `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/28-final-desktop-feedback.png`
  - `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/29-final-desktop-principle.png`
  - `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/38-final-mobile-hero-balanced.png`
- Viewports: `1440 × 1000`, `1280 × 800`, `390 × 844`, and compact-width verification at `320 × 700`.
- State: dark mode. Comparison captures use the prototype's `#qa-static` state to remove transition timing from the visual comparison. Interaction checks use the normal animated page.

## Comparison evidence

- Full-view hero and initial product-story comparison: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/35-comparison-hero.png`
- Focused feedback-story and testimonial-treatment comparison: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/36-comparison-story.png`
- Focused mobile image-legibility comparison: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/37-comparison-mobile.png`

## Findings

No actionable P0, P1, or P2 findings remain.

- Fonts and typography: the FortyOne hero and UI remain Inter with the existing weight and scale. Newsreader is limited to one green principle section and no longer competes with the product identity.
- Spacing and layout rhythm: the existing desktop hero split, 1400px product framing, restrained radii, and neutral surfaces remain intact. Complexus influences are limited to section spacing, editorial rails, thin rules, and one principle band.
- Colors and tokens: the prototype uses the existing FortyOne near-black, off-white, neutral border, and coral tokens. Green appears once; no additional accent system was introduced.
- Image quality and asset fidelity: the production FortyOne wordmark is used as the exact SVG asset. Product imagery uses real FortyOne screenshots rather than recreated UI, and the feedback capture has an accurate JPEG extension.
- Copy and content: the active FortyOne heading and shortened description remain unchanged. The product story explicitly connects feedback intake, planning, delivery, and roadmap progress.
- Responsiveness and accessibility: no horizontal overflow remains at 320, 390, 1280, or 1440px. The mobile hero keeps “built in.” together, feedback imagery uses a legible focused crop, anchor targets are present, images have descriptive alt text, focus styling and a skip link are present, and reduced-motion behavior is defined.

## Comparison history

1. Initial brand-fidelity pass — P1

   - Earlier finding: the first concept used an invented `41 + fortyone` lockup, a more expressive hero treatment, and broader color fields that made the page feel like a new brand.
   - Fix: replaced the lockup with the exact production wordmark, restored FortyOne's current header and split-hero grammar, returned screenshots to neutral frames, and limited color to coral details plus one green section.
   - Post-fix evidence: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/35-comparison-hero.png`.

2. Mobile screenshot-legibility pass — P2

   - Earlier finding: the feedback screenshot was scaled down until its content became unreadable on mobile.
   - Fix: used a focused 720px crop inside the mobile frame while preserving the desktop composition.
   - Before: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/12-mobile-feedback-before.png`.
   - Post-fix evidence: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/37-comparison-mobile.png`.

3. Final brand-guardrail pass — P1/P2

   - Earlier findings: the green statement was oversized, a decorative `41` read as a second logo, rail labels were over-styled, the footer introduced an extra black token, and the general text weight had drifted from FortyOne's baseline.
   - Fix: removed the decorative mark, capped the Newsreader statement at 56px desktop and 44px mobile, restored normally capitalized rail labels, returned the footer to the canvas token, and restored the 500 body weight.
   - Post-fix evidence: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/29-final-desktop-principle.png`.

4. Responsive and affordance pass — P2
   - Earlier findings: the mobile hero orphaned “in.”, reveal scaling could create a one-pixel root overflow, filter-heavy reveals were unnecessary on compact screens, static rows implied clickability on hover, the feedback asset had a mismatched file extension, and dark browser metadata was missing.
   - Fix: balanced and retuned the compact hero, moved screenshot scaling to a clipped inner frame, used transform-and-opacity reveals on mobile, removed hover affordances from static rows, corrected the asset extension, and added theme-color and favicon metadata.
   - Post-fix evidence: `/Users/joseph/.codex/visualizations/2026/07/20/019f7fcf-ed3e-7863-a16a-88d206f7b39b/editorial-homepage/38-final-mobile-hero-balanced.png`; browser measurements confirmed `scrollWidth === innerWidth` at both 320px and 390px.

## Interaction and runtime checks

- Hero `Get started free` navigates to `#system` and leaves the section heading below the fixed header.
- Header `Sign up` navigates to `#start`.
- Every internal anchor resolves to an existing target.
- All six rendered images loaded with non-zero natural dimensions.
- Desktop and mobile widths were checked for overflow.
- Browser console after page load and primary navigation checks: no errors.

## Follow-up polish

- P3: the live feedback screenshot includes test-like request copy. Before translating the concept into the production landing page, recapture the portal with a short, polished set of representative requests.

final result: passed
