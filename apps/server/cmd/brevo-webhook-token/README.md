# Brevo inbound webhook token

This command derives the Brevo inbound webhook header value from the dedicated
`APP_EMAIL_REPLY_SECURITY_KEY`. It never prints the key itself. It refuses the
development default and keys shorter than 32 bytes.

From `apps/server`, with the same environment used by the API process:

```sh
go run ./cmd/brevo-webhook-token
```

The output is ready to add to the Brevo inbound webhook as a custom header:

```text
X-FortyOne-Webhook-Token: <derived value>
```

Configure the webhook URL as
`https://<api-host>/webhooks/brevo/inbound-email-processed`, with type
`inbound`, event `inboundEmailProcessed`, and the delegated inbound reply
domain. Rotating `APP_EMAIL_REPLY_SECURITY_KEY` also rotates this header and
the durable payload-encryption key, so drain pending email-reply work and
update the Brevo webhook as part of the same controlled deployment.
