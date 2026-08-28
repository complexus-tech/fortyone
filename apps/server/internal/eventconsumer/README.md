# Event consumer

`internal/eventconsumer` owns the API process's Redis Stream projection and
notification side effects. It consumes durable application events, delegates
policy to module services, and acknowledges a stream message only after the
handler succeeds.

## File map

| File                         | Responsibility                                                                            |
| ---------------------------- | ----------------------------------------------------------------------------------------- |
| `runtime.go`                 | Dependencies, stream decoding, dispatch, and deterministic notification dedupe keys       |
| `lifecycle.go`               | Consumer-group initialization, supervised read/reclaim loops, cancellation, and retry     |
| `story_events.go`            | Story notifications, schedule reconciliation, feedback bridging, and workspace broadcasts |
| `objective_events.go`        | Objective and key-result notifications                                                    |
| `feedback_events.go`         | Feedback notification projections                                                         |
| `collaboration_events.go`    | Comment, reply, and mention notifications                                                 |
| `email_events.go`            | Verification and invitation email delivery                                                |
| `workspace_notifications.go` | Workspace deletion and restoration email notifications                                    |

## Invariants

1. The consumer is internal application composition, not a reusable public
   library.
2. Stream payloads are decoded into typed events before dispatch. Unknown event
   types fail explicitly and are not acknowledged as successful.
3. Notification dedupe identity derives from immutable event facts so a
   reclaimed or retried message converges on the same write.
4. Recipient authorization and resource context are loaded through narrow
   module ports; this package does not own notification SQL.
5. Both the new-message and stale-message reclaim loops are supervised and
   joined. Cancellation does not leave a background goroutine behind.
6. Errors and logs contain safe identifiers, never verification tokens, email
   bodies, credentials, or raw provider payloads.

## Verification

```bash
go test -race ./internal/eventconsumer
go vet ./internal/eventconsumer
make architecture-check
```
