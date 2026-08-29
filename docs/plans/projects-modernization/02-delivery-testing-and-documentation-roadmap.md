# FortyOne Projects Modernization: Delivery, Testing, and Documentation Roadmap

**Status:** Active delivery plan; Phase 0 mechanical foundation implemented and
first invitations slice in verification  
**Snapshot date:** 2026-08-28  
**Scope:** Incremental modernization of `apps/projects`  
**Architecture:** [current state and target architecture](./01-current-state-and-target-architecture.md)  
**Evidence ledger:** [implementation status](./00-implementation-status.md)  
**Review standard:**
[`apps/projects/docs/architecture/standards.md`](../../../apps/projects/docs/architecture/standards.md)

This roadmap turns the target architecture into small, verifiable delivery
steps. Source migration has started with an invitations ownership slice: known
callers now use module-owned invitation actions, reads, and keys, and four
duplicate `src/lib` implementations are deleted. The Phase 0 scanner, baseline,
and Jest discovery repair are also implemented. This is not a claim that the
invitations capability has passed every acceptance gate, that the wider source
tree follows the target, or that the modernization is finished.

## 1. Starting point

The initial reviewed baseline was large enough that a folder-first rewrite would
hide risk rather than remove it:

- 1,622 TypeScript/TSX files and 184,890 physical lines under
  `apps/projects/src`, approximately 185,000 lines;
- 26 top-level directories under `src/modules`;
- one static strongly connected component containing 16 feature modules;
- about 250 authenticated request paths that call `auth()`;
- 218 Jest test files, while the default Jest command failed before test
  discovery at that initial snapshot;
- a passing Projects type check at the reviewed snapshot; and
- a React Doctor result of 57/100, used only as a diagnostic searchlight.

After the Phase 0 implementation and invitations slice, the current worktree has
1,624 TypeScript/TSX files, 184,854 physical lines, and 220 discoverable Jest
files. The architecture scanner covers 1,404 production files and still reports
one 16-module strongly connected component. These current counts are delivery
context; they do not change the initial architecture-readiness score by
themselves.

The current **4.7/10** score measures source and architecture readiness for safe
change. It is not a rating of product quality, feature quality, customer value,
or visual design.

| Dimension            |  Score |
| -------------------- | -----: |
| Architecture         | 4.0/10 |
| Correctness          | 6.0/10 |
| Security             | 5.5/10 |
| Performance          | 5.0/10 |
| Testing              | 4.0/10 |
| Maintainability      | 4.0/10 |
| Developer experience | 4.5/10 |

## 2. Delivery rules

Every phase and behavior slice follows these rules:

1. **No big-bang migration.** Keep one releasable application and move one
   complete behavior slice at a time.
2. **Security precedes relocation.** Record actor, workspace, role, validation,
   and failure behavior before changing an authenticated request path.
3. **Behavior is the compatibility contract.** Preserve URLs, permissions, API
   payloads, loading/error/empty states, optimistic behavior, realtime outcomes,
   analytics, and accessibility unless a separately approved product change
   says otherwise.
4. **The owner is singular.** One capability owns each resource's request
   mapping, query keys/options, cache policy, mutation reconciliation, and
   realtime semantics.
5. **Move complete slices.** Inventory, characterize, introduce the target
   owner, switch every caller, verify, and delete the superseded path.
6. **Debt ratchets downward.** Existing exact violations may be baselined; new
   cyclic edges, private cross-module imports, server leakage, or duplicate
   query owners fail the change.
7. **Evidence is scoped.** A passing type check is not runtime proof; a focused
   test is not a full-suite result; an unauthenticated browser check is not an
   authorization check; React Doctor is not a release gate.
8. **Documentation changes with ownership.** A moved source contract and its
   ownership, tests, compatibility state, and removal status land together.

## 3. Phase sequence

The phases are ordered by risk and dependency. Work may proceed in parallel only
when slices have separate owners and do not create a second implementation of
the same contract.

### Phase 0: Make the baseline trustworthy

**Objective:** Establish reproducible feedback and prevent architecture debt
from growing before moving source.

**Current delivery state:** ADR 0001 is accepted. Jest configuration startup and
discovery are repaired. The dependency-free scanner, eleven scanner tests, guarded
baseline writer, and initial exact debt baseline are implemented. Invitations
has been selected and cut over as the first ownership slice. The complete Jest
run passes 220/220 suites and 1,279/1,279 tests; the combined type check and
production build pass; and focused changed-scope ESLint passes. The broad lint
baseline remains red. A Projects pull-request workflow now configures the
architecture, type, Jest, and build gates, although it has not run remotely on
this worktree. The full Phase 0 inventory, semantic boundary checks, outbound
request hardening, and complete invitations characterization remain open.

