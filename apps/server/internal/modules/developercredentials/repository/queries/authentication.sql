-- name: LookupCredentialForAuthentication :one
SELECT
    credential.credential_id,
    credential.workspace_id,
    credential.principal_id AS principal_record_id,
    principal.kind AS principal_kind,
    principal.subject_user_id,
    CAST(
        CASE
            WHEN credential.kind = 'personal_access_token' THEN membership.role
            ELSE principal.workspace_role
        END
        AS text
    ) AS workspace_role,
    credential.kind AS credential_kind,
    credential.lookup_prefix,
    credential.secret_digest,
    credential.token_version,
    credential.digest_key_id,
    credential.digest_key_version,
    credential.expires_at,
    credential.last_used_at,
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
LEFT JOIN public.users AS account
    ON account.user_id = principal.subject_user_id
LEFT JOIN public.workspace_members AS membership
    ON membership.workspace_id = principal.workspace_id
    AND membership.user_id = principal.subject_user_id
WHERE credential.lookup_prefix = sqlc.arg(lookup_prefix)
  AND credential.kind = sqlc.arg(credential_kind)
  AND credential.token_version = sqlc.arg(token_version)
  AND credential.revoked_at IS NULL
  AND credential.expires_at > sqlc.arg(authenticated_at)
  AND principal.status = 'active'
  AND EXISTS (
      SELECT 1
      FROM public.api_credential_scopes AS granted_scope
      WHERE granted_scope.credential_id = credential.credential_id
  )
  AND (
      (
          credential.kind = 'personal_access_token'
          AND principal.kind = 'human_user'
          AND account.is_active = TRUE
          AND membership.user_id IS NOT NULL
      )
      OR
      (
          credential.kind = 'service_account_key'
          AND principal.kind = 'service_account'
          AND principal.workspace_role IS NOT NULL
      )
  );

-- name: ConfirmCredentialActiveAndTouch :one
WITH eligible AS MATERIALIZED (
    SELECT credential.credential_id
    FROM public.api_credentials AS credential
    INNER JOIN public.principals AS principal
        ON principal.principal_id = credential.principal_id
        AND principal.workspace_id = credential.workspace_id
    LEFT JOIN public.users AS account
        ON account.user_id = principal.subject_user_id
    LEFT JOIN public.workspace_members AS membership
        ON membership.workspace_id = principal.workspace_id
        AND membership.user_id = principal.subject_user_id
    WHERE credential.credential_id = sqlc.arg(credential_id)
      AND credential.revoked_at IS NULL
      AND credential.expires_at > sqlc.arg(used_at)
      AND principal.status = 'active'
      AND (
          (
              credential.kind = 'personal_access_token'
              AND principal.kind = 'human_user'
              AND account.is_active = TRUE
              AND membership.user_id IS NOT NULL
          )
          OR
          (
              credential.kind = 'service_account_key'
              AND principal.kind = 'service_account'
              AND principal.workspace_role IS NOT NULL
          )
      )
), touched AS (
    UPDATE public.api_credentials AS credential
    SET last_used_at = sqlc.arg(used_at)
    WHERE credential.credential_id IN (SELECT eligible.credential_id FROM eligible)
      AND (
          credential.last_used_at IS NULL
          OR credential.last_used_at <= sqlc.arg(touch_before)
      )
    RETURNING credential.credential_id
)
SELECT eligible.credential_id
FROM eligible;
