# Feedback Contributors, Updates, and Signed Widget Identity Plan

## Document Status

- Status: Draft implementation plan
- Scope: Next phase after anonymous feedback and the Feedback/Roadmap widget
- Primary surfaces:
  - `apps/server/internal/modules/feedback`
  - `apps/server/internal/modules/notifications`
  - `apps/projects/src/modules/public-portal`
  - `apps/projects/src/modules/feedback-widget`
  - `apps/projects/src/modules/settings/workspace/feedback`
- Goal: Let people participate without a FortyOne account while preserving a
  reliable update loop, then complete the public Updates surface and securely
  identify users inside third-party widget embeds.

## Current Baseline

The first release already provides:

- Portal-level `account_required` and `anonymous_allowed` participation modes
- Portal-scoped account and unlinkable anonymous contributors
- Truly anonymous, submit-only feedback with no name, email, or reusable
  identity
- A public tracking URL for anonymous submissions
- Account-gated votes and comments
- Signed anonymous ingress, rate limits, duplicate detection, and a honeypot
- A hosted widget with Feedback and Roadmap, bubble/custom/inline modes, themes,
  and a copyable installation snippet

The next phase must preserve these contracts. It must not silently turn a guest
into a global `users` row, promise notifications to contactless contributors, or
make a third-party iframe depend on the normal `SameSite=Lax` account cookie.

On the frontend, introduce a discriminated public participant type for
`account`, `verified_guest`, `external`, and `anonymous`. Do not represent a
verified guest as a fake `PublicPortalViewer`; account navigation and account
menus remain account-only.

## Product Decisions

### Participation modes

Expand the portal setting to three explicit choices:

1. `account_required`
   - A normal FortyOne account is required for submissions, comments, and votes.
2. `verified_guest`
   - A normal account or portal-scoped verified email identity may submit,
     comment, vote, follow, and receive updates.
3. `anonymous_allowed`
   - Account and verified guest participation remain available.
   - A person may alternatively make a truly anonymous submission.
   - Truly anonymous comments and votes remain disabled.

Do not use one `anonymous` boolean to represent all three cases.

### Public attribution

Keep contactability separate from public identity:

- A verified guest may show their chosen display name publicly.
- A verified guest may choose **Hide my name publicly** when the portal allows
  it. Administrators still see the verified identity for follow-up.
- A truly anonymous submission has no recoverable contributor identity and is
  shown as `Anonymous` to both the public and administrators.
- Public serializers must never use an email address as a display-name
  fallback.

Add a portal policy with these values:

- `show_identity`
- `allow_public_masking`
- `always_mask_guests`

Default existing portals to `show_identity` so the migration does not alter
current presentation.

### Update delivery

- Account contributors keep existing in-app notifications and email behavior.
- Verified guests receive contributor-scoped email delivery without becoming
  normal users.
- Item authors automatically follow their item.
- Commenters and voters are offered an explicit **Notify me about updates**
  choice; do not subscribe them silently.
- Notify followers for:
  - Direct replies and meaningful public comments
  - Merge into another feedback item
  - Planned
  - In progress
  - Completed or shipped
  - Closed, when the administrator supplies a public explanation
  - A published Update linked to the item
- Do not notify for internal-only triage or read/unread changes.
- Truly anonymous contributors receive no personal delivery. Their public item
  URL and the roadmap remain the complete update path.

## Phase 1: Verified Guest Identity and Sessions

### Schema

Use the next available forward migration. Applied migrations remain immutable.

Extend `feedback_contributors` with:

- `email`
- `email_verified_at`
- `display_name`
- `avatar_url`
- `external_id`
- `last_seen_at`
- `blocked_at`
- `blocked_reason`

Expand `kind` to:

- `account`
- `verified_guest`
- `anonymous`
- `external`

Add partial portal-scoped uniqueness for non-null `user_id`, normalized email,
and `external_id`. Constraints must ensure:

- Verified guests have a normalized email and `email_verified_at`.
- Anonymous contributors have no user, email, or external identity.
- External contributors have an `external_id`.
- Deleting a normal user does not delete their historical feedback.

Add `feedback_contributor_verifications`:

- `id`
- `portal_id`
- `email`
- `display_name`
- `token_hash`
- `code_hash`
- `source` (`portal` or `widget`)
- `expires_at`
- `consumed_at`
- `attempt_count`
- `created_at`

