# Background actor boundaries

Queue handlers and Redis event consumers do not inherit HTTP authentication.
Do not make ordinary story reads accept a missing actor, or add a global system
actor to every task. Use an explicit capability at the composition boundary.

- Story update notifications resolve only the title using the event's actor,
  story, and workspace IDs. SQL checks the actor's current membership or active
  system identity. Notification insertion separately authorizes the recipient.
  Database errors propagate for retry; deleted or inaccessible stories are skipped.
- Maya uses the same story adapter in the API and worker. Interactive reads retain
  the requesting actor. Background snapshots use the configured Maya identity;
  approved automation writes explicitly use Maya while retaining team restrictions.
  Assignment planning has a separate system workload read limited to one team.
- GitHub background reads use the configured GitHub identity. Webhook actions can
  retain a mapped human author while preserving the transport actor's scopes and
  team restrictions. An ordinary caller cannot substitute another author.
- Story mutations bind their validated actor before reading status metadata or
  other supporting data. System status and parent-comment reads use explicit
  system queries rather than human membership queries.
- Provider comment creation requires an active system user, the selected
  workspace, and comment write scope. System comments cannot synthesize mentions;
  comment update and deletion retain their existing user authorization.

The database regression test `TestBackgroundStoryOperationsUseExplicitActors`
exercises context-free assignment rules through inbox persistence, background
status updates, provider replies, team workload, and revoked/foreign identities.
Run it against an isolated database using the repository's integration testkit:

```sh
TEST_DATABASE_URL='<isolated local PostgreSQL URL>' go test -tags=integration \
  ./internal/modules/stories/repository \
  ./internal/modules/comments/repository \
  ./internal/modules/notifications/repository -count=1 -parallel=4
```

Deploy both API and worker binaries for these changes. The API owns the Redis
notification consumer. No schema migration or new environment variable is needed.
Existing archived tasks and previously missed notifications are not replayed by
this code change. Inspect current task errors before retrying archived work.
