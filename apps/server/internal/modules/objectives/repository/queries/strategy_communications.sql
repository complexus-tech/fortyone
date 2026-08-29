-- name: ListStrategyCommunicationAdministrators :many
SELECT
    user_account.user_id,
    membership.workspace_id,
    CAST(COALESCE(NULLIF(TRIM(user_account.timezone), ''), 'UTC') AS text) AS timezone
FROM public.workspace_members AS membership
INNER JOIN public.users AS user_account
    ON user_account.user_id = membership.user_id
   AND user_account.is_active = true
   AND user_account.is_system = false
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.role = 'admin'
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR membership.workspace_id > sqlc.arg(after_workspace_id)
      OR (
          membership.workspace_id = sqlc.arg(after_workspace_id)
          AND user_account.user_id > sqlc.arg(after_user_id)
      )
  )
ORDER BY membership.workspace_id, user_account.user_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: GetStrategyCommunicationFoundation :one
SELECT
    EXISTS (
        SELECT 1
        FROM public.workspace_strategies AS strategy
        WHERE strategy.workspace_id = sqlc.arg(workspace_id)
          AND NULLIF(TRIM(strategy.ultimate_goal), '') IS NOT NULL
    ) AS has_ultimate_goal,
    (
        SELECT COUNT(*)
        FROM public.strategic_pillars AS pillar
        WHERE pillar.workspace_id = sqlc.arg(workspace_id)
    ) AS pillar_count,
    (
        SELECT COUNT(*)
        FROM public.objectives AS objective
        LEFT JOIN public.objective_statuses AS status
            ON status.status_id = objective.status_id
        WHERE objective.workspace_id = sqlc.arg(workspace_id)
          AND (
              objective.start_date IS NULL
              OR objective.start_date < CAST(sqlc.arg(period_end) AS date)
          )
          AND (
              objective.end_date IS NULL
              OR objective.end_date >= CAST(sqlc.arg(period_start) AS date)
          )
          AND COALESCE(status.category, '') NOT IN ('completed', 'cancelled', 'paused')
    ) AS objective_count;

-- name: GetStrategyCommunicationMonthlySummary :one
WITH objective_data AS (
    SELECT
        objective.objective_id,
        CAST(objective.health AS text) AS health,
        alignment.pillar_id
    FROM public.objectives AS objective
    LEFT JOIN public.objective_statuses AS status
        ON status.status_id = objective.status_id
    LEFT JOIN public.strategy_objective_alignments AS alignment
        ON alignment.objective_id = objective.objective_id
    WHERE objective.workspace_id = sqlc.arg(workspace_id)
      AND COALESCE(status.category, '') NOT IN ('completed', 'cancelled', 'paused')
),
key_result_data AS (
    SELECT
        CASE
            WHEN CAST(key_result.measurement_type AS text) = 'percentage' THEN
                GREATEST(0.0, LEAST(100.0, key_result.current_value))
            WHEN CAST(key_result.measurement_type AS text) = 'number'
                 AND key_result.target_value = key_result.start_value THEN
                CASE
                    WHEN key_result.current_value = key_result.target_value THEN 100.0
                    ELSE 0.0
                END
            WHEN CAST(key_result.measurement_type AS text) = 'number' THEN
                GREATEST(
                    0.0,
                    LEAST(
                        100.0,
                        (
                            (key_result.current_value - key_result.start_value)
                            / NULLIF(key_result.target_value - key_result.start_value, 0)
                        ) * 100.0
                    )
                )
            WHEN CAST(key_result.measurement_type AS text) = 'boolean' THEN
                CASE
                    WHEN key_result.current_value = key_result.target_value THEN 100.0
                    ELSE 0.0
                END
            ELSE NULL
        END AS progress
    FROM public.key_results AS key_result
    INNER JOIN objective_data AS objective
        ON objective.objective_id = key_result.objective_id
)
SELECT
    (
        SELECT COUNT(*)
        FROM public.strategic_pillars AS pillar
        WHERE pillar.workspace_id = sqlc.arg(workspace_id)
    ) AS pillar_count,
    (
        SELECT COUNT(DISTINCT pillar_id)
        FROM objective_data
        WHERE pillar_id IS NOT NULL
          AND health IN ('At Risk', 'Off Track')
    ) AS pillars_needing_review,
    (SELECT COUNT(*) FROM objective_data) AS objective_count,
    (
        SELECT COUNT(*)
        FROM objective_data
        WHERE health IN ('At Risk', 'Off Track')
    ) AS at_risk_objectives,
    (
        SELECT COUNT(*)
        FROM objective_data
        WHERE pillar_id IS NULL
    ) AS unaligned_objectives,
    (SELECT COUNT(*) FROM key_result_data) AS key_result_count,
    (SELECT COUNT(progress) FROM key_result_data) AS key_result_progress_count,
    CAST(
        COALESCE(
            (SELECT AVG(progress) FROM key_result_data),
            CAST(0 AS numeric)
        )
        AS double precision
    ) AS key_result_progress,
    (
        SELECT COUNT(*)
        FROM public.stories AS story
        WHERE story.workspace_id = sqlc.arg(workspace_id)
          AND (story.objective_id IS NOT NULL OR story.key_result_id IS NOT NULL)
          AND story.completed_at >= CAST(sqlc.arg(period_start) AS timestamptz)
          AND story.completed_at < CAST(sqlc.arg(period_end) AS timestamptz)
          AND story.deleted_at IS NULL
    ) AS completed_stories;

