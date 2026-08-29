-- name: ListTeamsForActor :many
SELECT
    team.team_id,
    team.name,
    team.code,
    team.color,
    team.is_private,
    team.workspace_id,
    team.created_at,
    team.updated_at,
    CAST((
        SELECT COUNT(*)
        FROM team_members AS counted_member
        WHERE counted_member.team_id = team.team_id
    ) AS integer) AS member_count,
    COALESCE(sprint_settings.auto_create_sprints, FALSE) AS sprints_enabled
FROM teams AS team
LEFT JOIN user_team_orders AS team_order
    ON team_order.team_id = team.team_id
   AND team_order.user_id = sqlc.arg(actor_id)
   AND team_order.workspace_id = team.workspace_id
LEFT JOIN team_sprint_settings AS sprint_settings
    ON sprint_settings.team_id = team.team_id
   AND sprint_settings.workspace_id = team.workspace_id
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM workspace_members AS actor_membership
      INNER JOIN users AS actor ON actor.user_id = actor_membership.user_id
      WHERE actor_membership.workspace_id = team.workspace_id
        AND actor_membership.user_id = sqlc.arg(actor_id)
        AND actor.is_active = TRUE
  )
  AND (
      EXISTS (
          SELECT 1
          FROM team_members AS actor_team_membership
          WHERE actor_team_membership.team_id = team.team_id
            AND actor_team_membership.user_id = sqlc.arg(actor_id)
      )
      OR (
          CAST(sqlc.arg(joined_only) AS boolean) = FALSE
          AND EXISTS (
              SELECT 1
              FROM workspace_members AS admin_membership
              WHERE admin_membership.workspace_id = team.workspace_id
                AND admin_membership.user_id = sqlc.arg(actor_id)
                AND admin_membership.role = 'admin'
          )
      )
  )
  AND (
      CAST(sqlc.arg(search) AS text) = ''
      OR team.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
      OR team.code ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
  )
ORDER BY
    CASE WHEN team_order.order_index IS NULL THEN 1 ELSE 0 END,
    team_order.order_index ASC NULLS LAST,
    team.created_at DESC,
    team.team_id DESC
LIMIT CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_limit) AS integer)
    ELSE NULL
END
OFFSET CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_offset) AS integer)
    ELSE 0
END;

-- name: GetTeamForActor :one
SELECT
    team.team_id,
    team.name,
    team.code,
    team.color,
    team.is_private,
    team.workspace_id,
    team.created_at,
    team.updated_at,
    CAST((
        SELECT COUNT(*)
        FROM team_members AS counted_member
        WHERE counted_member.team_id = team.team_id
    ) AS integer) AS member_count,
    COALESCE(sprint_settings.auto_create_sprints, FALSE) AS sprints_enabled
FROM teams AS team
LEFT JOIN team_sprint_settings AS sprint_settings
    ON sprint_settings.team_id = team.team_id
   AND sprint_settings.workspace_id = team.workspace_id
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
  AND EXISTS (
      SELECT 1
      FROM workspace_members AS actor_membership
      INNER JOIN users AS actor ON actor.user_id = actor_membership.user_id
      WHERE actor_membership.workspace_id = team.workspace_id
        AND actor_membership.user_id = sqlc.arg(actor_id)
        AND actor.is_active = TRUE
  )
  AND (
      EXISTS (
          SELECT 1
          FROM team_members AS actor_team_membership
          WHERE actor_team_membership.team_id = team.team_id
            AND actor_team_membership.user_id = sqlc.arg(actor_id)
      )
      OR EXISTS (
          SELECT 1
          FROM workspace_members AS admin_membership
          WHERE admin_membership.workspace_id = team.workspace_id
            AND admin_membership.user_id = sqlc.arg(actor_id)
            AND admin_membership.role = 'admin'
      )
  );

-- name: ListPublicTeamsForActor :many
SELECT
    team.team_id,
    team.name,
    team.code,
    team.color,
    team.is_private,
    team.workspace_id,
    team.created_at,
    team.updated_at,
    CAST((
        SELECT COUNT(*)
        FROM team_members AS counted_member
        WHERE counted_member.team_id = team.team_id
    ) AS integer) AS member_count,
    COALESCE(sprint_settings.auto_create_sprints, FALSE) AS sprints_enabled
