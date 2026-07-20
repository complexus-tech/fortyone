# Design QA

## Scope

- Public feedback roadmap cards
- Contributor profile navigation and feedback history
- Bottom-right board-creation CTA
- Organization-specific public portal favicon

## Visual references

- Contributor profile reference: `/var/folders/vf/_ym913kj0gx47jx1c5nlky8w0000gn/T/TemporaryItems/NSIRD_screencaptureui_D18xP2/Screenshot 2026-07-20 at 7.38.23 AM.png`
- Roadmap implementation capture: `/tmp/fortyone-roadmap-qa.png`
- Contributor profile implementation capture: `/tmp/fortyone-author-profile-qa.png`

The reference and implementation captures were reviewed together. The reference is a profile-page direction rather than a pixel-identical FortyOne screen, so the implementation keeps FortyOne's existing navigation, spacing, colors, typography, and component scale.

## Checklist

- [x] Card title is the first element and uses the Projects Kanban title scale.
- [x] Author avatar uses the Kanban-style compact, squared avatar with a subtle outline.
- [x] Board is rendered as a color-tinted outlined pill.
- [x] Roadmap vote metadata is icon plus count with no container.
- [x] Author name opens a public contributor profile.
- [x] Contributor profile lists only that author's feedback and supports pagination.
- [x] Empty roadmap columns remain visually distinct without becoming prominent.
- [x] Bottom-right CTA is fixed, text-only, and reads “Create your own board”.
- [x] CTA preserves the feedback-settings destination through signup and onboarding.
- [x] Public portal metadata uses the organization logo for icon, shortcut icon, and Apple touch icon.
- [x] Portals without an organization logo use a 64 × 64 peach-dot PNG fallback.

## Functional verification

- Roadmap author link opened `/portal/art-circles/people/8a798112-90fe-495e-9f1c-f36655e3d8ab` successfully.
- The contributor page returned only Joseph Mukorivo's submitted feedback.
- The CTA resolved to signup with the nested onboarding and feedback-settings callback intact.
- The fallback favicon endpoint returned HTTP 200 with `image/png`, RGBA, and 64 × 64 dimensions.

## Result

Passed. No visible overflow, unintended card nesting, missing metadata, or broken primary navigation was found in the final browser pass.
