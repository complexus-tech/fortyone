DROP INDEX public.chat_mutation_approval_executions_lease_expiry_idx;
DROP INDEX public.chat_mutation_approval_executions_unresolved_fingerprint_key;

ALTER TABLE public.chat_mutation_approval_executions
    DROP CONSTRAINT chat_mutation_approval_executions_lifecycle_check,
    DROP CONSTRAINT chat_mutation_approval_executions_status_check,
    DROP CONSTRAINT chat_mutation_approval_reconciliation_resolution_check;

ALTER TABLE public.chat_mutation_approval_executions
    ADD CONSTRAINT chat_mutation_approval_executions_status_check
        CHECK (status IN (
            'ready',
            'retry_ready',
            'executing',
            'completed',
            'failed_uncertain'
        )),
    ADD CONSTRAINT chat_mutation_approval_reconciliation_resolution_check
        CHECK (
            last_reconciliation_resolution IS NULL
            OR last_reconciliation_resolution IN (
                'verified_completed',
                'verified_not_applied',
                'safe_retry_prepared'
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
                status = 'retry_ready'
                AND output IS NULL
                AND completed_at IS NULL
                AND (
                    (lease_token IS NULL AND lease_expires_at IS NULL)
                    OR
                    (lease_token IS NOT NULL AND lease_expires_at IS NOT NULL)
                )
                AND started_at IS NULL
                AND failed_at IS NOT NULL
                AND failure_code IS NOT NULL
                AND attempt_count >= 1
                AND reconciliation_count = 1
                AND last_reconciliation_resolution = 'safe_retry_prepared'
                AND last_reconciliation_evidence IS NOT NULL
                AND jsonb_typeof(last_reconciliation_evidence) = 'object'
                AND last_reconciled_at IS NOT NULL
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

-- A prepared retry remains globally quarantined. Only its original transcript
-- and ledger identity can lease it; a new chat or tool call cannot transfer it.
CREATE UNIQUE INDEX chat_mutation_approval_executions_unresolved_fingerprint_key
    ON public.chat_mutation_approval_executions (
        user_id,
        workspace_id,
        fingerprint
    )
    WHERE status IN ('ready', 'retry_ready', 'executing', 'failed_uncertain');

CREATE INDEX chat_mutation_approval_executions_lease_expiry_idx
    ON public.chat_mutation_approval_executions (lease_expires_at)
    WHERE status IN ('ready', 'retry_ready', 'executing');
