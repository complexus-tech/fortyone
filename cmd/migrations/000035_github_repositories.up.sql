-- 000035_github_repositories.up.sql
CREATE TABLE public.github_repositories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    installation_id bigint NOT NULL,
    workspace_id uuid NOT NULL,
    github_repo_id bigint NOT NULL,
    name text NOT NULL,
    full_name text NOT NULL,
    description text,
    private boolean DEFAULT false NOT NULL,
    default_branch text DEFAULT 'main'::text NOT NULL,
    clone_url text NOT NULL,
    ssh_url text NOT NULL,
    webhook_id bigint,
    webhook_secret text,
    is_active boolean DEFAULT true,
    last_synced_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.github_repositories ADD CONSTRAINT github_repositories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.github_repositories ADD CONSTRAINT github_repositories_workspace_id_github_repo_id_key UNIQUE (workspace_id, github_repo_id);

ALTER TABLE ONLY public.github_repositories 
    ADD CONSTRAINT github_repositories_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_github_repositories_integration ON public.github_repositories USING btree (installation_id);
CREATE INDEX idx_github_repositories_workspace ON public.github_repositories USING btree (workspace_id);
