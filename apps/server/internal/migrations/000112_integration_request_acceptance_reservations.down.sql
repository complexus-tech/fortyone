DROP INDEX IF EXISTS public.integration_requests_acceptance_recovery_idx;

ALTER TABLE public.integration_requests
    DROP CONSTRAINT IF EXISTS integration_requests_priority_check,
    DROP CONSTRAINT IF EXISTS integration_requests_acceptance_lifecycle_check,
    DROP CONSTRAINT IF EXISTS integration_requests_acceptance_started_by_user_id_fkey,
    DROP CONSTRAINT IF EXISTS integration_requests_acceptance_state_check,
    DROP COLUMN IF EXISTS acceptance_started_at,
    DROP COLUMN IF EXISTS acceptance_started_by_user_id,
    DROP COLUMN IF EXISTS acceptance_state;
