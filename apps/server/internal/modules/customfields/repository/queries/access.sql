-- name: AuthorizeTeam :one
SELECT member.role
FROM teams AS team
JOIN workspaces AS workspace ON workspace.workspace_id = team.workspace_id
JOIN workspace_members AS member ON member.workspace_id = team.workspace_id AND member.user_id = sqlc.arg(actor_id)
JOIN users AS actor ON actor.user_id = member.user_id
WHERE team.team_id = sqlc.arg(team_id) AND team.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL AND actor.is_active = TRUE
  AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM team_members AS joined WHERE joined.team_id = team.team_id AND joined.user_id = actor.user_id));

-- name: AuthorizeTeamMutation :one
SELECT member.role
FROM teams AS team
JOIN workspaces AS workspace ON workspace.workspace_id = team.workspace_id
JOIN workspace_members AS member ON member.workspace_id = team.workspace_id AND member.user_id = sqlc.arg(actor_id)
JOIN users AS actor ON actor.user_id = member.user_id
WHERE team.team_id = sqlc.arg(team_id) AND team.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL AND actor.is_active = TRUE
  AND (member.role = 'admin' OR EXISTS (SELECT 1 FROM team_members AS joined WHERE joined.team_id = team.team_id AND joined.user_id = actor.user_id))
FOR UPDATE OF team FOR SHARE OF workspace, member, actor;

-- name: LockActorTeamMembership :one
SELECT joined.user_id FROM team_members AS joined
WHERE joined.team_id = sqlc.arg(team_id) AND joined.user_id = sqlc.arg(actor_id)
FOR SHARE;

-- name: GetStoryScope :one
SELECT story.team_id, story.custom_fields_version, story.archived_at
FROM stories AS story
WHERE story.id = sqlc.arg(story_id) AND story.workspace_id = sqlc.arg(workspace_id) AND story.deleted_at IS NULL;

-- name: LockStoryScope :one
SELECT story.team_id, story.custom_fields_version, story.archived_at
FROM stories AS story
WHERE story.id = sqlc.arg(story_id) AND story.workspace_id = sqlc.arg(workspace_id) AND story.deleted_at IS NULL
FOR UPDATE;

-- name: AdvanceStoryCustomFieldVersion :one
UPDATE stories SET custom_fields_version = custom_fields_version + 1
WHERE id = sqlc.arg(story_id) AND workspace_id = sqlc.arg(workspace_id) AND team_id = sqlc.arg(team_id)
RETURNING custom_fields_version;

-- name: ValidatePerson :one
SELECT actor.user_id
FROM users AS actor
JOIN workspace_members AS member ON member.user_id = actor.user_id AND member.workspace_id = sqlc.arg(workspace_id)
JOIN team_members AS joined ON joined.user_id = actor.user_id AND joined.team_id = sqlc.arg(team_id)
WHERE actor.user_id = sqlc.arg(person_id) AND actor.is_active = TRUE
FOR SHARE OF actor, member, joined;
