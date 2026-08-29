# ADR 0001: Modular frontend boundaries

- Status: Accepted
- Date: 2026-08-28
- Owner: Projects frontend engineering

## Context

FortyOne Projects is one Next.js application with approximately 185,000 lines of
TypeScript/TSX across routing, workspace shell behavior, product modules, server
actions, browser data hooks, realtime updates, and shared UI. Product-oriented
modules are a useful foundation, but the reviewed import graph contains a
strongly connected component spanning 16 modules. Server/browser request logic,
query keys, cache updates, and reusable helpers have multiple competing homes.

Splitting the application into independently deployed microfrontends would add
runtime integration, routing, release, shared-state, and design-system failure
modes without resolving the ownership problems inside the current source graph.
A big-bang folder rewrite would create a large compatibility branch and make
behavioral regressions difficult to attribute.

## Decision

Keep Projects as one modular Next.js application. Organize production code by
the following dependency direction:

```text
app -> shell and module public boundaries -> shared foundations
```

The layers mean:

1. `app` owns Next.js route files, layouts, route handlers, metadata, and route-
   level composition. It does not own feature policy or reusable cache behavior.
2. `shell` owns application-wide runtime composition: workspace shell,
   providers, navigation, hydration boundaries, global overlays, and realtime
   connection lifecycle.
3. `modules/<capability>` owns one product capability's UI, domain types,
   server/browser data adapters, query options, mutations, cache policy, and
   focused tests.
4. `shared` owns stable app-wide technical primitives with no feature semantics.
   Workspace packages such as `ui`, `icons`, `api-client`, and `auth` remain
   explicit external foundations.

Dependencies point downward. `shared` never imports a feature, shell, or route.
Modules never import `app` or `shell`. Cross-module use goes through a narrow,
curated public boundary owned by the providing module or a consumer-owned
capability contract. A module must not import another module's private component,
hook, action, query key, or cache implementation.

Curated boundaries are not convenience barrels. Inside a module, use direct
relative imports. Across modules, import only an explicit browser, server, type,
or capability entrypoint. Do not create a root `index.ts` that re-exports an
entire feature and accidentally broadens bundles or contracts.

Server-only and browser-safe graphs are explicit:

- server adapters use `server-only`, may read authenticated request context, and
  never enter a client component graph;
- browser adapters contain no server token, cookie, secret, or Node-only import;
- server actions and route handlers authenticate and validate inputs as external
  mutation boundaries; and
- shared types are transport-neutral rather than request-session objects.

Each server-state resource has one owner for its query-key factory, normalized
inputs, query-options factory, mutation/cache policy, and realtime reducer. SSR
prefetch, hydration, client hooks, mutations, Maya invalidation, and realtime
updates consume that same contract. Realtime events are runtime validated before
dispatch; the shell owns the connection, while modules own event meaning.

Adoption is incremental. Migrate one complete behavior slice, characterize its
current behavior, preserve its route/API/permission/UX contracts, switch all
callers, and delete the superseded path. The existing 16-module cycle becomes a
ratcheting debt baseline: it may shrink, never grow.

## Enforcement and adoption

- Add a static import-graph gate for layer direction, cross-module private
  imports, server-only leakage, and newly introduced cycles.
- Record the reviewed debt by exact edge/path. A baseline update may accept only
  a reduction unless a separate ADR changes the rule.
- Add query-owner checks that reject unmanaged literal keys in migrated slices.
- Add server/client boundary tests and bundle checks for designated public
  entrypoints.
- Add runtime event decoding plus reducer tests for duplicate, stale,
  out-of-order, workspace-mismatched, grouped, infinite, and detail cache shapes.
- Keep React Doctor advisory. Type checking, focused Jest tests, production build,
  and browser evidence remain the acceptance gates appropriate to each slice.

The detailed adoption sequence is in
[`02-delivery-testing-and-documentation-roadmap.md`](../../../../../docs/plans/projects-modernization/02-delivery-testing-and-documentation-roadmap.md).

## Consequences

Feature ownership and server/browser safety become visible, query/cache behavior
has one source of truth, and source migration can proceed without a release fork.
Some temporary adapters and compatibility entrypoints will exist. They must be
named as migration debt, have an owner and removal condition, and cannot become a
second permanent architecture.

The architecture permits intentional cross-feature workflows; it does not permit
arbitrary internal imports. A small explicit adapter may add code, but it keeps
the dependency and bundle contract reviewable.

## Alternatives rejected

- **Microfrontends now:** operational complexity without first fixing ownership.
- **Technical layer-wide folders:** hides product ownership and recreates broad
  `hooks`, `services`, or `utils` dumping grounds.
- **Big-bang rewrite:** weak behavioral attribution, long-lived divergence, and
  an unsafe merge/deployment boundary.
- **One global query/cache service:** centralizes feature semantics and preserves
  the same coupling under a different name.

## Revisit when

A capability has independent product ownership, release cadence, runtime
isolation, performance budget, and deployment need that cannot be met safely in
the modular application. Source size or a difficult import graph alone is not a
microfrontend justification.
