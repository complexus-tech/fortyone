-- Ordinary user-facing story reads require an active actor who is a current
-- member of both the workspace and the story's team. The integration-only
-- repository method can replace membership with a pre-authorized, restricted
-- credential team scope; it never permits unrestricted credential reads.

-- name: GetVisibleStory :one
SELECT
    story.id,
    story.sequence_id,
    story.title,
    team.code AS team_code,
    team.name AS team_name,
    story.description,
    story.description_html,
    story.parent_id,
    story.objective_id,
    objective.name AS objective_name,
    objective.description AS objective_description,
    story.workspace_id,
    story.team_id,
    story.status_id,
    story.assignee_id,
    story.blocked_by_id,
    story.blocking_id,
    story.related_id,
    story.reporter_id,
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    story.sprint_id,
    sprint.name AS sprint_name,
    sprint.goal AS sprint_goal,
    sprint.start_date AS sprint_start_date,
    sprint.end_date AS sprint_end_date,
    story.key_result_id,
    story.estimate_unit,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.start_date,
    story.end_date,
    story.created_at,
    story.updated_at,
    story.deleted_at,
    story.archived_at,
    story.completed_at,
    story.external_creation_key,
    CAST(ARRAY(
        SELECT story_label.label_id
        FROM story_labels AS story_label
        WHERE story_label.story_id = story.id
        ORDER BY story_label.label_id
    ) AS uuid[]) AS label_ids,
    CAST(ARRAY(
        SELECT collaborator.user_id
        FROM story_collaborators AS collaborator
        WHERE collaborator.story_id = story.id
        ORDER BY collaborator.created_at, collaborator.user_id
    ) AS uuid[]) AS collaborator_ids,
    CAST(ARRAY(
        SELECT audience.user_id
        FROM (
            SELECT story.assignee_id AS user_id
            WHERE story.assignee_id IS NOT NULL
            UNION
            SELECT collaborator.user_id
            FROM story_collaborators AS collaborator
            WHERE collaborator.story_id = story.id
            UNION
            SELECT watcher.user_id
            FROM story_watchers AS watcher
            WHERE watcher.story_id = story.id
        ) AS audience
        INNER JOIN users AS audience_user
            ON audience_user.user_id = audience.user_id
           AND audience_user.is_active = TRUE
           AND audience_user.is_system = FALSE
        WHERE NOT EXISTS (
            SELECT 1
            FROM story_notification_mutes AS muted
            WHERE muted.story_id = story.id
              AND muted.user_id = audience.user_id
        )
        ORDER BY audience.user_id
    ) AS uuid[]) AS watcher_ids,
    CAST((
        SELECT COUNT(*)
        FROM (
            SELECT story.assignee_id AS user_id
            WHERE story.assignee_id IS NOT NULL
            UNION
            SELECT collaborator.user_id
            FROM story_collaborators AS collaborator
            WHERE collaborator.story_id = story.id
            UNION
            SELECT watcher.user_id
            FROM story_watchers AS watcher
            WHERE watcher.story_id = story.id
        ) AS audience
        INNER JOIN users AS audience_user
            ON audience_user.user_id = audience.user_id
           AND audience_user.is_active = TRUE
           AND audience_user.is_system = FALSE
        WHERE NOT EXISTS (
            SELECT 1
            FROM story_notification_mutes AS muted
            WHERE muted.story_id = story.id
              AND muted.user_id = audience.user_id
        )
    ) AS integer) AS watcher_count,
    (
        NOT EXISTS (
            SELECT 1
            FROM story_notification_mutes AS muted
            WHERE muted.story_id = story.id
              AND muted.user_id = sqlc.arg(actor_id)
        )
        AND (
            story.assignee_id = sqlc.arg(actor_id)
            OR EXISTS (
                SELECT 1
                FROM story_collaborators AS collaborator
                WHERE collaborator.story_id = story.id
                  AND collaborator.user_id = sqlc.arg(actor_id)
            )
            OR EXISTS (
                SELECT 1
                FROM story_watchers AS watcher
                WHERE watcher.story_id = story.id
                  AND watcher.user_id = sqlc.arg(actor_id)
            )
        )
    ) AS is_watching,
    CAST(COALESCE(CASE
        WHEN story.assignee_id = sqlc.arg(actor_id) THEN 'assignee'
        WHEN EXISTS (
            SELECT 1
            FROM story_collaborators AS collaborator
            WHERE collaborator.story_id = story.id
              AND collaborator.user_id = sqlc.arg(actor_id)
        ) THEN 'collaborator'
        WHEN EXISTS (
            SELECT 1
            FROM story_watchers AS watcher
            WHERE watcher.story_id = story.id
              AND watcher.user_id = sqlc.arg(actor_id)
        ) THEN 'watcher'
        ELSE NULL
    END, '') AS text) AS watching_reason
