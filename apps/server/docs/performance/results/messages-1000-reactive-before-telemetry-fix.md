# Reactive workload failed acceptance

The 2026-10-17 10:36 UTC run held 1,000 authenticated SSE connections for five
minutes after a 30-second ramp. Two hundred users additionally replayed the
Messages screen's 150ms coalescing and four HTTP reads per refresh; every team
was split across both API instances. Builds and the development watcher were
paused. PostgreSQL and Redis were disposable loopback processes and were removed.

The full report is `messages-1000-reactive-before-telemetry-fix.json`. This is a
failed result, despite healthy connection survival and successful final fanout,
outbox retry, idempotency, mutations and cursor catch-up probes.

| Observation | Result |
| --- | ---: |
| Live SSE clients at end | 1,000 |
| Periodic history reads offered | 6,000 / 20 per second |
| Reactive HTTP reads | 112,872 / 376.24 per second |
| Message-create requests | 1,573; 380 errors |
| Reaction requests | 300; 19 errors |
| Reactive-read transport errors | 18 |
| Create p95 / p99 | 2,801.9 / 3,494.3 ms |
| Periodic history p95 / p99 | 1,408.3 / 1,598.5 ms |
| Reactive endpoint p95 range | 1,040.5–1,341.7 ms |
| Steady SSE delivery p95 / p99 | 1,725.9 / 2,400.0 ms |
| Peak oldest pending outbox age | 2,331 ms |
| Peak pending outbox depth | 11 |
| Recovery after deliberate worker pause/publication failure | 7,001 ms |

The fixed gates require zero unexpected errors and HTTP p95 below 1,000ms /
p99 below 2,000ms for every operation. They were not relaxed. Error samples are
capped at 100 in JSON; full operation error counts remain available.

A lightweight `pg_stat_activity` snapshot during held traffic found six tuple
waits and three transaction waits in `RecordWorkspaceAccess`, three transaction
waits in `TrackBrowserSession`, and five tuple/transaction waits in
`SessionAccessState`. A second snapshot found 52 backend connections and zero
deadlocks/conflicts (retained in the adjacent database diagnostic file).

Source inspection established that every HTTP access updated the same workspace
activity row, while session authorization held shared locks on that workspace.
Concurrent reactive reads amplified those conflicting telemetry writes. The
session tracking upsert also wrote even when its CASE kept `last_seen_at`
unchanged. These are concrete contention sources to correct while preserving
fresh authorization checks; the snapshot alone does not attribute every error.
This run discarded private application logs, so a bounded error-log capture was
added to the harness for the next run to preserve any remaining server error.

The exact same users, active fraction, offered workload and gates must be rerun
after the telemetry fix. The earlier 20-read/s-only pass is not a substitute for
passing this reactive workload.
