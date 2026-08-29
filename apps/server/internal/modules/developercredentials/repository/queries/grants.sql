-- name: IsCurrentWorkspaceMember :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS account
        ON account.user_id = membership.user_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(user_id)
      AND account.is_active = TRUE
);

-- name: IsCurrentWorkspaceAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS account
        ON account.user_id = membership.user_id
    WHERE membership.workspace_id = sqlc.arg(workspace_id)
      AND membership.user_id = sqlc.arg(user_id)
      AND membership.role = 'admin'
      AND account.is_active = TRUE
);

-- name: ValidatePersonalTokenTeamRestrictions :one
SELECT cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = COUNT(DISTINCT team.team_id)
FROM public.teams AS team
INNER JOIN public.team_members AS team_member
    ON team_member.team_id = team.team_id
INNER JOIN public.workspace_members AS workspace_member
    ON workspace_member.workspace_id = team.workspace_id
    AND workspace_member.user_id = team_member.user_id
INNER JOIN public.users AS account
    ON account.user_id = workspace_member.user_id
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND team_member.user_id = sqlc.arg(user_id)
  AND team.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[]))
  AND account.is_active = TRUE;

-- name: ValidateServiceAccountTeamRestrictions :one
SELECT cardinality(CAST(sqlc.arg(team_ids) AS uuid[])) = COUNT(DISTINCT team.team_id)
FROM public.teams AS team
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND team.team_id = ANY(CAST(sqlc.arg(team_ids) AS uuid[]));

-- name: InsertCredentialScopes :exec
INSERT INTO public.api_credential_scopes (
    credential_id,
    scope
)
SELECT
    sqlc.arg(credential_id),
    requested.scope
FROM unnest(CAST(sqlc.arg(scopes) AS text[])) AS requested(scope);

-- name: InsertCredentialTeamRestrictions :exec
INSERT INTO public.api_credential_team_restrictions (
    credential_id,
    workspace_id,
    team_id
)
SELECT
    sqlc.arg(credential_id),
    sqlc.arg(workspace_id),
    requested.team_id
FROM unnest(CAST(sqlc.arg(team_ids) AS uuid[])) AS requested(team_id);

-- name: CopyCredentialScopes :exec
INSERT INTO public.api_credential_scopes (
    credential_id,
    scope
)
SELECT
    sqlc.arg(new_credential_id),
    existing_scope.scope
FROM public.api_credential_scopes AS existing_scope
WHERE existing_scope.credential_id = sqlc.arg(old_credential_id);

-- name: CopyCredentialTeamRestrictions :exec
INSERT INTO public.api_credential_team_restrictions (
    credential_id,
    workspace_id,
    team_id
)
SELECT
    sqlc.arg(new_credential_id),
    restriction.workspace_id,
    restriction.team_id
FROM public.api_credential_team_restrictions AS restriction
WHERE restriction.credential_id = sqlc.arg(old_credential_id);

-- name: GetCredentialGrants :one
SELECT
    CAST(ARRAY(
        SELECT scope.scope
        FROM public.api_credential_scopes AS scope
        WHERE scope.credential_id = sqlc.arg(credential_id)
        ORDER BY scope.scope
    ) AS text[]) AS scopes,
    CAST(ARRAY(
        SELECT restriction.team_id
        FROM public.api_credential_team_restrictions AS restriction
        WHERE restriction.credential_id = sqlc.arg(credential_id)
        ORDER BY restriction.team_id
    ) AS uuid[]) AS team_restrictions;
