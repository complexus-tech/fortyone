-- name: GetOrCreateOnboardingTourProgressForUser :one
INSERT INTO public.user_onboarding_tour_progress_global (
    user_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at
)
SELECT
    account.user_id,
    CAST(sqlc.arg(tour_key) AS text),
    CAST(sqlc.arg(tour_version) AS text),
    ARRAY[]::text[],
    ARRAY[]::text[],
    'active',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.users AS account
WHERE account.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, tour_key, tour_version) DO UPDATE
SET tour_key = user_onboarding_tour_progress_global.tour_key
RETURNING
    user_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at;

-- name: UpsertOnboardingTourProgressForUser :one
INSERT INTO public.user_onboarding_tour_progress_global (
    user_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at
)
SELECT
    account.user_id,
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
FROM public.users AS account
WHERE account.user_id = sqlc.arg(user_id)
  AND account.is_active = TRUE
ON CONFLICT (user_id, tour_key, tour_version) DO UPDATE
SET
    completed_step_ids = ARRAY(
        SELECT DISTINCT step_id.value
        FROM UNNEST(
            user_onboarding_tour_progress_global.completed_step_ids ||
            EXCLUDED.completed_step_ids
        ) AS step_id(value)
        ORDER BY step_id.value
    ),
    completed_action_ids = ARRAY(
        SELECT DISTINCT action_id.value
        FROM UNNEST(
            user_onboarding_tour_progress_global.completed_action_ids ||
            EXCLUDED.completed_action_ids
        ) AS action_id(value)
        ORDER BY action_id.value
    ),
    status = CASE
        WHEN user_onboarding_tour_progress_global.status = 'completed'
            OR EXCLUDED.status = 'completed'
            THEN 'completed'
        WHEN user_onboarding_tour_progress_global.status = 'skipped'
            OR EXCLUDED.status = 'skipped'
            THEN 'skipped'
        ELSE 'active'
    END,
    updated_at = CURRENT_TIMESTAMP
RETURNING
    user_id,
    tour_key,
    tour_version,
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at;