**Work**

- Keep the accepted target architecture, ADR 0001, and engineering standards
  current as implementation evidence evolves.
- Preserve the repaired Jest startup path without weakening strict TypeScript
  or silently excluding existing tests. The initial reviewed error was
  `TypeError: Cannot read properties of undefined (reading 'fileExists')` before
  discovery.
- Make the source inventory reproducible: routes, module import edges and
  cycles, public/private entrypoints, server/browser graphs, authenticated
  boundaries, query owners, literal query keys, realtime event owners, and test
  assets.
- Extend the active mechanical architecture check toward full
  `app -> shell/modules -> shared` enforcement, semantic private cross-module
  detection, and server-only leakage analysis. The current subset ratchets
  lower-layer-to-module imports, exact cross-module imports, cycle membership,
  broad barrels, legacy files, and oversized production files. The separate
  `await auth()` inventory remains analysis rather than ratcheted debt.
- Record existing violations by exact source edge rather than exempting a whole
  directory or rule.
- Complete characterization of the selected invitations pilot, which has a real
  request/query/cache path and no product redesign requirement.
- Close the `/api/metadata` SSRF boundary with per-hop destination checks,
  bounded content fetching, network egress protection, and adversarial tests.

**Acceptance**

- `pnpm --filter projects type-check` passes on the combined worktree.
- `pnpm --filter projects test -- --runInBand` reaches deterministic discovery
  and passes the intended 220 suites and 1,279 tests.
- The active architecture inventory commands are deterministic and reviewable in
  CI, and each planned semantic check is either implemented or explicitly
  tracked.
- Caller-controlled outbound requests cannot reach loopback, private,
  link-local, or cloud-metadata destinations through direct URLs, DNS answers,
  or redirects, and their time, size, and content type are bounded.
- The debt baseline rejects growth in every currently represented category,
  including expansion of the 16-module strongly connected component.
- The pilot's route, permission, request, cache, and UI behavior is documented,
  and any characterization gaps are closed before the capability is called
  architecture-complete.

Source migration did not begin merely because the governing documents existed;
it began after the initial mechanical gate, Jest discovery repair, and a scoped
pilot were available. Phase 0 remains partially open until the combined
acceptance evidence above is recorded.

### Phase 1: Establish foundations through one pilot slice

**Objective:** Prove the target with a complete behavior slice rather than
creating unused framework code.

**Current delivery state:** The invitations source cutover is implemented. Its
keys and duplicate request paths have one module owner, known callers are
switched, the obsolete files are deleted, and focused key/mapping tests pass.
Do not call the capability architecture-complete until the remaining request,
permission, cache, browser, and broad-gate evidence passes.

**Work**

- Introduce only the shared primitives required by the pilot:
  - a server-only request-context boundary that deduplicates request-local
    session/cookie work without hiding authorization;
  - a browser-safe request base with no token, cookie, secret, Node API, or
    server-only dependency;
  - typed query-key and query-options conventions;
  - runtime realtime-envelope validation and a module-handler registry if the
    pilot consumes events; and
  - focused test builders that preserve workspace and authorization semantics.
- Place the pilot's model, server/browser adapters, query options, mutation/cache
  policy, realtime reducer, and UI under one owning module.
- Expose only the public browser, server, types, or capability entrypoints that
  real consumers need.
- Use direct imports inside the owner and curated imports across owners. Avoid a
  broad root barrel.
- Switch SSR prefetch, hydration, client hooks, mutations, Maya paths, and
  realtime handling that address the pilot resource to the same query contract.
- Delete the superseded request/query/cache implementation once all callers
  switch.

**Acceptance**

- The pilot preserves route, permission, API, loading, error, empty, optimistic,
  realtime, and accessibility behavior.
- Server-only code cannot enter a client bundle; browser adapters contain no
  server credential logic.
- Every pilot resource key and query-options factory has one owner and is reused
  by server prefetch and browser consumers.
- Realtime payloads are runtime validated. Reducer tests cover the cache shapes
  the pilot actually supports and bounded invalidation handles unknown shapes.
- No new module-cycle member, forbidden layer edge, unmanaged literal query key,
  or private cross-module import is introduced.
