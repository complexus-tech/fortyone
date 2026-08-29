# FortyOne Projects Modernization: Current State and Target Architecture

**Status:** Active source of truth; Phase 0 foundation and first invitations
slice in progress  
**Scope:** `apps/projects`  
**Audience:** Projects engineers, reviewers, product engineers, and integration
authors  
**Decision:** Keep one modular Next.js application and migrate incrementally  
**Implementation status:** [evidence ledger](./00-implementation-status.md)  
**Delivery plan:** [delivery, testing, and documentation roadmap](./02-delivery-testing-and-documentation-roadmap.md)

## 1. Executive decision

Keep Projects as a **modular frontend**, not a collection of microfrontends and
not a layer-wide rewrite. The current challenge is not Next.js, React, or the
single deployment. It is inconsistent ownership across routes, the workspace
shell, feature modules, app-local shared code, server/browser requests, query
keys, cache updates, and realtime behavior.

The target makes the normal path predictable:

```text
Next.js route
  -> application shell or module public boundary
  -> module use case / query options / UI
  -> explicit browser-safe or server-only API adapter
  -> shared technical foundation or workspace package
```

Dependencies point in one direction:

```text
app -> shell/modules -> shared
```

These documents establish the destination and review rules. The initial
architecture ratchet and invitations ownership slice begin adoption, but they do
not mean that the wider source tree already follows the target.

## 2. What was reviewed

The current-state analysis covered:

- all TypeScript and TSX under `apps/projects/src`;
- the Next.js App Router tree and workspace layouts;
- 26 top-level feature directories under `src/modules`;
- root-level `components`, `context`, `hooks`, `lib`, `types`, `utils`, and
  `constants` ownership;
- server components, route handlers, and server actions;
- browser-side queries, mutations, optimistic updates, and invalidation;
- React Query prefetch/dehydrate/hydrate behavior;
- the server-sent-events connection and cache update logic;
- authentication call sites and backend request adapters;
- route handlers that perform server-side outbound requests;
- 218 Jest test files and the Jest startup failure in the initial analysis
  snapshot, plus the passing TypeScript check; and
- React Doctor output as an advisory diagnostic.

The initial quantitative snapshot was approximately 185,000 TypeScript/TSX lines
in 1,622 files. Static analysis found one strongly connected component
containing 16 feature modules and classified about 250 authenticated request
paths. The implementation ledger records current post-foundation counts and
verification separately so the initial score is not silently rewritten.

The baseline score is **4.7/10 for source and architecture readiness**:

| Dimension            |  Score |
| -------------------- | -----: |
| Architecture         | 4.0/10 |
| Correctness          | 6.0/10 |
| Security             | 5.5/10 |
| Performance          | 5.0/10 |
| Testing              | 4.0/10 |
| Maintainability      | 4.0/10 |
| Developer experience | 4.5/10 |

This is not a product-quality or feature-quality score. The detailed rationale
and evidence limitations are in the implementation-status ledger.

### 2.1 Current adoption checkpoint

ADR 0001 is accepted. Phase 0 now has a dependency-free scanner, eleven scanner
tests, a guarded exact-path debt baseline, deterministic Jest discovery, and a
pull-request workflow for the architecture, type, Jest, and build gates. The
first invitations ownership slice moves invitation keys and duplicate request
paths into `src/modules/invitations`, switches the known route, onboarding,
layout-prefetch, and UI callers, and deletes four superseded `src/lib` files.

This is an incremental source cutover, not a claim that invitations or Projects
already satisfy every target rule. The complete Jest run, combined type check,
production build, focused changed-scope ESLint, and changed-scope React Doctor
pass; React Doctor reports 90/100 with no issues across 29 changed files. The
broad lint baseline remains red with 240 findings in the final run, the full-app
React Doctor baseline remains 57/100, and semantic architecture, request,
permission, cache, and browser evidence still must be completed before
invitations can be called architecture-complete.