-- name: ListStrategyWeeklyCommunicationRecipients :many
SELECT
    user_account.user_id,
    membership.workspace_id,
    CAST(COALESCE(NULLIF(TRIM(user_account.timezone), ''), 'UTC') AS text) AS timezone
FROM public.workspace_members AS membership
INNER JOIN public.users AS user_account
    ON user_account.user_id = membership.user_id
   AND user_account.is_active = true
   AND user_account.is_system = false
INNER JOIN public.workspaces AS workspace
    ON workspace.workspace_id = membership.workspace_id
   AND workspace.deleted_at IS NULL
WHERE membership.role IN ('admin', 'member', 'guest')
  AND (
      NOT CAST(sqlc.arg(has_cursor) AS boolean)
      OR membership.workspace_id > sqlc.arg(after_workspace_id)
      OR (
          membership.workspace_id = sqlc.arg(after_workspace_id)
          AND user_account.user_id > sqlc.arg(after_user_id)
      )
  )
ORDER BY membership.workspace_id, user_account.user_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: ListStrategyWeeklyCommunicationRecords :many
WITH recipient AS MATERIALIZED (
    SELECT
        user_account.user_id,
        membership.workspace_id,
        membership.role,
        CAST(COALESCE(NULLIF(TRIM(user_account.timezone), ''), 'UTC') AS text) AS timezone
    FROM public.workspace_members AS membership
    INNER JOIN public.users AS user_account
        ON user_account.user_id = membership.user_id
       AND user_account.is_active = true
       AND user_account.is_system = false
    INNER JOIN public.workspaces AS workspace
        ON workspace.workspace_id = membership.workspace_id
       AND workspace.deleted_at IS NULL
    WHERE membership.user_id = sqlc.arg(user_id)
      AND membership.workspace_id = sqlc.arg(workspace_id)
),
eligible_objectives AS (
    SELECT
        recipient.user_id,
        recipient.workspace_id,
        recipient.timezone,
        objective.objective_id,
        objective.team_id,
        objective.name AS objective_name,
        objective.health IS NOT NULL AS objective_health_is_set,
        COALESCE(CAST(objective.health AS text), '') AS objective_health,
        status.status_id AS objective_status_id,
        status.name AS objective_status_name,
        status.category AS objective_status_category,
        objective.start_date AS objective_start_date,
        objective.end_date AS objective_end_date,
        objective.updated_at AS objective_updated_at,
        objective.updated_at < CAST(sqlc.arg(stale_before) AS timestamptz) AS is_stale_objective,
        COALESCE(CAST(objective.health AS text) IN ('At Risk', 'Off Track'), false) AS is_at_risk_objective
    FROM recipient
    INNER JOIN public.objectives AS objective
        ON objective.workspace_id = recipient.workspace_id
       AND objective.lead_user_id = recipient.user_id
    LEFT JOIN public.objective_statuses AS status
        ON status.status_id = objective.status_id
    WHERE objective.team_id IS NOT NULL
      AND COALESCE(status.category, '') NOT IN ('completed', 'cancelled', 'paused')
      AND (
          recipient.role = 'admin'
          OR EXISTS (
              SELECT 1
              FROM public.team_members AS team_membership
              WHERE team_membership.team_id = objective.team_id
                AND team_membership.user_id = recipient.user_id
          )
      )
),
stale_key_results AS (
    SELECT
        key_result.id,
        key_result.objective_id,
        key_result.name,
        CAST(key_result.measurement_type AS text) AS measurement_type,
        CAST(key_result.start_value AS double precision) AS start_value,
        CAST(key_result.current_value AS double precision) AS current_value,
        CAST(key_result.target_value AS double precision) AS target_value,
        key_result.start_date,
        key_result.end_date,
        key_result.updated_at
    FROM public.key_results AS key_result
    INNER JOIN eligible_objectives AS objective
        ON objective.objective_id = key_result.objective_id
    WHERE key_result.updated_at < CAST(sqlc.arg(stale_before) AS timestamptz)
      AND NOT COALESCE(
          CASE
              WHEN CAST(key_result.measurement_type AS text) IN ('percentage', 'number')
                   AND key_result.target_value >= key_result.start_value THEN
                  key_result.current_value >= key_result.target_value
              WHEN CAST(key_result.measurement_type AS text) IN ('percentage', 'number')
                   AND key_result.target_value < key_result.start_value THEN
                  key_result.current_value <= key_result.target_value
              WHEN CAST(key_result.measurement_type AS text) = 'boolean' THEN
                  key_result.current_value = key_result.target_value
              ELSE false
          END,
          false
      )
),
signal_rows AS (
    SELECT
        objective.user_id,
        objective.workspace_id,
        objective.timezone,
        objective.objective_id,
        objective.team_id,
        objective.objective_name,
        objective.objective_health_is_set,
        objective.objective_health,
        objective.objective_status_id,
        objective.objective_status_name,
        objective.objective_status_category,
        objective.objective_start_date,
        objective.objective_end_date,
        objective.objective_updated_at,
        objective.is_stale_objective,
        objective.is_at_risk_objective,
        key_result.id AS key_result_id,
        key_result.name AS key_result_name,
        key_result.measurement_type AS key_result_measurement_type,
        key_result.start_value AS key_result_start_value,
        key_result.current_value AS key_result_current_value,
        key_result.target_value AS key_result_target_value,
        key_result.start_date AS key_result_start_date,
        key_result.end_date AS key_result_end_date,
        key_result.updated_at AS key_result_updated_at,
        CASE WHEN key_result.id IS NULL THEN 1 ELSE 0 END AS key_result_null_rank,
        COALESCE(
            key_result.id,
            CAST('00000000-0000-0000-0000-000000000000' AS uuid)
        ) AS key_result_sort_id
    FROM eligible_objectives AS objective
    LEFT JOIN stale_key_results AS key_result
        ON key_result.objective_id = objective.objective_id
    WHERE (
        objective.is_stale_objective
        OR objective.is_at_risk_objective
        OR key_result.id IS NOT NULL
    )
      AND (
          NOT CAST(sqlc.arg(has_cursor) AS boolean)
          OR objective.objective_id > sqlc.arg(after_objective_id)
          OR (
              objective.objective_id = sqlc.arg(after_objective_id)
              AND CASE WHEN key_result.id IS NULL THEN 1 ELSE 0 END
                  > CAST(sqlc.arg(after_key_result_null_rank) AS integer)
          )
          OR (
              objective.objective_id = sqlc.arg(after_objective_id)
              AND CASE WHEN key_result.id IS NULL THEN 1 ELSE 0 END
                  = CAST(sqlc.arg(after_key_result_null_rank) AS integer)
              AND COALESCE(
                  key_result.id,
                  CAST('00000000-0000-0000-0000-000000000000' AS uuid)
              ) > sqlc.arg(after_key_result_id)
          )
      )
)
SELECT
    communication_signal.user_id,
    communication_signal.workspace_id,
    communication_signal.timezone,
    communication_signal.objective_id,
    communication_signal.team_id,
    communication_signal.objective_name,
    CAST(communication_signal.objective_health_is_set AS boolean) AS objective_health_is_set,
    CAST(communication_signal.objective_health AS text) AS objective_health,
    communication_signal.objective_status_id,
    communication_signal.objective_status_name,
    communication_signal.objective_status_category,
    communication_signal.objective_start_date,
    communication_signal.objective_end_date,
    communication_signal.objective_updated_at,
    communication_signal.is_stale_objective,
    CAST(communication_signal.is_at_risk_objective AS boolean) AS is_at_risk_objective,
    communication_signal.key_result_id,
    communication_signal.key_result_name,
    communication_signal.key_result_measurement_type,
    communication_signal.key_result_start_value,
    communication_signal.key_result_current_value,
    communication_signal.key_result_target_value,
    communication_signal.key_result_start_date,
    communication_signal.key_result_end_date,
    communication_signal.key_result_updated_at,
    communication_signal.key_result_null_rank,
    communication_signal.key_result_sort_id
FROM signal_rows AS communication_signal
ORDER BY
    communication_signal.objective_id,
    communication_signal.key_result_null_rank,
    communication_signal.key_result_sort_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);
