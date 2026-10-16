# Delivery quality and operations

## 1. Status, ownership and evidence

This is a proposed delivery and operating contract, dated 16 October 2026. The product owner confirmed **1,000 concurrent users** as the initial capacity target. Connection counts, traffic rates, dataset sizes, latency objectives, retention periods, recovery objectives and rollout observation periods below are planning assumptions to validate. They are not measured capacity or production guarantees.

Read this alongside the [product specification](01-product-and-interactions.md), [backend architecture](02-backend-data-and-realtime.md), [Maya design](03-maya-orchestration.md), and [client/notification design](04-clients-and-notifications.md). Each work package needs one named accountable owner before implementation. Role labels below allocate responsibility without assuming a particular staffing level.

Maintain a release evidence register containing commit/build, migration version, environment, configuration, dataset, workload, client versions, results, defects and reviewer. Distinguish inspected code, local tests, staging measurements and production canary evidence. A skipped PostgreSQL test, unset test database, simulated push, browser-only fixture or successful model response cannot satisfy a different gate.

Current gaps must remain visible: existing SSE has no native-chat replay contract; current attachment code does not ship a malware-scanner provider; native notification behaviour needs platform validation; provider acceptance is not end-device delivery; chat-specific push receipt/reconciliation coverage has not been demonstrated. The committed story-comment Maya implementation now provides a reusable durable invocation engine, but native chat still needs its surface adapter, shared-audience policy and qualification. Existing personal Maya, Slack and document collaboration do not establish native-chat readiness.

## 2. Dependency-ordered work packages

The critical path is contracts → durable human communication → grounded Maya reads → bounded actions → private/client completion → general availability. Client design, operational tooling and evaluation datasets can proceed alongside backend work after their contracts stabilise. Do not parallelise by inventing conflicting message, actor or replay schemas.

| Package                                 | Accountable role; dependencies                  | Deliverable and acceptance evidence                                                                                                                                                                                                                                                     |
| --------------------------------------- | ----------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| F1: product and security contracts      | Product with security lead; none                | Approve membership/history, guests, DMs, retention, edit/delete, notification and native action policies. Document Maya's existing identity and human authorisation boundary. Complete failure-state designs and privacy threat model.                                                  |
| F2: SSE delivery and native benchmark   | Realtime/client lead; F1                        | Validate the chosen SSE+HTTP transport over the actual proxy on web, hosted desktop, iOS and Android. Benchmark full authorised projections for active rooms plus coalesced sidebar hints, HTTP typing, subscription/history races, reconnect and revocation.                           |
| F3: schema and domain foundation        | Backend lead; F1                                | Versioned message/event/API schemas, tenant constraints, SQLC, transaction boundaries, durable send identity, conversation event ordering and separate root/thread read positions. Real PostgreSQL tests prove commit/retry/removal races.                                              |
| F4: relay, sync and operations baseline | Backend/operations lead; F2–F3                  | Fenced outbox, bounded cursor-vector sync, replay floor/reset, slow-client handling, shared Redis subscriptions, telemetry and kill switches. Crash/reconnect tests recover every acknowledged message.                                                                                 |
| H1: human team-chat slice               | Client lead with backend; F3–F4                 | Global Chat navigation, canonical team room, composer, messages, threads, reactions, edit/delete, pins/saves, drafts, source routes and current Maya avatar projection. Two-user acceptance passes with real authenticated clients.                                                     |
| H2: files, search and lifecycle         | Backend lead; H1                                | Conversation-bound uploads, scanning/quarantine, authorised previews/downloads, search and deletion propagation. Cross-tenant, retained-reference, poisoned-file and stale-search tests pass. Restricted file beta is explicit if scanner is unfinished.                                |
| H3: attention and notifications         | Notifications/client lead; H1                   | Independent root/thread reads, attention badges, preferences, mute/snooze, foreground suppression and deep links. Event/recipient dedupe and current-access checks hold through retries and removal.                                                                                    |
| M1: grounded Maya reads                 | AI lead; H1, F4, security policy                | Extend `agentinvocations` with native-chat acceptance, surface routing, evidence manifests and audience-safe questions/summaries. Reuse leases, continuations, cancellation, publication/recovery and usage. Preserve comment behaviour; frozen/adversarial evaluations pass.           |
| M2: bounded work creation               | AI/backend lead; M1, story capability review    | Extend shared `Store.Execute` beyond its single atomic Apply/outcome with fenced per-item transactions, checkpoints and partial receipts. Preserve human authority, story validation, conditional claim, previews and audit provenance through crash, ambiguity and cancellation gates. |
| C1: private communication               | Backend/client lead; H1–H3, M1                  | Canonical two-human DMs and explicit private Maya navigation. No silent Maya access to human DMs. Privacy, removal, export, notifications and duplicate-DM races pass.                                                                                                                  |
| C2: native and offline completion       | Native/client lead; F2, H1–H3, M2               | Mobile/desktop interaction parity, bounded persistent client outbox, account-separated storage, lifecycle recovery, badge and notification actions, accessible routes and upgrade compatibility. Actual devices pass.                                                                   |
| G1: operational qualification           | Operations/quality lead; all preceding packages | Load/soak/fault results, restore rehearsal, support tools, budgets, scanner and notification operations, production canary and signed release checklist. Every release-blocking defect resolved.                                                                                        |

