# Brevo inbound webhook token

This command derives the Brevo inbound webhook header value from the existing
`APP_AUTH_SECRET_KEY`. It never prints the application secret itself and does
not require a separate environment variable. It refuses the development
default and secrets shorter than 32 bytes.

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
domain. Rotating `APP_AUTH_SECRET_KEY` also rotates this header, so update the
Brevo webhook as part of the same deployment.
