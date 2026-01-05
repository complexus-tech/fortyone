-- 000031_workspace_settings.up.sql
CREATE TABLE public.workspace_settings (
    workspace_id uuid NOT NULL,
    story_term character varying(50) DEFAULT 'story'::character varying NOT NULL,
    sprint_term character varying(50) DEFAULT 'sprint'::character varying NOT NULL,
    objective_term character varying(50) DEFAULT 'objective'::character varying NOT NULL,
    key_result_term character varying(50) DEFAULT 'key result'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    objective_enabled boolean DEFAULT true NOT NULL,
    key_result_enabled boolean DEFAULT true NOT NULL
);

ALTER TABLE ONLY public.workspace_settings ADD CONSTRAINT workspace_terminology_pkey PRIMARY KEY (workspace_id);

ALTER TABLE ONLY public.workspace_settings 
    ADD CONSTRAINT workspace_terminology_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;
