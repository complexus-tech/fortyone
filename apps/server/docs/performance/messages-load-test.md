# Messages concurrent-user release test

The latest five-minute run passed on October 17, 2026 with Chat controlled by the
default-on workspace Features setting. See [current results, candidate fingerprints
and limitations](results/messages-1000-workspace-features.md). The earlier
[static-rollout results and synchronized-start diagnostic](results/messages-1000-final.md)
remain historical evidence. Earlier failures remain below and in the results directory.

Run the test from the repository root:

```sh
python3 apps/server/scripts/loadtest/run-chat-load.py \
  --users 1000 --ramp-seconds 30 --hold-seconds 300 --active-chat-percent 20 \
  --report /tmp/fortyone-chat-load.json
```

The runner needs Go, Python 3, Redis, and PostgreSQL 18 binaries. Pass `--pg-bin`
when PostgreSQL is installed somewhere other than the macOS Homebrew default.
It creates a separate temporary PostgreSQL cluster and Redis process bound only
to loopback, applies the real migration chain to a randomly named test database,
and removes both processes and their data on exit. It does not use the running
application's database, Redis, credentials, workspace, uploads, or users. The Go
test refuses non-loopback URLs and a non-empty Redis database; use the runner,
not an existing shared Redis instance. Redis Pub/Sub is not database-scoped.

A quick harness check is available with `--users 20 --hold-seconds 10
--ramp-seconds 1`. It is not a 1,000-user result.

## Executed system and workload

The test creates 1,000 distinct users and opaque browser sessions with real
Redis session records, PostgreSQL account version checks, live workspace
membership, session policy/tracking, SSO-policy checks, and team membership.
Users belong to ten teams with 100 members each and 60 initial messages per team.

Two independent HTTP servers, PostgreSQL connection pools (24 connections each),
and SSE hubs run the real production routes and middleware in one Go process.
An additional eight-connection PostgreSQL pool serves test setup and delivery.
The services/repositories are not replaced by fake message implementations.

The held workload targets:

- 1,000 simultaneous authenticated SSE connections, after a 30-second ramp.
- 20 history reads per second, distributed over all 1,000 users.
- 200 users additionally model an open Messages screen, with the app's 150ms
  invalidation coalescing. Each refresh makes four real HTTP reads: conversation
  list, conversation details, first history page and latest page (50 messages).
  Reads are triggered by actual authorized SSE hints, including reactions and
  replies. The other 800 users represent connected users elsewhere in the app.
  This is not a claim that all 1,000 users constantly have Messages open. The
  active fraction is configurable; test it again for a different workload.
  Each active user's initial reads mount during that user's connection ramp;
  reopening a stream also reconciles. Total reactive counts include that initial
  ramp, while `held_reactive_http_reads` and its rate isolate the timed hold.
  Latency/error gates include both phases. `--synchronized-chat-start` reproduces
  the separate diagnostic that mounts all 200 viewers together after the ramp
  (800 initial GET connections); no transport errors are retried away.
- Five top-level message creations per second, distributed over teams/users.
  Each message is short plain text (a fixed label plus a UUID), with no attachment
  or model invocation; richer payload/attachment workloads need separate sizing.
- One reaction per second and one thread reply every five seconds.
- 100 users disconnect and reconnect to the other API instance halfway through.
- Actual task enqueueing after commit, two Asynq worker instances with two
  concurrent tasks each, the production dispatcher, and the five-second recovery
  sweep. PostgreSQL owns the durable outbox; Redis carries the invalidation hints.
- Real notification persistence and queueing. Email/push jobs go to an unconsumed
  `load-external-disabled` queue. No external notification, AI, transcription, or
  object-store worker runs.

Post-load checks cover all-recipient cross-replica delivery, non-overlapping
cursor history, edit/read/delete, thread retrieval, read state, idempotent
message-creation replay, history catch-up for a message committed while the
client was disconnected, and recovery after temporarily pausing dispatch plus
injecting one Redis publication failure. The failed event must have more than
one attempt and a final `delivered_at` value.

The final probe also disables and re-enables Chat in the disposable workspace's
database settings. Both API replicas must immediately reflect capabilities and
deny/restore message access without restart, while retaining the shared platform
SSE streams. Explicit expected-403 probes do not relax normal request error gates.
These direct test-database writes do not exercise the admin settings endpoint or
settings broadcasts; separate backend and browser checks cover those paths.

## Acceptance assumptions

These are local release-test budgets, not advertised production capacity:

- Exactly 1,000 live connections after ramp and after reconnect/held traffic.
- Zero unexpected HTTP, stream, dispatcher, or generator errors.
- At least 95% of the offered 20 history requests per second must be achieved.
- HTTP p95 below 1 second and p99 below 2 seconds per operation.
  This includes each of the four reactive-read operations.
- During held traffic, SSE delivery p95 below 2 seconds and p99 below 5 seconds.
- A final probe must reach every authorized team member across both replicas.
- Durable outbox drains within 30 seconds after processing resumes.
- Lost SSE hints recover through authoritative cursor history.

