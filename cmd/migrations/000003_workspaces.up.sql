-- 000003_workspaces.up.sql
CREATE TABLE public.workspaces (
    workspace_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    slug character varying(255) NOT NULL,
    color character varying(50) DEFAULT '#EA6060'::character varying NOT NULL,
    team_size character varying(20) DEFAULT '2-10'::character varying NOT NULL,
    trial_ends_on timestamp with time zone,
    avatar_url character varying(255),
    deleted_at timestamp without time zone,
    deleted_by uuid,
    last_accessed_at timestamp without time zone DEFAULT now(),
    inactivity_warning_sent_at timestamp without time zone,
    created_by uuid
);

ALTER TABLE ONLY public.workspaces ADD CONSTRAINT workspaces_pkey PRIMARY KEY (workspace_id);

CREATE UNIQUE INDEX workspaces_slug_key ON public.workspaces USING btree (slug);
CREATE INDEX idx_workspaces_created_at ON public.workspaces USING btree (created_at);
