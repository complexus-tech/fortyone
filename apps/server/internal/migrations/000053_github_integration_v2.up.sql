-- 000053_github_integration_v2.up.sql

DROP TABLE IF EXISTS public.github_commits CASCADE;
DROP TABLE IF EXISTS public.story_github_links CASCADE;
DROP TABLE IF EXISTS public.team_github_automation_settings CASCADE;
DROP TABLE IF EXISTS public.repository_team_assignments CASCADE;
DROP TABLE IF EXISTS public.github_webhook_events CASCADE;
DROP TABLE IF EXISTS public.github_automation_preferences CASCADE;
DROP TABLE IF EXISTS public.github_repositories CASCADE;
DROP TABLE IF EXISTS public.github_installations CASCADE;

CREATE TABLE public.github_workspace_settings (
    workspace_id uuid NOT NULL,
    branch_format text NOT NULL DEFAULT 'username/identifier-title',
    link_commits_by_magic_words bool NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_workspace_settings_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    PRIMARY KEY (workspace_id)
);

CREATE TABLE public.github_installations (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    github_app_id bigint NOT NULL,
    github_installation_id bigint NOT NULL,
    account_id bigint NOT NULL,
    account_login text NOT NULL,
    account_type text NOT NULL,
    account_avatar_url text,
    repository_selection text NOT NULL,
    permissions jsonb NOT NULL DEFAULT '{}'::jsonb,
    events jsonb NOT NULL DEFAULT '[]'::jsonb,
    installed_by_user_id uuid,
    installed_by_github_user_id bigint,
    is_active bool NOT NULL DEFAULT true,
    suspended_at timestamptz,
    disconnected_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_installations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_installations_installed_by_user_id_fkey FOREIGN KEY (installed_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT github_installations_repository_selection_check CHECK (repository_selection IN ('all', 'selected')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_installations_github_installation_id_key
    ON public.github_installations USING btree (github_installation_id);
CREATE INDEX idx_github_installations_workspace_id
    ON public.github_installations USING btree (workspace_id);
CREATE INDEX idx_github_installations_account_login
    ON public.github_installations USING btree (account_login);

CREATE TABLE public.github_repositories (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    installation_id uuid NOT NULL,
    github_repository_id bigint NOT NULL,
    owner_id bigint NOT NULL,
    owner_login text NOT NULL,
    name text NOT NULL,
    full_name text NOT NULL,
    description text,
    html_url text NOT NULL,
    clone_url text NOT NULL,
    ssh_url text NOT NULL,
    default_branch text NOT NULL DEFAULT 'main',
    is_private bool NOT NULL DEFAULT false,
    is_archived bool NOT NULL DEFAULT false,
    is_disabled bool NOT NULL DEFAULT false,
    is_active bool NOT NULL DEFAULT true,
    last_synced_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_repositories_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_repositories_installation_id_fkey FOREIGN KEY (installation_id) REFERENCES public.github_installations(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_repositories_installation_id_github_repository_id_key
    ON public.github_repositories USING btree (installation_id, github_repository_id);
CREATE INDEX idx_github_repositories_workspace_id
    ON public.github_repositories USING btree (workspace_id);
CREATE INDEX idx_github_repositories_full_name
    ON public.github_repositories USING btree (full_name);

CREATE TABLE public.github_issue_sync_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    repository_id uuid NOT NULL,
    team_id uuid NOT NULL,
    sync_direction text NOT NULL DEFAULT 'inbound_only',
    is_active bool NOT NULL DEFAULT true,
    created_by_user_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_issue_sync_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_issue_sync_links_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE,
    CONSTRAINT github_issue_sync_links_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT github_issue_sync_links_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL,
    CONSTRAINT github_issue_sync_links_sync_direction_check CHECK (sync_direction IN ('inbound_only', 'bidirectional')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_issue_sync_links_active_team_id_key
    ON public.github_issue_sync_links USING btree (team_id)
    WHERE is_active = true;
CREATE UNIQUE INDEX github_issue_sync_links_active_repository_id_key
    ON public.github_issue_sync_links USING btree (repository_id)
    WHERE is_active = true;

CREATE TABLE public.github_team_workflow_rules (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    event_key text NOT NULL,
    target_status_id uuid,
    base_branch_pattern text,
    is_active bool NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_team_workflow_rules_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_team_workflow_rules_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE,
    CONSTRAINT github_team_workflow_rules_target_status_id_fkey FOREIGN KEY (target_status_id) REFERENCES public.statuses(status_id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_team_workflow_rules_team_event_branch_key
    ON public.github_team_workflow_rules USING btree (team_id, event_key, COALESCE(base_branch_pattern, ''));

CREATE TABLE public.github_story_links (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL,
    story_id uuid NOT NULL,
    repository_id uuid NOT NULL,
    external_type text NOT NULL,
    github_node_id text,
    github_id bigint,
    github_number integer,
    ref_name text,
    url text NOT NULL,
    title text,
    state text,
    sync_state text,
    last_synced_from text,
    last_sync_hash text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT github_story_links_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE,
    CONSTRAINT github_story_links_story_id_fkey FOREIGN KEY (story_id) REFERENCES public.stories(id) ON DELETE CASCADE,
    CONSTRAINT github_story_links_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.github_repositories(id) ON DELETE CASCADE,
    CONSTRAINT github_story_links_external_type_check CHECK (external_type IN ('issue', 'pull_request', 'branch', 'commit')),
    PRIMARY KEY (id)
);

CREATE INDEX idx_github_story_links_story_id
    ON public.github_story_links USING btree (story_id);
CREATE INDEX idx_github_story_links_repository_id
    ON public.github_story_links USING btree (repository_id);
CREATE UNIQUE INDEX github_story_links_unique_external_ref
    ON public.github_story_links USING btree (story_id, repository_id, external_type, COALESCE(github_id, 0), COALESCE(ref_name, ''));

CREATE TABLE public.github_webhook_events (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    delivery_id text NOT NULL,
    event_name text NOT NULL,
    action text,
    installation_external_id bigint,
    repository_external_id bigint,
    sender_external_id bigint,
    payload jsonb NOT NULL,
    processing_state text NOT NULL DEFAULT 'pending',
    attempts integer NOT NULL DEFAULT 0,
    error_message text,
    received_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    CONSTRAINT github_webhook_events_processing_state_check CHECK (processing_state IN ('pending', 'processed', 'ignored', 'failed')),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX github_webhook_events_delivery_id_key
    ON public.github_webhook_events USING btree (delivery_id);
CREATE INDEX idx_github_webhook_events_processing_state
    ON public.github_webhook_events USING btree (processing_state, received_at);