FROM teams AS team
LEFT JOIN team_sprint_settings AS sprint_settings
    ON sprint_settings.team_id = team.team_id
   AND sprint_settings.workspace_id = team.workspace_id
WHERE team.workspace_id = sqlc.arg(workspace_id)
  AND team.is_private = FALSE
  AND EXISTS (
      SELECT 1
      FROM workspace_members AS actor_membership
      INNER JOIN users AS actor ON actor.user_id = actor_membership.user_id
      WHERE actor_membership.workspace_id = team.workspace_id
        AND actor_membership.user_id = sqlc.arg(actor_id)
        AND actor.is_active = TRUE
  )
  AND NOT EXISTS (
      SELECT 1
      FROM team_members AS actor_team_membership
      WHERE actor_team_membership.team_id = team.team_id
        AND actor_team_membership.user_id = sqlc.arg(actor_id)
  )
  AND (
      CAST(sqlc.arg(search) AS text) = ''
      OR team.name ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
      OR team.code ILIKE '%' || CAST(sqlc.arg(search) AS text) || '%'
  )
ORDER BY team.created_at DESC, team.team_id DESC
LIMIT CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_limit) AS integer)
    ELSE NULL
END
OFFSET CASE
    WHEN CAST(sqlc.arg(page_limit) AS integer) > 0
        THEN CAST(sqlc.arg(page_offset) AS integer)
    ELSE 0
END;

-- name: CreateTeam :one
INSERT INTO teams (
    name,
    code,
    color,
    is_private,
    workspace_id
)
VALUES (
    CAST(sqlc.arg(name) AS text),
    CAST(sqlc.arg(code) AS text),
    CAST(sqlc.arg(color) AS text),
    sqlc.arg(is_private),
    sqlc.arg(workspace_id)
)
RETURNING
    team_id,
    name,
    code,
    color,
    is_private,
    workspace_id,
    created_at,
    updated_at;

-- name: CreateDefaultStoryAutomationSettings :execrows
INSERT INTO team_story_automation_settings (
    team_id,
    workspace_id,
    auto_close_inactive_enabled,
    auto_close_inactive_months,
    auto_archive_enabled,
    auto_archive_months
)
SELECT
    team.team_id,
    team.workspace_id,
    TRUE,
    3,
    TRUE,
    3
FROM teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id);

-- name: CreateDefaultStoryStatus :execrows
INSERT INTO statuses (
    name,
    category,
    order_index,
    color,
    team_id,
    workspace_id
)
SELECT
    CAST(sqlc.arg(name) AS text),
    CAST(sqlc.arg(category) AS text),
    CAST(sqlc.arg(order_index) AS integer),
    CAST(sqlc.arg(color) AS text),
    team.team_id,
    team.workspace_id
FROM teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id);

-- name: UpdateTeamForWorkspace :one
UPDATE teams AS team
SET
    name = CASE
        WHEN CAST(sqlc.arg(name) AS text) = '' THEN team.name
        ELSE CAST(sqlc.arg(name) AS text)
    END,
    code = CASE
        WHEN CAST(sqlc.arg(code) AS text) = '' THEN team.code
        ELSE CAST(sqlc.arg(code) AS text)
    END,
    color = CASE
        WHEN CAST(sqlc.arg(color) AS text) = '' THEN team.color
        ELSE CAST(sqlc.arg(color) AS text)
    END,
    is_private = sqlc.arg(is_private),
    updated_at = CURRENT_TIMESTAMP
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id)
RETURNING
    team_id,
    name,
    code,
    color,
    is_private,
    workspace_id,
    created_at,
    updated_at;

-- name: DeleteTeamForWorkspace :execrows
DELETE FROM teams AS team
WHERE team.team_id = sqlc.arg(team_id)
  AND team.workspace_id = sqlc.arg(workspace_id);