## 3. What should remain

The modernization preserves and strengthens these choices:

1. **One Next.js application.** A single route tree, design system, session,
   cache runtime, and deployment are appropriate for the product today.
2. **Product-oriented modules.** `stories`, `objectives`, `sprints`, `teams`,
   `strategy`, `feedback`, and similar names express ownership better than one
   application-wide `services` or `hooks` layer.
3. **React Server Components and server actions.** These are useful boundaries
   when server-only imports, authentication, serialization, and error behavior
   are explicit.
4. **React Query.** It is the browser server-state runtime. The problem is
   duplicated ownership, not the library.
5. **Workspace packages.** `ui`, `icons`, `api-client`, `auth`, and other shared
   packages are established foundations and should be consumed intentionally.
6. **Strict TypeScript.** The passing type check is a meaningful baseline and
   must not be weakened to make migration easier.
7. **Existing UX and routes.** A structural slice preserves permissions, URLs,
   API behavior, loading/error/empty states, cache semantics, and user workflows
   unless a product change is separately specified.

## 4. Problems the target must solve

### 4.1 Ownership is not predictable

Feature behavior can live in `src/modules`, `src/lib/actions`, `src/lib/queries`,
`src/lib/hooks`, root `components`, root `context`, or an App Router file. Some
features are vertical slices; others depend on a parallel global implementation.
An engineer cannot always answer from the path alone:

- who owns a backend resource;
- which query key is canonical;
- whether a function is safe in a client graph;
- which mutation or realtime handler updates its cache;
- which server boundary authenticates it; or
- which tests prove workspace and failure behavior.

### 4.2 Feature imports form a large cycle

The reviewed graph has a 16-module strongly connected component. A cycle of that
size obscures change impact, makes extraction and testing difficult, encourages
private imports, and allows accidental bundle coupling. The goal is not to ban
all cross-feature workflows; it is to make their direction and contracts
explicit.

### 4.3 Server and browser code are easy to mix

Next.js uses one TypeScript tree for server components, actions, route handlers,
and browser bundles. Without explicit entrypoints, a helper can acquire cookies,
tokens, Node APIs, or server-only dependencies and later be imported by a client
component. Conversely, browser concerns can spread into server data functions.

Separate static analysis finds 244 executable `await auth()` calls across 225
production files. That heuristic inventory shows that authentication is taken
seriously, but also that request context is assembled repeatedly. The target
deduplicates request-local work without hiding or weakening authorization.

### 4.4 Query identity and cache semantics have multiple owners

Query keys exist in `src/constants/keys.ts`, feature constants, local hook
objects, and array literals. Query functions and mutations are also split
between global `lib` paths and modules. A key is part of a data contract: its
shape decides deduplication, hydration, optimistic state, invalidation, and
realtime updates. Duplicate ownership can make a mutation appear successful in
one view while another remains stale.

### 4.5 Outbound request policy is incomplete

`src/app/api/metadata/route.ts` accepts a caller-supplied URL and fetches it on
the server with `cheerio.fromURL`. Authentication limits who can invoke the
route, but does not make the destination safe. The handler has no explicit
private-network, loopback, link-local, cloud-metadata, redirect-hop, content
type, or response-size policy. This is an SSRF boundary and should be remediated
before lower-risk structural cleanup.

### 4.6 Realtime behavior is centralized but not typed end to end

`src/app/server-sent-events.tsx` owns the browser connection, parses arbitrary
JSON, classifies several domain events, and directly mutates story caches while
invalidating other resources. Local type assertions do not validate runtime
payloads. Cache patching also has to understand every possible result shape,
including detail, flat list, grouped, paginated, and infinite data.

### 4.7 Broad feedback is not reliable enough

