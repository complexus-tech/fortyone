-- An uncertain mutation must remain quarantined when the same user opens a
-- different chat. Refuse the migration instead of silently discarding or
-- rewriting ambiguous executions if an earlier deployment already produced
-- duplicate unresolved rows across sessions.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.chat_mutation_approval_executions
        WHERE status IN ('ready', 'executing', 'failed_uncertain')
        GROUP BY user_id, workspace_id, fingerprint
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION
            'cannot enforce cross-session mutation quarantine while duplicate unresolved fingerprints exist';
    END IF;
END;
$$;

DROP INDEX public.chat_mutation_approval_executions_unresolved_fingerprint_key;

CREATE UNIQUE INDEX chat_mutation_approval_executions_unresolved_fingerprint_key
    ON public.chat_mutation_approval_executions (
        user_id,
        workspace_id,
        fingerprint
    )
    WHERE status IN ('ready', 'executing', 'failed_uncertain');
