CREATE TABLE public.admin_notes (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    target_type text NOT NULL,
    target_id uuid NOT NULL,
    workspace_id uuid,
    body text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT admin_notes_pkey PRIMARY KEY (id),
    CONSTRAINT admin_notes_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE RESTRICT,
    CONSTRAINT admin_notes_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE SET NULL,
    CONSTRAINT admin_notes_target_type_check
        CHECK (target_type IN ('workspace', 'user')),
    CONSTRAINT admin_notes_body_check
        CHECK (length(trim(body)) > 0)
);

CREATE INDEX idx_admin_notes_target_created
    ON public.admin_notes USING btree (target_type, target_id, created_at DESC);

CREATE INDEX idx_admin_notes_workspace_created
    ON public.admin_notes USING btree (workspace_id, created_at DESC)
    WHERE workspace_id IS NOT NULL;