`pnpm --filter projects type-check` passes on the combined worktree. The
default Jest command originally failed before discovery with `TypeError: Cannot
read properties of undefined (reading 'fileExists')`; the CommonJS `next/jest`
configuration now reaches deterministic discovery. The complete combined-
worktree run passes 220/220 suites and 1,279/1,279 tests with 0 snapshots in
33.601 seconds. The production build also passes, compiling in 39.1 seconds,
type-checking in 11.5 seconds, and generating 27 static pages. Focused ESLint over
every changed app and scanner file passes; the final broad lint run remains red
with 160 errors and 80 warnings. These are strong static/build gates, not a
production-readiness or complete runtime-verification claim. React Doctor's
57/100 result with 536 findings remains diagnostic rather than an architecture
specification or release verdict.

## 5. Architectural principles

The governance model deliberately follows the API modernization precedent:
[ADR 0001 for modular-monolith boundaries](../../../apps/server/docs/architecture/decisions/0001-modular-monolith-boundaries.md),
the server's [short engineering standard](../../../apps/server/docs/architecture/standards.md),
and its [incremental delivery roadmap](../api-modernization/03-delivery-testing-and-documentation-roadmap.md)
use narrow ownership, caller-shaped capabilities, exact architecture-debt
ratchets, complete behavior slices, and evidence-based acceptance. Projects
adopts those governing principles while defining frontend-specific layers,
server/browser graphs, query ownership, hydration, bundles, and realtime cache
semantics. It does not copy Go transport/service/repository layers into React.

Every new or migrated Projects change follows these rules:

1. **Preserve user behavior while changing ownership.** Structural work does not
   silently change routes, permissions, API payloads, interactions, or cache
   outcomes.
2. **Make runtime boundaries visible.** A file's path and entrypoint make its
   server-only, browser-safe, shared, or route-specific role clear.
3. **Keep product behavior with its owner.** UI, domain vocabulary, request
   mapping, query identity, cache policy, and realtime semantics belong to the
   capability that changes for the same product reason.
4. **Use one source of truth for server state.** SSR prefetch, client reads,
   mutations, Maya tool invalidation, and realtime events use the same owner.
5. **Keep cross-module contracts narrow.** Expose the capability a consumer
   needs, not another feature's internal folder tree.
6. **Prefer direct imports.** Use direct relative imports within an owner and
   explicit, curated boundary files across owners; avoid broad barrels that
   widen contracts or bundles.
7. **Authenticate server mutations like API endpoints.** Validate input,
   establish current actor/workspace authority, and fail closed on missing or
   stale access.
8. **Minimize client work.** Keep client boundaries low, serialize only required
   data, parallelize independent server work, and load heavy optional UI only
   when needed.
9. **Model cache shape deliberately.** Never assume every cache entry is a flat
   array; update only shapes the owner understands and otherwise invalidate.
10. **Make async lifecycles explicit.** Every subscription, listener, timer, and
    request has an owner, cancellation path, and error/recovery behavior.
11. **Promote stable semantics, not repeated syntax.** `shared` is not a dumping
    ground; a reusable primitive needs multiple real consumers and a stable
    feature-neutral contract.
12. **Migrate complete slices.** Characterize, move, switch callers, verify, and
    delete the old path. Do not build a parallel architecture indefinitely.
13. **Use evidence over labels.** A passing type check, focused test, build,
    browser trace, or bundle measurement proves only what it actually exercised.

## 6. Target repository layout

The default substantial capability shape is:

```text
apps/projects/src/
├── app/
│   ├── [workspaceSlug]/
│   └── api/
├── shell/
│   ├── workspace/
│   ├── providers/
│   ├── hydration/
│   ├── navigation/
│   └── realtime/
├── modules/
│   └── stories/
│       ├── public/
│       │   ├── client.ts
│       │   ├── server.ts
│       │   ├── types.ts
│       │   └── capabilities.ts
│       ├── api/
│       │   ├── browser/
│       │   └── server/
│       ├── model/
│       ├── queries/
│       ├── mutations/
│       ├── realtime/
│       ├── ui/
│       └── testing/
└── shared/
    ├── api/
    │   ├── browser/
    │   └── server/
    ├── auth/
    ├── query/
    ├── realtime/
    ├── validation/
    └── testing/
```

