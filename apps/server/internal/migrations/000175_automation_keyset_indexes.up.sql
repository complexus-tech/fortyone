-- Scheduled automation scans globally rather than through a tenant prefix.
-- These partial/keyset indexes match the jobs' stable ordering and eligibility
-- predicates so bounded batches do not degrade into full-table sorts as the
-- number of workspaces grows.
CREATE INDEX idx_stories_automation_updated_page
    ON public.stories (updated_at, id)
    WHERE deleted_at IS NULL
      AND archived_at IS NULL;

CREATE INDEX idx_team_sprint_settings_auto_create_page
    ON public.team_sprint_settings (workspace_id, team_id)
    INCLUDE (updated_at)
    WHERE auto_create_sprints = TRUE;

CREATE INDEX idx_sprints_automation_end_page
    ON public.sprints (end_date, sprint_id, workspace_id, team_id);
