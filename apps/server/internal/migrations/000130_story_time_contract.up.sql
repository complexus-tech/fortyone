ALTER TABLE public.stories
    ADD COLUMN estimated_duration_minutes integer,
    ADD COLUMN minimum_focus_block_minutes integer,
    ADD CONSTRAINT stories_estimated_duration_minutes_check
        CHECK (
            estimated_duration_minutes IS NULL
            OR estimated_duration_minutes BETWEEN 1 AND 2400
        ),
    ADD CONSTRAINT stories_minimum_focus_block_minutes_check
        CHECK (
            minimum_focus_block_minutes IS NULL
            OR minimum_focus_block_minutes BETWEEN 1 AND 2400
        ),
    ADD CONSTRAINT stories_focus_block_requires_duration_check
        CHECK (minimum_focus_block_minutes IS NULL OR estimated_duration_minutes IS NOT NULL),
    ADD CONSTRAINT stories_focus_block_within_duration_check
        CHECK (
            estimated_duration_minutes IS NULL
            OR minimum_focus_block_minutes IS NULL
            OR minimum_focus_block_minutes <= estimated_duration_minutes
        );

ALTER TABLE public.integration_requests
    ADD COLUMN estimated_duration_minutes integer,
    ADD COLUMN minimum_focus_block_minutes integer,
    ADD CONSTRAINT integration_requests_estimated_duration_minutes_check
        CHECK (
            estimated_duration_minutes IS NULL
            OR estimated_duration_minutes BETWEEN 1 AND 2400
        ),
    ADD CONSTRAINT integration_requests_minimum_focus_block_minutes_check
        CHECK (
            minimum_focus_block_minutes IS NULL
            OR minimum_focus_block_minutes BETWEEN 1 AND 2400
        ),
    ADD CONSTRAINT integration_requests_focus_block_requires_duration_check
        CHECK (minimum_focus_block_minutes IS NULL OR estimated_duration_minutes IS NOT NULL),
    ADD CONSTRAINT integration_requests_focus_block_within_duration_check
        CHECK (
            estimated_duration_minutes IS NULL
            OR minimum_focus_block_minutes IS NULL
            OR minimum_focus_block_minutes <= estimated_duration_minutes
        );

UPDATE public.stories AS story
SET
    estimated_duration_minutes = CASE story.estimate_unit
        WHEN 1 THEN 30
        WHEN 2 THEN 60
        WHEN 3 THEN 120
        WHEN 5 THEN 240
        WHEN 8 THEN 480
    END,
    estimate_unit = NULL
FROM public.teams AS team
LEFT JOIN public.team_estimation_settings AS settings
    ON settings.team_id = team.team_id
WHERE story.team_id = team.team_id
    AND story.workspace_id = team.workspace_id
    AND COALESCE(settings.scheme, 'hours') = 'hours'
    AND story.estimate_unit IS NOT NULL;

UPDATE public.stories AS story
SET
    estimated_duration_minutes = CASE story.estimate_unit
        WHEN 1 THEN 240
        WHEN 2 THEN 480
        WHEN 3 THEN 960
        WHEN 5 THEN 1440
        WHEN 8 THEN 2400
    END,
    estimate_unit = NULL
FROM public.teams AS team
INNER JOIN public.team_estimation_settings AS settings
    ON settings.team_id = team.team_id
WHERE story.team_id = team.team_id
    AND story.workspace_id = team.workspace_id
    AND settings.scheme = 'ideal_days'
    AND story.estimate_unit IS NOT NULL;

UPDATE public.integration_requests AS request
SET
    estimated_duration_minutes = CASE request.estimate_unit
        WHEN 1 THEN 30
        WHEN 2 THEN 60
        WHEN 3 THEN 120
        WHEN 5 THEN 240
        WHEN 8 THEN 480
    END,
    estimate_unit = NULL
FROM public.teams AS team
LEFT JOIN public.team_estimation_settings AS settings
    ON settings.team_id = team.team_id
WHERE request.team_id = team.team_id
    AND request.workspace_id = team.workspace_id
    AND COALESCE(settings.scheme, 'hours') = 'hours'
    AND request.estimate_unit IS NOT NULL;

UPDATE public.integration_requests AS request
SET
    estimated_duration_minutes = CASE request.estimate_unit
        WHEN 1 THEN 240
        WHEN 2 THEN 480
        WHEN 3 THEN 960
        WHEN 5 THEN 1440
        WHEN 8 THEN 2400
    END,
    estimate_unit = NULL
FROM public.teams AS team
INNER JOIN public.team_estimation_settings AS settings
    ON settings.team_id = team.team_id
WHERE request.team_id = team.team_id
    AND request.workspace_id = team.workspace_id
    AND settings.scheme = 'ideal_days'
    AND request.estimate_unit IS NOT NULL;

UPDATE public.team_estimation_settings
SET
    scheme = 'tshirt',
    updated_at = NOW()
WHERE scheme IN ('hours', 'ideal_days');

ALTER TABLE public.team_estimation_settings
    ALTER COLUMN scheme SET DEFAULT 'tshirt',
    DROP CONSTRAINT team_estimation_settings_scheme_check,
    ADD CONSTRAINT team_estimation_settings_scheme_check
        CHECK (scheme IN ('points', 'tshirt'));
