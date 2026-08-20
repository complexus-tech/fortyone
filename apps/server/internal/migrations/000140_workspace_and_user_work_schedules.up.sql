ALTER TABLE public.workspace_settings
    ADD COLUMN working_days smallint[] NOT NULL DEFAULT ARRAY[1, 2, 3, 4, 5]::smallint[],
    ADD COLUMN working_start_minute smallint NOT NULL DEFAULT 540,
    ADD COLUMN working_end_minute smallint NOT NULL DEFAULT 1020,
    ADD CONSTRAINT workspace_settings_working_days_check CHECK (
        cardinality(working_days) BETWEEN 1 AND 7
        AND working_days <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::smallint[]
    ),
    ADD CONSTRAINT workspace_settings_working_hours_check CHECK (
        working_start_minute BETWEEN 0 AND 1439
        AND working_end_minute BETWEEN 1 AND 1440
        AND working_end_minute > working_start_minute
    );

-- Preserve the most common existing team workweek as each workspace default.
WITH ranked_workweeks AS (
    SELECT
        settings.workspace_id,
        settings.working_days,
        ROW_NUMBER() OVER (
            PARTITION BY settings.workspace_id
            ORDER BY COUNT(*) DESC, settings.working_days::text
        ) AS position
    FROM public.team_sprint_settings AS settings
    GROUP BY settings.workspace_id, settings.working_days
)
UPDATE public.workspace_settings AS workspace_settings
SET working_days = ranked_workweeks.working_days
FROM ranked_workweeks
WHERE ranked_workweeks.workspace_id = workspace_settings.workspace_id
    AND ranked_workweeks.position = 1;

ALTER TABLE public.users
    ADD COLUMN working_days smallint[],
    ADD COLUMN working_start_minute smallint,
    ADD COLUMN working_end_minute smallint,
    ADD CONSTRAINT users_working_days_check CHECK (
        working_days IS NULL
        OR (
            cardinality(working_days) BETWEEN 1 AND 7
            AND working_days <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::smallint[]
        )
    ),
    ADD CONSTRAINT users_working_hours_check CHECK (
        (working_start_minute IS NULL AND working_end_minute IS NULL)
        OR (
            working_start_minute BETWEEN 0 AND 1439
            AND working_end_minute BETWEEN 1 AND 1440
            AND working_end_minute > working_start_minute
        )
    );
