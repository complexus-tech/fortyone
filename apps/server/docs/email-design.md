# Email design and assets

The production mailer now implements the approved compact email design: warm #fff6ef canvas, white rounded cards, landing-page wordmark, ticket divider, full-width CTA, Inter with Arial/Helvetica fallback, 23px desktop / 21px phone headings, and 15px body copy on a 21px line height. Invitation cards are 480px; other templates are 520px. Inner gutters are 32px desktop / 24px phone.

## Sources

- `templates/layouts/base.html` owns the shared table-based shell, Outlook conditionals, font declarations, action slot, and footer. The same source is embedded for Maya replies via `templates/embed.go`.
- `pkg/mailer/design.go` selects immutable artwork, widths, and CTA destinations from existing template data. Only invitation and acceptance get illustrations.
- `pkg/mailer/styles.go` supplies inline typography and compact row styles for notification and digest producers.
- `internal/modules/emailreply/service/reply_html.go` styles the agent's escaped paragraph, callout, and list blocks before using the shared shell. It preserves the actual agent copy, including CONFIRM/CANCEL instructions; fixture-only fields and content are not synthesized in live messages.
- Public assets live in `apps/landing/public/email-assets/v1` at repository root, including the source SVG, rasterized wordmark, two illustrations, and licensed Inter files.

## Release order

1. Deploy the landing site assets and the `/email-assets/v1/:path*` CORS/cache headers.
2. Verify the public logo, both illustrations, and font files return HTTP 200; font responses must allow cross-origin access.
3. Deploy both API and worker with the updated template files. The worker also embeds the Maya reply shell; it must be rebuilt.
4. Check representative messages in Gmail, Apple Mail, and Outlook, including images blocked and reply threading. No live email-client certification is implied by browser tests.

Already-delivered emails retain their original HTML. Maya deliveries whose provider payload was already frozen for retry also retain that payload; this preserves idempotency and the original message content.

## Local review

From `apps/server`, export fictional renders without sending email:

```sh
FORTYONE_EMAIL_PREVIEW_DIR="$PWD/../../design/email-preview/rendered" go test ./pkg/mailer -count=1
```

From the repository root:

```sh
node design/email-preview/build-integrated.mjs
```

The local email-preview server exposes `/integrated/`. This review copy changes only asset URLs to local files; `/rendered/` retains the production HTTPS asset URLs. Fixtures cover every file-backed template plus Maya reply and weekly notification layouts. The gallery intercepts CTA clicks. Sender configuration, actual destinations, expirations, notification selection/batching, and Maya thread headers still come from their existing producers.
