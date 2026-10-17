# Notification Settings Design QA

- Source visual truth: `/Users/joseph/Downloads/Screenshot 2026-09-19 at 17.12.23.png`
- Reviewed implementation: `/Users/joseph/Downloads/Screenshot 2026-09-19 at 17.22.13.png`
- Route: `fortyone://settings/notifications`
- Viewport: physical iPhone 12, portrait, dark appearance
- Screenshots: 1170 x 2532 pixels before display resizing

## Full-view comparison evidence

The physical-device screenshot confirmed the page hierarchy, grouped preference rows, native switches, dark appearance, and working scroll layout. The user then verified that the on-device test notification was received.

The reviewed screenshot represented the first implementation pass. Its prominent action pills, centered navigation title, always-visible preference list, and raw APNs entitlement exception were subsequently replaced based on physical-device feedback.

## Focused region comparison evidence

- Navigation: moved to the native stack header with a larger, left-aligned title beside the back control.
- Global state: replaced the two prominent pills with a single standard iOS switch row.
- Test action: replaced the secondary pill with a simple iOS link action.
- Preferences: retained grouped rows and native switches, but now reveals the section only when global notifications are enabled.
- Error state: removed the raw `aps-environment` exception and all developer-account copy from the interface. Local notification testing remains available.
- Safe area: the native navigation bar owns the top inset and `SafeContainer` owns the bottom inset.

## Findings

- [P2] A final screenshot of the revised pass is not available to compare exact text wrapping and spacing.
  - Location: settings and notification settings routes.
  - Impact: the revised structure is verified statically and by the live Metro bundle, but final pixels have not been recaptured.
  - Fix: capture the current notification page in both off and on states.

## Comparison history

- Pass 1: physical-device screenshot found raw entitlement copy, oversized pill actions, and navigation alignment that did not feel native.
- Pass 2: adopted the native stack header, grouped settings cards, a native master switch, progressive disclosure, a text-only test action, and concise single-line preference rows.

## Follow-up polish

- Confirm custom workspace terminology wraps cleanly beside switches at larger Dynamic Type sizes.
- Recheck the revised page in light appearance.

final result: blocked pending a screenshot of the revised pass

---

# Document library, editor, and sharing QA — 2026-10-17

## Scope and visual references

The approved generated concepts were implemented using FortyOne's existing UI components, icons, color tokens, and typography. The earlier notification QA above is unrelated and remains unchanged.

- Library reference: `/Users/joseph/.codex/generated_images/01a14967-57cf-7b42-8e3c-095d21480e74/exec-0ab2e1b9-39ce-4681-955b-3b3a9882f489.png`
- Editor reference: `/Users/joseph/.codex/generated_images/01a14967-57cf-7b42-8e3c-095d21480e74/exec-91a683ee-a3fe-47f1-bf51-c64ce97d03de.png`
- Sharing reference: `/Users/joseph/.codex/generated_images/01a14967-57cf-7b42-8e3c-095d21480e74/exec-e6cd48d2-47a7-4430-9656-ac23e554e1b4.png`
- Preview: `http://127.0.0.1:4174/preview/docs`
- Screenshots: `/Users/joseph/development/complexus/fortyone/output/playwright/document-ui/`
- Viewports: 1584 × 992 desktop and 390 × 844 phone, dark appearance.

The preview bundles actual repository components and styles with temporary sample-data hooks. Its small navigation strip and simplified application shell are fixture controls, not proposed product UI. Browser actions were performed against this fixture, not live documents.

## Comparison evidence

Each approved mockup and corresponding implementation screenshot was reviewed together. The library preserves document ownership, related work, access, updated time, orange favorites, and search/filter/sort hierarchy. Narrow rows wrap metadata without making text smaller. The editor preserves aligned checkboxes, visible save/access status, a selection toolbar below the header, and docked comments. Sharing separates workspace access, individual roles, and the public link with an explicit explanation that saved changes appear publicly.

Existing avatars, outline navigation, toolbar icons, surface treatments, and base font tokens intentionally differ from the generated images. No `text-xs`, `text-sm`, or equivalent small typography variants were added to the new library, sharing, header, insertion, or toolbar UI. Existing base text currently resolves to 13px in this fixture; the implementation reuses that application token rather than introducing a new scale.

Final evidence:

- `library-desktop.png` and `library-mobile.png`: responsive library, final header, rows and controls.
- `archived-desktop.png`: archived document with a restore action.
- `editor-desktop.png`: selected passage, formatting toolbar, aligned checkboxes, docked comments.
- `editor-mobile.png` and `comments-mobile.png`: readable document and dialog-based comments at phone width.
- `sharing-desktop.png` and `sharing-mobile.png`: access roles and public actions.
- `editor-insert.png`: insertion button opens the existing slash-command menu.
- `reviewer-selection.png` and `reviewer-comment.png`: native read-only text selection exposes Comment and opens a composer containing the selected quote.