FROM stories AS story
INNER JOIN teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
LEFT JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
LEFT JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
LEFT JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
LEFT JOIN objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
LEFT JOIN sprints AS sprint
    ON sprint.sprint_id = story.sprint_id
   AND sprint.workspace_id = story.workspace_id
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = story.team_id
   AND estimation.workspace_id = story.workspace_id
WHERE story.id = sqlc.arg(story_id)
  AND story.workspace_id = sqlc.arg(workspace_id)
  AND (
      CAST(sqlc.arg(bypass_actor_membership) AS boolean)
      OR (
          actor.user_id IS NOT NULL
          AND workspace_member.user_id IS NOT NULL
          AND team_member.user_id IS NOT NULL
      )
  )
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  );

-- name: GetVisibleStoryIDByRef :one
SELECT story.id
FROM stories AS story
INNER JOIN teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
LEFT JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
LEFT JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
LEFT JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND UPPER(team.code) = UPPER(CAST(sqlc.arg(team_code) AS text))
  AND story.sequence_id = sqlc.arg(sequence_id)
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(bypass_actor_membership) AS boolean)
      OR (
          actor.user_id IS NOT NULL
          AND workspace_member.user_id IS NOT NULL
          AND team_member.user_id IS NOT NULL
      )
  )
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY story.id
LIMIT 1;

-- name: ListMyVisibleStories :many
SELECT
    story.id,
    story.sequence_id,
    story.title,
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    story.estimate_unit,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.parent_id,
    story.objective_id,
    objective.name AS objective_name,
    objective.description AS objective_description,
    story.sprint_id,
    sprint.name AS sprint_name,
    sprint.goal AS sprint_goal,
    sprint.start_date AS sprint_start_date,
    sprint.end_date AS sprint_end_date,
    story.team_id,
    team.code AS team_code,
    team.name AS team_name,
    story.workspace_id,
    story.status_id,
    story.assignee_id,
    CAST((SELECT COUNT(*) FROM story_collaborators AS collaborator WHERE collaborator.story_id = story.id) AS integer) AS collaborator_count,
    story.reporter_id,
    story.key_result_id,
    story.start_date,
    story.end_date,
    story.created_at,
    story.updated_at,
    story.completed_at,
    story.deleted_at,
    story.archived_at,
    CAST(ARRAY(
        SELECT story_label.label_id
        FROM story_labels AS story_label
        WHERE story_label.story_id = story.id
        ORDER BY story_label.label_id
    ) AS uuid[]) AS label_ids