Verification links use random high-entropy opaque tokens and store only their
SHA-256 hashes. Short human-entered codes have less entropy, so authenticate
them with a domain-separated HMAC key derived from `APP_AUTH_SECRET_KEY` and
enforce bounded attempts. Expire unconsumed verification records quickly and
purge them on a schedule.

Add `feedback_contributor_sessions`:

- `id`
- `portal_id`
- `contributor_id`
- `token_hash`
- `source` (`portal` or `widget`)
- `expires_at`
- `revoked_at`
- `last_used_at`
- `created_at`

Store only a SHA-256 hash of each random session token. Bind every lookup to the
portal and reject cross-portal reuse.

### HTTP contract

Add:

- `POST /portals/{portalSlug}/feedback/verifications`
  - Accepts email, optional display name, and trusted source.
  - Always returns `202` to avoid disclosing whether an address already exists.
  - Sends both a portal magic link and a short code.
- `POST /portals/{portalSlug}/feedback/verifications/confirm`
  - Accepts a one-time token or code.
  - Creates or reuses the portal-scoped verified guest contributor.
  - Returns a contributor session exactly once with `Cache-Control: no-store`.
- `POST /portals/{portalSlug}/feedback/sessions/revoke`
  - Revokes the current contributor session.

The first-party portal may exchange the session for an HttpOnly feedback cookie.
The widget must use `Authorization: FeedbackSession <opaque-token>` because
third-party cookie availability is not reliable. The iframe owns this token and
must never post it to the embedding page.

### UI flow

For logged-out visitors on a `verified_guest` portal:

1. Let the person draft feedback before asking for identity.
2. Ask for email and optional display name at submit time.
3. Preserve the draft while showing **Check your email** and code-entry states.
4. Submit only after verification succeeds.
5. Explain whether the public will see their name.

Lock the participation intent when the composer opens. If an account, guest, or
widget identity appears or expires while the draft is open, require an explicit
retry or choice instead of silently converting the draft to another identity.

For `anonymous_allowed`, present two clearly different paths:

- **Continue with email** — can receive replies and status updates.
- **Submit anonymously** — no name/email and no personal notifications.

### Safety

- Rate-limit by portal, ingress fingerprint, and normalized email.
- Use enumeration-neutral responses.
- Enforce one-time use, short expiry, bounded code attempts, and session
  revocation.
- Do not persist raw IP addresses.
- Apply existing duplicate suggestions before verification so unnecessary email
  challenges are avoided.

## Phase 2: Contributor Comments, Votes, Following, and Delivery

### Canonical contributor identity

Add and backfill `contributor_id` on feedback comments and votes. Change vote
uniqueness from normal-user identity to `(item_id, contributor_id)`. Keep legacy
`user_id` nullable only where existing account notification compatibility needs
it; contributor identity becomes authoritative.

All public write services accept a resolved participant object. Never accept a
user or contributor ID from a public request body.

### Following

Add explicit tables rather than a weak polymorphic subscription:

- `feedback_item_followers`
  - `item_id`
  - `contributor_id`
  - `created_at`
  - `unsubscribed_at`
- `feedback_portal_followers`
  - `portal_id`
  - `contributor_id`
  - `created_at`
  - `unsubscribed_at`

An item author is followed atomically with item creation. When feedback is
merged, move active followers to the canonical item idempotently.

### Delivery

Do not fabricate a normal `users` record or insert guest recipients into the
existing user-FK notification table.

Add a contributor delivery pipeline with:

- Contributor-scoped recipient resolution
- Durable delivery records and a unique event/dedupe key
- Retryable worker tasks
- Per-item and portal-wide unsubscribe controls
- A one-click unsubscribe token stored as a hash
- Delivery status, attempt count, and final failure reason

Add per-item Follow/Unfollow controls and a contributor-scoped guest preference
route. An email management token must be consumed once and then redirect to a
token-free preference URL; guest preferences must not use the existing
workspace-member notification settings API.

Account contributors may continue through the existing notification service.
Verified guest email should use the existing mail worker and SMTP configuration.

Update feedback domain events to carry `recipient_contributor_id`; resolve the
linked `user_id` only when producing a normal in-app notification. Add a status
event bridge for linked-story workflow changes so projected Planned/In progress/
Completed transitions also produce delivery events.

## Phase 3: Published Updates and Changelog

