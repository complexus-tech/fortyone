-- 000034_github_installations.up.sql
CREATE TABLE public.github_installations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    workspace_id uuid NOT NULL,
    github_app_id bigint NOT NULL,
    installation_id bigint NOT NULL,
    account_type character varying(20) NOT NULL,
    account_login character varying(255) NOT NULL,
    account_id bigint NOT NULL,
    repository_selection character varying(10) NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    suspended_at timestamp with time zone,
    suspended_by character varying(255),
    permissions jsonb,
    events text[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.github_installations ADD CONSTRAINT github_installations_installation_id_key UNIQUE (installation_id);
ALTER TABLE ONLY public.github_installations ADD CONSTRAINT github_installations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.github_installations ADD CONSTRAINT github_installations_workspace_id_key UNIQUE (workspace_id);

ALTER TABLE ONLY public.github_installations 
    ADD CONSTRAINT github_installations_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_github_installations_account ON public.github_installations USING btree (account_type, account_login);
CREATE INDEX idx_github_installations_installation_id ON public.github_installations USING btree (installation_id);
CREATE INDEX idx_github_installations_workspace_id ON public.github_installations USING btree (workspace_id);