FROM stories AS story
INNER JOIN teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
INNER JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
LEFT JOIN objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
LEFT JOIN sprints AS sprint
    ON sprint.sprint_id = story.sprint_id
   AND sprint.workspace_id = story.workspace_id
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = story.team_id
   AND estimation.workspace_id = story.workspace_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.parent_id IS NULL
  AND (
      story.assignee_id = actor.user_id
      OR story.reporter_id = actor.user_id
      OR EXISTS (
          SELECT 1
          FROM story_collaborators AS collaborator
          WHERE collaborator.story_id = story.id
            AND collaborator.user_id = actor.user_id
      )
  )
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY story.created_at DESC, story.id DESC
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: ListVisibleStoriesByCategory :many
SELECT
    story.id,
    story.sequence_id,
    story.title,
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    story.estimate_unit,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.parent_id,
    story.objective_id,
    objective.name AS objective_name,
    objective.description AS objective_description,
    story.sprint_id,
    sprint.name AS sprint_name,
    sprint.goal AS sprint_goal,
    sprint.start_date AS sprint_start_date,
    sprint.end_date AS sprint_end_date,
    story.team_id,
    team.code AS team_code,
    team.name AS team_name,
    story.workspace_id,
    story.status_id,
    story.assignee_id,
    CAST((SELECT COUNT(*) FROM story_collaborators AS collaborator WHERE collaborator.story_id = story.id) AS integer) AS collaborator_count,
    story.reporter_id,
    story.key_result_id,
    story.start_date,
    story.end_date,
    story.created_at,
    story.updated_at,
    story.completed_at,
    story.deleted_at,
    story.archived_at,
    CAST(ARRAY(
        SELECT story_label.label_id
        FROM story_labels AS story_label
        WHERE story_label.story_id = story.id
        ORDER BY story_label.label_id
    ) AS uuid[]) AS label_ids
FROM stories AS story
INNER JOIN teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
INNER JOIN statuses AS status
    ON status.status_id = story.status_id
   AND status.team_id = story.team_id
   AND status.workspace_id = story.workspace_id
INNER JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
LEFT JOIN objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
LEFT JOIN sprints AS sprint
    ON sprint.sprint_id = story.sprint_id
   AND sprint.workspace_id = story.workspace_id
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = story.team_id
   AND estimation.workspace_id = story.workspace_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.team_id = sqlc.arg(team_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND (CAST(sqlc.arg(include_sub_stories) AS boolean) OR story.parent_id IS NULL)
  AND status.category = sqlc.arg(category)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY story.created_at DESC, story.id DESC
LIMIT CAST(sqlc.arg(page_limit) AS integer)
OFFSET CAST(sqlc.arg(page_offset) AS integer);

-- name: ListVisibleSubStories :many
SELECT
    story.id,
    story.sequence_id,
    story.title,
    CAST(COALESCE(story.priority, 'No Priority') AS text) AS priority,
    story.estimate_unit,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS estimate_scheme,
    story.estimated_duration_minutes,
    story.minimum_focus_block_minutes,
    story.auto_scheduling_enabled,
    story.auto_scheduling_locked,
    story.auto_scheduling_status,
    story.auto_scheduling_reason,
    story.auto_scheduling_updated_at,
    story.parent_id,
    story.objective_id,
    objective.name AS objective_name,
    objective.description AS objective_description,
    story.sprint_id,
    sprint.name AS sprint_name,
    sprint.goal AS sprint_goal,
    sprint.start_date AS sprint_start_date,
    sprint.end_date AS sprint_end_date,
    story.team_id,
    team.code AS team_code,
    team.name AS team_name,
    story.workspace_id,
    story.status_id,
    story.assignee_id,
    CAST((SELECT COUNT(*) FROM story_collaborators AS collaborator WHERE collaborator.story_id = story.id) AS integer) AS collaborator_count,
    story.reporter_id,
    story.key_result_id,
    story.start_date,
    story.end_date,
    story.created_at,
    story.updated_at,
    story.completed_at,
    story.deleted_at,
    story.archived_at,
    CAST(ARRAY(
        SELECT story_label.label_id
        FROM story_labels AS story_label
        WHERE story_label.story_id = story.id
        ORDER BY story_label.label_id
    ) AS uuid[]) AS label_ids
FROM stories AS story
INNER JOIN teams AS team
    ON team.team_id = story.team_id
   AND team.workspace_id = story.workspace_id
INNER JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = story.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN team_members AS team_member
    ON team_member.team_id = story.team_id
   AND team_member.user_id = actor.user_id
LEFT JOIN objectives AS objective
    ON objective.objective_id = story.objective_id
   AND objective.workspace_id = story.workspace_id
LEFT JOIN sprints AS sprint
    ON sprint.sprint_id = story.sprint_id
   AND sprint.workspace_id = story.workspace_id
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = story.team_id
   AND estimation.workspace_id = story.workspace_id
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.parent_id = ANY(CAST(sqlc.arg(parent_ids) AS uuid[]))
  AND story.deleted_at IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  )
