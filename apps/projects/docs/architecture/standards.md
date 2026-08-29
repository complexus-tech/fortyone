# Projects frontend engineering standards

The governing plans under `docs/plans/projects-modernization` define the target
and delivery sequence. [ADR 0001](decisions/0001-modular-frontend-boundaries.md)
fixes the modular-frontend decision. This page is the short review contract for
new code and each migrated behavior slice.

The source migration is not complete. Rules marked as target evidence require
the Phase 0 architecture gate before they can be called mechanically enforced.

| Concern                 | Required implementation                                                                                                                                                 | Evidence                                                                                       | Adoption rule                                                                                |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| Layer direction         | `app -> shell/modules -> shared`; modules never depend on `app` or `shell`; shared never depends on a feature                                                           | Target import-graph gate and cycle report                                                      | Reject new violations immediately; migrate existing edges slice by slice                     |
| Module ownership        | A product capability owns its UI, domain vocabulary, data adapters, query/cache behavior, realtime semantics, and focused tests                                         | Module inventory plus adjacent tests                                                           | Do not add an empty layer or a second home merely for symmetry                               |
| Cross-module use        | Consume a narrow public browser/server/type/capability boundary; never another module's private internals                                                               | Target public-boundary allowlist                                                               | Add a temporary adapter only with owner and removal condition                                |
| Imports and bundles     | Use direct relative imports inside an owner; use explicit curated boundary files across owners; no broad root barrels                                                   | ESLint/import graph and bundle inspection                                                      | Prefer direct or dynamically loaded heavy entrypoints; do not widen a facade for convenience |
| Server/browser boundary | Server adapters use `server-only`; client graphs contain no session token, cookie, secret, Node API, or server-only dependency                                          | Type/import checks, build, and boundary tests                                                  | Move a complete request path; do not duplicate server and browser policy                     |
| Authentication          | Every externally reachable server action/route handler authenticates and validates like an API endpoint; request-scoped helpers may deduplicate work but not authority  | Negative auth/workspace tests and request inventory                                            | Preserve fail-closed behavior during migration                                               |
| Outbound requests       | Caller-controlled server fetches validate every destination and redirect, reject private/metadata networks, and bound time, bytes, and content type                     | IPv4/IPv6/DNS/redirect tests plus network egress policy                                        | URL parsing or caller authentication alone is not an SSRF boundary                           |
| Data fetching           | Start independent server work early and await together; avoid client waterfalls; minimize serialized client props                                                       | Focused render/request tests and trace/bundle evidence where risk warrants                     | Do not introduce a new fetch wrapper when an owning adapter exists                           |
| Query ownership         | One resource owner defines keys, normalized inputs, query options, mutation/cache policy, and invalidation semantics                                                    | Target literal-key/duplicate-owner check plus cache tests                                      | Existing `src/constants/keys.ts` and local keys migrate by resource, not wholesale           |
| Realtime                | Shell owns connection lifecycle; modules own runtime-validated event types and shape-aware cache reducers                                                               | Decoder/reducer tests for stale, duplicate, order, tenant, grouped, infinite, and detail cases | Unknown or unsafe events fall back to bounded invalidation, never unchecked cache mutation   |
| Mutations               | Validate input, call one owned request boundary, model optimistic state deliberately, roll back exactly, then reconcile with authoritative data                         | Mutation and cache-transition tests                                                            | Reuse the owner's cache policy across UI, Maya, and realtime paths                           |
| Rendering               | Keep client boundaries as low as practical, avoid unnecessary serialized data, parallelize independent server fetches, and dynamically load genuinely heavy optional UI | Type check, build, focused performance evidence                                                | React Doctor is advisory and cannot replace source review                                    |
| Components              | Feature components stay with the feature; broadly reusable visual primitives use established `ui`/`icons` packages or a stable app-local shared owner                   | Story/component tests and usage inventory                                                      | Similar markup alone is not a shared semantic contract                                       |
| Tests                   | Pure logic, request policy, query/cache, component interaction, server boundary, and browser journey tests match the risk of the slice                                  | Jest after baseline repair, type check, build, and focused browser tests                       | Test files do not count as evidence until the runner executes them                           |
| Documentation           | Update ownership, public boundary, cache/event contract, migration status, and evidence with the source change                                                          | Documentation review and generated inventories                                                 | Do not mark an unrun check or unmigrated module complete                                     |

## Canonical target shape

```text
src/
├── app/                         Next.js route adapters only
├── shell/                       providers, workspace shell, hydration, realtime
├── modules/<capability>/        feature-owned vertical slice
│   ├── public/                  explicit client/server/type/capability boundaries
│   ├── api/browser/             browser-safe request adapters
│   ├── api/server/              server-only request adapters
│   ├── model/                   domain types, schemas, and pure policy
│   ├── queries/                 owned keys and query-options factories
│   ├── mutations/               mutation and cache reconciliation policy
│   ├── realtime/                validated events and cache reducers
│   └── ui/                      feature components
└── shared/                      stable feature-neutral foundations
```

This is a responsibility map, not a demand that every module contain every
directory. Small modules omit layers they do not own. Existing names can remain
where they already satisfy the contract.

## Architecture debt rule

The reviewed 16-module strongly connected component is a ceiling, not precedent.
The future architecture baseline records exact production edges. Normal updates
may remove findings but may not add a module to a cycle, introduce a forbidden
layer edge, leak a server-only import, or create a second query owner. A rule
change requires an ADR; refreshing a baseline is never the way to make growth
green.

## Review rule

A temporary exception names the exact import or operation, owner, behavioral
reason, tests, expiry/removal condition, and governing decision. File size,
deadline pressure, or migration convenience is not sufficient. A source move is
complete only after all callers use the new owner and the obsolete path is
deleted.