This is a responsibility map, not a directory quota. A static catalog may need
only model and UI. A route-only adapter may own no query. A capability with no
realtime behavior does not create an empty `realtime` directory. Existing files
that already satisfy the target can remain where they are.

### 6.1 Canonical ownership questions

| Question                                                  | Canonical target                                                          |
| --------------------------------------------------------- | ------------------------------------------------------------------------- |
| Where is a URL/layout/route handler defined?              | `app`                                                                     |
| Where are workspace-wide providers and overlays composed? | `shell`                                                                   |
| Where is a feature's public client entrypoint?            | `modules/<owner>/public/client.ts` or an explicitly public direct UI path |
| Where is a server-only feature capability exposed?        | `modules/<owner>/public/server.ts`                                        |
| Where are stable cross-feature types exposed?             | `modules/<owner>/public/types.ts`                                         |
| Where is a browser request implemented?                   | `modules/<owner>/api/browser`                                             |
| Where is an authenticated server request implemented?     | `modules/<owner>/api/server`                                              |
| Who owns a query key and options factory?                 | The resource-owning module's `queries`                                    |
| Who owns optimistic and authoritative reconciliation?     | The same module's `mutations`/cache policy                                |
| Who validates and applies a realtime event?               | The resource-owning module's `realtime`                                   |
| Who opens/closes the global realtime connection?          | `shell/realtime`                                                          |
| Where does a feature-neutral primitive live?              | `shared` or an established workspace package                              |

## 7. Dependency rules

### 7.1 Allowed direction

| Source                 | May depend on                                                              | Must not depend on                                                             |
| ---------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `app`                  | shell entrypoints, module public boundaries, shared foundations            | private module internals, feature cache implementation embedded in route files |
| `shell`                | module public client/realtime capabilities, shared foundations             | route-segment internals, private module implementation                         |
| module public boundary | its own internals, shared foundations, explicit capability types           | broad re-export of the module tree, server code from a client entrypoint       |
| module internals       | same module, shared foundations, another module's approved public boundary | `app`, `shell`, another module's private paths                                 |
| `shared`               | standard/runtime libraries and workspace packages                          | `app`, `shell`, or any product module                                          |
| server adapter         | server-safe shared primitives, backend client, request context             | browser hooks, DOM APIs, client-only dependencies                              |
| browser adapter        | browser-safe API client and transport types                                | cookies, secrets, session tokens, Node APIs, `server-only` modules             |

The production module graph must be acyclic. An intentional workflow crosses a
boundary through one directed capability; it does not create reciprocal feature
imports.

### 7.2 Curated facades and direct imports

A public boundary is a contract, not a shortcut around import review.

- Within an owner, import the leaf file directly with a relative path.
- Across owners, import only a named public server, client, type, or capability
  entrypoint.
- Keep client and server exports separate so a client import cannot traverse a
  server graph.
- Do not re-export third-party packages or every component from a module root.
- Keep heavy optional components on explicit paths and use `next/dynamic` when
  deferring them materially improves the bundle or interaction path.
- Type-only contracts use `import type` and must not acquire runtime side effects.

### 7.3 Compatibility adapters

During migration, an old path may delegate to the new owner. The adapter must:

- preserve its exact call contract;
- contain no new business logic;
- name its owning migration slice;
- have focused compatibility evidence; and
- be deleted when the last caller moves.

A compatibility adapter is not a second permanent public API.

## 8. Application shell

The shell owns concerns that are global in lifecycle but not global in business
meaning:

- workspace selection and shell access outcome;
- React Query client and hydration composition;
- application providers and theme;
- navigation and global overlays;
- global chat/Maya surfaces when their lifecycle spans routes;
- analytics identity wiring; and
- realtime connection lifecycle and dispatch.

