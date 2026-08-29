# Purpose-specific application keys

Browser sessions use opaque random tokens backed by versioned Redis records and
PostgreSQL account epochs; they do not use `APP_AUTH_SECRET_KEY` for signing or
verification. That variable remains the legacy application key for bounded
non-session protocols that have not yet moved to dedicated keys. New security
capabilities must use an independently generated key or keyring. The API and
worker validate required production values at startup and reject known
development material, keys shorter than 32 bytes, reuse between named
capabilities, and reuse of a credential-vault key.

## Active boundaries

| Configuration                                                        | Owner                 | Protects                                                                                                                           | Shared by                       |
| -------------------------------------------------------------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| `APP_AUTH_SECRET_KEY`                                                | legacy application    | Existing calendar, Maya, feedback, public-API cursor, and bounded Slack cutover cryptography; never browser-session authentication | API and worker legacy consumers |
| `APP_EMAIL_REPLY_SECURITY_KEY`                                       | Email Reply           | Brevo ingress proof and authenticated encryption of durable email-reply payloads                                                   | API and worker                  |
| `APP_MESSAGING_MUTATION_HMAC_KEY`                                    | Messaging             | short-lived assistant mutation proposals and confirmations                                                                         | API and worker                  |
| `APP_FEEDBACK_SECURITY_KEY`                                          | Feedback              | verification digests, unsubscribe links, and widget-secret encryption                                                              | API and worker                  |
| `APP_VERIFICATION_TOKEN_HMAC_KEY`                                    | Users                 | email-verification digests and opaque abuse-limit identifiers                                                                      | API                             |
| `APP_INVITATION_TOKEN_HMAC_KEY`                                      | Invitations           | invitation bearer-token digests                                                                                                    | API and worker                  |
| `APP_API_CREDENTIAL_HMAC_KEYS`                                       | Developer credentials | PAT and service-account secret digests                                                                                             | API                             |
| `APP_OAUTH_ACCESS_TOKEN_SIGNING_KEY` and `APP_OAUTH_TOKEN_HMAC_KEYS` | Developer OAuth       | access-token signatures and authorization/refresh-token digests                                                                    | API                             |
| `APP_CREDENTIAL_VAULT_KEYS`                                          | Integration platform  | provider credential envelope encryption                                                                                            | API and worker                  |
| `APP_<PROVIDER>_WEBHOOK_PAYLOAD_SECRET`                              | provider adapter      | authenticated encryption for one durable provider inbox                                                                            | API and worker                  |

Domain separation inside one capability prevents cross-protocol use; it does
not justify sharing that key with another row in the table. Provider signing
secrets, OAuth client secrets, storage credentials, and deployment identities
are also independent and are never accepted as substitutes for application
keys.

## Provisioning

Generate at least 32 random bytes per key in the managed secret store. Inject
the same value into API and worker task definitions only when the table says the
capability is shared. Never place plaintext values in source control, generated
configuration documentation, logs, traces, task payloads, or support tickets.

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
