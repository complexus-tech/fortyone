# Slack module

The Slack module is a first-party messaging adapter. It translates Slack's
OAuth, signed HTTP requests, conversations, modals, and Work Objects into
FortyOne use cases. It does not define a public plugin ABI, and Slack SDK types
do not leave the provider-facing service files.

Start here when changing Slack:

| Concern                                                          | Location                                                 |
| ---------------------------------------------------------------- | -------------------------------------------------------- |
| Transport routes and JSON mapping                                | `internal/modules/slack/http`                            |
| Transport-neutral records, commands, and errors                  | `internal/modules/slack/domain`                          |
| OAuth, account links, commands, assistant behavior, and delivery | `internal/modules/slack/service`                         |
| PostgreSQL adapter and transactions                              | `internal/modules/slack/repository`                      |
| Generated database methods                                       | `internal/modules/slack/repository/sqlc`                 |
| Typed SQL source                                                 | `internal/modules/slack/repository/queries`              |
| API and worker composition                                       | `internal/bootstrap/api` and `internal/bootstrap/worker` |
| Durable inbox lifecycle                                          | `internal/platform/webhooks`                             |
| Credential encryption                                            | `internal/platform/credentialvault`                      |

Generated SQLC structs are private repository details. HTTP models do not carry
database tags, domain commands do not import database or Slack SDK packages,
and bootstrap is the only place that connects Slack to concrete FortyOne
modules.

## Main flows

### Install a workspace

1. An active workspace administrator asks for an install session.
2. The service stores a single-use opaque OAuth nonce bound to that workspace
   and actor.
3. The callback consumes the nonce, exchanges the code, and validates the Slack
   team identity.
4. The bot credential is sealed in the shared vault with workspace, Slack team,
   credential type, and a fresh installation generation as authenticated data.
5. One repository transaction rechecks the actor, serializes the Slack
   installation lifecycle, rejects workspace/team conflicts, supersedes safe
   pending uninstall work, upserts the installation, links the installer when
   possible, and rebinds provider threads from the old generation.
6. Channel discovery and email-based member linking run after the installation
   commit. Their final writes recheck the same active installation generation.

The OAuth callback cannot make the provider exchange part of a PostgreSQL
transaction. If local installation persistence loses a race, the credential is
placed in the durable uninstall outbox and revoked through its guarded worker
flow instead of issuing an unsafe best-effort provider call.

### Receive an event

```text
Slack exact body
  -> bounded read
  -> signature and timestamp verification
  -> active installation + generation resolution
  -> context-bound payload encryption
  -> durable inbox receipt
  -> queue {provider, inbox_id}
  -> worker lease + generation recheck
  -> actor/team authorization recheck
  -> idempotent product behavior
  -> terminal safe outcome
```

URL verification is the provider-required short synchronous exception. Ordinary
events are acknowledged only after durable persistence. Raw bodies and tokens
never enter the queue. See [Slack webhook security](../security/slack-webhooks.md).

### Commands and interactions

Commands and interactions use Slack's short response deadline. The handler
verifies the exact request first, extracts only bounded fields, and starts the
longer work asynchronously where needed. Before a story, request, comment, or
assistant action is executed, the service resolves the linked FortyOne user and
rechecks current workspace and team access.

Modal private metadata is provider state, not authorization. It can carry a
source reference, but every selected workspace, team, story, objective, sprint,
label, assignee, or request is loaded again through a tenant-scoped capability.

### Deliver a message

Outbound effects use stable idempotency keys and the messaging delivery store.
The worker rechecks the active Slack installation generation immediately before
the provider call. Actor-scoped messages also recheck the actor's current
workspace/team membership; channel-scoped disclosures use the explicitly
configured public-team audience. A disconnect or reinstall cancels stale inbox
and outbound rows in the same installation transaction.

### Disconnect and uninstall

Disconnect is locally atomic:

1. lock and recheck the active administrator and installation;
2. copy the still-encrypted credential into the uninstall outbox;
3. cancel retryable inbound and outbound Slack work;
4. delete account links and the active installation; and
5. commit.

Remote Slack revocation happens after the commit and cannot be atomic with
PostgreSQL. Failure leaves a retryable encrypted outbox row. Claims use bounded
leases, attempt limits, stable ordering, and a terminal
`revocation_required` state for operator resolution. Completion clears the
retained credential.

## Authorization rules

Middleware is a useful early rejection, not the source of truth. Repository SQL
must observe the current database state at the operation boundary.

| Operation                                             | Required live state                                                                              |
| ----------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| Read installation/channels/account link               | active user + current workspace membership                                                       |
| Read logs/settings/audiences                          | active user + current workspace administrator                                                    |
| Install, resync, disconnect, update settings/audience | active user + current workspace administrator in the owning transaction                          |
| Link or unlink own Slack identity                     | active user + current workspace membership + matching actor                                      |
| Provider event processing                             | active installation + exact installation generation + current linked actor/team scope            |
| Channel disclosure                                    | current linked actor plus public, joined, explicitly mapped teams where configured               |
| Credential maintenance                                | exact installation/uninstall identity, generation, envelope version, and original ciphertext CAS |

Unknown, removed, inactive, downgraded, deleted-workspace, mismatched-tenant,
and stale-generation states fail closed. Expected absence maps to the Slack
domain not-found error; authorization denial maps to the domain forbidden error;
lifecycle and compare-and-swap races map to conflict. Raw PostgreSQL and provider
errors are logged safely and never returned as public details.

## Adding Slack behavior

1. Put the stable input/output in `domain` when it crosses a repository or
   transport boundary.
2. Add the smallest caller-owned capability interface to the service. Do not
   import another module's concrete service type.
3. Keep Slack request/response types in a provider adapter file and map them at
   the edge.
4. Add named, static, tenant-scoped SQL in `repository/queries`; do not construct
   SQL fragments or use maps as update intent.
5. Put a multi-statement invariant in one repository-owned pgx transaction.
6. Add unit tests for mapping/validation, HTTP tests for strict parsing/error
   mapping, and PostgreSQL 18 tests for authorization, concurrency, rollback,
   stable order, and tenant isolation.
7. Update this guide and the provider registry when a capability genuinely
   changes. Never advertise an unimplemented provider feature.

## Focused checks

```bash
go test -race -count=1 ./internal/modules/slack/...
go vet ./internal/modules/slack/...
make sqlc-check
make sqlc-vet
```

PostgreSQL tests require the disposable PostgreSQL 18 control database described
in [integration test infrastructure](../testing/integration-infrastructure.md).
