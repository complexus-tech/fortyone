# Public SDK generation

This guide explains where the preview clients come from and how to change them
without creating a second API contract. Start with the public
[`/api/v1` contract](../../api/openapi/v1/README.md) and
[ADR 0011](../architecture/decisions/0011-public-sdk-generation.md).

## Ownership map

| Concern                           | Source or generated location                                 | Edit directly?               |
| --------------------------------- | ------------------------------------------------------------ | ---------------------------- |
| Public paths and schemas          | `api/openapi/v1/**/*.yaml`                                   | Yes, through API review      |
| Bundled contract                  | temporary generation directory                               | No                           |
| Go generated client               | `sdk/go/client.gen.go`, `metadata.gen.go`                    | No                           |
| Go transport helpers              | `sdk/go/*.go` excluding `*.gen.go`                           | Yes                          |
| TypeScript generated declarations | `packages/sdk-typescript/src/generated`                      | No                           |
| TypeScript helpers                | `packages/sdk-typescript/src` excluding `generated`          | Yes                          |
| Runnable consumer                 | `examples/external-integration`                              | Yes; public SDK imports only |
| Tool pins                         | `scripts/openapi-common.sh`, `tools/sdk.lock`, SDK manifests | Yes, deliberately            |

The server, Go client, and TypeScript client all consume the same OpenAPI graph.
The SDK metadata generator also reads the bundled contract, so the package
version and default base URL cannot drift into handwritten constants.

## Make a contract change

1. Change the smallest YAML source under `api/openapi/v1`.
2. Run `make openapi-generate sdk-generate` from `apps/server`.
3. Review generated server and both client diffs. An unexpected SDK diff usually
   means the contract change is broader than intended.
4. Update the public docs and add a server contract test before changing an
   ergonomic helper.
5. Run `make openapi-check sdk-check`. The full `make generated-check` also
   verifies SQLC, configuration, route inventory, and migrations.

`sdk-check` does not write the repository. It regenerates into a temporary
directory, compares all four artifacts, verifies both nested Go modules are
tidy, runs their race tests, then type-checks and tests TypeScript.

## Allowed handwritten behavior

Helpers can make a documented operation safer to call:

- put one bearer credential in the `Authorization` header;
- require HTTPS except for explicitly enabled loopback tests;
- deny redirects and bound request timeouts;
- retry only `GET`/`HEAD` for network failures, `429`, or `503`;
- generate and validate cryptographically random `Idempotency-Key` values for
  operations whose OpenAPI contract requires them;
- retain error code, status, request ID, field errors, and `Retry-After` without
  retaining a raw response that might contain sensitive data;
- follow opaque cursors while detecting missing or repeated cursors;
- verify the exact raw webhook body, timestamp window, and every rotation
  signature before JSON decoding.

A helper cannot create an unpublished resource, loosen authorization, infer an
OAuth flow, or declare a write retryable without a proved server contract. The
story-create contract now supplies that server guarantee. The Go
`NewIdempotencyKey` and TypeScript `createIdempotencyKey` helpers generate one
valid key, but transports still do not retry `POST` automatically: only the
caller can durably retain the key, exact serialized request, retry budget, and
business recovery state.

For an idempotent write, generated clients must expose the required header and
typed request without normalizing away the server's exact-byte semantics. SDK
documentation must tell callers to generate once, persist before the first
attempt, reuse the same key and bytes after an unknown outcome, honor
`Retry-After`, and stop on `idempotency_key_reused`. A convenience retry helper
is not allowed unless it accepts durable operation state rather than silently
generating a new key in memory.

## Upgrade a generator

Read the generator's official release and migration notes, update the exact
pin, install dependencies, and regenerate once. Treat every generated change as
a contract review. Run the focused SDK checks and the complete server generated
gate. Never upgrade the generator and redesign the public schema in the same
change unless the coupling is unavoidable and explicitly documented.

The current TypeScript generator is paired with TypeScript 5.9.3 because its
published peer contract is TypeScript 5.x. Do not silently substitute the
monorepo's newer compiler. The selected Go runtime currently requires Go 1.24;
the server itself remains on Go 1.25.