The `feedback_updates` and `feedback_update_items` tables already exist, but
there is no service, repository, HTTP, or usable frontend data path.

### Data model

Add the minimum fields needed for stable public routes:

- Portal-scoped `slug`
- Optional `summary`
- Optional `cover_image_url`
- `published_at`
- `published_by_user_id`

Keep draft and published states. Drafts must never appear in a public snapshot,
search result, widget payload, or cache.

### Admin API and UI

Add workspace-authenticated create, edit, preview, publish, unpublish, and delete
operations under the existing feedback route family. Allow an update to link one
or more feedback items.

Build an Updates management surface with:

- Draft list
- Published list
- Title, summary, body, and optional cover image
- Linked feedback selector
- Preview
- Explicit publish confirmation

Publishing is one transaction that sets the publication timestamp, persists
linked items, and emits one idempotent `feedback.update.published` event.

### Public portal and widget

Add public list/detail endpoints and replace the current hard-coded empty Updates
array. Restore the widget Updates tab only after those endpoints return real
published records.

Until the first published Update exists, hide Updates from public navigation
rather than retaining the current permanently blank page.

Show an unread badge for a contactable contributor based on their last-seen
published timestamp. Anonymous visitors may see Updates but do not get a
personal unread state or email.

## Phase 4: Signed Widget Identity and Origin Allowlist

### Widget settings

Add `feedback_widget_settings` with:

- `portal_id`
- `enabled`
- `widget_key_id`
- `allowed_origins`
- `signing_secret_encrypted`
- `signing_secret_version`
- `created_at`
- `updated_at`

Allowed origins are exact HTTPS origins in the first release. Permit HTTP only
for localhost development. Do not support `*` or ambiguous wildcard domains.

The signing secret is generated per portal, shown once, encrypted at rest, and
rotatable. Keep the previous version for a short, explicit rotation grace
period.

### Customer-side identity assertion

The customer's server signs an HMAC-SHA256 assertion containing:

- Widget key id
- Stable external user id
- Email
- Display name
- Optional avatar URL
- Issued-at and expiry timestamps
- Nonce
- Exact parent origin

Assertions expire within five minutes. The browser receives the signed assertion,
never the signing secret. The loader passes it to the FortyOne iframe, which
exchanges it for a short-lived portal-scoped widget session.

The loader may hold the signed assertion in memory and resend it after the frame
announces `ready`. Never place the assertion in the iframe URL, local storage, or
session storage. The exchanged contributor session remains iframe-only.

Add:

- `POST /portals/{portalSlug}/feedback/widget/sessions`

The endpoint validates signature version, expiry, nonce replay, widget key id,
configured parent origin, and portal binding before creating or reusing an
`external` contributor. Use Redis or a durable nonce record for replay
protection.

### Browser security

- Validate both `event.origin` and `event.source` for every `postMessage`.
- Use exact target origins; never use `*`.
- Generate `frame-ancestors` from the configured allowlist in
  `apps/projects/src/proxy.ts` or a dedicated embed bootstrap response; this is
  not a React-component responsibility.
- Reject an unlisted parent origin before rendering or exchanging identity in
  `app/(public)/embed/feedback/[portalSlug]/page.tsx`.
- Keep writes same-origin inside the iframe; do not open the API CORS policy to
  arbitrary customer origins.
- Never expose contributor session tokens, verification tokens, or encrypted
  signing material to the host page.
- Preserve anonymous intent even if an account or widget session appears while
  a draft is open.

## Environment Variables

### Variables required by the current anonymous-feedback release

| Variable                            | Runtime              | Requirement                                                 | Purpose                                                                                                                                     |
| ----------------------------------- | -------------------- | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `FEEDBACK_INGRESS_SECRET`           | Projects web and API | Required; same value in both; at least 32 characters        | Signs trusted anonymous ingress fingerprints. `setup.sh` generates and preserves it for self-hosted installations.                          |
| `FEEDBACK_TRUSTED_CLIENT_IP_HEADER` | Projects web         | Required only for self-hosted production anonymous feedback | Names a header that a trusted reverse proxy always overwrites with the real client address. Leave blank on Vercel and in local development. |

Do not configure `FEEDBACK_TRUSTED_CLIENT_IP_HEADER` unless a sanitizing reverse
proxy owns and overwrites that header. A directly exposed web container must not
trust a client-supplied forwarding header.