ORDER BY story.parent_id, story.created_at, story.id;

-- name: ListVisibleStoryAssociations :many
SELECT
    association.id,
    association.from_story_id,
    association.to_story_id,
    CAST(association.association_type AS text) AS association_type,
    related.id AS related_id,
    related.sequence_id AS related_sequence_id,
    related.title AS related_title,
    CAST(COALESCE(related.priority, 'No Priority') AS text) AS related_priority,
    related.estimate_unit AS related_estimate_unit,
    CAST(COALESCE(estimation.scheme, 'tshirt') AS text) AS related_estimate_scheme,
    related.estimated_duration_minutes AS related_estimated_duration_minutes,
    related.minimum_focus_block_minutes AS related_minimum_focus_block_minutes,
    related.auto_scheduling_enabled AS related_auto_scheduling_enabled,
    related.auto_scheduling_locked AS related_auto_scheduling_locked,
    related.auto_scheduling_status AS related_auto_scheduling_status,
    related.auto_scheduling_reason AS related_auto_scheduling_reason,
    related.auto_scheduling_updated_at AS related_auto_scheduling_updated_at,
    related.parent_id AS related_parent_id,
    related.objective_id AS related_objective_id,
    related.sprint_id AS related_sprint_id,
    related.team_id AS related_team_id,
    related.workspace_id AS related_workspace_id,
    related.status_id AS related_status_id,
    related.assignee_id AS related_assignee_id,
    CAST((SELECT COUNT(*) FROM story_collaborators AS collaborator WHERE collaborator.story_id = related.id) AS integer) AS related_collaborator_count,
    related.reporter_id AS related_reporter_id,
    related.key_result_id AS related_key_result_id,
    related.start_date AS related_start_date,
    related.end_date AS related_end_date,
    related.created_at AS related_created_at,
    related.updated_at AS related_updated_at,
    related.completed_at AS related_completed_at,
    related.deleted_at AS related_deleted_at,
    related.archived_at AS related_archived_at,
    CAST(ARRAY(
        SELECT story_label.label_id
        FROM story_labels AS story_label
        WHERE story_label.story_id = related.id
        ORDER BY story_label.label_id
    ) AS uuid[]) AS related_label_ids
FROM story_associations AS association
INNER JOIN stories AS source
    ON source.id = sqlc.arg(story_id)
   AND source.workspace_id = sqlc.arg(workspace_id)
INNER JOIN stories AS related
    ON related.id = CASE
        WHEN association.from_story_id = source.id THEN association.to_story_id
        ELSE association.from_story_id
    END
   AND related.workspace_id = source.workspace_id
INNER JOIN teams AS source_team
    ON source_team.team_id = source.team_id
   AND source_team.workspace_id = source.workspace_id
INNER JOIN teams AS related_team
    ON related_team.team_id = related.team_id
   AND related_team.workspace_id = related.workspace_id
INNER JOIN users AS actor
    ON actor.user_id = sqlc.arg(actor_id)
   AND actor.is_active = TRUE
INNER JOIN workspace_members AS workspace_member
    ON workspace_member.workspace_id = source.workspace_id
   AND workspace_member.user_id = actor.user_id
INNER JOIN team_members AS source_team_member
    ON source_team_member.team_id = source.team_id
   AND source_team_member.user_id = actor.user_id
INNER JOIN team_members AS related_team_member
    ON related_team_member.team_id = related.team_id
   AND related_team_member.user_id = actor.user_id
LEFT JOIN team_estimation_settings AS estimation
    ON estimation.team_id = related.team_id
   AND estimation.workspace_id = related.workspace_id
WHERE association.workspace_id = source.workspace_id
  AND (association.from_story_id = source.id OR association.to_story_id = source.id)
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR (
          source.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
          AND related.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
      )
  )
ORDER BY association.created_at, association.id;
