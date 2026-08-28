DROP INDEX IF EXISTS public.idx_attachments_optimization_queue;

ALTER TABLE public.attachments
    DROP CONSTRAINT IF EXISTS attachments_optimization_attempts_check,
    DROP CONSTRAINT IF EXISTS attachments_optimization_status_check,
    DROP CONSTRAINT IF EXISTS attachments_scan_status_check,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS optimization_last_error,
    DROP COLUMN IF EXISTS optimization_lease_expires_at,
    DROP COLUMN IF EXISTS optimization_completed_at,
    DROP COLUMN IF EXISTS optimization_started_at,
    DROP COLUMN IF EXISTS optimization_attempts,
    DROP COLUMN IF EXISTS optimization_status,
    DROP COLUMN IF EXISTS scan_failure_reason,
    DROP COLUMN IF EXISTS scan_completed_at,
    DROP COLUMN IF EXISTS scan_status;