H1–H3 constitute the first human-team pilot; they do not imply DM/native completion. M1 can pilot before M2. Keep action execution disabled until its independent gate passes. Calendar estimates require named staffing, allocation, infrastructure readiness, existing-code integration cost and evaluation effort; do not infer a delivery date from this package count.

## 3. Capacity model and benchmark envelope

Model **users and connections separately**. A user can have multiple tabs or devices. The baseline assumes 1.5 active connections per concurrent user: **1,500 SSE connections**. Stress assumes three: **3,000 connections**. Neither ratio is confirmed customer behaviour. Include inactive-but-connected tabs and ensure hidden tabs do not create independent aggressive polling loops.

Use an initial synthetic dataset of 20 workspaces, 200 team rooms, 1,000 active users and five million retained messages, with skewed room sizes, long threads, edits, deletions and attachment references. These are test assumptions. Include a single workspace containing all 1,000 users; equal-sized tenants alone hide hot-workspace contention. Generate private-room/DM data for negative tests without copying production content.

| Workload              | Proposed input assumptions                                                                                                                                          | What must be measured                                                                                                                               |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Baseline conversation | 1,500 connections; 20 durable sends/second distributed across rooms; additional edits/reactions at 20% of send rate; 100 active typers updating every three seconds | Send/live-render/sync latency, authorisation cost, payload fanout, relay lag, database pool wait and connection memory.                             |
| Burst and hot room    | 3,000 connections; 100 sends/second for five minutes, including 20/second in one room with 300 viewers; separately, a room with 1,000 viewers                       | Room lock contention, fanout fairness, sync coalescing, client rendering and bounded queues.                                                        |
| Recovery storm        | Disconnect/reconnect 1,000 users over 30 seconds; each resumes up to 1,000 events in bounded pages                                                                  | Admission/backoff, replay correctness, database saturation and recovery time without duplicate sends.                                               |
| Lost final live event | Suppress the final active-room projection with no subsequent message; separately suppress an inactive-room sidebar hint                                             | Active conversation reconciles within the declared fallback interval; inactive summaries repair on next Chat/sidebar open, foreground or reconnect. |
| Mixed work            | Baseline chat plus 10 searches/second, 10 concurrent uploads and a separately controlled Maya workload                                                              | Human-chat isolation from indexing, storage, scanner and model activity.                                                                            |
| Endurance             | Baseline for eight hours; periodic workspace switches, reconnects, media failures and permission changes                                                            | Memory/goroutine leaks, stale subscriptions, lease recovery, autovacuum/index behaviour and telemetry volume.                                       |

