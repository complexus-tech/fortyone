DROP TABLE IF EXISTS public.team_key_result_sequences;

ALTER TABLE public.key_results
    DROP CONSTRAINT IF EXISTS key_results_team_sequence_unique,
    DROP CONSTRAINT IF EXISTS key_results_team_id_fkey,
    DROP COLUMN IF EXISTS sequence_id,
    DROP COLUMN IF EXISTS team_id;
