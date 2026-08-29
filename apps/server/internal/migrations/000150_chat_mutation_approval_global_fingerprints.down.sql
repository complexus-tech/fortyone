DROP INDEX public.chat_mutation_approval_executions_unresolved_fingerprint_key;

CREATE UNIQUE INDEX chat_mutation_approval_executions_unresolved_fingerprint_key
    ON public.chat_mutation_approval_executions (
        session_id,
        user_id,
        workspace_id,
        fingerprint
    )
    WHERE status IN ('ready', 'executing', 'failed_uncertain');