At 1,000 foreground users, one visible-room reconciliation every 30 seconds produces about 33 requests/second; the proposed uniform 25–30-second jitter averages about 36 requests/second and preserves the 30-second maximum interval. Do not poll every room. Active-room projections avoid a normal fetch for each delivered message; account separately for gap recovery, coalesced sidebar refreshes, read updates, typing and notification traffic.

Run a protocol load generator for connection/throughput tests and a smaller representative set of real browsers/devices for rendering and lifecycle measurements. Declare hardware, replica count, database size/settings, Redis limits, proxy configuration, network latency and client mix. Run an isolated backend profile and a representative higher-latency mobile profile; report both. Avoid artificial production load without a separately authorised exercise.

No Kafka, database sharding or large gateway fleet is required by the confirmed target alone. First measure shared-subscription fanout, SQL query plans, connection memory and per-room lock wait. Extract or partition components only after a measured limit and migration plan exist.

### Growth checkpoints, not user caps

The confirmed 1,000-user figure is the first qualification workload, not a configured maximum or a claim that user 1,001 causes failure. The following checkpoints expose growth implications without assigning unmeasured capacity to an instance. Connection ratios remain assumptions; actual active views, traffic and tenant skew must come from measurement.

| Concurrent users | Connections at assumed 1.5/user | Stress at assumed 3/user | Visible-room checks/second at mean 27.5 seconds, assuming one foreground view/user |
| ---------------- | ------------------------------: | -----------------------: | ---------------------------------------------------------------------------------: |
| 1,000            |                           1,500 |                    3,000 |                                                                                 36 |
| 5,000            |                           7,500 |                   15,000 |                                                                                182 |
| 10,000           |                          15,000 |                   30,000 |                                                                                364 |
| 50,000           |                          75,000 |                  150,000 |                                                                              1,818 |

These checks are lightweight watermark requests, not full-history scans; unchanged views return bounded results. Hidden-view suppression and same-device coordination may reduce them; independent foreground devices may increase them. Requests/second equal active reconciliation views divided by mean interval. Preserve the declared fallback guarantee when changing schedules.

At each checkpoint, rerun mixed traffic, reconnect storms, slow consumers, hot-room fanout and failure/drain tests. Report the maximum workload meeting SLOs and headroom, rather than claiming a user ceiling from connection count. At 5,000, scrutinise pool budgets and replica scaling; at 10,000, qualify fleet-wide subscription sharing, recovery admission and hotspot isolation; at 50,000, require a fresh topology/cost review and sustained fleet tests. These are review priorities, not mandatory architecture changes at those numbers. Separate a realtime gateway only if connection memory/draining needs independent scaling; partition hot fanout or storage only when measurements justify it. SSE remains viable as the connection fleet grows; WebSocket does not remove database, fanout, authorisation or recovery work.

### Measurable constraints in the current repository

The [ECS release runbook](../../../apps/server/docs/operations/ecs-release.md) and [release workflow](../../../.github/workflows/ecs-fargate-release.yml) describe ECS/Fargate API and worker deployment, with task definitions fetched from AWS. The repo does not establish live CPU/RAM allocation, desired replica counts, autoscaling bounds, database/Redis service tiers or the actual bill. Inspect those settings through an authorised infrastructure review before sizing production.

Current defaults are 25 database connections per API process and 25 per worker process; replica growth multiplies the aggregate budget. API/worker Redis pool defaults also differ. The worker config hardcodes Asynq concurrency at 10 per process, and the current SSE hub starts three Redis listeners per connected client. These are specific things to profile and redesign where needed, not proven production bottlenecks. Budget total pools across all workloads, make queue allocation explicit, and replace per-connection subscriptions with shared gateway subscriptions before claiming efficient scale. The invocation recovery path currently processes a bounded batch of three sequentially; qualify recovery fairness and surface routing under shared-engine load. See [API config](../../../apps/server/cmd/api/config.go), [worker config](../../../apps/server/internal/bootstrap/worker/config.go), [worker bootstrap](../../../apps/server/internal/bootstrap/worker/bootstrap.go), [invocation worker](../../../apps/server/internal/bootstrap/worker/agent_invocations.go), and [SSE hub](../../../apps/server/internal/sse/hub.go).

