# FortyOne email assets — v1

- `wordmark.svg`: exact Wordmark path from `src/components/ui/logo.tsx`, filled espresso (#25150e).
- `wordmark.png`: transparent 504 × 123 PNG rasterized from that SVG; emails display it at 126 × 31.
- `invitation.png` and `invitation-accepted.png`: approved original illustrations, 1040 × 520 PNGs. Display size: 416 × 208 on desktop, fluid on phones. Original prompt provenance is in `design/email-preview/assets/illustration-prompts.md` at repository root.
- `fonts/`: unmodified Inter v4.1 static WOFF2 weights 400, 500, 600, 700 and SIL Open Font License.

Public URL prefix: https://fortyone.app/email-assets/v1/

Deploy these files with the landing site **before** deploying the API and worker email changes. Public assets are accessible without authentication; Next.js applies CORS and immutable caching headers for this version. Keep v1 files stable so already-delivered emails continue to work. Use a new version directory for future artwork changes.

Inter remains optional in inboxes; Arial/Helvetica fallback and Outlook-specific Arial styles are included in the mailer. Essential copy stays selectable text and images have alt text.
