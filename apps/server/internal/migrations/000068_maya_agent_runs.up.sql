CREATE TABLE public.maya_agent_runs (
    run_id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    triggered_by_user_id uuid NOT NULL,
    trigger_type varchar(32) NOT NULL,
    status varchar(32) NOT NULL,
    summary text NOT NULL DEFAULT '',
    context jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_message text,
    started_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT maya_agent_runs_pkey PRIMARY KEY (run_id),
    CONSTRAINT maya_agent_runs_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_runs_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_runs_triggered_by_user_id_fkey
        FOREIGN KEY (triggered_by_user_id) REFERENCES public.users(user_id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_runs_trigger_type_check
        CHECK (trigger_type IN ('manual', 'event')),
    CONSTRAINT maya_agent_runs_status_check
        CHECK (status IN ('running', 'succeeded', 'failed'))
);

CREATE TABLE public.maya_agent_actions (
    action_id uuid NOT NULL DEFAULT gen_random_uuid(),
    run_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    action_type varchar(64) NOT NULL,
    status varchar(32) NOT NULL,
    reason text NOT NULL DEFAULT '',
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_message text,
    applied_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT maya_agent_actions_pkey PRIMARY KEY (action_id),
    CONSTRAINT maya_agent_actions_run_id_fkey
        FOREIGN KEY (run_id) REFERENCES public.maya_agent_runs(run_id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_actions_workspace_id_fkey
        FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_actions_story_id_fkey
        FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT maya_agent_actions_type_check
        CHECK (action_type IN ('assign_story', 'schedule_work_block', 'flag_schedule_risk')),
    CONSTRAINT maya_agent_actions_status_check
        CHECK (status IN ('proposed', 'applied', 'failed'))
);

CREATE INDEX idx_maya_agent_runs_workspace_story_created
    ON public.maya_agent_runs (workspace_id, story_id, created_at DESC);

CREATE INDEX idx_maya_agent_actions_run
    ON public.maya_agent_actions (run_id, created_at ASC);

CREATE INDEX idx_maya_agent_actions_workspace_status
    ON public.maya_agent_actions (workspace_id, status, created_at DESC);