## 4. Reliability objectives and measurements

The following are proposed launch objectives under the declared healthy-network baseline. Product and operations must approve them after the deployment benchmark. Use a rolling 30-day service window and per-deployment load reports; do not average away slow tenants or hot rooms.

| Journey                        | SLI and proposed objective                                                                                                                                                                                 | Accountable role   |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| Human send/read availability   | Successful eligible requests / eligible requests: ≥99.9%. Exclude malformed/forbidden requests, not unexpected overload or internal failures.                                                              | Backend/operations |
| Durable send                   | Ingress to committed response: p95 ≤300 ms, p99 ≤1 second. Report client-to-ack separately.                                                                                                                | Backend            |
| Visible delivery               | Commit to canonical render in an online foreground recipient: p95 ≤500 ms, p99 ≤2 seconds; includes authorised projection fanout and client merge/render. Report fallback/recovery separately.             | Realtime/client    |
| Replay                         | Reconcile 1,000 events: p95 ≤2 seconds under the declared test network; larger backlogs report progress.                                                                                                   | Backend/client     |
| Silent live-event loss         | Active conversation refreshed within 35 seconds under healthy fallback conditions: 30-second schedule plus bounded request allowance.                                                                      | Realtime/client    |
| Maya acknowledgement           | Visible durable queued/running state within 500 ms of accepted invocation when the client is online. Model completion is a separate SLI.                                                                   | AI/client          |
| Notification dispatch          | Eligible attention item reaches provider submission attempt within five seconds at p95 after its grace/quiet-hours delay expires. Measure source-event-to-attempt separately, including deliberate delays. | Notifications      |
| Privacy and mutation integrity | Zero unauthorised disclosure or duplicate committed action in qualification tests; any production occurrence is an incident, not an acceptable percentage.                                                 | Security/backend   |

Measure delivery using consenting test clients or privacy-safe sampled client acknowledgements of rendering. Keep fallback/recovery breakdowns, but do not remove delayed online-recipient samples from the overall delivery SLI merely because recovery was needed. A Redis publish, SSE write, push ticket or provider receipt is not a human read. Display device-delivery/read claims only when the protocol actually establishes them.

Document eligibility denominators and error classes before dashboards ship. Do not conceal throttled in-budget customers by classifying all 429s as success. Track intentional limit enforcement separately. This request-based 99.9% objective permits a 0.1% eligible-request failure budget; it cannot be converted into downtime minutes without a separately defined time-based availability measure. Production incidents consume the relevant budget even if unit tests passed.

## 5. Independent AI capacity and cost controls

The 1,000-user target does not imply 1,000 simultaneous model calls. Start qualification with an assumed one invocation/second and a five/second, 60-second burst. If average active run time is 20 seconds, the baseline implies approximately 20 concurrent runs before headroom. Measure actual duration, tool time and provider rate limits before setting worker concurrency.

Use separate queues/pools for human delivery, interactive Maya, notification dispatch and slower extraction/indexing. Reuse the shared invocation engine for comment and native-chat surfaces, with explicit fair capacity allocation; a separate queue is not a second run lifecycle. Apply workspace and requester admission, run/time/tool/token budgets, bounded retries and a global provider circuit breaker. Saturated AI displays a durable queued state or explicit capacity rejection while human messages continue normally. Waiting for human clarification/approval releases compute capacity. Stream previews must not hold a room lock or consume an unbounded token-event log.

Track queue age separately from inference duration and action execution. Proposed operating triggers are p95 interactive queue wait above five seconds and oldest eligible run above 30 seconds for a sustained interval. These are tuning inputs, not provider promises. Before enlarging concurrency, check token/request quotas, database/tool saturation and tenant fairness.

Budget with measured quantities:

