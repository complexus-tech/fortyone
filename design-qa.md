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
