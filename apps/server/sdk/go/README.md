# FortyOne Go SDK preview

This nested module is generated from the committed FortyOne OpenAPI v1
contract. It is a repository preview, not a published or compatibility-backed
Go module release.

Use `fortyone.New` for bearer authentication, HTTPS validation, redirect
protection, bounded safe-read retries, and the generated response-aware client.
`NewStoryPager` follows opaque story cursors. `NewWebhookVerifier` verifies the
exact raw body before JSON decoding.

`NewIdempotencyKey` generates a cryptographically random key for a supported
write, and `ValidateIdempotencyKey` checks retained values against the public
header contract. Generate one key per logical mutation and persist it with the
exact serialized request before the first attempt. The generated
`CreateStoryWithResponse` operation requires a PAT with `stories:write` and the
key in `CreateStoryParams`. The transport never retries `POST`; callers own the
retry budget and must reuse the same key and request bytes.

Regenerate from `apps/server` with `make sdk-generate`; verify drift, module
tidiness, race tests, and the consumer sample with `make sdk-check`. Never edit
`client.gen.go` or `metadata.gen.go` directly.
