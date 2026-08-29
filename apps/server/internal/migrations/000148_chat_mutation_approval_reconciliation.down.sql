ALTER TABLE public.chat_mutation_approval_executions
    DROP CONSTRAINT chat_mutation_approval_reconciliation_audit_check,
    DROP CONSTRAINT chat_mutation_approval_reconciliation_count_check,
    DROP CONSTRAINT chat_mutation_approval_reconciliation_resolution_check,
    DROP COLUMN reconciliation_count,
    DROP COLUMN last_reconciled_at,
    DROP COLUMN last_reconciliation_evidence,
    DROP COLUMN last_reconciliation_resolution;
