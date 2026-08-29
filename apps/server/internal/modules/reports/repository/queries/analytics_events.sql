-- name: CreateWorkspaceAnalyticsEvent :execrows
WITH actor_access AS (
    SELECT membership.role
    FROM workspace_members AS membership
    INNER JOIN users AS actor
        ON actor.user_id = membership.user_id
       AND actor.is_active = TRUE
       AND actor.is_system = FALSE
    INNER JOIN workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE membership.workspace_id = sqlc.arg(workspace_id)::uuid
      AND membership.user_id = sqlc.arg(actor_id)::uuid
      AND membership.role IN ('admin', 'member')
),
authorized_input AS (
    SELECT
        sqlc.arg(workspace_id)::uuid AS workspace_id,
        sqlc.arg(actor_id)::uuid AS user_id,
        sqlc.narg(team_id)::uuid AS team_id,
        sqlc.narg(story_id)::uuid AS story_id,
        sqlc.narg(objective_id)::uuid AS objective_id,
        sqlc.narg(sprint_id)::uuid AS sprint_id,
        sqlc.narg(key_result_id)::uuid AS key_result_id
    FROM actor_access
    WHERE (
          sqlc.narg(team_id)::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM teams AS team
              WHERE team.workspace_id = sqlc.arg(workspace_id)::uuid
                AND team.team_id = sqlc.narg(team_id)
                AND (
                    actor_access.role = 'admin'
                    OR team.is_private = FALSE
                    OR EXISTS (
                        SELECT 1
                        FROM team_members AS actor_team_membership
                        WHERE actor_team_membership.team_id = team.team_id
                          AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
                    )
                )
          )
      )
      AND (
          sqlc.narg(story_id)::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM stories AS story
              INNER JOIN teams AS team
                  ON team.team_id = story.team_id
                 AND team.workspace_id = story.workspace_id
              WHERE story.workspace_id = sqlc.arg(workspace_id)::uuid
                AND story.id = sqlc.narg(story_id)
                AND story.deleted_at IS NULL
                AND (
                    actor_access.role = 'admin'
                    OR team.is_private = FALSE
                    OR EXISTS (
                        SELECT 1
                        FROM team_members AS actor_team_membership
                        WHERE actor_team_membership.team_id = team.team_id
                          AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
                    )
                )
          )
      )
      AND (
          sqlc.narg(objective_id)::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM objectives AS objective
              LEFT JOIN teams AS team
                  ON team.team_id = objective.team_id
                 AND team.workspace_id = objective.workspace_id
              WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
                AND objective.objective_id = sqlc.narg(objective_id)
                AND (
                    objective.team_id IS NULL
                    OR (
                        team.team_id IS NOT NULL
                        AND (
                            actor_access.role = 'admin'
                            OR team.is_private = FALSE
                            OR EXISTS (
                                SELECT 1
                                FROM team_members AS actor_team_membership
                                WHERE actor_team_membership.team_id = objective.team_id
                                  AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
                            )
                        )
                    )
                )
          )
      )
      AND (
          sqlc.narg(sprint_id)::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM sprints AS sprint
              INNER JOIN teams AS team
                  ON team.team_id = sprint.team_id
                 AND team.workspace_id = sprint.workspace_id
              WHERE sprint.workspace_id = sqlc.arg(workspace_id)::uuid
                AND sprint.sprint_id = sqlc.narg(sprint_id)
                AND (
                    actor_access.role = 'admin'
                    OR team.is_private = FALSE
                    OR EXISTS (
                        SELECT 1
                        FROM team_members AS actor_team_membership
                        WHERE actor_team_membership.team_id = team.team_id
                          AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
                    )
                )
          )
      )
      AND (
          sqlc.narg(key_result_id)::uuid IS NULL
          OR EXISTS (
              SELECT 1
              FROM key_results AS key_result
              INNER JOIN objectives AS objective ON objective.objective_id = key_result.objective_id
              LEFT JOIN teams AS team
                  ON team.team_id = objective.team_id
                 AND team.workspace_id = objective.workspace_id
              WHERE objective.workspace_id = sqlc.arg(workspace_id)::uuid
                AND key_result.id = sqlc.narg(key_result_id)
                AND (
                    objective.team_id IS NULL
                    OR (
                        team.team_id IS NOT NULL
                        AND (
                            actor_access.role = 'admin'
                            OR team.is_private = FALSE
                            OR EXISTS (
                                SELECT 1
                                FROM team_members AS actor_team_membership
                                WHERE actor_team_membership.team_id = objective.team_id
                                  AND actor_team_membership.user_id = sqlc.arg(actor_id)::uuid
                            )
                        )
                    )
                )
          )
      )
)
INSERT INTO workspace_analytics_events (
    workspace_id,
    user_id,
    team_id,
    story_id,
    objective_id,
    sprint_id,
    key_result_id,
    event_name,
    surface,
    properties,
    occurred_at
)
SELECT
    authorized_input.workspace_id,
    authorized_input.user_id,
    authorized_input.team_id,
    authorized_input.story_id,
    authorized_input.objective_id,
    authorized_input.sprint_id,
    authorized_input.key_result_id,
    sqlc.arg(event_name),
    sqlc.arg(surface),
    CAST(sqlc.arg(properties) AS jsonb),
    sqlc.arg(occurred_at)
FROM authorized_input;