### Next-phase decision

No additional global FortyOne environment variable is required for this plan.

- Verification, contributor sessions, and unsubscribe links use random opaque
  tokens whose hashes are stored in Postgres.
- Email delivery uses the existing `APP_EMAIL_*` variables.
- Public verification links use the existing `APP_WEBSITE_URL`.
- Per-portal widget signing secrets are product data, not deployment-wide
  configuration. Generate them in FortyOne, encrypt them at rest with the
  existing `APP_AUTH_SECRET_KEY` through a domain-separated key derivation, and
  store the encrypted envelope in `feedback_widget_settings`.

The customer embedding the widget will normally put the revealed per-portal
secret in their own server environment, for example
`FORTYONE_FEEDBACK_WIDGET_SECRET`. That is a customer-application variable, not
a new variable required by the FortyOne deployment.

Do not add a separate `APP_FEEDBACK_SECRET_KEY` unless FortyOne adopts a broader
key-management and rotation strategy. Introducing an isolated key without a
rotation and re-encryption path would add operational risk without improving the
token design above.

## Delivery Sequence

1. Add contributor identity, verification, and session migrations behind no
   selectable UI mode.
2. Deploy API support and verification email tasks.
3. Add verified contributor submissions, comments, votes, following, and
   unsubscribe behavior.
4. Expose `verified_guest` in admin settings only after email delivery and
   revocation are working in production.
5. Implement Updates CRUD and public endpoints.
6. Restore Updates in the widget after published data is available.
7. Add widget settings, exact origin allowlists, secret rotation, and signed
   identify.
8. Run live portal, email, and third-party iframe verification before general
   availability.

## Test Plan

### Backend

- Forward migration and backfill tests for contributors, comments, and votes
- Participation-mode matrix across account, verified guest, external, and
  anonymous participants
- One-time, expired, reused, malformed, and wrong-portal verification tests
- Enumeration-neutral verification responses
- Contributor session revocation and cross-portal rejection
- Account and guest vote uniqueness
- Follow-on-create and follower migration after merge
- Delivery dedupe, retries, unsubscribe, and permanent failure behavior
- Linked-story status transition event tests
- Draft Updates excluded from every public query
- Atomic publish and linked-item event tests
- Widget signature, expiry, nonce replay, rotation, portal binding, and exact
  origin tests

### Frontend

- Draft survives verification and failed confirmation
- Clear verified-versus-fully-anonymous choice
- Public masking copy and serializer behavior
- Guest comment/vote/follow behavior
- Unsubscribe and notification-preference states
- Updates draft, preview, publish, public list, and detail flows
- Widget Feedback/Roadmap/Updates navigation once Updates are live
- Loader and iframe executable protocol tests, including hostile origins
- Assertions queued before and after widget initialization
- Assertion absence from iframe URLs, local storage, and session storage
- Responsive popup, mobile sheet, keyboard, focus, and Escape behavior

### Live release checks

- Real SMTP delivery and one-click unsubscribe
- Portal magic-link verification
- Widget code verification in a third-party iframe
- Signed customer identity from a sample server application
- Secret rotation while the prior secret is in its grace period
- Allowed-origin rejection from an unlisted host
- Published Update email, badge, portal page, and widget rendering

## Observability and Operations

Track:

- Verification requested, confirmed, expired, and rate-limited
- Active contributor sessions by source
- Guest submission, comment, vote, and follow counts
- Email queued, sent, bounced, suppressed, unsubscribed, and permanently failed
- Update publish and follower-delivery counts
- Widget session signature failures grouped by safe reason code
- Origin rejection and nonce replay counts

Logs must not contain raw verification tokens, contributor session tokens,
widget signing secrets, or full email addresses.

## Release Gates

The next phase is ready only when:

- Verified guests never create global user accounts.
- A guest can submit, revisit, comment, vote, follow, and unsubscribe without a
  normal account.
- Truly anonymous feedback remains unlinkable and receives no personal
  notifications.
- Every public identity mode has accurate UI copy and API serialization.
- Linked-story status changes generate the same meaningful delivery events as
  direct feedback status changes.
- Draft Updates are impossible to fetch publicly.
- Widget identity works with third-party cookies blocked.
- Origin and replay checks fail closed.
- Existing account-required and anonymous-only portals remain backward
  compatible.
