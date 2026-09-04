# FortyOne email design

The approved prototype is in `emails/`. The design is now implemented in the application mailer; Go-rendered review pages are in `integrated/`, with original HTTPS asset references retained in `rendered/`.

## Preview

```sh
python3 -m http.server 4178 --bind 127.0.0.1
```

- `http://127.0.0.1:4178/`: approved eight-message prototype.
- `http://127.0.0.1:4178/integrated/`: application templates with fictional data, including authentication, invitations, notifications, workspace lifecycle, and Maya.

Both galleries support desktop / 390px / 320px, metadata, and an images-off check. Links inside the gallery are intercepted. Standalone HTML links retain fixture destinations; nothing here sends email.

## Rebuild

Run `node build.mjs` to rebuild the original prototype. To refresh application renders, follow `apps/server/docs/email-design.md` at repository root, then run `node build-integrated.mjs` from this directory. The integrated builder only localizes image and font paths; layout and content markup come from the Go renderer.

## Design and assets

Invitation cards are 480px; other cards are 520px. Inter headings are 23px desktop / 21px phone, with 15px body copy on a 21px line height. The warm background fills the browser canvas. Category eyebrows are removed; compact content labels and story identifiers remain.

Canonical deployable assets are in `apps/landing/public/email-assets/v1`: SVG source, rasterized wordmark PNG, invitation and acceptance PNGs, and Inter v4.1 WOFF2 files with their SIL Open Font License. PNGs have descriptive alt text; illustrations never carry required copy.

Inter is progressive enhancement. Custom font support varies by inbox, so Arial/Helvetica fallback and Outlook-specific Arial styling remain. Rounded corners and ticket notches may simplify in older Outlook clients. Browser review is not email-client certification.

## Release boundary

The production template source has been updated locally. Deploy the landing assets before releasing the API and worker. These changes have not been committed, deployed, or sent through an email provider as part of this task. Real destinations, expiration rules, sender identities, and Maya reply-thread headers remain supplied by the existing application.
