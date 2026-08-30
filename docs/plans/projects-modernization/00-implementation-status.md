# FortyOne Projects Modernization: Implementation Status

**Status:** Phase 0 foundation, request-boundary hardening, and multiple
behavior-preserving cohesion slices implemented
**Snapshot date:** 2026-08-30
**Scope:** `apps/projects` source architecture, server/browser boundaries, data
ownership, realtime cache behavior, tests, developer workflow, and internal
documentation  
**Target:** [current state and target architecture](./01-current-state-and-target-architecture.md)  
**Delivery plan:** [delivery, testing, and documentation roadmap](./02-delivery-testing-and-documentation-roadmap.md)

## Completion update — 2026-08-30

This update takes precedence over the historical snapshots below. The latest
local architecture check scans 1,618 production TypeScript/TSX files, reports
an acyclic module graph, and finds no production file at or above 700 lines.
Focused Strategy Map and Roadmap scale tests pass with 50 objectives and key
results. This is local source-and-test validation only; it does not establish
authenticated-browser, deployed, or production behavior. No GitHub Actions were
added or changed, at the user's request.

This is an evidence ledger, not a completion announcement. The governing
documents, enforceable architecture ratchet, deterministic Jest discovery,
invitations ownership slice, and clean repository-wide lint gate now exist.
This checkpoint hardens outbound metadata, AI-suggestion, Maya chat, and
reasoning-Markdown boundaries; refactors several high-churn screens; and adds
focused regression coverage for the moved behavior. The wider Projects source
tree still has a 16-module cycle, 14 oversized files, incomplete semantic
boundary checks, and no authenticated-browser or deployment verification. No
commit, push, deployment, or production-state change is claimed here.

## 1. Executive checkpoint

| Area                        | Current state                                       | Evidence                                                                                                                                                                         | Acceptance still required                                                                                                    |
| --------------------------- | --------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Architecture definition     | Foundation documented and ADR accepted              | One target architecture, one delivery roadmap, one review standard, and accepted ADR 0001                                                                                        | Apply the standard to each slice and keep the evidence ledger current                                                        |
| Architecture ratchet        | Enforced mechanical subset active                   | `architecture:test` passes 21 scanner tests; `architecture:check` scans 1,515 production files and passes against the guarded baseline                                           | Add client/server public-export safety, server-leakage, query-owner, and shell/cache checks                                  |
| Pull-request enforcement    | Core quality gates configured locally and in CI     | `.github/workflows/projects-quality.yml` runs architecture integrity, changed-file lint/format checks, type check, complete Jest, and production build for relevant changes      | Observe required checks on a protected pull request; retain CODEOWNERS/reviewer enforcement outside repository source        |
| Source migration            | Ten focused cohesion/ownership slices implemented   | Story options, New Story, Story subtasks, filters, Integration Requests, Calendar, Strategy map, invitations, and Maya policy each have explicit seams and focused tests         | Continue one complete behavior slice at a time; do not confuse lower line counts with complete module ownership              |
| Module boundaries           | Coupling is declining, still material               | Cross-module imports fell to 271 units; one strongly connected component still contains 16 feature modules                                                                       | Stop new cross-boundary imports, reduce the component monotonically, then reach an acyclic production graph                  |
| Server/browser API boundary | High-risk routes hardened; wider policy distributed | Metadata uses validated/pinned outbound fetches; AI suggestions reload canonical server resources; malformed Maya messages fail before reservation; reasoning HTML is sanitized  | One request-context contract, egress defense in depth, and per-slice authorization/browser tests                             |
| Query/cache ownership       | Inconsistent overall; invitations consolidated      | Invitation keys moved from `src/constants/keys.ts` to the owning module, while other keys and data functions remain split across `src/constants`, `src/lib`, and feature modules | One key/options owner per remaining server-state resource and one cache policy shared by SSR, hooks, mutations, and realtime |
| Realtime behavior           | Functional, centrally coupled                       | `src/app/server-sent-events.tsx` parses unvalidated JSON and directly patches stories while invalidating calendar and notification data                                          | Runtime-validated events, module-owned cache reducers, ordering/idempotency tests, and a shell-owned connection lifecycle    |
| Maya request policy         | Explicit server-only seams and request decoder      | Active-tool selection, approval policy, execution ledger, and stream finalization are split; malformed UI-message structures return 400 before downstream work                   | Measure real request/token impact; keep approval, authorization, and scheduling policy tests as hard compatibility evidence  |
| Type feedback               | Passing                                             | `pnpm --filter projects type-check` passes on the combined worktree                                                                                                              | Keep it required on every migration checkpoint                                                                               |
| Production build            | Passing                                             | `pnpm --filter projects build` compiled in 11.5s, type-checked in 1.635s, and generated 28 static pages                                                                          | Add authenticated runtime and browser evidence where the slice requires it                                                   |
| Lint feedback               | Repository-wide gate green                          | `pnpm --filter projects lint` completes with 0 errors and 0 warnings                                                                                                             | Keep the green gate required; do not reintroduce a warning baseline                                                          |
| Jest feedback               | Complete suite passing                              | CommonJS `next/jest` configuration reaches deterministic discovery; the complete run passes 260/260 suites and 1,469/1,469 tests                                                 | Keep the complete suite in CI and treat future failures as owned regressions                                                 |
| Test assets                 | Discoverable and executing                          | 255 `*.test.ts`/`*.test.tsx` files exist, including route-boundary, dialog, OTP, calendar, Strategy, and onboarding regressions                                                  | Strengthen permission, cache, realtime, accessibility, and authenticated-browser risk coverage                               |
| Heuristic diagnostics       | Improved, advisory only                             | React Doctor reports 88/100 on the final changed scope with no errors; remaining findings are reviewed performance/maintainability hypotheses                                    | Correlate each finding with source evidence; do not use this score as a release gate or truth metric                         |

