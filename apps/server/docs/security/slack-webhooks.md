# Slack webhook security

Slack webhook ingress has two different keys with different jobs:

| Key | Purpose | Source | Where it is used |
| --- | --- | --- | --- |
| `SLACK_SIGNING_SECRET` | Authenticates the exact HTTP request Slack sent | Slack application configuration | API verifier only |
| Slack webhook payload key | Encrypts the verified body retained in the durable inbox | Derived from `APP_AUTH_SECRET_KEY` with a versioned Slack purpose label | API ingress and worker |

Neither key is a Slack OAuth token or provider-credential key. The signing
secret remains an external Slack credential. FortyOne derives the payload key
internally, so operators provision only the same stable application root on the
API and worker.

## Receive and processing boundary

The API reads a bounded copy of the exact request bytes, verifies Slack's
signature and timestamp, resolves the current installation, and then encrypts
the body before the durable receipt is committed. The queue message contains
only the provider and inbox receipt ID; it never contains the body or bot token.

The encrypted `slack-webhook.v2` envelope authenticates all of these database
facts:

- provider (`slack`);
- Slack delivery ID;
- FortyOne workspace ID;
- Slack installation ID; and
- installation generation.

The worker reconstructs that binding from the claimed database row. Moving a
valid ciphertext to a different receipt, workspace, installation, or generation
therefore fails authentication. This is also the reinstall fence: an event from
an old installation cannot be replayed under the new installation generation.

Before applying product behavior, the worker acquires the inbox lease and
rechecks the active installation generation and current actor or team scope.
Completion is idempotent and stores only a bounded outcome code. Raw request
bodies, decrypted payloads, bot tokens, signing material, and provider error
text must never appear in logs or queue payloads.

## Legacy cutover

Normal webhook processing understands only `slack-webhook.v2`. It deliberately
does not fall back to the former `APP_AUTH_SECRET_KEY` ciphertext. This makes a
missed rollout step visible instead of silently extending legacy key access.

During the coordinated replacement-worker startup, the bounded cutover scans
only retryable Slack inbox rows in stable UUID order. For each row it:

1. validates the complete durable identity;
2. decrypts the legacy payload with the old application secret;
3. verifies the Slack `team_id` and `event_id` against the row;
4. reseals the exact body with the derived Slack payload key; and
5. replaces the old value with an exact ciphertext-and-identity compare-and-swap.

A concurrent update wins and is never overwritten. Any malformed identity,
authentication failure, strict decode failure, reseal failure, or database
failure aborts startup before the worker reports ready or consumes jobs. After
all credential and payload legacy counts are zero, remove the old application
secret from the Slack cutover path. Do not add a runtime fallback.

## Rotation and incident response

The payload codec currently uses one derived key version rather than a keyring.
Changing the application root or derivation version requires a coordinated
maintenance window: stop ingress, drain or reseal every retryable retained
Slack payload with a purpose-built bounded migration, deploy API and worker
together, and then resume traffic. Changing the root without handling retained
receipts makes them intentionally unreadable.

If either Slack secret may be compromised, pause Slack ingress and delivery,
replace the affected secret, revoke or reinstall exposed Slack credentials when
appropriate, and inspect safe receipt IDs and generations. Do not copy payloads,
ciphertext, tokens, or keys into tickets or logs.

## Verification

The focused suite must prove exact-body verification, timestamp replay
rejection, request-size bounds, durable persistence before dispatch, terminal
deduplication, wrong-key and ciphertext tamper rejection, every binding-field
swap, stale installation generation rejection, legacy runtime rejection, and
legacy cutover compare-and-swap behavior.

```bash
go test -race -count=1 ./internal/modules/slack/...

TEST_DATABASE_URL='postgresql://<createdb-role>@<postgres-18-host>/<control-db>?sslmode=disable' \
  go test -race -count=1 -tags=integration ./internal/modules/slack/...
```

The PostgreSQL command uses a disposable control database as documented in
[integration test infrastructure](../testing/integration-infrastructure.md).
