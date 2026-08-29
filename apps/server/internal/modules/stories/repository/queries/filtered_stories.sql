-- Static, actor-scoped filter surface used by ordinary lists, grouped lists,
-- and group pagination. Boolean flags distinguish an omitted scalar filter
-- from a zero value; arrays are empty when their filter is omitted.

-- name: ListVisibleFilteredStoryRows :many
WITH read_args AS (
    SELECT
        CAST(sqlc.arg(actor_id) AS uuid) AS actor_id,
        CAST(sqlc.arg(workspace_id) AS uuid) AS workspace_id,
        CAST(sqlc.arg(unrestricted_team_access) AS boolean) AS unrestricted_team_access,
        CAST(sqlc.arg(allowed_team_ids) AS uuid[]) AS allowed_team_ids,
        CAST(sqlc.arg(status_ids) AS uuid[]) AS status_ids,
        CAST(sqlc.arg(excluded_status_ids) AS uuid[]) AS excluded_status_ids,
        CAST(sqlc.arg(assignee_ids) AS uuid[]) AS assignee_ids,
        CAST(sqlc.arg(excluded_assignee_ids) AS uuid[]) AS excluded_assignee_ids,
        CAST(sqlc.arg(collaborator_ids) AS uuid[]) AS collaborator_ids,
        CAST(sqlc.arg(reporter_ids) AS uuid[]) AS reporter_ids,
        CAST(sqlc.arg(excluded_reporter_ids) AS uuid[]) AS excluded_reporter_ids,
        CAST(sqlc.arg(has_title_contains) AS boolean) AS has_title_contains,
        CAST(sqlc.arg(title_contains) AS text) AS title_contains,
        CAST(sqlc.arg(has_title_not_contains) AS boolean) AS has_title_not_contains,
        CAST(sqlc.arg(title_not_contains) AS text) AS title_not_contains,
        CAST(sqlc.arg(priorities) AS text[]) AS priorities,
        CAST(sqlc.arg(excluded_priorities) AS text[]) AS excluded_priorities,
        CAST(sqlc.arg(categories) AS text[]) AS categories,
        CAST(sqlc.arg(team_ids) AS uuid[]) AS team_ids,
        CAST(sqlc.arg(excluded_team_ids) AS uuid[]) AS excluded_team_ids,
        CAST(sqlc.arg(sprint_ids) AS uuid[]) AS sprint_ids,
        CAST(sqlc.arg(excluded_sprint_ids) AS uuid[]) AS excluded_sprint_ids,
        CAST(sqlc.arg(label_ids) AS uuid[]) AS label_ids,
        CAST(sqlc.arg(excluded_label_ids) AS uuid[]) AS excluded_label_ids,
        CAST(sqlc.arg(estimate_values) AS smallint[]) AS estimate_values,
        CAST(sqlc.arg(excluded_estimate_values) AS smallint[]) AS excluded_estimate_values,
        CAST(sqlc.arg(has_parent) AS boolean) AS has_parent,
        CAST(sqlc.arg(parent_id) AS uuid) AS parent_id,
        CAST(sqlc.arg(has_objective) AS boolean) AS has_objective,
        CAST(sqlc.arg(objective_id) AS uuid) AS objective_id,
        CAST(sqlc.arg(has_excluded_objective) AS boolean) AS has_excluded_objective,
        CAST(sqlc.arg(excluded_objective_id) AS uuid) AS excluded_objective_id,
        CAST(sqlc.arg(has_key_result) AS boolean) AS has_key_result,
        CAST(sqlc.arg(key_result_id) AS uuid) AS key_result_id,
        CAST(sqlc.arg(has_no_assignee) AS boolean) AS has_no_assignee,
        CAST(sqlc.arg(has_assignee) AS boolean) AS has_assignee,
        CAST(sqlc.arg(has_blocked_by) AS boolean) AS has_blocked_by,
        CAST(sqlc.arg(assigned_to_me) AS boolean) AS assigned_to_me,
        CAST(sqlc.arg(collaborating_with_me) AS boolean) AS collaborating_with_me,
        CAST(sqlc.arg(created_by_me) AS boolean) AS created_by_me,
        CAST(sqlc.arg(show_sub_stories) AS boolean) AS show_sub_stories,
        CAST(sqlc.arg(has_created_after) AS boolean) AS has_created_after,
        CAST(sqlc.arg(created_after) AS timestamptz) AS created_after,
        CAST(sqlc.arg(has_created_before) AS boolean) AS has_created_before,
        CAST(sqlc.arg(created_before) AS timestamptz) AS created_before,
        CAST(sqlc.arg(has_updated_after) AS boolean) AS has_updated_after,
        CAST(sqlc.arg(updated_after) AS timestamptz) AS updated_after,
        CAST(sqlc.arg(has_updated_before) AS boolean) AS has_updated_before,
        CAST(sqlc.arg(updated_before) AS timestamptz) AS updated_before,
        CAST(sqlc.arg(has_start_date_after) AS boolean) AS has_start_date_after,
        CAST(sqlc.arg(start_date_after) AS date) AS start_date_after,
        CAST(sqlc.arg(has_start_date_before) AS boolean) AS has_start_date_before,
        CAST(sqlc.arg(start_date_before) AS date) AS start_date_before,
        CAST(sqlc.arg(has_start_date_not) AS boolean) AS has_start_date_not,
        CAST(sqlc.arg(start_date_not) AS date) AS start_date_not,
        CAST(sqlc.arg(has_deadline_after) AS boolean) AS has_deadline_after,
        CAST(sqlc.arg(deadline_after) AS date) AS deadline_after,
        CAST(sqlc.arg(has_deadline_before) AS boolean) AS has_deadline_before,
        CAST(sqlc.arg(deadline_before) AS date) AS deadline_before,
        CAST(sqlc.arg(has_deadline_not) AS boolean) AS has_deadline_not,
        CAST(sqlc.arg(deadline_not) AS date) AS deadline_not,
        CAST(sqlc.arg(has_completed_after) AS boolean) AS has_completed_after,
        CAST(sqlc.arg(completed_after) AS timestamptz) AS completed_after,
        CAST(sqlc.arg(has_completed_before) AS boolean) AS has_completed_before,
        CAST(sqlc.arg(completed_before) AS timestamptz) AS completed_before,
        CAST(sqlc.arg(is_completed) AS boolean) AS is_completed,
        CAST(sqlc.arg(is_not_completed) AS boolean) AS is_not_completed,
        CAST(sqlc.arg(include_archived) AS boolean) AS include_archived,
        CAST(sqlc.arg(include_deleted) AS boolean) AS include_deleted,
        CAST(sqlc.arg(group_by) AS text) AS group_by,
        CAST(sqlc.arg(order_by) AS text) AS order_by,
        CAST(sqlc.arg(order_direction) AS text) AS order_direction,
        CAST(sqlc.arg(apply_group_filter) AS boolean) AS apply_group_filter,
        CAST(sqlc.arg(group_key) AS text) AS group_key,
        CAST(sqlc.arg(read_mode) AS text) AS read_mode,
        CAST(sqlc.arg(result_limit) AS integer) AS result_limit,
        CAST(sqlc.arg(result_offset) AS integer) AS result_offset
),
candidate_stories AS (
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
        CAST((
            SELECT COUNT(*)
            FROM story_collaborators AS collaborator
            WHERE collaborator.story_id = story.id
        ) AS integer) AS collaborator_count,
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
        ) AS uuid[]) AS label_ids,
        CAST(CASE args.group_by
            WHEN 'status' THEN COALESCE(CAST(story.status_id AS text), 'null')
            WHEN 'assignee' THEN COALESCE(CAST(story.assignee_id AS text), 'null')
            WHEN 'priority' THEN COALESCE(story.priority, 'No Priority')
            WHEN 'team' THEN CAST(story.team_id AS text)
            WHEN 'sprint' THEN COALESCE(CAST(story.sprint_id AS text), 'null')
            ELSE 'none'
        END AS text) AS group_key,
        args.order_by,
        args.order_direction,
        args.apply_group_filter,
        args.group_key AS requested_group_key,
        args.read_mode,
        args.result_limit,
        args.result_offset
    FROM stories AS story
    CROSS JOIN read_args AS args
    INNER JOIN teams AS team
        ON team.team_id = story.team_id
       AND team.workspace_id = story.workspace_id
    INNER JOIN users AS actor
        ON actor.user_id = args.actor_id
       AND actor.is_active = TRUE
    INNER JOIN workspace_members AS workspace_member
        ON workspace_member.workspace_id = story.workspace_id
       AND workspace_member.user_id = actor.user_id
    INNER JOIN team_members AS team_member
        ON team_member.team_id = story.team_id
       AND team_member.user_id = actor.user_id
    LEFT JOIN statuses AS status
        ON status.status_id = story.status_id
       AND status.team_id = story.team_id
       AND status.workspace_id = story.workspace_id
    LEFT JOIN objectives AS objective
        ON objective.objective_id = story.objective_id
       AND objective.workspace_id = story.workspace_id
    LEFT JOIN sprints AS sprint
        ON sprint.sprint_id = story.sprint_id
       AND sprint.workspace_id = story.workspace_id
    LEFT JOIN team_estimation_settings AS estimation
        ON estimation.team_id = story.team_id
       AND estimation.workspace_id = story.workspace_id
    WHERE story.workspace_id = args.workspace_id
      AND (args.unrestricted_team_access OR story.team_id = ANY(args.allowed_team_ids))
      AND (cardinality(args.status_ids) = 0 OR story.status_id = ANY(args.status_ids))
      AND (cardinality(args.excluded_status_ids) = 0 OR story.status_id IS NULL OR story.status_id <> ALL(args.excluded_status_ids))
      AND (cardinality(args.assignee_ids) = 0 OR story.assignee_id = ANY(args.assignee_ids))
      AND (cardinality(args.excluded_assignee_ids) = 0 OR story.assignee_id IS NULL OR story.assignee_id <> ALL(args.excluded_assignee_ids))
      AND (cardinality(args.collaborator_ids) = 0 OR EXISTS (
          SELECT 1
          FROM story_collaborators AS collaborator_filter
          WHERE collaborator_filter.story_id = story.id
            AND collaborator_filter.user_id = ANY(args.collaborator_ids)
      ))
      AND (cardinality(args.reporter_ids) = 0 OR story.reporter_id = ANY(args.reporter_ids))
      AND (cardinality(args.excluded_reporter_ids) = 0 OR story.reporter_id IS NULL OR story.reporter_id <> ALL(args.excluded_reporter_ids))
      AND (NOT args.has_title_contains OR (
          story.title ILIKE '%' || args.title_contains || '%'
          OR COALESCE(story.description, '') ILIKE '%' || args.title_contains || '%'
          OR COALESCE(story.description_html, '') ILIKE '%' || args.title_contains || '%'
      ))
      AND (NOT args.has_title_not_contains OR NOT (
          story.title ILIKE '%' || args.title_not_contains || '%'
          OR COALESCE(story.description, '') ILIKE '%' || args.title_not_contains || '%'
          OR COALESCE(story.description_html, '') ILIKE '%' || args.title_not_contains || '%'
      ))
      AND (cardinality(args.priorities) = 0 OR COALESCE(story.priority, 'No Priority') = ANY(args.priorities))
      AND (cardinality(args.excluded_priorities) = 0 OR COALESCE(story.priority, 'No Priority') <> ALL(args.excluded_priorities))
      AND (cardinality(args.categories) = 0 OR CAST(status.category AS text) = ANY(args.categories))
      AND (cardinality(args.team_ids) = 0 OR story.team_id = ANY(args.team_ids))
      AND (cardinality(args.excluded_team_ids) = 0 OR story.team_id <> ALL(args.excluded_team_ids))
      AND (cardinality(args.sprint_ids) = 0 OR story.sprint_id = ANY(args.sprint_ids))
      AND (cardinality(args.excluded_sprint_ids) = 0 OR story.sprint_id IS NULL OR story.sprint_id <> ALL(args.excluded_sprint_ids))
      AND (cardinality(args.label_ids) = 0 OR EXISTS (
          SELECT 1
          FROM story_labels AS included_label
          WHERE included_label.story_id = story.id
            AND included_label.label_id = ANY(args.label_ids)
      ))
      AND (cardinality(args.excluded_label_ids) = 0 OR NOT EXISTS (
          SELECT 1
          FROM story_labels AS excluded_label
          WHERE excluded_label.story_id = story.id
            AND excluded_label.label_id = ANY(args.excluded_label_ids)
      ))
      AND (cardinality(args.estimate_values) = 0 OR story.estimate_unit = ANY(args.estimate_values))
      AND (cardinality(args.excluded_estimate_values) = 0 OR story.estimate_unit IS NULL OR story.estimate_unit <> ALL(args.excluded_estimate_values))
      AND (NOT args.has_parent OR story.parent_id = args.parent_id)
      AND (args.has_parent OR args.show_sub_stories OR story.parent_id IS NULL)
      AND (NOT args.has_objective OR story.objective_id = args.objective_id)
      AND (NOT args.has_excluded_objective OR story.objective_id IS DISTINCT FROM args.excluded_objective_id)
      AND (NOT args.has_key_result OR story.key_result_id = args.key_result_id)
      AND (NOT args.has_no_assignee OR story.assignee_id IS NULL)
      AND (NOT args.has_assignee OR story.assignee_id IS NOT NULL)
      AND (NOT args.has_blocked_by OR story.blocked_by_id IS NOT NULL)
      AND (
          (NOT args.assigned_to_me AND NOT args.collaborating_with_me AND NOT args.created_by_me)
          OR (args.assigned_to_me AND story.assignee_id = args.actor_id)
          OR (args.created_by_me AND story.reporter_id = args.actor_id)
          OR (args.collaborating_with_me AND EXISTS (
              SELECT 1
              FROM story_collaborators AS actor_collaboration
              WHERE actor_collaboration.story_id = story.id
                AND actor_collaboration.user_id = args.actor_id
          ))
      )
      AND (NOT args.has_created_after OR story.created_at >= args.created_after)
      AND (NOT args.has_created_before OR story.created_at <= args.created_before)
      AND (NOT args.has_updated_after OR story.updated_at >= args.updated_after)
      AND (NOT args.has_updated_before OR story.updated_at <= args.updated_before)
      AND (NOT args.has_start_date_after OR story.start_date >= args.start_date_after)
      AND (NOT args.has_start_date_before OR story.start_date <= args.start_date_before)
      AND (NOT args.has_start_date_not OR story.start_date IS DISTINCT FROM args.start_date_not)
      AND (NOT args.has_deadline_after OR story.end_date >= args.deadline_after)
      AND (NOT args.has_deadline_before OR story.end_date <= args.deadline_before)
      AND (NOT args.has_deadline_not OR story.end_date IS DISTINCT FROM args.deadline_not)
      AND (NOT args.has_completed_after OR story.completed_at >= args.completed_after)
      AND (NOT args.has_completed_before OR story.completed_at <= args.completed_before)
      AND (NOT args.is_completed OR story.completed_at IS NOT NULL)
      AND (NOT args.is_not_completed OR story.completed_at IS NULL)
      AND ((args.include_archived AND story.archived_at IS NOT NULL) OR (NOT args.include_archived AND story.archived_at IS NULL))
      AND ((args.include_deleted AND story.deleted_at IS NOT NULL) OR (NOT args.include_deleted AND story.deleted_at IS NULL))
),
filtered_stories AS (
    SELECT *
    FROM candidate_stories
    WHERE NOT apply_group_filter OR group_key = requested_group_key
),
ranked_stories AS (
    SELECT
        filtered_stories.*,
        CAST(COUNT(*) OVER (PARTITION BY group_key) AS integer) AS total_count,
        ROW_NUMBER() OVER (
            PARTITION BY group_key
            ORDER BY
                CASE WHEN order_by = 'created' AND order_direction = 'asc' THEN created_at END ASC,
                CASE WHEN order_by = 'created' AND order_direction = 'desc' THEN created_at END DESC,
                CASE WHEN order_by = 'updated' AND order_direction = 'asc' THEN updated_at END ASC,
                CASE WHEN order_by = 'updated' AND order_direction = 'desc' THEN updated_at END DESC,
                CASE WHEN order_by = 'priority' AND order_direction = 'asc' THEN CASE priority
                    WHEN 'Urgent' THEN 1 WHEN 'High' THEN 2 WHEN 'Medium' THEN 3 WHEN 'Low' THEN 4 WHEN 'No Priority' THEN 5 ELSE 6 END
                END ASC,
                CASE WHEN order_by = 'priority' AND order_direction = 'desc' THEN CASE priority
                    WHEN 'Urgent' THEN 1 WHEN 'High' THEN 2 WHEN 'Medium' THEN 3 WHEN 'Low' THEN 4 WHEN 'No Priority' THEN 5 ELSE 6 END
                END DESC,
                CASE WHEN order_by = 'deadline' AND order_direction = 'asc' THEN end_date END ASC NULLS LAST,
                CASE WHEN order_by = 'deadline' AND order_direction = 'desc' THEN end_date END DESC NULLS LAST,
                CASE WHEN order_by = 'completed' AND order_direction = 'asc' THEN completed_at END ASC NULLS LAST,
                CASE WHEN order_by = 'completed' AND order_direction = 'desc' THEN completed_at END DESC NULLS LAST,
                created_at DESC,
                id ASC
        ) AS row_number
    FROM filtered_stories
)
SELECT
    id, sequence_id, title, priority, estimate_unit, estimate_scheme,
    estimated_duration_minutes, minimum_focus_block_minutes,
    auto_scheduling_enabled, auto_scheduling_locked, auto_scheduling_status,
    auto_scheduling_reason, auto_scheduling_updated_at, parent_id,
    objective_id, objective_name, objective_description, sprint_id,
    sprint_name, sprint_goal, sprint_start_date, sprint_end_date, team_id,
    team_code, team_name, workspace_id, status_id, assignee_id,
    collaborator_count, reporter_id, key_result_id, start_date, end_date,
    created_at, updated_at, completed_at, deleted_at, archived_at, label_ids,
    group_key, total_count, row_number
