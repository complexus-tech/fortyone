ALTER TABLE public.chat_mutation_approval_executions
    ADD COLUMN last_reconciliation_resolution text,
    ADD COLUMN last_reconciliation_evidence jsonb,
    ADD COLUMN last_reconciled_at timestamptz,
    ADD COLUMN reconciliation_count integer NOT NULL DEFAULT 0,
    ADD CONSTRAINT chat_mutation_approval_reconciliation_resolution_check
        CHECK (
            last_reconciliation_resolution IS NULL
            OR last_reconciliation_resolution IN (
                'verified_completed',
                'verified_not_applied'
            )
        ),
    ADD CONSTRAINT chat_mutation_approval_reconciliation_count_check
        CHECK (reconciliation_count >= 0),
    ADD CONSTRAINT chat_mutation_approval_reconciliation_audit_check
        CHECK (
            (
                reconciliation_count = 0
                AND last_reconciliation_resolution IS NULL
                AND last_reconciliation_evidence IS NULL
                AND last_reconciled_at IS NULL
            )
            OR
            (
                reconciliation_count >= 1
                AND last_reconciliation_resolution IS NOT NULL
                AND last_reconciliation_evidence IS NOT NULL
                AND jsonb_typeof(last_reconciliation_evidence) = 'object'
                AND last_reconciled_at IS NOT NULL
            )
        );
