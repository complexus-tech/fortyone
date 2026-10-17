# Diagnostic run excluded from capacity acceptance

On 2026-10-17, a diagnostic with 1,000 users, a 30-second ramp and a planned
300-second hold ran while a Next production build, a Docker linux/amd64 Go build,
and repeated local Air builds/linkers competed for the same 18GB M3 Pro host.
The parent investigation found and stopped 16 duplicate Air build/link processes.

This diagnostic used the old three-socket-per-client SSE hub and a 100ms direct
outbox-dispatch loop. It is not the final Asynq-path acceptance test.

The run failed after 343.35 seconds. Redis reported many loopback Pub/Sub write
timeouts around 10:09:28–10:09:33 UTC, and the held-load assertion reported:

```text
expected: 1000
actual:   565
all streams remain connected after held traffic/reconnect
```

The test stopped at that assertion, before writing its final JSON metrics.
Temporary PostgreSQL/Redis processes and data were removed by the runner.
The failure remains recorded here; it must not be presented as a pass or silently
removed from the test history. Build contention is a confounding factor, not
proof that the application was fault-free. A controlled rerun is required.

The preceding 30-second diagnostic passed and is retained separately in
`messages-1000-baseline-30s.json`; its fixed-loop timing is also not the final
production-queue capacity result. The final harness writes partial reports on
failure, uses actual Asynq wakeups and the five-second recovery sweep, and tests
the hardened one-socket hub after competing builds have finished.
