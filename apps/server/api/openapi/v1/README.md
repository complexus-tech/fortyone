# FortyOne API v1 contract

`openapi.yaml` is the root OpenAPI 3.1 document. Paths and reusable schemas are
split by responsibility so reviews stay small and ownership remains obvious.

Edit the YAML source, then run `make openapi-generate` from `apps/server`.
Generated Go belongs only in `internal/generated/openapi/v1` and must not be
edited directly. `make openapi-check` is the CI drift check.

Pull-request CI also runs `make openapi-breaking-check` against the immutable
base commit. It uses pinned `oasdiff` compatibility rules and rejects both
definite and potential client-breaking changes. A deliberately incompatible
change therefore needs a version/deprecation decision; it must not be hidden by
regenerating clients in the same pull request.

The same committed contract generates the public preview SDKs. Run
`make sdk-generate` after a reviewed contract change and `make sdk-check` before
handoff. Go client output lives in `sdk/go`; TypeScript output lives in
`packages/sdk-typescript/src/generated`. The handwritten helpers beside those
files may configure transport behavior, but they must never add a contract that
is absent here. See [ADR 0011](../../../docs/architecture/decisions/0011-public-sdk-generation.md).

The generator is pinned to `oapi-codegen` v2.8.0. It produces models, an
embedded bundled spec, strict server request/response adapters, and the Go 1.22
standard-library router. Handwritten HTTP adapters depend on services, never
repositories.
