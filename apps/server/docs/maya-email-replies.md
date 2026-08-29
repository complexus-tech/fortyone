# Maya email replies

Maya guidance emails are sent from `Maya, AI Agent <maya@fortyone.app>` and use a
thread-specific reply address such as
`maya+<opaque-token>@reply.fortyone.app`. The token is bound to one user,
workspace, and durable conversation. Raw tokens are not stored in the reply
token table.

## Production setup

### 1. Authenticate the sender

In Brevo, authenticate `fortyone.app` and verify `maya@fortyone.app` as a
transactional sender. Install the SPF, DKIM, and DMARC records Brevo provides.
Do not publish a second SPF record; merge the required include into the
existing record instead.

The Maya sender has built-in defaults. These environment variables are only
overrides:

```text
APP_EMAIL_MAYA_FROM_ADDRESS=maya@fortyone.app
APP_EMAIL_MAYA_FROM_NAME="Maya, AI Agent"
```

### 2. Delegate the inbound reply subdomain

Create these DNS records for the dedicated receiving subdomain:

| Name                 | Type | Priority | Value                      |
| -------------------- | ---- | -------: | -------------------------- |
| `reply.fortyone.app` | MX   |       10 | `inbound1.sendinblue.com.` |
| `reply.fortyone.app` | MX   |       20 | `inbound2.sendinblue.com.` |

