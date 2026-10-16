# Delivery quality and operations

## 1. Status, ownership and evidence

This is a proposed delivery and operating contract, dated 16 October 2026. The product owner confirmed **1,000 concurrent users** as the initial capacity target. Connection counts, traffic rates, dataset sizes, latency objectives, retention periods, recovery objectives and rollout observation periods below are planning assumptions to validate. They are not measured capacity or production guarantees.

Read this alongside the [product specification](01-product-and-interactions.md), [backend architecture](02-backend-data-and-realtime.md), [Maya design](03-maya-orchestration.md), and [client/notification design](04-clients-and-notifications.md). Each work package needs one named accountable owner before implementation. Role labels below allocate responsibility without assuming a particular staffing level.

Maintain a release evidence register containing commit/build, migration version, environment, configuration, dataset, workload, client versions, results, defects and reviewer. Distinguish inspected code, local tests, staging measurements and production canary evidence. A skipped PostgreSQL test, unset test database, simulated push, browser-only fixture or successful model response cannot satisfy a different gate.

Current gaps must remain visible: existing SSE has no native-chat replay contract; current attachment code does not ship a malware-scanner provider; native notification behaviour needs platform validation; provider acceptance is not end-device delivery; chat-specific push receipt/reconciliation coverage has not been demonstrated. Existing personal Maya, Slack and document collaboration do not establish native-chat readiness.

## 2. Dependency-ordered work packages

The critical path is contracts → durable human communication → grounded Maya reads → bounded actions → private/client completion → general availability. Client design, operational tooling and evaluation datasets can proceed alongside backend work after their contracts stabilise. Do not parallelise by inventing conflicting message, actor or replay schemas.

| Package                                 | Accountable role; dependencies                  | Deliverable and acceptance evidence                                                                                                                                                                                                                         |
| --------------------------------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| F1: product and security contracts      | Product with security lead; none                | Approve membership/history, guests, DMs, retention, edit/delete, notification and native action policies. Document Maya's existing identity and human authorisation boundary. Complete failure-state designs and privacy threat model.                      |
| F2: transport and native spike          | Realtime/client lead; F1                        | Prove authenticated SSE over the actual proxy on web, hosted desktop, iOS and Android. Exercise foreground/background, buffering, reconnect and revocation. Validate HTTP typing. Retain SSE unless a measured limitation justifies a reviewed alternative. |
| F3: schema and domain foundation        | Backend lead; F1                                | Versioned message/event/API schemas, tenant constraints, SQLC, transaction boundaries, durable send identity, conversation event ordering and separate root/thread read positions. Real PostgreSQL tests prove commit/retry/removal races.                  |
| F4: relay, sync and operations baseline | Backend/operations lead; F2–F3                  | Fenced outbox, bounded cursor-vector sync, replay floor/reset, slow-client handling, shared Redis subscriptions, telemetry and kill switches. Crash/reconnect tests recover every acknowledged message.                                                     |
| H1: human team-chat slice               | Client lead with backend; F3–F4                 | Global Chat navigation, canonical team room, composer, messages, threads, reactions, edit/delete, pins/saves, drafts, source routes and current Maya avatar projection. Two-user acceptance passes with real authenticated clients.                         |
| H2: files, search and lifecycle         | Backend lead; H1                                | Conversation-bound uploads, scanning/quarantine, authorised previews/downloads, search and deletion propagation. Cross-tenant, retained-reference, poisoned-file and stale-search tests pass. Restricted file beta is explicit if scanner is unfinished.    |
| H3: attention and notifications         | Notifications/client lead; H1                   | Independent root/thread reads, attention badges, preferences, mute/snooze, foreground suppression and deep links. Event/recipient dedupe and current-access checks hold through retries and removal.                                                        |
| M1: grounded Maya reads                 | AI lead; H1, F4, security policy                | Durable run lifecycle, exact source manifests, questions/summaries, shared-audience retrieval, citations, cancellation and resumable results. Frozen evaluation set and adversarial cases meet the AI document's thresholds.                                |
| M2: bounded work creation               | AI/backend lead; M1, story capability review    | Typed actions, actual story-service validation, human authority, stable per-item receipts, conditional claim, real previews and audit provenance. Crash, ambiguity, partial-failure and cancellation gates pass.                                            |
| C1: private communication               | Backend/client lead; H1–H3, M1                  | Canonical two-human DMs and explicit private Maya navigation. No silent Maya access to human DMs. Privacy, removal, export, notifications and duplicate-DM races pass.                                                                                      |
| C2: native and offline completion       | Native/client lead; F2, H1–H3, M2               | Mobile/desktop interaction parity, bounded persistent client outbox, account-separated storage, lifecycle recovery, badge and notification actions, accessible routes and upgrade compatibility. Actual devices pass.                                       |
| G1: operational qualification           | Operations/quality lead; all preceding packages | Load/soak/fault results, restore rehearsal, support tools, budgets, scanner and notification operations, production canary and signed release checklist. Every release-blocking defect resolved.                                                            |

