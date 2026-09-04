-- name: DeleteExpiredGoogleDriveOAuthStates :execrows
DELETE FROM public.google_drive_oauth_states
WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: SaveGoogleDriveOAuthState :execrows
INSERT INTO public.google_drive_oauth_states (
    state_hash,
    workspace_id,
    user_id,
    workspace_slug,
    return_url,
    code_verifier,
    expires_at
)
SELECT
    sqlc.arg(state_hash),
    workspace.workspace_id,
    membership.user_id,
    workspace.slug,
    CAST(sqlc.narg(return_url) AS text),
    sqlc.arg(code_verifier),
    sqlc.arg(expires_at)
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
   AND membership.user_id = sqlc.arg(user_id)
   AND membership.role IN ('member', 'admin')
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.slug = sqlc.arg(workspace_slug)
  AND workspace.deleted_at IS NULL;

-- name: ConsumeGoogleDriveOAuthState :one
DELETE FROM public.google_drive_oauth_states
WHERE state_hash = sqlc.arg(state_hash)
  AND expires_at > sqlc.arg(consumed_at)
RETURNING state_hash, workspace_id, user_id, workspace_slug, return_url,
    code_verifier, expires_at;

-- name: LockGoogleDriveUserLifecycle :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'google-drive:' || CAST(CAST(sqlc.arg(user_id) AS uuid) AS text),
        0
    )
);

-- name: LockGoogleDriveSubjectLifecycle :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        'google-drive-subject:' || sqlc.arg(google_subject),
        0
    )
);

-- name: AcquireGoogleDriveProviderLifecycleLock :exec
SELECT pg_advisory_lock(hashtextextended(sqlc.arg(lock_key), 0));

-- name: ReleaseGoogleDriveProviderLifecycleLock :one
SELECT pg_advisory_unlock(hashtextextended(sqlc.arg(lock_key), 0));

-- name: UpsertGoogleDriveAccount :one
INSERT INTO public.google_drive_accounts (
    user_id,
    google_subject,
    email,
    display_name,
    credential_payload,
    credential_key_version,
    installation_generation,
    scopes,
    expires_at
)
SELECT
    actor.user_id,
    sqlc.arg(google_subject),
    sqlc.arg(email),
    CAST(sqlc.narg(display_name) AS text),
    sqlc.arg(credential_payload),
    sqlc.arg(credential_key_version),
    sqlc.arg(installation_generation),
    sqlc.arg(scopes),
    sqlc.arg(expires_at)
FROM public.users AS actor
WHERE actor.user_id = sqlc.arg(user_id)
  AND actor.is_active = TRUE
ON CONFLICT (user_id, google_subject) WHERE revoked_at IS NULL
DO UPDATE SET
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    credential_payload = EXCLUDED.credential_payload,
    credential_key_version = EXCLUDED.credential_key_version,
    installation_generation = EXCLUDED.installation_generation,
    scopes = EXCLUDED.scopes,
    expires_at = EXCLUDED.expires_at,
    requires_reauthorization = FALSE,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
RETURNING account_id, user_id, google_subject, email, display_name,
    credential_payload, credential_key_version, installation_generation,
    scopes, expires_at, requires_reauthorization, created_at, updated_at;

-- name: UpsertGoogleDriveWorkspaceConnection :execrows
INSERT INTO public.google_drive_workspace_connections (
    workspace_id,
    user_id,
    account_id
)
SELECT
    workspace.workspace_id,
    membership.user_id,
    account.account_id
FROM public.workspaces AS workspace
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = workspace.workspace_id
   AND membership.user_id = sqlc.arg(user_id)
   AND membership.role IN ('member', 'admin')
INNER JOIN public.google_drive_accounts AS account
    ON account.account_id = sqlc.arg(account_id)
   AND account.user_id = membership.user_id
   AND account.revoked_at IS NULL
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL
ON CONFLICT (workspace_id, user_id)
DO UPDATE SET
    account_id = EXCLUDED.account_id,
    connected_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE google_drive_workspace_connections.account_id = EXCLUDED.account_id;