`src/app/[workspaceSlug]/layout.tsx` is a reference migration seam: today it
authenticates, reads workspaces, redirects, starts non-critical prefetches,
awaits critical prefetches, hydrates the cache, and renders several global
providers/overlays. The target route layout remains a Next.js adapter while
shell-owned composition moves behind named boundaries. Independent server work
starts early and is awaited together; moving code must not introduce a serial
waterfall.

The shell does not own story, calendar, notification, or Maya cache semantics.
It composes the capabilities exported by those owners.

## 9. Server and browser API boundaries

### 9.1 Server request context

Provide one server-only, request-scoped context for data that is safe and useful
across server reads:

```ts
type ServerRequestContext = {
  session: AuthenticatedSession | null;
  workspaceSlug?: string;
  cookieHeader: string;
  requestId?: string;
};
```

The exact type may evolve. Construction uses a server-only module and
request-scoped memoization where appropriate so repeated reads of auth/cookies
within one render do not repeat work. It must not become a service locator or a
way to bypass operation-level authorization.

### 9.2 Server actions and route handlers

Treat every externally reachable action/handler as an API boundary:

1. authenticate the current request;
2. validate the transport input with a bounded typed schema;
3. establish workspace and resource authority;
4. call one module-owned use case/request adapter;
5. return a safe typed result or mapped error; and
6. expose no token, cookie, internal backend response, or sensitive log field.

Client-side hiding and route layout checks are not authorization evidence.

### 9.3 Browser requests

Browser adapters receive only the public input needed for the call. They rely on
the approved browser credential transport and never accept a raw server token.
They normalize backend errors and DTOs at the owner boundary. Components do not
construct arbitrary URLs, authorization headers, or response-shape casts.

### 9.4 Outbound requests

Caller-controlled server fetches accept only required schemes, validate every
redirect destination, reject private and metadata networks for IPv4 and IPv6,
and bound time, bytes, and content types. Application validation complements a
network egress policy; it cannot by itself eliminate DNS-rebinding risk.

### 9.5 Shared transport types

Generated or backend-client types can remain transport inputs, but modules map
them into deliberate UI/domain values when nullability, dates, enums, identity,
or compatibility semantics matter. Do not make a generated response type the
universal component model merely to avoid a small mapper.

## 10. Query and cache ownership

Each server-state resource has one owner that exports:

- a hierarchical key factory;
- normalization for every key input;
- a query-options factory usable by server prefetch and browser hooks;
- finite result types for detail/list/grouped/infinite variants;
- mutation reconciliation and rollback helpers;
- invalidation scope for related resources; and
- realtime event reducers or invalidation policy when applicable.

For example, a route and hook consume the same options rather than reconstructing
the contract:

```ts
const options = storyQueries.detail({ workspaceSlug, storyId });
await queryClient.prefetchQuery(options);
const result = useQuery(options);
```

The example is illustrative. Server and client query functions may need
different request adapters, but they retain the same normalized identity and
result contract.

Required rules:

- include workspace and every result-changing filter in the key;
- use stable normalized objects or primitives;
- distinguish finite list, grouped, and infinite result shapes;
- never create a literal key outside the owner for a migrated resource;
- never broaden invalidation merely because the exact owner is unknown;
- reuse the same cache transformation for optimistic mutation and equivalent
  realtime events where their semantics match; and
- treat backend response data as authoritative after settlement.

The current global `src/constants/keys.ts` is migration input, not a file to
delete in one sweep. Move one resource only after all prefetch, read, mutation,
Maya, and realtime callers are inventoried.

## 11. Typed realtime cache ownership

The shell owns one connection per authenticated workspace lifecycle. It parses an
envelope, validates the event against a bounded discriminated union, checks the
workspace, and dispatches to registered owner capabilities.

A module realtime capability defines:

