-- name: AddTeamMemberForWorkspace :one
WITH eligible_member AS (
    SELECT team.team_id, workspace_membership.user_id
    FROM teams AS team
    INNER JOIN workspace_members AS workspace_membership
        ON workspace_membership.workspace_id = team.workspace_id
       AND workspace_membership.user_id = sqlc.arg(user_id)
    INNER JOIN users AS member
        ON member.user_id = workspace_membership.user_id
       AND member.is_active = TRUE
    WHERE team.team_id = sqlc.arg(team_id)
      AND team.workspace_id = sqlc.arg(workspace_id)
), inserted_member AS (
    INSERT INTO team_members (team_id, user_id, created_at, updated_at)
    SELECT eligible_member.team_id, eligible_member.user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    FROM eligible_member
    ON CONFLICT (team_id, user_id) DO NOTHING
    RETURNING team_id
)
SELECT
    EXISTS(SELECT 1 FROM eligible_member) AS eligible,
    EXISTS(SELECT 1 FROM inserted_member) AS added;

-- name: JoinPublicTeamForActor :one
WITH eligible_team AS (
    SELECT team.team_id, actor_membership.user_id
    FROM teams AS team
    INNER JOIN workspace_members AS actor_membership
        ON actor_membership.workspace_id = team.workspace_id
       AND actor_membership.user_id = sqlc.arg(actor_id)
    INNER JOIN users AS actor
        ON actor.user_id = actor_membership.user_id
       AND actor.is_active = TRUE
    WHERE team.team_id = sqlc.arg(team_id)
      AND team.workspace_id = sqlc.arg(workspace_id)
      AND team.is_private = FALSE
), inserted_member AS (
    INSERT INTO team_members (team_id, user_id, created_at, updated_at)
    SELECT eligible_team.team_id, eligible_team.user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    FROM eligible_team
    ON CONFLICT (team_id, user_id) DO NOTHING
    RETURNING team_id
)
SELECT
    EXISTS(SELECT 1 FROM eligible_team) AS eligible,
    EXISTS(SELECT 1 FROM inserted_member) AS joined;

-- name: RemoveTeamMemberForWorkspace :execrows
DELETE FROM team_members AS team_membership
USING teams AS team
WHERE team_membership.team_id = team.team_id
  AND team_membership.team_id = sqlc.arg(team_id)
  AND team_membership.user_id = sqlc.arg(user_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM workspace_members AS target_workspace_membership
      INNER JOIN users AS target_user
          ON target_user.user_id = target_workspace_membership.user_id
         AND target_user.is_active = TRUE
      WHERE target_workspace_membership.workspace_id = team.workspace_id
        AND target_workspace_membership.user_id = team_membership.user_id
  );

-- name: LeaveTeamForActor :execrows
DELETE FROM team_members AS team_membership
USING teams AS team, workspace_members AS actor_membership, users AS actor
WHERE team_membership.team_id = team.team_id
  AND team_membership.team_id = sqlc.arg(team_id)
  AND team_membership.user_id = sqlc.arg(actor_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND actor_membership.workspace_id = team.workspace_id
  AND actor_membership.user_id = sqlc.arg(actor_id)
  AND actor.user_id = actor_membership.user_id
  AND actor.is_active = TRUE;

-- name: UpdateTeamMemberAIContextForWorkspace :execrows
UPDATE team_members AS team_membership
SET
    ai_role_title = CAST(sqlc.arg(ai_role_title) AS text),
    ai_role_description = CAST(sqlc.arg(ai_role_description) AS text),
    updated_at = CURRENT_TIMESTAMP
FROM teams AS team
WHERE team_membership.team_id = team.team_id
  AND team_membership.team_id = sqlc.arg(team_id)
  AND team_membership.user_id = sqlc.arg(user_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM workspace_members AS target_workspace_membership
      INNER JOIN users AS target_user
          ON target_user.user_id = target_workspace_membership.user_id
         AND target_user.is_active = TRUE
      WHERE target_workspace_membership.workspace_id = team.workspace_id
        AND target_workspace_membership.user_id = team_membership.user_id
  );
