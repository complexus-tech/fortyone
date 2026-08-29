# FortyOne Projects Modernization: Implementation Status

**Status:** Phase 0 mechanical foundation implemented; first invitations slice
in verification  
**Snapshot date:** 2026-08-28  
**Scope:** `apps/projects` source architecture, server/browser boundaries, data
ownership, realtime cache behavior, tests, developer workflow, and internal
documentation  
**Target:** [current state and target architecture](./01-current-state-and-target-architecture.md)  
**Delivery plan:** [delivery, testing, and documentation roadmap](./02-delivery-testing-and-documentation-roadmap.md)

This is an evidence ledger, not a completion announcement. The governing
documents, an initial architecture-debt ratchet, deterministic Jest discovery,
and the first invitations ownership slice now exist. The invitations change
moves duplicate actions and queries into the owning module, gives invitation
query keys one owner, switches the known callers, and deletes four superseded
`src/lib` files. The wider Projects source tree is not architecture-clean or
migrated, the full verification matrix is not yet green, and no commit, push,
deployment, or production-state change is claimed here.

## 1. Executive checkpoint

| Area                        | Current state                                  | Evidence                                                                                                                                                                         | Acceptance still required                                                                                                    |
| --------------------------- | ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Architecture definition     | Foundation documented and ADR accepted         | One target architecture, one delivery roadmap, one review standard, and accepted ADR 0001                                                                                        | Apply the standard to each slice and keep the evidence ledger current                                                        |
| Architecture ratchet        | Initial mechanical subset active               | `architecture:test` passes 11 scanner tests; `architecture:check` scans 1,404 production files and passes against `architecture-debt-baseline.json`                              | Add semantic public/private-boundary, server-leakage, query-owner, and shell/cache checks                                    |
| Pull-request enforcement    | Core local gates configured in CI              | `.github/workflows/projects-quality.yml` runs architecture tests/check, type check, complete Jest, and the production build for relevant changes                                 | Observe the workflow on a pull request; add broad lint only after its existing baseline is repaired                          |
| Source migration            | First scoped ownership slice implemented       | Invitation actions, reads, and keys now live under `src/modules/invitations`; known callers switched; four duplicate `src/lib` files deleted                                     | Finish broad verification and continue one complete behavior slice at a time                                                 |
| Module boundaries           | High coupling remains                          | Static inventory found a strongly connected component containing 16 feature modules                                                                                              | Stop new cross-boundary imports, reduce the component monotonically, then reach an acyclic production graph                  |
| Server/browser API boundary | Present but distributed                        | Separate analysis finds 244 executable `await auth()` calls across 225 production files; `/api/metadata` also fetches a caller-supplied URL without a private-network policy     | One request-context contract, explicit server/browser adapters, SSRF-safe outbound fetching, and per-slice security tests    |
| Query/cache ownership       | Inconsistent overall; invitations consolidated | Invitation keys moved from `src/constants/keys.ts` to the owning module, while other keys and data functions remain split across `src/constants`, `src/lib`, and feature modules | One key/options owner per remaining server-state resource and one cache policy shared by SSR, hooks, mutations, and realtime |
| Realtime behavior           | Functional, centrally coupled                  | `src/app/server-sent-events.tsx` parses unvalidated JSON and directly patches stories while invalidating calendar and notification data                                          | Runtime-validated events, module-owned cache reducers, ordering/idempotency tests, and a shell-owned connection lifecycle    |
| Maya request budget         | Static system prompt reduced                   | `src/app/api/chat/system.ts` is reduced from 12,758 to 9,884 bytes while its focused policy tests and the complete Jest suite pass                                               | Measure real request/token impact; keep approval, authorization, and scheduling policy tests as hard compatibility evidence  |
| Type feedback               | Passing on the combined worktree               | `pnpm --filter projects type-check` completed successfully after the implementation changes                                                                                      | Keep it required on every migration checkpoint                                                                               |
| Production build            | Passing                                        | `pnpm --filter projects build` compiled in 39.1s, type-checked in 11.5s, and generated 27 static pages                                                                           | Add authenticated runtime and browser evidence where the slice requires it                                                   |
| Lint feedback               | Changed scope green; broad baseline red        | Focused ESLint passes for every changed app, Jest, invitations, and scanner file; the final broad run reports 240 findings: 160 errors and 80 warnings                           | Repair the broad baseline without hiding existing findings                                                                   |
| Jest feedback               | Complete suite passing                         | CommonJS `next/jest` configuration reaches deterministic discovery; the complete run passes 220/220 suites and 1,279/1,279 tests                                                 | Keep the complete suite in CI and treat future failures as owned regressions                                                 |
| Test assets                 | Discoverable and executing                     | 220 `*.test.ts`/`*.test.tsx` files exist; the full suite and the two focused invitations tests pass                                                                              | Strengthen request, permission, cache, realtime, and browser risk coverage                                                   |
| Heuristic diagnostics       | Changed scope green; broad baseline critical   | React Doctor reports 90/100 with no issues across 29 changed files; its full-app scan remains 57/100 with 536 findings: 15 errors and 521 warnings                               | Correlate broad findings with source evidence and focused tests; do not use either score as a release gate or truth metric   |

