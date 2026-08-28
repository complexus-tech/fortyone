-- name: GetInvitationByToken :one
SELECT
    invitation.invitation_id,
    invitation.workspace_id,
    invitation.inviter_id,
    invitation.email,
    invitation.role,
    invitation.expires_at,
    invitation.used_at,
    invitation.created_at,
    invitation.updated_at,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    workspace.color AS workspace_color,
    CAST(ARRAY(
        SELECT assignment.team_id
        FROM public.workspace_invitation_teams AS assignment
        WHERE assignment.invitation_id = invitation.invitation_id
        ORDER BY assignment.team_id
    ) AS uuid[]) AS team_ids
FROM public.workspace_invitations AS invitation
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = invitation.workspace_id
WHERE (
        CAST(sqlc.arg(token_key_id) AS text) <> ''
        AND invitation.token_key_id = sqlc.arg(token_key_id)
        AND invitation.token_version = sqlc.arg(token_version)
        AND invitation.token_digest = sqlc.arg(token_digest)
    )
    OR (
        CAST(sqlc.arg(legacy_token) AS text) <> ''
        AND invitation.token_digest IS NULL
        AND invitation.token = sqlc.arg(legacy_token)
    )
LIMIT 1;

-- name: LockInvitationByToken :one
SELECT
    invitation.invitation_id,
    invitation.workspace_id,
    invitation.inviter_id,
    invitation.email,
    invitation.role,
    invitation.expires_at,
    invitation.used_at,
    invitation.created_at,
    invitation.updated_at,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    workspace.color AS workspace_color,
    CAST(ARRAY(
        SELECT assignment.team_id
        FROM public.workspace_invitation_teams AS assignment
        WHERE assignment.invitation_id = invitation.invitation_id
        ORDER BY assignment.team_id
    ) AS uuid[]) AS team_ids
FROM public.workspace_invitations AS invitation
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = invitation.workspace_id
WHERE (
        CAST(sqlc.arg(token_key_id) AS text) <> ''
        AND invitation.token_key_id = sqlc.arg(token_key_id)
        AND invitation.token_version = sqlc.arg(token_version)
        AND invitation.token_digest = sqlc.arg(token_digest)
    )
    OR (
        CAST(sqlc.arg(legacy_token) AS text) <> ''
        AND invitation.token_digest IS NULL
        AND invitation.token = sqlc.arg(legacy_token)
    )
LIMIT 1
FOR UPDATE OF invitation;

-- name: LockActiveWorkspaceAdmin :one
SELECT TRUE AS authorized
FROM public.workspace_members AS member
INNER JOIN public.users AS actor
    ON actor.user_id = member.user_id
   AND actor.is_active = TRUE
WHERE member.workspace_id = sqlc.arg(workspace_id)
  AND member.user_id = sqlc.arg(actor_id)
  AND member.role = 'admin'
FOR UPDATE OF member, actor;

-- name: ListWorkspaceInvitations :many
SELECT
    invitation.invitation_id,
    invitation.workspace_id,
    invitation.inviter_id,
    invitation.email,
    invitation.role,
    invitation.expires_at,
    invitation.used_at,
    invitation.created_at,
    invitation.updated_at,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    workspace.color AS workspace_color,
    CAST(ARRAY(
        SELECT assignment.team_id
        FROM public.workspace_invitation_teams AS assignment
        WHERE assignment.invitation_id = invitation.invitation_id
        ORDER BY assignment.team_id
    ) AS uuid[]) AS team_ids
FROM public.workspace_invitations AS invitation
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = invitation.workspace_id
WHERE invitation.workspace_id = sqlc.arg(workspace_id)
  AND invitation.used_at IS NULL
  AND invitation.expires_at > sqlc.arg(now)
ORDER BY invitation.created_at DESC, invitation.invitation_id DESC;

-- name: ListInvitationsByEmail :many
SELECT
    invitation.invitation_id,
    invitation.workspace_id,
    invitation.inviter_id,
    invitation.email,
    invitation.role,
    invitation.expires_at,
    invitation.used_at,
    invitation.created_at,
    invitation.updated_at,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    workspace.color AS workspace_color,
    CAST(ARRAY(
        SELECT assignment.team_id
        FROM public.workspace_invitation_teams AS assignment
        WHERE assignment.invitation_id = invitation.invitation_id
        ORDER BY assignment.team_id
    ) AS uuid[]) AS team_ids
FROM public.workspace_invitations AS invitation
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = invitation.workspace_id
WHERE lower(invitation.email) = lower(sqlc.arg(email))
  AND invitation.used_at IS NULL
  AND invitation.expires_at > sqlc.arg(now)
ORDER BY invitation.created_at DESC, invitation.invitation_id DESC;

-- name: RevokePendingInvitationsForEmail :many
WITH revoked AS (
    UPDATE public.workspace_invitations AS invitation
    SET
        used_at = sqlc.arg(revoked_at),
        updated_at = sqlc.arg(revoked_at)
    WHERE invitation.workspace_id = sqlc.arg(workspace_id)
      AND lower(invitation.email) = lower(sqlc.arg(email))
      AND invitation.used_at IS NULL
      AND invitation.expires_at > sqlc.arg(revoked_at)
    RETURNING invitation.invitation_id
), cancelled AS (
    UPDATE public.workspace_invitation_outbox AS outbox
    SET
        status = 'cancelled',
        next_attempt_at = NULL,
        claim_token = NULL,
        claimed_at = NULL,
        completed_at = sqlc.arg(revoked_at),
        last_error = NULL,
        updated_at = sqlc.arg(revoked_at)
    FROM revoked
    WHERE outbox.invitation_id = revoked.invitation_id
      AND outbox.event_type = 'invitation.email'
      AND outbox.status IN ('pending', 'retrying')
    RETURNING outbox.outbox_id
)
SELECT revoked.invitation_id
FROM revoked;

