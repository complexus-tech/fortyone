-- name: EnsureHumanPrincipal :one
WITH current_account AS (
    SELECT
        account.user_id,
        COALESCE(NULLIF(btrim(account.full_name), ''), account.email) AS display_name
    FROM public.users AS account
    INNER JOIN public.workspace_members AS membership
        ON membership.user_id = account.user_id
    WHERE account.user_id = sqlc.arg(user_id)
      AND membership.workspace_id = sqlc.arg(workspace_id)
      AND account.is_active = TRUE
), inserted AS (
    INSERT INTO public.principals (
        principal_id,
        workspace_id,
        kind,
        name,
        subject_user_id,
        status,
        created_by_user_id,
        created_at,
        updated_at
    )
    SELECT
        sqlc.arg(principal_id),
        sqlc.arg(workspace_id),
        'human_user',
        current_account.display_name,
        current_account.user_id,
        'active',
        current_account.user_id,
        sqlc.arg(created_at),
        sqlc.arg(created_at)
    FROM current_account
    ON CONFLICT (workspace_id, subject_user_id) WHERE kind = 'human_user'
    DO NOTHING
    RETURNING principal_id
)
SELECT inserted.principal_id
FROM inserted
UNION ALL
SELECT principal.principal_id
FROM public.principals AS principal
INNER JOIN current_account
    ON current_account.user_id = principal.subject_user_id
WHERE principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'human_user'
  AND principal.status = 'active'
LIMIT 1;

-- name: ResolveHumanPrincipal :one
SELECT principal.principal_id
FROM public.principals AS principal
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.subject_user_id = sqlc.arg(user_id)
  AND principal.kind = 'human_user'
  AND principal.status = 'active'
  AND account.is_active = TRUE;

-- name: InsertPersonalAccessToken :one
INSERT INTO public.api_credentials (
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    secret_digest,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    created_by_user_id,
    created_at
)
SELECT
    sqlc.arg(credential_id),
    principal.workspace_id,
    principal.principal_id,
    'personal_access_token',
    sqlc.arg(name),
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(token_version),
    sqlc.arg(digest_key_id),
    sqlc.arg(digest_key_version),
    sqlc.arg(expires_at),
    sqlc.arg(user_id),
    sqlc.arg(created_at)
FROM public.principals AS principal
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE principal.principal_id = sqlc.arg(principal_id)
  AND principal.workspace_id = sqlc.arg(workspace_id)
  AND principal.kind = 'human_user'
  AND principal.status = 'active'
  AND principal.subject_user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
RETURNING
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    last_used_at,
    rotated_from_id,
    rotated_at,
    revoked_at,
    revoked_by_user_id,
    revoked_reason,
    created_by_user_id,
    created_at;

-- name: ListPersonalAccessTokens :many
SELECT
    credential.credential_id,
    credential.workspace_id,
    credential.principal_id,
    credential.kind,
    credential.name,
    credential.lookup_prefix,
    credential.token_version,
    credential.expires_at,
    credential.last_used_at,
    credential.rotated_from_id,
    credential.rotated_at,
    credential.revoked_at,
    credential.revoked_by_user_id,
    credential.revoked_reason,
    credential.created_by_user_id,
    credential.created_at,
    CAST(ARRAY(
        SELECT scope.scope
        FROM public.api_credential_scopes AS scope
        WHERE scope.credential_id = credential.credential_id
        ORDER BY scope.scope
    ) AS text[]) AS scopes,
    CAST(ARRAY(
        SELECT restriction.team_id
        FROM public.api_credential_team_restrictions AS restriction
        WHERE restriction.credential_id = credential.credential_id
        ORDER BY restriction.team_id
    ) AS uuid[]) AS team_restrictions
FROM public.api_credentials AS credential
INNER JOIN public.principals AS principal
    ON principal.principal_id = credential.principal_id
    AND principal.workspace_id = credential.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.kind = 'personal_access_token'
  AND principal.kind = 'human_user'
  AND principal.subject_user_id = sqlc.arg(user_id)
  AND principal.status = 'active'
  AND account.is_active = TRUE
