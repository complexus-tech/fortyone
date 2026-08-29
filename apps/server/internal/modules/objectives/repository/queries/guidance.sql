-- name: ListOverdueObjectiveGuidanceRecipients :many
SELECT DISTINCT
    objective.lead_user_id,
    lead.email AS lead_email,
    COALESCE(NULLIF(lead.full_name, ''), lead.username) AS lead_name,
    workspace.workspace_id,
    workspace.name AS workspace_name,
    workspace.slug AS workspace_slug,
    CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) AS email_enabled
FROM objectives AS objective
JOIN users AS lead
    ON lead.user_id = objective.lead_user_id
JOIN workspaces AS workspace
    ON workspace.workspace_id = objective.workspace_id
JOIN workspace_members AS membership
    ON membership.workspace_id = objective.workspace_id
    AND membership.user_id = objective.lead_user_id
    AND membership.role IN ('admin', 'member', 'guest')
JOIN objective_statuses AS status
    ON status.status_id = objective.status_id
JOIN workspace_settings AS settings
    ON settings.workspace_id = objective.workspace_id
LEFT JOIN notification_preferences AS preference
    ON preference.user_id = objective.lead_user_id
    AND preference.workspace_id = objective.workspace_id
WHERE (
        objective.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '7 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days'
        OR (
            settings.key_result_enabled = true
            AND EXISTS (
                SELECT 1
                FROM key_results AS key_result
                WHERE key_result.objective_id = objective.objective_id
                    AND key_result.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '7 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days'
                    AND NOT (
                        (
                            key_result.measurement_type IN ('percentage', 'number')
                            AND (
                                (
                                    key_result.target_value >= key_result.start_value
                                    AND key_result.current_value >= key_result.target_value
                                )
                                OR (
                                    key_result.target_value < key_result.start_value
                                    AND key_result.current_value <= key_result.target_value
                                )
                            )
                        )
                        OR (
                            key_result.measurement_type = 'boolean'
                            AND key_result.current_value = key_result.target_value
                        )
                    )
            )
        )
    )
    AND objective.lead_user_id IS NOT NULL
    AND lead.is_active = true
    AND lead.is_system = false
    AND NULLIF(TRIM(lead.email), '') IS NOT NULL
    AND workspace.deleted_at IS NULL
    AND (
        membership.role = 'admin'
        OR EXISTS (
            SELECT 1
            FROM team_members AS team_membership
            WHERE team_membership.team_id = objective.team_id
                AND team_membership.user_id = objective.lead_user_id
        )
    )
    AND status.category NOT IN ('completed', 'cancelled', 'paused')
    AND settings.objective_enabled = true
    AND CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) = true
    AND (
        NOT CAST(sqlc.arg(has_cursor) AS boolean)
        OR objective.lead_user_id > sqlc.arg(after_lead_user_id)
        OR (
            objective.lead_user_id = sqlc.arg(after_lead_user_id)
            AND objective.workspace_id > sqlc.arg(after_workspace_id)
        )
    )
ORDER BY objective.lead_user_id, workspace.workspace_id
LIMIT CAST(sqlc.arg(result_limit) AS integer);