H1–H3 constitute the first human-team pilot; they do not imply DM/native completion. M1 can pilot before M2. Keep action execution disabled until its independent gate passes. Calendar estimates require named staffing, allocation, infrastructure readiness, existing-code integration cost and evaluation effort; do not infer a delivery date from this package count.

## 3. Capacity model and benchmark envelope

Model **users and connections separately**. A user can have multiple tabs or devices. The baseline assumes 1.5 active connections per concurrent user: **1,500 SSE connections**. Stress assumes three: **3,000 connections**. Neither ratio is confirmed customer behaviour. Include inactive-but-connected tabs and ensure hidden tabs do not create independent aggressive polling loops.

Use an initial synthetic dataset of 20 workspaces, 200 team rooms, 1,000 active users and five million retained messages, with skewed room sizes, long threads, edits, deletions and attachment references. These are test assumptions. Include a single workspace containing all 1,000 users; equal-sized tenants alone hide hot-workspace contention. Generate private-room/DM data for negative tests without copying production content.

| Workload              | Proposed input assumptions                                                                                                                                          | What must be measured                                                                                         |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| Baseline conversation | 1,500 connections; 20 durable sends/second distributed across rooms; additional edits/reactions at 20% of send rate; 100 active typers updating every three seconds | Send/sync latency, authorisation cost, relay lag, database pool wait and connection memory.                   |
| Burst and hot room    | 3,000 connections; 100 sends/second for five minutes, including 20/second in one room with 300 viewers; separately, a room with 1,000 viewers                       | Room lock contention, fanout fairness, sync coalescing, client rendering and bounded queues.                  |
| Recovery storm        | Disconnect/reconnect 1,000 users over 30 seconds; each resumes up to 1,000 events in bounded pages                                                                  | Admission/backoff, replay correctness, database saturation and recovery time without duplicate sends.         |
| Lost final hint       | Suppress the final room hint with no subsequent message                                                                                                             | Foreground reconciliation discovers the change within the declared fallback interval.                         |
| Mixed work            | Baseline chat plus 10 searches/second, 10 concurrent uploads and a separately controlled Maya workload                                                              | Human-chat isolation from indexing, storage, scanner and model activity.                                      |
| Endurance             | Baseline for eight hours; periodic workspace switches, reconnects, media failures and permission changes                                                            | Memory/goroutine leaks, stale subscriptions, lease recovery, autovacuum/index behaviour and telemetry volume. |

At 1,000 foreground users, one visible-room reconciliation every 30 seconds produces about 33 requests/second; the proposed uniform 25–30-second jitter averages about 36 requests/second and preserves the 30-second maximum interval. Do not poll every room. Account separately for normal hint-triggered sync, read updates, typing and notification traffic.

Run a protocol load generator for connection/throughput tests and a smaller representative set of real browsers/devices for rendering and lifecycle measurements. Declare hardware, replica count, database size/settings, Redis limits, proxy configuration, network latency and client mix. Run an isolated backend profile and a representative higher-latency mobile profile; report both. Avoid artificial production load without a separately authorised exercise.

No Kafka, database sharding or large gateway fleet is required by the confirmed target alone. First measure shared-subscription fanout, SQL query plans, connection memory and per-room lock wait. Extract or partition components only after a measured limit and migration plan exist.

## 4. Reliability objectives and measurements

The following are proposed launch objectives under the declared healthy-network baseline. Product and operations must approve them after the spike. Use a rolling 30-day service window and per-deployment load reports; do not average away slow tenants or hot rooms.