Do not add a normal mailbox provider or competing MX record to this subdomain.
It is deliberately separate from the domain used to send email. Brevo's
current setup instructions are in its
[inbound parsing documentation](https://developers.brevo.com/docs/inbound-parse-webhooks).

### 3. Create the Brevo inbound webhook

Configure this webhook:

```json
{
  "type": "inbound",
  "events": ["inboundEmailProcessed"],
  "domain": "reply.fortyone.app",
  "url": "https://api.fortyone.app/webhooks/brevo/inbound-email-processed",
  "headers": [
    {
      "key": "X-FortyOne-Webhook-Token",
      "value": "<derived value>"
    }
  ]
}
```

Generate the header value with the same dedicated
`APP_EMAIL_REPLY_SECURITY_KEY` used by the API and worker:

```sh
cd apps/server
go run ./cmd/brevo-webhook-token
```

The command prints the complete header name and derived value without printing
the application secret. Brevo supports custom authentication headers on
webhooks as described in its
[secured webhook documentation](https://developers.brevo.com/docs/secured-webhooks).

Rotating `APP_EMAIL_REPLY_SECURITY_KEY` also rotates the webhook header and the
key used to encrypt queued email payloads. Drain pending email-reply work and
update the Brevo header as part of the same controlled rotation.

### 4. Existing runtime configuration

Production needs the existing values for Postgres, Redis, SMTP, and OpenAI,
plus the dedicated email-reply key shared by the API and worker:

```text
APP_EMAIL_REPLY_SECURITY_KEY
APP_ENVIRONMENT=production
APP_EMAIL_HOST
APP_EMAIL_PORT
APP_EMAIL_USERNAME
APP_EMAIL_PASSWORD
APP_EMAIL_ENVIRONMENT=production
OPENAI_API_KEY
```

`APP_ENVIRONMENT` is the authoritative deployment mode for security policy.
`APP_EMAIL_ENVIRONMENT` controls email-delivery behavior only. In production,
`APP_EMAIL_REPLY_SECURITY_KEY` must be a unique value of at least 32 bytes. The
API and worker refuse to start with the development default, a short key, or a
value reused by another security capability because it protects both the
webhook capability and queued email payloads.

`OPENAI_MODEL` is optional and defaults to `gpt-5.6-luna`. The worker manages
conversation state in FortyOne and calls the model with provider storage
disabled.

## Runtime boundaries

The ingress route authenticates the derived webhook header before reading the
body. It accepts JSON only, reads at most 5 MiB, allows at most 100 provider
items and 100 mailboxes per item, and gives durable inbox handoff 15 seconds.
Brevo receives `429` plus `Retry-After: 600` when a transient dependency or the
deadline prevents a safe acknowledgement. Malformed permanent input is not
retried.

Only the data needed to authenticate and interpret the current reply enters the
encrypted inbox. HTML, signatures, carbon-copy recipients, attachment download
capabilities, provider display names, and raw quoted history are removed. The
visible reply is limited to 32,000 Unicode code points; canonical inbound and
sealed delivery state are each limited to 256 KiB. Redis tasks contain only the
opaque workspace/thread scope and provider event ID. The worker task itself has
a 45-second deadline, and post-cancellation receipt/delivery bookkeeping gets a
separate five-second bounded context rather than an unbounded detached context.

Authorization is deliberately repeated at each trust transition:

- ingress verifies the derived HMAC capability, opaque reply token, exact
  sender address, thread, workspace, and user binding;
- the worker reloads current workspace membership and current team access after
  acquiring the per-thread lease;
- proposal confirmation reloads the target and validates its current team,
  version, selected status, and selected assignee;
- finite typed objective, key-result, story, and feedback commands adapt to the
  existing domain CAS methods; arbitrary model-produced maps never cross the
  email-reply mutation boundary;
- a frozen retryable delivery records the authorized team IDs and is rejected
  if the actor has lost any required access before the send retry.

## Code map

- `internal/modules/emailreply/http/emailreply.go` owns webhook authentication,
  media-type enforcement, body limits, the ingress deadline, and HTTP mapping.
- `internal/modules/emailreply/service/service.go` owns durable ingress,
  encryption, sender/thread binding, deduplication, and recovery.
- `internal/modules/emailreply/service/context_loader.go` rebuilds current
  workspace, team, target, status, and assignee authorization.
- `internal/modules/emailreply/service/mutation_applier.go` and
  `mutation_ports.go` own confirmed finite mutation commands and domain adapters.
- `processor_agent_ports.go` and `processor_conversation_ports.go` define the
  emailreply-owned decision, summary, rendering, conversation, proposal, and
  delivery contracts. They contain neutral finite types and do not import the
  concrete emailagent or messaging services.
- `processor.go`, `processor_proposals.go`, `processor_delivery.go`, and
  `processor_history.go` separate orchestration, confirmation, delivery, and
  bounded conversation history.
- `internal/bootstrap/worker/email_reply_agent_adapter.go` and
  `email_reply_store_adapter.go` are the only worker composition adapters that
  translate those contracts to emailagent, emailthread, and messaging types.

## Product flow by example

### Update an objective

Maya sends:

> Activation is marked At Risk and has not been updated recently. Has the
> position changed? Tell me the latest health or blocker and I will help you
> keep it current.

The user replies:

> The launch blocker is resolved. Set Activation to On Track.

The worker rechecks the user's current access and reloads the objective. Maya
does not write yet. It sends an exact preview:

> I can set “Activation” from At Risk to On Track and add this check-in:
> “The launch blocker is resolved.” Reply CONFIRM to apply this change, or
> CANCEL to leave it unchanged.

An exact `CONFIRM` reply reauthorizes the user, compares the objective version,
and applies the update through the objective service. Maya then sends a receipt.
An exact `CANCEL` leaves the product unchanged.

### Correct a pending change

If the user replies before confirming:

> Actually, keep it At Risk and change the check-in to “Launch moved to
> Friday.”

Maya creates a new complete preview. The previous pending proposal is marked
superseded, so a later confirmation cannot accidentally apply the old version.

### Continue a longer conversation

Every inbound and outbound turn is stored in sequence. Maya receives the most
recent turns verbatim. Once a thread grows, older turns are incrementally
condensed into a persisted factual summary that preserves corrections,
decisions, commitments, and open proposals. The original messages remain in
the durable history; the summary is only the bounded model context.

Responses are rendered from typed email blocks into both HTML and plain-text
MIME alternatives. Model output is never treated as raw HTML and is never sent
as Markdown.

### Stale or unauthorized changes

If the entity changes after Maya's preview, confirmation does not overwrite the
newer state. Maya explains that the preview is stale and asks for the latest
desired state. If the user no longer has access, no entity details are exposed
and no action is applied.

## Supported email actions in v1

- Set one objective's health, optionally with the user's check-in text.
- Set one key result's current value, optionally with the user's check-in text.
- Change one task's due date, status, and/or assignee.
- Change one feedback item's status.

Every write is one entity, previewed first, confirmed explicitly, reauthorized,
and guarded by the entity's last-updated version. Deletes, permission changes,
billing, workspace or team moves, and batch mutations stay in the product.

## Smoke test

1. Apply migration `000118_messaging_email_conversations.up.sql`.
2. Deploy the matching API and worker.
3. Run `dig MX reply.fortyone.app +short` and confirm both Brevo hosts.
4. Trigger a real Maya guidance email to a test user.
5. Verify its From, opaque Reply-To, Message-ID, HTML/plain alternatives, SPF,
   DKIM, and DMARC.
6. Reply from the exact recipient address with a specific update.
7. Confirm Maya sends a preview and the product has not changed.
8. Reply with only `CONFIRM`; confirm the product update and receipt.
9. Repeat after changing the entity in the portal; confirm the stale preview is
   rejected.
10. Test Gmail, Outlook, and Apple Mail reply extraction, including signatures
    and quoted history.