## Interaction checks and corrections

Passed in the component fixture: favorite state stays consistent between the sidebar and list; search narrows results; archive moves a document to Archived; restore returns it to active results; sharing exposes accurate role labels; blank-paragraph insertion opens the existing block menu and inserts a checklist; native reviewer selection opens a comment-only toolbar and exact-quote composer. The phone library's page width equals the viewport width (390px).

Corrections made during QA:

1. The phone library title was squeezed beside actions. Title and description now occupy their own row, above the action buttons.
2. Read-only ProseMirror selection did not update the comment toolbar. A selection-only bridge now synchronizes native text selection without focusing or changing document content.
3. Permission changes could leave stale formatting controls visible. A controlled formatting permission now updates the toolbar immediately.

No remaining P0–P2 visual issue was found in the captured states.

## Automated validation and limits

- 13 targeted Jest suites passed: 78 tests covering library controls, sharing/header behavior, archive/restore mutations, favorites, public rendering, comments, slash commands, and reviewer editing permissions. The final native-selection bridge also passed its six-test reviewer suite.
- Projects TypeScript check, shared UI component build, scoped projects lint, and `git diff --check` passed.
- Document backend HTTP, repository, and service unit tests passed. Integration tests compile; their database execution is unverified because `TEST_DATABASE_URL` is unset.
- Standard shared UI lint is blocked by the existing ESLint configuration/module-resolution problem; a legacy configuration attempt also reports existing file-level violations. This is not a clean full-repository lint result.
- React Doctor reported 89/100 with warnings about component complexity and existing code, with no errors.
- The main local application's authentication service was unavailable. Authenticated end-to-end behavior, database integration, and production rendering therefore remain unverified. No deployment or commit was performed.

Document UI result: passed for the tested component layouts and interactions; authenticated integration remains blocked by the local authentication service and missing test database configuration.

## Follow-up refinements — 2026-10-17

The user's subsequent direction supersedes the docked comments and public-link explanatory copy above. Document rows now use `py-3` with compact circular owner avatars; sidebar search and create buttons have identical measured dimensions. The search bar and font scale are unchanged. Sharing uses smaller buttons and avatars, a right-aligned access selector on the label row, and the label “Members.” Both public-link explanatory messages were removed. Comments use the previous `DocumentCommentsOverlay` presentation instead of resizing the document into a docked panel.

Verified actual-component fixtures at 1584 × 992 and 390 × 844. Evidence: `library-compact.png`, `sharing-compact.png`, `sharing-compact-mobile.png`, and `comments-restored.png` in the screenshot directory above. Scoped lint, projects type-check, nine existing sharing/header tests, and whitespace checks passed. React Doctor remains at 89/100 with the same existing warnings. These are local changes; no commit or deployment was performed.

### Avatar variants and hover favorites refinement

The latest user direction replaces the custom compact avatar dimensions and font overrides with the shared `Avatar size="sm"` variant in the library, sharing panel, and document header. Rows now use `py-2`. The favorite button is absolutely positioned in the left gutter, hidden at rest and revealed on row hover or keyboard focus; its tooltip is exactly “Favorite.” It no longer takes space in the document column. Active document menus include Open document and Favorite/Remove favorite, with Archive document still restricted to the owner. The menu preserves favorite access on touch screens.

Browser verification passed for the resting and hovered desktop row, keyboard tooltip, Open document navigation, and mobile menu favorite toggling. Evidence: `library-hover-default.png`, `document-row-menu.png`, and `document-row-menu-mobile.png`. Scoped lint, projects type-check, the nine sharing/header tests, and whitespace checks passed; React Doctor retained the same score and warnings. Search-bar layout is unchanged.

### Hover options and compact insert button

The row options button now hides at rest on hover-capable devices, revealing on row hover, keyboard focus, or while its menu is open. Touch devices retain a visible menu entry. The sidebar create button retains the shared small variant with the original plus icon geometry and stroke. The document insertion control now uses the same small button and icon dimensions with the naked variant.

The component fixture confirmed options opacity changes from 0 at rest to 1 on hover and while open. Sidebar create and insert buttons both measured 28.34px square at the fixture's existing font scale, with matching 15.52px icons and stroke width 2. The insert button has no border or background and still opens the existing block menu. Evidence: `library-hover-actions.png` and `editor-insert-compact.png`. Scoped lint, formatting, fixture build, and whitespace checks passed. React Doctor remains at 89/100 with the same seven existing warnings.