| Journey                        | SLI and proposed objective                                                                                                                                                                                 | Accountable role   |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| Human send/read availability   | Successful eligible requests / eligible requests: ≥99.9%. Exclude malformed/forbidden requests, not unexpected overload or internal failures.                                                              | Backend/operations |
| Durable send                   | Ingress to committed response: p95 ≤300 ms, p99 ≤1 second. Report client-to-ack separately.                                                                                                                | Backend            |
| Visible delivery               | Commit to canonical render in an online foreground recipient: p95 ≤500 ms, p99 ≤2 seconds; includes hint batching and sync round trip.                                                                     | Realtime/client    |
| Replay                         | Reconcile 1,000 events: p95 ≤2 seconds under the declared test network; larger backlogs report progress.                                                                                                   | Backend/client     |
| Silent hint loss               | Active conversation refreshed within 35 seconds under healthy fallback conditions: 30-second schedule plus bounded request allowance.                                                                      | Realtime/client    |
| Maya acknowledgement           | Visible durable queued/running state within 500 ms of accepted invocation when the client is online. Model completion is a separate SLI.                                                                   | AI/client          |
| Notification dispatch          | Eligible attention item reaches provider submission attempt within five seconds at p95 after its grace/quiet-hours delay expires. Measure source-event-to-attempt separately, including deliberate delays. | Notifications      |
| Privacy and mutation integrity | Zero unauthorised disclosure or duplicate committed action in qualification tests; any production occurrence is an incident, not an acceptable percentage.                                                 | Security/backend   |

Measure delivery using consenting test clients or privacy-safe sampled client acknowledgements of rendering. A Redis publish, SSE write, push ticket or provider receipt is not a human read. Display device-delivery/read claims only when the protocol actually establishes them.

Document eligibility denominators and error classes before dashboards ship. Do not conceal throttled in-budget customers by classifying all 429s as success. Track intentional limit enforcement separately. This request-based 99.9% objective permits a 0.1% eligible-request failure budget; it cannot be converted into downtime minutes without a separately defined time-based availability measure. Production incidents consume the relevant budget even if unit tests passed.

## 5. Independent AI capacity and cost controls

The 1,000-user target does not imply 1,000 simultaneous model calls. Start qualification with an assumed one invocation/second and a five/second, 60-second burst. If average active run time is 20 seconds, the baseline implies approximately 20 concurrent runs before headroom. Measure actual duration, tool time and provider rate limits before setting worker concurrency.

Use separate queues/pools for human delivery, interactive Maya, notification dispatch and slower extraction/indexing. Apply workspace and requester admission, fair scheduling, run/time/tool/token budgets, bounded retries and a global provider circuit breaker. Saturated AI displays a durable queued state or explicit capacity rejection while human messages continue normally. Stream previews must not hold a room lock or consume an unbounded token-event log.

Track queue age separately from inference duration and action execution. Proposed operating triggers are p95 interactive queue wait above five seconds and oldest eligible run above 30 seconds for a sustained interval. These are tuning inputs, not provider promises. Before enlarging concurrency, check token/request quotas, database/tool saturation and tenant fairness.

Budget with measured quantities:

- Daily model cost = runs × mean input tokens × input price + runs × mean output tokens × output price, using provider billing units and separate cached/reasoning/tool charges where applicable.
- Retained storage = daily messages × measured stored bytes × retention days × measured index/replica overhead, plus attachments, revisions, events and backups.
- Delivery traffic = event/hint bytes × actual eligible online recipients, plus sync responses, reconnects and previews; do not equate sends with delivered bytes.
- Capacity = arrival rate × mean service time, then add measured burst and failure headroom. Service times differ for summaries, retrieval and multi-action runs.

Record model/version and pricing date in the budget sheet at implementation time. Set workspace spend ceilings, warning thresholds and graceful budget-exhaustion behaviour; no secret switch to a less-tested action model. Cost dashboards use bounded dimensions, not message text.

## 6. Verification gates

### Persistence and communication

Real PostgreSQL/Redis tests must cover concurrent room writes, rollback, identical and conflicting retries, ambiguous HTTP responses, outbox claim fencing, crash after publish, expired leases and consumer dedupe. Prove per-room event commit order, separate root/reply positions and no global-sequence assumption. Test snapshot/subscription races, overlapping sync pages, out-of-order hints, replay expiry, lost final hint, server drain and Redis subscription restart.

Read tests cover multiple devices, background tabs, partial visibility, search jumps, manual unread and independent thread reads. Edit/delete tests preserve local conflict text, thread tombstones and source provenance. A generic idempotency receipt expiring must not make an old client-outbox send duplicate its committed message.

### Privacy, security and lifecycle