## 2. Quantitative baseline

The initial analysis snapshot contained 1,622 TypeScript/TSX files, 184,890
physical lines, and 218 Jest files. After the current implementation slices, the
source worktree contains:

- 1,775 TypeScript and TSX files under `apps/projects/src`;
- 198,708 physical lines in those files, approximately 199,000 lines;
- 26 top-level directories under `src/modules`;
- one static module-import strongly connected component with 16 members;
- 1,515 production TypeScript/TSX files scanned by the architecture check;
- 245 executable `await auth()` calls across 226 production files; and
- 255 Jest test files.

The active ratchet records 368 lower-layer import keys / 370 imports, 271
cross-module keys / 271 imports, one cycle / 16 modules, 1,064 broad-barrel keys
/ 1,080 imports, 170 legacy files, and 14 oversized files / 14,885 combined
lines. The `await auth()` count is a separate analysis inventory, not a
ratcheted debt category.

These counts describe code shape, not product quality, coverage, or security
approval. The complete Jest and production-build passes are meaningful evidence,
but they do not exercise authenticated browser journeys or every architecture
and authorization contract.

## 3. Architecture-readiness score

The score is **7.3/10**. It measures source and architecture readiness for safe,
predictable change. It is not a rating of FortyOne's product, feature set,
visual quality, or customer value.

| Dimension            |  Score | Evidence-based interpretation                                                                                                                                              |
| -------------------- | -----: | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Architecture         | 6.2/10 | Exact public paths, a hard source-size ratchet, and 25 fewer private cross-module edges prevent new mechanical debt, but the 16-module cycle and mixed root layers remain. |
| Correctness          | 7.8/10 | Strict TypeScript, 255 Jest suites, production build, and regression tests for high-risk state flows pass; authenticated browser evidence remains incomplete.              |
| Security             | 8.0/10 | Metadata SSRF, chat-message decoding, canonical AI suggestions, and raw reasoning HTML are hardened with adversarial tests; authorization ownership remains distributed.   |
| Performance          | 6.8/10 | Settings request work is parallelized and throwaway QueryClient prefetching is fixed, but bundle analysis and authenticated interaction traces remain incomplete.          |
| Testing              | 7.5/10 | The full suite is green and new tests cover security, dialogs, calendar/Strategy pure behavior, onboarding races, and Suspense; browser and accessibility evidence remain. |
| Maintainability      | 7.2/10 | Lint is clean and major high-churn surfaces are cohesive, but 14 oversized files, broad roots, and a module cycle still raise change cost.                                 |
| Developer experience | 7.8/10 | Architecture integrity, lint, type, Jest, build, formatting, and baseline verification are repeatable; semantic gates and remote CI evidence remain incomplete.            |

The rounded arithmetic mean is 7.3. Future updates must change a score only
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

`src/app/api/metadata/route.ts` now authenticates before outbound work and uses
an explicit public-target policy. HTTP(S) only, default ports, all-public DNS
answers, per-socket address pinning, redirect-hop revalidation, an eight-second
total timeout, identity encoding, HTML content types, and a one-MiB body ceiling
are enforced. Transport failures can fall through the complete validated DNS
answer set without performing a second unvalidated lookup.