- Daily model cost = sum across billable provider inference attempts of input/output tokens × their applicable unit prices, including retries and separate cached/reasoning/tool charges where applicable. A per-invocation estimate must aggregate all its attempts. Exact assent or other control transitions that make no inference call incur no new model-token spend.
- Retained storage = daily messages × measured stored bytes × retention days × measured index/replica overhead, plus attachments, revisions, events and backups.
- Delivery traffic = event/hint bytes × actual eligible online recipients, plus sync responses, reconnects and previews; do not equate sends with delivered bytes.
- Capacity = arrival rate × mean service time, then add measured burst and failure headroom. Service times differ for summaries, retrieval and multi-action runs.

Record model/version and pricing date in the budget sheet at implementation time. Set workspace spend ceilings, warning thresholds and graceful budget-exhaustion behaviour; no secret switch to a less-tested action model. Cost dashboards use bounded dimensions, not message text.

### Hosting cost model

Separate the existing platform baseline, incremental human-chat hosting, and incremental Maya/provider spend. Do not quote dollar totals from concurrency alone. A useful monthly model is API/gateway/worker task CPU and memory hours + load-balancer hours and measured capacity usage + database/Redis capacity and storage + object storage/requests/egress + logs/metrics/backups + scanning and any applicable NAT/cross-zone traffic. Apply current region/service prices to an observed workload and chosen redundancy; record both average and peak task counts. Frontend hosting is a separate line, and the current browser connects its SSE stream to the Go API rather than holding a frontend function open.

At 5,000/10,000/50,000 users, estimate each component from message rate, audience fanout, retained bytes, connections, CPU/memory per connection, recovery rate, attachment traffic and AI adoption. Cost need not scale exactly with users: shared fixed capacity, redundancy steps, large rooms and heavy AI usage change the curve. Neither SSE nor WebSocket is inherently cheaper; compare the actual payloads, HTTP command volume, avoided fetches, heartbeat traffic, connection behaviour and operational complexity. Benchmark the selected full-projection/hint split before changing protocol. Live infrastructure sizes and negotiated provider prices remain unknown until verified.

## 6. Verification gates

### Persistence and communication

Real PostgreSQL/Redis tests must cover concurrent room writes, rollback, identical and conflicting retries, ambiguous HTTP responses, outbox claim fencing, crash after publish, expired leases and consumer dedupe. Prove per-room event commit order, separate root/reply positions and no global-sequence assumption. Test acknowledged subscription/history races, bounded buffers, cross-replica subscription commands, overlapping sync/live events, out-of-order full projections, missing sequence gaps, replay expiry, lost final projection/hint, server drain and Redis subscription restart. A newer entity projection must never skip the contiguous replay cursor over a missing event.

Read tests cover multiple devices, background tabs, partial visibility, search jumps, manual unread and independent thread reads. Edit/delete tests preserve local conflict text, thread tombstones and source provenance. A generic idempotency receipt expiring must not make an old client-outbox send duplicate its committed message.

### Privacy, security and lifecycle

Maintain a matrix of active/removed/deactivated actors across two workspaces, private/public teams and DMs. Check messages, members, counts, snippets, replay, attachment IDs, previews, citations, exports and notification excerpts. Remove access while a stream, upload, search, model run and action are active. No unrestricted workspace broadcast may expose private room IDs or activity.

Adversarial cases include malicious rich text, mention spoofing, quoted instructions, prompt injection in files/links, decompression bombs, infected files, SSRF, cross-tenant attachment reuse and stale signed links. Verify deletion propagation through revisions, search, derived summaries, AI evidence, pending delivery and object references. Restore testing must reapply deletion/access tombstones before recovered data becomes visible.

### Maya and actions

Freeze evaluation cases before tuning, maintain a held-out set and run model/prompt/tool changes against it. Include disagreement, sarcasm, absent owners/dates, duplicate work, exact last-ten selection, permission changes, stale citations and malicious source instructions. A readable summary is insufficient if it changes a tentative idea into a decision.

Run existing story-comment engine regressions alongside native-chat adapter tests. Prove atomic source/invocation acceptance, surface-isolated claims and controls, immutable source identity, continuation ownership, post-commit dispatch recovery, publication dedupe and usage attribution. Old comment-only workers must not claim chat runs; new routing must preserve existing comment execution and recovery. Reuse does not waive the new shared-audience and multi-action safety gates.

