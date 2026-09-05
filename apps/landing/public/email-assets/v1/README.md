# FortyOne email assets — v1

- `wordmark.svg`: exact Wordmark path from `src/components/ui/logo.tsx`, filled espresso (#25150e).
- `wordmark.png`: transparent 504 × 123 PNG rasterized from that SVG; emails display it at 94px wide with proportional height.
- `invitation.png` and `invitation-accepted.png`: approved original illustrations, 1040 × 520 PNGs. Display size: 416 × 208 on desktop, fluid on phones.
- `fonts/`: legacy Inter v4.1 files and SIL Open Font License, retained for already-delivered emails. New emails do not reference them.

Public URL prefix: https://fortyone.app/email-assets/v1/

Deploy these files with the landing site **before** deploying the API and worker email changes. Public assets are accessible without authentication; Next.js applies CORS and immutable caching headers for this version. Keep v1 files stable so already-delivered emails continue to work. Use a new version directory for future artwork changes.

New emails use Georgia with Times/serif fallback throughout, including Outlook. No web-font download is required. Essential copy stays selectable text and images have alt text.