- Targeted tests, complete Jest, type check, lint, production build, and the
  pilot's critical browser journey pass. If a repository-wide baseline outside
  the slice remains red, the exact pre-existing failure is reported rather than
  presented as a pass.
- All pilot callers use the new owner and its obsolete implementation is
  removed. A compatibility adapter may remain only with an owner and deletion
  condition.

### Phase 2: Separate the application shell from route adapters

**Objective:** Make routing thin and give workspace-wide runtime composition a
single explicit owner.

**Work**

- Extract provider composition, workspace shell state, hydration boundaries,
  global overlays, navigation, and realtime connection lifecycle from App Router
  files into `src/shell`.
- Keep `src/app` responsible for Next.js routes, layouts, metadata, route
  handlers, redirects, and route-level composition.
- Preserve the independent parallel prefetch behavior currently visible in
  `src/app/[workspaceSlug]/layout.tsx`; do not replace it with sequential fetches
  or a client waterfall.
- Move connection lifecycle out of `src/app/server-sent-events.tsx` while
  dispatching validated events to module-owned handlers.
- Keep client boundaries low and serialize only data needed by the client
  graph. Dynamically load genuinely heavy optional UI where measurement supports
  it.

**Acceptance**

- Route files no longer own reusable feature query/cache/realtime policy.
- Modules do not import `app` or `shell`, and `shared` imports neither.
- Authentication, missing-workspace redirects, error paths, hydration, and
  provider behavior remain characterized and green.
- Server request timing does not regress through newly sequential work; compare
  a trace or focused timing evidence where the move changes orchestration.
- One shell connection is opened and closed predictably, while module handlers
  own event meaning and cache changes.

### Phase 3: Break the core work-graph cycle

**Objective:** Reduce the 16-module strongly connected component monotonically
while migrating the capabilities that participate in the central planning and
execution graph.

**Work**

- Generate the exact cycle edges and select the next slice by dependency impact,
  security risk, test evidence, and change frequency rather than directory size.
- Migrate complete resource slices across stories, objectives, key results,
  sprints, teams, roadmap, and related work surfaces in the order supported by
  the current graph.
- Replace cross-module internal imports with one of:
  - a provider-owned public capability;
  - a stable transport-neutral public type;
  - a consumer-owned narrow interface; or
  - a workflow/orchestration owner when behavior genuinely spans modules.
- Consolidate each resource's keys/options, optimistic transitions,
  invalidations, and realtime reducers during its slice; do not postpone cache
  ownership to a later global rewrite.
- Keep public entrypoints narrow enough that moving one feature does not pull its
  full UI or server graph into another bundle.

**Acceptance**

- The largest strongly connected component shrinks or, for a neutral enabling
  slice, its exact membership and edge count do not grow. The program exit gate
  is an acyclic production module graph.
- Each migrated resource has one server-state contract shared by prefetch,
  client reads, mutations, Maya, and realtime paths that use it.
- Cross-feature workflows are expressed through named capabilities rather than
  another module's private hooks, components, actions, or query constants.
- Characterization, request-policy, query/cache, realtime, component, and
  browser tests pass in proportion to the slice's risk.
- Superseded files and compatibility imports are deleted as their final callers
  move.

### Phase 4: Complete secondary capabilities and integrations

**Objective:** Apply the proven structure to remaining collaboration, discovery,
configuration, integration, and assistant surfaces.

Candidate areas include notifications, search, team feedback, strategy,
documents, settings, calendar, integration requests, Maya, and public-portal
flows. Their exact order must come from the current dependency and risk
inventory; this list does not pre-assign ownership or imply they are internally
homogeneous.

**Acceptance**

- Each capability has an explicit owner and only the layers it actually needs.
- Settings and integration mutations preserve role/workspace enforcement and
  fail closed in negative tests.
- Search, notification, calendar, and assistant-triggered cache behavior consume
  resource-owned keys/options rather than constructing parallel identities.
- Public or unauthenticated routes do not acquire authenticated shell or private
  module dependencies accidentally.
- Heavy optional feature entrypoints are isolated based on bundle evidence, not
  intuition alone.

### Phase 5: Remove compatibility debt and make governance routine

**Objective:** Finish the migration rather than preserve two architectures.

**Work**

- Remove temporary adapters, legacy barrels, duplicate action/query folders,
  stale key constants, and broad architecture exceptions after their final
  consumer moves.
- Drive the production module graph to zero cycles and enable the target rules
  without a legacy baseline.
