# Application root-secret recovery

FortyOne derives its provider-credential and webhook-inbox encryption keys from
the stable `APP_AUTH_SECRET_KEY`. Routine integration operation therefore has
one application root to manage and no credential-vault rotation variables.

## Normal operation

- Keep the same strong root in the production secret store.
- Give the API and worker the same value.
- Do not expose it in logs, task payloads, tickets, or shell history.
- Do not rotate it as ordinary housekeeping.

Purpose-separated HKDF derivation prevents one integration key from being used
in another protocol. It does not make an old encrypted value readable after the
root itself changes.

## If the root must change

Treat a root change as an incident and a credential migration, not as a normal
deployment toggle.

1. Pause integration ingress and provider mutations.
2. Stop API and worker tasks that use the old root.
3. Preserve a recoverable database backup and identify retained provider
   credentials and pending webhook receipts.
4. Revoke and reconnect affected GitHub, Slack, and Figma credentials, or ship
   a reviewed one-time old-root-to-new-root migration before changing the
   production value.
5. Deploy the API and worker together with the new root.
6. Exercise provider connection, refresh, webhook, disconnect, and retry flows.
7. Retire the old root only after the migration or reconnection proof succeeds.

Never deploy a new root first and hope existing ciphertext can be recovered
afterward. If an attacker may have obtained a provider token, re-encryption is
not enough; revoke that token at the provider.

## Future upgrade point

If independent online key rotation becomes a real operational requirement, add
a managed versioned keyring or KMS implementation behind the existing vault
contract. Do not reintroduce several manually coordinated environment variables
without an operational owner and tested rotation procedure.
