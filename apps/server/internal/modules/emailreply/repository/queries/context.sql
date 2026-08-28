-- name: StoryStatusExists :one
SELECT EXISTS (
    SELECT 1
    FROM statuses
    WHERE workspace_id = sqlc.arg(workspace_id)
      AND team_id = sqlc.arg(team_id)
      AND status_id = sqlc.arg(status_id)
);

-- name: StoryAssigneeExists :one
SELECT EXISTS (
    SELECT 1
    FROM team_members AS membership
    INNER JOIN users AS actor ON actor.user_id = membership.user_id
    WHERE membership.team_id = sqlc.arg(team_id)
      AND membership.user_id = sqlc.arg(user_id)
      AND actor.is_active = true
      AND actor.is_system = false
);

-- name: GetObjectiveVersion :one
SELECT updated_at
FROM objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective_id = sqlc.arg(entity_id);

-- name: GetKeyResultVersion :one
SELECT result.updated_at
FROM key_results AS result
INNER JOIN objectives AS objective ON objective.objective_id = result.objective_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND result.id = sqlc.arg(entity_id);

-- name: GetStoryVersion :one
SELECT updated_at
FROM stories
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND id = sqlc.arg(entity_id)
  AND deleted_at IS NULL;

-- name: GetFeedbackVersion :one
SELECT updated_at
FROM feedback_items
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND id = sqlc.arg(entity_id)
  AND deleted_at IS NULL;

-- name: GetObjectiveHealth :one
SELECT CAST(COALESCE(CAST(health AS text), '') AS text) AS health
FROM objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective_id = sqlc.arg(entity_id);

-- name: GetKeyResultCurrentValue :one
SELECT CAST(result.current_value AS double precision) AS current_value
FROM key_results AS result
INNER JOIN objectives AS objective ON objective.objective_id = result.objective_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND result.id = sqlc.arg(entity_id);

-- name: GetFeedbackStatus :one
SELECT status
FROM feedback_items
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(entity_id)
  AND deleted_at IS NULL;

-- name: GetStoryReconciliationState :one
SELECT status_id, assignee_id, end_date
FROM stories
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(entity_id)
  AND deleted_at IS NULL;

-- name: GetEmailReplyActorWorkspace :one
SELECT workspace.slug, member.role
FROM workspaces AS workspace
INNER JOIN workspace_members AS member
    ON member.workspace_id = workspace.workspace_id
    AND member.user_id = sqlc.arg(user_id)
INNER JOIN users AS actor ON actor.user_id = member.user_id
WHERE workspace.workspace_id = sqlc.arg(workspace_id)
  AND workspace.deleted_at IS NULL
  AND actor.is_active = true
  AND actor.is_system = false
  AND member.role IN ('admin', 'member', 'guest');

-- name: ListEmailReplyActorTeams :many
SELECT team.team_id
FROM teams AS team
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.arg(actor_role) AS text) = 'admin'
      OR EXISTS (
          SELECT 1
          FROM team_members AS membership
          WHERE membership.team_id = team.team_id
            AND membership.user_id = sqlc.arg(user_id)
      )
  )
ORDER BY team.team_id;

-- name: GetEmailReplyObjectiveTarget :one
SELECT objective_id AS id,
       team_id,
       name,
       CAST(COALESCE(CAST(health AS text), '') AS text) AS health,
       start_date,
       end_date,
       updated_at
FROM objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective_id = sqlc.arg(entity_id);

-- name: GetEmailReplyKeyResultTarget :one
SELECT result.id,
       objective.team_id,
       result.name,
       result.measurement_type,
       CAST(result.current_value AS double precision) AS current_value,
       CAST(result.target_value AS double precision) AS target_value,
       result.start_date,
       result.end_date,
       result.updated_at
FROM key_results AS result
INNER JOIN objectives AS objective ON objective.objective_id = result.objective_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND result.id = sqlc.arg(entity_id);

-- name: GetEmailReplyStoryTarget :one
SELECT story.id,
       story.team_id,
       story.title,
       state.name AS status_name,
       COALESCE(actor.full_name, actor.email, '') AS assignee_name,
       story.end_date,
       story.updated_at
FROM stories AS story
LEFT JOIN statuses AS state ON state.status_id = story.status_id
LEFT JOIN users AS actor ON actor.user_id = story.assignee_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.id = sqlc.arg(entity_id)
  AND story.deleted_at IS NULL;

-- name: GetEmailReplyFeedbackTarget :one
SELECT item.id,
       board.team_id,
       item.title,
       item.status,
       item.updated_at
FROM feedback_items AS item
INNER JOIN feedback_boards AS board
    ON board.id = item.board_id
    AND board.workspace_id = item.workspace_id
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(entity_id)
  AND item.deleted_at IS NULL;

-- name: ListEmailReplyStoryStatuses :many
SELECT status_id AS id, name
FROM statuses
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND team_id = sqlc.arg(team_id)
ORDER BY order_index, status_id;

-- name: ListEmailReplyStoryAssignees :many
SELECT actor.user_id AS id,
       COALESCE(NULLIF(actor.full_name, ''), actor.email) AS name
FROM team_members AS member
INNER JOIN users AS actor ON actor.user_id = member.user_id
WHERE member.team_id = sqlc.arg(team_id)
  AND actor.is_active = true
  AND actor.is_system = false
ORDER BY name, actor.user_id
LIMIT sqlc.arg(row_limit);

-- name: GetObjectiveTeam :one
SELECT team_id
FROM objectives
WHERE workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND objective_id = sqlc.arg(entity_id);

-- name: GetKeyResultTeam :one
SELECT objective.team_id
FROM key_results AS result
INNER JOIN objectives AS objective ON objective.objective_id = result.objective_id
WHERE objective.workspace_id = CAST(sqlc.arg(workspace_id) AS uuid)
  AND result.id = sqlc.arg(entity_id);

-- name: GetStoryTeam :one
SELECT team_id
FROM stories
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id = sqlc.arg(entity_id)
  AND deleted_at IS NULL;

-- name: GetFeedbackTeam :one
SELECT board.team_id
FROM feedback_items AS item
INNER JOIN feedback_boards AS board
    ON board.id = item.board_id
    AND board.workspace_id = item.workspace_id
WHERE item.workspace_id = sqlc.arg(workspace_id)
  AND item.id = sqlc.arg(entity_id)
  AND item.deleted_at IS NULL;