ORDER BY credential.created_at DESC, credential.credential_id DESC;

-- name: LockPersonalAccessTokenForRotation :one
SELECT credential.credential_id
FROM public.api_credentials AS credential
INNER JOIN public.principals AS principal
    ON principal.principal_id = credential.principal_id
    AND principal.workspace_id = credential.workspace_id
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE credential.credential_id = sqlc.arg(credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.kind = 'personal_access_token'
  AND credential.revoked_at IS NULL
  AND credential.expires_at > sqlc.arg(rotated_at)
  AND credential.rotated_at IS NULL
  AND principal.kind = 'human_user'
  AND principal.subject_user_id = sqlc.arg(user_id)
  AND principal.status = 'active'
  AND account.is_active = TRUE
FOR UPDATE OF credential;

-- name: InsertRotatedPersonalAccessToken :one
INSERT INTO public.api_credentials (
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    secret_digest,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    rotated_from_id,
    created_by_user_id,
    created_at
)
SELECT
    sqlc.arg(new_credential_id),
    credential.workspace_id,
    credential.principal_id,
    credential.kind,
    credential.name,
    sqlc.arg(lookup_prefix),
    sqlc.arg(secret_digest),
    sqlc.arg(token_version),
    sqlc.arg(digest_key_id),
    sqlc.arg(digest_key_version),
    sqlc.arg(expires_at),
    credential.credential_id,
    sqlc.arg(user_id),
    sqlc.arg(created_at)
FROM public.api_credentials AS credential
WHERE credential.credential_id = sqlc.arg(old_credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.kind = 'personal_access_token'
  AND credential.revoked_at IS NULL
  AND credential.rotated_at IS NULL
RETURNING
    credential_id,
    workspace_id,
    principal_id,
    kind,
    name,
    lookup_prefix,
    token_version,
    digest_key_id,
    digest_key_version,
    expires_at,
    last_used_at,
    rotated_from_id,
    rotated_at,
    revoked_at,
    revoked_by_user_id,
    revoked_reason,
    created_by_user_id,
    created_at;

-- name: RevokePersonalAccessToken :one
UPDATE public.api_credentials AS credential
SET
    revoked_at = sqlc.arg(revoked_at),
    revoked_by_user_id = sqlc.arg(user_id),
    revoked_reason = sqlc.arg(revoked_reason)
FROM public.principals AS principal
INNER JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
INNER JOIN public.users AS account
    ON account.user_id = membership.user_id
WHERE credential.principal_id = principal.principal_id
  AND credential.workspace_id = principal.workspace_id
  AND credential.credential_id = sqlc.arg(credential_id)
  AND credential.workspace_id = sqlc.arg(workspace_id)
  AND credential.kind = 'personal_access_token'
  AND credential.revoked_at IS NULL
  AND principal.kind = 'human_user'
  AND principal.subject_user_id = sqlc.arg(user_id)
  AND principal.status = 'active'
  AND account.is_active = TRUE
RETURNING credential.credential_id;

-- name: MarkCredentialRotated :execrows
UPDATE public.api_credentials
SET
    rotated_at = CAST(sqlc.arg(rotated_at) AS timestamptz),
    expires_at = CASE
        WHEN CAST(sqlc.arg(revoke_immediately) AS boolean) THEN expires_at
        ELSE LEAST(expires_at, CAST(sqlc.arg(overlap_expires_at) AS timestamptz))
    END,
    revoked_at = CASE
        WHEN CAST(sqlc.arg(revoke_immediately) AS boolean) THEN CAST(sqlc.arg(rotated_at) AS timestamptz)
        ELSE NULL
    END,
    revoked_by_user_id = CASE
        WHEN CAST(sqlc.arg(revoke_immediately) AS boolean) THEN CAST(sqlc.arg(user_id) AS uuid)
        ELSE NULL
    END,
    revoked_reason = CASE
        WHEN CAST(sqlc.arg(revoke_immediately) AS boolean) THEN 'rotated'
        ELSE NULL
    END
WHERE credential_id = sqlc.arg(credential_id)
  AND rotated_at IS NULL;