FROM ranked_stories
WHERE read_mode = 'page' OR row_number <= result_limit
ORDER BY
    CASE WHEN read_mode = 'grouped' THEN group_key END ASC,
    CASE WHEN read_mode = 'grouped' THEN row_number END ASC,
    CASE WHEN order_by = 'created' AND order_direction = 'asc' THEN created_at END ASC,
    CASE WHEN order_by = 'created' AND order_direction = 'desc' THEN created_at END DESC,
    CASE WHEN order_by = 'updated' AND order_direction = 'asc' THEN updated_at END ASC,
    CASE WHEN order_by = 'updated' AND order_direction = 'desc' THEN updated_at END DESC,
    CASE WHEN order_by = 'priority' AND order_direction = 'asc' THEN CASE priority
        WHEN 'Urgent' THEN 1 WHEN 'High' THEN 2 WHEN 'Medium' THEN 3 WHEN 'Low' THEN 4 WHEN 'No Priority' THEN 5 ELSE 6 END
    END ASC,
    CASE WHEN order_by = 'priority' AND order_direction = 'desc' THEN CASE priority
        WHEN 'Urgent' THEN 1 WHEN 'High' THEN 2 WHEN 'Medium' THEN 3 WHEN 'Low' THEN 4 WHEN 'No Priority' THEN 5 ELSE 6 END
    END DESC,
    CASE WHEN order_by = 'deadline' AND order_direction = 'asc' THEN end_date END ASC NULLS LAST,
    CASE WHEN order_by = 'deadline' AND order_direction = 'desc' THEN end_date END DESC NULLS LAST,
    CASE WHEN order_by = 'completed' AND order_direction = 'asc' THEN completed_at END ASC NULLS LAST,
    CASE WHEN order_by = 'completed' AND order_direction = 'desc' THEN completed_at END DESC NULLS LAST,
    created_at DESC,
    id ASC
