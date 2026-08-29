CREATE INDEX idx_users_admin_created_id
    ON public.users USING btree (created_at DESC, user_id DESC);

CREATE INDEX idx_workspaces_admin_created_id
    ON public.workspaces USING btree (created_at DESC, workspace_id DESC);

CREATE INDEX idx_admin_audit_logs_created_id
    ON public.admin_audit_logs USING btree (created_at DESC, id DESC);

CREATE INDEX idx_admin_notes_created_id
    ON public.admin_notes USING btree (created_at DESC, id DESC);
