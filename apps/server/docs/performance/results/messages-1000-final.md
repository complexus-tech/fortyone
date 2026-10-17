# Final local Messages release load test

The five-minute representative run and the separate synchronized-start
diagnostic both passed their unchanged acceptance budgets on October 17, 2026.
This is local evidence for a controlled pilot, not a production capacity claim.

## Exact candidate and environment

Both JSON reports identify the same **uncommitted working-tree candidate**:

- Git HEAD: `524cef6cd62401d3b87b7a452262c5966789a7b8`; `working_tree_dirty=true`.
- Runtime/migration SHA256: `b0cdb1a912e05eddae09a3708192855d57968a30f755dd3f1bb688ddcebaf664` (2,774 files).
- Harness SHA256: `8878337903b05aa8c53167a766855f22b1ce1da6c17afec6828388045f82fae8` (four files).

The [runner](../../../scripts/loadtest/run-chat-load.py) documents and computes the
filename/content fingerprint before execution. HEAD alone does not contain these
changes. The test ran natively on macOS/arm64, Go 1.26.9, 11 reported CPUs and 18GiB
host memory. PostgreSQL 18 and Redis were separate, disposable loopback processes;
their data was removed afterward. Existing application databases were untouched.
Two API servers/hubs and the generator share one Go process. Linux/amd64 production
images were built and smoke-tested separately; this did not load-test those images.

No competing release builds or development rebuild loop ran during measurement.
This was still a developer laptop with normal operating-system background work.

## Workload and acceptance

There were **1,000 distinct authenticated connected users**, ten teams with 100
members each, and **200 active Messages viewers**. The other 800 represented
connected platform users. Every team had recipients on both API replicas.
Each team started with 60 short text messages; this is not a deep-history storage
benchmark or a 1,000-simultaneous-active-chat-viewer test.

The representative run ramped connections and initial page reads over 30 seconds,
then held traffic for 300 seconds. During the hold it completed:

| Work | Completed | Achieved rate |
| --- | ---: | ---: |
| Periodic history reads | 5,999 | 19.997/s |
| Top-level messages | 1,499 | 4.997/s |
| Thread replies | 59 | 0.197/s |
| Reactions | 299 | 0.997/s |
| Reactive HTTP reads | 137,840 | 459.467/s |

Reactive reads follow actual authorized SSE hints and the app's 150ms coalescing:
conversation list, conversation detail, first history page, and latest page.
There were 138,640 total reactive reads including 800 initial-ramp reads, and
146,620 total ordinary HTTP requests including post-load probes. All 1,101 SSE
connections succeeded (initial users, 100 cross-replica reconnects, and catch-up).
The 50ms workload scheduler produced 5,999 rather than 6,000 ticks; the predeclared
minimum remained 19 history requests/s (95% of target), and no HTTP failures were
retried by the generator.

Other predeclared gates were zero unexpected errors, all 1,000 streams live after
ramp/reconnect/hold, each HTTP operation p95 under 1s and p99 under 2s, steady SSE
p95 under 2s and p99 under 5s, complete final recipient probes, and durable outbox
recovery within 30s. Every gate passed. Application error capture was also empty.

## Measured latency and resources

From [the five-minute JSON](messages-1000-final-ramped.json), recorded at
`2026-10-17T11:14:03Z`:

| Operation | p95 | p99 |
| --- | ---: | ---: |
| Create message | 150.3ms | 262.6ms |
| Periodic history read | 117.2ms | 213.0ms |
| Reaction | 101.0ms | 182.2ms |
| Reactive conversation list | 166.9ms | 255.0ms |
| Reactive conversation detail | 192.8ms | 297.2ms |
| Reactive history | 220.2ms | 333.8ms |
| Reactive latest page | 220.8ms | 337.3ms |
| SSE connection | 15.6ms | 19.0ms |
| Steady message-to-SSE delivery | 861.9ms | 1,076.7ms |

