# Global feedback profile design QA

- Source visual truth: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_QcEVWK/Screenshot 2026-08-11 at 11.57.15 PM.png`
- Implementation screenshot: `/tmp/fortyone-global-profile-normalized.png`
- Combined comparison: `/tmp/fortyone-profile-comparison.png` (source above, implementation below)
- Viewport: 1267 x 620 CSS pixels at device scale factor 1
- Dimensions: source 1900 x 930 pixels normalized to 1267 x 620; implementation 1267 x 620 pixels
- State: dark theme, desktop, Feedback tab, one feedback contribution

## Full-view comparison evidence

The implementation carries over the existing contributor profile composition: identity header, Feedback and Comments tabs, inline contribution totals, border-separated activity rows, status treatment, vote count, and the right-hand contributor card. The global account header remains above the profile so the user retains account navigation; the reference captures only the contributor content area. Cross-portal copy and workspace labels are intentional changes required by the global scope.

The normalized comparison shows no actionable P0, P1, or P2 layout, hierarchy, color, typography, or component-style mismatch. A separate focused crop was not needed because the normalized full view keeps the tabs, contribution row, and sidebar text readable at the comparison size.

## Required fidelity surfaces

- Fonts and typography: uses the same system font stack, weights, sizes, line heights, and muted hierarchy as the existing contributor screen.
- Spacing and layout rhythm: reuses the same 78rem grid, 19rem sidebar, 2.5rem column gap, tab spacing, row padding, radii, borders, and sticky behavior.
- Colors and visual tokens: uses the existing background, surface, border, muted text, hover, and semantic feedback-status tokens in dark and light themes.
- Image quality and asset fidelity: avatars use the existing Avatar component and the live user's stored profile image; the QA fixture intentionally falls back to initials. Icons come from the existing icon package.
- Copy and content: contributor language is adapted from one named workspace to all FortyOne portals, and each contribution identifies its originating workspace.

## Interaction and console checks

- Feedback and Comments tabs switch correctly and preserve the selected tab in the URL.
- Feedback rows navigate to the originating portal request.
- Account menu and theme switching work in the rendered page.
- No browser console warnings or errors were present during the checked states.

## Comparison history

1. Initial capture used the local default light theme, which did not match the dark reference. Switched through the existing Appearance menu and recaptured in dark mode.
2. The source was a 1900 x 930 high-density capture. Normalized it to 1267 x 620 and recaptured the implementation at the same pixel and CSS dimensions before the final comparison.

## Findings

No actionable P0, P1, or P2 findings remain.

## Follow-up polish

No P3 follow-up is required for this pass.

final result: passed
