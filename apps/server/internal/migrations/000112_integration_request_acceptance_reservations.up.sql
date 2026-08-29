ALTER TABLE public.integration_requests
    ADD COLUMN acceptance_state text NOT NULL DEFAULT 'idle',
    ADD COLUMN acceptance_started_by_user_id uuid,
    ADD COLUMN acceptance_started_at timestamptz,
    ADD CONSTRAINT integration_requests_acceptance_state_check
        CHECK (acceptance_state IN ('idle', 'reserved')),
    ADD CONSTRAINT integration_requests_acceptance_started_by_user_id_fkey
        FOREIGN KEY (acceptance_started_by_user_id) REFERENCES public.users(user_id),
    ADD CONSTRAINT integration_requests_acceptance_lifecycle_check
        CHECK (
            (acceptance_state = 'idle'
                AND acceptance_started_by_user_id IS NULL
                AND acceptance_started_at IS NULL)
            OR
            (acceptance_state = 'reserved'
                AND status = 'pending'
                AND acceptance_started_by_user_id IS NOT NULL
                AND acceptance_started_at IS NOT NULL)
        );

UPDATE public.integration_requests
SET priority = CASE lower(btrim(priority))
    WHEN 'no priority' THEN 'No Priority'
    WHEN 'low' THEN 'Low'
    WHEN 'medium' THEN 'Medium'
    WHEN 'high' THEN 'High'
    WHEN 'urgent' THEN 'Urgent'
    ELSE 'No Priority'
END;

ALTER TABLE public.integration_requests
    ADD CONSTRAINT integration_requests_priority_check
        CHECK (priority IN ('No Priority', 'Low', 'Medium', 'High', 'Urgent'));

CREATE INDEX integration_requests_acceptance_recovery_idx
    ON public.integration_requests USING btree (workspace_id, acceptance_started_at)
    WHERE status = 'pending' AND acceptance_state = 'reserved';
