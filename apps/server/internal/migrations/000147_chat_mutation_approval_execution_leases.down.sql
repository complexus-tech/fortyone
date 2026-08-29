DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.chat_mutation_approval_executions
        WHERE status <> 'completed'
    ) THEN
        RAISE EXCEPTION
            'cannot roll back mutation approval leases while unresolved or uncertain executions exist';
    END IF;
END;
$$;

DROP INDEX IF EXISTS public.chat_mutation_approval_executions_lease_expiry_idx;
DROP INDEX IF EXISTS public.chat_mutation_approval_executions_unresolved_fingerprint_key;

ALTER TABLE public.chat_mutation_approval_executions
    DROP CONSTRAINT chat_mutation_approval_executions_lifecycle_check,
    DROP CONSTRAINT chat_mutation_approval_executions_failure_code_check,
    DROP CONSTRAINT chat_mutation_approval_executions_attempt_count_check,
    DROP CONSTRAINT chat_mutation_approval_executions_status_check;

ALTER TABLE public.chat_mutation_approval_executions
    DROP COLUMN attempt_count,
    DROP COLUMN failure_code,
    DROP COLUMN failed_at,
    DROP COLUMN started_at,
    DROP COLUMN lease_expires_at,
    DROP COLUMN lease_token,
    ALTER COLUMN status SET DEFAULT 'in_progress';

ALTER TABLE public.chat_mutation_approval_executions
    ADD CONSTRAINT chat_mutation_approval_executions_status_check
        CHECK (status IN ('in_progress', 'completed')),
    ADD CONSTRAINT chat_mutation_approval_executions_completion_check
        CHECK (
            (status = 'in_progress' AND output IS NULL AND completed_at IS NULL)
            OR
            (status = 'completed' AND output IS NOT NULL AND completed_at IS NOT NULL)
        );
