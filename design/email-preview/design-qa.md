# Email preview QA

final result: passed

Scope: browser-rendered review collection, not production email-client certification.

## Application template migration

- Exported the SVG landing wordmark to a 504 × 123 transparent PNG. Canonical assets are now in `apps/landing/public/email-assets/v1`, with both approved illustrations and four licensed Inter weights.
- Applied the approved shell to all 12 file-backed templates and embedded that same layout for Maya replies. Invites match the approved art/card composition; actual notifications and Maya retain the content vocabulary and data supplied by their producers rather than inventing fixture-only fields.
- The action slot preserves each template's actual destination. Authentication codes and expirations, workspace recovery instructions, notification preferences, and Maya CONFIRM/CANCEL copy and thread handling remain covered by focused tests.
- All seven focused Go package suites pass: mailer, jobs, emailreply, emailagent, taskhandlers, worker bootstrap, and eventconsumer. Asset existence and exact SVG-path agreement with the landing wordmark are checked.
- Next.js configuration was loaded and its asset CORS/cache rule verified. Public hosting is not deployed or live-verified.
- All 14 application review variants passed at 838px, 390px, and 320px (42 measurements): no horizontal overflow; fonts and images loaded. One acceptance-image measurement initially occurred before loading completed; a settled recheck passed. Evidence: `qa/integrated-layout-checks.json`.
- Visual evidence: `qa/integrated-invitation-desktop.png` and `qa/integrated-maya-confirmation-mobile.png`. `integrated/` changes only asset URLs for local loading; `rendered/` contains actual production URL references.
- Deployment order and local reproduction are in `apps/server/docs/email-design.md`. No mail was sent and no deployment or commit was performed.

## Latest spacing refinement

- Reduced main titles by 1px: 23px desktop / 21px phone, with 29px / 27px line heights. Item headings are 15px; larger section headings are 17px.
- Body copy remains 15px, with line height reduced from 24px to 21px. Paragraph spacing is 12px; closing copy and signatures use 16px.
- Notification row padding reduced from 18px to 12px; conversation and Maya sections use 16px instead of 24–25px. Section starts and title-to-copy spacing are tighter.
- Desktop card heights before → after: Maya weekly 764.5 → 638.5px; Maya confirmation 720.5 → 602.5px; conversation 512 → 432px; story updates 639 → 553px. All copy remains present.
- All eight emails rechecked at 838px, 390px, and 320px with no horizontal overflow or failed image loads. Measurements: `qa/tight-spacing-checks.json`. Visual captures: `qa/maya-weekly-tight-desktop.png` and `qa/maya-confirmation-tight-mobile.png`.

## Previous compact revision

The user requested smaller cards and type, Inter, removal of category eyebrows, and a full-width warm preview background. These instructions supersede the original mockup dimensions and the initial typography findings below.

- Invitations and acceptance: 480px maximum card width; all other cards: 520px (previously 600px).
- Desktop inner gutters: 32px; phones: 24px. Headings: 24px / 22px; body: 15px / 24px line height. Image measure: 416 × 208px on desktop.
- All category eyebrows were removed from fixtures and email markup. Story identifiers and useful content labels remain.
- Inter v4.1 is bundled in four static weights, with Arial/Helvetica fallbacks and an explicit Outlook Windows Arial override. Browser font loading was confirmed, including `document.fonts.check('600 22px Inter')`. Inbox support is still conditional.
- The desktop gallery has no surrounding email-window border or width cap. Its warm canvas fills the available area and viewport height; phone controls retain their device-width frame.
- Rechecked all eight standalone emails at 838px, 390px, and 320px: no horizontal overflow, all images loaded, and expected card/title sizes. Evidence: `qa/compact-layout-checks.json`.
- Current visual captures: `qa/invitation-compact-desktop.png`, `qa/maya-follow-up-compact-mobile.png`, and `qa/gallery-compact-desktop.png`. Older captures document the previous design.

## Original visual targets (historical)

Sources are the user-selected final email family in `/Users/joseph/.codex/generated_images/01a06e05-f542-7ff3-bca8-12cf723596a3/`:

| Email              | Source filename                               | Comparison evidence                  |
| ------------------ | --------------------------------------------- | ------------------------------------ |
| Invitation         | exec-d008a256-8f3a-48c5-b76a-1ad87f7a99fe.png | qa/invitation-comparison.png         |
| Story updates      | exec-dcee53c9-0ba9-4b39-aa66-368365e39d81.png | qa/story-updates-comparison.png      |
| Scheduled deletion | exec-2eea2136-a4c4-4607-841f-c13fde3fafa9.png | qa/workspace-deletion-comparison.png |
| Conversation reply | exec-48bbca82-f01f-4c29-a83f-dfd017ba1b88.png | qa/conversation-comparison.png       |
| Maya weekly        | exec-a9aecdc7-2bef-4386-8a1d-563fd64dcf0d.png | qa/maya-weekly-comparison.png        |
| Maya follow-up     | exec-ac602021-51b5-4109-90c1-bd0fad08a773.png | qa/maya-follow-up-comparison.png     |

Invitation accepted and workspace inactivity extend the same selected family. Their source content was checked against the existing invitation acceptance and inactivity templates.

## Original capture and normalization (historical)

