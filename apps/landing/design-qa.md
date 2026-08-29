# AI Planning hero design QA

**Source visual truth**

- Current homepage hero at the matching desktop viewport: `/private/tmp/fortyone-ai-planning-hero/source-homepage-desktop.png`
- Earlier selected layout: `/Users/joseph/.codex/generated_images/01a030ae-fde1-7120-b472-1656078a49c6/exec-d0f2f33d-5f11-4447-9b48-471024018ace.png`
- Product media: `public/images/product/maya-delivery-brief-light.webp` and `public/images/product/maya-delivery-brief-dark.webp`

**Implementation evidence**

- Desktop: `/private/tmp/fortyone-ai-planning-hero/implementation-desktop-concise.png`
- Mobile: `/private/tmp/fortyone-ai-planning-hero/implementation-mobile-concise.png`
- Route: `http://localhost:3001/features/ai-planning`
- State: light theme, page top, hero loaded, no menus open.

**Viewport and normalization**

- Homepage style anchor: 1440 x 1051 px capture from a 1440 x 1100 CSS viewport at 1x density.
- Desktop implementation: 1440 x 1051 px capture from the same 1440 x 1100 CSS viewport at 1x density.
- Mobile implementation: 390 x 844 px capture from a 390 x 844 CSS viewport at 1x density.
- The homepage anchor and implementation were viewed together at the same desktop viewport. The earlier generated concept was used only for the already-approved image-led composition; the implementation intentionally preserves the supplied 2998 x 1788 media aspect ratio.

**Full-view comparison evidence**

- The heading, description, and two provider signup actions sit above the full-width Maya visual, matching the current homepage hero's vertical composition.
- The feature headline intentionally departs from the homepage's colored handwritten accents: it uses the normal heading font and wraps across exactly two desktop lines, as requested.
- The shell, typefaces, colors, radii, shadows, and browser framing come from the existing homepage and pricing-page primitives rather than approximated styles.
- Only the approved hero slice was implemented. The existing detail content remains below it until the supporting card layouts are supplied.

**Focused-region comparison evidence**

- Desktop heading: 651 x 116.25 CSS px, exactly two lines at 58.125 px line height. It contains no accent or handwritten child elements.
- Desktop actions: `Continue with Google` and `Continue with Microsoft` reuse the homepage signup controls and share one row above the unchanged full-width product media.
- Mobile hero: one-column flow with stacked actions, a readable headline, a cropped product view, and no horizontal overflow (`scrollWidth` and `innerWidth` are both 390 px).
- Focused browser checks confirmed one `h1`, the exact approved headline, both destinations, and that no `Book a demo` control remains.

**Required fidelity surfaces**

- Fonts and typography: passed. The heading uses the normal site heading stack with no handwritten or colored spans, and its 651 px desktop measure produces exactly two lines.
- Spacing and layout rhythm: passed. Existing landing frame gutters, hero offsets, shell radii, and responsive container spacing are reused.
- Colors and visual tokens: passed. The live `landing-hero-shell` surface and existing button treatments are reused; no words in the feature headline are highlighted.
- Image quality and asset fidelity: passed. The current light/dark Maya captures and shared browser frame are used at their natural aspect ratio; no placeholder or reconstructed product art is present.
- Copy and content: passed. The concise planning-specific headline, brief description, and two familiar provider actions are present with one semantic page heading.

**Comparison history**

1. Initial mobile capture exposed a P2 min-content overflow: the cropped product media widened its grid track and clipped the headline. The media grid item received `min-w-0`; the next capture measured 390 px content width in a 390 px viewport with no clipping.
2. Initial desktop capture exposed a P2 proportion mismatch: inherited nested container padding reduced the product visual to 663 px wide. Desktop padding was removed at the nested screenshot container; the next capture measured the product media at 756 px wide with the original aspect ratio intact.
3. The user rejected the side-by-side composition as a P1 direction mismatch. The grid was removed, the heading and description were moved above the media using the exact homepage hero structure, and `Book a demo` was removed. The final desktop capture shows the full-width 1250 px product visual below the copy; the final mobile capture has no overflow and no demo CTA.
4. The user requested normal hero typography, a title of at least two lines, and two contextual actions. All handwritten accent markup was removed, the heading was replaced with a longer planning-specific sentence constrained to an 868 px measure, and `Plan with Maya` plus `See AI planning in action` were added. The post-fix desktop capture measures the heading at exactly two lines; the mobile capture remains overflow-free with both actions stacked above the unchanged media.
5. The user found the heading and button copy too long, then requested the homepage provider buttons. The title was shortened to `Plan your team’s next move with Maya.` and constrained to 651 px, producing exactly two desktop lines and three compact mobile lines. The custom actions were replaced with the existing Google and Microsoft continuation buttons; the post-fix mobile capture remains 390 px wide with no overflow.

