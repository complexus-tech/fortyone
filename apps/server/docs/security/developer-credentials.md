# Developer credentials

Personal access tokens (PATs) and service-account keys are the show-once
credential families for the versioned FortyOne API. This document is the
security contract for their generation, storage, authentication, authorization,
and incident response. User-authorized OAuth is a separate, audience-bound
bearer family documented in [Developer OAuth](developer-oauth.md).

## Release boundary

Credential management is available to the first-party application on the
unversioned management routes listed below. The strict developer bearer
middleware is mounted only by the released `/api/v1` route group and is **not
mounted on legacy application routes**. The first preview operations expose
workspace/team/story and related collaboration reads, story creation, and
outbound-webhook management. User-backed operations accept a PAT or an OAuth
token for the exact `<APP_API_PUBLIC_URL>/api/v1` resource. Service-account
keys fail closed on user-only operations; the current explicit non-human
contract is idempotent story creation. Creating any credential never makes
legacy routes callable with it.

This boundary prevents a broad authentication fallback from silently changing
the security semantics of hundreds of existing browser routes.

## Credential forms

| Kind                | Text header   | Principal recorded in `Actor.PrincipalID` | `Actor.CredentialID` |
| ------------------- | ------------- | ----------------------------------------- | -------------------- |
| PAT                 | `f41_pat_v1_` | underlying user ID                        | PAT credential ID    |
| Service-account key | `f41_sak_v1_` | service-account principal ID              | key credential ID    |

The remainder is a 12-character random hexadecimal lookup prefix, an
underscore, and a 43-character unpadded base64url encoding of 32 random bytes.
The complete bearer is shown only in the successful create or rotate response.
It must never be returned by list endpoints.

The lookup prefix is non-secret and may appear in an administration UI, audit
workflow, or support conversation. It is not sufficient to authenticate.

## Storage and verification

The API stores:

- the random lookup prefix;
- HMAC-SHA-256 of a domain-separated message containing token version,
  credential kind, credential ID, prefix, and secret;
- digest key ID and key version;
- principal/workspace ownership, scopes, optional team restrictions, expiry,
  rotation/revocation state, and coarse last-used metadata.

It never stores the plaintext token. The dedicated
`APP_API_CREDENTIAL_HMAC_*` keyring is separate from browser auth, verification
tokens, invitation tokens, and provider credential encryption. Production
startup rejects the public development key, missing active generations,
duplicate key material, malformed keyrings, and cross-purpose key reuse.

Authentication uses this fail-closed sequence:

1. accept exactly one `Authorization: Bearer <token>` header;
2. reject cookies, query parameters, legacy JWT fallback, malformed forms, and
   unknown versions;
3. look up an unrevoked, unexpired record by kind/version/prefix;
4. recheck the principal, user, and current workspace membership in PostgreSQL;
5. calculate the version-selected HMAC and compare it in constant time;
6. construct the explicit PAT or service-account actor with scopes, workspace,
   team restrictions, and credential ID;
7. recheck active state and update `last_used_at` at most once per 15 minutes.

The final active-state query closes the verification/revocation race. A revoke,
membership deletion, user deactivation, or service-account disable wins before
the actor enters the request context.

## Least privilege

Scopes narrow product authorization; they never create workspace membership,
role, ownership, or team membership. Team restrictions narrow a credential
again and are not a substitute for resource-level authorization.

PATs inherit the user's current workspace role and stop authenticating when the
user or membership is inactive. Service accounts carry an explicit role, which
is restricted to `guest` or `member` in both service validation and a database
constraint. They can never be workspace admins.

Service-account keys cannot receive `service_accounts:manage`. The service
rejects that grant and a database trigger rejects direct or future alternate
writes. Credential management handlers and services independently allow only a
`human_user`, so adding a broad scope to a machine credential in the future
cannot make it self-replicating.

## Management routes

These routes use the existing first-party browser session, current workspace
middleware, and mutation rate limit. They do not accept machine credentials.

| Method and path                                              | Policy                                             | Result                                                            |
| ------------------------------------------------------------ | -------------------------------------------------- | ----------------------------------------------------------------- |
| `GET /workspaces/{slug}/personal-access-tokens`              | current human member                               | redacted metadata list                                            |
| `POST /workspaces/{slug}/personal-access-tokens`             | current human member                               | metadata plus show-once token                                     |
| `POST /workspaces/{slug}/personal-access-tokens/{id}/rotate` | current owning human                               | replacement plus show-once token; old token revoked immediately   |
| `DELETE /workspaces/{slug}/personal-access-tokens/{id}`      | current owning human                               | revoke                                                            |
| `GET/POST /workspaces/{slug}/service-accounts`               | current human admin plus `service_accounts:manage` | list/create                                                       |
| `DELETE /workspaces/{slug}/service-accounts/{id}`            | same                                               | disable and revoke all keys                                       |
| `GET/POST /workspaces/{slug}/service-accounts/{id}/keys`     | same                                               | redacted list/create                                              |
| `POST .../keys/{keyId}/rotate`                               | same                                               | replacement plus show-once token; optional overlap up to 24 hours |
| `DELETE .../keys/{keyId}`                                    | same                                               | revoke key                                                        |

Create payloads contain `name`, a non-empty `scopes` array, optional `teamIds`,
and an explicit `expiresAt`. Expiry must be between one minute and 365 days.
Service-account creation additionally requires `workspaceRole` equal to
`guest` or `member`. Rotation requires a new expiry. Service-account key
rotation may include `overlapSeconds` from zero through 86,400; PAT rotation is
always immediate.

## Principal registry capability

Modules that need a durable human-principal foreign key must depend on the
narrow `EnsureHumanPrincipal` service capability; they must not write
`principals` directly. A first-party human actor may resolve or provision its
row. A PAT actor may resolve only the already-existing row backing its subject
user, after current membership checks. Service accounts and every other actor
kind are denied. This preserves the distinction between PAT actor attribution
(the user ID) and database foreign keys (the principal record ID).
Concurrent first-party provisioning uses bounded retries around serializable
transactions, so callers converge on the one workspace/user principal without
weakening the unique database invariant.

## Audit and safe observability

Every credential mutation and first-party principal provision writes an event
to `developer_credential_audit_events` in the same transaction. The ledger
stores UUID facts, actor kind, optional actor credential ID, operation, result,
request ID, and bounded metadata. It intentionally has no mutable foreign keys,
so account deletion cannot rewrite history. A trigger rejects update and delete.

Never log or trace Authorization headers, token text, HMAC keys, digests, or
create/rotate response bodies. `PlaintextToken` renders as `[REDACTED]` through
string and structured logging. Authentication failures use one public response
for unknown, expired, revoked, malformed, and bad-secret credentials.

## Security review checklist

- Route is below `/api/v1` and installs machine authentication explicitly.
- Allowed principal kinds are listed in the service policy.
- Required scopes, current workspace role, team restriction, ownership, and
  workspace predicates are all evaluated independently.
- Repository SQL repeats tenant and current-state checks.
- Response models have no digest, digest key, or plaintext fields.
- Tests cover two tenants, membership loss, expiry/revoke/disable, forbidden
  scope/role, one-winner rotation, last-used coalescing, and secret non-leakage.
- New scope strings are added to the auth catalog, database constraint, API
  documentation, and policy matrices together.
