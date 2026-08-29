# FortyOne TypeScript SDK preview

This private workspace package is generated from the committed FortyOne
OpenAPI v1 contract. It is not published to npm and does not yet carry a public
compatibility promise.

`createFortyOneClient` adds bearer authentication, validates HTTPS base URLs,
and retries only safe reads. The package also exports safe structured errors,
opaque-cursor story iterators, and exact-byte webhook verification for
server-side runtimes.

`createIdempotencyKey` generates a cryptographically random key for a supported
write, and `validateIdempotencyKey` checks a retained key. Generate one key per
logical mutation and persist it with the exact serialized request before the
first attempt. The typed story-create operation requires a PAT with
`stories:write`. The transport never retries `POST`; callers own the retry
budget and must reuse the same key and request bytes.

From the repository root, run:

```sh
pnpm --dir packages/sdk-typescript type-check
pnpm --dir packages/sdk-typescript test
```

Generation is owned by `apps/server`: run `make sdk-generate` or the read-only
`make sdk-check` there. Never edit `src/generated` directly.
