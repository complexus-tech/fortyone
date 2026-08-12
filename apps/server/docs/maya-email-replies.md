# Maya email replies

Maya guidance emails are sent from `Maya <maya@fortyone.app>` and use a
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
APP_EMAIL_MAYA_FROM_NAME=Maya
```

### 2. Delegate the inbound reply subdomain

Create these DNS records for the dedicated receiving subdomain:

| Name | Type | Priority | Value |
| --- | --- | ---: | --- |
| `reply.fortyone.app` | MX | 10 | `inbound1.sendinblue.com.` |
| `reply.fortyone.app` | MX | 20 | `inbound2.sendinblue.com.` |

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

Generate the header value with the same `APP_AUTH_SECRET_KEY` used by the API
and worker:

```sh
cd apps/server
go run ./cmd/brevo-webhook-token
```

The command prints the complete header name and derived value without printing
the application secret. Brevo supports custom authentication headers on
webhooks as described in its
[secured webhook documentation](https://developers.brevo.com/docs/secured-webhooks).

Rotating `APP_AUTH_SECRET_KEY` also rotates the webhook header and the key used
to encrypt queued email payloads. Drain pending email-reply work and update the
Brevo header as part of the same controlled rotation.

### 4. Existing runtime configuration

No new reply-specific environment variable is required. Production still
needs the existing values for Postgres, Redis, the shared application secret,
SMTP, and OpenAI:

```text
APP_AUTH_SECRET_KEY
APP_EMAIL_HOST
APP_EMAIL_PORT
APP_EMAIL_USERNAME
APP_EMAIL_PASSWORD
APP_EMAIL_ENVIRONMENT=production
OPENAI_API_KEY
```

In production, `APP_AUTH_SECRET_KEY` must be a unique value of at least 32
bytes. The API and worker refuse to start with the development default or a
short secret because this value protects both the webhook capability and
queued email payloads.

`OPENAI_MODEL` is optional and defaults to `gpt-5.6-luna`. The worker manages
conversation state in FortyOne and calls the model with provider storage
disabled.

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