- Make ownership and graph inventories part of ordinary pull-request feedback.
- Establish bundle, request-waterfall, accessibility, and critical-journey
  regression budgets from measured baselines.
- Archive the migration ledger as a dated outcome and keep the engineering
  standard and ADRs as ongoing governance.

**Acceptance**

- The production feature graph is acyclic.
- No module consumes another module's private internals.
- No server-only dependency is reachable from a client graph.
- Each server-state resource has one query/options/cache owner; unmanaged legacy
  key literals and duplicate ownership are eliminated.
- Realtime events are runtime validated and dispatched to typed, module-owned
  cache policies.
- Compatibility paths and temporary architecture exceptions are zero or have a
  separately accepted ADR with a durable rationale.
- Type check, lint, Jest, production build, and defined critical browser journeys
  are green on the same revision.

## 4. Standard behavior-slice workflow

Every source migration uses the same reviewable loop:

1. **Inventory the behavior.** List routes, server components/actions/handlers,
   browser hooks, request calls, authentication and workspace checks, query
   keys, cache shapes, mutations, realtime events, components, analytics, and
   tests.
2. **Characterize high-risk outcomes.** Add or repair tests for access denial,
   stale data, optimistic rollback, error/loading/empty states, event ordering,
   and critical user interaction before moving implementation.
3. **Name the owner and public contract.** Record which module owns the resource
   and which browser, server, type, or capability entrypoints consumers need.
4. **Separate runtime graphs.** Move authenticated server access behind a
   server-only adapter and browser requests behind a browser-safe adapter.
5. **Unify server state.** Define one normalized key factory and options contract;
   reuse it for prefetch, hydration, hooks, mutations, invalidation, Maya, and
   realtime.
6. **Switch callers completely.** Move route, UI, action, hook, and event
   consumers while temporary adapters remain thin and observable.
7. **Delete the old path.** Remove duplicate owners, keys, cache policies, and
   compatibility exports when the final caller moves.
8. **Verify and record evidence.** Run the checks appropriate to the slice,
   update the status ledger, and state any environment or coverage limitation.

A folder move without ownership consolidation, caller migration, behavioral
evidence, and deletion is not a completed slice.

## 5. Testing and verification strategy

### 5.1 Static and build feedback

| Check              | What it can establish                                                                                                 | What it cannot establish                                                                                                                        |
| ------------------ | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Type check         | Type compatibility and server/client declarations represented in types                                                | Runtime validation, authorization, cache outcomes, or browser behavior                                                                          |
| Lint               | Configured code-quality, import, React, and accessibility rules                                                       | Product correctness or a complete dependency graph unless a dedicated rule exists                                                               |
| Architecture check | Current exact lower-layer/cross-module imports, cycle membership, broad-barrel use, legacy files, and oversized files | Semantic public/private boundaries, server leakage, query ownership, auth policy, runtime payload validity, or permission correctness by itself |
| Production build   | Next.js compilation, route/build integration, and bundle-graph failures visible to the builder                        | Authenticated workflows, data correctness, or acceptable performance by itself                                                                  |
| Bundle inspection  | Whether a change pulls an unexpected feature or heavy dependency into an entrypoint                                   | Runtime responsiveness or network latency by itself                                                                                             |

The static architecture check should report exact importing and imported paths,
the violated rule, current baseline status, and the owning module. A broad
directory exemption is not an acceptable debt record.

### 5.2 Jest layers

With discovery restored, classify and strengthen the current 220 test files
around these risks:

- pure model, schema, and permission policy;
- authenticated server adapter and action behavior, including negative
  workspace/role cases;
- query-key normalization and options contracts;
- optimistic cache transition, rollback, invalidation, and authoritative
  reconciliation;
- realtime decoder and reducer behavior for malformed, duplicate, stale,
  out-of-order, workspace-mismatched, flat, detail, grouped, paginated, and
  infinite cache cases that a resource supports;
- component loading, error, empty, disabled, retry, and interaction states; and
- shell provider, hydration, subscription cleanup, and route adapter behavior.

Not every resource supports every cache shape. Tests document the owner's real
contract; unknown shapes use bounded invalidation instead of speculative
mutation.

### 5.3 Browser and request evidence

Use focused browser tests for the critical journeys touched by a slice. At
minimum, verify the relevant combination of:

- unauthenticated redirect or denial;
- wrong-workspace and insufficient-role denial;
- server-rendered first load and hydrated interaction;
- create/update/delete success and server rejection;
- optimistic rollback and subsequent reconciliation;
- reconnect, duplicate event, stale event, and cross-workspace event behavior;
- navigation without leaked state or duplicate subscriptions; and
- keyboard, focus, labeling, loading, error, and empty-state behavior.

Authenticated browser evidence must use a real authenticated test context. A
static render, mocked session, or public-route smoke test must be described by
that narrower scope.

### 5.4 React Doctor

React Doctor's reviewed 57/100 result is a diagnostic, not truth. Use individual
findings to locate possible oversized components, effect misuse, rendering
instability, or boundary problems. Accept or reject each material finding from
source evidence and focused verification. Do not fail or approve a release from
the aggregate score. The final changed-scope run reports 90/100 with no issues
across 29 files; that narrower result does not supersede the full-app baseline.

## 6. Proposed feedback pipeline

Phase 0 has established the architecture commands, restored Jest discovery, and
configured architecture, type, complete Jest, and build checks in
`.github/workflows/projects-quality.yml`. That workflow has not executed
remotely on this worktree. The pipeline becomes a complete required gate only
after broad lint, browser, security, and slice evidence are also recorded on the
same revision.

| Cadence       | Required evidence                                                                                                                            |
| ------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Local/focused | Formatter check for touched files, focused Jest tests, type check for contract changes, and the slice's architecture report                  |
| Pull request  | Type check, lint, architecture ratchet, complete Jest suite, production build, changed-slice browser journey, and documentation/status check |
| Scheduled     | Full dependency/query/auth/realtime inventory, cycle trend, bundle trend for key routes, and critical authenticated journey suite            |
| Advisory      | React Doctor findings correlated with changed source; never the sole approval or rejection signal                                            |

Run independent checks in parallel where CI capacity permits, but report each
result separately. Never summarize a green focused test and a red repository
baseline as “tests pass.”

## 7. Program measures

Measurements are guardrails and trend evidence, not productivity targets.

| Measure                         |                                                                                               Current checkpoint | Direction / exit condition                                                         |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------: | ---------------------------------------------------------------------------------- |
| Largest feature SCC             |                                                                                                       16 modules | Never grow; reduce to zero production cycles                                       |
| Source size                     |                                                                       1,624 TS/TSX files; 184,854 physical lines | Context only; no target to delete code for its own sake                            |
| Authenticated request inventory |                                                                        244 `await auth()` calls across 225 files | Preserve explicit enforcement; make ownership and request-context policy auditable |
| Test assets                     |                                                                        220/220 suites; 1,279/1,279 tests passing | Track passing behavior and risk coverage, not counts alone                         |
| Type check                      |                                                                                     Passing on combined worktree | Keep required on every migration revision                                          |
| Production build                |                                                           Pass: compile 39.1s; TypeScript 11.5s; 27 static pages | Preserve on every releasable checkpoint                                            |
| Lint                            |                              Focused changed scope passes; final broad run 240 findings: 160 errors, 80 warnings | Repair the repository-wide baseline                                                |
| React Doctor                    |                                                                                                           57/100 | Advisory only; no target score without source-correlated value                     |
| Module import debt              |                                                                                                One 16-module SCC | Exact edges ratchet down; final graph acyclic                                      |
| Architecture ratchet            | 370/372 lower-layer; 298/298 cross-module; 1/16 cycles; 1,075/1,095 barrels; 170/170 legacy; 25/29,019 oversized | No represented category grows; add the planned semantic checks                     |
| Query ownership                 |                                                                Split across constants, `lib`, hooks, and modules | One key/options/cache owner per migrated resource, then every resource             |
| Realtime typing                 |                                                                      Parsed JSON asserted into local event types | Runtime-validated envelope and module-owned typed reducers for every handled event |

Re-run inventories from a pinned revision and record the command, exclusions,
and generated timestamp. Counts from different definitions must not be charted
as a trend.

## 8. Documentation contract

The following records evolve with implementation:

- `00-implementation-status.md` records current phase, completed slices, exact
  verification, known red baselines, and limitations.
- `01-current-state-and-target-architecture.md` describes the governing target;
  change it only when the architecture changes, not for normal progress.
- This roadmap records phase and acceptance policy.
- `apps/projects/docs/architecture/standards.md` is the concise code-review
  contract.
- `apps/projects/docs/architecture/decisions` records durable choices and their
  consequences.
- Each migrated capability documents its owner, public entrypoints, request and
  permission boundary, server-state contract, realtime behavior, and any
  temporary compatibility adapter.