LIMIT CASE WHEN CAST(sqlc.arg(read_mode) AS text) = 'page' THEN CAST(sqlc.arg(result_limit) AS integer) ELSE 2147483647 END
OFFSET CASE WHEN CAST(sqlc.arg(read_mode) AS text) = 'page' THEN CAST(sqlc.arg(result_offset) AS integer) ELSE 0 END;

-- name: CountVisibleStories :one
SELECT CAST(COUNT(*) AS integer)
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
WHERE story.workspace_id = sqlc.arg(workspace_id)
  AND story.deleted_at IS NULL
  AND story.archived_at IS NULL
  AND (
      CAST(sqlc.arg(unrestricted_team_access) AS boolean)
      OR story.team_id = ANY(CAST(sqlc.arg(allowed_team_ids) AS uuid[]))
  );

-- name: ListVisibleStoryGroupCatalog :many
WITH catalog_args AS (
    SELECT
        CAST(sqlc.arg(actor_id) AS uuid) AS actor_id,
        CAST(sqlc.arg(workspace_id) AS uuid) AS workspace_id,
        CAST(sqlc.arg(unrestricted_team_access) AS boolean) AS unrestricted_team_access,
        CAST(sqlc.arg(allowed_team_ids) AS uuid[]) AS allowed_team_ids,
        CAST(sqlc.arg(team_ids) AS uuid[]) AS team_ids,
        CAST(sqlc.arg(excluded_team_ids) AS uuid[]) AS excluded_team_ids,
        CAST(sqlc.arg(assignee_ids) AS uuid[]) AS assignee_ids,
        CAST(sqlc.arg(group_by) AS text) AS group_by
),
visible_teams AS (
    SELECT team.team_id, team.name
    FROM teams AS team
    CROSS JOIN catalog_args AS args
    INNER JOIN users AS actor
        ON actor.user_id = args.actor_id
       AND actor.is_active = TRUE
    INNER JOIN workspace_members AS workspace_member
        ON workspace_member.workspace_id = team.workspace_id
       AND workspace_member.user_id = actor.user_id
    INNER JOIN team_members AS team_member
        ON team_member.team_id = team.team_id
       AND team_member.user_id = actor.user_id
    WHERE team.workspace_id = args.workspace_id
      AND (args.unrestricted_team_access OR team.team_id = ANY(args.allowed_team_ids))
      AND (cardinality(args.team_ids) = 0 OR team.team_id = ANY(args.team_ids))
      AND (cardinality(args.excluded_team_ids) = 0 OR team.team_id <> ALL(args.excluded_team_ids))
),
catalog AS (
    SELECT CAST(status.status_id AS text) AS group_key, CAST(COALESCE(status.order_index, 2147483647) AS bigint) AS sort_rank, CAST(status.name AS text) AS sort_name
    FROM statuses AS status
    INNER JOIN visible_teams AS team ON team.team_id = status.team_id
    CROSS JOIN catalog_args AS args
    WHERE args.group_by = 'status' AND status.workspace_id = args.workspace_id

    UNION ALL

    SELECT CAST(team.team_id AS text), CAST(0 AS bigint), CAST(team.name AS text)
    FROM visible_teams AS team
    CROSS JOIN catalog_args AS args
    WHERE args.group_by = 'team'

    UNION ALL

    SELECT priority.group_key, priority.sort_rank, priority.group_key
    FROM (VALUES
        (CAST('Urgent' AS text), CAST(1 AS bigint)), (CAST('High' AS text), CAST(2 AS bigint)),
        (CAST('Medium' AS text), CAST(3 AS bigint)), (CAST('Low' AS text), CAST(4 AS bigint)),
        (CAST('No Priority' AS text), CAST(5 AS bigint))
    ) AS priority(group_key, sort_rank)
    CROSS JOIN catalog_args AS args
    WHERE args.group_by = 'priority'

    UNION ALL

    SELECT DISTINCT CAST(account.user_id AS text), CAST(0 AS bigint), CAST(account.username AS text)
    FROM users AS account
    INNER JOIN team_members AS membership ON membership.user_id = account.user_id
    INNER JOIN visible_teams AS team ON team.team_id = membership.team_id
    INNER JOIN workspace_members AS account_workspace_member
        ON account_workspace_member.user_id = account.user_id
    CROSS JOIN catalog_args AS args
    WHERE args.group_by = 'assignee'
      AND account_workspace_member.workspace_id = args.workspace_id
      AND account.is_active = TRUE
      AND account.is_system = FALSE
      AND (cardinality(args.assignee_ids) = 0 OR account.user_id = ANY(args.assignee_ids))

    UNION ALL

    SELECT CAST('null' AS text), CAST(1 AS bigint), CAST('Unassigned' AS text)
    FROM catalog_args AS args
    WHERE args.group_by = 'assignee' AND cardinality(args.assignee_ids) = 0

    UNION ALL

    SELECT CAST(sprint.sprint_id AS text), CAST(0 AS bigint), CAST(sprint.name AS text)
    FROM sprints AS sprint
    INNER JOIN visible_teams AS team ON team.team_id = sprint.team_id
    CROSS JOIN catalog_args AS args
    WHERE args.group_by = 'sprint' AND sprint.workspace_id = args.workspace_id

    UNION ALL

    SELECT CAST('null' AS text), CAST(1 AS bigint), CAST('No Sprint' AS text)
    FROM catalog_args AS args
    WHERE args.group_by = 'sprint'

    UNION ALL

    SELECT CAST('none' AS text), CAST(0 AS bigint), CAST('none' AS text)
    FROM catalog_args AS args
    WHERE args.group_by = 'none'
)
SELECT group_key
FROM catalog
GROUP BY group_key, sort_rank, sort_name
ORDER BY sort_rank, sort_name, group_key
-- One look-ahead row lets the adapter reject an oversized catalog without
-- loading an unbounded number of group identifiers into memory.
LIMIT CAST(sqlc.arg(catalog_limit) AS integer);