-- name: GetGoogleDriveConnection :one
SELECT
    connection.workspace_id,
    account.account_id,
    account.user_id,
    account.google_subject,
    account.email,
    account.display_name,
    account.credential_payload,
    account.credential_key_version,
    account.installation_generation,
    account.scopes,
    account.expires_at,
    account.requires_reauthorization,
    account.created_at,
    account.updated_at
FROM public.google_drive_workspace_connections AS connection
INNER JOIN public.google_drive_accounts AS account
    ON account.account_id = connection.account_id
   AND account.user_id = connection.user_id
   AND account.revoked_at IS NULL
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = connection.workspace_id
   AND workspace.deleted_at IS NULL
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = connection.workspace_id
   AND membership.user_id = connection.user_id
INNER JOIN public.users AS actor
    ON actor.user_id = membership.user_id
   AND actor.is_active = TRUE
WHERE connection.workspace_id = sqlc.arg(workspace_id)
  AND connection.user_id = sqlc.arg(user_id);

-- name: GetActiveGoogleDriveAccountBySubject :one
SELECT account_id, user_id, google_subject, email, display_name,
    credential_payload, credential_key_version, installation_generation,
    scopes, expires_at, requires_reauthorization, created_at, updated_at
FROM public.google_drive_accounts
WHERE google_subject = sqlc.arg(google_subject)
  AND revoked_at IS NULL;

-- name: CompareAndSwapGoogleDriveCredential :execrows
UPDATE public.google_drive_accounts
SET credential_payload = sqlc.arg(next_payload),
    credential_key_version = sqlc.arg(credential_key_version),
    expires_at = sqlc.arg(expires_at),
    requires_reauthorization = FALSE,
    last_error_code = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE account_id = sqlc.arg(account_id)
  AND installation_generation = sqlc.arg(installation_generation)
  AND credential_payload = sqlc.arg(previous_payload)
  AND revoked_at IS NULL;

-- name: MarkGoogleDriveReauthorizationRequired :execrows
UPDATE public.google_drive_accounts
SET requires_reauthorization = TRUE,
    last_error_code = sqlc.arg(error_code),
    updated_at = CURRENT_TIMESTAMP
WHERE account_id = sqlc.arg(account_id)
  AND installation_generation = sqlc.arg(installation_generation)
  AND revoked_at IS NULL;

-- name: DeleteGoogleDriveWorkspaceConnection :one
DELETE FROM public.google_drive_workspace_connections
WHERE workspace_id = sqlc.arg(workspace_id)
  AND user_id = sqlc.arg(user_id)
RETURNING account_id;

-- name: CountGoogleDriveAccountConnections :one
SELECT COUNT(*)
FROM public.google_drive_workspace_connections
WHERE account_id = sqlc.arg(account_id);

-- name: RevokeUnusedGoogleDriveAccount :execrows
WITH candidate AS MATERIALIZED (
    SELECT account.*
    FROM public.google_drive_accounts AS account
    WHERE account.account_id = sqlc.arg(account_id)
      AND account.user_id = sqlc.arg(user_id)
      AND account.revoked_at IS NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.google_drive_workspace_connections AS connection
          WHERE connection.account_id = account.account_id
      )
    FOR UPDATE
), staged AS MATERIALIZED (
    SELECT public.stage_google_drive_account_revocation(
        candidate.account_id,
        candidate.user_id,
        candidate.google_subject,
        candidate.installation_generation,
        candidate.credential_payload,
        candidate.credential_key_version
    ) AS staged
    FROM candidate
)
UPDATE public.google_drive_accounts AS account
SET credential_payload = '',
    google_subject = '',
    email = '',
    display_name = NULL,
    scopes = '{}'::text[],
    requires_reauthorization = TRUE,
    last_error_code = NULL,
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
FROM candidate
WHERE account.account_id = candidate.account_id
  AND (SELECT COUNT(*) FROM staged) = 1;
