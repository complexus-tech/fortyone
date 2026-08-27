ALTER TABLE public.chat_mutation_approval_executions
    ADD COLUMN lease_token uuid,
    ADD COLUMN lease_expires_at timestamptz,
    ADD COLUMN started_at timestamptz,
    ADD COLUMN failed_at timestamptz,
    ADD COLUMN failure_code text,
    ADD COLUMN attempt_count integer NOT NULL DEFAULT 0;

ALTER TABLE public.chat_mutation_approval_executions
    DROP CONSTRAINT chat_mutation_approval_executions_status_check,
    DROP CONSTRAINT chat_mutation_approval_executions_completion_check;

-- Version 146 could not distinguish a process that crashed before its tool call
-- from one that crashed after the mutation. Preserve those rows as terminal and
-- uncertain so an upgrade can never replay them.
UPDATE public.chat_mutation_approval_executions
SET status = 'failed_uncertain',
    failed_at = CURRENT_TIMESTAMP,
    failure_code = 'legacy_in_progress',
    updated_at = CURRENT_TIMESTAMP
WHERE status = 'in_progress';

ALTER TABLE public.chat_mutation_approval_executions
    ALTER COLUMN status SET DEFAULT 'ready',
    ADD CONSTRAINT chat_mutation_approval_executions_status_check
        CHECK (status IN ('ready', 'executing', 'completed', 'failed_uncertain')),
    ADD CONSTRAINT chat_mutation_approval_executions_attempt_count_check
        CHECK (attempt_count >= 0),
    ADD CONSTRAINT chat_mutation_approval_executions_failure_code_check
        CHECK (
            failure_code IS NULL
            OR char_length(btrim(failure_code)) BETWEEN 1 AND 64
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

-- A freshly prepared tool call may not bypass an unresolved identical mutation
-- by choosing a new tool_call_id. Completed operations are excluded so a user
-- may intentionally perform the same action again after a known result.
CREATE UNIQUE INDEX chat_mutation_approval_executions_unresolved_fingerprint_key
    ON public.chat_mutation_approval_executions (
        session_id,
        user_id,
        workspace_id,
        fingerprint
    )
    WHERE status IN ('ready', 'executing', 'failed_uncertain');

CREATE INDEX chat_mutation_approval_executions_lease_expiry_idx
    ON public.chat_mutation_approval_executions (lease_expires_at)
    WHERE status IN ('ready', 'executing');
