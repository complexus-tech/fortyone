-- Scheduled retention and inactivity jobs scan globally rather than through a
-- tenant prefix. These partial indexes match each job's eligibility predicate
-- and stable keyset order so table growth does not turn bounded batches into
-- full-table sorts.
CREATE INDEX idx_stories_retention_deleted_page
    ON public.stories (deleted_at, id)
    WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_workspaces_inactivity_warning_page
    ON public.workspaces (last_accessed_at, workspace_id)
    WHERE deleted_at IS NULL
      AND last_accessed_at IS NOT NULL
      AND inactivity_warning_sent_at IS NULL;

CREATE INDEX idx_workspaces_inactivity_deletion_page
    ON public.workspaces (last_accessed_at, workspace_id)
    WHERE deleted_at IS NULL
      AND last_accessed_at IS NOT NULL
      AND inactivity_warning_sent_at IS NOT NULL;

CREATE INDEX idx_workspaces_deleted_retention_page
    ON public.workspaces (deleted_at, workspace_id)
    WHERE deleted_at IS NOT NULL;

CREATE INDEX idx_users_inactivity_warning_page
    ON public.users (last_login_at, user_id)
    WHERE is_active = TRUE
      AND is_system = FALSE
      AND last_login_at IS NOT NULL
      AND inactivity_warning_sent_at IS NULL;

CREATE INDEX idx_users_inactivity_deactivation_page
    ON public.users (inactivity_warning_sent_at, last_login_at, user_id)
    WHERE is_active = TRUE
      AND is_system = FALSE
      AND last_login_at IS NOT NULL
      AND inactivity_warning_sent_at IS NOT NULL;