An ADR is required to reverse dependency direction, introduce a new shared
state/runtime pattern, split deployment boundaries, weaken server/browser
separation, or accept durable architecture debt. Routine refactors that follow
the accepted target update the implementation ledger, not the ADR.

## 9. Risk register

| Risk                                                         | Control                                                                                                  | Evidence                                                           |
| ------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| Hidden authorization regression during request consolidation | Inventory actor/workspace/role behavior first; keep actions and handlers fail closed; add negative tests | Server adapter/action tests and authenticated browser denial paths |
| Two permanent architectures                                  | Migrate complete slices, switch every caller, require owner/removal condition for adapters               | Import inventory and deletion checklist                            |
| Cache identity drift                                         | One owner for keys/options/mutations/realtime; normalize inputs once                                     | Query-contract and cache-transition tests                          |
| Unsafe realtime mutation                                     | Runtime validation, tenant check, ordering policy, known-shape reducer, invalidation fallback            | Decoder/reducer tests and connection lifecycle tests               |
| New client waterfalls or bundle growth                       | Preserve parallel server work, keep client boundaries low, curate imports, measure risky changes         | Production build, request trace, and bundle comparison             |
| Architecture checker becomes ceremonial                      | Baseline exact paths/edges; reject growth; require ADR for rule change                                   | CI diff with owner and violation details                           |
| The repaired Jest baseline regresses                         | Run and own the complete suite on every migration checkpoint                                             | Complete command output on the same revision                       |
| Heuristic score drives low-value churn                       | Treat React Doctor as advisory and correlate findings with runtime/source risk                           | Focused evidence attached to accepted changes                      |
| Migration expands into product redesign                      | Separate behavior-preserving architecture slices from product changes                                    | Explicit scope and acceptance contract per slice                   |

## 10. Evidence limitations and reporting language

The reviewed evidence supports an architecture plan, not a production-readiness
claim:

- physical LOC and file counts describe size, not complexity or quality;
- the static import graph may miss computed runtime imports and requires type-only
  and compatibility edges to be classified;
- 244 executable `await auth()` calls across 225 production files demonstrate
  distributed authentication work, not that every authorization rule is correct
  or that checks should be removed;
- the 220/220 suite and 1,279/1,279 test pass is meaningful Jest evidence, but it
  does not prove complete product behavior, coverage, authorization, or browser
  correctness;
- the passing type check does not execute permissions, requests, hydration,
  optimistic transitions, realtime events, or browser journeys;
- the production build passes, but it does not execute authenticated journeys or
  prove runtime data correctness;
- focused changed-scope ESLint passes, while the final broad run remains red with
  240 findings: 160 errors and 80 warnings;
- React Doctor's 57/100 is a heuristic result, not a factual quality score; and
- no authenticated browser suite, bundle analysis, accessibility audit, network
  trace, deployment, or production verification is claimed here.

Status updates must use precise language such as “focused tests passed,” “type
check passed,” “Jest discovery completed,” or “220/220 Jest suites passed,” with
the command and revision.
Do not use “migration complete,” “tests pass,” “secure,” or “production ready”
until the corresponding exit evidence exists.

## 11. Program completion gate

The Projects modernization is complete only when all of the following are true
on one releasable revision:

1. production dependencies follow `app -> shell/modules -> shared` and the
   feature-module graph is acyclic;
2. cross-module consumers use curated public boundaries or explicit
   consumer-owned capabilities, with no private internal imports;
3. server-only and browser-safe request graphs are mechanically separated, and
   externally reachable actions/handlers retain explicit authentication,
   authorization, and validation;
4. every server-state resource has one owner for query keys/options, mutations,
   cache reconciliation, and realtime semantics;
5. the shell owns realtime connection lifecycle while modules runtime-validate
   and apply their own events safely;
6. compatibility paths, duplicate owners, legacy broad barrels, and temporary
   debt exceptions are removed or governed by a separate accepted ADR;
7. type check, lint, architecture checks, complete Jest, production build, and
   defined critical authenticated browser journeys pass on the same revision;
8. current architecture, ownership, decisions, operational behavior, and
   limitations are documented; and
9. the final ledger cites the exact evidence and distinguishes verified behavior
   from work that was not exercised.

Until then, report the program by its actual phase and completed behavior slices.
The present state is **Phase 0 mechanical foundation implemented; first
invitations slice in verification; no full source migration**.
