# Purpose-specific application keys

Browser sessions use opaque random tokens backed by versioned Redis records and
PostgreSQL account epochs; they do not use `APP_AUTH_SECRET_KEY` for signing or
verification. That variable is the stable application root for legacy
protocols and versioned, purpose-separated integration keys. New integration
capabilities should derive a labelled key through `internal/platform/appkeys`
instead of adding another operator-managed environment variable. Capabilities
that issue externally durable credentials keep their explicit versioned
keyrings so an application-root change cannot silently invalidate them.

## Active boundaries

| Configuration                                                        | Owner                 | Protects                                                                                                                           | Shared by                       |
| -------------------------------------------------------------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| `APP_AUTH_SECRET_KEY`                                                | application root      | Legacy protocols and HKDF-derived integration keys; never browser-session authentication                                           | API and worker                  |
| `APP_EMAIL_REPLY_SECURITY_KEY`                                       | Email Reply           | Brevo ingress proof and authenticated encryption of durable email-reply payloads                                                   | API and worker                  |
| `APP_MESSAGING_MUTATION_HMAC_KEY`                                    | Messaging             | short-lived assistant mutation proposals and confirmations                                                                         | API and worker                  |
| `APP_FEEDBACK_SECURITY_KEY`                                          | Feedback              | verification digests, unsubscribe links, and widget-secret encryption                                                              | API and worker                  |
| `APP_VERIFICATION_TOKEN_HMAC_KEY`                                    | Users                 | email-verification digests and opaque abuse-limit identifiers                                                                      | API                             |
| `APP_INVITATION_TOKEN_HMAC_KEY`                                      | Invitations           | invitation bearer-token digests                                                                                                    | API and worker                  |
| `APP_API_CREDENTIAL_HMAC_KEYS`                                       | Developer credentials | PAT and service-account secret digests                                                                                             | API                             |
| `APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY` and `APP_OAUTH_TOKEN_HMAC_KEYS` | Developer OAuth       | access-token signatures and authorization/refresh-token digests                                                                    | API                             |
| Internally derived integration keys                                 | Integration platform  | provider credential envelopes and GitHub, Slack, and Figma durable webhook inboxes                                                  | API and worker                  |

The integration derivation uses HKDF-SHA256 with fixed, versioned purpose
labels. This provides domain separation without asking an operator to
provision four additional secrets. Provider signing secrets, OAuth client
secrets, storage credentials, and deployment identities remain external
provider credentials and are never accepted as application keys.

## Provisioning

Generate at least 32 random bytes for each operator-managed key in the managed
secret store. Keep `APP_AUTH_SECRET_KEY` identical across API and worker tasks;
the application derives integration keys locally and never exports them. Never
place plaintext values in source control, generated configuration
documentation, logs, traces, task payloads, or support tickets.

Before deployment, run:

```bash
make config-check
go test ./cmd/api ./internal/bootstrap/worker -run 'TestValidateRuntimeConfig' -count=1
```

Startup errors name configuration variables but deliberately never include
their values.

## Email-reply key rotation

The current email-reply envelope does not carry a key-generation identifier.
Treat rotation as a controlled drain:

1. stop new Brevo ingress and pause the email-reply recovery schedule;
2. wait until the durable inbound and outbound email-reply queues are empty;
3. configure the new key on every API and worker replica;
4. derive the new header with `go run ./cmd/brevo-webhook-token` and update
   Brevo before reopening ingress;
5. deploy API and worker together, resume recovery, and verify one synthetic
   reply end to end;
6. retire the previous value from the secret store after the rollback window.

Do not rotate this key during an unresolved email-reply incident. A future
versioned envelope may support read-old/write-new overlap; until then, draining
is the explicit safety contract.

## Messaging mutation key rotation

Mutation confirmations are intentionally short-lived. Deploy the new
`APP_MESSAGING_MUTATION_HMAC_KEY` to API and worker replicas together. Pending
proposals signed under the old key will fail closed and must be proposed again;
no mutation is applied merely because a key changed. Verify proposal,
confirmation, cancellation, expiry, and replay behavior after rollout.

## Incident response

For suspected exposure, identify the exact capability, revoke or rotate only
its key, and inspect immutable audit/request identifiers without copying secret
material. Expanding the incident to unrelated keys is based on evidence of
reuse or broader secret-store compromise, not on the former shared-auth-key
design. Record the new key generation, deployment correlation, validation
evidence, affected credential lifetime, and retirement time in the incident.
