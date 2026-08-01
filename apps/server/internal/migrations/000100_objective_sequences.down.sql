DROP TABLE IF EXISTS public.team_objective_sequences;

ALTER TABLE public.objectives
    DROP CONSTRAINT IF EXISTS objectives_team_sequence_unique,
    DROP COLUMN IF EXISTS sequence_id;
