# Stripe webhook security

## Trust sequence

The public `POST /webhooks/stripe` route is intentionally unauthenticated. It becomes trusted only after this sequence:

1. read at most 64 KiB without altering the signed bytes;
2. require `Stripe-Signature` and verify it with the configured webhook secret;
3. claim the Stripe event ID in PostgreSQL before running a billing effect;
4. process through a bounded lease and provider-identity workspace binding;
5. persist `processed` or `failed` with the exact lease token;
6. return non-2xx for any uncertain state so Stripe retries.

Never parse, normalize, log, or persist the body before signature verification. Logs may include opaque event/type/workspace IDs and attempts; they must not include signatures, customer email, invoice URLs, payment data, or raw payloads.

## Required properties

| Threat                   | Control                                                                                     |
| ------------------------ | ------------------------------------------------------------------------------------------- |
| forged delivery          | exact-body Stripe signature verification                                                    |
| oversized request        | 64 KiB route limit before verification                                                      |
| duplicate/replay         | event-ID primary key and terminal-state deduplication                                       |
| concurrent delivery      | atomic claim and one active lease token                                                     |
| crashed handler          | lease expiry and retryable failed state                                                     |
| stale worker             | compare-and-set terminal update by lease token                                              |
| event-ID type confusion  | immutable event-type comparison on duplicate claim                                          |
| cross-tenant invoice     | workspace plus bound Stripe customer predicate                                              |
| provider ID reassignment | immutable subscription/invoice workspace binding                                            |
| out-of-order state       | current-provider snapshots plus semantic priority cursor; event IDs are not treated as time |
| malicious return URL     | exact configured first-party billing origin                                                 |
| hidden Stripe price      | explicit paid-price lookup-key catalog                                                      |

## Authorization matrix

| Operation                  | Guest                  | Member            | Workspace admin              | Stripe webhook           | System/operator                  |
| -------------------------- | ---------------------- | ----------------- | ---------------------------- | ------------------------ | -------------------------------- |
| Read subscription/invoices | workspace route policy | current workspace | current workspace            | no                       | explicit workspace               |
| Checkout/portal            | denied                 | denied            | allowed in current workspace | no                       | no                               |
| Change/cancel plan         | denied                 | denied            | allowed in current workspace | no                       | explicit use case                |
| Apply provider snapshot    | no                     | no                | no                           | signed provider identity | explicit sync includes workspace |
| Claim/complete inbox       | no                     | no                | no                           | signed event only        | recovery tooling only            |

Billing mutations and access to the Stripe customer portal require a current workspace administrator. Repository tenant enforcement remains a second, independent boundary.

## Failure behavior

- Invalid/missing signatures and malformed bounded bodies return 4xx.
- A delivery already being processed, a handler failure, claim loss, or database failure returns 5xx so Stripe retries.
- A terminal duplicate returns 2xx without running the handler.
- Unsupported signed event types are durably recorded as ignored and return 2xx.
- Cross-workspace provider identity conflicts fail closed and remain visible to operators; they are never silently rebound.

## Review checklist

- failure then retry and duplicate/concurrent delivery tests remain green;
- old events cannot overwrite a newer snapshot or deletion;
- terminal rows carry workspace identity for handled tenant events;
- no SQLx/raw application SQL returns to the module;
- redirect hosts and price lookup keys are never accepted from arbitrary client input;
- checkout and portal bearer URLs never appear in application logs;
- `make sqlc-check`, SQLC database vetting, PostgreSQL 18 race tests, and architecture checks pass.