- event names and payload schemas;
- version/ordering or updated-at semantics when available;
- idempotency behavior for duplicate delivery;
- supported cache shapes and exact patch rules;
- related-resource invalidation; and
- safe behavior for unknown, stale, malformed, or unsupported data.

An event must never be applied through `JSON.parse(...) as EventType` alone.
Malformed or mismatched events are reported without sensitive payloads. When the
owner cannot prove a shape-safe patch, it performs the narrowest authoritative
invalidation.

Tests cover:

- valid event to detail and list caches;
- grouped and infinite caches;
- absent cache entries;
- duplicate and out-of-order events;
- stale mutation/SSE races;
- wrong workspace or resource;
- unknown version/type; and
- unmount, reconnect, and handler cleanup.

## 12. Component and rendering boundaries

- Route files compose; feature components express feature behavior; shared visual
  primitives remain feature-neutral.
- Add `"use client"` only at the lowest boundary requiring browser state,
  effects, context, or event handlers.
- Keep serialized server-to-client props minimal and avoid passing duplicate
  representations of the same data.
- Start independent server fetches in parallel. Await late when a branch may not
  need the result.
- Avoid effects for state that can be derived during render or handled directly
  in an interaction.
- Deduplicate global listeners in shell-owned providers and clean them up
  deterministically.
- Load editor, chart, AI, or other heavy optional surfaces conditionally when
  measurement shows meaningful bundle or interaction benefit.
- Keep established loading, empty, error, optimistic, disabled, and permission
  states during structural moves.

React Doctor may highlight candidates, but a refactor is justified by source and
behavior evidence, not by the score alone.

## 13. Module archetypes

| Archetype             | Typical ownership                                                          | Examples                                              | Required evidence                                                |
| --------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------- |
| Resource module       | model, server/browser data, queries, mutations, UI, optional realtime      | stories, objectives, sprints, teams                   | request, cache, permission, and UI interaction tests             |
| Orchestration module  | use case and UI over several capability ports; little/no owned persistence | roadmap, my-work, summary                             | directed capability imports and orchestration tests              |
| Integration module    | install/settings/provider UI and request adapters                          | calendar, integration requests, Slack/GitHub settings | credential-safe boundary, revoke/retry, and provider-state tests |
| Shell capability      | global lifecycle and composition                                           | workspace shell, realtime connection, providers       | lifecycle, cleanup, hydration, and access tests                  |
| Static/catalog module | pure values or deterministic rendering                                     | terminology/catalog-like data                         | pure tests and explicit owner                                    |
| Public surface        | unauthenticated or optional-auth route/UI with its own identity rules      | public portal, feedback widget                        | origin/identity/privacy/authorization contract tests             |

Do not force every archetype into the same folders. Uniform ownership matters;
empty architecture does not.

## 14. Architecture enforcement

Phase 0 now includes a read-only mechanical gate that parses production
TypeScript/TSX static imports, re-exports, and literal dynamic imports. Its
initial exact-path baseline ratchets:

- imports from current lower-layer roots (`components`, `constants`, `context`,
  `hooks`, `lib`, `types`, and `utils`) into feature modules;
- all exact cross-module imports and the membership of module dependency cycles;
- imports from designated broad root barrels;
- files under the current legacy `src/lib/actions`, `src/lib/hooks`, and
  `src/lib/queries` roots; and
- handwritten production files above 700 lines.

The separate 244-call `await auth()` analysis remains useful for request-boundary
planning, but the gate deliberately does not classify required authentication as
architecture debt.

This is a mechanical subset of the target enforcement, not the complete gate.
Semantic public-versus-private module entrypoints, server-only leakage reachable
from client graphs, unmanaged query-key literals, duplicate resource owners, and
shell or route files implementing module cache semantics still require dedicated
checks.

The initial baseline records existing findings by exact path/edge. It is a
ratchet:

- a new file or edge fails;
- adding a member or edge to an existing cycle fails;
- reductions pass;
- refreshing the snapshot cannot accept growth; and
- changing a rule requires a reviewed ADR.

