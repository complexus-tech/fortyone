# One-time public price increase

Replaces the current USD per-seat prices on their existing Stripe products:

| Lookup key | New unit amount |
| --- | --- |
| `pro_monthly` | $9/month |
| `pro_yearly` | $86.40/year ($7.20/month equivalent) |
| `business_monthly` | $15/month |
| `business_yearly` | $144/year ($12/month equivalent) |

Annual prices retain the existing 20% discount. Enterprise is unaffected.

## Deployment order

1. Deploy the API's `fortyone_lookup_key` metadata fallback first. Transferring a
   lookup key clears it on the old price; historical subscriptions must still be
   recognized by subscription sync and webhook processing.
2. Coordinate the website and in-app pricing deployment with the Stripe change.
   Public prices share `packages/lib/src/subscription-pricing.ts`. The billing
   summary requests `subscription?includePrice=true` to display the current
   Stripe subscription item's actual rate, including legacy prices.
3. From `apps/server`, run `make stripe-pricing-sync` and inspect all four prices,
   product IDs, and `livemode` before applying.
4. Run `make stripe-pricing-sync ARGS="--apply"` against the intended environment.
5. Rerun the dry run to verify the four amounts are unchanged. Verify a test-mode
   checkout and historical subscription synchronization before production rollout.

The command loads `.env` with the same config parser as the application. It uses
`STRIPE_SECRET_KEY`, falling back to `APP_STRIPE_SECRET_KEY`. Credentials select
live or test mode; no credentials are printed. There is no database dependency.

## Behavior

Stripe does not allow a Price's unit amount to be edited. The command validates
all four existing prices before writing, preserves plan identity in metadata,
creates replacement prices, and atomically transfers each lookup key. It keeps
product identity, tax behavior, and metadata. Unexpected billing shapes, missing
keys, inactive products/prices, and extra currencies stop the command for review.

Existing subscriptions retain their existing price IDs and charges. New checkout
and explicit plan/interval changes use the new lookup targets. No subscriptions,
subscription schedules, invoices, or prorations are modified by this command.
Adding or removing members updates only the existing subscription item's quantity,
so new seats retain that subscription's original per-seat rate. Normal seat-change
proration still applies. The existing plan limits remain in effect.

The old prices remain active, without their public lookup keys. This avoids
changing product defaults or existing payment links implicitly; any manually
configured Stripe portal offerings, default prices, or payment links must be
reviewed separately before rollout.

Each price is an independent operation (Stripe has no catalog-wide transaction).
The output records old/new IDs. Completed replacements are skipped on rerun, and
creation requests have stable idempotency keys to recover from uncertain replies.
Do not run concurrent copies. Reverting the source code does not revert Stripe.