## 2. Quantitative baseline

The initial analysis snapshot contained 1,622 TypeScript/TSX files, 184,890
physical lines, and 218 Jest files. After the mechanical foundation and
invitations slice, the current source worktree contains:

- 1,624 TypeScript and TSX files under `apps/projects/src`;
- 184,854 physical lines in those files, approximately 185,000 lines;
- 26 top-level directories under `src/modules`;
- one static module-import strongly connected component with 16 members;
- 1,404 production TypeScript/TSX files scanned by the architecture check;
- 244 executable `await auth()` calls across 225 production files; and
- 220 Jest test files.

The active ratchet records 370 lower-layer import keys / 372 imports, 298
cross-module keys / 298 imports, one cycle / 16 modules, 1,075 broad-barrel keys
/ 1,095 imports, 170 legacy files, and 25 oversized files / 29,019 combined
lines. The `await auth()` count is a separate analysis inventory, not a
ratcheted debt category.

These counts describe code shape, not product quality, coverage, or security
approval. The complete Jest and production-build passes are meaningful evidence,
but they do not exercise authenticated browser journeys or every architecture
and authorization contract.

## 3. Architecture-readiness score

The score is **4.7/10**. It measures source and architecture readiness for safe,
predictable change. It is not a rating of FortyOne's product, feature set,
visual quality, or customer value.

| Dimension            |  Score | Evidence-based interpretation                                                                                                                                                            |
| -------------------- | -----: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Architecture         | 4.0/10 | Product-oriented modules exist, but ownership is weakened by a 16-module cycle, mixed root-level layers, and cross-feature internals.                                                    |
| Correctness          | 6.0/10 | Strict TypeScript, the complete Jest suite, and the production build pass on the combined worktree; authenticated runtime evidence is still incomplete.                                  |
| Security             | 5.5/10 | Authentication is widespread, but authorization policy is scattered and `/api/metadata` accepts arbitrary outbound URLs without private-network, redirect, and response-budget controls. |
| Performance          | 5.0/10 | The production build passes and the app has solid prefetch, hydration, React Query, and compiler foundations; waterfall, bundle, and interaction evidence remains incomplete.            |
| Testing              | 4.0/10 | Jest startup and discovery are repaired and all 220 suites pass, but architecture, authorization, realtime, and authenticated-browser coverage remain incomplete.                        |
| Maintainability      | 4.0/10 | Naming often follows product concepts, yet duplicate query ownership, broad shared areas, literal keys, and cyclic feature dependencies make changes expensive to reason about.          |
| Developer experience | 4.5/10 | Architecture, type, Jest, build, and focused-lint feedback now work, but the broad lint baseline, full semantic gate set, and boundary navigation still need repair.                     |

The rounded arithmetic mean is 4.7. Future updates must change a score only
with cited evidence and must preserve the distinction between architecture
readiness and product quality.

## 4. What already works in the target's favor

- The application is already organized around recognizable product concepts such
  as stories, objectives, sprints, teams, notifications, strategy, feedback, and
  Maya. The modernization should strengthen those domains, not replace them with
  technical layer-wide folders.
- Next.js 16, React 19, strict TypeScript, React Query, server components, server
  actions, and workspace packages provide appropriate primitives. A framework
  rewrite is not required.
- `src/app/[workspaceSlug]/layout.tsx` already parallelizes critical prefetches
  and hydrates React Query. The issue is contract ownership, not the existence of
  server prefetching.
