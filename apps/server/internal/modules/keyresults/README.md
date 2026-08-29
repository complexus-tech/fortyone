# Key results module

Start with [`docs/database/key-results.md`](../../../docs/database/key-results.md)
for the data model, security rules, transaction boundaries, typed patch model,
integration seam, and verification commands.

Package responsibilities:

- `domain`: dependency-free values, validation, access scope, typed commands,
  and typed update intent.
- `service`: use cases plus caller-owned `Repository` and `EventPublisher`
  ports. No SQLC types belong here.
- `repository`: handwritten pgx adapter and transaction ownership.
- `repository/queries`: reviewed SQLC input. All workspace/resource access is
  scoped here as well as in the service.
- `repository/sqlc`: generated code. Never edit it by hand.
- `http`: transport decoding, error-to-status mapping, response conversion, and
  cache invalidation after successful mutations.

When adding behavior, begin in the domain/use case and keep provider-specific
integration code outside this module.