- Local gallery: `http://127.0.0.1:4178/`.
- Standalone implementations: `emails/{id}.html`.
- Desktop CSS viewport: 838 × 1100; native screenshots: 838 × 1100 pixels. Device reports DPR 2, but the native screenshot API supplies CSS-sized pixels. No additional density scaling is applied.
- Original mockup widths vary (1122 pixels for invitation, story updates, deletion, and Maya weekly; 1047 for conversation; 1196 for Maya follow-up). Comparison sources are resized proportionally to 838px width. The raw source canvas and the consistent 600px email shell are not pixel-identical. Card scale, typography hierarchy, spacing, colors, and content are compared with this difference acknowledged.
- Full comparison images place source on the left and rendered HTML on the right. The source aspect ratio is preserved.
- Authoritative implementation captures: `qa/{id}-desktop-native.png`; mobile examples: `qa/{id}-mobile-native.png`.
- Mobile layout measurements cover every email at 320px and 390px. See `qa/browser-layout-checks.json` for 24 desktop/mobile measurements. No horizontal overflow or failed image loads were found.
- Mobile screenshots use a taller viewport to show complete content; the layout breakpoint is driven by the 320px width. The final deletion fix is captured in `qa/workspace-deletion-mobile-native.png` at 320 × 1400.
- Early `*-raw.png` captures from the browser full-page screenshot helper had an inconsistent content scale. They were not used as authoritative visual evidence. Native screenshot captures replaced them.

## Findings and iteration history

1. Initial typography was too small relative to the selected family, especially the wordmark and title. Increased the wordmark from 126px to 144px, headline from 26px to 30px, body from 16px to 17px, and adopted the locally available Helvetica Neue with Helvetica/Arial fallbacks. Kept 24px mobile headings. Fresh native captures were compared after the change.
2. [P2, resolved] At 320px, the deletion requester's address wrapped awkwardly in a two-column detail table. Added narrow-screen stacked detail cells below 420px. Post-fix evidence shows `joseph@example.com` on one line, with no page overflow. The DOM measured a 240px value cell and a 320px document width.
3. Narrowed desktop heading measure to keep longer titles readable and emphasized the exact **Cancel scheduled deletion** action in the warning copy. Maya health columns received additional separation.

No remaining actionable P0/P1/P2 findings in the preview scope.

## Original fidelity surfaces (historical)

- **Typography:** readable 17px body, 30px desktop / 24px mobile titles, and compact metadata. The system font stack is an intentional email-compatible approximation of the generated mock typography. Different installed fonts can produce different line breaks. Controls retain readable contrast and explicit line heights.
- **Spacing:** consistent 600px shell, 40px desktop / 24px mobile inner gutters, rounded white card, external wordmark/footer, and ticket divider. Natural email lengths differ by content; the implementation does not preserve arbitrary mockup canvas whitespace.
- **Colors:** solid apricot `#fff1e7`, white surface, espresso `#25150e`, warm neutral secondary copy, and restrained warm divider. Warning and health text use darker semantic colors for readability. No reliance on gradients or background imagery.
- **Assets:** exact landing-page Wordmark path rasterized as PNG. Both illustrations are independently generated assets in the approved visual style, resized proportionally to 1040 × 520 without cropping. They render at up to 520 × 260. Only invitation/acceptance contain illustration art. Essential copy remains text.
- **Copy:** all six selected concepts are represented; added acceptance/inactivity copy uses current product semantics. 48-hour requested deletion is distinct from 30-day inactivity. Maya confirmation says no change has been made and retains literal CONFIRM/CANCEL replies. Notification excerpts and counts are fictional fixtures, not claims about live events.

Full-view paired captures and direct inspection of mobile text, the logo, artwork, CTA, and confirmation instructions were sufficient for this typography-led email family; no additional image zoom crops were necessary.

## Browser behavior checked

- Sidebar navigation updates the email, subject, preheader, metadata, plain-text link, and HTML download link.
- Desktop / 390px / 320px preview modes resize the actual iframe.
- Metadata disclosure displays sample From, To, Reply-To, source path, and fields.
- Images-off check removes art while preserving essential content and action.
- The gallery intercepts email links and displays a sample destination dialog; closing it returns to the preview.
- Gallery page itself fits a 390px browser viewport with no document overflow.
- Standalone HTML pages produced no console warnings/errors during the direct-page pass. The gallery inspection session logged a `MutationObserver.observe` diagnostic while browser tooling inspected iframe content. The authored gallery uses `ResizeObserver`, not `MutationObserver`; the message is consistent with browser inspection instrumentation, though its originating script was not identified. It did not prevent navigation, rendering, controls, or downloads from being exercised. No claim of an entirely clean gallery tooling log is made.

## Structural checks and release boundaries

- Eight HTML documents have plain-text alternatives and load all local image paths.
- HTML files stay below 16KB. Table layouts have presentation roles; images have explicit width and alt text. No scripts, forms, iframes, SVG, external stylesheets are embedded in the emails.
- Outlook VML buttons, conditional wrappers matching the 480px / 520px card widths, and DPI settings are present. Their real Outlook rendering has not been verified.
- All destinations use example.com and all email identities are fictional. No messages were sent, no production templates were edited, and no code was committed or deployed.
- Public HTTPS image hosting, real field binding, provider headers, Gmail/Outlook/Apple Mail rendering, forced dark mode, and authenticated Maya reply delivery remain production-integration checks after design approval.

## Follow-up polish

Decorative notches and rounded corners may simplify in older Outlook engines. Client rendering tests should determine whether to retain that graceful fallback or supply a more exact alternative. This does not block browser review of the requested HTML.
