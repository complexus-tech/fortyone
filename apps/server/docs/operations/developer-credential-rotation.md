# Developer credential key rotation

This runbook rotates the server-side HMAC keyring used to digest PATs and
service-account keys. It is different from rotating an individual user's token
or service-account key.

## Properties

- Existing credential rows record their digest key ID and version.
- New credentials use exactly the configured active generation.
- Verification can read every retained generation.
- Digest keys cannot recover plaintext credentials.
- Removing a generation immediately makes every credential using it
  unauthenticatable, even if its database row is otherwise active.

## Prepare

1. Generate independent 32-byte random material in the managed secret store.
2. Choose a stable non-secret key ID and positive version, for example
   `api-credentials-2026-09@2`.
3. Confirm the material is not used for auth JWTs, verification/invitation
   HMACs, provider encryption, or any previous digest generation.
4. Append the base64 material to `APP_API_CREDENTIAL_HMAC_KEYS` on every API
   instance, leaving the old active generation configured.
5. Deploy and verify readiness. An unknown or malformed keyring must prevent
   startup.

## Activate

1. Set `APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_ID` and
   `APP_API_CREDENTIAL_HMAC_ACTIVE_KEY_VERSION` to the new entry.
2. Roll API instances gradually. Mixed instances remain compatible because all
   instances have both generations before activation.
3. Create a short-lived canary PAT, record only its ID/prefix, and verify its
   database row references the new generation.
4. Authenticate through a read-only `/api/v1` canary endpoint, then revoke the
   canary.

Do not attempt to re-HMAC existing credentials: the plaintext is intentionally
unavailable. Users rotate individual credentials to move them to the new
generation.

## Retire an old generation

Query the exact population first:

```sql
SELECT kind, digest_key_id, digest_key_version, COUNT(*)
FROM api_credentials
WHERE revoked_at IS NULL
  AND expires_at > now()
GROUP BY kind, digest_key_id, digest_key_version
ORDER BY kind, digest_key_id, digest_key_version;
```

Retire only when no unexpired, unrevoked row references the generation. Remove
it from the encoded keyring, deploy gradually, and repeat the query plus
authentication smoke test. Keep secret-manager recovery history under the
organization's incident policy; do not paste keyring values into tickets or
logs.

## Individual credential rotation

PAT rotation is immediate: the replacement is shown once and the old PAT is
revoked atomically. Service-account keys may request zero to 24 hours of
overlap. Prefer zero. Use overlap only when a coordinated consumer rollout
cannot replace the key atomically; shorten it to the smallest operational
window and verify old-key rejection after it closes.

## Incident response

For one exposed credential, revoke that credential (or disable the service
account), verify the next request fails, issue a replacement, and inspect the
immutable audit plus `last_used_at`. For suspected HMAC-key exposure, keep the
compromised generation temporarily readable only long enough to revoke/expire
all rows that use it, activate a new independent generation, and then remove
the old generation. Credential text, digests, and HMAC material must not enter
incident chat, screenshots, logs, or exported traces.
