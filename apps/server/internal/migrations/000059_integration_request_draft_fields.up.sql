ALTER TABLE public.integration_requests
    ADD COLUMN status_id uuid,
    ADD COLUMN priority text NOT NULL DEFAULT 'No Priority',
    ADD COLUMN assignee_id uuid,
    ADD CONSTRAINT integration_requests_status_id_fkey FOREIGN KEY (status_id) REFERENCES public.statuses(status_id) ON DELETE SET NULL,
    ADD CONSTRAINT integration_requests_assignee_id_fkey FOREIGN KEY (assignee_id) REFERENCES public.users(user_id) ON DELETE SET NULL;
