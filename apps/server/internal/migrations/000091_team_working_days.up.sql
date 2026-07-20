ALTER TABLE public.team_sprint_settings
    ADD COLUMN working_days smallint[] NOT NULL DEFAULT ARRAY[1, 2, 3, 4, 5]::smallint[],
    ADD CONSTRAINT team_sprint_settings_working_days_check CHECK (
        cardinality(working_days) BETWEEN 1 AND 7
        AND working_days <@ ARRAY[1, 2, 3, 4, 5, 6, 7]::smallint[]
    );