**Findings**

- No actionable P0, P1, or P2 differences remain in the approved hero scope.
- The real Maya screenshot remains landscape, matching the homepage treatment and preserving current product UI without distortion.
- The browser console had no application errors. It retains existing Next Image quality/LCP warnings from the legacy mesh media below the hero; the updated hero does not use that asset.

**Primary interactions tested**

- `Continue with Google` resolves through the existing Google signup flow.
- `Continue with Microsoft` resolves through the existing Microsoft signup flow.
- `Book a demo` is absent as requested.
- Existing feature, use-case, integration, and comparison routes retain their original hero and exactly one `h1`.

**Implementation checklist**

- [x] Approved hero composition implemented.
- [x] Existing landing/pricing visual system reused.
- [x] Plain two-line feature heading verified.
- [x] Existing Google and Microsoft signup actions verified.
- [x] Current light and dark Maya assets wired.
- [x] Desktop and mobile captures verified.
- [x] CTA destinations verified.
- [x] Unrelated marketing detail pages regression-checked.

final result: passed

---

# Homepage customer stories and integrations design QA

**Source visual truth**

- Customer-story interaction and composition reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_GaHNrh/Screenshot 2026-08-24 at 11.37.00 AM.png`
- Customer-story accordion overview reference: `/Users/joseph/Downloads/Screenshot 2026-08-24 at 8.07.24 AM.png`
- Integrations composition reference: `/Users/joseph/Downloads/Screenshot 2026-08-24 at 10.34.24 AM.png`
- Live Maisa desktop idle state: `/private/tmp/fortyone-homepage-qa/maisa-stories-idle-1280x720.png`
- Live Maisa second-card hover state: `/private/tmp/fortyone-homepage-qa/maisa-stories-indigo-hover-1280x720.png`
- Live Maisa mobile state: `/private/tmp/fortyone-homepage-qa/maisa-stories-mobile-390x844.png`
- The references establish the image-led accordion, content hierarchy, and centered icon-cloud direction. FortyOne typography, copy, imagery, and controls intentionally remain native to the product.

**Implementation evidence**

- Customer stories, desktop idle: `/private/tmp/fortyone-homepage-qa/fortyone-stories-idle-exact2-1440x900.png`
- Customer stories, second-card hover: `/private/tmp/fortyone-homepage-qa/fortyone-stories-miningo-hover-1280x720.png`
- Customer stories, reset after pointer leave: `/private/tmp/fortyone-homepage-qa/fortyone-stories-reset-1280x720.png`
- Customer stories, mobile: `/private/tmp/fortyone-homepage-qa/fortyone-stories-mobile-390x844.png`
- Customer stories, desktop dark: `/private/tmp/fortyone-homepage-qa/testimonials-dark-1440x900.png`
- Integrations, desktop light: `/private/tmp/fortyone-homepage-qa/integrations-final-light-1494x795.png`
- Integrations, mobile: `/private/tmp/fortyone-homepage-qa/integrations-mobile-390x844.png`
- Integrations, desktop dark: `/private/tmp/fortyone-homepage-qa/integrations-dark-1440x900.png`
- Same-viewport comparisons: `/private/tmp/fortyone-homepage-qa/maisa-fortyone-stories-comparison-1280.jpg`, `/private/tmp/fortyone-homepage-qa/maisa-fortyone-stories-hover-comparison-1280.jpg`, and `/private/tmp/fortyone-homepage-qa/integrations-reference-comparison.png`
- Route: `http://localhost:3001/#customer-stories`

