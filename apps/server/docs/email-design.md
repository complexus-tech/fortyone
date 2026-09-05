# Email design and assets

The production mailer now implements the approved compact email design: warm #fff6ef canvas, white square-cornered cards, 94px landing-page wordmark above the white card, aligned with its left edge, dashed divider with semicircular ticket cutouts, full-width square-cornered CTA, Georgia with Times/serif fallback, 22px desktop / 20px phone headings, and 14.5px body copy on a 20px line height. Invitation cards are 480px; other templates are 520px. Inner gutters are 32px desktop / 24px phone. Primary text and button backgrounds use `#3d1c10`; muted labels, supporting text and footers use `#6c605b`; button labels stay white and avatars keep their name-based colors and circular shape. Cards and illustrations have square corners. All text, including Outlook buttons, avatar initials and verification codes, uses the shared Georgia stack. Body text, detail values, button labels and signature text share the 14.5px size. Button labels use regular weight, including Outlook. Headings and labels have no added letter spacing. Small labels use their original casing without uppercase transformation; label sizes and weights are unchanged.

## Sources

- `templates/layouts/base.html` owns the shared table-based shell, Outlook conditionals, font declarations, action slot, and footer. The same source is embedded for Maya replies via `templates/embed.go`.
- `pkg/mailer/design.go` selects immutable artwork, widths, and CTA destinations from existing template data. Only invitation and acceptance get illustrations.
- `pkg/mailer/styles.go` supplies inline typography and compact row styles for notification and digest producers.
- `internal/modules/emailreply/service/reply_html.go` styles the agent's escaped paragraph, callout, and list blocks before using the shared shell. It preserves the actual agent copy, including CONFIRM/CANCEL instructions; fixture-only fields and content are not synthesized in live messages.
- Public assets live in `apps/landing/public/email-assets/v1` at repository root, including the source SVG, rasterized wordmark, two illustrations, and calendar/comment icons. Legacy licensed Inter assets remain available for already-delivered emails; new emails do not download fonts.

## Release order

1. Deploy the landing site assets and the `/email-assets/v1/:path*` CORS/cache headers.
2. Verify the public logo, both illustrations, and icons return HTTP 200.
3. Apply forward migrations `000184_routine_email_deliveries` and `000185_email_avatar_handles` to the release database after verifying it in the PostgreSQL integration suite.
4. Deploy the API, then the worker with the updated template files. Set the worker’s `APP_API_PUBLIC_URL` to the same public HTTPS API origin used by the API. The worker also embeds the Maya reply shell; it must be rebuilt.
5. Check representative messages in Gmail, Apple Mail, and Outlook, including images blocked and reply threading. No live email-client certification is implied by browser tests.

Already-delivered emails retain their original HTML. Maya deliveries whose provider payload was already frozen for retry also retain that payload; this preserves idempotency and the original message content.

## Verification

Run the application template and delivery tests from `apps/server`:

```sh
go test ./pkg/mailer ./internal/modules/emailreply/service ./internal/taskhandlers ./internal/bootstrap/worker
```

The tests render templates in memory and check the shared layout, assets, escaping, actions, footer, avatars and consolidated delivery behavior. Design galleries, generated HTML collections and export scripts have been removed after approval. Production rendering uses only application templates and the versioned landing assets listed above.

## Consolidated delivery (September 2026)

- The existing one-hour, recipient/workspace activity batch remains the delivery unit. Legacy single-notification tasks now consume that same batch. All included notification IDs, including details represented by the overflow link, are marked covered together after successful delivery. Read items and disabled categories are excluded by the existing scoped queries.
- Maya priorities now join the existing one-hour unread activity batch. There is no standalone morning briefing or weekly note schedule. Retired queued morning/deadline/weekly tasks are consumed without sending. If there is no unread activity, no routine email is sent solely for reminders.
- The first eligible weekday activity batch can include story deadlines, Monday objective/key-result reminders and Monday weekly overview, preserving category preferences and access checks. Saved IANA timezones determine the local date; missing/invalid zones use UTC. Priorities are included at most once per recipient/workspace/local date, with coverage recorded atomically alongside notification IDs after successful delivery. Later batches contain only new unread activity.
- Activity batches share recipient send claims, preventing concurrent workers from consuming the same snapshot. Claims expire after ten minutes; reclaimed attempts get a new ID so an old attempt cannot complete a newer one. SMTP and the database cannot commit atomically: a crash after SMTP accepts the message but before completion can still cause a retry duplicate. This is not an exactly-once provider guarantee.
- Each notification digest or briefing section contains at most five details, with a count and destination for overflow. Activity overflow goes to Notifications; work/objective/strategy overflow goes to its existing product page because email-only reminders are not inbox records. Feedback digests retain their five-submission limit.
- Feedback reviewer digests retain their separate daily/weekly/off board preferences. Strategy snapshot generation and its existing cadence are retained; strategy notifications can share a batch/briefing with other eligible activity. Immediate auth, invitation, lifecycle and direct Maya reply deliveries are preserved. This release does not introduce the proposed two-email daily budget, change activity quiet hours, or persist entity-version coverage across strategy and work reminders.

## Avatars and footer

Activity actors are rendered inline at 20px with 8px initials and a 4px gap. Colors use the same NFKC/whitespace/lowercase normalization, UTF-16 FNV-1a hash and six-color palette as `packages/lib/src/avatar-color.ts`. Unicode parity fixtures were derived from the frontend implementation and remain in the Go tests. Copy and URLs are escaped; model output cannot inject image markup. Calendar/comment PNGs are packaged in the existing versioned asset folder.

Emails now store a stable application URL, `/media/email-avatars/{opaque-handle}/avatar`, rather than an AWS/Azure signature. A handle is created only for a visible actor with a photo and is stored once in `email_avatar_handles`; it has no time limit and does not depend on an authentication signing key. The unauthenticated image endpoint accepts only that opaque handle, loads the current active user's profile reference, and redirects with `Cache-Control: no-store` to a newly generated five-minute storage URL. Storage stays private. No arbitrary object-key, bucket or URL input is accepted by this endpoint.

Changing the profile photo keeps the same email URL and resolves the replacement file. Removing the photo or deleting/deactivating the account stops photo resolution; the email retains initials as image alt text. Preserve the handle table and public API origin when deploying or rotating keys. The worker needs `APP_API_PUBLIC_URL` set to the public HTTPS API origin; invalid/non-HTTPS configuration falls back to initials and logs a diagnostic. The stable URL applies to newly generated mail; already-delivered HTML cannot be rewritten.

This follows Art Circles' public media redirect pattern, with opaque profile handles in place of raw storage keys. Unit tests reopen the same URL after eight simulated days, check that it requests a fresh signature, and verify it follows a replaced photo. Live cloud-storage resolution remains a release check.

The shared footer places Manage notifications on the left and “A product of Complexus” on the right, at 13px to match footer notes. When there is no settings link, the company credit is the only column and aligns left. Footer notes also align left. The footer has no horizontal padding, so its text aligns with the outer edges of the card on desktop and mobile. Existing explicit destinations take precedence, including feedback board settings and explicit omission for external contributors. Known workspace URLs use `/settings/account/notifications`. Maya reply footers use the already-authorized thread's workspace slug; sign-in-only messages without workspace context omit the workspace settings link.

Login and feedback verification codes render directly on the white card, without a nested panel or repeated code label. Browser review was completed during design approval; live email-client and cloud-storage checks remain release steps.
