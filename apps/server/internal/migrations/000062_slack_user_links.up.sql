CREATE TABLE public.slack_user_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    slack_workspace_id uuid NOT NULL,
    slack_team_id text NOT NULL,
    slack_user_id text NOT NULL,
    user_id uuid NOT NULL,
    linked_via text NOT NULL DEFAULT 'email_match',
    linked_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT slack_user_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT slack_user_links_slack_workspace_id_fkey FOREIGN KEY (slack_workspace_id) REFERENCES public.slack_workspaces(id) ON DELETE CASCADE,
    CONSTRAINT slack_user_links_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX slack_user_links_workspace_team_user_key
    ON public.slack_user_links USING btree (workspace_id, slack_team_id, slack_user_id);

CREATE INDEX idx_slack_user_links_workspace_user_id
    ON public.slack_user_links USING btree (workspace_id, user_id);

CREATE INDEX idx_slack_user_links_slack_workspace
    ON public.slack_user_links USING btree (slack_workspace_id, slack_user_id);