All 1,000 streams remained live. There were 157,098 observed message-recipient
deliveries, including 155,800 from held traffic. A final probe reached every one
of the 1,000 authorized recipients across the two APIs. The create operation
recorded 1,571 successes, including 13 post-load probe messages; the separate
idempotent create/replay checks also passed.

Held-workload sampled peaks were 1,073 Redis connections, 27,565,872 bytes of Redis
memory, 240,267,824 bytes of combined Go heap, and 11,419 goroutines. Heap includes
the generator and both API replicas, not their individual RSS. Peak pending outbox
depth was ten and oldest pending age was 1,208ms.

After deliberately pausing dispatch and failing one real publication, the durable
event retried and drained in 9,000ms. The overall delivery distribution, including
that intentional fault, was p95 862.4ms / p99 1,078.5ms / max 10,010.8ms; steady
max was 1,324.5ms. Cursor catch-up for a disconnected client, non-overlapping history
pages, thread reads, edits, deletion, read state and creation-id replay all passed.
This fault probe pauses processing; it is not an operating-system worker crash.

## Separate synchronized-start diagnostic

[The final diagnostic JSON](messages-1000-final-synchronized.json), recorded at
`2026-10-17T11:16:01Z`, used the identical source fingerprints and thresholds.
After ramping 1,000 SSE users, it mounted all 200 active viewers simultaneously
(800 initial GETs) and held traffic for 30 seconds. It passed with zero resets,
HTTP errors, stream errors or application errors, and all 1,000 streams live.

It completed 600 periodic reads, 150 top-level messages, six replies, 30 reactions,
and 14,000 reactive GETs (466.7/s). Create p95/p99 was 264.3/448.3ms, history
225.9/496.5ms, and steady SSE 809.4/1,026.5ms. The highest reactive-read p99 was
818.6ms. Its deliberate outbox fault recovered in 7,301ms.

This short passing burst does not replace the five-minute run or prove reliable
reconnect storms through the deployed proxy. Earlier burst failures remain in
the evidence set; no retry or threshold change concealed them.

## Findings fixed and retained failures

- Three Redis Pub/Sub sockets per browser became one subscription containing the
  same three authorized channels. Subscription readiness, workspace/recipient
  filtering, typing/notification semantics and cleanup have regression coverage.
- Slow SSE consumers/writes now close within bounded deadlines. A Redis automatic
  resubscription closes SSE so the client's normal reconnect reconciles its caches.
- Per-request workspace/user/session activity writes caused global row-lock
  contention. Age-gated telemetry now avoids those writes while permission and
  revocation checks remain authoritative; concurrent and revocation tests passed.
- Classified burst logging proved PostgreSQL `40001` was exhausting three immediate
  retries. Bounded, cancelable jittered retries preserve Serializable isolation and
  atomic receipt/message/outbox commits. Focused unit and PostgreSQL tests cover
  rollback, commit failures, cancellation and exactly-once after-commit callbacks.

The [failed pre-telemetry run](messages-1000-reactive-before-telemetry-fix.md), both
post-telemetry synchronized diagnostics, the short ramped pass, and the earlier
build-contaminated diagnostic are retained. The first full reactive run had 380
create errors, 19 reaction errors and 18 transport resets; create p95/p99 was
2,802/3,494ms. Later logging separated serialization failures from TCP resets.
The host listener backlog was 128, but the earlier reset cause was not conclusively
isolated. Neither error class appeared in the final measured runs.

## Production boundaries

Before enabling the pilot, verify actual proxy/TLS streaming and idle deadlines,
Redis connection quotas (roughly 1,100 connections for this single-tab workload,
plus reconnect/multiple-tab headroom), API file descriptors, database pools,
worker restarts and queue alerts on the intended staging resources. Repeat with
the real active-chat fraction, room sizes, payloads and retention depth.

These runs do not exercise browser rendering, production replicas in separate
processes, load-balancer failover, process crashes, long network partitions,
uploads/object storage, PDFs/images, audio/transcription, models, or external
email/push delivery. No production migration, deployment or external notification
was performed. A limited workspace rollout and deployment-specific validation
remain required.
