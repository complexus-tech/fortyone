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
- Existing feature, use-case, integration, comparison, and AI Project Manager routes retain their original hero and exactly one `h1`.

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