- Shared workspace packages such as `ui`, `icons`, `api-client`, and `auth` are
  useful boundaries when consumed deliberately.
- Focused tests around mutations, cache behavior, authorization, Maya tools, and
  realtime-adjacent behavior can become characterization evidence for incremental
  moves now that Jest discovery is restored.

## 5. Confirmed architecture debt

### 5.1 Layer and module direction

`src/app`, feature modules, app-local `components`, `context`, `hooks`, and `lib`
can currently depend on one another through several paths. The 16-module cycle is
the clearest aggregate signal: a change in one feature can require understanding
a large mutually dependent graph. The cycle count is a baseline to reduce, not a
license for new edges.

### 5.2 Server and browser requests

Authentication at many call sites is a security strength, but the number and
distribution of those calls makes the full request policy difficult to inspect.
The target does not remove authentication from actions or route handlers. It
introduces a request-scoped, server-only context that deduplicates session and
cookie work while keeping authorization explicit at every externally reachable
mutation and read boundary.

### 5.3 Outbound URL handling

`src/app/api/metadata/route.ts` authenticates the caller, then passes the
caller-supplied `url` to `cheerio.fromURL`. Parsing the URL is not an SSRF
boundary: the route does not reject loopback, link-local, private, or
cloud-metadata destinations; revalidate redirects; or bound response size and
content type. Treat this as the highest-priority security remediation. A safe
implementation must validate every hop, apply network-level egress protection,
bound time and bytes, and test IPv4, IPv6, DNS, and redirect cases.

### 5.4 Query keys and cache policy

Resource keys exist both in `src/constants/keys.ts` and feature-owned constants,
while some consumers still construct array literals. Query functions also span
`src/lib/queries`, `src/lib/hooks`, and module-local `queries` and `hooks`.
Consequently, prefetch, client reads, optimistic mutation, invalidation, and SSE
updates can disagree about key or result shape.

### 5.5 Realtime ownership

`src/app/server-sent-events.tsx` currently owns connection lifecycle, event
classification, event types, analytics error reporting, and cache mutation for
several domains. It casts parsed JSON into local types and assumes particular
story cache shapes. The target keeps one shell-level connection but delegates
validated event semantics and cache changes to the resource owner.

### 5.6 Test feedback

The initial TypeScript run passed. Jest's TypeScript-configuration startup error
is repaired, and the complete combined-worktree run passes 220/220 suites and
1,279/1,279 tests with 0 snapshots in 33.601 seconds. That is a green Jest gate,
not a claim that type check, lint, production build, authorization, browser, or
deployment verification also passed. React Doctor remains a useful searchlight
for large components, effects, unstable rendering patterns, and boundary smells,
but its 57/100 score is heuristic and repository-sensitive.

## 6. Foundation and first slice delivered in this stage

This stage adds:

- a current-state inventory and target modular frontend architecture;
- an explicit `app -> shell/modules -> shared` dependency direction;
- an accepted decision to remain one modular Next.js application;
- explicit server-only and browser-safe API boundaries;
- query key/options and typed realtime cache ownership rules;
- an incremental migration and testing roadmap;
- a short review standard for new and migrated code;
- a dependency-free architecture scanner, eleven scanner tests, an exact debt
  baseline, and guarded baseline updates;
- a Projects pull-request workflow for the architecture ratchet, type check,
  complete Jest suite, and production build;
- a Jest configuration that reaches deterministic discovery;
- a 22.5% reduction in Maya's static system-prompt bytes while preserving the
  tested approval, authorization, scheduling, and presentation policies; and
- an invitations ownership slice that centralizes its keys and request paths,
  switches known callers, preserves the onboarding role mapping, and deletes
  four duplicate `src/lib` implementations.

The scanner is deliberately a Phase 0 mechanical subset. It currently ratchets
lower-layer-to-module imports, exact cross-module imports, module-cycle
membership, broad-barrel imports, legacy files, and production files above 700
lines. It does **not** classify required boundary authentication as architecture
debt and does not yet prove semantic public boundaries, server-only leakage,
unique query ownership, or shell/module cache ownership. This stage also does
not introduce the target `shell` or `shared` trees, relocate SSE, migrate other
capabilities, or establish production readiness.

## 7. Evidence and limitations

