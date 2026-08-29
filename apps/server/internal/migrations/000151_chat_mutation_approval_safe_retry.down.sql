-- Once a safe retry has been prepared, migration 150 has no truthful state in
-- which to preserve its one-retry audit. Refuse to erase that security state.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.chat_mutation_approval_executions
        WHERE status = 'retry_ready'
            OR last_reconciliation_resolution = 'safe_retry_prepared'
    ) THEN
        RAISE EXCEPTION
            'cannot roll back safe mutation retries after retry preparation has been used';
    END IF;
END;
$$;

DROP INDEX public.chat_mutation_approval_executions_lease_expiry_idx;
DROP INDEX public.chat_mutation_approval_executions_unresolved_fingerprint_key;

ALTER TABLE public.chat_mutation_approval_executions
    DROP CONSTRAINT chat_mutation_approval_executions_lifecycle_check,
    DROP CONSTRAINT chat_mutation_approval_executions_status_check,
    DROP CONSTRAINT chat_mutation_approval_reconciliation_resolution_check;

ALTER TABLE public.chat_mutation_approval_executions
    ADD CONSTRAINT chat_mutation_approval_executions_status_check
        CHECK (status IN ('ready', 'executing', 'completed', 'failed_uncertain')),
    ADD CONSTRAINT chat_mutation_approval_reconciliation_resolution_check
        CHECK (
            last_reconciliation_resolution IS NULL
            OR last_reconciliation_resolution IN (
                'verified_completed',
                'verified_not_applied'
            )
        ),
    ADD CONSTRAINT chat_mutation_approval_executions_lifecycle_check
        CHECK (
            (
                status = 'ready'
                AND output IS NULL
                AND completed_at IS NULL
                AND lease_token IS NOT NULL
                AND lease_expires_at IS NOT NULL
                AND started_at IS NULL
                AND failed_at IS NULL
                AND failure_code IS NULL
                AND attempt_count >= 1
            )
            OR
            (
                status = 'executing'
                AND output IS NULL
                AND completed_at IS NULL
                AND lease_token IS NOT NULL
                AND lease_expires_at IS NOT NULL
                AND started_at IS NOT NULL
                AND failed_at IS NULL
                AND failure_code IS NULL
                AND attempt_count >= 1
            )
            OR
            (
                status = 'completed'
                AND output IS NOT NULL
                AND completed_at IS NOT NULL
                AND lease_token IS NULL
                AND lease_expires_at IS NULL
                AND failed_at IS NULL
                AND failure_code IS NULL
            )
            OR
            (
                status = 'failed_uncertain'
                AND output IS NULL
                AND completed_at IS NULL
                AND lease_token IS NULL
                AND lease_expires_at IS NULL
                AND failed_at IS NOT NULL
                AND failure_code IS NOT NULL
            )
        );

CREATE UNIQUE INDEX chat_mutation_approval_executions_unresolved_fingerprint_key
    ON public.chat_mutation_approval_executions (
        user_id,
        workspace_id,
        fingerprint
    )
    WHERE status IN ('ready', 'executing', 'failed_uncertain');

CREATE INDEX chat_mutation_approval_executions_lease_expiry_idx
    ON public.chat_mutation_approval_executions (lease_expires_at)
    WHERE status IN ('ready', 'executing');
