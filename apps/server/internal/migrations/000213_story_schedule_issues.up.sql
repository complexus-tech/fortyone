-- An unresolved scheduling problem survives transient planning state and
-- changing reservations. Its identity changes only after resolution or when
-- the affected owner or actionable cause changes.
CREATE TABLE public.story_schedule_issues (
    workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    story_id uuid NOT NULL REFERENCES public.stories(id) ON DELETE CASCADE,
    issue_id uuid NOT NULL UNIQUE,
    owner_id uuid NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    cause_code text NOT NULL,
    opened_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    resolved_at timestamptz,
    PRIMARY KEY (workspace_id, story_id),
    CONSTRAINT story_schedule_issues_cause_check CHECK (length(btrim(cause_code)) BETWEEN 1 AND 100)
);
