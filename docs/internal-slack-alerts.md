# Internal Slack alerts

The existing FortyOne Slack app posts two operations alerts to the internal
workspace's general channel:

- A new account: name, email, signup time.
- A successful, live, nonzero invoice payment: billing customer, email, amount
  and currency, workspace, payment time, and Stripe dashboard invoice link.

Signing in, linking an external identity to an existing account, creating a
workspace, and joining a workspace do not produce alerts. There is no historical
account or invoice backfill. Account notifications cover both application
registration and new external-identity accounts.

## Enable

1. Apply migration `000188_internal_slack_alerts` before deploying the API or
   worker changes. Signup and invoice writes persist their pending alerts in the
   same database statement/transaction as the business change.
2. Ensure the existing Slack app is connected to internal team `T015B85FC6R`
   through FortyOne's Slack integration. The worker uses that installation's
   existing encrypted credential; no additional bot token is needed.
3. Add `invoice.payment_succeeded` to the existing **live** Stripe webhook
   endpoint's selected events. Keep its existing events, including `invoice.paid`.
   Both reconcile invoices, but only the successful-payment event records an
   alert. This excludes test events, zero-value invoices, and manual/out-of-band
   settlement. See [Stripe's invoice event semantics](https://docs.stripe.com/invoicing/integration).
4. Set these variables on the worker and restart it:

   ```dotenv
   APP_INTERNAL_SLACK_ENABLED=true
   APP_INTERNAL_SLACK_TEAM_ID=T015B85FC6R
   APP_INTERNAL_SLACK_CHANNEL_ID=C014XSVSRF1
   ```

The worker must consume the `notifications` queue. The existing app's requested
`chat:write`, `chat:write.public`, and `channels:read` permissions cover delivery
and channel verification, subject to the installed token actually having those
scopes. The dispatcher checks the token's team and verifies that the channel is
the unarchived, unshared general channel before posting. It never chooses a
destination from the signup, invoice, or customer workspace.

## Delivery and disabling

The worker claims one pending alert every five seconds. Concurrent workers use
database leases, failed attempts use bounded exponential backoff and Slack's
`Retry-After`, and expired leases are recoverable. Signup deduplication uses the
user ID; payment deduplication uses the Stripe invoice ID. Repeated webhook
deliveries cannot create another alert for the same invoice.

Successful deliveries keep a receipt and clear their personal-data payload.
Deleting the source user/workspace also deletes its alert records. Receipts
remain to protect against delayed webhook replays. The Slack message itself
follows the internal workspace's Slack retention policy.

Set `APP_INTERNAL_SLACK_ENABLED=false` and restart the worker to stop scheduling
and claiming internal alerts. This leaves customer-facing Slack features
unchanged. Disabling pauses delivery: pending alerts continue to be recorded,
and enabling later resumes them. A destination is fixed on its first delivery
attempt, so changing configuration cannot reroute already-attempted alerts to
another team/channel.

As with any external message API, a crash after Slack accepts a message but
before its receipt is saved creates an uncertain outcome. Retries reuse the
same `client_msg_id`; this is not a claim of guaranteed exactly-once delivery
across Slack and PostgreSQL.

## Verification

Local tests use a mock Slack server and disposable PostgreSQL databases. They
cover signup rollback, external-identity reuse, no workspace alerts, invoice
deduplication and atomicity, concurrent claims, stale leases, destination
validation, rate limits, message escaping, and disabling. These tests do not
send real Slack messages or prove deployed configuration.

To remove the feature later, remove the worker registration, signup outbox CTEs,
and invoice outbox CTE/payment metadata before dropping its table with a new
forward migration. The customer Slack routes and installation lifecycle do not
depend on this dispatcher.