-- name: ListOverdueObjectiveGuidanceItems :many
WITH objective_deadlines AS (
    SELECT
        objective.objective_id,
        objective.name,
        objective.end_date,
        objective.lead_user_id,
        objective.workspace_id,
        objective.team_id,
        lead.email AS lead_email,
        COALESCE(NULLIF(lead.full_name, ''), lead.username) AS lead_name,
        workspace.name AS workspace_name,
        workspace.slug AS workspace_slug,
        CASE
            WHEN objective.end_date = CAST(sqlc.arg(as_of) AS date) THEN 'due_today'
            WHEN objective.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '1 day' THEN 'due_tomorrow'
            WHEN objective.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days' THEN 'due_in_7_days'
            WHEN objective.end_date < CAST(sqlc.arg(as_of) AS date) THEN 'overdue'
            ELSE 'future'
        END AS deadline_status,
        CASE
            WHEN objective.end_date < CAST(sqlc.arg(as_of) AS date) THEN CAST(CAST(sqlc.arg(as_of) AS date) - objective.end_date AS integer)
            ELSE CAST(objective.end_date - CAST(sqlc.arg(as_of) AS date) AS integer)
        END AS days_difference,
        CASE
            WHEN settings.key_result_enabled = true THEN COALESCE(
                (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'id', key_result.id,
                            'name', key_result.name,
                            'end_date', key_result.end_date,
                            'measurement_type', key_result.measurement_type,
                            'start_value', key_result.start_value,
                            'current_value', key_result.current_value,
                            'target_value', key_result.target_value,
                            'is_completed', CASE
                                WHEN key_result.measurement_type IN ('percentage', 'number')
                                    AND key_result.target_value >= key_result.start_value
                                    AND key_result.current_value >= key_result.target_value THEN true
                                WHEN key_result.measurement_type IN ('percentage', 'number')
                                    AND key_result.target_value < key_result.start_value
                                    AND key_result.current_value <= key_result.target_value THEN true
                                WHEN key_result.measurement_type = 'boolean'
                                    AND key_result.current_value = key_result.target_value THEN true
                                ELSE false
                            END,
                            'deadline_status', CASE
                                WHEN key_result.end_date = CAST(sqlc.arg(as_of) AS date) THEN 'due_today'
                                WHEN key_result.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '1 day' THEN 'due_tomorrow'
                                WHEN key_result.end_date = CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days' THEN 'due_in_7_days'
                                WHEN key_result.end_date < CAST(sqlc.arg(as_of) AS date) THEN 'overdue'
                                ELSE 'future'
                            END,
                            'days_difference', CASE
                                WHEN key_result.end_date < CAST(sqlc.arg(as_of) AS date) THEN CAST(CAST(sqlc.arg(as_of) AS date) - key_result.end_date AS integer)
                                ELSE CAST(key_result.end_date - CAST(sqlc.arg(as_of) AS date) AS integer)
                            END
                        )
                    )
                    FROM key_results AS key_result
                    WHERE key_result.objective_id = objective.objective_id
                        AND key_result.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '7 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days'
                        AND NOT (
                            (
                                key_result.measurement_type IN ('percentage', 'number')
                                AND (
                                    (
                                        key_result.target_value >= key_result.start_value
                                        AND key_result.current_value >= key_result.target_value
                                    )
                                    OR (
                                        key_result.target_value < key_result.start_value
                                        AND key_result.current_value <= key_result.target_value
                                    )
                                )
                            )
                            OR (
                                key_result.measurement_type = 'boolean'
                                AND key_result.current_value = key_result.target_value
                            )
                        )
                ),
                CAST('[]' AS jsonb)
            )
            ELSE CAST('[]' AS jsonb)
        END AS key_results
    FROM objectives AS objective
    JOIN users AS lead
        ON lead.user_id = objective.lead_user_id
    JOIN workspaces AS workspace
        ON workspace.workspace_id = objective.workspace_id
    JOIN workspace_members AS membership
        ON membership.workspace_id = objective.workspace_id
        AND membership.user_id = objective.lead_user_id
        AND membership.role IN ('admin', 'member', 'guest')
    JOIN objective_statuses AS status
        ON status.status_id = objective.status_id
    JOIN workspace_settings AS settings
        ON settings.workspace_id = objective.workspace_id
    LEFT JOIN notification_preferences AS preference
        ON preference.user_id = objective.lead_user_id
        AND preference.workspace_id = objective.workspace_id
    WHERE objective.lead_user_id = sqlc.arg(lead_user_id)
        AND objective.workspace_id = sqlc.arg(workspace_id)
        AND workspace.deleted_at IS NULL
        AND (
            membership.role = 'admin'
            OR EXISTS (
                SELECT 1
                FROM team_members AS team_membership
                WHERE team_membership.team_id = objective.team_id
                    AND team_membership.user_id = objective.lead_user_id
            )
        )
        AND (
            objective.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '7 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days'
            OR (
                settings.key_result_enabled = true
                AND EXISTS (
                    SELECT 1
                    FROM key_results AS key_result
                    WHERE key_result.objective_id = objective.objective_id
                        AND key_result.end_date BETWEEN CAST(sqlc.arg(as_of) AS date) - INTERVAL '7 days' AND CAST(sqlc.arg(as_of) AS date) + INTERVAL '7 days'
                        AND NOT (
                            (
                                key_result.measurement_type IN ('percentage', 'number')
                                AND (
                                    (
                                        key_result.target_value >= key_result.start_value
                                        AND key_result.current_value >= key_result.target_value
                                    )
                                    OR (
                                        key_result.target_value < key_result.start_value
                                        AND key_result.current_value <= key_result.target_value
                                    )
                                )
                            )
                            OR (
                                key_result.measurement_type = 'boolean'
                                AND key_result.current_value = key_result.target_value
                            )
                        )
                )
            )
        )
        AND lead.is_active = true
        AND lead.is_system = false
        AND NULLIF(TRIM(lead.email), '') IS NOT NULL
        AND status.category NOT IN ('completed', 'cancelled', 'paused')
        AND settings.objective_enabled = true
        AND CAST(COALESCE(preference.preferences -> 'reminders' ->> 'email', 'true') AS boolean) = true
)
SELECT
    objective_id,
    name,
    end_date,
    lead_user_id,
    workspace_id,
    team_id,
    lead_email,
    lead_name,
    workspace_name,
    workspace_slug,
    deadline_status,
    days_difference,
    key_results
FROM objective_deadlines
WHERE deadline_status IN ('due_today', 'due_tomorrow', 'due_in_7_days', 'overdue')
    OR jsonb_array_length(key_results) > 0
ORDER BY deadline_status, end_date, objective_id;