Test multiple pending Maya questions/proposals in one thread. The persisted exact reply target and expected source/proposal versions must select only the intended continuation; retries cannot substitute another target, and a generic thread reply must not become inferred approval.

Action tests interrupt every boundary: before mutation, after commit before receipt response, between batch items and before publication. Retry remaining items using stable action IDs. Test simultaneous ticket claims, cancellation after partial completion and model failure after successful creation. Execute domain changes under the initiating human's current authority; Maya's system identity/provenance must not bypass normal story-service validation. Verify actor, initiating human, timestamp, before/after and source links in visible audit output.

### Clients, notifications and accessibility

Use real two-user journeys across supported web browsers, Safari, hosted desktop and physical iOS/Android devices. Verify background/suspended/terminated launches, OS permission denial, rotated/invalid push tokens, expired sessions, account/workspace switches, offline restart, duplicate taps and old installed versions. Persist pending content with account/workspace isolation; never send it under the next signed-in account.

Exercise VoiceOver and NVDA, keyboard-only use, IME, touch, reduced motion, 200–400% zoom, contrast and long/localised content. Verify focus restoration, virtualised-feed navigation, upload controls and restrained live announcements. Real device acceptance and accessibility review are release gates, not follow-up polish.

Notification evidence distinguishes eligibility, queued, provider accepted, receipt/error and app-open events. Implement and test chat-specific receipt reconciliation where the provider supports it; the current repository does not prove this complete. Test one coherent Maya result notification, mute/quiet-hours, cross-device suppression and deep links to removed content. Keep routine email aligned with the established weekly preference unless the user explicitly changes it.

## 7. Migrations and mixed-version deployment

The refreshed repository baseline is HEAD `0ee7c80448b722eeca817e60b76c630cfdf957f1`, with a clean working tree before this documentation update. Story-comment collaboration/attachments **000219/000220** are committed, followed by **000221** agent invocations, **000222** native Maya update proposals, **000223** assistant usage, **000224** thread reads, and **000225/000226** reaction catalogue/expansion. Preserve that chain and its generated contracts. This replaces the earlier uncommitted-work observation; it does not prove migrations or features are deployed. Recheck HEAD and migration ownership when implementation begins, then allocate fresh versions without rewriting any applied migration.

Follow the repository migration manifest and operations contract. Each chat migration requires ownership, classification, rollout mode, API/worker compatibility, bounded backfill/index strategy and recovery instructions. Extend the existing agent-invocation schema additively for evidence/actions rather than adding a parallel lifecycle. Use additive schema-first deployment with chat flags off. Backfill canonical team rooms idempotently in bounded batches; concurrent team creation must use the same uniqueness invariant.

Deploy readers that tolerate the additive schema, then compatible workers/API, then clients, then enable writes. Version message bodies/events and negotiate installed-client capability. Test the oldest supported client against the newest compatible API. Unknown optional events request a safe refresh; incompatible required changes need a version gate. Do not allow old workers to claim tasks whose payload or fencing semantics they cannot honor.

Once chat data exists, rollback normally means flags and compatible binaries, not destructive down migrations. Preserve messages, AI action receipts and deletion ledgers. A forward fix is mandatory where downgrading would discard acknowledged data or remove an authorisation fence.

## 8. Phased exposure and rollback

Use server-enforced, workspace-scoped switches for chat reads/admission, live delivery, media, search, Maya reads, Maya actions, DMs and external notifications. Flags must not bypass permissions. Maintain a canary allowlist and an audited record of changes. Chat-specific Maya flags govern the new surface; they must not disable existing story-comment Maya as an accidental consequence.

