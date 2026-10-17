# Messages load regression with workspace Features settings

The final five-minute local regression passed on October 17, 2026 after Chat
rollout moved to the default-on `workspace_settings.chat_enabled` setting and
settings saves gained an optimistic version guard. The unchanged acceptance
budgets passed with **1,000 authenticated SSE users, including 200 active Messages
viewers**. This is local workload evidence, not a production capacity guarantee.

## Candidate and environment

The [raw result](messages-1000-workspace-features-ramped.json), recorded at
`2026-10-17T11:49:57Z`, identifies this uncommitted working-tree candidate:

- Git HEAD: `524cef6cd62401d3b87b7a452262c5966789a7b8`; `working_tree_dirty=true`.
- Runtime/migration SHA256: `9652b9daa10e9f5efc056a2dc5392fc463988c910144fa7fc934955099acf148` (2,775 files).
- Harness SHA256: `5311e862a7c05587031aa4abc427264dad2b754bfd6739074fe9a6820bb4d79a` (five files).

Both fingerprints were recomputed after the run and still matched. HEAD alone
does not contain these changes. The [runner](../../../scripts/loadtest/run-chat-load.py)
records the exact filename/content fingerprint algorithm and scope in the JSON.

The test ran natively on macOS/arm64 with Go 1.26.9, 11 reported CPUs and 18GiB
host memory. It created disposable PostgreSQL 18 and Redis processes bound to
loopback, applied the real migration chain including migration 234, and removed
their data and processes afterward. Existing application databases were untouched.
Two actual API servers, their independent PostgreSQL pools and SSE hubs, and the
load generator shared one Go process. This was not a load test of Linux images.
The development reload watcher was paused, the local scanner stopped, and no
release builds ran during measurement. Normal laptop background work remained.

## Workload and unchanged acceptance

Ten teams each contained 100 users and started with 60 short text messages. All
users had real authenticated sessions and live authorization checks; each team
had recipients on both API replicas. Connections and the active viewers' initial
reads ramped over 30 seconds, followed by a 300-second hold. The other 800 users
represented connected platform users outside Messages.

| Work during the hold | Completed | Achieved rate |
| --- | ---: | ---: |
| Periodic history reads | 5,999 | 19.997/s |
| Top-level messages | 1,499 | 4.997/s |
| Thread replies | 59 | 0.197/s |
| Reactions | 299 | 0.997/s |
| Reactive HTTP reads | 136,720 | 455.733/s |

Reactive reads follow actual SSE hints with the app's 150ms coalescing. Each cycle
reads conversation list, conversation detail, history and latest messages. There
were 137,520 reactive reads overall, including 800 initial-ramp requests, and
145,510 ordinary HTTP requests including post-load checks. One hundred users
reconnected to the other replica halfway through. All 1,101 SSE connections
succeeded, including those reconnects and the later cursor catch-up probe.

The fixed gates were all 1,000 streams live after ramp/reconnect/hold, zero
unexpected errors, at least 19 periodic history reads/s, each HTTP operation p95
below 1s and p99 below 2s, steady SSE p95 below 2s and p99 below 5s, complete final
recipient probes, and durable outbox recovery within 30s. Every gate passed.
The generator did not retry HTTP failures. Intentional disabled-feature HTTP 403
responses were checked explicitly under separate operation names; any other
response would fail those probes. Application error capture was empty.

## Latency, delivery and resources

| Operation | p95 | p99 |
| --- | ---: | ---: |
| Create message | 200.0ms | 441.1ms |
| Periodic history read | 127.5ms | 283.9ms |
| Reaction | 133.8ms | 229.4ms |
| Reactive conversation list | 189.7ms | 349.2ms |
| Reactive conversation detail | 219.0ms | 388.0ms |
| Reactive history | 250.3ms | 452.9ms |
| Reactive latest page | 251.2ms | 453.9ms |
| SSE connection | 15.6ms | 18.0ms |
| Steady message-to-SSE delivery | 854.3ms | 1,102.4ms |

All 1,000 streams remained connected. There were 157,098 observed message-recipient
deliveries, including 155,800 during held traffic. Every one of the 1,000 final
recipient probes passed. The create operation recorded 1,571 successful requests,
including 13 post-load probes; a separate idempotent create/replay pair also passed.

Sampled held-workload peaks were 1,073 Redis clients, 27,539,376 bytes of Redis
memory, 239,248,536 bytes of combined Go heap, and 11,786 goroutines. Heap includes
the generator and both APIs, not per-replica RSS. The pending outbox peaked at seven
events and its oldest sampled age was 1,148ms.

After deliberately pausing dispatch and failing one Redis publication, the durable
event retried and drained in 7,500ms. The overall delivery distribution including
that deliberate fault was p95 854.5ms / p99 1,126.4ms / max 8,433.3ms; steady max
was 1,346.9ms. Cursor catch-up for a disconnected client, non-overlapping history
pages, thread reads, edits, deletion, read state and creation-id replay passed.
The fault probe pauses processing; it does not simulate an operating-system crash.

## Live workspace setting regression

After the held traffic and delivery checks, the harness changed only the
disposable workspace's database setting. Both already-running API replicas then:

- Reported disabled capabilities and rejected message reads and creates with 403.
- Immediately reported enabled capabilities and allowed history reads after the
  setting was restored, without a process restart.
- Preserved all 1,000 shared platform SSE connections across both setting changes.

These probes use direct setting writes restricted to the disposable database.
They validate live reads and API enforcement, not the administrative settings
endpoint, settings-change broadcasts or optimistic-save conflicts. Those paths,
durable worker filtering and existing-stream event gating have separate focused
backend regressions. The root task separately verified the Features toggle in
the local browser after applying migration 234, and left Chat enabled. The local
development watcher and scanner were restored after this measurement.

## Boundaries and historical evidence

This result does not establish capacity for 1,000 simultaneously active Messages
viewers, larger teams, deep histories, richer payloads, attachments, voice/AI jobs,
browser rendering, external notifications, production TLS/proxies or independent
process failover. Verify managed Redis connection quotas, API descriptors, database
pools, streaming deadlines, worker restarts and queue alerts on staging with the
intended resources and realistic active-viewer fraction. This workload needed
roughly 1,100 Redis connections before multiple-tab/reconnect headroom.

The earlier [static-rollout candidate results](messages-1000-final.md), including
its separate synchronized-start diagnostic, remain unchanged as historical
evidence. That diagnostic predates this database-setting implementation; no new
synchronized-start measurement was performed for this candidate. Earlier failed
and build-contaminated reports also remain preserved. No budgets were relaxed to
pass this regression, and no production migration or deployment was performed.
