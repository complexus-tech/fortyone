# Feedback contributor delivery security

This document is the security and recovery contract for contributor
verification, unsubscribe links, feedback widget secrets, and contributor email
delivery. These capabilities share one feedback-owned cryptographic root, but
do not reuse browser authentication, developer credentials, OAuth, or the
provider credential vault.

## Key boundary

`APP_FEEDBACK_SECURITY_KEY` is required by both the API and worker. Production
startup requires at least 32 random bytes, rejects the checked-in development
value, and rejects reuse with every configured auth, verification, invitation,
OAuth-signing, developer-credential, or credential-vault key.

The feedback service derives purpose-separated keys for:

- contributor verification-code digests;
- deterministic unsubscribe-token digests;
- feedback widget signing-secret encryption.

The purpose labels are versioned and domain-separated. Derived material never
crosses these capability boundaries. `FEEDBACK_INGRESS_SECRET` is a different
secret used only to authenticate the Projects application's public feedback
rate-limit proof; it must not be reused as the feedback security key.

## Delivery and queue invariant

PostgreSQL owns the durable delivery record and stores only the keyed digest of
the unsubscribe token. Redis/Asynq receives this identity-only payload:

```json
{ "deliveryId": "<uuid>" }
```

No unsubscribe token, email credential, provider token, or cryptographic key is
placed in a queue payload. The worker atomically claims the delivery, rechecks
that the contributor remains eligible, loads the unconsumed and unexpired
digest, reconstructs the deterministic token from its server-held key, and
compares the digest in constant time before building the email URL.

Blocked, unsubscribed, expired-token, missing-token, and otherwise ineligible
recipients are marked `suppressed` before any email is sent. A successful send
transitions `processing` to `sent`. Provider failure records a bounded reason
and transitions to `retrying` or `failed` according to the Asynq retry budget.
The stored and task-visible reason is an application-owned category; raw SMTP
or provider errors are not persisted or returned because they may contain a
recipient address, provider response, or credential-bearing diagnostic.
Recovery re-enqueues only eligible queued/retrying deliveries and processing
leases older than 15 minutes. A digest-integrity failure is terminal and never
reconstructs or emits a URL.

Worker handlers depend on the feedback-owned `ContributorDeliveryStore`; they
do not import SQLx, execute SQL, or create a database fallback. If bootstrap
does not supply the pgx repository, delivery and recovery fail closed before
touching the mail provider. Scheduled retention and digest jobs follow the same
port-injection rule, including all five contributor/widget artifact cleanup
statements.

## Rotation and incident response

Changing `APP_FEEDBACK_SECURITY_KEY` immediately invalidates outstanding
verification and unsubscribe tokens and prevents decryption of existing widget
signing secrets. Therefore this key is not rotated as an uncoordinated
environment change.

For a planned rotation:

1. pause feedback publication and contributor-email dispatch;
2. deploy a reviewed keyring/rewrap implementation that can read the old
   generation and write only the new generation;
3. re-encrypt widget signing secrets and verify counts/checksums;
4. allow outstanding short-lived verification and unsubscribe tokens to expire
   or explicitly invalidate and regenerate them;
5. deploy API and worker with the same active generation;
6. resume dispatch, verify suppression/send/recovery metrics, then retire the
   old generation after the documented overlap.

Until that keyring work exists, an emergency key change is a deliberate
fail-closed incident action. Expect contributors to request fresh verification
links and administrators to rotate widget signing secrets.

## Verification checklist

- `make config-check` confirms API, worker, `.env.example`, and generated
  configuration documentation agree.
- Unit tests prove the task payload contains only `deliveryId` and that token
  reconstruction fails under the wrong key or digest.
- Service tests prove only the digest is persisted and only the delivery ID is
  enqueued.
- Worker tests assert blocked/unsubscribed/expired recipients are suppressed
  by the claim query before delivery.
- Logs, traces, queue inspection, and errors must never include raw contributor
  tokens, destination query credentials, email bodies, or configured key
  material.