-- name: LockInvitationRecipient :exec
SELECT pg_advisory_xact_lock(
    hashtextextended(
        CAST(CAST(sqlc.arg(workspace_id) AS uuid) AS text)
        || chr(31)
        || lower(CAST(sqlc.arg(email) AS text)),
        0
    )
);

-- name: CreateInvitation :one
INSERT INTO public.workspace_invitations (
    invitation_id,
    workspace_id,
    inviter_id,
    email,
    role,
    token,
    token_digest,
    token_nonce,
    token_key_id,
    token_version,
    expires_at,
    created_at,
    updated_at
) VALUES (
    sqlc.arg(invitation_id),
    sqlc.arg(workspace_id),
    sqlc.arg(inviter_id),
    sqlc.arg(email),
    CAST(sqlc.arg(role) AS public.user_role),
    NULL,
    sqlc.arg(token_digest),
    sqlc.arg(token_nonce),
    sqlc.arg(token_key_id),
    sqlc.arg(token_version),
    sqlc.arg(expires_at),
    sqlc.arg(created_at),
    sqlc.arg(created_at)
)
RETURNING invitation_id, created_at, updated_at;

-- name: AddInvitationTeam :execrows
INSERT INTO public.workspace_invitation_teams (invitation_id, team_id)
SELECT sqlc.arg(invitation_id), team.team_id
FROM public.teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
ON CONFLICT (invitation_id, team_id) DO NOTHING;

-- name: RevokeInvitation :execrows
UPDATE public.workspace_invitations AS invitation
SET
    used_at = sqlc.arg(revoked_at),
    updated_at = sqlc.arg(revoked_at)
WHERE invitation.invitation_id = sqlc.arg(invitation_id)
  AND invitation.workspace_id = sqlc.arg(workspace_id)
  AND invitation.used_at IS NULL;

-- name: ActiveInviteeMatchesInvitation :one
SELECT EXISTS (
    SELECT 1
    FROM public.users AS invitee
    WHERE invitee.user_id = sqlc.arg(user_id)
      AND invitee.is_active = TRUE
      AND lower(invitee.email) = lower(sqlc.arg(email))
) AS matches;

-- name: WorkspaceMembershipExists :one
SELECT EXISTS (
    SELECT 1
    FROM public.workspace_members AS member
    WHERE member.workspace_id = sqlc.arg(workspace_id)
      AND member.user_id = sqlc.arg(user_id)
) AS exists;

-- name: AddWorkspaceMembership :execrows
INSERT INTO public.workspace_members (workspace_id, user_id, role, created_at, updated_at)
VALUES (
    sqlc.arg(workspace_id),
    sqlc.arg(user_id),
    CAST(sqlc.arg(role) AS public.user_role),
    sqlc.arg(accepted_at),
    sqlc.arg(accepted_at)
)
ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- name: AddInvitationTeamMemberships :execrows
INSERT INTO public.team_members (team_id, user_id, created_at, updated_at)
SELECT assignment.team_id, sqlc.arg(user_id), sqlc.arg(accepted_at), sqlc.arg(accepted_at)
FROM public.workspace_invitation_teams AS assignment
INNER JOIN public.teams AS team
    ON team.team_id = assignment.team_id
   AND team.workspace_id = sqlc.arg(workspace_id)
WHERE assignment.invitation_id = sqlc.arg(invitation_id)
ON CONFLICT (team_id, user_id) DO NOTHING;

-- name: UpdateInviteeLastWorkspace :execrows
UPDATE public.users AS invitee
SET
    last_used_workspace_id = sqlc.arg(workspace_id),
    updated_at = sqlc.arg(accepted_at)
WHERE invitee.user_id = sqlc.arg(user_id)
  AND invitee.is_active = TRUE
  AND EXISTS (
      SELECT 1
      FROM public.workspace_members AS member
      WHERE member.workspace_id = sqlc.arg(workspace_id)
        AND member.user_id = invitee.user_id
  );

-- name: ConsumeInvitation :execrows
UPDATE public.workspace_invitations AS invitation
SET
    used_at = sqlc.arg(accepted_at),
    updated_at = sqlc.arg(accepted_at)
WHERE invitation.invitation_id = sqlc.arg(invitation_id)
  AND invitation.used_at IS NULL;

-- name: GetInvitationAcceptedEventDetails :one
SELECT
    inviter.email AS inviter_email,
    COALESCE(NULLIF(inviter.full_name, ''), inviter.username) AS inviter_name,
    invitee.email AS invitee_email,
    COALESCE(NULLIF(invitee.full_name, ''), invitee.username) AS invitee_name,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug
FROM public.workspace_invitations AS invitation
INNER JOIN public.users AS inviter ON inviter.user_id = invitation.inviter_id
INNER JOIN public.users AS invitee ON invitee.user_id = sqlc.arg(invitee_id)
INNER JOIN public.workspaces AS workspace ON workspace.workspace_id = invitation.workspace_id
WHERE invitation.invitation_id = sqlc.arg(invitation_id);
