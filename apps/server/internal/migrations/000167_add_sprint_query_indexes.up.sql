-- Sprint lists are always tenant scoped and use a stable end-date/id order.
CREATE INDEX idx_sprints_workspace_end_id
    ON sprints (workspace_id, end_date DESC, sprint_id DESC);

-- Team-filtered sprint lists are a common navigation path.
CREATE INDEX idx_sprints_workspace_team_end_id
    ON sprints (workspace_id, team_id, end_date DESC, sprint_id DESC);

-- Sprint summaries aggregate active stories by tenant, sprint, and status.
CREATE INDEX idx_stories_workspace_sprint_status_active
    ON stories (workspace_id, sprint_id, status_id)
    WHERE deleted_at IS NULL
      AND archived_at IS NULL
      AND sprint_id IS NOT NULL;
