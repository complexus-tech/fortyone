DROP INDEX IF EXISTS public.idx_integration_requests_sprint_id;
DROP INDEX IF EXISTS public.idx_integration_requests_objective_id;

ALTER TABLE public.integration_requests
    DROP CONSTRAINT IF EXISTS integration_requests_sprint_id_fkey,
    DROP CONSTRAINT IF EXISTS integration_requests_key_result_id_fkey,
    DROP CONSTRAINT IF EXISTS integration_requests_objective_id_fkey,
    DROP CONSTRAINT IF EXISTS integration_requests_estimate_unit_check,
    DROP COLUMN IF EXISTS end_date,
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS sprint_id,
    DROP COLUMN IF EXISTS key_result_id,
    DROP COLUMN IF EXISTS objective_id,
    DROP COLUMN IF EXISTS estimate_unit;
