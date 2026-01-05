-- 000021_labels.up.sql
CREATE TABLE public.labels (
    label_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    team_id uuid,
    workspace_id uuid,
    color character varying(56) DEFAULT 'red'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE ONLY public.labels ADD CONSTRAINT labels_pkey PRIMARY KEY (label_id);

ALTER TABLE ONLY public.labels 
    ADD CONSTRAINT labels_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

ALTER TABLE ONLY public.labels 
    ADD CONSTRAINT labels_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_labels_label_id ON public.labels USING btree (label_id);
CREATE INDEX idx_labels_team_created ON public.labels USING btree (team_id, created_at DESC);
CREATE INDEX idx_labels_team_workspace ON public.labels USING btree (team_id, workspace_id);
CREATE INDEX idx_labels_workspace_created ON public.labels USING btree (workspace_id, created_at DESC);
