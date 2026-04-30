ALTER TABLE public.integration_requests
    DROP CONSTRAINT IF EXISTS integration_requests_assignee_id_fkey,
    DROP CONSTRAINT IF EXISTS integration_requests_status_id_fkey,
    DROP COLUMN IF EXISTS assignee_id,
    DROP COLUMN IF EXISTS priority,
    DROP COLUMN IF EXISTS status_id;
