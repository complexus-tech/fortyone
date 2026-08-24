# Design QA: Decide What Matters Showcase

Date: 2026-08-24

## Scope

- Replacement for the previous `HowItWorks` section.
- Reference: `Screenshot 2026-08-24 at 7.08.32 AM.png`.
- Desktop comparison viewport: 1502 x 894.
- Mobile verification viewport: 390 x 844.

## Result

No blocking visual, responsive, or accessibility findings remain in the scoped
section.

## Comparison notes

- Preserved the reference structure: concise introduction, three portrait media
  cards, centered product UI, and supporting copy beneath each card.
- Applied the requested FortyOne typography scale and muted text tokens instead
  of the reference site's serif display treatment.
- Removed the reference's left rule and the added group-level product-card
  surface. The real product UI surfaces remain legible over the imagery.
- Used the texture tool's built-in Risograph image consistently across all three
  cards, with no CSS blur.
- Standardized all three product mockups to the first mockup's width.
- Retired the previous `HowItWorks` section after the replacement was approved.

## Responsive and accessibility checks

- Desktop: three equal media columns; no clipping or awkward copy wrapping.
- Tablet: two-column layout with the third card centered on its own row.
- Mobile: single-column layout; the goal label stays on one line and the product
  mockup remains readable.
- Texture images are decorative (`alt=""`); static product mockups are marked
  `aria-hidden` to avoid duplicating the card descriptions.
- Section and card headings use semantic `h2` and `h3` elements.

## Evidence

- `reference-vs-risograph.png`
- `implementation-desktop-risograph.png`
- `implementation-mobile-risograph.png`