**Viewport and responsive checks**

- Customer stories were compared against the live source at a matching 1280 x 720 desktop viewport and verified at 390 x 844 mobile.
- Integrations were verified at 1494 x 795 and 1440 x 900 desktop viewports, plus a 390 x 844 mobile viewport.
- Mobile `scrollWidth` equals `innerWidth` at 390 px; neither section introduces horizontal overflow.
- Desktop customer stories use a three-card 2:1:1 expanding row with 10 px gutters and 540 px cards. Mobile presents three fully expanded 310 x 416 px stories with 48 px gaps so the content does not depend on hover.
- The integration tiles wrap from the intended 6 + 3 desktop grouping into a centered two-column mobile layout.

**Focused customer-story checks**

- Live source inspection confirmed the original spring values: stiffness 500, damping 60, mass 1, with the first card selected by default.
- The desktop cards use FortyOne's `rounded-lg` token, 36 px content insets, a fixed 730 x 788 px image plane, a 350 px copy measure, and a 36 x 36 px glass control with a 24 px icon and `rounded-md` corners.
- Every shortened desktop quote renders in three lines at the 350 px source measure. Copy remains left-aligned between the logo and bottom attribution.
- Pointer hover opens the corresponding story, the plus fades away, and leaving the entire row restores the first card. The captured second-card state and reset state confirm both paths.
- The image plane is fixed while the card reveals more of it. Card widths animate through numeric flex growth instead of transform-based layout projection, so expansion does not zoom, shrink, or stretch customer imagery.
- Touch and keyboard activation remain available through the collapsed-card control. `MotionConfig` honors the user's reduced-motion preference, and pointer-triggered selection ignores touch input.
- The approved image-led customer stories now replace the legacy rotating testimonial section.

**Focused integrations checks**

- Nine icon-only tiles are centered with larger cards and consistent icon sizing: six on the first desktop row and three on the second.
- Labels, eyebrow copy, and the integrations CTA are removed from the visible presentation. Accessible names and native titles retain each integration's identity and purpose.
- Gmail is described as replying to Maya from the user's inbox, avoiding an unsupported direct Gmail OAuth/API claim.
- The heading wraps with `team already uses` on the second desktop line, and the body copy uses the standard homepage body scale and normal text color.

**Comparison history**

1. The initial testimonial treatment was progressively tightened to match the supplied image-led reference: reduced radii and height, centered vertical hierarchy, left-aligned constrained copy, hover expansion, delayed fade-and-slide content, stationary imagery, and a smaller glass plus control.
2. Quote length was reduced after the final copy review. The browser measurement confirms every quote now occupies exactly three desktop lines without changing the approved type scale.
3. The initial integration treatment was simplified to icon-only cards, then adjusted to larger tiles and spacing with a centered 6 + 3 desktop wrap. The heading measure was narrowed so `team` begins the second line.
4. The live Maisa HTML and generated Framer bundle were inspected after the user supplied the source URL. The local implementation was updated from an approximation to the exact 2:1:1 spring accordion, 1200 px desktop breakpoint, fixed image canvas, reset-on-leave behavior, and stacked mobile variant.
5. Same-viewport idle and hover comparisons were reviewed after the final motion pass. A moving image wrapper initially exposed a blank strip in collapsed cards; converting the non-animated image plane to a normal layout element removed the artifact while preserving the source reveal behavior.
6. Transform-based layout projection made the fixed image appear to compress during hover. The accordion now animates the actual flex-grow values while keeping the image plane outside Motion layout transforms, preserving its pixel scale throughout the transition.

**Findings**

- No actionable P0, P1, or P2 visual or interaction defects remain in the approved scope.
- The generated customer imagery and the provisional Zimboriginal quote still require final customer approval before publication.
- The implementation adapts the source interaction using FortyOne-owned copy, customer identities, and local imagery; no Maisa testimonial copy, brand marks, or customer imagery are shipped.
- Focused ESLint, TypeScript, Prettier, React Doctor, and browser-console checks are recorded after the final validation pass.

final result: passed
