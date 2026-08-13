# Feedback widget design QA

Validated on 2026-08-13 against the supplied UserJot Home, feedback-list, and roadmap screenshots at a 408-430px widget width in both light and dark themes.

## Result

- P0: none.
- P1: none.
- P2: none.
- P3: none.

## Verified states

- Home: defaults as the first widget section; preserves FortyOne language; summarizes the strongest feedback and active roadmap items; and omits updates when the portal has no published updates.
- Home structure: the header action, feedback prompt, section headers, summary rows, simple icon-over-label navigation, and padded attribution footer track the supplied reference without copying its editorial content.
- Feedback list: search, Top/Newest/Oldest ordering, all public status filters, empty/error/loading states, inline voting, and the FortyOne feedback-module navigation icon.
- Roadmap: In progress, Planned, and Completed lanes use the established FortyOne status colors; each lane expands independently and can fetch subsequent pages.
- Feedback detail: the persistent back-button surface, status badge, combined downvote/upvote control, enlarged comment radius, and small-button sizing were verified in the live dark theme.
- Chrome: compact Add feedback action, visible current-color icon in dark mode, larger navigation labels, and a padded 12px Powered by FortyOne footer.
- Theme: explicit light and dark widget themes now control the iframe document tokens; the translucent shell and sticky layers use backdrop blur.
- Radius system: React widget UI uses semantic Tailwind radius utilities only. Full circles remain limited to circular badges and success artwork; the raw loader keeps its fully round launcher CSS because it runs outside Tailwind in the shadow DOM.

## Interaction and runtime checks

- Add feedback remained inside the embed URL and opened the widget composer without the removed helper copy.
- Feedback and roadmap navigation remained in-widget; Planned's Show more control fetched and revealed the remaining lane items.
- The detail view exposed both Downvote feedback and Upvote feedback controls without exercising either mutation against the local portal.
- Browser console check: no warnings or errors across Home, roadmap, composer, and feedback-detail states.

## Evidence

- Home source: `/Users/joseph/Downloads/Screenshot 2026-08-12 at 11.32.27 PM.png` (816×1570).
- Home implementation: `/Users/joseph/.codex/visualizations/2026/08/12/019ff7e9-4bb7-7ee2-8539-768008a9ad47/feedback-widget-home-final.png` (430×820, dark theme, local portal data).
- Combined Home comparison: `/Users/joseph/.codex/visualizations/2026/08/12/019ff7e9-4bb7-7ee2-8539-768008a9ad47/feedback-widget-home-comparison.png`.
- Feedback detail: `/Users/joseph/.codex/visualizations/2026/08/12/019ff7e9-4bb7-7ee2-8539-768008a9ad47/feedback-widget-detail-final.png` (430×820, dark theme).
- Feedback comparison: `/Users/joseph/Downloads/Screenshot 2026-08-12 at 11.33.48 PM.png` and `/Users/joseph/.codex/visualizations/2026/08/12/019ff7e9-4bb7-7ee2-8539-768008a9ad47/feedback-widget-dark-final-full.jpg`.
- Roadmap comparison: `/Users/joseph/Downloads/Screenshot 2026-08-12 at 11.34.04 PM.png` and `/Users/joseph/.codex/visualizations/2026/08/12/019ff7e9-4bb7-7ee2-8539-768008a9ad47/feedback-widget-roadmap-final-full.jpg`.
- Browser interaction checks: ordering/status menus remained inside the widget bounds; lane expansion reduced the visible Show more count from three to two and revealed the fourth item in only the selected lane.

## Iteration history

1. Restored the simple navigation treatment, added Home as the default section, and wired the exact supplied Home icon through the shared icon package.
2. Added the Home summary using live portal feedback and roadmap data while conditionally excluding unpublished updates.
3. Matched feedback-detail control sizing, added the downvote action, increased comment-input radius, and made the back-button surface persistent.
4. Compared the 430×820 dark implementation with the supplied Home reference, then verified roadmap expansion, inline composing, and a clean browser console.

Final result: passed.