1. **Foundation:** internal synthetic accounts only. Pass ordering, isolation, replay and transport gates before real team use.
2. **Human team pilot:** selected internal teams, then invited workspaces. Enable complete human journeys and observe usage/failure reports. Keep unavailable media explicitly restricted.
3. **Maya read pilot:** source-linked questions/summaries with evaluation, budget and disclosure controls. Leave actions disabled.
4. **Bounded action pilot:** enable only approved native-chat operations after crash/retry and audit gates. Personal Maya and Slack policies remain separately controlled.
5. **Private/client completion:** gate DMs, full native lifecycle and persistent offline sends independently. Clear parity/privacy defects before claiming the complete planned feature.
6. **General availability:** expand workspaces gradually, for example 1%, 5%, 25%, then all eligible workspaces. Percentages and proposed 24–48-hour observation windows are rollout assumptions; advance on evidence and representative traffic, not elapsed time alone.

Pause expansion on SLO burn, unexpected queue growth, repeated client resets, scanner/privacy failure or unexplained action outcomes. Disable Maya actions immediately for duplicate or uncertain execution defects; preserve receipt reconciliation. Disable affected notification/media paths independently. For human-send integrity faults, stop new admission, preserve pending drafts, and retain authorised history where safe. Do not erase data or replay all jobs blindly to recover.

## 9. Dashboards, runbooks and recovery

Provide dashboards for user journeys, database/pool/lock pressure, SSE health/subscriptions, outbox age, replay/reset rate, client rendering, queue fairness, media scan age, index lag, notification outcomes and Maya usage. Alert on sustained user-impacting symptoms with actionable saturation signals. Use multiple-window error-budget burn alerts; queue-age alarms alone do not establish a customer-visible outage.

Correlate request → message → event → run/action/delivery through safe IDs in restricted logs/traces. Follow the repository's sensitive-data rules: no message bodies, raw prompts, signed URLs, cookies, credentials or unrestricted provider errors. Metrics must not use message/user/workspace IDs as unbounded labels. Define access, retention and sampling for exceptional AI diagnostics.

| Incident                                   | Required runbook actions                                                                                                                                                |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Redis/gateway outage                       | Check readiness policy, stop misleading healthy-stream state, use bounded sync/reconnect, restore subscriptions, reconcile watermarks and watch recovery load.          |
| Database outage or saturation              | Reject uncommitted sends; preserve client identities; inspect pool/locks and stop excess admission. Reconcile ambiguous sends from canonical records after recovery.    |
| Outbox/worker stall                        | Inspect oldest age and lease owner, repair dependency, recover fenced claims in bounded batches, quarantine poison work. Never bypass dedupe.                           |
| Maya/provider failure or duplicate concern | Disable actions as needed; inspect durable receipts and authority; reconcile actual resources before retry. Report partial results accurately.                          |
| Notification/scanner failure               | Pause the affected external path, preserve durable intent, check current access before retry, keep unsafe files unavailable.                                            |
| Disclosure or deletion defect              | Contain affected reads/publication, revoke caches/URLs where possible, preserve restricted evidence, follow the security incident process and validate deletion repair. |

Proposed disaster-recovery goals are **RPO ≤5 minutes and RTO ≤60 minutes** for a regional/database loss, subject to actual PostgreSQL backup/WAL/replication infrastructure and restore rehearsal. These goals are not verified and are distinct from zero loss of acknowledged writes during ordinary application/worker crashes against a surviving database. Stronger regional durability requires an explicit infrastructure/cost decision.

Rehearse restoration into an isolated environment, then verify membership, deletion tombstones, canonical message/action identities, outbox state and attachment references before reopening traffic. Record measured recovery time and recovered-data boundary. Do not report successful backup creation as proof of restore readiness.

## 10. Definition of release completeness

General availability requires named owners, approved semantics, full product acceptance journeys, real database/fault tests, AI quality/security gates, accessible web/native flows, migration/upgrade rehearsal, scanner operations, notification reconciliation, retention/deletion verification, measured initial 1,000-user qualification, restore evidence and healthy production canary results. Publish the tested workload/headroom and an owner for requalification before projected traffic exceeds it; the initial qualification is not a permanent user cap. Support must be able to investigate a failed send/run from a safe reference ID without reading unrelated conversations.

Record known limitations explicitly and restrict the corresponding capability until its gate passes. A partial pilot can be useful, but screenshots, compilation, happy-path model demos and successful deployment do not establish the complete feature.
