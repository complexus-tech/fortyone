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

1. Pause integration ingress and provider mutations, including Google Drive
   connect/disconnect callbacks and new revocation claims.
2. Let any in-flight Google Drive provider lifecycle gates finish, then stop API
   and worker tasks that use the old root.
3. Preserve a recoverable database backup and identify retained provider
   credentials and pending webhook receipts.
4. While the old root is still available to both API and worker, drain pending
   and processing Google Drive revocation tombstones and reconcile retained
   failed tombstones, including source-account-free cleanup jobs created by
   failed OAuth callbacks. Do not delete or rewrite a sealed job merely because
   its account or user no longer exists. Confirm terminal completed/superseded
   rows no longer hold an envelope. Revoke and reconnect affected GitHub, Slack, Figma, and Google
   Drive credentials, or ship a reviewed one-time
   old-root-to-new-root migration before changing the production value.
5. Deploy the API and worker together with the new root.
6. Exercise provider connection, refresh, webhook, disconnect, and retry flows.
   For Google Drive, verify Picker selection, a bounded content read, final
   workspace disconnect, and revocation-outbox completion.
7. Retire the old root only after the migration or reconnection proof succeeds.

Never discard an encrypted Google Drive revocation envelope merely because its
user or account row was deleted. It is deliberately foreign-key independent and
must be completed, safely superseded by a newer same-subject generation, or
rewrapped with its original authenticated context before the old root retires.

Never deploy a new root first and hope existing ciphertext can be recovered
afterward. If an attacker may have obtained a provider token, re-encryption is
not enough; revoke that token at the provider.

## Future upgrade point

If independent online key rotation becomes a real operational requirement, add
a managed versioned keyring or KMS implementation behind the existing vault
contract. Do not reintroduce several manually coordinated environment variables
without an operational owner and tested rotation procedure.