Maintain a matrix of active/removed/deactivated actors across two workspaces, private/public teams and DMs. Check messages, members, counts, snippets, replay, attachment IDs, previews, citations, exports and notification excerpts. Remove access while a stream, upload, search, model run and action are active. No unrestricted workspace broadcast may expose private room IDs or activity.

Adversarial cases include malicious rich text, mention spoofing, quoted instructions, prompt injection in files/links, decompression bombs, infected files, SSRF, cross-tenant attachment reuse and stale signed links. Verify deletion propagation through revisions, search, derived summaries, AI evidence, pending delivery and object references. Restore testing must reapply deletion/access tombstones before recovered data becomes visible.

### Maya and actions

Freeze evaluation cases before tuning, maintain a held-out set and run model/prompt/tool changes against it. Include disagreement, sarcasm, absent owners/dates, duplicate work, exact last-ten selection, permission changes, stale citations and malicious source instructions. A readable summary is insufficient if it changes a tentative idea into a decision.

Action tests interrupt every boundary: before mutation, after commit before receipt response, between batch items and before publication. Retry remaining items using stable action IDs. Test simultaneous ticket claims, cancellation after partial completion and model failure after successful creation. Execute domain changes under the initiating human's current authority; Maya's system identity/provenance must not bypass normal story-service validation. Verify actor, initiating human, timestamp, before/after and source links in visible audit output.

### Clients, notifications and accessibility

Use real two-user journeys across supported web browsers, Safari, hosted desktop and physical iOS/Android devices. Verify background/suspended/terminated launches, OS permission denial, rotated/invalid push tokens, expired sessions, account/workspace switches, offline restart, duplicate taps and old installed versions. Persist pending content with account/workspace isolation; never send it under the next signed-in account.

Exercise VoiceOver and NVDA, keyboard-only use, IME, touch, reduced motion, 200–400% zoom, contrast and long/localised content. Verify focus restoration, virtualised-feed navigation, upload controls and restrained live announcements. Real device acceptance and accessibility review are release gates, not follow-up polish.

Notification evidence distinguishes eligibility, queued, provider accepted, receipt/error and app-open events. Implement and test chat-specific receipt reconciliation where the provider supports it; the current repository does not prove this complete. Test one coherent Maya result notification, mute/quiet-hours, cross-device suppression and deep links to removed content. Keep routine email aligned with the established weekly preference unless the user explicitly changes it.

## 7. Migrations and mixed-version deployment

The audited checkout contains concurrent uncommitted story-comment collaboration and attachment work, including migrations **000219/000220**, generated SQLC, desktop notification and client changes. Preserve it. Reconcile its final state before assigning chat migration numbers or extracting shared primitives; do not rewrite applied migrations or assume these changes are deployed.

Follow the repository migration manifest and operations contract. Each chat migration requires ownership, classification, rollout mode, API/worker compatibility, bounded backfill/index strategy and recovery instructions. Use additive schema-first deployment with chat flags off. Backfill canonical team rooms idempotently in bounded batches; concurrent team creation must use the same uniqueness invariant.

Deploy readers that tolerate the additive schema, then compatible workers/API, then clients, then enable writes. Version message bodies/events and negotiate installed-client capability. Test the oldest supported client against the newest compatible API. Unknown optional events request a safe refresh; incompatible required changes need a version gate. Do not allow old workers to claim tasks whose payload or fencing semantics they cannot honor.

Once chat data exists, rollback normally means flags and compatible binaries, not destructive down migrations. Preserve messages, AI action receipts and deletion ledgers. A forward fix is mandatory where downgrading would discard acknowledged data or remove an authorisation fence.

## 8. Phased exposure and rollback

Use server-enforced, workspace-scoped switches for chat reads/admission, live hints, media, search, Maya reads, Maya actions, DMs and external notifications. Flags must not bypass permissions. Maintain a canary allowlist and an audited record of changes.

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

General availability requires named owners, approved semantics, full product acceptance journeys, real database/fault tests, AI quality/security gates, accessible web/native flows, migration/upgrade rehearsal, scanner operations, notification reconciliation, retention/deletion verification, measured 1,000-user qualification, restore evidence and healthy production canary results. Support must be able to investigate a failed send/run from a safe reference ID without reading unrelated conversations.

Record known limitations explicitly and restrict the corresponding capability until its gate passes. A partial pilot can be useful, but screenshots, compilation, happy-path model demos and successful deployment do not establish the complete feature.
