# OKR activities module

This module owns the objective/key-result audit stream. Read
[`docs/database/key-results.md`](../../../docs/database/key-results.md) for its
SQL security contract, deterministic pagination, transaction role, and test
commands.

Package responsibilities:

- `domain`: activity enums, validation, bounded batches, and list queries.
- `service`: the caller-owned repository port and authenticated actor binding.
- `repository`: the native pgx transaction and generated-query adapter.
- `repository/queries`: reviewed, tenant- and membership-scoped SQL.
- `repository/sqlc`: generated code; never edit it by hand.

Key-result create, update, and delete operations write activities inside the
key-result repository transaction. The standalone `CreateBatch` path is also
atomic. Provider-specific notifications and webhook delivery do not belong in
this audit repository.