Favicon and social-image URLs are no longer returned as third-party browser
destinations. An authenticated same-origin proxy applies the same destination
and redirect policy, admits a raster MIME allowlist, rejects SVG and compressed
responses, caps the body at two MiB, and returns defensive response headers.
Sixty-six focused metadata tests cover URL/IP/DNS/redirect policy, byte and MIME
budgets, address fallback, proxy authentication, and error mapping. Network
egress controls remain a defense-in-depth deployment responsibility; source
validation is not a substitute for infrastructure policy.

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
is repaired, and the final combined-worktree run passes 260/260 suites and
1,469/1,469 tests with 0 snapshots in 26.953 seconds. That is a green Jest gate,
not a claim that authorization, authenticated browser behavior, or deployment
verification also passed. React Doctor remains a useful searchlight for large
components, effects, unstable rendering patterns, and boundary smells; its
final changed-scope score is 88/100 with no errors, but it is heuristic and
repository-sensitive.

### 5.7 High-density strategy and roadmap rendering

The Strategy Map and Roadmap no longer mount or query every objective merely
because it is present in the workspace. The Strategy Map culls offscreen cards
and connector paths, reserves key-result geometry from the server-provided
count, and activates key-result queries only for visible, selected, dragged, or
persisted-position owners. Roadmap Gantt uses one shared vertical window for
its sidebar and timeline; list and Kanban use measured virtual windows and pin
selected, focused, or actively dragged items.

Two deterministic component workloads provide the current regression evidence:
`strategy-map-scale.test.tsx` renders 50 objectives with two key results each,
then moves to a middle viewport while keeping at most four active objective
queries and eight mounted key-result cards. `roadmap-large-dataset.test.tsx`
uses 50 objectives (25 with two key results), selects its midpoint in Gantt,
list, and Kanban, and verifies bounded mounted items and key-result
subscriptions. These tests establish no-crash and bounded-work behavior for
that representative dataset; they do not replace authenticated browser
profiling, variable-height drag/auto-scroll testing, or production telemetry.

The supporting cohesion work keeps the high-density behavior reviewable: the
Base Gantt shell is 271 lines (from 1,279), the Roadmap board host is 325
lines (from 761), and the Strategy Map coordinator is 386 lines (from 985).
The architecture scan consequently records 14 oversized production files rather
than 17 before this slice. These reductions preserve the public Gantt API and
the page-owned Strategy/Roadmap selection contracts; they are not a claim that
the remaining architecture debt is resolved.

## 6. Foundation and first slice delivered in this stage

This stage adds:

- a current-state inventory and target modular frontend architecture;
- an explicit `app -> shell/modules -> shared` dependency direction;
- an accepted decision to remain one modular Next.js application;
- explicit server-only and browser-safe API boundaries;
- query key/options and typed realtime cache ownership rules;
- an incremental migration and testing roadmap;
- a short review standard for new and migrated code, including cohesion review
  at 400–700 lines and behavior-slice refactoring above the 700-line ceiling;
- a dependency-free architecture scanner, twenty-one scanner tests, an exact
  debt baseline, guarded baseline updates, and baseline-integrity verification;
- a Projects pull-request workflow for the architecture ratchet, changed-file
  lint/format checks, type check, complete Jest suite, and production build;
- a Jest configuration that reaches deterministic discovery;
- a hardened metadata HTML/image fetch path with per-hop SSRF policy and 66
  focused adversarial tests;
- a non-overridable source-size ceiling and explicit module-public-boundary
  policy serialized into the baseline;
- cohesive Story options, Story subtask suggestions, New Story, filters,
  Integration Requests, Calendar, and Strategy-map extractions that remove
  nine high-churn files from oversized debt while preserving public behavior;
- Maya active-tool and mutation-approval policy seams, a structural request
  decoder that rejects malformed messages before any downstream work, and
  focused approval/ledger regression tests;
- server-owned AI suggestion endpoints that load canonical workspace resources,
  reject guests, bound input, and avoid trusting browser-supplied prompt data;
- sanitization for the dormant reasoning Markdown renderer after raw HTML
  parsing, with malicious-content regressions; and
- an invitations ownership slice that centralizes its keys and request paths,
  switches known callers, preserves the onboarding role mapping, and deletes
  four duplicate `src/lib` implementations.

