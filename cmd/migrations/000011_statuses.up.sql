-- 000011_statuses.up.sql
CREATE TABLE public.statuses (
    status_id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    category character varying(255),
    order_index integer,
    workspace_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    team_id uuid NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    color character varying(16) DEFAULT '#6b665c'::character varying
);

ALTER TABLE ONLY public.statuses ADD CONSTRAINT statuses_pkey PRIMARY KEY (status_id);

ALTER TABLE ONLY public.statuses 
    ADD CONSTRAINT fk_statuses_team FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.statuses 
    ADD CONSTRAINT statuses_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE;

CREATE INDEX idx_statuses_category ON public.statuses USING btree (category);
CREATE INDEX idx_statuses_team ON public.statuses USING btree (team_id);
CREATE INDEX idx_statuses_workspace_order ON public.statuses USING btree (workspace_id, order_index);
CREATE UNIQUE INDEX unique_default_status_per_team ON public.statuses USING btree (team_id) WHERE (is_default = true);