Latency uses real response completion and received SSE events; SSE delivery lag
starts at the message's database creation timestamp. JSON includes counts,
percentiles, maximums, errors, active connections, retries, Redis clients/memory,
and sampled Go heap/goroutine peaks. Go heap includes the load generator and both
API instances; it is not per-replica RSS. The full delivery distribution includes
intentional outage/retry probes; `steady_delivery_lag_ms` isolates held traffic.
Resource and queue peaks are sampled during the held workload; the deliberate
worker outage is reported separately through its pending-event/recovery fields.
The runner adds `candidate` metadata with Git HEAD, the working-tree dirty flag,
and separate SHA256 fingerprints for runtime/migration inputs and the harness.
These local changes are uncommitted; the recorded HEAD alone does not contain
the tested implementation. The JSON describes the exact fingerprint algorithm
and file scope so the working-tree candidate can be identified independently.

## Boundaries before production

This is a laptop loopback test, not a production load balancer/CDN/TLS test or a
certified production capacity figure. It does not measure browser rendering,
file uploads, PDF/image work, audio/transcription, AI, email/push delivery,
long network partitions, or rolling deployments. API instances share a process
and Redis client; a process crash/restart was not simulated. It proves route,
authorization, queue, database, outbox and fanout behavior under this workload.

The hardened hub uses one Redis Pub/Sub connection per connected browser tab
for its three authorized channels, plus pooled command and worker connections.
One thousand users therefore need roughly 1,100 Redis connections, with extra
headroom for reconnect overlap and multiple tabs. The previous implementation
used three sockets per tab (roughly 3,100 connections). Subscription setup is
acknowledged before the SSE connection is accepted; a failed subscription or
slow subscriber closes the stream so the normal reconnect/history path can recover. Verify the actual managed Redis connection quota, API file-descriptor
budget, proxy streaming settings and deployed replica capacity before enabling
a workspace. Increasing the user target requires a new measured workload and
likely a shared-subscription hub, not simply changing this test's user count.

After Redis automatically resubscribes, the hub closes the affected SSE stream
so the existing EventSource reconnect handler reconciles its authoritative
caches. Each HTTP event/heartbeat write and flush has a ten-second deadline;
successful writes clear that deadline while the healthy stream is idle. Focused
race tests cover resubscription, blocked socket writes/flushes, recipient/workspace
filtering, typing, notifications and stream lifecycle.

## Earlier controlled run

`results/messages-1000-pre-final-stream-guards.json` passed a five-minute hold
with all 1,000 streams live: 6,000 periodic history reads, 1,573 timed message
creates and 157,298 observed message-recipient deliveries. Create p95/p99 was
39.3/84.8ms, history 13.8/40.9ms, and steady SSE delivery 818.5/1,033.9ms.
Peak Redis usage was 1,073 connections and 27.4MB; combined Go heap was 168.4MB.
This run preceded the final write-deadline/resubscription guards and did not
include reactive UI reads. Keep it as intermediate evidence, not the final
candidate result. The final rerun includes the 20% active-chat workload above.

The first reactive run failed and is retained in
`results/messages-1000-reactive-before-telemetry-fix.json` and its accompanying
analysis. It found real workspace/session telemetry contention under amplified
reads, plus transport resets in the synchronized startup burst. The fixed
HTTP/error budgets were not relaxed. The normal ramped-arrival workload and the
optional synchronized-start diagnostic are distinguished explicitly.

After age-gating workspace/user/session telemetry, a separate 60-second
ramped-arrival diagnostic passed with no errors and all 1,000 clients live
(`results/messages-1000-postfix-ramped-diagnostic.json`). It made 28,080 reactive
reads, including 27,280 during the hold. Create p95/p99 fell to 182/259ms and
history to 135/211ms. This short result does not replace the five-minute gate.

The synchronized-start comparisons remain in
`results/messages-1000-postfix-synchronized-diagnostic.json` and
`results/messages-1000-postfix-synchronized-classified.json`. They still failed:
respectively two/four TCP connection resets and two/one create HTTP 500s. The
classified run correlated the create failure with PostgreSQL SQLSTATE `40001`
(serializable retry exhaustion). The startup resets are distinct from that
database failure; the test does not silently retry either. All 200 viewers
mounting at once opens 800 GET requests on this laptop's loopback listeners;
the normal acceptance scenario mounts viewers during the same 30-second user
arrival ramp. A production proxy/reconnect-storm test remains necessary.

The candidate keeps Serializable isolation and now retries aborted chat units
with bounded, context-cancelable exponential jitter (eight attempts, at most
630ms total backoff). It retries only classified serialization/deadlock failures,
never ambiguous commit/transport failures. Receipts, messages and outbox writes
roll back together before the next attempt; after-commit work runs once. Focused
transaction lifecycle tests and real PostgreSQL concurrent-write/replay tests
cover these invariants. The load generator still never retries HTTP failures.