The scanner is deliberately a Phase 0 mechanical subset. It currently ratchets
lower-layer-to-module imports, cross-module imports outside exact `public.ts` or
`public/**` boundaries, module-cycle membership including public dependencies,
broad-barrel imports, legacy files, and production files above 700 lines.
Existing oversized entries may only shrink; new or growing entries cannot be
accepted by the emergency ADR-backed baseline writer. It does **not** classify
required boundary authentication as architecture debt and does not yet prove
public-export browser/server safety, server-only leakage, unique query ownership,
or shell/module cache ownership. This stage also does not introduce the target
`shell` or `shared` trees, relocate SSE, migrate other capabilities, or establish
production readiness.

## 7. Evidence and limitations

| Evidence                                                  | Result                                                                                                           | Limitation                                                                                                                        |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Static TypeScript/TSX inventory under `apps/projects/src` | 1,775 files and 198,708 physical lines                                                                           | Physical LOC includes tests and type declarations; it is not complexity or productivity.                                          |
| Top-level feature inventory                               | 26 directories under `src/modules`                                                                               | Directory count does not prove coherent ownership.                                                                                |
| Static import graph                                       | One 16-module strongly connected component                                                                       | Static analysis can miss runtime-computed imports and can overstate type-only or compatibility edges unless classified carefully. |
| Architecture production scan                              | 1,515 files; one 16-module SCC                                                                                   | The active categories are mechanical debt signals, not proof of all target boundaries.                                            |
| Architecture ratchet counts                               | 368/370 lower-layer; 271/271 cross-module; 1/16 cycles; 1,064/1,080 barrels; 170/170 legacy; 14/14,885 oversized | Counts are keys/units; the oversized unit is combined lines, not a violation count.                                               |
| `pnpm --filter projects architecture:test`                | Pass: 21/21 scanner tests                                                                                        | Tests prove scanner behavior represented by those cases, not application architecture correctness.                                |
| `pnpm --filter projects architecture:check`               | Pass against the guarded exact baseline                                                                          | The baseline remains a ceiling; integrity verification rejects ordinary self-blessing of new debt.                                |
| `.github/workflows/projects-quality.yml`                  | Configures architecture integrity, changed lint/format, type, complete Jest, and build checks                    | The workflow file is locally validated but has not executed remotely on this uncommitted worktree.                                |
| Executable `await auth()` inventory                       | 245 calls across 226 production files                                                                            | Count is not proof that authorization is correct, nor evidence that every call should be removed.                                 |
| Test file inventory and discovery                         | 255 files discovered                                                                                             | Discovery alone is not a complete-suite pass or coverage claim.                                                                   |
| Focused invitations Jest run                              | Pass: 2 suites and 2 tests                                                                                       | Covers key identity and onboarding mapping, not every request, permission, cache, or browser path.                                |
| `pnpm --filter projects type-check`                       | Pass on the combined worktree                                                                                    | Type checking does not execute requests, cache transitions, hydration, or browser behavior.                                       |
| Complete `pnpm --filter projects test -- --runInBand`     | Pass: 260/260 suites; 1,469/1,469 tests; 0 snapshots; 26.953s                                                    | Jest does not establish build, authenticated-browser, deployment, or complete architecture correctness.                           |
| `pnpm --filter projects build`                            | Pass: compiled in 11.5s; TypeScript 1.635s; 28 static pages                                                      | Build integration does not execute authenticated journeys or prove runtime data correctness.                                      |
| `pnpm --filter projects lint`                             | Pass: 0 errors; 0 warnings                                                                                       | Lint is a static quality gate, not runtime or authorization proof.                                                                |
| React Doctor `--scope changed`                            | 88/100; 0 errors; 66 advisory warnings                                                                           | The remaining findings require source-by-source ownership decisions; this is not a release gate.                                  |

The production build, full lint, type check, architecture integrity, and full
Jest suite are green. No authenticated browser journey, infrastructure
egress-policy test, bundle analysis, accessibility audit, network trace, remote
CI run, or deployment verification is claimed by this document.

## 8. Next acceptance milestone

The accepted ADR, initial architecture ratchet, Jest discovery repair, and first
invitations source cutover satisfy part of Phase 0 and start Phase 1. The next
acceptance milestone is to:

1. add infrastructure egress enforcement around the now-hardened metadata
   request path and exercise it in deployment-level security verification;
2. reduce the remaining 16-module cycle, 14 oversized files, legacy roots, and
   broad barrels under the guarded mechanical baseline;
3. finish risk-proportionate invitations request, permission, cache, and browser
   characterization before calling that capability architecture-complete;
4. add the planned semantic public/private-boundary, server-leakage,
   query-owner, and shell/cache enforcement; and
5. add authenticated browser, accessibility, bundle, and deployment-level
   verification for the high-risk journeys now covered by source and Jest tests.

No commit, push, deployment, or production-state change is implied.
