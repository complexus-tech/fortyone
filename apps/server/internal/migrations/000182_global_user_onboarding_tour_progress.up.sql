CREATE TABLE public.user_onboarding_tour_progress_global (
    user_id uuid NOT NULL,
    tour_key text NOT NULL,
    tour_version text NOT NULL,
    completed_step_ids text[] NOT NULL DEFAULT ARRAY[]::text[],
    completed_action_ids text[] NOT NULL DEFAULT ARRAY[]::text[],
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_onboarding_tour_progress_global_pkey
        PRIMARY KEY (user_id, tour_key, tour_version),
    CONSTRAINT user_onboarding_tour_progress_global_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT user_onboarding_tour_progress_global_status_check
        CHECK (status IN ('active', 'completed', 'skipped'))
);

WITH progress_groups AS (
    SELECT
        progress.user_id,
        progress.tour_key,
        progress.tour_version,
        CASE
            WHEN BOOL_OR(progress.status = 'completed') THEN 'completed'
            WHEN BOOL_OR(progress.status = 'skipped') THEN 'skipped'
            ELSE 'active'
        END AS status,
        MIN(progress.created_at) AS created_at,
        MAX(progress.updated_at) AS updated_at
    FROM public.user_onboarding_tour_progress AS progress
    GROUP BY progress.user_id, progress.tour_key, progress.tour_version
), step_groups AS (
    SELECT
        progress.user_id,
        progress.tour_key,
        progress.tour_version,
        COALESCE(
            ARRAY_AGG(DISTINCT step.value ORDER BY step.value)
                FILTER (WHERE step.value IS NOT NULL),
            ARRAY[]::text[]
        ) AS completed_step_ids
    FROM public.user_onboarding_tour_progress AS progress
    LEFT JOIN LATERAL UNNEST(progress.completed_step_ids) AS step(value) ON TRUE
    GROUP BY progress.user_id, progress.tour_key, progress.tour_version
), action_groups AS (
    SELECT
        progress.user_id,
        progress.tour_key,
        progress.tour_version,
        COALESCE(
            ARRAY_AGG(DISTINCT action.value ORDER BY action.value)
                FILTER (WHERE action.value IS NOT NULL),
            ARRAY[]::text[]
        ) AS completed_action_ids
    FROM public.user_onboarding_tour_progress AS progress
    LEFT JOIN LATERAL UNNEST(progress.completed_action_ids) AS action(value) ON TRUE
    GROUP BY progress.user_id, progress.tour_key, progress.tour_version
)
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
    progress_groups.user_id,
    progress_groups.tour_key,
    progress_groups.tour_version,
    step_groups.completed_step_ids,
    action_groups.completed_action_ids,
    progress_groups.status,
    progress_groups.created_at,
    progress_groups.updated_at
FROM progress_groups
INNER JOIN step_groups USING (user_id, tour_key, tour_version)
INNER JOIN action_groups USING (user_id, tour_key, tour_version);

-- Before versioned progress existed, completion of the original getting-started
-- walkthrough was recorded only on the user. Preserve that terminal state so
-- those users are not enrolled again after the global progress cutover.
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
    'workspace-getting-started',
    '1.0.0',
    ARRAY[]::text[],
    ARRAY[]::text[],
    'completed',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM public.users AS account
WHERE account.has_seen_walkthrough = TRUE
ON CONFLICT (user_id, tour_key, tour_version) DO UPDATE
SET
    status = 'completed',
    updated_at = GREATEST(
        user_onboarding_tour_progress_global.updated_at,
        EXCLUDED.updated_at
    );

-- Active Sprints previously shared the workspace-module-team key. Seed its new
-- dedicated key from that legacy state so this route split cannot replay a tour
-- a user already completed or dismissed.
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
    progress.user_id,
    'workspace-module-sprints',
    progress.tour_version,
    progress.completed_step_ids,
    progress.completed_action_ids,
    progress.status,
    progress.created_at,
    progress.updated_at
FROM public.user_onboarding_tour_progress_global AS progress
WHERE progress.tour_key = 'workspace-module-team'
  AND progress.tour_version = '1.0.0'
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
    created_at = LEAST(
        user_onboarding_tour_progress_global.created_at,
        EXCLUDED.created_at
    ),
    updated_at = GREATEST(
        user_onboarding_tour_progress_global.updated_at,
        EXCLUDED.updated_at
    );

CREATE FUNCTION public.mirror_user_onboarding_tour_progress_global()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
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
    VALUES (
        NEW.user_id,
        NEW.tour_key,
        NEW.tour_version,
        NEW.completed_step_ids,
        NEW.completed_action_ids,
        NEW.status,
        NEW.created_at,
        NEW.updated_at
    )
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
        created_at = LEAST(
            user_onboarding_tour_progress_global.created_at,
            EXCLUDED.created_at
        ),
        updated_at = GREATEST(
            user_onboarding_tour_progress_global.updated_at,
            EXCLUDED.updated_at
        );

    RETURN NEW;
END;
$$;

CREATE TRIGGER mirror_user_onboarding_tour_progress_global
AFTER INSERT OR UPDATE OF
    completed_step_ids,
    completed_action_ids,
    status,
    created_at,
    updated_at
ON public.user_onboarding_tour_progress
FOR EACH ROW
EXECUTE FUNCTION public.mirror_user_onboarding_tour_progress_global();
