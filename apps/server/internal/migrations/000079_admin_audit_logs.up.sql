CREATE TABLE public.admin_audit_logs (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL,
    target_type text NOT NULL,
    target_id uuid,
    workspace_id uuid,
    action text NOT NULL,
    field_name text,
    old_value jsonb,
    new_value jsonb,
    reason text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT admin_audit_logs_actor_user_id_fkey
        FOREIGN KEY (actor_user_id) REFERENCES public.users(user_id) ON DELETE RESTRICT,
    CONSTRAINT admin_audit_logs_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE SET NULL,
    CONSTRAINT admin_audit_logs_target_type_check
        CHECK (target_type IN ('workspace', 'user', 'subscription', 'system'))
);

CREATE INDEX idx_admin_audit_logs_created
    ON public.admin_audit_logs USING btree (created_at DESC);

CREATE INDEX idx_admin_audit_logs_actor_created
    ON public.admin_audit_logs USING btree (actor_user_id, created_at DESC);

CREATE INDEX idx_admin_audit_logs_workspace_created
    ON public.admin_audit_logs USING btree (workspace_id, created_at DESC)
    WHERE workspace_id IS NOT NULL;

CREATE INDEX idx_admin_audit_logs_target_created
    ON public.admin_audit_logs USING btree (target_type, target_id, created_at DESC)
    WHERE target_id IS NOT NULL;
