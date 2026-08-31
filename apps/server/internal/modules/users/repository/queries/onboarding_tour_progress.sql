-- name: GetOrCreateOnboardingTourProgressForMember :one
INSERT INTO public.user_onboarding_tour_progress (
    user_id,
    workspace_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at
)
SELECT
    membership.user_id,
    membership.workspace_id,
    CAST(sqlc.arg(tour_key) AS text),
    CAST(sqlc.arg(tour_version) AS text),
    ARRAY[]::text[],
    ARRAY[]::text[],
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE membership.user_id = sqlc.arg(user_id)
  AND membership.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, workspace_id, tour_key, tour_version) DO UPDATE
SET tour_key = user_onboarding_tour_progress.tour_key
RETURNING
    user_id,
    workspace_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at;

-- name: UpsertOnboardingTourProgressForMember :one
INSERT INTO public.user_onboarding_tour_progress (
    user_id,
    workspace_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at
)
SELECT
    membership.user_id,
    membership.workspace_id,
    CAST(sqlc.arg(tour_key) AS text),
    CAST(sqlc.arg(tour_version) AS text),
    CAST(sqlc.arg(completed_step_ids) AS text[]),
    CAST(sqlc.arg(completed_action_ids) AS text[]),
    CASE
        WHEN CAST(sqlc.arg(set_status) AS boolean)
            THEN CAST(sqlc.arg(status) AS text)
        ELSE 'active'
    END,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.workspace_members AS membership
INNER JOIN public.users AS account ON account.user_id = membership.user_id
WHERE membership.user_id = sqlc.arg(user_id)
  AND membership.workspace_id = sqlc.arg(workspace_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, workspace_id, tour_key, tour_version) DO UPDATE
SET
    completed_step_ids = ARRAY(
        SELECT DISTINCT step_id.value
        FROM unnest(
            user_onboarding_tour_progress.completed_step_ids || EXCLUDED.completed_step_ids
        ) AS step_id(value)
        ORDER BY step_id.value
    ),
    completed_action_ids = ARRAY(
        SELECT DISTINCT action_id.value
        FROM unnest(
            user_onboarding_tour_progress.completed_action_ids || EXCLUDED.completed_action_ids
        ) AS action_id(value)
        ORDER BY action_id.value
    ),
    status = CASE
        WHEN user_onboarding_tour_progress.status IN ('completed', 'skipped')
            THEN user_onboarding_tour_progress.status
        WHEN CAST(sqlc.arg(set_status) AS boolean)
            THEN EXCLUDED.status
        ELSE user_onboarding_tour_progress.status
    END,
    updated_at = CURRENT_TIMESTAMP
RETURNING
    user_id,
    workspace_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at;
