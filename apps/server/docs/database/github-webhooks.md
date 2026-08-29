# GitHub webhook persistence

GitHub webhook authorization reads are owned by
`internal/modules/github/repository/queries/webhooks.sql` and generated into
`internal/modules/github/repository/sqlc`. Handwritten service and HTTP code do
not import generated SQLC types.

## Migration 000160

`000160_code_host_webhook_fencing` adds two columns to
`github_installations`:

- `installation_generation uuid NOT NULL`: opaque authorization generation;
- `installation_authorized_at timestamptz NOT NULL`: when that generation
  became current.

The generation has a unique index. A `BEFORE UPDATE` trigger rotates it whenever
an authority-bearing installation field changes, including active,
disconnected, or suspended state. The application must never set or predict the
generation.

This migration is additive and should be applied schema-first. Existing API
instances ignore the columns. Replacement webhook workers require the columns
and must not start until the migration is visible on their database connection.

The down migration removes the trigger, function, index, and columns. Do not
roll it back while replacement workers or inbox records rely on generation
fencing; deploy a forward fix instead.

## Generated queries

`GetAuthorizedWebhookInstallation` resolves a signed external installation and
repository only when:

- the repository belongs to the same workspace and installation;
- the installation is active, not suspended, and not disconnected; and
- the repository is active.

`GetCurrentWebhookInstallation` additionally requires the internal
installation ID and exact generation captured by the durable inbox receipt.
This is the worker's time-of-use fence.

Both queries are static, parameterized SQLC statements over native pgx. Missing
or stale rows map to `githubshared.ErrWebhookInstallationNotFound`; callers use
that stable domain error instead of pgx or generated row types.

An infrastructure error is not mapped to the not-found sentinel. Receipt-time
verification preserves it as `webhooks.ErrVerificationUnavailable`, producing
a retryable 503 rather than a misleading 401. Worker-time lookup failures also
remain retryable; only an actual missing or stale grant is safely cancelled.

## Durable inbox relationship

The raw body is not stored in GitHub-specific tables. The shared
`messaging_inbound_events` inbox stores:

- immutable provider and delivery identity;
- workspace, installation, and generation;
- encrypted payload and bounded expiry;
- processing status, attempts, lease/recovery metadata, and safe outcome.

The encrypted payload contains a second authenticated binding to the same
identity. Copying valid ciphertext to another receipt, workspace, provider, or
generation fails closed.

## PostgreSQL 18 contract coverage

`webhooks_integration_test.go` creates an isolated fully migrated database and
proves:

- active installation/repository authorization resolves;
- deactivation makes the grant unavailable;
- the database rotates generation on deactivation and reauthorization;
- work captured under the old generation cannot resolve; and
- the new generation resolves after reauthorization.

Run it with:

```sh
TEST_DATABASE_URL='postgres://.../postgres?sslmode=disable' \
  go test -tags=integration -count=1 ./internal/modules/github/repository
```

## Complete GitHub SQLC scope

Webhook authorization is one part of a fully native GitHub repository. The same
module-local generated package now owns settings, installation/repository
reconciliation, sync links and workflow rules, story/comment ledgers, identity
mapping, credential compare-and-swap, and workspace authorization. Every
handwritten query lives under `internal/modules/github/repository/queries`, and
all service-facing values are mapped to GitHub-owned records before leaving the
repository adapter.

No new migration was required for this migration wave. The existing
`github_story_links_unique_external_ref` expression index is used directly by
the typed upsert. Historical migrations `000001` through `000169` remain
unchanged. The exact query inventory and remaining provider-lifecycle work are
maintained in the
[GitHub integration implementation inventory](../integrations/github-inventory.md#persistence-inventory).