| Evidence                                                  | Result                                                                                                           | Limitation                                                                                                                        |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Static TypeScript/TSX inventory under `apps/projects/src` | 1,624 files and 184,854 physical lines                                                                           | Physical LOC includes tests and type declarations; it is not complexity or productivity.                                          |
| Top-level feature inventory                               | 26 directories under `src/modules`                                                                               | Directory count does not prove coherent ownership.                                                                                |
| Static import graph                                       | One 16-module strongly connected component                                                                       | Static analysis can miss runtime-computed imports and can overstate type-only or compatibility edges unless classified carefully. |
| Architecture production scan                              | 1,404 files; one 16-module SCC                                                                                   | The active categories are mechanical debt signals, not proof of all target boundaries.                                            |
| Architecture ratchet counts                               | 370/372 lower-layer; 298/298 cross-module; 1/16 cycles; 1,075/1,095 barrels; 170/170 legacy; 25/29,019 oversized | Counts are keys/units; the oversized unit is combined lines, not a violation count.                                               |
| `pnpm --filter projects architecture:test`                | Pass: 11/11 scanner tests                                                                                        | Tests prove scanner behavior represented by those cases, not application architecture correctness.                                |
| `pnpm --filter projects architecture:check`               | Pass against the initial exact baseline                                                                          | The baseline was captured after the invitations cutover; it is a ceiling, not evidence of that slice's before/after reduction.    |
| `.github/workflows/projects-quality.yml`                  | Configures architecture, type, complete Jest, and build checks for relevant pull requests                        | The workflow file is locally validated but has not executed remotely on this uncommitted worktree.                                |
| Executable `await auth()` inventory                       | 244 calls across 225 production files                                                                            | Count is not proof that authorization is correct, nor evidence that every call should be removed.                                 |
| Test file inventory and discovery                         | 220 files discovered                                                                                             | Discovery alone is not a complete-suite pass or coverage claim.                                                                   |
| Focused invitations Jest run                              | Pass: 2 suites and 2 tests                                                                                       | Covers key identity and onboarding mapping, not every request, permission, cache, or browser path.                                |
| `pnpm --filter projects type-check`                       | Pass on the combined worktree                                                                                    | Type checking does not execute requests, cache transitions, hydration, or browser behavior.                                       |
| Complete `pnpm --filter projects test -- --runInBand`     | Pass: 220/220 suites; 1,279/1,279 tests; 0 snapshots; 33.601s                                                    | Jest does not establish build, authenticated-browser, deployment, or complete architecture correctness.                           |
| `pnpm --filter projects build`                            | Pass: compiled in 39.1s; TypeScript 11.5s; 27 static pages                                                       | Build integration does not execute authenticated journeys or prove runtime data correctness.                                      |
| Focused ESLint over all changed app/scanner files         | Pass                                                                                                             | A focused pass is not a repository-wide lint pass.                                                                                |
| `pnpm --filter projects lint`                             | Final broad run red: 240 findings; 160 errors; 80 warnings                                                       | The changed files pass focused ESLint; untouched baseline debt prevents calling the repository-wide gate green.                   |
| React Doctor `--scope changed`                            | 90/100 across 29 changed files; no issues                                                                        | Advisory heuristic only; it does not assess untouched architecture debt or runtime behavior.                                      |
| React Doctor full-app scan                                | 57/100; 536 findings; 15 errors; 521 warnings                                                                    | Advisory heuristic only; each broad finding requires source correlation and focused verification.                                 |

The production build is green and focused changed-scope lint passes. The broad
lint baseline remains red. No authenticated browser journey, outbound-network
security test, bundle analysis, accessibility audit, network trace, remote CI
run, or deployment verification is claimed by this document.

## 8. Next acceptance milestone

The accepted ADR, initial architecture ratchet, Jest discovery repair, and first
invitations source cutover satisfy part of Phase 0 and start Phase 1. The next
acceptance milestone is to:

1. close the `/api/metadata` SSRF boundary with per-hop destination validation,
   network egress controls, response budgets, and adversarial tests;
2. repair and re-run the repository-wide lint baseline while preserving the
   green focused checks;
3. finish risk-proportionate invitations request, permission, cache, and browser
   characterization before calling that capability architecture-complete;
4. add the planned semantic public/private-boundary, server-leakage,
   query-owner, and shell/cache enforcement; and
5. choose the next low-risk complete behavior slice without expanding the
   invitations compatibility surface.

No commit, push, deployment, or production-state change is implied.