File size is a review signal, not the architecture itself. For handwritten
production files, target fewer than 400 lines, review 400–700 for cohesion, and
require an explicit exception or split above 700 before substantial behavior is
added. Generated files and data-only fixtures are exempt. Split by use case or
component responsibility, never `part2` naming.

## 15. Testing standard

A persisted authenticated behavior slice normally needs:

- pure tests for normalization, validation, policy, and cache reducers;
- request-adapter tests for auth context, parameters, response mapping, and safe
  errors;
- React Query tests for key identity, optimistic update, rollback, concurrent
  settlement, invalidation, and all owned cache shapes;
- component tests for loading, empty, error, permission, interaction, and
  accessibility behavior;
- server action/route tests for missing, stale, wrong-workspace, and allowed
  authority; and
- a focused browser journey when RSC hydration, navigation, realtime, editor,
  drag-and-drop, or provider behavior crosses runtime boundaries.

Tests must not use arbitrary sleeps, share mutable cache/session state between
parallel cases, or replace an authorization test with a hidden button assertion.

Jest startup and discovery repair are part of Phase 0 and are now implemented.
The complete checkpoint passes 220/220 suites and 1,279/1,279 tests. Every
subsequent modernization checkpoint must repeat the complete run and distinguish
that Jest evidence from type, build, browser, authorization, and deployment
evidence.

## 16. Migration rule

The unit of migration is one complete behavior, not one directory:

1. inventory its route, UI, server/browser request, query keys, cache shapes,
   mutations, realtime/Maya callers, permissions, and tests;
2. add characterization evidence before moving risky behavior;
3. establish the target owner and explicit public boundaries;
4. move server and browser data paths without changing the external contract;
5. move query/options, mutations, and realtime behavior to the same owner;
6. switch all callers through temporary adapters where necessary;
7. run the risk-proportionate gates; and
8. delete old paths and reduce the architecture baseline.

Do not create a new `shared`, `shell`, or module tree and then leave existing
behavior duplicated indefinitely.

## 17. Definition of done

A capability is architecture-complete only when every applicable statement is
true:

- its owner and archetype are documented;
- routes/layouts contain only route-level composition;
- client and server public graphs are explicit and do not leak;
- server actions/handlers authenticate, validate, and authorize current scope;
- cross-module callers use approved narrow boundaries;
- it owns one query-key/options contract per resource;
- SSR, browser hooks, mutations, Maya, and realtime use that contract;
- realtime input is validated and cache updates are shape-aware;
- old global/lib/private paths are removed after caller cutover;
- focused unit, request, cache, component, and browser evidence passes as
  applicable;
- type check, lint, production build, and restored Jest gates pass for the
  affected scope and required broad checkpoint;
- bundle/rendering impact is measured when a boundary or heavy dependency moves;
- documentation and architecture inventory are current; and
- no new cycle, forbidden edge, raw key, or unexplained exception is added.

## 18. Target scorecard

The modernization reaches its target only when:

| Dimension            | Target evidence                                                                                                            |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Architecture         | Production module graph is acyclic and follows `app -> shell/modules -> shared`.                                           |
| Correctness          | Query/cache/realtime contracts have deterministic success, failure, concurrency, and shape evidence.                       |
| Security             | Every server boundary has explicit actor/workspace/resource policy; browser graphs contain no secrets or server-only code. |
| Performance          | Critical routes have waterfall, serialization, bundle, render, and interaction evidence appropriate to their risk.         |
| Testing              | Jest is deterministic, risk-layered tests run in CI, and critical browser journeys have focused coverage.                  |
| Maintainability      | A route-to-request-to-query/cache-to-test trace is predictable from ownership and generated inventory.                     |
| Developer experience | One documented workflow and actionable gates let an engineer change a small capability without tribal knowledge.           |

No score is awarded merely for moving files. The outcome is safer reasoning,
clearer ownership, faster feedback, and preserved product behavior.
